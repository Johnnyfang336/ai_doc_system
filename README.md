# AI Document System

## Project Structure

```
AI_doc_system/
├── backend/                 # Go backend
│   ├── cmd/                # Application entry point
│   ├── internal/           # Internal packages
│   │   ├── api/           # API handlers and routes
│   │   ├── auth/          # Authentication middleware
│   │   ├── database/      # Database connection and migration
│   │   ├── services/      # Business logic services
│   │   └── utils/         # Utility functions
│   ├── migrations/        # Database migration files
│   ├── storage/          # File storage directory
│   ├── go.mod            # Go module dependencies
│   └── go.sum
├── frontend/              # React frontend
│   ├── public/           # Static resources
│   ├── src/              # Source code
│   │   ├── components/   # React components
│   │   ├── pages/        # Page components
│   │   ├── services/     # API services
│   │   ├── types/        # TypeScript type definitions
│   │   ├── hooks/        # React Hooks
│   │   └── utils/        # Utility functions
│   ├── package.json      # Frontend dependencies
│   └── tsconfig.json     # TypeScript configuration
└── README.md             # Project documentation
```

## Features

### User Management
- User registration and login
- JWT authentication
- User profile management
- Administrator permission control

### File Management
- File upload and download
- File rename and delete
- Storage space limit and usage statistics
- Support for multiple file types

### Friend System
- Send and accept friend requests
- Friend group management
- User search functionality

### Messaging Features
- Private messages between friends
- Message read status
- Chat history management

### File Sharing
- Share files with friends
- Create public sharing links
- Sharing permission management

## Technology Stack

### Backend
- **Go 1.19+** - Main programming language
- **Gin** - Web framework
- **PostgreSQL** - Database
- **JWT** - Authentication
- **bcrypt** - Password encryption

### Frontend
- **React 18** - Frontend framework
- **TypeScript** - Type safety
- **Material-UI** - UI component library
- **React Router** - Route management
- **Axios** - HTTP client

## Quick Start

### Requirements
- Go 1.19+
- Node.js 16+
- PostgreSQL 12+

### Backend Setup

1. Enter backend directory:
```bash
cd backend
```

2. Install dependencies:
```bash
go mod tidy
```

3. Configure environment variables:
```bash
export DATABASE_URL="postgres://username:password@localhost/ai_doc_system?sslmode=disable"
export JWT_SECRET="your-jwt-secret-key"
export PORT="8080"
```

4. Run the application:
```bash
go run cmd/main.go
```

### Frontend Setup

1. Enter frontend directory:
```bash
cd frontend
```

2. Install dependencies:
```bash
npm install
```

3. Start development server:
```bash
npm start
```

The frontend application will start at http://localhost:80, and the backend API at http://localhost:8080

## API Documentation

### Authentication Endpoints
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login

### User Endpoints
- `GET /api/profile` - Get user profile
- `PUT /api/profile` - Update user profile

### File Endpoints
- `POST /api/files/upload` - Upload file
- `GET /api/files` - Get user file list
- `GET /api/files/:id` - Get file information
- `GET /api/files/:id/download` - Download file
- `DELETE /api/files/:id` - Delete file
- `PUT /api/files/:id/rename` - Rename file

### Friend Endpoints
- `POST /api/friends/request` - Send friend request
- `POST /api/friends/accept/:id` - Accept friend request
- `POST /api/friends/reject/:id` - Reject friend request
- `GET /api/friends` - Get friend list
- `DELETE /api/friends/:id` - Delete friend

### Message Endpoints
- `POST /api/messages` - Send message
- `GET /api/messages/:friend_id` - Get chat history
- `GET /api/chats` - Get chat list

### File Sharing Endpoints
- `POST /api/shares/friend` - Share file with friend
- `POST /api/shares/public` - Create public share
- `GET /api/shares/with-me` - Get files shared with me
- `GET /api/shares/my-shares` - Get my shared files

## Database Design

The system uses PostgreSQL database with the following main tables:

- `users` - User information
- `files` - File information
- `friendships` - Friend relationships
- `friend_groups` - Friend groups
- `messages` - Message records
- `file_shares` - File sharing records

## Deployment

### Deploy with Docker

1. Build backend image:
```bash
cd backend
docker build -t ai-doc-system-backend .
```

2. Build frontend image:
```bash
cd frontend
docker build -t ai-doc-system-frontend .
```

3. Start with docker-compose:
```bash
docker-compose up -d
```

## Contributing

1. Fork the project
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.


