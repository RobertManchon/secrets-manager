-- Création de la base de données pour Vault
CREATE DATABASE IF NOT EXISTS docksky_vault CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- 0. Création de l'utilisateur principal
CREATE USER IF NOT EXISTS 'vault_mysql_user'@'%' IDENTIFIED BY 'v4u1t@dm1nKlm5';
GRANT ALL PRIVILEGES ON docksky_vault.* TO 'vault_mysql_user'@'%';

-- Création des utilisateurs dédiés avec des placeholders pour les mots de passe
-- 1. Utilisateur dédié pour le stockage Vault
CREATE USER IF NOT EXISTS 'vault_storage_user'@'%' IDENTIFIED BY 'c1e7$2Sv2Kb8o&';
GRANT SELECT, INSERT, UPDATE, DELETE ON docksky_vault.vault_storage TO 'vault_storage_user'@'%';

-- 2. Utilisateur pour l'API de gestion des secrets
CREATE USER IF NOT EXISTS 'secrets_api_user'@'%' IDENTIFIED BY 'gF!P&#sr03k%Nu';
-- Droits sur les tables de métadonnées
GRANT SELECT, INSERT, UPDATE, DELETE ON docksky_vault.users TO 'secrets_api_user'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON docksky_vault.organizations TO 'secrets_api_user'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON docksky_vault.projects TO 'secrets_api_user'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON docksky_vault.environments TO 'secrets_api_user'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON docksky_vault.secret_metadata TO 'secrets_api_user'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON docksky_vault.subscriptions TO 'secrets_api_user'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON docksky_vault.plans TO 'secrets_api_user'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON docksky_vault.user_organizations TO 'secrets_api_user'@'%';
GRANT INSERT ON docksky_vault.audit_logs TO 'secrets_api_user'@'%';

-- 3. Utilisateur pour la surveillance (monitoring)
CREATE USER IF NOT EXISTS 'monitoring_user'@'%' IDENTIFIED BY '5Kx9*qzfog$8tG';
GRANT SELECT ON docksky_vault.* TO 'monitoring_user'@'%';
-- Retirer l'accès à la table des secrets
REVOKE SELECT ON docksky_vault.vault_storage FROM 'monitoring_user'@'%';

-- Utiliser la base docksky_vault
USE docksky_vault;

-- Table pour le stockage Vault
CREATE TABLE IF NOT EXISTS vault_storage (
  vault_key VARCHAR(512) NOT NULL,
  vault_value MEDIUMBLOB NOT NULL,
  PRIMARY KEY (vault_key)
);

-- Tables pour votre application de gestion de secrets
CREATE TABLE IF NOT EXISTS users (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  hashed_password VARCHAR(255) NOT NULL,
  first_name VARCHAR(255),
  last_name VARCHAR(255),
  role VARCHAR(50) NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS organizations (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  plan_id VARCHAR(36),
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  owner_id VARCHAR(36) NOT NULL,
  FOREIGN KEY (owner_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS projects (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  organization_id VARCHAR(36) NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  created_by VARCHAR(36) NOT NULL,
  FOREIGN KEY (organization_id) REFERENCES organizations(id),
  FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS environments (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  project_id VARCHAR(36) NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  FOREIGN KEY (project_id) REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS secret_metadata (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  organization_id VARCHAR(36) NOT NULL,
  project_id VARCHAR(36) NOT NULL,
  environment VARCHAR(36) NOT NULL,
  created_by VARCHAR(36) NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  version INT NOT NULL DEFAULT 1,
  FOREIGN KEY (organization_id) REFERENCES organizations(id),
  FOREIGN KEY (project_id) REFERENCES projects(id),
  FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS subscriptions (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  organization_id VARCHAR(36) NOT NULL,
  plan_id VARCHAR(36) NOT NULL,
  status VARCHAR(50) NOT NULL,
  secrets_limit INT NOT NULL,
  start_date DATETIME NOT NULL,
  end_date DATETIME NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  FOREIGN KEY (organization_id) REFERENCES organizations(id)
);

CREATE TABLE IF NOT EXISTS plans (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  price DECIMAL(10,2) NOT NULL,
  billing_cycle VARCHAR(50) NOT NULL,
  secrets_limit INT NOT NULL,
  features JSON,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS user_organizations (
  user_id VARCHAR(36) NOT NULL,
  organization_id VARCHAR(36) NOT NULL,
  role VARCHAR(50) NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  PRIMARY KEY (user_id, organization_id),
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (organization_id) REFERENCES organizations(id)
);

CREATE TABLE IF NOT EXISTS audit_logs (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  user_id VARCHAR(36) NOT NULL,
  organization_id VARCHAR(36) NOT NULL,
  action VARCHAR(50) NOT NULL,
  resource_type VARCHAR(50) NOT NULL,
  resource_id VARCHAR(36) NOT NULL,
  timestamp DATETIME NOT NULL,
  ip_address VARCHAR(45),
  user_agent TEXT,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (organization_id) REFERENCES organizations(id)
);

-- Table pour les statistiques d'utilisation
CREATE TABLE IF NOT EXISTS usage_statistics (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  organization_id VARCHAR(36) NOT NULL,
  secret_count INT NOT NULL DEFAULT 0,
  api_calls INT NOT NULL DEFAULT 0,
  last_updated DATETIME NOT NULL,
  FOREIGN KEY (organization_id) REFERENCES organizations(id)
);

-- Réinitialiser les privilèges
FLUSH PRIVILEGES;