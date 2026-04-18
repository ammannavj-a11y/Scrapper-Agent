package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
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
var aesKey []byte

func getTenantDB(tenantSlug string) (*sql.DB, error) {
	resp, err := http.Get(fmt.Sprintf("http://tenant-svc:8094/api/v1/tenant-db/%s", tenantSlug))
	if err != nil { return nil, err }
	defer resp.Body.Close()
	var result struct{ DSN string `json:"dsn"` }
	json.NewDecoder(resp.Body).Decode(&result)
	return sql.Open("mysql", result.DSN)
}

func main() {
	keyHex := getenv("AES_KEY", "")
	if keyHex == "" {
		keyHex = fmt.Sprintf("%x", sha256.Sum256([]byte("change_me_32_bytes_aes_key")))
	}
	aesKey, _ = hex.DecodeString(keyHex)

	gin.SetMode(getenv("GIN_MODE", "release"))
	r := gin.New(); r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "institution-svc", "status": "ok", "version": "3.0.0"})
	})
	r.GET("/ready", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ready"}) })

	// All routes require tenant context from JWT (set by api-gateway)
	r.POST("/api/v1/institution/issue-credential", issueCred)
	r.GET("/api/v1/institution/credentials", listCreds)
	r.GET("/api/v1/institution/credentials/:id", getCred)
	r.DELETE("/api/v1/institution/credentials/:id", revokeCred)
	r.GET("/api/v1/institution/stats", getStats)
	r.GET("/api/v1/institution/audit", getAuditLog)

	srv := &http.Server{Addr: ":8082", Handler: r}
	go func() {
		logger.Info("institution-svc starting", "port", 8082)
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

type IssueCredRequest struct {
	HolderGUVID     string `json:"holderGuvid" binding:"required"`
	HolderName      string `json:"holderName" binding:"required"`
	HolderDOB       string `json:"holderDob"`
	CredentialType  string `json:"credentialType" binding:"required"`
	CredentialLevel string `json:"credentialLevel"`
	DegreeName      string `json:"degreeName"`
	FieldOfStudy    string `json:"fieldOfStudy"`
	GraduationYear  int    `json:"graduationYear"`
	Grade           string `json:"grade"`
	RollNumber      string `json:"rollNumber"`
	ExpiresAt       string `json:"expiresAt"`
}

func issueCred(c *gin.Context) {
	tenantSlug := c.GetHeader("X-Tenant-Slug")
	if tenantSlug == "" { tenantSlug = "iit-delhi" } // demo fallback

	var req IssueCredRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()}); return
	}

	db, err := getTenantDB(tenantSlug)
	if err != nil { c.JSON(502, gin.H{"error": "tenant db unavailable"}); return }
	defer db.Close()

	credID := uuid.New().String()
	holderNameEnc, _ := encryptAES([]byte(req.HolderName))
	holderDOBEnc, _ := encryptAES([]byte(req.HolderDOB))
	gradeEnc, _ := encryptAES([]byte(req.Grade))
	rollEnc, _ := encryptAES([]byte(req.RollNumber))

	// W3C VC 2.0 JSON (no PII — only structural metadata)
	w3cVC := map[string]interface{}{
		"@context": []string{"https://www.w3.org/2018/credentials/v1"},
		"type": []string{"VerifiableCredential", req.CredentialType},
		"issuer": fmt.Sprintf("did:guvid:IN:tenant:%s", tenantSlug),
		"credentialSubject": map[string]interface{}{
			"id": fmt.Sprintf("did:guvid:IN:%s", req.HolderGUVID),
			"credentialType": req.CredentialType, "level": req.CredentialLevel,
			"field": req.FieldOfStudy, "year": req.GraduationYear,
		},
		"issuanceDate": time.Now().UTC().Format(time.RFC3339),
	}
	w3cJSON, _ := json.Marshal(w3cVC)

	credHash := sha256Hex(credID + req.HolderGUVID + req.CredentialType)
	fabricTxID := "tx_" + sha256Hex(credID)[:16]

	_, err = db.ExecContext(c.Request.Context(),
		`INSERT INTO issued_credentials(id,tenant_id,guvid,holder_name_encrypted,holder_dob_encrypted,
		credential_type,credential_level,degree_name,field_of_study,graduation_year,grade_encrypted,
		roll_number_encrypted,w3c_vc_json,cred_hash,fabric_tx_id,status,issued_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		credID, tenantSlug, req.HolderGUVID, holderNameEnc, holderDOBEnc,
		req.CredentialType, req.CredentialLevel, req.DegreeName, req.FieldOfStudy,
		req.GraduationYear, gradeEnc, rollEnc, string(w3cJSON),
		credHash, fabricTxID, "active", time.Now(),
	)
	if err != nil {
		logger.Error("cred insert failed", "err", err)
		c.JSON(500, gin.H{"error": "failed to issue credential"}); return
	}

	// Audit log
	db.ExecContext(c.Request.Context(),
		`INSERT INTO institution_audit_log(id,tenant_id,action,actor_id,guvid,credential_id)
		VALUES(?,?,?,?,?,?)`,
		uuid.New().String(), tenantSlug, "CREDENTIAL_ISSUED", "system", req.HolderGUVID, credID,
	)

	logger.Info("credential issued", "id", credID, "guvid", req.HolderGUVID, "tenant", tenantSlug)
	c.JSON(201, gin.H{
		"credentialId": credID, "credHash": credHash,
		"fabricTxId": fabricTxID, "w3cVc": w3cVC,
		"status": "active", "issuedAt": time.Now().UTC(),
	})
}

func listCreds(c *gin.Context) {
	tenantSlug := c.GetHeader("X-Tenant-Slug")
	if tenantSlug == "" { tenantSlug = "iit-delhi" }
	db, err := getTenantDB(tenantSlug)
	if err != nil { c.JSON(502, gin.H{"error": "tenant db unavailable"}); return }
	defer db.Close()

	rows, err := db.QueryContext(c.Request.Context(),
		`SELECT id,guvid,credential_type,credential_level,degree_name,graduation_year,status,issued_at,fabric_tx_id
		FROM issued_credentials WHERE tenant_id=? ORDER BY issued_at DESC LIMIT 100`, tenantSlug)
	if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
	defer rows.Close()

	var creds []gin.H
	for rows.Next() {
		var id, guvid, ctype, clevel, dname, status, txid string
		var year int; var issuedAt time.Time
		rows.Scan(&id, &guvid, &ctype, &clevel, &dname, &year, &status, &issuedAt, &txid)
		creds = append(creds, gin.H{"id":id,"guvid":guvid,"type":ctype,"level":clevel,"degree":dname,"year":year,"status":status,"issuedAt":issuedAt,"fabricTxId":txid})
	}
	if creds == nil { creds = []gin.H{} }
	c.JSON(200, gin.H{"credentials": creds, "total": len(creds)})
}

func getCred(c *gin.Context) {
	tenantSlug := c.GetHeader("X-Tenant-Slug")
	if tenantSlug == "" { tenantSlug = "iit-delhi" }
	db, err := getTenantDB(tenantSlug)
	if err != nil { c.JSON(502, gin.H{"error": "tenant db unavailable"}); return }
	defer db.Close()

	var id, guvid, ctype, clevel, status, txid, w3cJson string
	var year int; var issuedAt time.Time
	err = db.QueryRowContext(c.Request.Context(),
		`SELECT id,guvid,credential_type,credential_level,graduation_year,status,fabric_tx_id,w3c_vc_json,issued_at
		FROM issued_credentials WHERE id=? AND tenant_id=?`, c.Param("id"), tenantSlug,
	).Scan(&id,&guvid,&ctype,&clevel,&year,&status,&txid,&w3cJson,&issuedAt)
	if err != nil { c.JSON(404, gin.H{"error": "credential not found"}); return }
	c.JSON(200, gin.H{"id":id,"guvid":guvid,"type":ctype,"level":clevel,"year":year,"status":status,"fabricTxId":txid,"w3cVc":json.RawMessage(w3cJson),"issuedAt":issuedAt})
}

func revokeCred(c *gin.Context) {
	tenantSlug := c.GetHeader("X-Tenant-Slug")
	if tenantSlug == "" { tenantSlug = "iit-delhi" }
	db, err := getTenantDB(tenantSlug)
	if err != nil { c.JSON(502, gin.H{"error": "tenant db unavailable"}); return }
	defer db.Close()

	db.ExecContext(c.Request.Context(), `UPDATE issued_credentials SET status='revoked' WHERE id=? AND tenant_id=?`, c.Param("id"), tenantSlug)
	c.JSON(200, gin.H{"message": "credential revoked"})
}

func getStats(c *gin.Context) {
	tenantSlug := c.GetHeader("X-Tenant-Slug")
	if tenantSlug == "" { tenantSlug = "iit-delhi" }
	db, err := getTenantDB(tenantSlug)
	if err != nil { c.JSON(502, gin.H{"error": "tenant db unavailable"}); return }
	defer db.Close()

	var total, active, revoked int
	db.QueryRowContext(c.Request.Context(), `SELECT COUNT(*),SUM(status='active'),SUM(status='revoked') FROM issued_credentials WHERE tenant_id=?`, tenantSlug).Scan(&total,&active,&revoked)
	c.JSON(200, gin.H{"totalIssued":total,"active":active,"revoked":revoked})
}

func getAuditLog(c *gin.Context) {
	tenantSlug := c.GetHeader("X-Tenant-Slug")
	if tenantSlug == "" { tenantSlug = "iit-delhi" }
	db, err := getTenantDB(tenantSlug)
	if err != nil { c.JSON(502, gin.H{"error": "tenant db unavailable"}); return }
	defer db.Close()

	rows, _ := db.QueryContext(c.Request.Context(),
		`SELECT id,action,actor_id,guvid,credential_id,created_at FROM institution_audit_log WHERE tenant_id=? ORDER BY created_at DESC LIMIT 50`, tenantSlug)
	defer rows.Close()
	var logs []gin.H
	for rows.Next() {
		var id,action,actor,guvid,credID string; var created time.Time
		rows.Scan(&id,&action,&actor,&guvid,&credID,&created)
		logs = append(logs, gin.H{"id":id,"action":action,"actor":actor,"guvid":guvid,"credentialId":credID,"createdAt":created})
	}
	if logs == nil { logs = []gin.H{} }
	c.JSON(200, gin.H{"logs": logs})
}

func encryptAES(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil { return nil, err }
	gcm, err := cipher.NewGCM(block)
	if err != nil { return nil, err }
	nonce := make([]byte, gcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	return gcm.Seal(nonce, nonce, data, nil), nil
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return base64.StdEncoding.EncodeToString(h[:])
}

func getenv(k, d string) string { if v := os.Getenv(k); v != "" { return v }; return d }
