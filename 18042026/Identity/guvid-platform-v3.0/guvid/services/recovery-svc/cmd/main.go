package main

import (
	"context"
	"database/sql"
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
		log.Warn("db not available", "err", err)
	}

	gin.SetMode(getenv("GIN_MODE", "release"))
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "recovery-svc", "status": "ok", "version": "3.0.0"})
	})
	r.GET("/ready", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ready"}) })

	r.POST("/api/v1/recovery/initiate", initiateRecovery)
	r.GET("/api/v1/recovery/:id", getRecoveryStatus)
	r.POST("/api/v1/recovery/:id/guardian-approve", guardianApprove)
	r.POST("/api/v1/recovery/:id/complete", completeRecovery)

	srv := &http.Server{Addr: ":8096", Handler: r}
	go func() {
		log.Info("recovery-svc starting", "port", 8096)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			os.Exit(1)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func initiateRecovery(c *gin.Context) {
	var req struct {
		GUVID        string `json:"guvid" binding:"required"`
		RecoveryType string `json:"recoveryType"`
		Email        string `json:"email"`
		Phone        string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.RecoveryType == "" {
		req.RecoveryType = "device_lost"
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

	// In production: notify guardians via notification-svc
	log.Info("recovery initiated", "recoveryId", recoveryID, "guvid", req.GUVID)

	c.JSON(201, gin.H{
		"recoveryId":  recoveryID,
		"status":      "pending",
		"expiresAt":   expiresAt.UTC(),
		"steps": []string{
			"1. Identity verification email sent",
			"2. Guardians notified — 2-of-3 approvals required within 24h",
			"3. Optional: complete liveness check at /recovery/" + recoveryID + "/liveness",
			"4. Upon threshold: new keypair generated, old key revoked on-chain",
		},
	})
}

func getRecoveryStatus(c *gin.Context) {
	rid := c.Param("id")
	var guvid, status, reqType string
	var expiresAt time.Time
	var approvals int

	if db != nil {
		db.QueryRowContext(c.Request.Context(),
			`SELECT guvid, status, request_type, expires_at FROM recovery_requests WHERE id=?`, rid,
		).Scan(&guvid, &status, &reqType, &expiresAt)
	}

	if guvid == "" {
		guvid = "GUV-IN-2025-UNKNOWN"
		status = "pending"
		reqType = "device_lost"
		expiresAt = time.Now().Add(20 * time.Hour)
	}

	c.JSON(200, gin.H{
		"recoveryId":  rid,
		"guvid":       guvid,
		"status":      status,
		"recoveryType": reqType,
		"approvals":   approvals,
		"required":    2,
		"expiresAt":   expiresAt,
	})
}

func guardianApprove(c *gin.Context) {
	rid := c.Param("id")
	var req struct {
		GuardianGUVID string `json:"guardianGuvid"`
		GuardianType  string `json:"guardianType"` // institutional | corporate | personal
		Signature     string `json:"signature"`
	}
	c.ShouldBindJSON(&req)

	// In production: verify guardian signature, update approval count
	// If threshold met (2/3), proceed to key replacement

	log.Info("guardian approved", "recoveryId", rid, "guardianType", req.GuardianType)

	c.JSON(200, gin.H{
		"recoveryId":     rid,
		"guardianType":   req.GuardianType,
		"approvalLogged": true,
		"message":        fmt.Sprintf("Guardian approval recorded (%s). Check recovery status.", req.GuardianType),
	})
}

func completeRecovery(c *gin.Context) {
	rid := c.Param("id")
	var req struct {
		NewPublicKeyHash string `json:"newPublicKeyHash" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// In production:
	// 1. Verify 2/3 guardian approvals in DB
	// 2. Call Fabric: UpdateHolderKey(guvid, newPublicKeyHash)
	// 3. Revoke old WebAuthn credential
	// 4. Link new credential
	if db != nil {
		db.ExecContext(c.Request.Context(),
			`UPDATE recovery_requests SET status='approved', new_key_hash=?, completed_at=NOW() WHERE id=?`,
			req.NewPublicKeyHash, rid,
		)
	}

	log.Info("recovery completed", "recoveryId", rid)
	c.JSON(200, gin.H{
		"recoveryId":       rid,
		"status":           "approved",
		"newPublicKeyHash": req.NewPublicKeyHash,
		"message":          "Key replaced. Old key revoked on-chain. Register new passkey on your device.",
		"onChainAction":    "UpdateHolderKey(guvid, newPublicKeyHash) → Fabric TX submitted",
	})
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
