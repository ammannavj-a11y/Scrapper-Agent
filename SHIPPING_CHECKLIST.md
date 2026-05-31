# PrivacyShield — Pre-Launch Shipping Checklist

**Status:** MVP Ready (v1.0.0-rc1)  
**Target GA Date:** June 8, 2026  
**Last Updated:** 2026-05-23

---

## CRITICAL PATH BLOCKERS (Must Complete by May 31)

### Authentication & Account Management
- [ ] **FR-3: Password Reset Flow**
  - Backend: POST /auth/forgot-password (generate 24h token)
  - Backend: POST /auth/reset-password (verify token + update pwd)
  - Frontend: ForgotPassword.tsx component
  - Email template: password_reset.html
  - Test: auth reset flow E2E
  - **Owner:** Backend + Frontend
  - **Deadline:** May 27
  - **Est. Effort:** 4 hours

- [ ] **FR-2: JTI Blocklist for Token Revocation (P1 Security)**
  - Backend: Add JTI blocklist to Redis on logout
  - Backend: Check JTI in token_decode() → raise TokenRevokedError
  - Test: logout invalidates token immediately
  - **Owner:** Backend
  - **Deadline:** May 26
  - **Est. Effort:** 2 hours
  - **Security Impact:** CRITICAL

### Billing & Payment
- [ ] **FR-19: Payment Processing Integration**
  - [ ] Stripe account setup (or Kill-Bill if using OSS)
  - [ ] Create Stripe webhook endpoint: POST /webhooks/stripe
  - [ ] Handle: `payment_intent.succeeded`, `payment_intent.payment_failed`, `customer.subscription.deleted`
  - [ ] Backend: Create Subscription model + update User.tier on success
  - [ ] Frontend: CheckoutPage with Stripe Elements
  - [ ] Email: Invoice template (receipt) → send via Mailhog
  - [ ] Test: End-to-end payment flow (test Stripe API key)
  - [ ] **Owner:** Backend + Frontend
  - **Deadline:** May 31
  - **Est. Effort:** 16 hours
  - **Revenue Impact:** CRITICAL

- [ ] **FR-20: Usage Quota Enforcement**
  - [ ] API: Return X-RateLimit-* headers on every response
  - [ ] Error response when quota exceeded (402 Payment Required)
  - [ ] Frontend: Warn at 80% quota, disable at 100%
  - [ ] Test: Validate quota for each tier
  - [ ] **Owner:** Backend + Frontend
  - **Deadline:** May 28
  - **Est. Effort:** 3 hours

### Data Privacy & Compliance
- [ ] **FR-25/26: GDPR/DPDP Erasure Pipeline (P0 Legal)**
  - [ ] Backend: Implement user_service.erase_user_data(user_id)
    - Hard delete: Scans, RemovalRequests, APIKeys
    - Cascade delete: Refresh tokens
    - Purge Redis: user:* keys
    - Anonymize: Audit logs (replace email with hash)
  - [ ] Migration: Add `deleted_at` trigger to mark users for erasure
  - [ ] Email: Confirm account deletion email
  - [ ] Test: Verify all user data purged post-deletion
  - [ ] **Owner:** Backend
  - **Deadline:** May 29
  - **Est. Effort:** 8 hours
  - **Legal Impact:** CRITICAL

- [ ] **FR-28: Legal Documents**
  - [ ] Privacy Policy (templates: Termly, iubenda)
    - Data handling, DPDP Act compliance, rights explanation
  - [ ] Terms of Service
    - Limitation of liability, usage restrictions, dispute resolution
  - [ ] Data Processing Agreement (DPA) — for enterprise customers
  - [ ] Legal review (3-5 days)
  - [ ] Host on website (static pages)
  - [ ] Frontend: Link in footer + onboarding
  - [ ] **Owner:** Legal (external counsel)
  - **Deadline:** May 25
  - **Est. Effort:** 2-3 days (legal review)
  - **Legal Impact:** CRITICAL

