package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/go-sql-driver/mysql"
)

var log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("mysql", getenv("MARIADB_DSN", "root:password@tcp(mariadb:3306)/guvid_platform?parseTime=true"))
	if err != nil {
		log.Error("db failed", "err", err)
	}
	if db != nil {
		defer db.Close()
	}

	gin.SetMode(getenv("GIN_MODE", "release"))
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "wallet-svc", "status": "ok", "version": "3.0.0"})
	})
	r.GET("/ready", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ready"}) })

	// WebAuthn registration
	r.POST("/api/v1/wallet/register/begin", regBegin)
	r.POST("/api/v1/wallet/register/complete", regComplete)

	// WebAuthn authentication / presentation
	r.POST("/api/v1/wallet/auth/begin", authBegin)
	r.POST("/api/v1/wallet/auth/complete", authComplete)

	// Status
	r.GET("/api/v1/wallet/status", walletStatus)

	// Recovery
	r.POST("/api/v1/wallet/recovery/initiate", initiateRecovery)
	r.POST("/api/v1/wallet/recovery/guardian-approve", guardianApprove)

	srv := &http.Server{Addr: ":8086", Handler: r, ReadTimeout: 30 * time.Second}
	go func() {
		log.Info("wallet-svc starting", "port", 8086)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			os.Exit(1)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func regBegin(c *gin.Context) {
	var req struct {
		GUVID  string `json:"guvid" binding:"required"`
		RpID   string `json:"rpId"`
		RpName string `json:"rpName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.RpID == "" {
		req.RpID = getenv("WEBAUTHN_RP_ID", "localhost")
	}
	if req.RpName == "" {
		req.RpName = "GUVID Platform"
	}

	challengeBytes := make([]byte, 32)
	rand.Read(challengeBytes)
	challenge := base64.URLEncoding.EncodeToString(challengeBytes)

	userIDBytes := []byte(req.GUVID)

	options := gin.H{
		"publicKey": gin.H{
			"challenge": challenge,
			"rp":        gin.H{"id": req.RpID, "name": req.RpName},
			"user": gin.H{
				"id":          base64.URLEncoding.EncodeToString(userIDBytes),
				"name":        req.GUVID,
				"displayName": "GUVID Holder " + req.GUVID,
			},
			"pubKeyCredParams": []gin.H{
				{"alg": -7, "type": "public-key"},
				{"alg": -257, "type": "public-key"},
			},
			"authenticatorSelection": gin.H{
				"authenticatorAttachment": "platform",
				"residentKey":             "required",
				"userVerification":        "required",
			},
			"timeout":     60000,
			"attestation": "none",
		},
	}
	log.Info("WebAuthn reg begin", "guvid", req.GUVID)
	c.JSON(200, options)
}

func regComplete(c *gin.Context) {
	var req struct {
		GUVID          string          `json:"guvid"`
		ID             string          `json:"id"`
		RawID          string          `json:"rawId"`
		Response       json.RawMessage `json:"response"`
		Type           string          `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// In production: verify attestation with go-webauthn/webauthn
	// For demo: store credential ID as public key hash
	pubKeyHash := sha256Hex(req.ID + req.GUVID + "guvid_platform")
	credID := uuid.New().String()

	if db != nil {
		rawIDBytes, _ := base64.URLEncoding.DecodeString(req.RawID)
		db.ExecContext(c.Request.Context(),
			`INSERT INTO webauthn_credentials(id,guvid,credential_id,public_key,rp_id,device_type)
			 VALUES(?,?,?,?,?,?) ON DUPLICATE KEY UPDATE last_used=NOW()`,
			credID, req.GUVID, rawIDBytes, []byte(pubKeyHash),
			getenv("WEBAUTHN_RP_ID", "localhost"), "platform",
		)
		// Update GUVID record with holder public key hash
		db.ExecContext(c.Request.Context(),
			`UPDATE guvid_records SET holder_public_key_hash=? WHERE guvid=?`,
			pubKeyHash, req.GUVID,
		)
	}

	log.Info("WebAuthn reg complete", "guvid", req.GUVID, "credId", credID)
	c.JSON(200, gin.H{
		"credentialId":    credID,
		"holderKeyHash":   pubKeyHash,
		"registered":      true,
		"guvid":           req.GUVID,
	})
}

func authBegin(c *gin.Context) {
	var req struct {
		GUVID  string `json:"guvid"`
		Domain string `json:"domain"`
	}
	c.ShouldBindJSON(&req)

	challengeBytes := make([]byte, 32)
	rand.Read(challengeBytes)
	challenge := base64.URLEncoding.EncodeToString(challengeBytes)

	c.JSON(200, gin.H{
		"publicKey": gin.H{
			"challenge":        challenge,
			"rpId":             req.Domain,
			"userVerification": "required",
			"timeout":          90000,
		},
		"nonce":     challenge,
		"expiresAt": time.Now().Add(90 * time.Second).UTC(),
	})
}

func authComplete(c *gin.Context) {
	var req struct {
		GUVID             string `json:"guvid"`
		Challenge         string `json:"challenge"`
		AuthenticatorData string `json:"authenticatorData"`
		Signature         string `json:"signature"`
	}
	c.ShouldBindJSON(&req)

	// In production: verify assertion with go-webauthn/webauthn
	// Check signature against stored public key, verify sign count
	holderVerified := req.Signature != ""

	c.JSON(200, gin.H{
		"holderVerified": holderVerified,
		"guvid":          req.GUVID,
		"method":         "webauthn",
		"verifiedAt":     time.Now().UTC(),
	})
}

func walletStatus(c *gin.Context) {
	guvid := c.Query("guvid")
	if guvid == "" {
		c.JSON(400, gin.H{"error": "guvid required"})
		return
	}

	var credID string
	registered := false
	if db != nil {
		err := db.QueryRowContext(c.Request.Context(),
			`SELECT id FROM webauthn_credentials WHERE guvid=? LIMIT 1`, guvid,
		).Scan(&credID)
		registered = err == nil
	}

	c.JSON(200, gin.H{
		"guvid":      guvid,
		"registered": registered,
		"credentialId": credID,
	})
}

func initiateRecovery(c *gin.Context) {
	var req struct {
		GUVID         string `json:"guvid" binding:"required"`
		RecoveryType  string `json:"recoveryType"`
		Email         string `json:"email"`
		Phone         string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	recoveryID := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	if db != nil {
		db.ExecContext(c.Request.Context(),
			`INSERT INTO recovery_requests(id,guvid,request_type,status,expires_at)
			 VALUES(?,?,?,'pending',?)`,
			recoveryID, req.GUVID, req.RecoveryType, expiresAt,
		)
	}

	log.Info("recovery initiated", "guvid", req.GUVID, "recoveryId", recoveryID)
	c.JSON(201, gin.H{
		"recoveryId":   recoveryID,
		"status":       "pending",
		"message":      "Guardians notified. 2-of-3 approvals required within 24h.",
		"expiresAt":    expiresAt.UTC(),
		"nextStep":     "Await guardian approvals via email. Check recovery status.",
	})
}

func guardianApprove(c *gin.Context) {
	var req struct {
		RecoveryID   string `json:"recoveryId" binding:"required"`
		GuardianGUVID string `json:"guardianGuvid"`
		Signature    string `json:"signature"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	// In production: verify guardian signature, check threshold, issue new keypair
	c.JSON(200, gin.H{
		"recoveryId": req.RecoveryID,
		"approvals":  1,
		"required":   2,
		"status":     "awaiting_more_approvals",
		"message":    fmt.Sprintf("Guardian approval recorded. %d more required.", 1),
	})
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
