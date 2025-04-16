# Documentation de la Structure de Base de Données - Gestionnaire de Secrets

## Vue d'ensemble

Cette documentation décrit la structure complète de la base de données `docksky_vault`, conçue pour stocker les données du gestionnaire de secrets. La base de données comprend une série de tables interconnectées permettant de gérer les utilisateurs, organisations, projets, secrets et abonnements, avec une séparation claire entre les métadonnées et les valeurs des secrets elles-mêmes.

## Utilisateurs et Accès

La base de données utilise plusieurs utilisateurs MySQL avec des privilèges distincts :

| Utilisateur | Description | Privilèges |
|-------------|-------------|-----------|
| `vault_mysql_user` | Utilisateur principal | Tous droits sur docksky_vault |
| `vault_storage_user` | Utilisateur dédié pour Vault | SELECT, INSERT, UPDATE, DELETE uniquement sur vault_storage |
| `secrets_api_user` | Utilisateur pour l'API | Accès aux tables de métadonnées uniquement |
| `monitoring_user` | Utilisateur pour la surveillance | SELECT sur toutes les tables sauf vault_storage |

## Structure des Tables

### 1. Stockage des Secrets

#### `vault_storage`
- `vault_key` VARCHAR(512) [PK] - Clé unique pour le secret
- `vault_value` MEDIUMBLOB - Valeur chiffrée du secret

### 2. Gestion des Utilisateurs

#### `users`
- `id` VARCHAR(36) [PK] - Identifiant unique
- `email` VARCHAR(255) [UQ] - Email (unique)
- `hashed_password` VARCHAR(255) - Mot de passe haché
- `first_name` VARCHAR(255) - Prénom
- `last_name` VARCHAR(255) - Nom
- `role` VARCHAR(50) - Rôle global
- `created_at` DATETIME - Date de création
- `updated_at` DATETIME - Date de mise à jour

### 3. Organisations et Projets

#### `organizations`
- `id` VARCHAR(36) [PK] - Identifiant unique
- `name` VARCHAR(255) - Nom de l'organisation
- `description` TEXT - Description
- `plan_id` VARCHAR(36) - Référence au plan
- `created_at` DATETIME - Date de création
- `updated_at` DATETIME - Date de mise à jour
- `owner_id` VARCHAR(36) [FK] - Référence à l'utilisateur propriétaire

#### `projects`
- `id` VARCHAR(36) [PK] - Identifiant unique
- `name` VARCHAR(255) - Nom du projet
- `description` TEXT - Description
- `organization_id` VARCHAR(36) [FK] - Référence à l'organisation
- `created_at` DATETIME - Date de création
- `updated_at` DATETIME - Date de mise à jour
- `created_by` VARCHAR(36) [FK] - Référence à l'utilisateur créateur

#### `environments`
- `id` VARCHAR(36) [PK] - Identifiant unique
- `name` VARCHAR(255) - Nom de l'environnement
- `description` TEXT - Description
- `project_id` VARCHAR(36) [FK] - Référence au projet
- `created_at` DATETIME - Date de création
- `updated_at` DATETIME - Date de mise à jour

#### `user_organizations`
- `user_id` VARCHAR(36) [PK, FK] - Référence à l'utilisateur
- `organization_id` VARCHAR(36) [PK, FK] - Référence à l'organisation
- `role` VARCHAR(50) - Rôle dans l'organisation
- `created_at` DATETIME - Date de création
- `updated_at` DATETIME - Date de mise à jour

### 4. Métadonnées des Secrets

#### `secret_metadata`
- `id` VARCHAR(36) [PK] - Identifiant unique
- `name` VARCHAR(255) - Nom du secret
- `description` TEXT - Description
- `organization_id` VARCHAR(36) [FK] - Référence à l'organisation
- `project_id` VARCHAR(36) [FK] - Référence au projet
- `environment` VARCHAR(36) - Environnement
- `created_by` VARCHAR(36) [FK] - Référence à l'utilisateur créateur
- `created_at` DATETIME - Date de création
- `updated_at` DATETIME - Date de mise à jour
- `version` INT - Version du secret

### 5. Facturation et Abonnements

#### `plans`
- `id` VARCHAR(36) [PK] - Identifiant unique
- `name` VARCHAR(255) - Nom du plan
- `description` TEXT - Description
- `price` DECIMAL(10,2) - Prix
- `billing_cycle` VARCHAR(50) - Cycle de facturation
- `secrets_limit` INT - Limite de secrets
- `features` JSON - Fonctionnalités incluses
- `created_at` DATETIME - Date de création
- `updated_at` DATETIME - Date de mise à jour

#### `subscriptions`
- `id` VARCHAR(36) [PK] - Identifiant unique
- `organization_id` VARCHAR(36) [FK] - Référence à l'organisation
- `plan_id` VARCHAR(36) - Référence au plan
- `status` VARCHAR(50) - Statut (active, canceled, etc.)
- `secrets_limit` INT - Limite de secrets
- `start_date` DATETIME - Date de début
- `end_date` DATETIME - Date de fin
- `created_at` DATETIME - Date de création
- `updated_at` DATETIME - Date de mise à jour

### 6. Audit et Surveillance

#### `audit_logs`
- `id` VARCHAR(36) [PK] - Identifiant unique
- `user_id` VARCHAR(36) [FK] - Référence à l'utilisateur
- `organization_id` VARCHAR(36) [FK] - Référence à l'organisation
- `action` VARCHAR(50) - Action effectuée
- `resource_type` VARCHAR(50) - Type de ressource
- `resource_id` VARCHAR(36) - Identifiant de la ressource
- `timestamp` DATETIME - Date et heure
- `ip_address` VARCHAR(45) - Adresse IP
- `user_agent` TEXT - Agent utilisateur

#### `usage_statistics`
- `id` VARCHAR(36) [PK] - Identifiant unique
- `organization_id` VARCHAR(36) [FK] - Référence à l'organisation
- `secret_count` INT - Nombre de secrets
- `api_calls` INT - Nombre d'appels API
- `last_updated` DATETIME - Dernière mise à jour

## Diagramme des Relations

```
users 1 --- * user_organizations * --- 1 organizations
organizations 1 --- * projects
projects 1 --- * environments
organizations 1 --- * secret_metadata
projects 1 --- * secret_metadata
users 1 --- * secret_metadata (created_by)
organizations 1 --- * subscriptions
```

## Configuration Vault

Vault devra être configuré pour utiliser MySQL comme backend de stockage avec les paramètres suivants:

```hcl
storage "mysql" {
  username = "vault_storage_user"
  password = "***"
  database = "docksky_vault"
  table = "vault_storage"
}
```

## Considérations pour le Développement

1. **Chemins Vault** - Format recommandé: 
   `organizations/{org_id}/projects/{project_id}/environments/{env}/secrets/{secret_name}`

2. **Rotation des secrets** - Utiliser le champ `version` dans `secret_metadata` pour suivre les versions

3. **Limites d'utilisation** - Vérifier `subscriptions.secrets_limit` avant de créer de nouveaux secrets

4. **Audit** - Enregistrer toutes les opérations de lecture/écriture dans `audit_logs`

5. **Statistiques** - Mettre à jour `usage_statistics` après chaque opération

## Prochaines Étapes

1. Configurer Vault pour utiliser cette base de données
2. Développer les services API pour interagir avec Vault et la base de données
3. Implémenter la couche d'authentification et d'autorisation
4. Développer l'interface utilisateur Flutter
