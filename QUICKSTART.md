# 🚀 OneChat - Quickstart Guide

This guide will get you up and running in 15 minutes!

## Prerequisites Checklist

- [ ] Go 1.21+ installed
- [ ] Flutter 3.0+ installed
- [ ] PostgreSQL installed (or Render account)
- [ ] Gemini API Key ([Get one here](https://makersuite.google.com/app/apikey))
- [ ] Cloudinary Account ([Sign up](https://cloudinary.com/users/register/free))

## 🎯 Quick Setup (Local Development)

### Step 1: Backend Setup (5 minutes)

```bash
# Navigate to backend
cd backend

# Install dependencies
go mod download

# Create .env file
cp .env.example .env

# Edit .env with your values
nano .env  # or use your favorite editor
```

**Minimum .env configuration:**
```env
DATABASE_URL=postgres://postgres:password@localhost:5432/onechat?sslmode=disable
JWT_SECRET=my-super-secret-jwt-key-12345
REFRESH_SECRET=my-super-secret-refresh-key-12345
GEMINI_API_KEY=your-gemini-api-key-here
CLOUDINARY_URL=cloudinary://api_key:api_secret@cloud_name
PORT=8080
```

**Start PostgreSQL and create database:**
```bash
# Using psql
psql -U postgres
CREATE DATABASE onechat;
\q

# Or using Docker
docker run --name postgres -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres
docker exec -it postgres psql -U postgres -c "CREATE DATABASE onechat;"
```

**Run backend:**
```bash
go run cmd/server/main.go
```

You should see: `Server starting on port 8080` ✅

### Step 2: Frontend Setup (5 minutes)

```bash
# Navigate to frontend
cd ../frontend

# Install dependencies
flutter pub get

# Update API URLs for local development
# Edit lib/utils/constants.dart and change:
# baseUrl = 'http://localhost:8080'
# wsUrl = 'ws://localhost:8080'

# Run on emulator or device
flutter run
```

The app should launch! ✅

### Step 3: Test It Out (5 minutes)

1. **Register a new account**
   - Open the app
   - Click "Register"
   - Enter phone number (e.g., +1234567890)
   - Enter username
   - Create password
   - Click Register

2. **Test messaging**
   - Register another account (use a different phone/username)
   - Start a chat
   - Send a message
   - See real-time delivery

3. **Test AI features**
   - Tap the AI assistant button
   - Ask a question
   - Get AI-powered response

## 🚀 Production Deployment (Render)

### Backend Deployment

1. **Push code to GitHub**
```bash
git init
git add .
git commit -m "Initial commit"
git branch -M main
git remote add origin https://github.com/yourusername/onechat.git
git push -u origin main
```

2. **Deploy on Render**
   - Go to [Render Dashboard](https://dashboard.render.com/)
   - Click "New +" → "Web Service"
   - Connect your GitHub repository
   - Render auto-detects `render.yaml`
   - Add environment variables:
     - `GEMINI_API_KEY`: Your Gemini API key
     - `CLOUDINARY_URL`: Your Cloudinary URL
   - Click "Create Web Service"

3. **Wait for deployment** (3-5 minutes)
   - Render will:
     - Create PostgreSQL database
     - Build Docker image
     - Deploy backend
     - Run migrations

4. **Get your backend URL**
   - Copy the URL: `https://your-app.onrender.com`

### Frontend Deployment (APK Build)

1. **Update API URLs**
```dart
// lib/utils/constants.dart
static const String baseUrl = 'https://your-app.onrender.com';
static const String wsUrl = 'wss://your-app.onrender.com';
```

2. **Build APK**
```bash
cd frontend
flutter build apk --release
```

3. **Locate APK**
```bash
build/app/outputs/flutter-apk/app-release.apk
```

4. **Distribute**
   - Share APK via Google Drive, Dropbox, etc.
   - Or upload to Play Store for production

## 🧪 Testing

### Test Authentication
```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+1234567890",
    "username": "testuser",
    "password": "password123"
  }'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+1234567890",
    "password": "password123"
  }'
```

### Test AI Features
```bash
# Get your access token from login response
TOKEN="your-access-token-here"

# AI Research
curl -X POST http://localhost:8080/api/v1/ai/research \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "query": "What is quantum computing?"
  }'

# Extract Event
curl -X POST http://localhost:8080/api/v1/ai/extract-event \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "message_text": "Let'\''s meet for coffee at Starbucks on Friday at 3pm"
  }'
```

## 📱 Getting API Keys

### Gemini API Key
1. Go to [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Sign in with Google account
3. Click "Create API Key"
4. Copy the key

### Cloudinary Setup
1. Go to [Cloudinary](https://cloudinary.com/users/register/free)
2. Sign up for free account
3. Go to Dashboard
4. Copy "API Environment variable" (format: `cloudinary://...`)

## 🐛 Troubleshooting

### Backend won't start
- **Database connection failed**: Check PostgreSQL is running
- **Port already in use**: Change PORT in .env
- **Dependencies error**: Run `go mod tidy`

### Frontend build fails
- **Dependencies error**: Run `flutter pub get`
- **Build error**: Run `flutter clean` then rebuild

### API not connecting
- **CORS error**: Check backend CORS settings
- **Connection refused**: Verify backend URL in constants.dart
- **401 Unauthorized**: Check JWT token is valid

### WebSocket not connecting
- **Connection failed**: Verify WebSocket URL uses `ws://` or `wss://`
- **Disconnects immediately**: Check token is being passed correctly

## 📊 Database Schema

Tables are auto-created on first run:
- `users` - User accounts
- `chats` - Private and group chats
- `messages` - All messages
- `groups` - Group information
- `group_members` - Group membership
- `events` - Calendar events
- `media` - Uploaded media tracking
- `message_status` - Delivery/read status

## 🔥 What's Next?

1. **Customize the UI** - Edit Flutter theme in `main.dart`
2. **Add more AI features** - Extend AI service
3. **Implement P2P** - Add Bluetooth/WiFi Direct (Phase 3)
4. **Add notifications** - Integrate FCM
5. **Deploy to Play Store** - Follow Google's guidelines

## 💡 Tips

- **Free tier limits**: Render free tier sleeps after inactivity
- **Database**: Render free PostgreSQL has 1GB limit
- **Media**: Cloudinary free tier has 25GB storage
- **AI**: Gemini API has free quota limits

## 🆘 Need Help?

- Check backend logs: `docker logs <container-id>`
- Check database: `psql -U postgres -d onechat`
- Flutter logs: `flutter logs`
- Backend health: `curl http://localhost:8080/health`

## 🎉 Success!

If you see:
- ✅ Backend running on port 8080
- ✅ Flutter app launched
- ✅ Can register and login
- ✅ Can send messages
- ✅ AI responds to queries

**You're all set! Start building amazing features! 🚀**
