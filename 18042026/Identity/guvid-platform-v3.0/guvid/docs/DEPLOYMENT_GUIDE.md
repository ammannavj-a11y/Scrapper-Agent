# GUVID v3.0 — Deployment, Testing & Monitoring Guide

## One-Command Deployment

```bash
git clone https://github.com/YOUR_ORG/guvid-platform
cd guvid-platform
cp .env.example .env
make quickstart
```

Open http://localhost:3000

---

## Demo Credentials (password: Admin@123 for all)

| Role | Email | Tenant Slug |
|------|-------|-------------|
| HR (Google) | hr@google.com | google-hr |
| HR (Infosys) | hr@infosys.com | infosys-hr |
| Institution (IIT Delhi) | registrar@iitd.ac.in | iit-delhi |
| Institution (MIT) | registrar@mit.edu | mit-edu |
| Regulatory | regulator@dpdp.gov.in | india-regulator |
| Fraud L1 | analyst@fraudmonitoring.in | fraud-monitoring |

---

## Architecture: Per-Tenant Isolated Databases

Each organisation has its own isolated MariaDB database:

```
guvid_platform    ← Shared: tenants, users, sessions, api_keys
db_google_hr      ← Google HR: verifications, candidates, audit_log
db_infosys_hr     ← Infosys HR: verifications, candidates, audit_log
db_iit_delhi      ← IIT Delhi: issued_credentials, institution_audit_log
db_mit_edu        ← MIT: issued_credentials, institution_audit_log
db_india_reg      ← India Regulator: regulatory_reports, compliance_events
db_fraud_l1       ← Fraud L1: fraud_incidents, fraud_graph_nodes/edges
```

When an HR user logs in, the JWT contains `dbName=db_google_hr`.
Every API call routes to that tenant's database via `tenant-svc`.

---

## Production Deployment

### Docker Compose (single server)
```bash
docker compose up -d
docker compose logs -f
```

### Kubernetes (Helm)
```bash
helm install guvid ./helm \
  --set global.jwtSecret=$(openssl rand -hex 32) \
  --set global.dbPassword=your_db_password \
  --set global.countryCode=IN
```

### Environment Setup for Production
```bash
# Generate ECDSA signing key
openssl ecparam -name prime256v1 -genkey -noout -out signing.key
export GUVID_SIGNING_PRIVATE_KEY=$(cat signing.key)

# Generate AES key
export AES_KEY=$(openssl rand -hex 32)

# Set JWT secret
export JWT_SECRET=$(openssl rand -hex 32)
```

---

## Adding a New Country

```bash
# 1. Create adapter YAML
cp adapters/IN/adapter.yaml adapters/KE/adapter.yaml
# Edit adapters/KE/adapter.yaml with Kenya API details

# 2. Implement adapter Go files
cp adapters/IN/identity_adapter.go adapters/KE/identity_adapter.go
# Modify to use Huduma Namba / NSSF / KNEC endpoints

# 3. Add mock fixtures
mkdir adapters/KE/fixtures
# Create identity.json, education.json, employment.json

# 4. Add env vars to .env
echo "KENYAN_NIIMS_API_BASE=http://mock-integrations:8099/ke" >> .env

# 5. Restart adapter-registry-svc — auto-discovers new adapter
docker compose restart adapter-registry-svc
```

---

## Health Checks

```bash
# All services
curl http://localhost:8080/health  # api-gateway
curl http://localhost:8094/health  # tenant-svc
curl http://localhost:8082/health  # institution-svc
curl http://localhost:8091/health  # hr-portal-svc
curl http://localhost:8093/health  # regulatory-svc
curl http://localhost:8085/health  # verify-svc
curl http://localhost:8086/health  # wallet-svc

# Platform health summary
curl http://localhost:8080/health | jq .
```

---

## Monitoring

| Dashboard | URL | Credentials |
|-----------|-----|-------------|
| Frontend | http://localhost:3000 | Role-based login |
| Grafana | http://localhost:3001 | admin/admin123 |
| Prometheus | http://localhost:9090 | — |

### Key Metrics to Monitor
- `guvid_issue_total` — GUVIDs issued per country
- `guvid_verify_total` — Verifications per tenant
- `fraud_block_total` — Fraud blocks triggered
- `kafka_consumer_lag` — Kafka consumer lag per topic
- HTTP latency p99 across all services

---

## Testing

### Smoke Test (post-deploy)
```bash
# Health check all services
for port in 8080 8081 8082 8083 8084 8085 8086 8087 8088 8091 8093 8094 8099; do
  status=$(curl -sf http://localhost:$port/health | jq -r .status 2>/dev/null || echo "FAIL")
  echo "Port $port: $status"
done

# Full citizen journey
TOKEN=$(curl -sf -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"hr@google.com","password":"Admin@123","tenantSlug":"google-hr"}' | jq -r .token)

curl -sf http://localhost:8080/api/v1/hr/verify \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-Slug: google-hr" \
  -H "Content-Type: application/json" \
  -d '{"guvid":"GUV-IN-2025-X7K2M9PQ","verifyType":"quick"}'
```

### Load Test
```bash
# Install k6
k6 run --vus 50 --duration 60s scripts/load-test.js
```
