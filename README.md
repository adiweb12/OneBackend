# OneChat - AI-Powered Messaging Platform

A modern, full-stack messaging application with AI capabilities, group chat, event scheduling, and offline P2P communication.

## 🌟 Features

### Phase 1 (MVP) - Implemented
- ✅ Phone-based authentication with JWT
- ✅ 1-to-1 private messaging
- ✅ Real-time WebSocket communication
- ✅ Group chat with admin controls (up to 256 members)
- ✅ Media upload with Cloudinary (auto-delete after 10 days)
- ✅ Message status tracking (sent/delivered/read)
- ✅ User search and profile management

### Phase 2 (AI Features) - Implemented
- ✅ AI Research Assistant (Gemini API integration)
- ✅ Smart Event Extraction from messages
- ✅ Calendar integration for event scheduling

### Phase 3 (Offline P2P) - Structure Ready
- 🔲 Bluetooth P2P chat
- 🔲 WiFi Direct P2P chat
- 🔲 6-digit pairing system

### Phase 4 (Advanced) - Planned
- 🔲 Status/Stories feature
- 🔲 Advanced search
- 🔲 Push notifications
- 🔲 Backup and restore
- 🔲 Analytics

## 🏗️ Architecture

### Backend (Go)
```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── database/
│   │   └── database.go          # Database initialization
│   ├── models/
│   │   └── models.go            # Data models
│   ├── services/
│   │   ├── auth_service.go      # Authentication logic
│   │   ├── chat_service.go      # Chat operations
│   │   ├── group_service.go     # Group management
│   │   ├── ai_service.go        # Gemini AI integration
│   │   ├── media_service.go     # Cloudinary integration
│   │   └── event_service.go     # Event scheduling
│   ├── handlers/
│   │   ├── auth_handler.go      # Auth HTTP handlers
│   │   ├── chat_handler.go      # Chat HTTP handlers
│   │   ├── group_handler.go     # Group HTTP handlers
│   │   ├── ai_handler.go        # AI HTTP handlers
│   │   ├── media_handler.go     # Media HTTP handlers
│   │   ├── event_handler.go     # Event HTTP handlers
│   │   └── websocket_handler.go # WebSocket handler
│   ├── middleware/
│   │   └── auth.go              # JWT authentication middleware
│   └── websocket/
│       └── hub.go               # WebSocket hub for real-time messaging
├── Dockerfile                    # Docker configuration
├── render.yaml                   # Render deployment config
├── go.mod                        # Go dependencies
└── .env.example                  # Environment variables template
```

### Frontend (Flutter)
```
frontend/
├── lib/
│   ├── main.dart                # App entry point
│   ├── models/
│   │   └── models.dart          # Data models
│   ├── providers/
│   │   ├── auth_provider.dart   # Authentication state
│   │   ├── chat_provider.dart   # Chat state
│   │   ├── websocket_provider.dart
│   │   ├── group_provider.dart
│   │   ├── ai_provider.dart
│   │   └── event_provider.dart
│   ├── services/
│   │   ├── api_service.dart     # HTTP client
│   │   └── storage_service.dart # Local storage
│   ├── screens/
│   │   ├── splash_screen.dart
│   │   ├── auth/
│   │   ├── home/
│   │   ├── chat/
│   │   ├── group/
│   │   ├── ai/
│   │   └── profile/
│   ├── widgets/
│   │   └── ...                  # Reusable widgets
│   └── utils/
│       └── constants.dart        # API endpoints & constants
└── pubspec.yaml                  # Flutter dependencies
```

## 🚀 Setup Instructions

