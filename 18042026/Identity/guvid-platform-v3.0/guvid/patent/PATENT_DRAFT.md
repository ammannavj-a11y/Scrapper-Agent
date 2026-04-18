# Patent Application Draft

**Title:** SYSTEM AND METHOD FOR ISSUING, ANCHORING, AND CRYPTOGRAPHICALLY
BINDING PORTABLE SOVEREIGN IDENTITY CREDENTIALS TO PHYSICAL HOLDERS USING
DECENTRALISED LEDGER TECHNOLOGY AND HARDWARE-BOUND PASSKEYS

**Applicant:** [Applicant Name / Organisation]
**Filing Jurisdiction:** Indian Patent Office (first filing) + PCT (international)
**Priority Date:** [Date of First Filing]
**Classification:** G06F 21/33; G06F 21/44; H04L 9/32; G06Q 50/10

---

## FIELD OF THE INVENTION

This invention relates to digital identity systems and, more particularly, to a
country-agnostic platform for issuing portable verifiable credentials anchored to
a permissioned distributed ledger, bound to physical holders via hardware-resident
cryptographic keys, and verifiable across sovereign jurisdictions without
re-verification of source government data.

---

## BACKGROUND

Existing identity verification systems operate as centralised, on-demand services
requiring repeated API calls to government data sources for every verification
event. Systems such as eKYC aggregators require raw national identifiers to pass
through intermediary systems, creating privacy risks and violating data minimisation
principles of GDPR, India's DPDP Act 2023, and equivalent statutes.

No prior art discloses: (a) a country-adapter plugin architecture enabling
jurisdiction-agnostic deployment of a unified verification platform; (b) a
composite trust score computed from three independent verification dimensions;
(c) WebAuthn hardware passkey binding of credentials to physical holders on a
blockchain-anchored record; (d) a multi-tenant per-organisation database isolation
architecture combined with zero-PII on-chain storage.

---

## SUMMARY OF THE INVENTION

The invention provides a system and method comprising:

1. A **Country Adapter Plugin System** wherein each sovereign jurisdiction is
   represented by a declarative capability manifest (adapter.yaml) and three Go
   interface implementations (IdentityProvider, EducationProvider, EmploymentProvider),
   enabling the core platform to operate unchanged across any country.

2. A **Composite Verification Hash Algorithm** computing:
   ```
   identityHash  = SHA256(primaryIDHash || secondaryIDHash || countryCode || salt || envSalt)
   educationHash = SHA256(institutionID || degreeType || year || countryCode || salt)
   employmentHash = SHA256(employerID || employmentType || startYear || countryCode || salt)
   compositeHash = SHA256(identityHash || educationHash || employmentHash || countryCode || salt)
   GUVID         = "GUV-" || countryCode || "-" || year || "-" || Base36(compositeHash[0:8])
   ```
   wherein raw national identifiers are irreversibly transformed at the point of
   ingestion and never stored, transmitted, or logged thereafter.

3. A **Trust Score Computation** combining dimension scores with country-configurable
   weights: `T = w_i × S_identity + w_e × S_education + w_m × S_employment`
   producing a normalised score classifying credentials as HIGH (≥85), MEDIUM (≥65),
   LOW (≥40), or UNVERIFIED (<40).

4. A **Distributed Ledger Anchoring Method** using Hyperledger Fabric 2.5 wherein
   only cryptographic hashes, trust scores, status, country codes, and holder public
   key hashes are stored on-chain, with zero personally identifiable information.

5. A **Hardware-Bound Holder Verification Method** using W3C WebAuthn Level 3
   wherein at credential issuance the holder's device generates an asymmetric
   ECDSA P-256 keypair in a Trusted Platform Module or Secure Enclave; the public
   key hash is anchored on-chain; and at presentation the holder signs a
   verifier-issued challenge proving physical device possession without revealing
   the private key.

6. A **Multi-Tenant Isolated Database Architecture** wherein each participating
   organisation (HR, institution, regulator) maintains a dedicated database
   instance, query-routed by the tenant-svc at runtime via JWT claims, ensuring
   complete data isolation between competing organisations.

