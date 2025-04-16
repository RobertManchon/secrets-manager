# Tableau des utilisateurs de base de données pour le gestionnaire de secrets

| Utilisateur | Description | Base de données | Droits | Utilisé par | Mot de passe (var env) |
|-------------|-------------|----------------|--------|-------------|------------------------|
| `root` | Administrateur global | Toutes | Tous droits sur toutes les bases | Maintenance uniquement | `MYSQL_ROOT_PASSWORD` |
| `vault_mysql_user` | Utilisateur par défaut créé au démarrage | `docksky_vault` | Tous droits sur docksky_vault | Configuration initiale | `VAULT_MYSQL_PASSWORD` |
| `vault_storage_user` | Utilisateur dédié pour Vault | `docksky_vault` | SELECT, INSERT, UPDATE, DELETE uniquement sur la table vault_storage | Service Vault | `VAULT_STORAGE_PASSWORD` |
| `secrets_api_user` | Utilisateur pour l'API | `docksky_vault` | Droits limités aux opérations nécessaires sur les tables de métadonnées | API de gestion des secrets | `SECRETS_API_PASSWORD` |
| `monitoring_user` | Utilisateur pour la surveillance | `docksky_vault` | SELECT uniquement (lecture seule) | Outils de monitoring | `MONITORING_PASSWORD` |

## Droits détaillés par utilisateur

### vault_storage_user
```sql
GRANT SELECT, INSERT, UPDATE, DELETE ON docksky_vault.vault_storage TO 'vault_storage_user'@'%';
```
Cet utilisateur ne doit accéder qu'à la table de stockage de Vault, sans accès aux tables de métadonnées.

### secrets_api_user
```sql
GRANT SELECT, INSERT, UPDATE, DELETE ON docksky_vault.users TO 'secrets_api_user'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON docksky_vault.organizations TO 'secrets_api_user'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON docksky_vault.projects TO 'secrets_api_user'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON docksky_vault.environments TO 'secrets_api_user'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON docksky_vault.secret_metadata TO 'secrets_api_user'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON docksky_vault.subscriptions TO 'secrets_api_user'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON docksky_vault.plans TO 'secrets_api_user'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON docksky_vault.user_organizations TO 'secrets_api_user'@'%';
GRANT INSERT ON docksky_vault.audit_logs TO 'secrets_api_user'@'%';
```
Cet utilisateur gère les métadonnées mais ne doit pas accéder directement aux valeurs des secrets.

### monitoring_user
```sql
GRANT SELECT ON docksky_vault.* TO 'monitoring_user'@'%';
REVOKE SELECT ON docksky_vault.vault_storage FROM 'monitoring_user'@'%';
```
Cet utilisateur peut lire toutes les tables sauf vault_storage pour la surveillance.

## Éléments supplémentaires à considérer

1. **Rotation des mots de passe**
   - Établir une politique de rotation régulière des mots de passe
   - Automatiser ce processus si possible

2. **Audit de base de données**
   - Activer l'audit MySQL pour suivre les accès à la base de données
   - Configurer des alertes en cas d'activité suspecte

3. **Chiffrement des données**
   - Activer le chiffrement des données au repos dans MySQL
   - Configuration TLS pour les connexions à la base de données

4. **Restrictions d'accès réseau**
   - Limiter les connexions à la base de données uniquement aux conteneurs nécessaires
   - Utiliser des réseaux Docker dédiés pour isoler les composants

5. **Sauvegarde et récupération**
   - Configurer des sauvegardes régulières de la base docksky_vault
   - Tester les procédures de restauration

6. **Séparation des préoccupations**
   - Vault gère le stockage sécurisé des secrets
   - Votre API gère les métadonnées, les autorisations et la logique métier
   - Aucun accès direct aux secrets en dehors de Vault

7. **Surveillance spécifique**
   - Surveiller le nombre de secrets par organisation (pour le respect des quotas)
   - Surveiller les tentatives d'accès non autorisées
   - Alerter en cas d'activité anormale

Cette structure respecte le principe du moindre privilège et sépare clairement les responsabilités entre les différents composants de votre système de gestion de secrets.