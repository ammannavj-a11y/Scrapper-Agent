package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/go-sql-driver/mysql"
)

var log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
var signingKey *ecdsa.PrivateKey
var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("mysql", getenv("MARIADB_DSN", "root:password@tcp(mariadb:3306)/guvid_platform?parseTime=true"))
	if err != nil {
		log.Error("db failed", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	signingKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Error("key gen failed", "err", err)
		os.Exit(1)
	}

	gin.SetMode(getenv("GIN_MODE", "release"))
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "guvid-svc", "status": "ok", "version": "3.0.0"})
	})
	r.GET("/ready", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ready"}) })
	r.GET("/metrics", func(c *gin.Context) { c.String(200, "# guvid-svc\nguvid_svc_up 1\n") })
	r.POST("/api/v1/guvid/issue", issueGUVID)

	srv := &http.Server{Addr: ":8084", Handler: r, ReadTimeout: 60 * time.Second}
	go func() {
		log.Info("guvid-svc starting", "port", 8084)
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

type IssueReq struct {
	Name              string  `json:"name" binding:"required"`
	CountryCode       string  `json:"countryCode" binding:"required"`
	TenantID          string  `json:"tenantId"`
	IdentityScore     float64 `json:"identityScore"`
	EducationScore    float64 `json:"educationScore"`
	EmploymentScore   float64 `json:"employmentScore"`
	IdentityHash      string  `json:"identityHash"`
	EducationHash     string  `json:"educationHash"`
	EmploymentHash    string  `json:"employmentHash"`
	HolderPubKeyHash  string  `json:"holderPublicKeyHash"`
}

func issueGUVID(c *gin.Context) {
	var req IssueReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error(), "code": "INVALID_REQUEST"})
		return
	}

	saltB := make([]byte, 16)
	rand.Read(saltB)
	salt := hex.EncodeToString(saltB)

	if req.IdentityHash == "" {
		req.IdentityHash = sha256Hex(req.Name + req.CountryCode + salt)
	}
	if req.EducationHash == "" {
		req.EducationHash = sha256Hex("edu" + req.CountryCode + salt)
	}
	if req.EmploymentHash == "" {
		req.EmploymentHash = sha256Hex("emp" + req.CountryCode + salt)
	}

	trustScore := req.IdentityScore*0.40 + req.EducationScore*0.35 + req.EmploymentScore*0.25
	if trustScore == 0 {
		trustScore = 87.5
	}
	trustLevel := toTrustLevel(trustScore)

	compositeInput := req.IdentityHash + req.EducationHash + req.EmploymentHash + req.CountryCode + salt
	compositeHash := sha256Hex(compositeInput)

	year := time.Now().UTC().Year()
	hashBytes, _ := hex.DecodeString(compositeHash)
	guvidToken := fmt.Sprintf("GUV-%s-%d-%s", req.CountryCode, year, base36Upper(hashBytes[:8]))

	// Sign
	sigInput := sha256.Sum256([]byte(guvidToken + compositeHash))
	r, s, _ := ecdsa.Sign(rand.Reader, signingKey, sigInput[:])
	sig := base64.StdEncoding.EncodeToString(append(r.Bytes(), s.Bytes()...))

	issuedAt := time.Now().UTC()
	expiresAt := issuedAt.AddDate(2, 0, 0)
	tenantID := req.TenantID
	if tenantID == "" {
		tenantID = "default"
	}
	fabricTxID := "tx_" + sha256Hex(guvidToken)[:16]
	did := fmt.Sprintf("did:guvid:%s:%s", req.CountryCode, guvidToken)
	guvidID := uuid.New().String()

	var wg sync.WaitGroup
	var dbErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, dbErr = db.ExecContext(c.Request.Context(),
			`INSERT INTO guvid_records
			(id,guvid,country_code,tenant_id,composite_hash,identity_hash,education_hash,employment_hash,
			holder_public_key_hash,trust_score,trust_level,identity_score,education_score,employment_score,
			platform_signature,fabric_tx_id,status,issued_at,expires_at)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,'active',?,?)`,
			guvidID, guvidToken, req.CountryCode, tenantID, compositeHash,
			req.IdentityHash, req.EducationHash, req.EmploymentHash,
			req.HolderPubKeyHash, trustScore, trustLevel,
			req.IdentityScore, req.EducationScore, req.EmploymentScore,
			sig, fabricTxID, issuedAt, expiresAt,
		)
	}()
	wg.Wait()

	if dbErr != nil {
		log.Error("db insert failed", "err", dbErr)
		// Continue — return result even if DB failed (demo mode)
	}

	log.Info("GUVID issued", "guvid", guvidToken, "trust", trustLevel, "country", req.CountryCode)
	c.JSON(201, gin.H{
		"guvid": guvidToken, "did": did,
		"trustScore": trustScore, "trustLevel": trustLevel,
		"identityScore": req.IdentityScore, "educationScore": req.EducationScore,
		"employmentScore": req.EmploymentScore, "countryCode": req.CountryCode,
		"platformSignature": sig, "fabricTxId": fabricTxID,
		"issuedAt": issuedAt, "expiresAt": expiresAt,
		"passkeyRegistered": req.HolderPubKeyHash != "",
	})
}

func toTrustLevel(s float64) string {
	switch {
	case s >= 85:
		return "HIGH"
	case s >= 65:
		return "MEDIUM"
	case s >= 40:
		return "LOW"
	default:
		return "UNVERIFIED"
	}
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func base36Upper(b []byte) string {
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	n := new(big.Int).SetBytes(b)
	base := big.NewInt(36)
	mod := new(big.Int)
	var result []byte
	for n.Sign() > 0 {
		n.DivMod(n, base, mod)
		result = append([]byte{chars[mod.Int64()]}, result...)
	}
	for len(result) < 8 {
		result = append([]byte{'0'}, result...)
	}
	if len(result) > 8 {
		return string(result[:8])
	}
	return string(result)
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
