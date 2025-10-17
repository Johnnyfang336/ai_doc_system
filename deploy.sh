#!/bin/bash

# AI Document System Deployment Script
# For Ubuntu servers

set -e  # Exit on error

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Log functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}


# Check system requirements
check_system() {
    log_info "Checking system requirements..."
    
    # Check operating system
    if [[ ! -f /etc/os-release ]]; then
        log_error "Unable to detect operating system version"
        exit 1
    fi
    
    . /etc/os-release
    if [[ "$ID" != "ubuntu" ]]; then
        log_warning "This script is designed for Ubuntu, current system: $ID"
    fi
    
    log_success "System check completed"
}

# Install Docker and Docker Compose
install_docker() {
    log_info "Checking Docker installation status..."
    
    if command -v docker &> /dev/null; then
        log_success "Docker is installed: $(docker --version)"
    else
        log_info "Installing Docker..."
        
        # Update package index
        sudo apt-get update
        
        # Install necessary packages
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
        
        # Add current user to docker group
        sudo usermod -aG docker $USER
        
        log_success "Docker installation completed"
    fi
    
    # Check Docker Compose
    if command -v docker-compose &> /dev/null; then
        log_success "Docker Compose is installed: $(docker-compose --version)"
    else
        log_info "Installing Docker Compose..."
        
        # Download Docker Compose
        sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
        
        # Add execute permission
        sudo chmod +x /usr/local/bin/docker-compose
        
        log_success "Docker Compose installation completed"
    fi
}



# Setup environment variables
setup_environment() {
    log_info "Setting up environment variables..."
    
    if [[ ! -f .env ]]; then
        log_info "Creating environment variables file..."
        
        # Generate random passwords
        POSTGRES_PASSWORD=$(openssl rand -base64 32)
        JWT_SECRET=$(openssl rand -base64 64)
        
        cat > .env << EOF
# Database configuration
POSTGRES_PASSWORD=$POSTGRES_PASSWORD

# JWT secret key
JWT_SECRET=$JWT_SECRET

# Other configurations
GIN_MODE=release
EOF
        
        log_success "Environment variables file creation completed"
    else
        log_success "Environment variables file already exists"
    fi
}

# Setup SSL certificates
setup_ssl() {
    log_info "Setting up SSL certificates..."
    
    mkdir -p ssl
    
    if [[ ! -f ssl/cert.pem ]] || [[ ! -f ssl/key.pem ]]; then
        log_warning "SSL certificates do not exist, generating self-signed certificates..."
        
        # Generate self-signed certificates (for testing only)
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout ssl/key.pem \
            -out ssl/cert.pem \
            -subj "/C=CN/ST=State/L=City/O=Organization/CN=localhost"
        
        log_warning "Self-signed certificates generated, please use official certificates in production environment"
    else
        log_success "SSL certificates already exist"
    fi
}

# Build and start services
deploy_services() {
    log_info "Building and starting services..."
    
    # Stop existing services
    docker-compose down 2>/dev/null || true
    
    # Build images
    log_info "Building Docker images..."
    docker-compose build --no-cache
    
    # Start services
    log_info "Starting services..."
    docker-compose up -d
    
    # Wait for services to start
    log_info "Waiting for services to start..."
    sleep 30
    
    # Check service status
    docker-compose ps
    
    log_success "Service deployment completed"
}

# Check service health status
check_health() {
    log_info "Checking service health status..."
    
    # Check backend API
    if curl -f http://localhost:8080/api/health &>/dev/null; then
        log_success "Backend API service is normal"
    else
        log_error "Backend API service is abnormal"
    fi
    
    # Check frontend
    if curl -f http://localhost:80 &>/dev/null; then
        log_success "Frontend service is normal"
    else
        log_error "Frontend service is abnormal"
    fi
    
    # Check database
    if docker-compose exec -T postgres pg_isready -U postgres &>/dev/null; then
        log_success "Database service is normal"
    else
        log_error "Database service is abnormal"
    fi
}

# Show deployment information
show_info() {
    log_success "Deployment completed!"
    echo
    echo "Service access addresses:"
    echo "  Frontend: http://$(hostname -I | awk '{print $1}'):80"
    echo "  Backend API: http://$(hostname -I | awk '{print $1}'):8080"
    echo "  HTTPS: https://$(hostname -I | awk '{print $1}'):443 (requires nginx profile configuration)"
    echo
    echo "Management commands:"
    echo "  View logs: docker-compose logs -f"
    echo "  Restart services: docker-compose restart"
    echo "  Stop services: docker-compose down"
    echo "  Update services: docker-compose pull && docker-compose up -d"
    echo
    echo "Configuration file location: $(pwd)"
}

# Main function
main() {
    log_info "Starting AI Document System deployment..."
    
    check_system
    install_docker
    setup_environment
    setup_ssl
    deploy_services
    check_health
    show_info
    
    log_success "Deployment script execution completed!"
}

# Run main function
main "$@"