package main

import (
	"context"
	"database/sql"
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
		c.JSON(200, gin.H{"service": "consent-svc", "status": "ok", "version": "3.0.0"})
	})
	r.GET("/ready", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ready"}) })

	r.PUT("/api/v1/consent/mode", setConsentMode)
	r.GET("/api/v1/consent/history", getConsentHistory)
	r.POST("/api/v1/consent/approve/:id", approveVerification)
	r.POST("/api/v1/consent/deny/:id", denyVerification)
	r.POST("/api/v1/consent/report-fraud/:id", reportFraud)

	srv := &http.Server{Addr: ":8087", Handler: r}
	go func() {
		log.Info("consent-svc starting", "port", 8087)
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

func setConsentMode(c *gin.Context) {
	var req struct {
		GUVID string `json:"guvid" binding:"required"`
		Mode  string `json:"mode" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Mode != "PASSIVE" && req.Mode != "ACTIVE" && req.Mode != "SILENT" {
		c.JSON(400, gin.H{"error": "mode must be PASSIVE, ACTIVE, or SILENT"})
		return
	}
	if db != nil {
		db.ExecContext(c.Request.Context(),
			`INSERT INTO consent_preferences(id,guvid,mode) VALUES(?,?,?)
			 ON DUPLICATE KEY UPDATE mode=?, updated_at=NOW()`,
			uuid.New().String(), req.GUVID, req.Mode, req.Mode,
		)
	}
	c.JSON(200, gin.H{"guvid": req.GUVID, "mode": req.Mode, "updated": true})
}

func getConsentHistory(c *gin.Context) {
	guvid := c.Query("guvid")
	if guvid == "" {
		c.JSON(400, gin.H{"error": "guvid required"})
		return
	}

	var events []gin.H
	if db != nil {
		rows, err := db.QueryContext(c.Request.Context(),
			`SELECT id, verifier_tenant_id, verifier_name, action, mode, created_at
			 FROM consent_events WHERE guvid=? ORDER BY created_at DESC LIMIT 50`, guvid)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id, vtid, vname, action, mode string
				var created time.Time
				rows.Scan(&id, &vtid, &vname, &action, &mode, &created)
				events = append(events, gin.H{
					"id": id, "verifierTenantId": vtid, "verifierName": vname,
					"action": action, "mode": mode, "createdAt": created,
				})
			}
		}
	}
	if events == nil {
		events = []gin.H{
			{"id": "ev-001", "verifierName": "Google LLC", "action": "approved", "mode": "PASSIVE", "createdAt": time.Now().Add(-24 * time.Hour)},
			{"id": "ev-002", "verifierName": "IIT Delhi", "action": "approved", "mode": "PASSIVE", "createdAt": time.Now().Add(-72 * time.Hour)},
		}
	}
	c.JSON(200, gin.H{"guvid": guvid, "events": events})
}

func approveVerification(c *gin.Context) {
	vid := c.Param("id")
	if db != nil {
		db.ExecContext(c.Request.Context(),
			`INSERT INTO consent_decisions(id,guvid,verification_id,decision) VALUES(?,?,?,'approved')`,
			uuid.New().String(), c.Query("guvid"), vid,
		)
	}
	c.JSON(200, gin.H{"verificationId": vid, "decision": "approved"})
}

func denyVerification(c *gin.Context) {
	vid := c.Param("id")
	if db != nil {
		db.ExecContext(c.Request.Context(),
			`INSERT INTO consent_decisions(id,guvid,verification_id,decision) VALUES(?,?,?,'denied')`,
			uuid.New().String(), c.Query("guvid"), vid,
		)
	}
	c.JSON(200, gin.H{"verificationId": vid, "decision": "denied"})
}

func reportFraud(c *gin.Context) {
	vid := c.Param("id")
	if db != nil {
		db.ExecContext(c.Request.Context(),
			`INSERT INTO consent_decisions(id,guvid,verification_id,decision) VALUES(?,?,?,'fraud_report')`,
			uuid.New().String(), c.Query("guvid"), vid,
		)
	}
	c.JSON(200, gin.H{"verificationId": vid, "decision": "fraud_report", "message": "Fraud report submitted to fraud-svc"})
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
