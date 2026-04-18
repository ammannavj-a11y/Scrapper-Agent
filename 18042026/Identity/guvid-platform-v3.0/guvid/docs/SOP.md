# GUVID v3.0 — Standard Operating Procedures (SOP)

**Document ID:** GUVID-SOP-001 | **Version:** 3.0 | **Classification:** Internal

---

## 1. System Overview

GUVID (Global Universal Verification & Identity Deployment) is a sovereign-grade
identity verification platform that issues portable cryptographic identity credentials
anchored to Hyperledger Fabric. Each credential (GUVID token) binds a citizen's
verified identity, education, and employment to a WebAuthn passkey held on their
device.

---

## 2. User Roles & Responsibilities

| Role | Access | Responsibilities |
|------|--------|-----------------|
| **Citizen** | Self-service portal | Issue GUVID, manage consent, view audit trail, initiate recovery |
| **Institution** | Institution portal | Issue W3C VC 2.0 credentials (degrees, certificates) |
| **HR Organisation** | HR portal | Verify candidates, batch verify, view audit history |
| **Regulatory Authority** | Read-only dashboard | Platform oversight, compliance reports, cross-org audit |
| **Fraud Analyst L1** | Fraud dashboard | Review AI-detected incidents, resolve/escalate, graph analysis |
| **Platform Admin** | Full access | Tenant management, adapter config, system health |

---

## 3. Onboarding Procedures

### 3.1 Onboarding a New Institution
1. Admin creates tenant via `POST /api/v1/tenants` with `org_type=institution`
2. System provisions isolated MariaDB database (`db_{slug}`)
3. Run migration: `mysql -u root db_{slug} < migrations/tenant/002_tenant_template.sql`
4. Admin creates institution user via `POST /api/v1/tenant-users`
5. Institution receives credentials; logs in at `http://platform-url?tenant={slug}`
6. Institution tests credential issuance with demo GUVID

### 3.2 Onboarding a New HR Organisation
1. Same as 3.1 with `org_type=hr`
2. HR admin verifies at least one demo GUVID to test integration
3. Review rate limits (default 10,000 verifications/day on standard plan)

### 3.3 Citizen Self-Registration
1. Citizen visits `/` and selects "Citizen" login type
2. Completes 5-step wizard: Country → Primary ID OTP → Secondary ID → Education → Employment
3. System issues GUVID token and registers WebAuthn passkey to device
4. Citizen receives GUVID string (e.g. `GUV-IN-2025-X7K2M9PQ`)

---

## 4. Credential Issuance SOP

1. Institution logs in with org credentials
2. Navigates to "Issue Credential"
3. Enters holder's GUVID (verified or from student record)
4. Fills credential metadata (degree, field, year, grade)
5. Toggles consent confirmation
6. Clicks "Issue to Fabric" — system:
   a. Encrypts PII fields with AES-256-GCM
   b. Computes credential hash
   c. Generates W3C VC 2.0 JSON (no PII)
   d. Anchors hash + metadata to Hyperledger Fabric
   e. Stores encrypted record in institution's isolated DB
7. Institution receives credential ID + Fabric TX ID

---

## 5. Candidate Verification SOP (HR)

1. HR logs in to HR portal
2. Enters candidate's GUVID + expected name + position
3. Selects verification type (FULL / QUICK / EDU ONLY)
4. Clicks "Verify" — system:
   a. Calls verify-svc → checks Fabric for GUVID status
   b. Returns trust score, dimension scores, status
   c. Optionally: triggers WebAuthn challenge for holder binding
   d. Stores result in HR organisation's isolated DB
   e. Logs event on Hyperledger Fabric (immutable audit)
5. HR views result with trust level badge + dimension breakdown
6. Result stored permanently in HR org's DB for compliance

---

## 6. Revocation SOP

### 6.1 GUVID Revocation
- Trigger: Fraud detection, user request, court order
- Action: `PUT /api/v1/guvid/revoke` with reason
- On-chain: `RevokeGUVID(guvid, reason, revokedBy)` → Fabric
- Result: All future verifications return `status: revoked`

### 6.2 Credential Revocation
- Trigger: Degree rescinded, error correction, disciplinary action
- Action: `DELETE /api/v1/institution/credentials/{id}`
- On-chain: `RevokeCredential(credID, reason)` → Fabric

---

## 7. Incident Response

| Severity | Response Time | Escalation |
|----------|--------------|-----------|
| Critical fraud | 15 minutes | Fraud L1 → Platform Admin → Regulatory |
| High fraud | 2 hours | Fraud L1 → Platform Admin |
| Compliance breach | 24 hours | Regulatory team |
| Service degradation | 1 hour | On-call engineer |

---

## 8. Data Retention Policy

- GUVID records: 7 years post-expiry
- Verification logs: 7 years
- Fraud checks: 7 years
- Consent events: 3 years
- Session tokens: 24 hours post-expiry
- Raw OTP transactions: NOT stored (hashes only, 1 hour)
