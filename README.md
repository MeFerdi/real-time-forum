# 💬 Real-Time Forum

A modern, real-time forum application featuring instant messaging, user authentication, and dynamic content management. Built with Go backend, WebSocket communication, and a responsive JavaScript frontend.

## 🚀 Features

### 🔐 **Authentication & User Management**
- Secure user registration and login system
- Session-based authentication with HTTP-only cookies
- User profiles with personal information
- Secure password hashing with bcrypt

### 📝 **Forum Functionality**
- Create and manage forum posts with categories
- Real-time commenting system
- Like/unlike posts and comments
- Category-based content organization
- User-specific post filtering (My Posts, Liked Posts)

### 💬 **Real-Time Private Messaging**
- **Instant messaging** with WebSocket technology
- **Online/offline status** indicators
- **Message history** with pagination (10 messages at a time)
- **Typing indicators** for active conversations
- **Conversation management** organized by recent activity
- **Unread message counters**
- **Mobile-responsive chat interface**

### 🎨 **Modern UI/UX**
- Single Page Application (SPA) architecture
- Responsive design for all device sizes
- Real-time updates without page refreshes
- Intuitive navigation and user interface
- Connection status indicators

## 🏗️ **Architecture**

### **Backend (Go)**
```
backend/
├── cmd/api/main.go              # Application entry point
├── internal/
│   ├── auth/                    # Authentication middleware & sessions
│   ├── database/                # Database schema & operations
│   ├── handlers/                # HTTP & WebSocket handlers
│   │   ├── user_handler.go      # User authentication endpoints
│   │   ├── post_handler.go      # Forum post endpoints
│   │   ├── message_handler.go   # Private message API
│   │   └── websocket_handler.go # Real-time WebSocket handling
│   └── models/                  # Data models & business logic
│       ├── user.go              # User model & operations
│       ├── post.go              # Post & comment models
│       └── message.go           # Private message models
├── go.mod                       # Go module dependencies
└── go.sum                       # Dependency checksums
```

### **Frontend (JavaScript)**
```
frontend/
├── index.html                   # Single HTML file (SPA)
├── css/styles.css              # Responsive styling
└── js/
    ├── api.js                  # API client for backend communication
    ├── websocket.js            # WebSocket client with reconnection
    ├── chat.js                 # Chat UI and messaging logic
    ├── views.js                # UI management and navigation
    ├── router.js               # Client-side routing
    └── app.js                  # Application initialization
```

## 🗄️ **Database Schema**

**SQLite database with the following tables:**

| Table | Description |
|-------|-------------|
| `users` | User accounts and profile information |
| `sessions` | Authentication sessions with expiration |
| `posts` | Forum posts with titles and content |
| `comments` | Post comments and replies |
| `categories` | Post categorization system |
| `post_categories` | Many-to-many relationship for post categories |
| `likes` | Like tracking for posts and comments |
| `messages` | Private messages between users |

## 🚀 **Getting Started**

### **Prerequisites**
- Go 1.19 or higher
- Modern web browser with WebSocket support

### **Installation & Setup**

1. **Clone the repository**
   ```bash
   git clone https://github.com/MeFerdi/real-time-forum.git
   cd real-time-forum
   ```

2. **Install Go dependencies**
   ```bash
   cd backend
   go mod tidy
   ```

3. **Run the application**
   ```bash
   go run ./cmd/api/main.go
   ```

4. **Access the application**
   - Open your browser and navigate to: `http://localhost:8080`
   - The server will automatically create the SQLite database on first run

### **Usage**

1. **Register** a new account or **login** with existing credentials
2. **Create posts** and engage with the community through comments and likes
3. **Start private conversations** by clicking the Chat button in navigation
4. **Send real-time messages** with instant delivery and read receipts
5. **See who's online** with live status indicators

## 🛠️ **Technology Stack**

### **Backend**
- **Go 1.19+** - High-performance backend language
- **Gorilla WebSocket** - Real-time WebSocket communication
- **SQLite** - Lightweight, embedded database
- **bcrypt** - Secure password hashing
- **UUID** - Session token generation

### **Frontend**
- **Vanilla JavaScript** - No framework dependencies
- **WebSocket API** - Real-time communication
- **CSS Grid & Flexbox** - Modern responsive layouts
- **Font Awesome** - Icon library

### **Key Features**
- **Real-time messaging** with WebSocket technology
- **Automatic reconnection** with exponential backoff
- **Responsive design** for mobile and desktop
- **Session-based authentication** with secure cookies
- **Message pagination** with scroll-to-load functionality
- **Typing indicators** and online status
- **Single Page Application** architecture

## 🔧 **Development**

### **Project Structure**
The application follows clean architecture principles:

- **Separation of concerns** between layers
- **Dependency injection** for testability
- **RESTful API design** with WebSocket enhancement
- **Modular frontend** with clear component separation

### **API Endpoints**

#### **Authentication**
- `POST /api/register` - User registration
- `POST /api/login` - User login
- `POST /api/logout` - User logout
- `GET /api/profile` - Get user profile

#### **Forum**
- `GET /api/posts` - List posts with optional filtering
- `POST /api/posts/create` - Create new post
- `GET /api/posts/get?post_id=X` - Get specific post
- `POST /api/posts/like` - Like/unlike post
- `POST /api/comments/like` - Like/unlike comment

#### **Messaging**
- `GET /api/messages/conversations` - Get user conversations
- `GET /api/messages/history` - Get conversation history
- `POST /api/messages/send` - Send message (HTTP fallback)
- `POST /api/messages/mark-read` - Mark messages as read
- `GET /api/messages/users` - Get all users for chat

#### **WebSocket**
- `WS /ws` - Real-time messaging and status updates

## 📱 **Mobile Support**

The application is fully responsive and optimized for:
- **Desktop browsers** (Chrome, Firefox, Safari, Edge)
- **Mobile devices** (iOS Safari, Android Chrome)
- **Tablet devices** with touch-optimized controls

## 🔒 **Security Features**

- **Secure session management** with HTTP-only cookies
- **Password hashing** with bcrypt
- **SQL injection prevention** with prepared statements
- **XSS protection** with proper input sanitization
- **CSRF protection** with SameSite cookie attributes

## 🚀 **Performance**

- **Efficient WebSocket management** with connection pooling
- **Message pagination** to handle large conversation histories
- **Throttled scrolling** and **debounced search** for smooth UX
- **Optimized database queries** with proper indexing
- **Minimal JavaScript bundle** with no external frameworks

---

**Built with ❤️ for real-time communication and community engagement.**
