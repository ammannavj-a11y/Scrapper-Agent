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
		log.Warn("db not available", "err", err)
	}

	gin.SetMode(getenv("GIN_MODE", "release"))
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "notification-svc", "status": "ok", "version": "3.0.0"})
	})
	r.GET("/ready", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ready"}) })

	r.POST("/api/v1/notify/send", sendNotification)
	r.GET("/api/v1/notify/history", getHistory)

	srv := &http.Server{Addr: ":8095", Handler: r}
	go func() {
		log.Info("notification-svc starting", "port", 8095)
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

func sendNotification(c *gin.Context) {
	var req struct {
		GUVID       string `json:"guvid"`
		TenantID    string `json:"tenantId"`
		Channel     string `json:"channel"` // email | webhook | none
		MessageType string `json:"messageType"`
		Payload     interface{} `json:"payload"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	notifID := uuid.New().String()

	// In production: send via SMTP / webhook / push based on channel
	// For now: log and store
	log.Info("notification sent", "id", notifID, "type", req.MessageType, "channel", req.Channel, "guvid", req.GUVID)

	if db != nil {
		db.ExecContext(c.Request.Context(),
			`INSERT INTO notification_log(id,guvid,tenant_id,channel,message_type,status,sent_at)
			 VALUES(?,?,?,?,?,'sent',NOW())`,
			notifID, req.GUVID, req.TenantID, req.Channel, req.MessageType,
		)
	}

	c.JSON(200, gin.H{"notificationId": notifID, "status": "sent", "channel": req.Channel})
}

func getHistory(c *gin.Context) {
	guvid := c.Query("guvid")
	var logs []gin.H
	if db != nil && guvid != "" {
		rows, err := db.QueryContext(c.Request.Context(),
			`SELECT id, channel, message_type, status, sent_at FROM notification_log
			 WHERE guvid=? ORDER BY sent_at DESC LIMIT 20`, guvid)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id, ch, mt, st string
				var sa time.Time
				rows.Scan(&id, &ch, &mt, &st, &sa)
				logs = append(logs, gin.H{"id": id, "channel": ch, "messageType": mt, "status": st, "sentAt": sa})
			}
		}
	}
	if logs == nil {
		logs = []gin.H{}
	}
	c.JSON(200, gin.H{"logs": logs})
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
