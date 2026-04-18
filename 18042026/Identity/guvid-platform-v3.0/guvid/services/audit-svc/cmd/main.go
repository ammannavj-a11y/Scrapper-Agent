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
		c.JSON(200, gin.H{"service": "audit-svc", "status": "ok", "version": "3.0.0"})
	})
	r.GET("/ready", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ready"}) })

	r.GET("/api/v1/audit/trail", getAuditTrail)
	r.POST("/api/v1/audit/log", logEvent)

	srv := &http.Server{Addr: ":8097", Handler: r}
	go func() {
		log.Info("audit-svc starting", "port", 8097)
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

func getAuditTrail(c *gin.Context) {
	var events []gin.H
	if db != nil {
		rows, _ := db.QueryContext(c.Request.Context(),
			`SELECT id, action, actor_id, resource_type, resource_id, created_at
			 FROM audit_log ORDER BY created_at DESC LIMIT 100`)
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var id, action, actor, rtype, rid string
				var created time.Time
				rows.Scan(&id, &action, &actor, &rtype, &rid, &created)
				events = append(events, gin.H{
					"id": id, "action": action, "actorId": actor,
					"resourceType": rtype, "resourceId": rid, "createdAt": created,
				})
			}
		}
	}
	if events == nil {
		events = []gin.H{}
	}
	c.JSON(200, gin.H{"events": events})
}

func logEvent(c *gin.Context) {
	var req struct {
		Action       string `json:"action"`
		ActorID      string `json:"actorId"`
		ResourceType string `json:"resourceType"`
		ResourceID   string `json:"resourceId"`
		Details      interface{} `json:"details"`
	}
	c.ShouldBindJSON(&req)
	if db != nil {
		db.ExecContext(c.Request.Context(),
			`INSERT INTO audit_log(action, actor_id, resource_type, resource_id) VALUES(?,?,?,?)`,
			req.Action, req.ActorID, req.ResourceType, req.ResourceID,
		)
	}
	c.JSON(200, gin.H{"logged": true})
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
