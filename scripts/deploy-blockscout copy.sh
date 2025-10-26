#!/bin/bash
# Deploy Blockscout explorer for Halo Chain using Docker

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "🔍 Halo Chain - Blockscout Deployment"
echo "======================================"
echo ""

# Check if docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    echo "   Visit: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if docker-compose is installed
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    echo "   Visit: https://docs.docker.com/compose/install/"
    exit 1
fi

# Detect which docker compose command to use
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

echo "✅ Docker installed: $(docker --version)"
echo "✅ Docker Compose installed: $($DOCKER_COMPOSE version --short 2>/dev/null || echo 'unknown')"
echo ""

# Check if Halo node is running
echo "🔍 Checking if Halo Chain node is running..."
if curl -s -X POST -H "Content-Type: application/json" \
    --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
    http://localhost:8545 > /dev/null 2>&1; then
    echo "   ✅ Halo Chain node is running"

    # Get current block number
    BLOCK_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
        --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
        http://localhost:8545)
    if [[ $BLOCK_RESPONSE == *"result"* ]]; then
        echo "   Node is syncing blocks"
    fi
else
    echo "   ⚠️  Halo Chain node is not running on localhost:8545"
    echo "   Please start your node first or update the RPC URL in docker-compose-blockscout.yml"
    echo ""
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi
echo ""

# Generate secret key if not already set
SECRET_KEY_FILE="$PROJECT_DIR/.blockscout-secret"
if [ -f "$SECRET_KEY_FILE" ]; then
    echo "🔑 Using existing secret key"
    SECRET_KEY=$(cat "$SECRET_KEY_FILE")
else
    echo "🔑 Generating new secret key..."
    SECRET_KEY=$(openssl rand -base64 64 | tr -d '\n')
    echo "$SECRET_KEY" > "$SECRET_KEY_FILE"
    chmod 600 "$SECRET_KEY_FILE"
    echo "   ✅ Secret key saved to $SECRET_KEY_FILE"
fi
echo ""

# Update docker-compose file with secret key
echo "⚙️  Configuring Blockscout..."
if grep -q "CHANGEME_GENERATE_WITH_OPENSSL_RAND_BASE64_64" "$PROJECT_DIR/docker-compose-blockscout.yml"; then
    sed -i.bak "s|SECRET_KEY_BASE: 'CHANGEME_GENERATE_WITH_OPENSSL_RAND_BASE64_64'|SECRET_KEY_BASE: '$SECRET_KEY'|g" \
        "$PROJECT_DIR/docker-compose-blockscout.yml"
    echo "   ✅ Secret key configured"
fi
echo ""

# Create necessary directories
echo "📁 Creating directories..."
mkdir -p "$PROJECT_DIR/blockscout-logs"
mkdir -p "$PROJECT_DIR/verifier-cache"
echo "   ✅ Directories created"
echo ""

# Pull Docker images
echo "📥 Pulling Docker images..."
echo "   (This may take a while on first run)"
$DOCKER_COMPOSE -f "$PROJECT_DIR/docker-compose-blockscout.yml" pull
echo ""

# Start services
echo "🚀 Starting Blockscout services..."
$DOCKER_COMPOSE -f "$PROJECT_DIR/docker-compose-blockscout.yml" up -d
echo ""

# Wait for services to be healthy
echo "⏳ Waiting for services to be healthy..."
sleep 5

# Check service status
echo ""
echo "📊 Service Status:"
$DOCKER_COMPOSE -f "$PROJECT_DIR/docker-compose-blockscout.yml" ps
echo ""

# Wait for Blockscout to be ready
echo "⏳ Waiting for Blockscout to be ready..."
MAX_WAIT=60
COUNTER=0
while [ $COUNTER -lt $MAX_WAIT ]; do
    if curl -s http://localhost:4000 > /dev/null 2>&1; then
        echo ""
        echo "✅ Blockscout is ready!"
        break
    fi
    echo -n "."
    sleep 1
    COUNTER=$((COUNTER + 1))
done

if [ $COUNTER -eq $MAX_WAIT ]; then
    echo ""
    echo "⚠️  Blockscout did not start within $MAX_WAIT seconds"
    echo "   Check logs with: $DOCKER_COMPOSE -f docker-compose-blockscout.yml logs blockscout"
    exit 1
fi

echo ""
echo "======================================"
echo "✅ Blockscout deployed successfully!"
echo ""
echo "📍 Access Points:"
echo "   Blockscout UI: http://localhost:4000"
echo "   PostgreSQL:    localhost:5432"
echo "   Redis:         localhost:6379"
echo ""
echo "📋 Useful Commands:"
echo "   View logs:       $DOCKER_COMPOSE -f docker-compose-blockscout.yml logs -f"
echo "   Stop services:   $DOCKER_COMPOSE -f docker-compose-blockscout.yml down"
echo "   Restart:         $DOCKER_COMPOSE -f docker-compose-blockscout.yml restart"
echo "   Remove all data: $DOCKER_COMPOSE -f docker-compose-blockscout.yml down -v"
echo ""
echo "⚠️  IMPORTANT:"
echo "   - The explorer will start indexing blocks from your node"
echo "   - Initial indexing may take time depending on block height"
echo "   - Configure RPC endpoints in docker-compose-blockscout.yml if needed"
echo ""
