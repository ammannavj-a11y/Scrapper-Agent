package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var log = slog.New(slog.NewJSONHandler(os.Stdout, nil))

// Backend routing table
var backends = map[string]string{
	"/api/v1/auth/":            "http://tenant-svc:8094",
	"/api/v1/tenant":           "http://tenant-svc:8094",
	"/api/v1/identity/":        "http://identity-svc:8081",
	"/api/v1/education/":       "http://education-svc:8082",
	"/api/v1/employment/":      "http://employment-svc:8083",
	"/api/v1/guvid/issue":      "http://guvid-svc:8084",
	"/api/v1/guvid/verify":     "http://verify-svc:8085",
	"/api/v1/guvid/present":    "http://verify-svc:8085",
	"/api/v1/chain/":           "http://verify-svc:8085",
	"/api/v1/wallet/":          "http://wallet-svc:8086",
	"/api/v1/consent/":         "http://consent-svc:8087",
	"/api/v1/fraud/":           "http://fraud-svc:8088",
	"/api/v1/hr/":              "http://hr-portal-svc:8091",
	"/api/v1/institution/":     "http://institution-svc:8082",
	"/api/v1/regulatory/":      "http://regulatory-svc:8093",
	"/api/v1/adapter-registry/": "http://adapter-registry-svc:8092",
}

// In-memory rate limiter (per IP, per minute)
var rateMap = map[string][]time.Time{}

