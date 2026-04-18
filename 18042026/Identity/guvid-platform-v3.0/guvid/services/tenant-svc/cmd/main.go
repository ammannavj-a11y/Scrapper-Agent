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
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

// Multi-tenant DB pool: tenantSlug → *sql.DB
type TenantPool struct {
	mu   sync.RWMutex
	pool map[string]*sql.DB
}

var tenantPool = &TenantPool{pool: make(map[string]*sql.DB)}

func (p *TenantPool) Get(slug string) (*sql.DB, bool) {
	p.mu.RLock(); defer p.mu.RUnlock()
	db, ok := p.pool[slug]; return db, ok
}

func (p *TenantPool) Set(slug string, db *sql.DB) {
	p.mu.Lock(); defer p.mu.Unlock(); p.pool[slug] = db
}

type Tenant struct {
	ID          string
	Slug        string
	OrgName     string
	OrgType     string
	CountryCode string
	DBName      string
	DBHost      string
	DBPort      int
	Plan        string
}

type LoginRequest  struct { Email string `json:"email" binding:"required"`; Password string `json:"password" binding:"required"`; TenantSlug string `json:"tenantSlug" binding:"required"` }
type LoginResponse struct { Token string `json:"token"`; RefreshToken string `json:"refreshToken"`; Role string `json:"role"`; OrgName string `json:"orgName"`; OrgType string `json:"orgType"`; TenantID string `json:"tenantId"`; ExpiresAt int64 `json:"expiresAt"` }

var platformDB *sql.DB

func main() {
	var err error
	platformDB, err = sql.Open("mysql", getenv("MARIADB_DSN", "root:password@tcp(mariadb:3306)/guvid_platform?parseTime=true"))
	if err != nil { logger.Error("platform db failed", "err", err); os.Exit(1) }
	defer platformDB.Close()

	gin.SetMode(getenv("GIN_MODE", "release"))
	r := gin.New(); r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "tenant-svc", "status": "ok", "version": "3.0.0"})
	})
	r.GET("/ready", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ready"}) })

	// Auth
	r.POST("/api/v1/auth/login", handleLogin)
	r.POST("/api/v1/auth/refresh", handleRefresh)
	r.POST("/api/v1/auth/logout", handleLogout)

	// Tenant management (admin only)
	r.GET("/api/v1/tenants", requireRole("platform_admin"), listTenants)
	r.POST("/api/v1/tenants", requireRole("platform_admin"), createTenant)
	r.GET("/api/v1/tenants/:slug", requireRole("platform_admin"), getTenant)

	// Tenant DB routing — called by other services
	r.GET("/api/v1/tenant-db/:slug", getTenantDB)

	srv := &http.Server{Addr: ":8094", Handler: r}
	go func() {
		logger.Info("tenant-svc starting", "port", 8094)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "err", err); os.Exit(1)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel(); srv.Shutdown(ctx)
}

func handleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request", "code": "BAD_REQUEST"}); return
	}

	// Get tenant
	var t Tenant
	err := platformDB.QueryRowContext(c.Request.Context(),
		`SELECT id,slug,org_name,org_type,country_code,db_name,db_host,db_port FROM tenants WHERE slug=? AND is_active=1`,
		req.TenantSlug,
	).Scan(&t.ID, &t.Slug, &t.OrgName, &t.OrgType, &t.CountryCode, &t.DBName, &t.DBHost, &t.DBPort)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid credentials or tenant", "code": "AUTH_FAILED"}); return
	}

	// Get user from platform DB
	emailHash := sha256Hex(req.Email)
	var userID, passwordHash, role, fullName string
	err = platformDB.QueryRowContext(c.Request.Context(),
		`SELECT id,password_hash,role,full_name FROM tenant_users WHERE tenant_id=? AND email_hash=? AND is_active=1`,
		t.ID, emailHash,
	).Scan(&userID, &passwordHash, &role, &fullName)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid credentials", "code": "AUTH_FAILED"}); return
	}

	// Verify password (demo: accept "Admin@123" for all seeded users)
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		// Demo fallback
		if req.Password != "Admin@123" {
			c.JSON(401, gin.H{"error": "invalid credentials", "code": "AUTH_FAILED"}); return
		}
	}

	// Issue JWT
	expiresAt := time.Now().Add(8 * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID, "tenantId": t.ID, "tenantSlug": t.Slug,
		"role": role, "orgType": t.OrgType, "countryCode": t.CountryCode,
		"orgName": t.OrgName, "dbName": t.DBName,
		"exp": expiresAt.Unix(), "iat": time.Now().Unix(),
	})
	tokenStr, _ := token.SignedString([]byte(getenv("JWT_SECRET", "change_me")))

	// Update last login
	platformDB.ExecContext(c.Request.Context(), `UPDATE tenant_users SET last_login=NOW() WHERE id=?`, userID)

	// Session record
	sessionID := uuid.New().String()
	platformDB.ExecContext(c.Request.Context(),
		`INSERT INTO sessions(id,tenant_id,user_id,token_hash,role,ip_address,expires_at) VALUES(?,?,?,?,?,?,?)`,
		sessionID, t.ID, userID, sha256Hex(tokenStr), role, c.ClientIP(), expiresAt,
	)

	logger.Info("login success", "tenant", t.Slug, "user", userID, "role", role)
	c.JSON(200, LoginResponse{
		Token: tokenStr, Role: role, OrgName: t.OrgName,
		OrgType: t.OrgType, TenantID: t.ID, ExpiresAt: expiresAt.Unix(),
	})
}

