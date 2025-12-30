#!/bin/bash

# WhatsApp Multi-Session Docker + MySQL Setup Script
# This script helps you set up the application with MySQL database

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_header() {
    echo -e "\n${BLUE}🚀 WhatsApp Multi-Session Docker + MySQL Setup${NC}\n"
}

# Check if Docker and Docker Compose are installed
check_requirements() {
    print_info "Checking requirements..."
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    print_success "Docker and Docker Compose are installed"
}

# Setup environment file
setup_environment() {
    print_info "Setting up environment configuration..."
    
    if [[ -f ".env" ]]; then
        print_warning ".env file already exists. Creating backup..."
        cp .env .env.backup.$(date +%Y%m%d_%H%M%S)
    fi
    
    # Copy template
    cp .env.docker.mysql .env
    print_success "Environment template copied to .env"
    
    # Generate random passwords
    MYSQL_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)
    MYSQL_ROOT_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)
    JWT_SECRET=$(openssl rand -base64 64 | tr -d "=+/" | cut -c1-50)
    ADMIN_PASSWORD=$(openssl rand -base64 16 | tr -d "=+/" | cut -c1-20)
    
    # Replace placeholders in .env file
    sed -i.bak "s/your_secure_mysql_password_here/${MYSQL_PASSWORD}/g" .env
    sed -i.bak "s/very_secure_root_password_123/${MYSQL_ROOT_PASSWORD}/g" .env
    sed -i.bak "s/your-super-secret-jwt-key-change-this-in-production-mysql-2024/${JWT_SECRET}/g" .env
    sed -i.bak "s/SuperSecureAdminPassword123!/${ADMIN_PASSWORD}/g" .env
    
    # Remove backup file created by sed
    rm .env.bak
    
    print_success "Random secure passwords generated"
    
    # Display credentials
    echo -e "\n${GREEN}🔐 Generated Credentials:${NC}"
    echo -e "Admin Username: ${BLUE}admin${NC}"
    echo -e "Admin Password: ${BLUE}${ADMIN_PASSWORD}${NC}"
    echo -e "MySQL User: ${BLUE}whatsapp_user${NC}"
    echo -e "MySQL Password: ${BLUE}${MYSQL_PASSWORD}${NC}"
    echo -e "MySQL Root Password: ${BLUE}${MYSQL_ROOT_PASSWORD}${NC}\n"
    
    # Save credentials to file
    cat > .credentials << EOF
WhatsApp Multi-Session - Generated Credentials
Generated on: $(date)

Admin Login:
- Username: admin
- Password: ${ADMIN_PASSWORD}
- URL: http://localhost:18080

MySQL Database:
- User: whatsapp_user
- Password: ${MYSQL_PASSWORD}
- Root Password: ${MYSQL_ROOT_PASSWORD}
- Database: whatsapp_multi_session
- Host: localhost
- Port: 3306

phpMyAdmin (if enabled):
- URL: http://localhost:8081
- Username: whatsapp_user
- Password: ${MYSQL_PASSWORD}

JWT Secret: ${JWT_SECRET}
EOF
    
    print_success "Credentials saved to .credentials file"
}

# Ask user for configuration preferences
configure_options() {
    print_info "Configuration Options:"
    
    echo -e "\n1. Database Logging:"
    echo "   - Enabled: Logs stored in MySQL database + web interface"
    echo "   - Disabled: Console logging only (better performance)"
    
    read -p "Enable database logging? (Y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Nn]$ ]]; then
        sed -i.bak "s/ENABLE_DATABASE_LOG=true/ENABLE_DATABASE_LOG=false/g" .env
        rm .env.bak
        print_info "Database logging disabled"
    else
        print_info "Database logging enabled"
    fi
    
    echo -e "\n2. Additional Services:"
    echo "   - phpMyAdmin: Web-based MySQL administration"
    echo "   - Redis: Session storage (optional)"
    echo "   - Monitoring: Prometheus + Grafana"
    
    COMPOSE_PROFILES=""
    
    read -p "Install phpMyAdmin for database management? (Y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        COMPOSE_PROFILES="admin"
        print_info "phpMyAdmin will be installed"
    fi
    
    read -p "Install Redis for session storage? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        COMPOSE_PROFILES="${COMPOSE_PROFILES} redis"
        print_info "Redis will be installed"
    fi
    
    read -p "Install monitoring stack (Prometheus + Grafana)? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        COMPOSE_PROFILES="${COMPOSE_PROFILES} monitoring"
        print_info "Monitoring stack will be installed"
    fi
    
    # Export profiles for docker-compose
    if [[ -n "$COMPOSE_PROFILES" ]]; then
        export COMPOSE_PROFILES=$(echo $COMPOSE_PROFILES | xargs)
        echo "COMPOSE_PROFILES=\"$COMPOSE_PROFILES\"" >> .env
        print_info "Docker Compose profiles: $COMPOSE_PROFILES"
    fi
}