func main() {
	gin.SetMode(getenv("GIN_MODE", "release"))
	r := gin.New()
	r.Use(gin.Recovery(), requestLogger(), corsMiddleware())

	// Health / readiness
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "api-gateway", "status": "ok",
			"country": getenv("COUNTRY_CODE", "IN"), "version": "3.0.0",
		})
	})
	r.GET("/ready", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ready"}) })
	r.GET("/metrics", func(c *gin.Context) { c.String(200, "# api-gateway\napi_gateway_up 1\n") })

	// ── Public routes (no JWT required) ─────────────────────────────────────
	r.POST("/api/v1/auth/login", rateLimiter(20), proxyTo("http://tenant-svc:8094"))
	r.POST("/api/v1/auth/refresh", proxyTo("http://tenant-svc:8094"))
	r.POST("/api/v1/auth/logout", proxyTo("http://tenant-svc:8094"))
	r.GET("/api/v1/guvid/verify", rateLimiter(100), proxyTo("http://verify-svc:8085"))
	r.GET("/api/v1/guvid/verify-quick", rateLimiter(200), proxyTo("http://verify-svc:8085"))
	r.POST("/api/v1/guvid/present/challenge", rateLimiter(100), proxyTo("http://verify-svc:8085"))
	r.POST("/api/v1/guvid/present", rateLimiter(100), proxyTo("http://verify-svc:8085"))
	r.GET("/api/v1/chain/status", proxyTo("http://verify-svc:8085"))
	r.GET("/api/v1/chain/history/:guvid", proxyTo("http://verify-svc:8085"))
	r.GET("/api/v1/chain/stats", proxyTo("http://verify-svc:8085"))
	r.GET("/api/v1/did/resolve/*did", proxyTo("http://verify-svc:8085"))

	// ── Authenticated routes (JWT required) ──────────────────────────────────
	auth := r.Group("/api/v1")
	auth.Use(jwtMiddleware())

	// Identity
	auth.POST("/identity/challenge", rateLimiter(10), proxyTo("http://identity-svc:8081"))
	auth.POST("/identity/verify-challenge", rateLimiter(10), proxyTo("http://identity-svc:8081"))
	auth.POST("/identity/verify-secondary", rateLimiter(10), proxyTo("http://identity-svc:8081"))

	// Education / Employment
	auth.POST("/education/verify", proxyTo("http://identity-svc:8081"))
	auth.POST("/employment/verify", proxyTo("http://identity-svc:8081"))

	// GUVID issuance (rate-limited heavily)
	auth.POST("/guvid/issue", rateLimiter(5), proxyTo("http://guvid-svc:8084"))

	// Wallet / WebAuthn
	auth.POST("/wallet/register/begin", proxyTo("http://wallet-svc:8086"))
	auth.POST("/wallet/register/complete", proxyTo("http://wallet-svc:8086"))
	auth.POST("/wallet/auth/begin", proxyTo("http://wallet-svc:8086"))
	auth.POST("/wallet/auth/complete", proxyTo("http://wallet-svc:8086"))
	auth.GET("/wallet/status", proxyTo("http://wallet-svc:8086"))
	auth.POST("/wallet/recovery/initiate", proxyTo("http://wallet-svc:8086"))
	auth.POST("/wallet/recovery/guardian-approve", proxyTo("http://wallet-svc:8086"))

	// Consent
	auth.PUT("/consent/mode", proxyTo("http://consent-svc:8087"))
	auth.GET("/consent/history", proxyTo("http://consent-svc:8087"))
	auth.POST("/consent/approve/:id", proxyTo("http://consent-svc:8087"))
	auth.POST("/consent/deny/:id", proxyTo("http://consent-svc:8087"))
	auth.POST("/consent/report-fraud/:id", proxyTo("http://consent-svc:8087"))

	// HR Portal
	auth.POST("/hr/verify", proxyTo("http://hr-portal-svc:8091"))
	auth.POST("/hr/batch-verify", proxyTo("http://hr-portal-svc:8091"))
	auth.GET("/hr/candidates", proxyTo("http://hr-portal-svc:8091"))
	auth.GET("/hr/stats", proxyTo("http://hr-portal-svc:8091"))
	auth.GET("/hr/audit", proxyTo("http://hr-portal-svc:8091"))

	// Institution
	auth.POST("/institution/issue-credential", proxyTo("http://institution-svc:8082"))
	auth.GET("/institution/credentials", proxyTo("http://institution-svc:8082"))
	auth.GET("/institution/credentials/:id", proxyTo("http://institution-svc:8082"))
	auth.DELETE("/institution/credentials/:id", proxyTo("http://institution-svc:8082"))
	auth.GET("/institution/stats", proxyTo("http://institution-svc:8082"))
	auth.GET("/institution/audit", proxyTo("http://institution-svc:8082"))

	// Regulatory
	auth.GET("/regulatory/overview", proxyTo("http://regulatory-svc:8093"))
	auth.GET("/regulatory/institutions", proxyTo("http://regulatory-svc:8093"))
	auth.GET("/regulatory/hr-orgs", proxyTo("http://regulatory-svc:8093"))
	auth.GET("/regulatory/compliance", proxyTo("http://regulatory-svc:8093"))
	auth.GET("/regulatory/audit-trail", proxyTo("http://regulatory-svc:8093"))
	auth.GET("/regulatory/fraud-summary", proxyTo("http://regulatory-svc:8093"))
	auth.POST("/regulatory/reports", proxyTo("http://regulatory-svc:8093"))
	auth.GET("/regulatory/reports", proxyTo("http://regulatory-svc:8093"))

	// Fraud L1
	auth.GET("/fraud/incidents", proxyTo("http://regulatory-svc:8093"))
	auth.PUT("/fraud/incidents/:id", proxyTo("http://regulatory-svc:8093"))
	auth.GET("/fraud/graph", proxyTo("http://regulatory-svc:8093"))
	auth.GET("/fraud/stats", proxyTo("http://regulatory-svc:8093"))

	// Tenant admin
	auth.GET("/tenant-db/:slug", proxyTo("http://tenant-svc:8094"))
	auth.GET("/tenants", proxyTo("http://tenant-svc:8094"))
	auth.POST("/tenants", proxyTo("http://tenant-svc:8094"))

	// Kafka / analytics
	auth.GET("/kafka/lag", proxyTo("http://regulatory-svc:8093"))

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	go func() {
		log.Info("api-gateway starting", "port", 8080)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("api-gateway shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

// ── Middleware ─────────────────────────────────────────────────────────────

func jwtMiddleware() gin.HandlerFunc {
	secret := getenv("JWT_SECRET", "change_me_jwt_secret_32_chars")
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing authorization header", "code": "UNAUTHORIZED"})
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid or expired token", "code": "TOKEN_INVALID"})
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("tenantId", claims["tenantId"])
			c.Set("tenantSlug", claims["tenantSlug"])
			c.Set("userId", claims["sub"])
			c.Set("role", claims["role"])
			c.Set("orgType", claims["orgType"])
			c.Set("countryCode", claims["countryCode"])
			c.Set("dbName", claims["dbName"])
		}
		c.Next()
	}
}

func rateLimiter(rpm int) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		now := time.Now()
		window := now.Add(-time.Minute)
		var valid []time.Time
		for _, t := range rateMap[key] {
			if t.After(window) {
				valid = append(valid, t)
			}
		}
		valid = append(valid, now)
		rateMap[key] = valid
		if len(valid) > rpm {
			c.AbortWithStatusJSON(429, gin.H{"error": "rate limit exceeded", "code": "RATE_LIMITED"})
			return
		}
		c.Next()
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,X-API-Key,X-Tenant-Slug")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"ms", time.Since(start).Milliseconds(),
			"ip", c.ClientIP(),
		)
	}
}

// ── Proxy ──────────────────────────────────────────────────────────────────

func proxyTo(target string) gin.HandlerFunc {
	remote, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Error("proxy error", "target", target, "err", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(502)
		w.Write([]byte(`{"error":"upstream service unavailable","code":"BAD_GATEWAY"}`))
	}
	return func(c *gin.Context) {
		// Forward tenant slug from JWT into header for downstream services
		if slug, ok := c.Get("tenantSlug"); ok {
			c.Request.Header.Set("X-Tenant-Slug", fmt.Sprintf("%v", slug))
		}
		if dbName, ok := c.Get("dbName"); ok {
			c.Request.Header.Set("X-DB-Name", fmt.Sprintf("%v", dbName))
		}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