7. A **Guardian-Based Recovery System** wherein recovery of a compromised holder
   key requires threshold approval (2-of-3) from designated institutional,
   corporate, and personal guardians, with on-chain recording of
   `UpdateHolderKey(guvid, newKeyHash)` and revocation of the prior key.

---

## CLAIMS

**Claim 1:** A computer-implemented system for issuing portable identity credentials,
comprising: (a) a country adapter plugin module configured to normalise identity
verification data from any sovereign jurisdiction's government APIs; (b) a hash
computation module configured to compute dimension hashes from normalised data
without retaining source identifiers; (c) a credential issuance module configured
to compute a composite hash, generate a globally unique token, and sign said token
with an ECDSA P-256 key; (d) a distributed ledger anchoring module configured to
record the composite hash and token on a permissioned blockchain without personally
identifiable information.

**Claim 2:** The system of Claim 1, wherein the country adapter plugin module
exposes three standardised interfaces: IdentityProvider, EducationProvider, and
EmploymentProvider, each independently implementable per jurisdiction without
modification of core platform code.

**Claim 3:** The system of Claim 1, further comprising a WebAuthn holder binding
module configured to: anchor a hardware-resident public key hash at credential
issuance; issue cryptographic challenges at presentation; verify holder assertions
against the anchored hash; and detect replay attacks via monotonically increasing
signature counters.

**Claim 4:** A method for verifying a portable identity credential, comprising:
receiving a credential token from a presenting party; retrieving the associated
composite hash and holder public key hash from a distributed ledger; issuing a
nonce-based challenge to the presenting party's device; receiving a WebAuthn
assertion signed by the device's hardware-resident private key; verifying the
assertion signature against the on-chain public key hash; and returning a
verification result comprising trust level, dimension scores, and holder binding
status, wherein no government API call is required.

**Claim 5:** The method of Claim 4, wherein the nonce expires after 90 seconds
and the signature counter is verified to exceed the stored count, preventing replay
attacks.

**Claim 6:** A method for recovering a compromised holder key, comprising:
receiving a recovery request from a credential holder; notifying a pre-registered
set of guardians comprising at least one institutional guardian, one corporate
guardian, and one personal guardian; collecting signed approval messages from said
guardians; upon receiving threshold approvals, generating a new asymmetric key
pair; recording `UpdateHolderKey(credentialToken, newPublicKeyHash)` on the
distributed ledger; and invalidating the prior public key hash.

**Claim 7:** The system of Claim 1, wherein multi-tenant data isolation is achieved
by: encoding a database name in a JSON Web Token claim at authentication; routing
all database queries to the tenant-specific database instance at runtime; and
preventing cross-tenant data access through query-level scope enforcement.

---

## ABSTRACT

A country-agnostic platform issues portable sovereign identity credentials
(GUVID tokens) anchored to Hyperledger Fabric with zero personally identifiable
information on-chain. Country-specific verification adapters normalise government
API responses into standardised records; composite SHA-256 hashes bind three
verification dimensions (identity, education, employment) into a single token.
WebAuthn Level 3 passkeys bound at issuance enable holder binding — proving the
presenter physically holds the registered device — without re-verification of
government sources. Multi-tenant database isolation ensures each participating
organisation's data is segregated. A guardian-based threshold recovery system
enables key replacement upon device compromise. The platform is compatible with
W3C Verifiable Credentials 2.0 and DID Core 1.0 standards.

---

## JURISDICTIONS FOR FILING

1. **India** — Indian Patent Office (IPO), Form 1 + Form 2, provisional → complete
2. **United States** — USPTO, via PCT national phase (35 U.S.C. § 371)
3. **European Patent Office** — EPO, via PCT Chapter II
4. **PCT** — WIPO PCT international application (priority from India filing)

**Recommended prosecution strategy:** File provisional in India → File PCT within
12 months → Enter national phase in US and EP within 30 months of priority date.