### Core Feature Completion
- [ ] **FR-12: Data Broker Removal Rules (15 → 100)**
  - [ ] Define removal rules for top 100 brokers (ranked by frequency)
    - Format: JSON with form selectors, field mapping, submit strategy
    - Example: `{ "broker": "spokeo", "form_selector": "#removal-form", "fields": { "name": "input[name='name']" }, "submit": "button[type='submit']" }`
  - [ ] Priority: US (Spokeo, Whitepages, BeenVerified) + India (JustDial, IndiaStack)
  - [ ] Test: Playwright dry-run on each broker (no real submission)
  - [ ] **Owner:** Data Eng
  - **Deadline:** June 5 (can extend to Sprint 1)
  - **Est. Effort:** 20-30 hours (research + rule definition)
  - **Product Impact:** HIGH
  - **Note:** MVP can ship with 15 brokers if needed; backlog expansion to 500

### Testing & QA
- [ ] **Load Testing @ 500 Concurrent Users**
  - [ ] Setup: JMeter or Locust test plan
  - [ ] Scenario: Register → Login → Initiate Scan → Get Results (5 min duration)
  - [ ] Ramp: 0 → 500 CCU over 10 minutes
  - [ ] Metrics:
    - [ ] p50 latency < 500ms
    - [ ] p95 latency < 1s
    - [ ] p99 latency < 2s
    - [ ] Error rate < 1%
    - [ ] Peak throughput ≥ 100 req/s
  - [ ] Document: Bottlenecks + scaling recommendations
  - [ ] **Owner:** QA
  - **Deadline:** May 30
  - **Est. Effort:** 8 hours

- [ ] **Unit Test Coverage ≥ 90%**
  - [ ] Backend: `pytest tests/ --cov=app --cov-report=html`
    - Target: 90%+ coverage on `app/services/` and `app/api/v1/`
  - [ ] Frontend: `npm run test` → coverage report
  - [ ] GitHub: Block merge if coverage < 85%
  - [ ] **Owner:** All contributors
  - **Deadline:** May 30
  - **Est. Effort:** Ongoing (as features are built)

- [ ] **Security Scan — CI/CD Integration**
  - [ ] Bandit: `bandit -r app/ --skip B101,B601 --exit-code 1`
  - [ ] Safety: `safety check -r requirements.txt --exit-code 1`
  - [ ] pip-audit: `pip-audit --desc -r requirements.txt`
  - [ ] Trivy: `trivy image privacyshield:latest --severity CRITICAL,HIGH --exit-code 1`
  - [ ] GitHub Actions: Fail build if any HIGH/CRITICAL found
  - [ ] **Owner:** DevOps + Security
  - **Deadline:** May 26
  - **Est. Effort:** 4 hours

---

## HIGH-PRIORITY SECURITY FIXES (P1 — Before GA)

### Backend Security Improvements
- [ ] **FR-7: robots.txt Compliance (P1 Ethical)**
  - [ ] Add `urllib.robotparser.RobotFileParser` check before fetching
  - [ ] Implement: `can_fetch(broker_url) → bool`
  - [ ] Update crawler to skip URLs blocked by robots.txt
  - [ ] Test: Mock robots.txt with blocked paths
  - [ ] **Owner:** Backend
  - **Deadline:** June 5
  - **Est. Effort:** 2 hours
  - **Ethical Impact:** HIGH (DMCA compliance)

- [ ] **FR-7: DNS Rebinding SSRF Protection (P1 Security)**
  - [ ] Add host re-resolution after DNS lookup (before HTTP request)
  - [ ] Block if resolved IP is in private range
  - [ ] Add to blocked patterns: 169.254.0.0/16 (AWS metadata)
  - [ ] Test: Mock DNS rebinding attack scenario
  - [ ] **Owner:** Backend
  - **Deadline:** June 5
  - **Est. Effort:** 3 hours
  - **Security Impact:** CRITICAL

- [ ] **Backend: Playwright Input Sanitisation (P2 Security)**
  - [ ] Strip dangerous characters from user_name, user_email before form-fill
  - [ ] Regex: `re.sub(r"[<>\"'`;;&|\\]", "", value)`
  - [ ] Test: Verify no injection into Playwright commands
  - [ ] **Owner:** Backend
  - **Deadline:** June 5
  - **Est. Effort:** 1 hour

### Frontend Security
- [ ] **Frontend: Content Security Policy (CSP) Headers**
  - [ ] Set CSP: `default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com`
  - [ ] Test: Verify CSP headers in response (Nginx config)
  - [ ] **Owner:** DevOps
  - **Deadline:** May 28
  - **Est. Effort:** 1 hour

