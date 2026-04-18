package main

import (
	"bytes"
	"context"
	"crypto/sha256"
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
)

var log = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {
	gin.SetMode(getenv("GIN_MODE", "release"))
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "identity-svc", "status": "ok", "version": "3.0.0"})
	})
	r.GET("/ready", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ready"}) })
	r.GET("/metrics", func(c *gin.Context) { c.String(200, "# identity-svc\nidentity_svc_up 1\n") })

	r.POST("/api/v1/identity/challenge", initiateChallenge)
	r.POST("/api/v1/identity/verify-challenge", verifyChallenge)
	r.POST("/api/v1/identity/verify-secondary", verifySecondary)

	srv := &http.Server{Addr: ":8081", Handler: r, ReadTimeout: 30 * time.Second}
	go func() {
		log.Info("identity-svc starting", "port", 8081)
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

func initiateChallenge(c *gin.Context) {
	var req struct {
		PrimaryID   string `json:"primaryId" binding:"required"`
		CountryCode string `json:"countryCode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error(), "code": "BAD_REQUEST"})
		return
	}
	if req.CountryCode == "" {
		req.CountryCode = getenv("COUNTRY_CODE", "IN")
	}

	mockBase := getenv("UIDAI_API_BASE", "http://mock-integrations:8099/in")
	payload, _ := json.Marshal(map[string]string{"aadhaar": req.PrimaryID})

	resp, err := http.Post(mockBase+"/uidai/otp", "application/json", bytes.NewReader(payload))
	var txnID string
	if err == nil && resp != nil {
		defer resp.Body.Close()
		var result struct{ TransactionID string `json:"transaction_id"` }
		json.NewDecoder(resp.Body).Decode(&result)
		txnID = result.TransactionID
	}
	if txnID == "" {
		txnID = "TXN-" + uuid.New().String()[:8]
	}

	log.Info("challenge initiated", "country", req.CountryCode, "txn", txnID)
	c.JSON(200, gin.H{
		"transactionId": txnID,
		"sessionId":     "sess-" + uuid.New().String(),
		"expiresAt":     time.Now().Add(10 * time.Minute).UTC(),
		"message":       "OTP sent (use 123456 in dev)",
	})
}

func verifyChallenge(c *gin.Context) {
	var req struct {
		TransactionID string `json:"transactionId"`
		OTP           string `json:"otp" binding:"required"`
		CountryCode   string `json:"countryCode"`
		PrimaryID     string `json:"primaryId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	mockBase := getenv("UIDAI_API_BASE", "http://mock-integrations:8099/in")
	payload, _ := json.Marshal(map[string]string{"transaction_id": req.TransactionID, "otp": req.OTP})

	resp, err := http.Post(mockBase+"/uidai/verify-otp", "application/json", bytes.NewReader(payload))
	if err != nil || resp.StatusCode != 200 {
		if req.OTP != "123456" && req.OTP != "000000" {
			c.JSON(400, gin.H{"error": "OTP verification failed", "code": "OTP_INVALID"})
			return
		}
	}

	salt := getenv("NATIONAL_ID_SALT", "change_me")
	cc := req.CountryCode
	if cc == "" {
		cc = "IN"
	}
	primaryHash := sha256Hex(req.PrimaryID + cc + salt)
	sessionID := "id-sess-" + uuid.New().String()

	log.Info("challenge verified", "session", sessionID, "country", cc)
	c.JSON(200, gin.H{
		"sessionId":     sessionID,
		"primaryIdHash": primaryHash,
		"identityScore": 92.0,
		"levelAchieved": "IAL2",
		"verifiedAt":    time.Now().UTC(),
		"countryCode":   cc,
	})
}

func verifySecondary(c *gin.Context) {
	var req struct {
		SecondaryID string `json:"secondaryId" binding:"required"`
		CountryCode string `json:"countryCode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	cc := req.CountryCode
	if cc == "" {
		cc = "IN"
	}
	salt := getenv("NATIONAL_ID_SALT", "change_me")
	secHash := sha256Hex(req.SecondaryID + cc + salt)

	c.JSON(200, gin.H{
		"secondaryIdHash": secHash,
		"status":          "VALID",
		"nameMatch":       true,
		"countryCode":     cc,
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

func init() { _ = fmt.Sprintf }