### Prerequisites
- Go 1.21+
- Flutter 3.0+
- PostgreSQL (or use Render's managed database)
- Gemini API Key
- Cloudinary Account

### Backend Setup

1. **Clone and navigate to backend directory**
```bash
cd backend
```

2. **Install dependencies**
```bash
go mod download
```

3. **Set up environment variables**
```bash
cp .env.example .env
# Edit .env with your actual values
```

Required environment variables:
```env
DATABASE_URL=postgres://user:password@localhost:5432/onechat?sslmode=disable
JWT_SECRET=your-super-secret-jwt-key
REFRESH_SECRET=your-super-secret-refresh-key
GEMINI_API_KEY=your-gemini-api-key
CLOUDINARY_URL=cloudinary://api_key:api_secret@cloud_name
PORT=8080
```

4. **Run locally**
```bash
go run cmd/server/main.go
```

### Frontend Setup

1. **Navigate to frontend directory**
```bash
cd frontend
```

2. **Install dependencies**
```bash
flutter pub get
```

3. **Update API URLs**
Edit `lib/utils/constants.dart`:
```dart
static const String baseUrl = 'http://localhost:8080'; // For local testing
static const String wsUrl = 'ws://localhost:8080';     // For local testing
```

4. **Run the app**
```bash
flutter run
```

## 📦 Deployment

### Deploy Backend to Render

1. **Create a new Web Service on Render**
   - Connect your GitHub repository
   - Select the `backend` directory
   - Render will auto-detect the `render.yaml`

2. **Add Environment Variables in Render Dashboard**
   - `GEMINI_API_KEY`: Your Gemini API key
   - `CLOUDINARY_URL`: Your Cloudinary URL
   - Other secrets are auto-generated

3. **Create PostgreSQL Database**
   - Render auto-creates this via `render.yaml`
   - Connection string is automatically set as `DATABASE_URL`

4. **Deploy**
   - Render automatically builds and deploys
   - Your backend will be live at `https://your-app.onrender.com`

### Build Flutter APK

1. **Update API URLs**
```dart
// lib/utils/constants.dart
static const String baseUrl = 'https://your-app.onrender.com';
static const String wsUrl = 'wss://your-app.onrender.com';
```

2. **Build APK**
```bash
flutter build apk --release
```

3. **Find your APK**
```
build/app/outputs/flutter-apk/app-release.apk
```

## 🔑 API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/refresh` - Refresh access token

### Users
- `GET /api/v1/users/me` - Get current user profile
- `PUT /api/v1/users/me` - Update profile
- `GET /api/v1/users/search?q=query` - Search users

### Chats
- `GET /api/v1/chats` - Get all chats
- `POST /api/v1/chats` - Create new chat
- `GET /api/v1/chats/:chatId/messages` - Get messages
- `POST /api/v1/chats/:chatId/messages` - Send message
- `PUT /api/v1/chats/messages/:messageId/status` - Update message status
- `DELETE /api/v1/chats/messages/:messageId` - Delete message

### Groups
- `POST /api/v1/groups` - Create group
- `GET /api/v1/groups/:groupId` - Get group details
- `PUT /api/v1/groups/:groupId` - Update group
- `DELETE /api/v1/groups/:groupId` - Delete group
- `POST /api/v1/groups/:groupId/members` - Add member
- `DELETE /api/v1/groups/:groupId/members/:userId` - Remove member
- `PUT /api/v1/groups/:groupId/members/:userId/role` - Update member role

### AI
- `POST /api/v1/ai/research` - AI research query
- `POST /api/v1/ai/extract-event` - Extract event from text

### Media
- `POST /api/v1/media/upload` - Upload file (multipart/form-data)

### Events
- `GET /api/v1/events` - Get user events
- `POST /api/v1/events` - Create event
- `PUT /api/v1/events/:eventId` - Update event
- `DELETE /api/v1/events/:eventId` - Delete event

### WebSocket
- `GET /ws?token=<jwt_token>` - WebSocket connection

## 🎨 UI/UX Features

- **Material 3 Design** with custom dark theme
- **Gradient accents** for modern feel
- **Smooth animations** with Hero transitions
- **Glass morphism** for AI assistant modal
- **Shimmer loading** effects
- **Floating action buttons** for quick actions
- **Bottom navigation** for easy switching

## 🔐 Security Features

- **JWT Authentication** with access and refresh tokens
- **bcrypt** password hashing
- **CORS** configuration
- **Input validation** on both frontend and backend
- **SQL injection** prevention via GORM
- **XSS** protection

## 📱 Message Types Supported

- Text messages
- Images
- Videos
- Audio
- Documents
- Replies/Quotes

## 🤖 AI Features

### Research Mode
1. User taps 🤖 button
2. Enters query
3. Backend calls Gemini API
4. Returns formatted response
5. User can copy to chat

### Event Extraction
1. Long-press message
2. Select "Schedule Event"
3. AI extracts: title, date, time, location
4. Creates calendar event
5. Saves to database

## 🎯 Next Steps

1. **Test thoroughly** with real users
2. **Add push notifications** (FCM)
3. **Implement offline P2P** (Phases 3)
4. **Add status/stories** feature
5. **Optimize performance**
6. **Add end-to-end encryption**
7. **Publish to Play Store**

## 📄 License

This project is created for educational and portfolio purposes.

## 👨‍💻 Developer

Built by Sachin with ❤️

## 🙏 Acknowledgments

- Anthropic's Claude for architecture guidance
- Flutter community for amazing packages
- Go community for robust libraries
- Gemini API for AI capabilities
- Cloudinary for media management
- Render for easy deployment