### Infrastructure Security
- [ ] **AWS KMS Encryption for RDS at Rest**
  - [ ] Enable AWS KMS key for EKS etcd encryption
  - [ ] Update Terraform: `encryption_config.provider.keyArn`
  - [ ] Test: Verify RDS backups are encrypted
  - [ ] **Owner:** DevOps
  - **Deadline:** June 3
  - **Est. Effort:** 2 hours

---

## MEDIUM-PRIORITY FEATURES (P2 — Post-Launch Sprint 1)

### Nice-to-Have Before Launch (Lower Priority)
- [ ] **FR-4: 2FA/TOTP Support**
  - Backend: pyotp library for TOTP generation
  - Endpoint: POST /auth/2fa/enable, POST /auth/2fa/verify
  - QR code generation
  - Backup codes
  - **Deadline:** June 20
  - **Est. Effort:** 8 hours

- [ ] **FR-11: Weekly Re-Scans (Celery Beat)**
  - Scheduler: `@periodic_task(run_every=crontab(hour=1, minute=0, day_of_week='mon'))`
  - Logic: Query all Pro/Enterprise users, trigger run_scan_task
  - Diff detection: Compare new results vs. old
  - Notification: Email diff summary to user
  - **Deadline:** June 20
  - **Est. Effort:** 6 hours

- [ ] **FR-16/17: Browser Extension — Real-Time Alerts**
  - Content script: Detect data broker domains
  - Query backend: POST /extension/analyse
  - Show alert: "Your name found on Spokeo"
  - One-click removal queue
  - **Deadline:** June 20
  - **Est. Effort:** 8 hours

- [ ] **FR-14/15: Manual Removal Fallback + Proof Storage**
  - Email link: Pre-filled removal form on broker
  - Screenshot storage in MinIO (90-day retention)
  - **Deadline:** June 20
  - **Est. Effort:** 6 hours

---

## DEPLOYMENT & INFRASTRUCTURE

### Pre-Production Environment (Staging)
- [ ] **Terraform Infrastructure**
  - [ ] EKS cluster (3 nodes, auto-scale 2-10)
  - [ ] RDS PostgreSQL 16 (Multi-AZ)
  - [ ] ElastiCache Redis 7 (Cluster mode)
  - [ ] S3 buckets (reports, backups)
  - [ ] CloudFront CDN
  - [ ] Route53 DNS failover
  - [ ] IAM roles + SGs
  - [ ] **Owner:** DevOps
  - **Deadline:** May 29
  - **Est. Effort:** 8 hours

- [ ] **Kubernetes Manifests (k8s/)**
  - [ ] Namespace: `privacyshield` with pod security policy
  - [ ] ConfigMap + ExternalSecrets (fetch from AWS Secrets Manager)
  - [ ] Deployment: backend (3 replicas), worker (2 replicas), frontend (2 replicas)
  - [ ] StatefulSet: PostgreSQL, Redis (for local dev)
  - [ ] Ingress: Nginx with SSL termination
  - [ ] HPA: CPU 70%, Memory 80% threshold
  - [ ] NetworkPolicy: Default-deny + explicit allow rules
  - [ ] **Owner:** DevOps
  - **Deadline:** May 29
  - **Est. Effort:** 6 hours

### Monitoring & Observability
- [ ] **Prometheus Alerts**
  - [ ] API p99 latency > 2s → WARN
  - [ ] Error rate > 5% → ALERT
  - [ ] Celery queue depth > 100 → WARN
  - [ ] DB connection pool > 80% → WARN
  - [ ] PVC disk usage > 85% → WARN
  - [ ] **Owner:** DevOps
  - **Deadline:** May 28
  - **Est. Effort:** 3 hours

- [ ] **Grafana Dashboards**
  - [ ] Backend: QPS, latency (p50/p95/p99), errors
  - [ ] Database: Query latency, connections, slow queries
  - [ ] Cache: Hit rate, evictions, memory
  - [ ] Tasks: Queue depth, completion rate, failures
  - [ ] **Owner:** DevOps
  - **Deadline:** May 28
  - **Est. Effort:** 4 hours

