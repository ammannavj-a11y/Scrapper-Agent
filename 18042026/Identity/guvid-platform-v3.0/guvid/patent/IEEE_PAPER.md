# GUVID: A Country-Agnostic Portable Identity Credential System Using Distributed
# Ledger Anchoring and Hardware-Bound Holder Verification

**Abstract** — Existing identity verification systems require repeated access to
government data sources for each verification event, creating privacy risks and
operational inefficiencies. We present GUVID (Global Universal Verification and
Identity Deployment), a platform that issues portable cryptographic identity
credentials anchored to Hyperledger Fabric 2.5 with zero personally identifiable
information on-chain. A country-adapter plugin architecture enables deployment
across any sovereign jurisdiction without modifying core platform logic. Three
verification dimensions — identity, education, and employment — are combined into
a composite trust score using jurisdiction-configurable weights. W3C WebAuthn
Level 3 passkeys bound at issuance solve the holder binding problem by requiring
hardware-resident cryptographic proof of device possession at presentation. We
demonstrate the system achieves ≥40 TPS for credential issuance and ≥200 TPS for
read-only verification under Hyperledger Caliper benchmarks, with p99 latency
<2000ms under 200 TPS load.

**Keywords** — Digital Identity, Verifiable Credentials, Hyperledger Fabric,
WebAuthn, Decentralised Identity, FIDO2, W3C DID, Zero-Knowledge

---

## I. INTRODUCTION

Digital identity verification underpins hiring decisions, academic credential
validation, and financial onboarding across every modern economy. The dominant
model — verification-as-a-service — requires each relying party to independently
call government APIs and third-party aggregators for every verification event.
In India alone, the identity verification market exceeds $2.1B annually [1],
with each KYC event costing between ₹50–₹500 depending on depth.

This model suffers three fundamental deficiencies. First, it necessitates
repeated transmission of raw national identifiers (Aadhaar numbers, SSNs, NINOs)
through intermediary systems, violating the data minimisation principle codified
in GDPR Article 5(1)(c) and India's Digital Personal Data Protection Act 2023
(DPDP) Section 6 [2][3]. Second, it creates economic redundancy: a job applicant
submitting to ten employers undergoes ten independent verifications of the same
underlying data. Third, the absence of a portable, citizen-controlled credential
prevents individuals from accumulating and presenting a verified identity profile
across contexts.

Prior work on self-sovereign identity (SSI) using DID and Verifiable Credentials
[4] addresses citizen ownership but does not resolve the holder binding problem:
a text credential string can be copied and presented by any party regardless of
identity. Blockchain-anchored document verification systems [5][6] address
immutability but retain centralised verification flows. No prior system combines:
(a) country-agnostic adapter architecture; (b) composite multi-dimensional trust
scoring; (c) hardware-bound holder verification; (d) multi-tenant per-organisation
data isolation; and (e) threshold guardian recovery.

This paper presents GUVID, addressing all five dimensions.

---

## II. SYSTEM ARCHITECTURE

### A. Country Adapter Plugin System

The core innovation enabling country-agnosticism is the adapter plugin architecture.
Each jurisdiction is represented by a declarative YAML capability manifest and
three Go interface implementations:

```
IdentityProvider  { InitiateChallenge, VerifyChallenge, VerifySecondaryID }
EducationProvider { FetchAcademicRecord, VerifyCredential, CrossCheckIssuer }
EmploymentProvider { FetchEmploymentHistory, VerifyTaxRecord, VerifyEmployer }
```

The core platform invokes only these interfaces; government-specific API logic
is entirely encapsulated within adapter packages. Adding a new country requires
implementing three Go files and one YAML manifest — zero changes to core services.

### B. Hash-Based PII Elimination

At the point of ingestion from government APIs, all national identifiers undergo
irreversible transformation:

```
primaryHash   = SHA256(nationalID || countryCode || randomSalt || envSecret)
identityHash  = SHA256(primaryHash || secondaryHash || countryCode || salt || envSalt)
educationHash = SHA256(institutionID || degreeType || graduationYear || salt)
employmentHash = SHA256(employerID || employmentType || startYear || salt)
compositeHash  = SHA256(identityHash || educationHash || employmentHash || salt)
```

Raw national identifiers are neither stored, transmitted post-hashing, nor logged.
Only the composite hash and dimension hashes are propagated to downstream services
and on-chain.

### C. GUVID Token Generation

A globally unique, country-namespaced token is generated:

```
GUVID = "GUV-" || ISO3166_CC || "-" || YEAR || "-" || Base36(compositeHash[0:8])
```

e.g., `GUV-IN-2025-X7K2M9PQ`, `GUV-US-2025-A3BF7RNV`. The token is signed with
ECDSA P-256 over SHA256(GUVID || compositeHash), providing non-repudiation.

### D. Distributed Ledger Anchoring

The Hyperledger Fabric 2.5 chaincode records the following on-chain:

```
{ guvid, compositeHash, countryCode, holderPublicKeyHash,
  trustScore, trustLevel, status, issuedAt, expiresAt }
```

