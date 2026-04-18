-- GUVID v3.0 Platform DB — runs once on platform schema
CREATE DATABASE IF NOT EXISTS guvid_platform CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE guvid_platform;

CREATE TABLE tenants (
  id VARCHAR(36) PRIMARY KEY,
  slug VARCHAR(64) UNIQUE NOT NULL,
  org_name VARCHAR(256) NOT NULL,
  org_type ENUM('hr','institution','regulatory','regulatory_fraud','citizen_portal','government') NOT NULL,
  country_code VARCHAR(2) NOT NULL DEFAULT 'IN',
  db_name VARCHAR(64) NOT NULL,
  db_host VARCHAR(128) NOT NULL DEFAULT 'mariadb',
  db_port INT NOT NULL DEFAULT 3306,
  plan ENUM('free','standard','enterprise') NOT NULL DEFAULT 'standard',
  logo_url VARCHAR(512),
  primary_color VARCHAR(7) DEFAULT '#00d4ff',
  custom_domain VARCHAR(256),
  is_active TINYINT(1) NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_slug(slug), INDEX idx_type(org_type), INDEX idx_country(country_code)
) ENGINE=InnoDB;

CREATE TABLE tenant_users (
  id VARCHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  email VARCHAR(256) NOT NULL,
  email_hash VARCHAR(64) NOT NULL,
  password_hash VARCHAR(128) NOT NULL,
  full_name VARCHAR(256),
  role VARCHAR(50) NOT NULL DEFAULT 'member',
  is_active TINYINT(1) NOT NULL DEFAULT 1,
  last_login DATETIME,
  mfa_secret VARCHAR(64),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_tenant_email(tenant_id, email_hash),
  INDEX idx_tenant(tenant_id), INDEX idx_email_hash(email_hash)
) ENGINE=InnoDB;

CREATE TABLE sessions (
  id VARCHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  user_id VARCHAR(36) NOT NULL,
  token_hash VARCHAR(64) NOT NULL UNIQUE,
  role VARCHAR(50) NOT NULL,
  ip_address VARCHAR(45),
  user_agent VARCHAR(512),
  expires_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_token(token_hash), INDEX idx_user(user_id)
) ENGINE=InnoDB;

CREATE TABLE api_keys (
  id VARCHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  key_hash VARCHAR(64) NOT NULL UNIQUE,
  label VARCHAR(128),
  scopes JSON,
  rate_limit_rpm INT NOT NULL DEFAULT 1000,
  is_active TINYINT(1) NOT NULL DEFAULT 1,
  expires_at DATETIME,
  last_used DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_tenant(tenant_id)
) ENGINE=InnoDB;

-- Seed demo tenants
INSERT INTO tenants VALUES
('t-hr-google','google-hr','Google LLC','hr','US','db_google_hr','mariadb',3306,'enterprise',NULL,'#4285F4',NULL,1,NOW()),
('t-hr-infosys','infosys-hr','Infosys Technologies','hr','IN','db_infosys_hr','mariadb',3306,'enterprise',NULL,'#007CC3',NULL,1,NOW()),
('t-inst-iit','iit-delhi','IIT Delhi','institution','IN','db_iit_delhi','mariadb',3306,'enterprise',NULL,'#B22222',NULL,1,NOW()),
('t-inst-mit','mit-edu','MIT','institution','US','db_mit_edu','mariadb',3306,'enterprise',NULL,'#A31F34',NULL,1,NOW()),
('t-reg-india','india-regulator','DPDP Authority India','regulatory','IN','db_india_reg','mariadb',3306,'enterprise',NULL,'#FF9933',NULL,1,NOW()),
('t-fraud-l1','fraud-monitoring','Fraud Monitoring L1','regulatory_fraud','IN','db_fraud_l1','mariadb',3306,'enterprise',NULL,'#FF4D6D',NULL,1,NOW());

-- Seed users (password: Admin@123 → bcrypt hash placeholder)
INSERT INTO tenant_users VALUES
('u-google-1','t-hr-google','hr@google.com',SHA2('hr@google.com',256),'$2b$12$placeholder_google','Google HR Admin','admin',1,NULL,NULL,NOW()),
('u-infosys-1','t-hr-infosys','hr@infosys.com',SHA2('hr@infosys.com',256),'$2b$12$placeholder_infosys','Infosys HR Admin','admin',1,NULL,NULL,NOW()),
('u-iit-1','t-inst-iit','registrar@iitd.ac.in',SHA2('registrar@iitd.ac.in',256),'$2b$12$placeholder_iit','IIT Delhi Registrar','admin',1,NULL,NULL,NOW()),
('u-mit-1','t-inst-mit','registrar@mit.edu',SHA2('registrar@mit.edu',256),'$2b$12$placeholder_mit','MIT Registrar','admin',1,NULL,NULL,NOW()),
('u-reg-1','t-reg-india','regulator@dpdp.gov.in',SHA2('regulator@dpdp.gov.in',256),'$2b$12$placeholder_reg','DPDP Authority','admin',1,NULL,NULL,NOW()),
('u-fraud-1','t-fraud-l1','analyst@fraudmonitoring.in',SHA2('analyst@fraudmonitoring.in',256),'$2b$12$placeholder_fraud','Fraud Analyst L1','analyst',1,NULL,NULL,NOW());
