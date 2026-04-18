package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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

	gin.SetMode(getenv("GIN_MODE", "release"))
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "fraud-svc", "status": "ok", "version": "3.0.0"})
	})
	r.GET("/ready", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ready"}) })

	r.POST("/api/v1/fraud/check", fraudCheck)
	r.POST("/api/v1/fraud/report", reportFraud)
	r.GET("/api/v1/fraud/incidents", listIncidents)
	r.GET("/api/v1/fraud/stats", fraudStats)

	srv := &http.Server{Addr: ":8088", Handler: r}
	go func() {
		log.Info("fraud-svc starting", "port", 8088)
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

type FraudCheckReq struct {
	IdentityHash    string `json:"identityHash"`
	CountryCode     string `json:"countryCode"`
	GUVID           string `json:"guvid"`
	IPAddress       string `json:"ipAddress"`
	DeviceFingerprint string `json:"deviceFingerprint"`
	Action          string `json:"action"` // issue | verify | login
}

func fraudCheck(c *gin.Context) {
	var req FraudCheckReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	riskScore := 0.0
	signals := []string{}
	result := "pass"

	// Velocity check (simplified — use Redis in production)
	if db != nil {
		var attempts int
		db.QueryRowContext(c.Request.Context(),
			`SELECT COUNT(*) FROM fraud_checks WHERE identity_hash=? AND checked_at > DATE_SUB(NOW(), INTERVAL 24 HOUR)`,
			req.IdentityHash,
		).Scan(&attempts)
		if attempts >= 3 {
			riskScore += 40
			signals = append(signals, fmt.Sprintf("velocity_breach:attempts=%d", attempts))
		}
	}

	// Blacklist check
	if db != nil {
		var blacklisted int
		db.QueryRowContext(c.Request.Context(),
			`SELECT COUNT(*) FROM fraud_checks WHERE identity_hash=? AND result='block'`,
			req.IdentityHash,
		).Scan(&blacklisted)
		if blacklisted > 0 {
			riskScore += 60
			signals = append(signals, "blacklisted")
		}
	}

	if riskScore >= 70 {
		result = "block"
	} else if riskScore >= 40 {
		result = "flag"
	}

	checkID := uuid.New().String()
	if db != nil {
		db.ExecContext(c.Request.Context(),
			`INSERT INTO fraud_checks(id,country_code,identity_hash,check_type,result,risk_score)
			 VALUES(?,?,?,'composite',?,?)`,
			checkID, req.CountryCode, req.IdentityHash, result, riskScore,
		)
	}

	log.Info("fraud check", "result", result, "riskScore", riskScore, "country", req.CountryCode)
	c.JSON(200, gin.H{
		"checkId":   checkID,
		"result":    result,
		"riskScore": riskScore,
		"signals":   signals,
		"allowed":   result == "pass",
	})
}

func reportFraud(c *gin.Context) {
	var req struct {
		GUVID       string `json:"guvid"`
		ReportType  string `json:"reportType"`
		Description string `json:"description"`
	}
	c.ShouldBindJSON(&req)

	reportID := uuid.New().String()
	if db != nil {
		db.ExecContext(c.Request.Context(),
			`INSERT INTO fraud_reports(id,guvid,report_type,details,status) VALUES(?,?,?,?,'open')`,
			reportID, req.GUVID, req.ReportType, req.Description,
		)
	}
	c.JSON(201, gin.H{"reportId": reportID, "status": "open", "message": "Fraud report logged and queued for review"})
}

func listIncidents(c *gin.Context) {
	c.JSON(200, gin.H{"incidents": []gin.H{
		{"id": "inc-001", "type": "velocity_breach", "severity": "high", "riskScore": 78.5, "status": "open", "detectedAt": time.Now().Add(-2 * time.Hour)},
		{"id": "inc-002", "type": "synthetic_identity", "severity": "critical", "riskScore": 92.1, "status": "investigating", "detectedAt": time.Now().Add(-4 * time.Hour)},
	}})
}

func fraudStats(c *gin.Context) {
	var total, blocked, flagged int
	if db != nil {
		db.QueryRowContext(c.Request.Context(),
			`SELECT COUNT(*), SUM(result='block'), SUM(result='flag') FROM fraud_checks`).
			Scan(&total, &blocked, &flagged)
	}
	c.JSON(200, gin.H{
		"totalChecks": total, "blocked": blocked, "flagged": flagged,
		"blockRate": safeDiv(float64(blocked), float64(total)),
	})
}

func safeDiv(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
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