Zero PII fields. CouchDB state database enables rich queries (e.g.,
`QueryHighTrustGUVIDs(minScore, countryCode)`).

---

## III. HOLDER BINDING VIA WEBAUTHN

The holder binding problem — any party possessing the GUVID string can present it
— is solved by WebAuthn Level 3 [7].

At issuance, `navigator.credentials.create()` generates an asymmetric ECDSA P-256
keypair within the device's TPM/Secure Enclave. SHA256(publicKey) is anchored on-chain
as `holderPublicKeyHash`. The private key never leaves the device.

At presentation, the verifier issues a challenge `{nonce, domain, guvid, expiry}`.
The holder's device signs the challenge via `navigator.credentials.get()`. The
verify-svc validates:
1. WebAuthn assertion signature against on-chain public key hash
2. Challenge freshness (nonce TTL: 90 seconds, Redis-enforced)
3. Sign counter monotonicity (replay attack prevention)

This protocol provides three security properties: (1) proof of device possession;
(2) freshness (anti-replay); (3) domain binding (phishing prevention).

---

## IV. TRUST SCORE MODEL

Three dimension scores (0–100) are aggregated with jurisdiction-configurable weights:

```
T = w_i × S_identity + w_e × S_education + w_m × S_employment
```

Default weights reflect regulatory guidance: identity (0.40) carries highest weight
per NIST SP 800-63B IAL2 requirements; education (0.35) per academic credential
verification norms; employment (0.25) as corroborating evidence.

Trust levels: HIGH (T≥85), MEDIUM (T≥65), LOW (T≥40), UNVERIFIED (T<40).

---

## V. MULTI-TENANT ISOLATION ARCHITECTURE

Each participating organisation — HR company, university, regulator — maintains a
physically isolated MariaDB database instance. The `tenant-svc` maps organisation
slugs to database connection strings, injecting `dbName` into JWT claims at
authentication. Every downstream service extracts `dbName` from the JWT and routes
queries accordingly — no shared tables, no cross-tenant data leakage.

This architecture satisfies GDPR's purpose limitation principle: Google HR's
candidate verification data is never accessible to Infosys HR, even on shared
infrastructure.

---

## VI. PERFORMANCE EVALUATION

Benchmarks conducted with Hyperledger Caliper 0.5.0 on a 3-peer Fabric network
(CouchDB state DB, Raft consensus):

| Benchmark | Target TPS | Achieved TPS | p99 Latency |
|-----------|-----------|-------------|-------------|
| GUVID Issue (sustained) | ≥40 | 47.3 | 1,840ms |
| GUVID Verify (read-only) | ≥200 | 234.7 | 410ms |
| HR Audit Log Write | 30 | 31.8 | 620ms |
| Multi-country concurrent | 100 | 108.4 | 1,620ms |

Read-heavy workloads benefit from `EvaluateTransaction` (no consensus required)
and Redis L1 cache, achieving sub-500ms p99 at 200+ TPS.

---

## VII. SECURITY ANALYSIS

**Replay attacks:** Mitigated by nonce TTL (90s) + sign counter monotonicity.

**Key theft:** Private key is hardware-bound (TPM/Secure Enclave). Not extractable.

**Guardian collusion:** 2-of-3 threshold with mandatory guardian diversity
(institutional + corporate + personal). 24-hour delay + liveness check option.

**Synthetic identity:** AI velocity checks + graph-based cluster detection.
Identities sharing device fingerprints across >3 GUVIDs trigger graph alerts.

**PII breach:** No PII on-chain. AES-256-GCM encrypted in tenant DB. Breach of
one tenant DB exposes only that tenant's encrypted records; no cross-tenant exposure.

---

## VIII. CONCLUSION

GUVID demonstrates that a single platform can serve as a universal identity trust
layer across sovereign jurisdictions, resolving the holder binding problem through
hardware passkeys, eliminating PII from distributed ledgers through composite
hashing, and enabling enterprise-grade multi-tenant isolation. The country adapter
plugin system reduces the cost of supporting a new jurisdiction to implementing
three Go interface files. Future work includes zero-knowledge proof integration for
privacy-preserving age/credential verification without revealing specific values.

---

## REFERENCES

[1] "India Digital KYC Market Report," NASSCOM, 2024.
[2] General Data Protection Regulation (GDPR), EU 2016/679, Art. 5(1)(c).
[3] Digital Personal Data Protection Act 2023, Ministry of Electronics and IT, India.
[4] W3C, "Decentralized Identifiers (DIDs) v1.0," W3C Recommendation, July 2022.
[5] S. Nakamoto, "Bitcoin: A Peer-to-Peer Electronic Cash System," 2008.
[6] Hyperledger Foundation, "Hyperledger Fabric 2.5 Architecture," 2023.
[7] W3C, "Web Authentication: An API for accessing Public Key Credentials Level 3," 2023.
[8] NIST, "Digital Identity Guidelines," SP 800-63B, 2022.
[9] W3C, "Verifiable Credentials Data Model v2.0," W3C Recommendation, 2024.
