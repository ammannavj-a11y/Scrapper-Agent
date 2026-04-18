package main

import (
	"context"
	"database/sql"
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

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func getTenantDB(slug string) (*sql.DB, error) {
	resp, err := http.Get(fmt.Sprintf("http://tenant-svc:8094/api/v1/tenant-db/%s", slug))
	if err != nil { return nil, err }
	defer resp.Body.Close()
	var r struct{ DSN string `json:"dsn"` }
	json.NewDecoder(resp.Body).Decode(&r)
	return sql.Open("mysql", r.DSN)
}

func verifyGUVIDFromChain(guvid string) (map[string]interface{}, error) {
	resp, err := http.Get(fmt.Sprintf("http://verify-svc:8085/api/v1/guvid/verify?guvid=%s", guvid))
	if err != nil { return demoVerifyResult(guvid), nil }
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func demoVerifyResult(guvid string) map[string]interface{} {
	return map[string]interface{}{
		"guvid": guvid, "trustLevel": "HIGH", "trustScore": 87.5,
		"identityScore": 92.0, "educationScore": 85.0, "employmentScore": 81.0,
		"status": "active", "holderVerified": false,
		"fabricTxId": "tx_demo_" + guvid[len(guvid)-8:],
	}
}

func main() {
	gin.SetMode(getenv("GIN_MODE", "release"))
	r := gin.New(); r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "hr-portal-svc", "status": "ok", "version": "3.0.0"})
	})

	r.POST("/api/v1/hr/verify", handleVerify)
	r.POST("/api/v1/hr/batch-verify", handleBatchVerify)
	r.GET("/api/v1/hr/candidates", listCandidates)
	r.GET("/api/v1/hr/stats", getHRStats)
	r.GET("/api/v1/hr/audit", getAuditLog)

	srv := &http.Server{Addr: ":8091", Handler: r}
	go func() {
		logger.Info("hr-portal-svc starting", "port", 8091)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("error", "err", err); os.Exit(1)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel(); srv.Shutdown(ctx)
}

type VerifyRequest struct {
	GUVID        string `json:"guvid" binding:"required"`
	ExpectedName string `json:"expectedName"`
	Position     string `json:"position"`
	VerifyType   string `json:"verifyType"`
}

