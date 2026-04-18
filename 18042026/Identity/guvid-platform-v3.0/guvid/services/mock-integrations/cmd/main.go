package main

import (
	"context"; "log/slog"; "net/http"; "os"; "os/signal"; "syscall"; "time"
	"math/rand"; "fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode("release"); r := gin.New(); r.Use(gin.Recovery())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service":"mock-integrations","status":"ok"})
	})
	// India
	r.POST("/in/uidai/otp", func(c *gin.Context) {
		c.JSON(200, gin.H{"transaction_id": fmt.Sprintf("TXN-%d", rand.Intn(1000000)), "status": "OTP_SENT"})
	})
	r.POST("/in/uidai/verify-otp", func(c *gin.Context) {
		var req struct{ OTP string `json:"otp"` }; c.ShouldBindJSON(&req)
		if req.OTP == "123456" || req.OTP == "000000" {
			c.JSON(200, gin.H{"name":"Rahul Kumar Sharma","dob":"1990-05-15","gender":"M","status":"SUCCESS"})
		} else { c.JSON(400, gin.H{"error":"invalid OTP","code":"OTP_MISMATCH"}) }
	})
	r.POST("/in/nsdl/pan-verify", func(c *gin.Context) {
		c.JSON(200, gin.H{"status":"VALID","name_match":true,"pan_holder":"RAHUL KUMAR SHARMA"})
	})
	r.GET("/in/nad/lookup", func(c *gin.Context) {
		c.JSON(200, gin.H{"institution_name":"IIT Delhi","degree":"Bachelor","field_of_study":"Computer Science","graduation_year":2015,"issuing_body":"AICTE"})
	})
	r.GET("/in/epfo/passbook", func(c *gin.Context) {
		c.JSON(200, gin.H{"uan":"100123456789","employer_name":"Infosys Technologies Limited","start_date":"2022-07-01","end_date":nil})
	})
	r.GET("/in/mca/company-lookup", func(c *gin.Context) {
		c.JSON(200, gin.H{"company_name":"Infosys Technologies Limited","cin":"L85110KA1981PLC013115","status":"Active"})
	})
	// US
	r.POST("/us/irs/tin-verify", func(c *gin.Context) {
		c.JSON(200, gin.H{"transaction_id":fmt.Sprintf("IRS-TXN-%d",rand.Intn(1000000)),"status":"ACCEPTED"})
	})
	r.GET("/us/nsc/enrollment-verify", func(c *gin.Context) {
		c.JSON(200, gin.H{"institution_name":"MIT","degree":"Bachelor","field":"Computer Science","year":2020})
	})
	r.GET("/us/ssa/earnings", func(c *gin.Context) {
		c.JSON(200, gin.H{"employer_name":"Google LLC","employment_type":"FT","start_year":2020,"tax_verified":true})
	})
	// GB
	r.POST("/gb/govuk/token", func(c *gin.Context) {
		c.JSON(200, gin.H{"access_token":"mock_gb_token_"+fmt.Sprintf("%d",rand.Intn(9999)),"token_type":"Bearer","expires_in":3600})
	})
	r.GET("/gb/hesa/student-verify", func(c *gin.Context) {
		c.JSON(200, gin.H{"institution":"University of Oxford","degree":"Bachelor","year":2019})
	})

	srv := &http.Server{Addr:":8099",Handler:r}
	slog.New(slog.NewJSONHandler(os.Stdout,nil)).Info("mock-integrations starting","port",8099)
	go func() { srv.ListenAndServe() }()
	quit := make(chan os.Signal,1); signal.Notify(quit,syscall.SIGINT,syscall.SIGTERM); <-quit
	ctx,cancel := context.WithTimeout(context.Background(),5*time.Second); defer cancel(); srv.Shutdown(ctx)
}
