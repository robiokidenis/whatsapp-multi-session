# WhatsApp Multi-Session - Utility Scripts

This directory contains utility scripts for managing and troubleshooting the WhatsApp Multi-Session application.

## Database Troubleshooting

### reset-whatsapp-db.sh
Resets the WhatsApp database to fix foreign key constraint errors.
- Backs up existing database
- Removes problematic database files
- Clears session files
- **Note:** All WhatsApp sessions will need re-authentication after running this script

### fix-docker-permissions.sh
Fixes Docker volume permission issues.
- Sets correct ownership (UID:GID 1001:1001) for container user
- Creates necessary directories
- Ensures write permissions
- **Requires:** sudo for production fixes

## Deployment & Setup

### deploy-production.sh
Production deployment script with multiple options:
- Option 1: Run container as root (simplest, recommended for production)
- Option 2: Fix host permissions (keep non-root container)
- Option 3: Use Docker volumes (most reliable)

### setup-docker-mysql.sh
Interactive setup script for Docker + MySQL configuration:
- Generates secure random passwords
- Configures environment variables
- Sets up phpMyAdmin (optional)
- Sets up Redis (optional)
- Sets up monitoring stack (optional)

## Database Maintenance

### fix-production-api-keys.sh
Fixes duplicate empty API key constraint errors in MySQL database:
- Checks for empty API keys
- Lists affected users
- Fixes empty API keys (sets to NULL)
- Verifies the fix
- **Supports:** Direct MySQL and Docker MySQL

## Testing Scripts

### test_send_api.sh
Test script for the `/api/send` endpoint:
- Tests sending using session ID
- Tests sending using phone number
- Tests sending using JID format
- **Note:** Update TOKEN variable with your auth token

### test_error_handling.sh
Test script for error handling with different HTTP status codes:
- Test non-existent session (404)
- Test disconnected session (503)
- Test already logged-in session (400)
- Test non-logged-in session (400)
- **Note:** Update TOKEN variable with your auth token

## Usage

Most scripts can be run directly:
```bash
./scripts/reset-whatsapp-db.sh
./scripts/deploy-production.sh
```

Some scripts require sudo:
```bash
sudo ./scripts/fix-docker-permissions.sh
```

Interactive scripts will guide you through the options:
```bash
./scripts/setup-docker-mysql.sh
```

## Important Notes

- Always backup your data before running maintenance scripts
- Scripts that modify the database will require re-authentication of sessions
- Keep your credentials secure (stored in `.credentials` file after setup)
- Check script headers for specific requirements and usage instructions