- [ ] **PagerDuty / On-Call Integration**
  - [ ] Webhook from Prometheus AlertManager → PagerDuty
  - [ ] Define on-call schedule (3 engineers rotating)
  - [ ] Escalation policy (15 min → page manager)
  - [ ] **Owner:** DevOps
  - **Deadline:** May 29
  - **Est. Effort:** 2 hours

### CI/CD Pipeline (GitHub Actions)
- [ ] **Test Pipeline**
  ```yaml
  - Run pytest (90%+ coverage required)
  - Run npm test (frontend)
  - Bandit + Safety scan
  - Trivy image scan
  - Lint: black, ruff, eslint
  ```
  - **Owner:** DevOps
  - **Deadline:** May 26

- [ ] **Build & Push**
  ```yaml
  - Build backend Docker image
  - Build frontend Docker image
  - Tag: :latest, :SHA, :date
  - Push to AWS ECR
  ```
  - **Owner:** DevOps
  - **Deadline:** May 27

- [ ] **Deploy to Staging**
  ```yaml
  - kubectl apply staging overlay
  - Run smoke tests (E2E)
  - Verify monitoring dashboards
  ```
  - **Owner:** DevOps
  - **Deadline:** May 27

- [ ] **Manual Approval → Prod Deploy**
  ```yaml
  - Slack notification to #release-approval
  - 2/3 approvals required
  - Blue-green deployment
  - Gradual traffic shift (10% → 50% → 100%)
  ```
  - **Owner:** DevOps
  - **Deadline:** May 27

---

## DOCUMENTATION

- [ ] **API Documentation**
  - [ ] Swagger UI at `/api/docs` (auto-generated)
  - [ ] ReDoc at `/api/redoc`
  - [ ] PostMan collection export
  - [ ] Example cURL requests
  - [ ] **Deadline:** May 28
  - [ ] **Est. Effort:** 2 hours (mostly auto-generated)

- [ ] **Deployment Runbook**
  - [ ] How to deploy to staging/prod
  - [ ] How to rollback
  - [ ] Emergency procedures
  - [ ] Log aggregation (ELK, CloudWatch)
  - [ ] **Deadline:** May 27
  - [ ] **Est. Effort:** 4 hours

- [ ] **Architecture Decision Records (ADRs)**
  - [ ] Why Celery (vs. Lambda)?
  - [ ] Why PostgreSQL (vs. DynamoDB)?
  - [ ] Why SearXNG (vs. Google API)?
  - [ ] **Deadline:** May 30
  - [ ] **Est. Effort:** 3 hours

- [ ] **Runbook for Common Incidents**
  - [ ] Database connection pool exhausted
  - [ ] Celery queue stuck
  - [ ] Payment webhook not triggering
  - [ ] High latency on /scans endpoint
  - [ ] **Deadline:** May 29
  - [ ] **Est. Effort:** 3 hours

---

## BETA TESTING & FEEDBACK

### Pre-Launch Beta (June 1-7)
- [ ] **Invite 50 Beta Testers**
  - [ ] Friends, advisors, early supporters
  - [ ] Mix of US, India, EU (test localization)
  - [ ] Feedback form: usability, bugs, feature requests
  - [ ] NPS survey at end
  - [ ] **Owner:** Product
  - **Deadline:** June 1

- [ ] **Monitor Beta Metrics**
  - [ ] Crash rate (< 1% acceptable)
  - [ ] Scan success rate (> 95%)
  - [ ] Payment success rate (> 98%)
  - [ ] API error rate (< 2%)
  - [ ] **Owner:** DevOps
  - **Daily Review:** June 1-7

- [ ] **Bug Triage**
  - [ ] P0 (crash): Fix within 4 hours
  - [ ] P1 (feature broken): Fix within 24 hours
  - [ ] P2 (minor UX): Fix before GA
  - [ ] **Owner:** Engineering Lead
  - **Daily:** June 1-7

---

## LAUNCH READINESS SIGN-OFF

### Final Pre-Launch Review (May 31)

**Engineering Checklist**
- [ ] All tests pass (90%+ coverage)
- [ ] Security scan: Zero CRITICAL/HIGH CVEs
- [ ] Load test: p99 < 2s, error rate < 1%
- [ ] Database migrations tested (forward + rollback)
- [ ] Backup/restore procedure tested
- [ ] Incident runbooks written
- **Owner:** Engineering Lead
- **Deadline:** May 31, 2 PM UTC

