#!/bin/bash

# Setup script for SSL certificates with Let's Encrypt
# Usage: ./setup-ssl.sh your-email@example.com

set -e

EXPLORER_DOMAIN="explorer.halofoundation.xyz"
RPC_DOMAIN="rpc.halofoundation.xyz"
EMAIL="${1:-admin@halofoundation.xyz}"

echo "=========================================="
echo "SSL Setup for Halo Foundation Services"
echo "Explorer Domain: $EXPLORER_DOMAIN"
echo "RPC Domain: $RPC_DOMAIN"
echo "Email: $EMAIL"
echo "=========================================="

# Check if email is provided
if [ -z "$EMAIL" ]; then
    echo "Error: Email address is required"
    echo "Usage: ./setup-ssl.sh your-email@example.com"
    exit 1
fi

# Create necessary directories
echo ""
echo "Creating directories..."
mkdir -p certbot/conf
mkdir -p certbot/www

# Get server IP
CURRENT_IP=$(curl -s ifconfig.me)
echo ""
echo "Your server IP: $CURRENT_IP"

# Check DNS for both domains
echo ""
echo "Checking DNS configuration..."
echo "----------------------------------------"

EXPLORER_IP=$(dig +short $EXPLORER_DOMAIN | tail -n1)
echo "Explorer domain ($EXPLORER_DOMAIN) resolves to: $EXPLORER_IP"

RPC_IP=$(dig +short $RPC_DOMAIN | tail -n1)
echo "RPC domain ($RPC_DOMAIN) resolves to: $RPC_IP"

DNS_OK=true
if [ "$CURRENT_IP" != "$EXPLORER_IP" ]; then
    echo "⚠ WARNING: Explorer domain DNS mismatch!"
    DNS_OK=false
fi

if [ "$CURRENT_IP" != "$RPC_IP" ]; then
    echo "⚠ WARNING: RPC domain DNS mismatch!"
    DNS_OK=false
fi

if [ "$DNS_OK" = false ]; then
    echo ""
    echo "Please make sure both domains point to $CURRENT_IP"
    echo ""
    read -p "Do you want to continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Create temporary nginx config for initial certificate
echo ""
echo "Creating temporary nginx config..."
cat > nginx-init.conf << 'EOF'
events {
    worker_connections 1024;
}

http {
    server {
        listen 80;
        server_name explorer.halofoundation.xyz rpc.halofoundation.xyz;

        location /.well-known/acme-challenge/ {
            root /var/www/certbot;
        }

        location / {
            return 200 "Server is ready for SSL setup\n";
            add_header Content-Type text/plain;
        }
    }
}
EOF

# Start nginx temporarily for certificate generation
echo ""
echo "Starting temporary nginx for certificate generation..."
docker run -d --name nginx-temp \
    -p 80:80 \
    -v $(pwd)/nginx-init.conf:/etc/nginx/nginx.conf:ro \
    -v $(pwd)/certbot/www:/var/www/certbot:ro \
    nginx:alpine

# Wait for nginx to start
echo "Waiting for nginx to start..."
sleep 5

# Get SSL certificate for Explorer domain
echo ""
echo "=========================================="
echo "Requesting SSL certificate for $EXPLORER_DOMAIN..."
echo "=========================================="
docker run --rm \
    -v $(pwd)/certbot/conf:/etc/letsencrypt \
    -v $(pwd)/certbot/www:/var/www/certbot \
    certbot/certbot certonly \
    --webroot \
    --webroot-path=/var/www/certbot \
    --email $EMAIL \
    --agree-tos \
    --no-eff-email \
    --force-renewal \
    -d $EXPLORER_DOMAIN

# Get SSL certificate for RPC domain
echo ""
echo "=========================================="
echo "Requesting SSL certificate for $RPC_DOMAIN..."
echo "=========================================="
docker run --rm \
    -v $(pwd)/certbot/conf:/etc/letsencrypt \
    -v $(pwd)/certbot/www:/var/www/certbot \
    certbot/certbot certonly \
    --webroot \
    --webroot-path=/var/www/certbot \
    --email $EMAIL \
    --agree-tos \
    --no-eff-email \
    --force-renewal \
    -d $RPC_DOMAIN

# Stop temporary nginx
echo ""
echo "Stopping temporary nginx..."
docker stop nginx-temp
docker rm nginx-temp

# Check if certificates were created
EXPLORER_CERT_OK=false
RPC_CERT_OK=false

if [ -f "certbot/conf/live/$EXPLORER_DOMAIN/fullchain.pem" ]; then
    echo "✓ Explorer SSL certificate obtained successfully"
    EXPLORER_CERT_OK=true
else
    echo "✗ Failed to obtain Explorer SSL certificate"
fi

if [ -f "certbot/conf/live/$RPC_DOMAIN/fullchain.pem" ]; then
    echo "✓ RPC SSL certificate obtained successfully"
    RPC_CERT_OK=true
else
    echo "✗ Failed to obtain RPC SSL certificate"
fi

echo ""
echo "=========================================="
echo "SSL Setup Summary"
echo "=========================================="

if [ "$EXPLORER_CERT_OK" = true ] && [ "$RPC_CERT_OK" = true ]; then
    echo "✓ All SSL certificates successfully obtained!"
    echo ""
    echo "Certificate locations:"
    echo "  Explorer: certbot/conf/live/$EXPLORER_DOMAIN/"
    echo "  RPC: certbot/conf/live/$RPC_DOMAIN/"
    echo ""
    echo "Next steps:"
    echo "1. Make sure your docker-compose.yml includes the nginx and certbot services"
    echo "2. Make sure nginx.conf is configured for both domains"
    echo "3. Run: docker-compose up -d"
    echo ""
    echo "Your services will be available at:"
    echo "  Explorer: https://$EXPLORER_DOMAIN"
    echo "  RPC: https://$RPC_DOMAIN"
    echo ""
elif [ "$EXPLORER_CERT_OK" = true ] || [ "$RPC_CERT_OK" = true ]; then
    echo "⚠ Partial success - Some certificates failed"
    echo ""
    echo "Please check DNS configuration and try again for failed domains"
    exit 1
else
    echo "✗ Failed to obtain SSL certificates"
    echo ""
    echo "Please check:"
    echo "1. Domain DNS is correctly pointing to this server ($CURRENT_IP)"
    echo "2. Port 80 is open and accessible from the internet"
    echo "3. No firewall is blocking Let's Encrypt validation"
    echo ""
    echo "Troubleshooting commands:"
    echo "  dig $EXPLORER_DOMAIN"
    echo "  dig $RPC_DOMAIN"
    echo "  curl http://$EXPLORER_DOMAIN"
    echo "  curl http://$RPC_DOMAIN"
    echo ""
    exit 1
fi