# Start the application
start_application() {
    print_info "Starting WhatsApp Multi-Session with MySQL..."
    
    # Pull latest images
    print_info "Pulling Docker images..."
    docker-compose -f docker-compose.yml -f docker-compose.mysql.enhanced.yml pull
    
    # Start services
    if [[ -n "$COMPOSE_PROFILES" ]]; then
        print_info "Starting with profiles: $COMPOSE_PROFILES"
        docker-compose -f docker-compose.yml -f docker-compose.mysql.enhanced.yml --profile $(echo $COMPOSE_PROFILES | sed 's/ / --profile /g') up -d
    else
        docker-compose -f docker-compose.yml -f docker-compose.mysql.enhanced.yml up -d
    fi
    
    print_success "Services started successfully!"
}

# Wait for services to be ready
wait_for_services() {
    print_info "Waiting for services to be ready..."
    
    # Wait for MySQL
    print_info "Waiting for MySQL to start..."
    until docker-compose exec -T mysql mysqladmin ping -h localhost --silent; do
        sleep 2
        echo -n "."
    done
    echo
    print_success "MySQL is ready"
    
    # Wait for application
    print_info "Waiting for application to start..."
    sleep 10
    
    # Check if application is responding
    for i in {1..30}; do
        if curl -f -s http://localhost:18080/api/health > /dev/null 2>&1; then
            print_success "Application is ready"
            break
        fi
        sleep 2
        echo -n "."
    done
    echo
}

# Display access information
show_access_info() {
    print_success "🎉 Setup completed successfully!"
    
    echo -e "\n${GREEN}📱 Application Access:${NC}"
    echo -e "Main Application: ${BLUE}http://localhost:18080${NC}"
    echo -e "Admin Username: ${BLUE}admin${NC}"
    echo -e "Admin Password: ${BLUE}$(grep 'Password:' .credentials | head -1 | cut -d' ' -f3)${NC}"
    
    if [[ "$COMPOSE_PROFILES" == *"admin"* ]]; then
        echo -e "\n${GREEN}🗄️  Database Management:${NC}"
        echo -e "phpMyAdmin: ${BLUE}http://localhost:8081${NC}"
        echo -e "Username: ${BLUE}whatsapp_user${NC}"
        echo -e "Password: ${BLUE}$(grep 'Password:' .credentials | tail -1 | cut -d' ' -f3)${NC}"
    fi
    
    if [[ "$COMPOSE_PROFILES" == *"monitoring"* ]]; then
        echo -e "\n${GREEN}📊 Monitoring:${NC}"
        echo -e "Grafana: ${BLUE}http://localhost:3000${NC} (admin/admin123)"
        echo -e "Prometheus: ${BLUE}http://localhost:9090${NC}"
    fi
    
    # Check database logging status
    if grep -q "ENABLE_DATABASE_LOG=true" .env; then
        echo -e "\n${GREEN}📋 Database Logging:${NC}"
        echo -e "Log Management: ${BLUE}http://localhost:18080/logs${NC} (admin only)"
        echo -e "Features: Filtering, Search, Auto-refresh, Cleanup"
    else
        echo -e "\n${YELLOW}📋 Console Logging Only:${NC}"
        echo -e "Database logging is disabled for better performance"
        echo -e "Logs are available in: ${BLUE}docker-compose logs whatsapp-multi-session${NC}"
    fi
    
    echo -e "\n${GREEN}🔧 Management Commands:${NC}"
    echo -e "View logs: ${BLUE}docker-compose logs -f whatsapp-multi-session${NC}"
    echo -e "Restart: ${BLUE}docker-compose restart whatsapp-multi-session${NC}"
    echo -e "Stop: ${BLUE}docker-compose down${NC}"
    echo -e "Update: ${BLUE}docker-compose pull && docker-compose up -d${NC}"
    
    echo -e "\n${BLUE}💾 Credentials saved in: .credentials${NC}"
    print_warning "Keep your credentials secure and change them in production!"
}

# Error handling
cleanup_on_error() {
    print_error "Setup failed. Cleaning up..."
    docker-compose -f docker-compose.yml -f docker-compose.mysql.enhanced.yml down
    exit 1
}

# Set trap for error handling
trap cleanup_on_error ERR

# Main execution
main() {
    print_header
    
    check_requirements
    setup_environment
    configure_options
    start_application
    wait_for_services
    show_access_info
    
    print_success "Setup completed! Enjoy using WhatsApp Multi-Session Manager! 🚀"
}

# Run main function
main "$@"