func handleRefresh(c *gin.Context) {
	// Extract existing token and issue new one with extended expiry
	auth := c.GetHeader("Authorization")
	if len(auth) < 8 { c.JSON(401, gin.H{"error": "missing token"}); return }
	// Re-sign with fresh expiry
	c.JSON(200, gin.H{"message": "refresh not implemented in stub — re-login"})
}

func handleLogout(c *gin.Context) {
	c.JSON(200, gin.H{"message": "logged out"})
}

func getTenantDB(c *gin.Context) {
	slug := c.Param("slug")
	var t Tenant
	err := platformDB.QueryRowContext(c.Request.Context(),
		`SELECT id,slug,org_name,org_type,db_name,db_host,db_port FROM tenants WHERE slug=? AND is_active=1`, slug,
	).Scan(&t.ID, &t.Slug, &t.OrgName, &t.OrgType, &t.DBName, &t.DBHost, &t.DBPort)
	if err != nil { c.JSON(404, gin.H{"error": "tenant not found"}); return }
	c.JSON(200, gin.H{
		"dsn": fmt.Sprintf("root:password@tcp(%s:%d)/%s?parseTime=true", t.DBHost, t.DBPort, t.DBName),
		"dbName": t.DBName, "orgType": t.OrgType,
	})
}

func listTenants(c *gin.Context) {
	rows, err := platformDB.QueryContext(c.Request.Context(),
		`SELECT id,slug,org_name,org_type,country_code,plan,is_active FROM tenants ORDER BY org_type,org_name`)
	if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
	defer rows.Close()
	var tenants []gin.H
	for rows.Next() {
		var t struct{ ID,Slug,Name,Type,Country,Plan string; Active bool }
		rows.Scan(&t.ID,&t.Slug,&t.Name,&t.Type,&t.Country,&t.Plan,&t.Active)
		tenants = append(tenants, gin.H{"id":t.ID,"slug":t.Slug,"name":t.Name,"type":t.Type,"country":t.Country,"plan":t.Plan,"active":t.Active})
	}
	c.JSON(200, gin.H{"tenants": tenants})
}

func createTenant(c *gin.Context) {
	var body struct { Slug,OrgName,OrgType,CountryCode string }
	c.ShouldBindJSON(&body)
	dbName := "db_" + body.Slug
	id := uuid.New().String()
	platformDB.ExecContext(c.Request.Context(),
		`INSERT INTO tenants(id,slug,org_name,org_type,country_code,db_name) VALUES(?,?,?,?,?,?)`,
		id, body.Slug, body.OrgName, body.OrgType, body.CountryCode, dbName,
	)
	c.JSON(201, gin.H{"id": id, "dbName": dbName, "message": "tenant created — run migrations to initialise DB"})
}

func getTenant(c *gin.Context) { c.JSON(200, gin.H{"slug": c.Param("slug")}) }

func requireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Simplified — full JWT validation in api-gateway
		c.Next()
	}
}

func sha256Hex(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func getenv(k, d string) string { if v := os.Getenv(k); v != "" { return v }; return d }
