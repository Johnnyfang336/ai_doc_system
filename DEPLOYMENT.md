# AI Document System Docker Deployment Guide

This guide will help you deploy the AI Document System using Docker containers on an Ubuntu server.

## System Requirements

### Hardware Requirements
- **CPU**: 2 cores or more
- **Memory**: 4GB RAM or more
- **Storage**: 20GB available space or more
- **Network**: Stable internet connection

### Software Requirements
- **Operating System**: Ubuntu 18.04 LTS or higher
- **Docker**: 20.10 or higher
- **Docker Compose**: 1.29 or higher

## Quick Deployment

### 1. Prepare Server

```bash
# Update system packages
sudo apt update && sudo apt upgrade -y

# Install necessary tools
sudo apt install -y curl wget git unzip
```

### 2. Upload Project Files

Upload the entire project directory to the server, or clone using git:

```bash
# Method 1: Clone using git (if you have a git repository)
git clone <your-repository-url> /opt/ai-doc-system

# Method 2: Manually upload project files to /opt/ai-doc-system
```

### 3. Run Deployment Script

```bash
cd /opt/ai-doc-system
chmod +x deploy.sh
./deploy.sh
```

The deployment script will automatically complete the following operations:
- Check system environment
- Install Docker and Docker Compose
- Set up project directory
- Configure environment variables
- Generate SSL certificates
- Build and start services

### 4. Verify Deployment

After deployment is complete, visit the following addresses to verify services:

- **Frontend Application**: http://your-server-ip:80
- **Backend API**: http://your-server-ip:8080/api/health
- **HTTPS Access**: https://your-server-ip:443 (requires nginx profile configuration)

## Manual Deployment

If you prefer to manually control the deployment process, follow these steps:

### 1. Install Docker

```bash
# Uninstall old versions
sudo apt-get remove docker docker-engine docker.io containerd runc

# Install dependencies
sudo apt-get update
sudo apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg \
    lsb-release

# Add Docker official GPG key
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# Set up stable repository
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker Engine
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io

# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker
```

### 2. Install Docker Compose

```bash
# Download Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose

# Add execute permission
sudo chmod +x /usr/local/bin/docker-compose

# Verify installation
docker-compose --version
```

### 3. Configure Environment Variables

```bash
# Copy environment variable example file
cp .env.example .env

# Edit environment variables
nano .env
```

**Important**: Please make sure to modify the following key configurations:
- `POSTGRES_PASSWORD`: Set a strong password
- `JWT_SECRET`: Set a randomly generated strong key

### 4. Set up SSL Certificates

```bash
# Create SSL directory
mkdir -p ssl

# Generate self-signed certificate (for testing only)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout ssl/key.pem \
    -out ssl/cert.pem \
    -subj "/C=CN/ST=State/L=City/O=Organization/CN=your-domain.com"
```

**Production Environment Recommendation**: Use Let's Encrypt or purchase official SSL certificates.

### 5. Start Services

```bash
# Build images
docker-compose build

# Start services
docker-compose up -d

# Check service status
docker-compose ps

# View logs
docker-compose logs -f
```

## Service Management

### Common Commands

```bash
# Check service status
docker-compose ps

# View real-time logs
docker-compose logs -f

# Restart all services
docker-compose restart

# Restart single service
docker-compose restart backend

# Stop all services
docker-compose down

# Stop and remove data volumes
docker-compose down -v

# Update services
./update.sh
```

### Service Configuration

#### Development Environment
```bash
# Start development environment (without nginx)
docker-compose up -d postgres backend frontend
```

#### Production Environment
```bash
# Start production environment (with nginx reverse proxy)
docker-compose --profile production up -d
```

## Data Management

### Data Backup

```bash
# Manual backup
./update.sh --backup

# Backup database
docker-compose exec postgres pg_dump -U postgres ai_doc_system > backup.sql

# Backup file storage
docker run --rm -v ai_doc_system_backend_storage:/source -v $(pwd):/backup alpine tar czf /backup/storage_backup.tar.gz -C /source .
```

### Data Recovery

```bash
# Restore database
docker-compose exec -T postgres psql -U postgres ai_doc_system < backup.sql

# Restore file storage
docker run --rm -v ai_doc_system_backend_storage:/target -v $(pwd):/backup alpine tar xzf /backup/storage_backup.tar.gz -C /target
```

### Data Migration

If you need to migrate to a new server:

1. Backup data on the old server
2. Deploy the system on the new server
3. Stop services on the new server
4. Restore backup data
5. Start services

## Monitoring and Maintenance

### Health Check

```bash
# Check service health status
curl http://localhost:8080/api/health
curl http://localhost:80

# Check database connection
docker-compose exec postgres pg_isready -U postgres
```

### Log Management

```bash
# View specific service logs
docker-compose logs backend
docker-compose logs frontend
docker-compose logs postgres

# Clean up logs
docker system prune -f
```

### Performance Monitoring

The system provides basic health check endpoints that you can integrate into monitoring systems:

- Backend health check: `/api/health`
- Database connection check: via docker health check
- Frontend availability check: HTTP status code check

## Security Configuration

### Firewall Settings

```bash
# Install ufw
sudo apt install ufw

# Allow SSH
sudo ufw allow ssh

# Allow HTTP and HTTPS
sudo ufw allow 80
sudo ufw allow 443

# Enable firewall
sudo ufw enable
```

### SSL/TLS Configuration

For production environments, it's recommended to use Let's Encrypt free SSL certificates:

```bash
# Install certbot
sudo apt install certbot

# Obtain certificate
sudo certbot certonly --standalone -d your-domain.com

# Update nginx configuration to use official certificate
# Edit nginx.conf file and update certificate paths
```

## Troubleshooting

### Common Issues

1. **Services fail to start**
   ```bash
   # Check logs
   docker-compose logs
   
   # Check port usage
   sudo netstat -tlnp | grep :80
   sudo netstat -tlnp | grep :8080
   ```

2. **Database connection failure**
   ```bash
   # Check database status
   docker-compose exec postgres pg_isready -U postgres
   
   # Restart database
   docker-compose restart postgres
   ```

3. **File upload failure**
   ```bash
   # Check storage volumes
   docker volume ls
   
   # Check permissions
   docker-compose exec backend ls -la /home/storage
   ```

4. **Out of memory**
   ```bash
   # Check system resources
   free -h
   df -h
   
   # Clean up Docker resources
   docker system prune -a
   ```

### Rollback Operations

If issues occur after an update, you can quickly rollback:

```bash
# Rollback to latest backup
./update.sh --rollback
```

## Updates and Upgrades

### Automatic Update

```bash
# Run update script
./update.sh
```

The update script will automatically:
1. Backup current data
2. Pull latest images
3. Rebuild services
4. Verify update results
5. Clean up old images

### Manual Update

```bash
# Backup data
./update.sh --backup

# Pull latest code
git pull origin main

# Rebuild and start
docker-compose build --no-cache
docker-compose up -d
```

## Contact Support

If you encounter issues during deployment, please:

1. Check log files
2. Check system resources
3. Refer to the troubleshooting section
4. Submit an issue to the project repository

---

**Note**: This deployment guide is for production environments. Please ensure thorough testing before deployment and regularly backup important data.