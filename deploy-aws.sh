#!/bin/bash

# Voltio Services Deployment Script for AWS EC2 Ubuntu
# This script sets up and deploys all Voltio services on a fresh Ubuntu EC2 instance

set -e

echo "🚀 Starting Voltio Services deployment on AWS EC2 Ubuntu..."

# Check if running as root
if [ "$EUID" -eq 0 ]; then
    echo "❌ Please don't run this script as root. Run as ubuntu user."
    exit 1
fi

# Update system
echo "📦 Updating system packages..."
sudo apt update && sudo apt upgrade -y

# Install Docker
echo "🐳 Installing Docker..."
sudo apt install -y apt-transport-https ca-certificates curl software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
sudo apt update
sudo apt install -y docker-ce

# Install Docker Compose
echo "🐙 Installing Docker Compose..."
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Add user to docker group
echo "👤 Adding user to docker group..."
sudo usermod -aG docker $USER

# Install Git
echo "📚 Installing Git..."
sudo apt install -y git

# Create application directory
echo "📁 Creating application directory..."
sudo mkdir -p /opt/voltio
sudo chown $USER:$USER /opt/voltio
cd /opt/voltio

# Clone repository (you'll need to replace with your actual repo)
echo "📥 Cloning repository..."
if [ ! -d "Voltio_CHMMA" ]; then
    git clone https://github.com/M1keTrike/Voltio_CHMMA.git
fi
cd Voltio_CHMMA

# Copy environment file
echo "⚙️ Setting up environment configuration..."
if [ ! -f ".env" ]; then
    cp .env.example .env
    echo "✏️ Please edit /opt/voltio/Voltio_CHMMA/.env with your configuration"
    echo "   You can use: nano /opt/voltio/Voltio_CHMMA/.env"
fi

# Create logs directory
echo "📋 Creating logs directory..."
mkdir -p logs

# Set up firewall
echo "🔥 Configuring firewall..."
sudo ufw allow ssh
sudo ufw allow 8081/tcp
sudo ufw --force enable

# Create systemd service for automatic startup
echo "⚡ Creating systemd service..."
sudo tee /etc/systemd/system/voltio-services.service > /dev/null <<EOF
[Unit]
Description=Voltio Services
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/voltio/Voltio_CHMMA
ExecStart=/usr/local/bin/docker-compose up -d
ExecStop=/usr/local/bin/docker-compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF

# Enable the service
sudo systemctl enable voltio-services.service

echo "✅ Installation completed!"
echo ""
echo "📋 Next steps:"
echo "1. Edit the environment file: nano /opt/voltio/Voltio_CHMMA/.env"
echo "2. Start the services: sudo systemctl start voltio-services"
echo "3. Check status: sudo systemctl status voltio-services"
echo "4. View logs: docker-compose logs -f"
echo ""
echo "🌐 Your services will be available at:"
echo "   - WebSocket Server: http://your-ec2-ip:8081"
echo ""
echo "⚠️ Remember to:"
echo "   - Configure your security groups to allow port 8081"
echo "   - Update your .env file with the correct configuration"
echo "   - Use 'newgrp docker' or logout/login to apply docker group changes"
