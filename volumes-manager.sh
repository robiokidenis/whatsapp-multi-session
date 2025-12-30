#!/bin/bash

# Docker Volume Management Script
# This script helps manage Docker volumes for WhatsApp Multi-Session

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "\n${BLUE}╔═══════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║  WhatsApp Multi-Session - Volume Manager            ║${NC}"
    echo -e "${BLUE}╚═══════════════════════════════════════════════════════╝${NC}\n"
}

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

# List all volumes
list_volumes() {
    print_info "Docker volumes for WhatsApp Multi-Session:\n"
    docker volume ls | grep -E "whatsapp|mysql" || echo "No volumes found"
}

# Show volume details
show_volume_details() {
    local volume_name=$1

    if [ -z "$volume_name" ]; then
        print_error "Volume name required"
        echo "Usage: $0 details <volume-name>"
        echo "Available volumes:"
        docker volume ls | grep -E "whatsapp|mysql" | awk '{print $2}'
        exit 1
    fi

    print_info "Volume details for: $volume_name\n"
    docker volume inspect $volume_name
}

# Backup a volume
backup_volume() {
    local volume_name=$1
    local backup_file=${2:-"backup_$(date +%Y%m%d_%H%M%S).tar.gz"}

    if [ -z "$volume_name" ]; then
        print_error "Volume name required"
        echo "Usage: $0 backup <volume-name> [backup-file.tar.gz]"
        exit 1
    fi

    print_info "Backing up volume: $volume_name to $backup_file..."

    # Run alpine container to backup volume
    docker run --rm \
        -v $volume_name:/data:ro \
        -v $(pwd):/backup \
        alpine tar czf /backup/$backup_file -C /data .

    if [ -f "$backup_file" ]; then
        local size=$(du -h $backup_file | cut -f1)
        print_success "Backup completed: $backup_file ($size)"
    else
        print_error "Backup failed"
        exit 1
    fi
}

# Restore a volume
restore_volume() {
    local backup_file=$1
    local volume_name=$2

    if [ -z "$backup_file" ] || [ -z "$volume_name" ]; then
        print_error "Backup file and volume name required"
        echo "Usage: $0 restore <backup-file.tar.gz> <volume-name>"
        exit 1
    fi

    if [ ! -f "$backup_file" ]; then
        print_error "Backup file not found: $backup_file"
        exit 1
    fi

    print_warning "This will REPLACE all data in volume: $volume_name"
    read -p "Are you sure? (yes/no): " confirm

    if [ "$confirm" != "yes" ]; then
        print_info "Restore cancelled"
        exit 0
    fi

    print_info "Restoring backup to volume: $volume_name..."

    # Run alpine container to restore volume
    docker run --rm \
        -v $volume_name:/data \
        -v $(pwd):/backup \
        alpine tar xzf /backup/$backup_file -C /data

    print_success "Restore completed"
}

# Delete a volume
delete_volume() {
    local volume_name=$1

    if [ -z "$volume_name" ]; then
        print_error "Volume name required"
        echo "Usage: $0 delete <volume-name>"
        exit 1
    fi

    print_warning "This will PERMANENTLY delete volume: $volume_name"
    read -p "Are you sure? (yes/no): " confirm

    if [ "$confirm" != "yes" ]; then
        print_info "Delete cancelled"
        exit 0
    fi

    print_info "Deleting volume: $volume_name..."
    docker volume rm $volume_name
    print_success "Volume deleted"
}

# Clear all volumes (reset to fresh state)
reset_all_volumes() {
    print_warning "⚠️  DANGER: This will DELETE ALL DATA including:"
    print_warning "  - All WhatsApp sessions"
    print_warning "  - All user accounts"
    print_warning "  - All message logs"
    print_warning "  - All application data"
    echo ""
    read -p "Type 'DELETE ALL' to confirm: " confirm

    if [ "$confirm" != "DELETE ALL" ]; then
        print_info "Reset cancelled"
        exit 0
    fi

    print_info "Stopping containers..."
    docker-compose -f docker-compose.yml -f docker-compose.volumes.yml down

    print_info "Deleting all volumes..."
    docker volume rm whatsapp-database whatsapp-logs whatsapp-data whatsapp-config whatsapp-mysql-data 2>/dev/null || true

    print_success "All volumes deleted"
    print_info "Start fresh with: docker-compose -f docker-compose.yml -f docker-compose.volumes.yml up -d"
}

