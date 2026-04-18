# ⚡ GUVID v3.0 — Global Universal Verification & Identity Deployment

> **Verify once. Present everywhere. Cryptographically.**

A country-agnostic sovereign identity platform that issues portable credentials
anchored to Hyperledger Fabric, bound to physical holders via WebAuthn passkeys.

---

## 🚀 One-Command Deploy

```bash
git clone https://github.com/YOUR_ORG/guvid-platform
cd guvid-platform
cp .env.example .env
make quickstart
```

Open **http://localhost:3000**

---

## 👥 Demo Logins (password: `Admin@123`)

| Role | Email | Org |
|------|-------|-----|
| 🏢 HR (Google) | hr@google.com | google-hr |
| 🏢 HR (Infosys) | hr@infosys.com | infosys-hr |
| 🎓 Institution (IIT Delhi) | registrar@iitd.ac.in | iit-delhi |
| 🎓 Institution (MIT) | registrar@mit.edu | mit-edu |
| ⚖️ Regulatory | regulator@dpdp.gov.in | india-regulator |
| 🚨 Fraud L1 | analyst@fraudmonitoring.in | fraud-monitoring |

---

## 🏗 Architecture

**15 microservices** · **Hyperledger Fabric 2.5** · **Apache Kafka** · **WebAuthn L3**

Each organisation gets its own **isolated database** (db_google_hr, db_iit_delhi, etc.)
No cross-tenant data leakage. Zero PII on blockchain.

```
GUVID Token format:   GUV-{CC}-{YEAR}-{HASH}
Example:              GUV-IN-2025-X7K2M9PQ
DID:                  did:guvid:IN:GUV-IN-2025-X7K2M9PQ
```

---

## 📊 Observability

| Service | URL |
|---------|-----|
| Frontend | http://localhost:3000 |
| Grafana | http://localhost:3001 (admin/admin123) |
| Prometheus | http://localhost:9090 |

---

## 📁 Key Documents

- `docs/SOP.md` — Standard Operating Procedures
- `docs/DEPLOYMENT_GUIDE.md` — Deployment, Testing & Monitoring
- `docs/PITCH_DECK.md` — Investor pitch
- `patent/PATENT_DRAFT.md` — Patent application (India + PCT)
- `patent/IEEE_PAPER.md` — Research paper (IEEE/ACM submission-ready)
