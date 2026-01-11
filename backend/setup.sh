#!/bin/bash

echo "🚀 Setting up Password Manager Backend..."

# Couleurs
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Vérifier si Go est installé
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed. Please install Go 1.21 or higher.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Go is installed: $(go version)${NC}"

# Vérifier si PostgreSQL est installé
if ! command -v psql &> /dev/null; then
    echo -e "${RED}❌ PostgreSQL is not installed.${NC}"
    echo "Please install PostgreSQL:"
    echo "  macOS: brew install postgresql@15"
    echo "  Ubuntu: sudo apt install postgresql postgresql-contrib"
    exit 1
fi

echo -e "${GREEN}✅ PostgreSQL is installed${NC}"

# Vérifier si Redis est installé
if ! command -v redis-cli &> /dev/null; then
    echo -e "${RED}⚠️  Redis is not installed (optional for import feature).${NC}"
    echo "To install Redis:"
    echo "  macOS: brew install redis"
    echo "  Ubuntu: sudo apt install redis-server"
fi

# Créer la structure des dossiers
echo -e "${BLUE}📁 Creating project structure...${NC}"
mkdir -p cmd/server
mkdir -p internal/{api/{handlers,middleware},models,repository,services,database,config}
mkdir -p migrations
mkdir -p config

# Initialiser le module Go
echo -e "${BLUE}📦 Initializing Go module...${NC}"
if [ ! -f "go.mod" ]; then
    go mod init github.com/tresor/password-manager
fi

# Installer les dépendances
echo -e "${BLUE}📥 Installing dependencies...${NC}"
go get -u github.com/gin-gonic/gin
go get -u github.com/jmoiron/sqlx
go get -u github.com/lib/pq
go get -u github.com/golang-jwt/jwt/v5
go get -u github.com/google/uuid
go get -u github.com/spf13/viper
go get -u golang.org/x/crypto
go get -u github.com/xlzd/gotp
go get -u github.com/skip2/go-qrcode
go get -u gopkg.in/gomail.v2

go mod tidy

echo -e "${GREEN}✅ Dependencies installed${NC}"

# Créer la base de données
echo -e "${BLUE}🗄️  Setting up database...${NC}"
read -p "Do you want to create the database? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    read -p "PostgreSQL username [postgres]: " PG_USER
    PG_USER=${PG_USER:-postgres}

    read -p "PostgreSQL password: " -s PG_PASSWORD
    echo

    read -p "Database name [password_manager]: " DB_NAME
    DB_NAME=${DB_NAME:-password_manager}

    # Créer la base de données
    PGPASSWORD=$PG_PASSWORD psql -U $PG_USER -h localhost -c "CREATE DATABASE $DB_NAME;" 2>/dev/null

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Database created${NC}"
    else
        echo -e "${BLUE}ℹ️  Database might already exist${NC}"
    fi

    # Exécuter les migrations
    echo -e "${BLUE}🔄 Running migrations...${NC}"
    PGPASSWORD=$PG_PASSWORD psql -U $PG_USER -h localhost -d $DB_NAME -f migrations/001_initial_schema.sql

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Migrations completed${NC}"
    else
        echo -e "${RED}❌ Migration failed${NC}"
        exit 1
    fi
fi

# Créer le fichier .env
if [ ! -f ".env" ]; then
    echo -e "${BLUE}📝 Creating .env file...${NC}"
    cat > .env << EOF
# Server
SERVER_PORT=8000
SERVER_MODE=debug

# Database
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
DATABASE_DBNAME=password_manager
DATABASE_SSLMODE=disable

# Redis (optional)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=$(openssl rand -hex 32)
JWT_EXPIRE_TIME=24

# Email (optional)
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USERNAME=
EMAIL_PASSWORD=
EMAIL_FROM=SecureVault <noreply@securevault.com>

# Have I Been Pwned (optional)
HIBP_API_KEY=
EOF
    echo -e "${GREEN}✅ .env file created${NC}"
fi

echo -e "${GREEN}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Setup completed successfully!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${NC}"
echo "Next steps:"
echo "1. Update the .env file with your credentials"
echo "2. Start Redis (if using import feature): redis-server"
echo "3. Run the server: go run cmd/server/main.go"
echo ""
echo "API will be available at: http://localhost:8000"
echo ""