# Export all volumes
export_all_volumes() {
    local export_dir=${1:-"./backups/$(date +%Y%m%d_%H%M%S)"}

    mkdir -p $export_dir

    print_info "Exporting all volumes to: $export_dir"

    for volume in whatsapp-database whatsapp-logs whatsapp-data whatsapp-config whatsapp-mysql-data; do
        if docker volume inspect $volume >/dev/null 2>&1; then
            print_info "Exporting $volume..."
            docker run --rm \
                -v $volume:/data:ro \
                -v $(pwd)/$export_dir:/backup \
                alpine tar czf /backup/${volume}.tar.gz -C /data .
        fi
    done

    print_success "Export completed to: $export_dir"
}

# Show volume usage
show_usage() {
    print_info "Volume usage:\n"
    docker system df -v | grep -A 3 "VOLUME NAME" | grep -E "whatsapp|mysql|VOLUME NAME"
}

# Main menu
show_menu() {
    print_header
    echo "1) List all volumes"
    echo "2) Show volume details"
    echo "3) Backup a volume"
    echo "4) Restore a volume"
    echo "5) Delete a volume"
    echo "6) Export all volumes"
    echo "7) Show volume usage"
    echo "8) Reset all volumes (⚠️  DELETES EVERYTHING)"
    echo "9) Exit"
    echo ""
}

# Interactive mode
interactive_mode() {
    while true; do
        show_menu
        read -p "Select an option [1-9]: " choice

        case $choice in
            1)
                list_volumes
                ;;
            2)
                read -p "Enter volume name: " vol_name
                show_volume_details $vol_name
                ;;
            3)
                read -p "Enter volume name: " vol_name
                read -p "Enter backup filename (optional): " backup_file
                backup_volume $vol_name $backup_file
                ;;
            4)
                read -p "Enter backup file: " backup_file
                read -p "Enter volume name: " vol_name
                restore_volume $backup_file $vol_name
                ;;
            5)
                read -p "Enter volume name: " vol_name
                delete_volume $vol_name
                ;;
            6)
                read -p "Enter export directory (optional): " export_dir
                export_all_volumes $export_dir
                ;;
            7)
                show_usage
                ;;
            8)
                reset_all_volumes
                ;;
            9)
                print_info "Goodbye!"
                exit 0
                ;;
            *)
                print_error "Invalid option"
                ;;
        esac

        echo ""
        read -p "Press Enter to continue..."
    done
}

# Command line mode
if [ $# -eq 0 ]; then
    interactive_mode
else
    case $1 in
        list)
            list_volumes
            ;;
        details)
            show_volume_details $2
            ;;
        backup)
            backup_volume $2 $3
            ;;
        restore)
            restore_volume $2 $3
            ;;
        delete)
            delete_volume $2
            ;;
        export)
            export_all_volumes $2
            ;;
        usage)
            show_usage
            ;;
        reset)
            reset_all_volumes
            ;;
        *)
            echo "Usage: $0 [list|details|backup|restore|delete|export|usage|reset]"
            echo ""
            echo "Commands:"
            echo "  list                          List all WhatsApp volumes"
            echo "  details <volume-name>         Show volume details"
            echo "  backup <volume-name> [file]   Backup volume to tar.gz"
            echo "  restore <file> <volume-name>  Restore volume from tar.gz"
            echo "  delete <volume-name>          Delete a volume"
            echo "  export [dir]                  Export all volumes"
            echo "  usage                         Show volume usage"
            echo "  reset                         Delete all volumes (⚠️  DANGER)"
            echo ""
            echo "Examples:"
            echo "  $0 list"
            echo "  $0 backup whatsapp-database my-backup.tar.gz"
            echo "  $0 restore my-backup.tar.gz whatsapp-database"
            echo "  $0 details whatsapp-mysql-data"
            exit 1
            ;;
    esac
fi
