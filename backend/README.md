# Password Manager Backend

Backend API pour le gestionnaire de mots de passe sécurisé construit avec Go.

## 🚀 Installation Rapide
```bash
# Cloner le repository
git clone <repository-url>
cd password-manager-backend

# Exécuter le script d'installation
./setup.sh

# Démarrer le serveur
make run
```

## 📋 Prérequis

- **Go 1.21+** - [Installer Go](https://golang.org/doc/install)
- **PostgreSQL 15+** - [Installer PostgreSQL](https://www.postgresql.org/download/)
- **Redis** (optionnel, pour les imports)

### Installation des prérequis

**macOS:**
```bash
brew install go postgresql@15
# Redis is optional: brew install redis
brew services start postgresql@15
# Start Redis only if you need import feature: brew services start redis
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install golang postgresql postgresql-contrib
# Redis is optional: sudo apt install redis-server
sudo systemctl start postgresql
# Start Redis only if you need import feature: sudo systemctl start redis
```

## 🛠️ Configuration

1. Copier le fichier `.env.example` vers `.env`:
```bash
cp .env.example .env
```

2. Modifier `.env` avec vos paramètres:
```env
DATABASE_USER=postgres
DATABASE_PASSWORD=votre_mot_de_passe
JWT_SECRET=votre_secret_jwt
```

## 🗄️ Base de données

### Créer la base de données
```bash
# Se connecter à PostgreSQL
psql -U postgres

# Créer la base de données
CREATE DATABASE password_manager;
\q
```

### Exécuter les migrations
```bash
make migrate
```

Ou manuellement:
```bash
psql -U postgres -d password_manager -f migrations/001_initial_schema.sql
```

## ▶️ Démarrage
```bash
# Mode développement
make run

# Avec hot reload (nécessite air)
go install github.com/cosmtrek/air@latest
make dev

# Build et exécution
make build
./bin/server
```

Le serveur démarre sur `http://localhost:8000`

## 📚 API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Créer un compte
- `POST /api/v1/auth/login` - Se connecter

### Vault
- `GET /api/v1/vault` - Liste des mots de passe
- `POST /api/v1/vault` - Créer un mot de passe
- `GET /api/v1/vault/:id` - Détails d'un mot de passe
- `PUT /api/v1/vault/:id` - Modifier un mot de passe
- `DELETE /api/v1/vault/:id` - Supprimer un mot de passe
- `POST /api/v1/vault/generate-password` - Générer un mot de passe

### Health
- `GET /api/v1/health/report` - Rapport de santé des mots de passe
- `POST /api/v1/vault/scan-all` - Scanner tous les mots de passe

### Sharing
- `POST /api/v1/share` - Partager un mot de passe
- `GET /api/v1/shared` - Liste des partages
- `GET /api/v1/shared/:token` - Accéder à un mot de passe partagé

### 2FA
- `POST /api/v1/2fa/enable` - Activer 2FA
- `POST /api/v1/2fa/verify` - Vérifier un code 2FA

### Import
- `POST /api/v1/import/upload` - Uploader un fichier d'import
- `POST /api/v1/import/confirm/:session_id` - Confirmer l'import

## 🧪 Tests
```bash
# Exécuter tous les tests
make test

# Avec couverture
go test -cover ./...
```

## 📦 Structure du projet
```
password-manager-backend/
├── cmd/
│   └── server/
│       └── main.go                 # Point d'entrée
├── internal/
│   ├── api/
│   │   ├── handlers/              # Gestionnaires de requêtes
│   │   ├── middleware/            # Middleware (auth, CORS, etc.)
│   │   └── router.go              # Configuration des routes
│   ├── models/                    # Modèles de données
│   ├── repository/                # Accès à la base de données
│   ├── services/                  # Logique métier
│   ├── database/                  # Configuration DB
│   └── config/                    # Configuration
├── migrations/                    # Migrations SQL
├── config/
│   └── config.yaml               # Configuration
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 🔐 Sécurité

- Chiffrement AES-256-GCM
- Dérivation de clé Argon2id
- Architecture zero-knowledge
- JWT pour l'authentification
- Rate limiting
- 2FA avec TOTP

## 📝 Commandes utiles
```bash
make help       # Voir toutes les commandes
make run        # Lancer le serveur
make build      # Compiler
make test       # Tests
make clean      # Nettoyer
make migrate    # Migrations
```

## 🐛 Dépannage

### Erreur de connexion PostgreSQL
```bash
# Vérifier que PostgreSQL est démarré
pg_isready

# Redémarrer PostgreSQL
brew services restart postgresql@15  # macOS
sudo systemctl restart postgresql    # Linux
```

### Port 8000 déjà utilisé
```bash
# Changer le port dans .env
SERVER_PORT=8001
```

## 📄 Licence

MIT License