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

# Check if Halo node is running with archive mode
echo "🔍 Checking if Halo Chain node is running..."
if curl -s -X POST -H "Content-Type: application/json" \
    --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
    http://localhost:8545 2>&1 | grep -q "result"; then
    echo "   ✅ Halo Chain node is running"

    # Get current block number
    BLOCK_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
        --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
        http://localhost:8545)
    
    if [[ $BLOCK_RESPONSE == *"result"* ]]; then
        BLOCK_HEX=$(echo $BLOCK_RESPONSE | grep -o '"result":"[^"]*"' | cut -d'"' -f4)
        BLOCK_NUM=$((16#${BLOCK_HEX#0x}))
        echo "   Current block: $BLOCK_NUM"
    fi

    # Check if archive mode is enabled (required for Blockscout)
    echo ""
    echo "🔍 Checking if archive mode is enabled..."
    DEBUG_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
        --data '{"jsonrpc":"2.0","method":"debug_traceBlockByNumber","params":["latest",{}],"id":1}' \
        http://localhost:8545 2>&1)
    
    if [[ $DEBUG_RESPONSE == *"result"* ]]; then
        echo "   ✅ Archive mode is enabled (debug_traceBlockByNumber works)"
    elif [[ $DEBUG_RESPONSE == *"the method debug_traceBlockByNumber does not exist"* ]]; then
        echo "   ⚠️  WARNING: Archive mode may not be enabled or debug API not exposed"
        echo "   Blockscout requires: --gcmode=archive --http.api=eth,net,web3,debug,txpool"
        echo ""
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    else
        echo "   ⚠️  Could not verify archive mode"
    fi
else
    echo "   ❌ Halo Chain node is not running on localhost:8545"
    echo "   Please start your node with:"
    echo "   geth --gcmode=archive --syncmode=full --http --http.addr=0.0.0.0 \\"
    echo "        --http.port=8545 --http.api=eth,net,web3,debug,txpool \\"
    echo "        --http.vhosts='*' --http.corsdomain='*'"
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
COMPOSE_FILE="$PROJECT_DIR/docker-compose-blockscout.yml"

if [ ! -f "$COMPOSE_FILE" ]; then
    echo "❌ docker-compose-blockscout.yml not found in $PROJECT_DIR"
    exit 1
fi

if grep -q "CHANGEME_GENERATE_WITH_OPENSSL_RAND_BASE64_64\|M6w5Z1YG3uYeRLYIEzadExlBc5Crik7neBmcQc/hzE+2JP6MX/D4CmKNp0CeqQHCiJV26eqGw0d5cBRaxQpWEw==" "$COMPOSE_FILE"; then
    # Create backup
    cp "$COMPOSE_FILE" "$COMPOSE_FILE.bak"
    
    # Replace secret key
    sed -i.tmp "s|SECRET_KEY_BASE: 'CHANGEME_GENERATE_WITH_OPENSSL_RAND_BASE64_64'|SECRET_KEY_BASE: '$SECRET_KEY'|g" "$COMPOSE_FILE"
    sed -i.tmp "s|SECRET_KEY_BASE: 'M6w5Z1YG3uYeRLYIEzadExlBc5Crik7neBmcQc/hzE+2JP6MX/D4CmKNp0CeqQHCiJV26eqGw0d5cBRaxQpWEw=='|SECRET_KEY_BASE: '$SECRET_KEY'|g" "$COMPOSE_FILE"
    
    # Remove sed backup files
    rm -f "$COMPOSE_FILE.tmp" "$COMPOSE_FILE.bak"
    
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
$DOCKER_COMPOSE -f "$COMPOSE_FILE" pull
echo ""

# Start services
echo "🚀 Starting Blockscout services..."
$DOCKER_COMPOSE -f "$COMPOSE_FILE" up -d
echo ""

# Wait for services to initialize
echo "⏳ Waiting for services to initialize (30 seconds)..."
sleep 30

# Check service status
echo ""
echo "📊 Service Status:"
$DOCKER_COMPOSE -f "$COMPOSE_FILE" ps
echo ""

# Wait for backend health endpoint to respond (using correct endpoint)
echo "⏳ Waiting for Blockscout backend to be ready..."
MAX_WAIT=120
COUNTER=0
BACKEND_READY=false

while [ $COUNTER -lt $MAX_WAIT ]; do
    # Try new health endpoint first (/api/health)
    if curl -sf http://localhost:4000/api/health > /dev/null 2>&1; then
        BACKEND_READY=true
        break
    fi
    # Fallback to checking if API root responds
    if curl -sf http://localhost:4000/api > /dev/null 2>&1; then
        BACKEND_READY=true
        break
    fi
    echo -n "."
    sleep 2
    COUNTER=$((COUNTER + 2))
done

echo ""

if [ "$BACKEND_READY" = true ]; then
    echo "✅ Blockscout backend is ready!"
else
    echo "⚠️  Backend did not become ready within $MAX_WAIT seconds"
    echo "   This is normal for first-time setup. Backend may still be initializing."
    echo ""
    echo "   Check backend logs with:"
    echo "   $DOCKER_COMPOSE -f $COMPOSE_FILE logs backend"
fi

# Wait for frontend to be ready
echo ""
echo "⏳ Waiting for Blockscout frontend to be ready..."
MAX_WAIT=60
COUNTER=0
FRONTEND_READY=false

while [ $COUNTER -lt $MAX_WAIT ]; do
    if curl -sf http://localhost:3000 > /dev/null 2>&1; then
        FRONTEND_READY=true
        break
    fi
    echo -n "."
    sleep 2
    COUNTER=$((COUNTER + 2))
done

echo ""

if [ "$FRONTEND_READY" = true ]; then
    echo "✅ Blockscout frontend is ready!"
else
    echo "⚠️  Frontend did not become ready within $MAX_WAIT seconds"
    echo "   Check frontend logs with:"
    echo "   $DOCKER_COMPOSE -f $COMPOSE_FILE logs frontend"
fi

echo ""
echo "======================================"
echo "✅ Blockscout deployment complete!"
echo ""
echo "📍 Access Points:"
echo "   Frontend UI:        http://localhost:3000"
echo "   Backend API:        http://localhost:4000"
echo "   Stats Service:      http://localhost:8080"
echo "   Visualizer:         http://localhost:8081"
echo "   Smart Contract Ver: http://localhost:8050"
echo "   PostgreSQL:         localhost:5433"
echo "   Redis:              localhost:6379"
echo ""
echo "📋 Useful Commands:"
echo "   View all logs:       $DOCKER_COMPOSE -f $COMPOSE_FILE logs -f"
echo "   View backend logs:   $DOCKER_COMPOSE -f $COMPOSE_FILE logs -f backend"
echo "   View frontend logs:  $DOCKER_COMPOSE -f $COMPOSE_FILE logs -f frontend"
echo "   Stop services:       $DOCKER_COMPOSE -f $COMPOSE_FILE down"
echo "   Restart service:     $DOCKER_COMPOSE -f $COMPOSE_FILE restart <service_name>"
echo "   Remove all data:     $DOCKER_COMPOSE -f $COMPOSE_FILE down -v"
echo ""
echo "🔍 Testing the Setup:"
echo "   Test backend API:    curl http://localhost:4000/api/health"
echo "   Test search API:     curl 'http://localhost:4000/api/v2/search?q=0x'"
echo "   Test frontend:       curl http://localhost:3000"
echo ""
echo "⚠️  IMPORTANT NOTES:"
echo "   - Initial indexing may take time depending on your blockchain height"
echo "   - Make sure your Geth node is running with archive mode:"
echo "     --gcmode=archive --syncmode=full"
echo "   - The frontend search requires the backend API to be fully initialized"
echo "   - If search doesn't work immediately, wait 5-10 minutes for indexing to start"
echo ""

# Final health check
echo "🏥 Running final health checks..."
echo ""

# Check backend health
if curl -sf http://localhost:4000/api/health > /dev/null 2>&1; then
    echo "   ✅ Backend health check: PASSED"
else
    echo "   ⚠️  Backend health check: PENDING (may still be initializing)"
fi

# Check if backend can connect to node
NODE_CHECK=$(curl -sf http://localhost:4000/api/v2/stats 2>&1)
if [[ $NODE_CHECK == *"error"* ]] || [ -z "$NODE_CHECK" ]; then
    echo "   ⚠️  Backend-to-Node connection: Check logs if issues persist"
else
    echo "   ✅ Backend-to-Node connection: OK"
fi

# Check frontend
if curl -sf http://localhost:3000 > /dev/null 2>&1; then
    echo "   ✅ Frontend health check: PASSED"
else
    echo "   ⚠️  Frontend health check: PENDING"
fi

echo ""
echo "======================================"
echo "🎉 Setup complete! Access the explorer at http://localhost:3000"
echo ""