**Product Checklist**
- [ ] User flows E2E (register → scan → removal → payment)
- [ ] Happy path & sad path tested
- [ ] Error messages clear & actionable
- [ ] Mobile responsive (iOS Safari, Android Chrome)
- [ ] Accessibility: WCAG 2.1 AA (focus, contrast, alt text)
- **Owner:** Product Manager
- **Deadline:** May 31, 2 PM UTC

**Legal & Compliance Checklist**
- [ ] Privacy Policy approved by lawyer
- [ ] Terms of Service approved by lawyer
- [ ] GDPR/DPDP compliance verified
- [ ] Audit log retention policy: 1 year
- [ ] Data deletion procedure verified
- **Owner:** Legal Counsel
- **Deadline:** May 25, 12 PM UTC

**Security Checklist**
- [ ] Penetration testing: None yet (schedule Q3)
- [ ] All P1 security fixes merged
- [ ] Secrets not in Git (checked)
- [ ] TLS 1.2+ enforced
- [ ] HSTS header set
- [ ] CSP headers configured
- [ ] Rate limiting tested
- **Owner:** Security Lead
- **Deadline:** May 31, 2 PM UTC

**DevOps Checklist**
- [ ] Staging environment mirrors prod
- [ ] Monitoring dashboards green
- [ ] Alerting rules tested
- [ ] On-call schedule active
- [ ] Disaster recovery plan documented
- [ ] Blue-green deployment tested
- **Owner:** DevOps Lead
- **Deadline:** May 31, 2 PM UTC

---

## LAUNCH DAY (June 8)

### Go/No-Go Decision (June 8, 10 AM UTC)
- All sign-offs complete? → **YES**
- Known critical issues? → **NONE**
- Monitoring ready? → **YES**
- Team on-call? → **YES**
- **Decision:** 🟢 GO FOR LAUNCH

### Launch Sequence
1. **10:30 AM UTC** — Enable open registration
2. **10:45 AM UTC** — Announce on Twitter/LinkedIn
3. **11:00 AM UTC** — Submit to ProductHunt
4. **12:00 PM UTC** — Monitor: Error rate, latency, payment flow
5. **2:00 PM UTC** — First user celebration 🎉
6. **4:00 PM UTC** — Daily standup: bugs, metrics, user feedback

### Post-Launch SLA
- **P0 (Crash/Data Loss):** Fix within 1 hour
- **P1 (Feature broken):** Fix within 4 hours
- **P2 (Minor UX):** Fix within 24 hours
- **Uptime Target:** 99.9%
- **On-Call Rotation:** 24/7 for first 2 weeks

---

## SUCCESS METRICS (First 30 Days)

- [ ] 1,000+ registered users
- [ ] 500+ completed scans
- [ ] 50+ paid subscribers (Basic+)
- [ ] <1% crash rate
- [ ] >95% scan success rate
- [ ] Avg NPS > 40
- [ ] <5% refund rate
- [ ] <2% uptime incidents

---

## ROLLBACK PLAN (If Critical Issues Found)

**Immediate Rollback:**
```bash
kubectl set image deployment/backend backend=privacyshield:previous-sha -n privacyshield
kubectl rollout status deployment/backend
```

**Database Rollback (Schema):**
```bash
docker compose exec backend alembic downgrade -1
# Or: `alembic downgrade 2025-05-31-initial-schema`
```

**If Payment Broken:**
- Disable Stripe webhook → manual payment processing
- Revert to free tier for all
- Notify users: "Payments temporarily offline"

**Communication:**
- Post on status page: status.privacyshield.ai
- Email all users: "Service incident"
- Twitter: "We're experiencing issues"
- ETA to resolution + regular updates

---

## SIGN-OFF

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Engineering Lead | [Name] | ________ | May 31 |
| Product Manager | [Name] | ________ | May 31 |
| Security Lead | [Name] | ________ | May 31 |
| DevOps Lead | [Name] | ________ | May 31 |
| Legal Counsel | [Name] | ________ | May 25 |

**Approval Status:** ⏳ Pending (Expected May 31)

---

**Last Updated:** 2026-05-23  
**Next Update:** Daily (as tasks progress)  
**Questions?** Slack #privacyshield-launch
