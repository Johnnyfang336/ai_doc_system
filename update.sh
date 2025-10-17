#!/bin/bash

# AI Document System Update Script

set -e

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# Check project directory
check_project_dir() {
    if [[ ! -f docker-compose.yml ]]; then
        log_error "docker-compose.yml file not found, please ensure running this script in project root directory"
        exit 1
    fi
}

# Backup data
backup_data() {
    log_info "Backing up data..."
    
    BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$BACKUP_DIR"
    
    # Backup database
    if docker-compose ps postgres | grep -q "Up"; then
        log_info "Backing up database..."
        docker-compose exec -T postgres pg_dump -U postgres ai_doc_system > "$BACKUP_DIR/database.sql"
        log_success "Database backup completed: $BACKUP_DIR/database.sql"
    fi
    
    # Backup file storage
    if docker volume ls | grep -q "ai_doc_system_backend_storage"; then
        log_info "Backing up file storage..."
        docker run --rm -v ai_doc_system_backend_storage:/source -v "$(pwd)/$BACKUP_DIR":/backup alpine tar czf /backup/storage.tar.gz -C /source .
        log_success "File storage backup completed: $BACKUP_DIR/storage.tar.gz"
    fi
}

# Update services
update_services() {
    log_info "Updating services..."
    
    # Pull latest images
    log_info "Pulling latest images..."
    docker-compose pull
    
    # Rebuild local images
    log_info "Rebuilding images..."
    docker-compose build --no-cache
    
    # Stop services
    log_info "Stopping existing services..."
    docker-compose down
    
    # Start updated services
    log_info "Starting updated services..."
    docker-compose up -d
    
    # Wait for services to start
    log_info "Waiting for services to start..."
    sleep 30
}

# Check update results
check_update() {
    log_info "Checking update results..."
    
    # Show service status
    docker-compose ps
    
    # Check service health status
    if curl -f http://localhost:8080/api/health &>/dev/null; then
        log_success "Backend service updated successfully"
    else
        log_error "Backend service update failed"
        return 1
    fi
    
    if curl -f http://localhost:80 &>/dev/null; then
        log_success "Frontend service updated successfully"
    else
        log_error "Frontend service update failed"
        return 1
    fi
}

# Clean up old images
cleanup() {
    log_info "Cleaning up old images..."
    
    # Remove unused images
    docker image prune -f
    
    # Remove unused containers
    docker container prune -f
    
    log_success "Cleanup completed"
}

# Rollback functionality
rollback() {
    log_warning "Starting rollback..."
    
    # Find latest backup
    LATEST_BACKUP=$(ls -1 backups/ | tail -1)
    
    if [[ -z "$LATEST_BACKUP" ]]; then
        log_error "No backup files found"
        exit 1
    fi
    
    log_info "Using backup: $LATEST_BACKUP"
    
    # Stop services
    docker-compose down
    
    # Restore database
    if [[ -f "backups/$LATEST_BACKUP/database.sql" ]]; then
        log_info "Restoring database..."
        docker-compose up -d postgres
        sleep 10
        docker-compose exec -T postgres psql -U postgres -c "DROP DATABASE IF EXISTS ai_doc_system;"
        docker-compose exec -T postgres psql -U postgres -c "CREATE DATABASE ai_doc_system;"
        docker-compose exec -T postgres psql -U postgres ai_doc_system < "backups/$LATEST_BACKUP/database.sql"
    fi
    
    # Restore file storage
    if [[ -f "backups/$LATEST_BACKUP/storage.tar.gz" ]]; then
        log_info "Restoring file storage..."
        docker run --rm -v ai_doc_system_backend_storage:/target -v "$(pwd)/backups/$LATEST_BACKUP":/backup alpine tar xzf /backup/storage.tar.gz -C /target
    fi
    
    # Start services
    docker-compose up -d
    
    log_success "Rollback completed"
}

# Show help information
show_help() {
    echo "AI Document System Update Script"
    echo
    echo "Usage: $0 [options]"
    echo
    echo "Options:"
    echo "  -h, --help     Show help information"
    echo "  -b, --backup   Backup data only"
    echo "  -r, --rollback Rollback to latest backup"
    echo "  -c, --cleanup  Clean up old images only"
    echo
    echo "Default behavior: Backup data -> Update services -> Check results -> Cleanup"
}

# Main function
main() {
    case "${1:-}" in
        -h|--help)
            show_help
            exit 0
            ;;
        -b|--backup)
            check_project_dir
            backup_data
            exit 0
            ;;
        -r|--rollback)
            check_project_dir
            rollback
            exit 0
            ;;
        -c|--cleanup)
            cleanup
            exit 0
            ;;
        "")
            # Default update process
            check_project_dir
            backup_data
            update_services
            if check_update; then
                cleanup
                log_success "Update completed!"
            else
                log_error "Update failed, you can use $0 --rollback to rollback"
                exit 1
            fi
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
}

main "$@"