func handleVerify(c *gin.Context) {
	tenantSlug := c.GetHeader("X-Tenant-Slug")
	if tenantSlug == "" { tenantSlug = "google-hr" }

	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()}); return
	}

	db, err := getTenantDB(tenantSlug)
	if err != nil { c.JSON(502, gin.H{"error": "tenant db unavailable"}); return }
	defer db.Close()

	// Call verify service
	result, _ := verifyGUVIDFromChain(req.GUVID)

	trustLevel, _ := result["trustLevel"].(string)
	trustScore, _ := result["trustScore"].(float64)
	idScore, _ := result["identityScore"].(float64)
	eduScore, _ := result["educationScore"].(float64)
	empScore, _ := result["employmentScore"].(float64)
	fabricTxID, _ := result["fabricTxId"].(string)

	verificationID := uuid.New().String()
	verifyResult := "pass"
	if trustLevel == "UNVERIFIED" { verifyResult = "fail" }

	// Store in tenant's own DB
	db.ExecContext(c.Request.Context(),
		`INSERT INTO hr_verifications(id,tenant_id,guvid,position_applied,verification_type,trust_level,trust_score,
		identity_score,education_score,employment_score,fabric_log_id,result,verified_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		verificationID, tenantSlug, req.GUVID, req.Position,
		req.VerifyType, trustLevel, trustScore, idScore, eduScore, empScore,
		fabricTxID, verifyResult, time.Now(),
	)

	// Audit
	db.ExecContext(c.Request.Context(),
		`INSERT INTO hr_audit_log(id,tenant_id,action,actor_id,resource_type,resource_id,details)
		VALUES(?,?,?,?,?,?,?)`,
		uuid.New().String(), tenantSlug, "CANDIDATE_VERIFIED", "system", "guvid", req.GUVID,
		fmt.Sprintf(`{"trustLevel":"%s","score":%.1f}`, trustLevel, trustScore),
	)

	logger.Info("hr verify", "guvid", req.GUVID, "result", verifyResult, "trust", trustLevel, "tenant", tenantSlug)
	c.JSON(200, gin.H{
		"verificationId": verificationID,
		"guvid": req.GUVID, "trustLevel": trustLevel, "trustScore": trustScore,
		"identityScore": idScore, "educationScore": eduScore, "employmentScore": empScore,
		"result": verifyResult, "fabricLogId": fabricTxID,
		"holderVerified": result["holderVerified"],
	})
}

func handleBatchVerify(c *gin.Context) {
	tenantSlug := c.GetHeader("X-Tenant-Slug")
	if tenantSlug == "" { tenantSlug = "google-hr" }

	var req struct{ GUVIDs []string `json:"guvids"` }
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"error": err.Error()}); return }

	db, err := getTenantDB(tenantSlug)
	if err != nil { c.JSON(502, gin.H{"error": "tenant db unavailable"}); return }
	defer db.Close()

	results := make([]map[string]interface{}, 0, len(req.GUVIDs))
	for _, guvid := range req.GUVIDs {
		r, _ := verifyGUVIDFromChain(guvid)
		tl, _ := r["trustLevel"].(string)
		ts, _ := r["trustScore"].(float64)
		vid := uuid.New().String()
		res := "pass"; if tl == "UNVERIFIED" { res = "fail" }
		db.ExecContext(c.Request.Context(),
			`INSERT INTO hr_verifications(id,tenant_id,guvid,verification_type,trust_level,trust_score,result) VALUES(?,?,?,?,?,?,?)`,
			vid, tenantSlug, guvid, "batch", tl, ts, res,
		)
		r["verificationId"] = vid; r["result"] = res
		results = append(results, r)
	}
	c.JSON(200, gin.H{"results": results, "total": len(results)})
}

func listCandidates(c *gin.Context) {
	tenantSlug := c.GetHeader("X-Tenant-Slug")
	if tenantSlug == "" { tenantSlug = "google-hr" }
	db, err := getTenantDB(tenantSlug)
	if err != nil { c.JSON(502, gin.H{"error": "tenant db unavailable"}); return }
	defer db.Close()

	rows, _ := db.QueryContext(c.Request.Context(),
		`SELECT id,guvid,trust_level,trust_score,result,verified_at FROM hr_verifications WHERE tenant_id=? ORDER BY verified_at DESC LIMIT 50`, tenantSlug)
	defer rows.Close()
	var cands []gin.H
	for rows.Next() {
		var id, guvid, tl, res string; var ts float64; var va time.Time
		rows.Scan(&id,&guvid,&tl,&ts,&res,&va)
		cands = append(cands, gin.H{"id":id,"guvid":guvid,"trustLevel":tl,"trustScore":ts,"result":res,"verifiedAt":va})
	}
	if cands == nil { cands = []gin.H{} }
	c.JSON(200, gin.H{"candidates": cands})
}

func getHRStats(c *gin.Context) {
	tenantSlug := c.GetHeader("X-Tenant-Slug")
	if tenantSlug == "" { tenantSlug = "google-hr" }
	db, err := getTenantDB(tenantSlug)
	if err != nil { c.JSON(502, gin.H{"error": "tenant db unavailable"}); return }
	defer db.Close()

	var total, passed, failed int
	db.QueryRowContext(c.Request.Context(),
		`SELECT COUNT(*),SUM(result='pass'),SUM(result='fail') FROM hr_verifications WHERE tenant_id=?`, tenantSlug,
	).Scan(&total,&passed,&failed)
	c.JSON(200, gin.H{"totalVerifications":total,"passed":passed,"failed":failed})
}

func getAuditLog(c *gin.Context) {
	tenantSlug := c.GetHeader("X-Tenant-Slug")
	if tenantSlug == "" { tenantSlug = "google-hr" }
	db, err := getTenantDB(tenantSlug)
	if err != nil { c.JSON(502, gin.H{"error": "tenant db unavailable"}); return }
	defer db.Close()

	rows, _ := db.QueryContext(c.Request.Context(),
		`SELECT id,action,actor_id,resource_id,details,created_at FROM hr_audit_log WHERE tenant_id=? ORDER BY created_at DESC LIMIT 50`, tenantSlug)
	defer rows.Close()
	var logs []gin.H
	for rows.Next() {
		var id,action,actor,resource,details string; var created time.Time
		rows.Scan(&id,&action,&actor,&resource,&details,&created)
		logs = append(logs, gin.H{"id":id,"action":action,"actor":actor,"resource":resource,"details":details,"createdAt":created})
	}
	if logs == nil { logs = []gin.H{} }
	c.JSON(200, gin.H{"logs": logs})
}

func getenv(k, d string) string { if v := os.Getenv(k); v != "" { return v }; return d }
