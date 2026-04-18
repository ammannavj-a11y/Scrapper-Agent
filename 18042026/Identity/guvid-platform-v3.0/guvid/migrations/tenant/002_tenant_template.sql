-- Template: runs per-tenant with DB_NAME substituted
-- Each org (Google HR, IIT Delhi, DPDP Regulator) gets their own isolated DB

CREATE DATABASE IF NOT EXISTS {{DB_NAME}} CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE {{DB_NAME}};

-- ── HR Tenant Tables ─────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS hr_verifications (
  id VARCHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  guvid VARCHAR(32) NOT NULL,
  candidate_name_encrypted VARBINARY(512),
  position_applied VARCHAR(256),
  verification_type ENUM('full','identity_only','education_only','quick') NOT NULL DEFAULT 'full',
  trust_level ENUM('HIGH','MEDIUM','LOW','UNVERIFIED'),
  trust_score DECIMAL(5,2),
  identity_score DECIMAL(5,2),
  education_score DECIMAL(5,2),
  employment_score DECIMAL(5,2),
  holder_verified TINYINT(1) NOT NULL DEFAULT 0,
  name_match TINYINT(1),
  fabric_log_id VARCHAR(128),
  verifier_user_id VARCHAR(36),
  result ENUM('pass','fail','pending') NOT NULL DEFAULT 'pending',
  notes TEXT,
  verified_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_guvid(guvid), INDEX idx_result(result), INDEX idx_verified_at(verified_at)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS hr_candidates (
  id VARCHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  guvid VARCHAR(32),
  name_encrypted VARBINARY(512),
  email_hash VARCHAR(64),
  position VARCHAR(256),
  department VARCHAR(128),
  status ENUM('pending','cleared','rejected','watchlist') NOT NULL DEFAULT 'pending',
  last_verified DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_guvid(guvid), INDEX idx_status(status)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS hr_audit_log (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  action VARCHAR(100) NOT NULL,
  actor_id VARCHAR(36) NOT NULL,
  resource_type VARCHAR(50),
  resource_id VARCHAR(36),
  details JSON,
  ip_address VARCHAR(45),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_actor(actor_id), INDEX idx_created(created_at)
) ENGINE=InnoDB;

-- ── Institution Tenant Tables ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS issued_credentials (
  id VARCHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  guvid VARCHAR(32) NOT NULL,
  holder_name_encrypted VARBINARY(512),
  holder_dob_encrypted VARBINARY(256),
  credential_type VARCHAR(100) NOT NULL,
  credential_level VARCHAR(50),
  degree_name VARCHAR(256),
  field_of_study VARCHAR(128),
  graduation_year YEAR,
  grade_encrypted VARBINARY(256),
  roll_number_encrypted VARBINARY(256),
  w3c_vc_json JSON,
  cred_hash VARCHAR(64) UNIQUE,
  fabric_tx_id VARCHAR(128),
  status ENUM('active','revoked','expired') NOT NULL DEFAULT 'active',
  issued_by VARCHAR(36),
  issued_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at DATETIME,
  INDEX idx_guvid(guvid), INDEX idx_status(status), INDEX idx_issued_at(issued_at)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS institution_audit_log (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  action VARCHAR(100) NOT NULL,
  actor_id VARCHAR(36) NOT NULL,
  guvid VARCHAR(32),
  credential_id VARCHAR(36),
  details JSON,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_guvid(guvid), INDEX idx_created(created_at)
) ENGINE=InnoDB;

-- ── Regulatory Tenant Tables ──────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS regulatory_reports (
  id VARCHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  report_type ENUM('compliance','audit','fraud','stats','breach') NOT NULL,
  period_start DATE,
  period_end DATE,
  total_users BIGINT DEFAULT 0,
  total_institutions INT DEFAULT 0,
  total_hr_orgs INT DEFAULT 0,
  total_credentials_issued BIGINT DEFAULT 0,
  total_verifications BIGINT DEFAULT 0,
  total_fraud_blocked INT DEFAULT 0,
  compliance_score DECIMAL(5,2),
  report_json JSON,
  generated_by VARCHAR(36),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_type(report_type), INDEX idx_period(period_start, period_end)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS compliance_events (
  id VARCHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  event_type VARCHAR(100) NOT NULL,
  severity ENUM('info','warning','critical') NOT NULL DEFAULT 'info',
  description TEXT,
  affected_entity_type VARCHAR(50),
  affected_entity_id VARCHAR(36),
  resolved TINYINT(1) NOT NULL DEFAULT 0,
  resolved_by VARCHAR(36),
  resolved_at DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_severity(severity), INDEX idx_resolved(resolved)
) ENGINE=InnoDB;

-- ── Fraud Monitoring Tables ───────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS fraud_incidents (
  id VARCHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  incident_type ENUM('synthetic_identity','velocity_breach','replay_attack','guardian_collusion','device_compromise','graph_cluster') NOT NULL,
  severity ENUM('low','medium','high','critical') NOT NULL,
  risk_score DECIMAL(5,2) NOT NULL,
  guvid VARCHAR(32),
  identity_hash VARCHAR(64),
  ip_address VARCHAR(45),
  device_fingerprint VARCHAR(128),
  ai_signals JSON,
  graph_evidence JSON,
  status ENUM('open','investigating','resolved','false_positive') NOT NULL DEFAULT 'open',
  assigned_to VARCHAR(36),
  resolution_notes TEXT,
  detected_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  resolved_at DATETIME,
  INDEX idx_severity(severity), INDEX idx_status(status), INDEX idx_guvid(guvid), INDEX idx_detected(detected_at)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS fraud_graph_nodes (
  id VARCHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  node_type ENUM('guvid','device','ip','verifier','guardian') NOT NULL,
  node_hash VARCHAR(64) NOT NULL,
  risk_score DECIMAL(5,2) NOT NULL DEFAULT 0,
  flags JSON,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_node(node_type, node_hash)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS fraud_graph_edges (
  id VARCHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  source_hash VARCHAR(64) NOT NULL,
  target_hash VARCHAR(64) NOT NULL,
  edge_type VARCHAR(50) NOT NULL,
  weight DECIMAL(5,2) NOT NULL DEFAULT 1.0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_source(source_hash), INDEX idx_target(target_hash)
) ENGINE=InnoDB;

-- ── Citizen Tables ────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS guvid_records (
  id VARCHAR(36) PRIMARY KEY,
  guvid VARCHAR(32) UNIQUE NOT NULL,
  country_code VARCHAR(2) NOT NULL,
  composite_hash VARCHAR(64) UNIQUE NOT NULL,
  identity_hash VARCHAR(64) NOT NULL,
  education_hash VARCHAR(64),
  employment_hash VARCHAR(64),
  holder_public_key_hash VARCHAR(64),
  did_document TEXT,
  trust_score DECIMAL(5,2) NOT NULL,
  trust_level ENUM('HIGH','MEDIUM','LOW','UNVERIFIED') NOT NULL,
  identity_score DECIMAL(5,2),
  education_score DECIMAL(5,2),
  employment_score DECIMAL(5,2),
  platform_signature TEXT NOT NULL,
  fabric_tx_id VARCHAR(128),
  status ENUM('active','revoked','expired','suspended') NOT NULL DEFAULT 'active',
  issued_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at DATETIME NOT NULL,
  INDEX idx_trust(trust_level), INDEX idx_status(status)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS webauthn_credentials (
  id VARCHAR(36) PRIMARY KEY,
  guvid VARCHAR(32) NOT NULL,
  credential_id VARBINARY(1024) NOT NULL,
  public_key VARBINARY(2048) NOT NULL,
  sign_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
  rp_id VARCHAR(256),
  device_type VARCHAR(50),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_used DATETIME,
  UNIQUE KEY uk_cred(credential_id(255)), INDEX idx_guvid(guvid)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS recovery_requests (
  id VARCHAR(36) PRIMARY KEY,
  guvid VARCHAR(32) NOT NULL,
  request_type ENUM('device_lost','key_compromise','account_recovery') NOT NULL,
  status ENUM('pending','guardian_approval','biometric_check','approved','rejected','expired') NOT NULL DEFAULT 'pending',
  new_key_hash VARCHAR(64),
  guardian_approvals JSON,
  liveness_check_passed TINYINT(1) DEFAULT 0,
  risk_score DECIMAL(5,2),
  ip_address VARCHAR(45),
  expires_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  completed_at DATETIME,
  INDEX idx_guvid(guvid), INDEX idx_status(status)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS guardians (
  id VARCHAR(36) PRIMARY KEY,
  guvid VARCHAR(32) NOT NULL,
  guardian_type ENUM('institutional','corporate','personal') NOT NULL,
  guardian_guvid VARCHAR(32),
  guardian_name_encrypted VARBINARY(512),
  guardian_contact_encrypted VARBINARY(512),
  reputation_score DECIMAL(5,2) NOT NULL DEFAULT 100.0,
  is_active TINYINT(1) NOT NULL DEFAULT 1,
  added_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_guvid(guvid)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS consent_events (
  id VARCHAR(36) PRIMARY KEY,
  guvid VARCHAR(32) NOT NULL,
  verifier_tenant_id VARCHAR(36),
  verifier_name VARCHAR(256),
  action ENUM('approved','denied','fraud_report','blocked_verifier') NOT NULL,
  mode ENUM('PASSIVE','ACTIVE','SILENT') NOT NULL DEFAULT 'PASSIVE',
  fabric_log_id VARCHAR(128),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_guvid(guvid), INDEX idx_created(created_at)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS citizen_audit_trail (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  guvid VARCHAR(32) NOT NULL,
  event_type VARCHAR(100) NOT NULL,
  verifier_org VARCHAR(256),
  verifier_type VARCHAR(50),
  result VARCHAR(20),
  trust_level_at_time VARCHAR(20),
  fabric_ref VARCHAR(128),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_guvid(guvid), INDEX idx_created(created_at)
) ENGINE=InnoDB;
