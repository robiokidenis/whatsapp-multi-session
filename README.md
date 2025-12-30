# WhatsApp Multi-Session Manager
<img width="1512" height="804" alt="image" src="https://github.com/user-attachments/assets/7b08b51b-c666-41a8-a569-ee85aa5ef828" />
n8n node support
<img width="1249" height="768" alt="image" src="https://github.com/user-attachments/assets/9ee66ffd-b890-4fc8-a84e-16e9bf666412" />

A modern WhatsApp multi-session manager built with Go and Vue.js using the whatsmeow library.

## Features

✅ **Multiple WhatsApp Sessions** - Manage multiple WhatsApp accounts simultaneously  
✅ **Real-time QR Code** - WebSocket-powered QR code generation and updates  
✅ **Modern Vue.js Frontend** - Beautiful, responsive web interface  
✅ **RESTful API** - Complete REST API for programmatic access  
✅ **Session Persistence** - SQLite database for session storage  
✅ **Real-time Updates** - Live status updates for all sessions  

## Quick Start

### Option 1: Docker (Recommended)

The easiest way to run the application:

```bash
# Start with Docker
make start

# Or using docker-compose directly
docker-compose up -d

# Access the web interface
open http://localhost:8080
```

**Default credentials:**
- Username: `admin`
- Password: `admin123`

### Option 2: Build from Source

Requires Go 1.24 or higher:

```bash
# Install dependencies
go mod tidy

# Build the application
go build -o whatsapp-multi-session main.go

# Run the application
./whatsapp-multi-session
```

### Common Commands

```bash
make help              # Show all available commands
make start            # Start the application
make stop             # Stop the application
make logs             # Show application logs
make status           # Check application status
make restart          # Restart the application
make volumes-recreate # Reset all volumes (fresh start)
```

### Prerequisites

**For Docker deployment:**
- Docker
- Docker Compose

**For building from source:**
- Go 1.24 or higher
- SQLite3

## Usage

### Web Interface

1. **Add a Session:**
   - Enter phone number in international format (e.g., `6281234567890@s.whatsapp.net`)
   - Optional: Add a friendly name for the session
   - Click "Add Session"

2. **Login to WhatsApp:**
   - Click "Show QR" for your new session
   - Scan the QR code with WhatsApp on your phone
   - The session status will update to "Connected" and "Logged In"

3. **Send Messages:**
   - Click "Send" on any logged-in session
   - Enter recipient number (e.g., `6281234567890@s.whatsapp.net`)
   - Type your message and send

### API Endpoints

#### Sessions Management
- `GET /api/sessions` - List all sessions
- `POST /api/sessions` - Create new session
  ```json
  {
    "phone": "6281234567890@s.whatsapp.net",
    "name": "My WhatsApp"
  }
  ```
- `GET /api/sessions/{id}` - Get session details
- `DELETE /api/sessions/{id}` - Delete session

#### Session Operations
- `POST /api/sessions/{id}/login` - Initiate login
- `POST /api/sessions/{id}/logout` - Logout session
- `GET /api/sessions/{id}/qr` - Get QR code (REST)
- `WS /api/ws/{id}` - Real-time QR codes (WebSocket)

#### Messaging
- `POST /api/sessions/{id}/send` - Send message
  ```json
  {
    "to": "6281234567890@s.whatsapp.net",
    "message": "Hello from WhatsApp Multi-Session!"
  }
  ```

## Project Structure

```
whatsapp-multi-session/
├── main.go                    # Main application entry point
├── docker-compose.yml          # Docker compose configuration
├── Dockerfile                  # Docker image definition
├── Makefile                    # Build and deployment commands
├── go.mod / go.sum             # Go dependencies
├── internal/                   # Application source code
│   ├── handlers/              # HTTP request handlers
│   ├── services/              # Business logic services
│   ├── models/                # Data models
│   ├── repository/            # Database layer
│   └── middleware/            # HTTP middleware
├── frontend/                   # Vue.js web interface
├── scripts/                    # Utility scripts
│   ├── deploy-production.sh   # Production deployment
│   ├── reset-whatsapp-db.sh  # Database reset
│   └── README.md             # Scripts documentation
├── deployment/                 # Docker compose overrides
│   ├── docker-compose.mysql.yml
│   ├── docker-compose.mysql.enhanced.yml
│   └── docker-compose.prod.yml
├── migrations/                 # Database migrations
├── docs/                       # Documentation
├── examples/                   # Example integrations
└── tests/                      # Test files
```

## Technology Stack

### Backend
- **Go** - High-performance backend
- **whatsmeow** - WhatsApp Web Multi-Device API
- **Gorilla Mux** - HTTP router
- **Gorilla WebSocket** - Real-time communication
- **SQLite** - Session persistence

### Frontend
- **Vue.js 3** - Reactive web framework
- **Tailwind CSS** - Modern utility-first CSS
- **Axios** - HTTP client
- **QRCode.js** - QR code generation
- **Font Awesome** - Icons

## Key Features Explained

### Session Manager
- Thread-safe management of multiple WhatsApp clients
- Automatic event handling for connection status
- Persistent storage of session data

### Real-time QR Codes
- WebSocket connections for live QR code updates
- Automatic expiration handling
- Success notifications

### Modern UI
- Responsive design works on all devices
- Real-time status indicators
- Modal dialogs for QR codes and messaging
- Auto-refresh functionality

### Robust API
- RESTful design following best practices
- CORS enabled for cross-origin requests
- Proper error handling and status codes

## Development

### Adding New Features

1. **API Endpoints:** Add new routes in `main.go`
2. **Frontend:** Modify `frontend/index.html`
3. **Database:** Session data is automatically managed

### Building for Production

```bash
# Build optimized binary
go build -ldflags="-s -w" -o whatsapp-multi .

# Run in production
./whatsapp-multi
```

## Troubleshooting

### Common Issues

1. **"No Go files" error:** Make sure you're in the correct directory
2. **SQLite errors:** Ensure write permissions in the directory
3. **WhatsApp connection issues:** Check your internet connection
4. **QR code not showing:** Try refreshing the page or clearing browser cache

### Logs

Check `server.log` for detailed application logs:
```bash
tail -f server.log
```

## License

This project is open source and available under the MIT License.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## Support

For issues and questions:
- Check the troubleshooting section
- Review the server logs
- Open an issue on GitHub

---

Built with ❤️ using Go and Vue.js
