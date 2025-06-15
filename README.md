# Real-Time Forum

A real-time forum application built with Go and SQLite.

## Project Structure

```
.
├── cmd
│   └── api
│       └── main.go
└── internal
    ├── config
    ├── database
    ├── handlers
    └── models
```

## Features

- User authentication and sessions
- Posts and comments
- Categories and tags
- Likes system
- Private messaging
- Real-time updates

## Database Schema

The application uses SQLite with the following tables:

- `users`: Store user information
- `sessions`: Manage user sessions
- `posts`: Store forum posts
- `comments`: Store post comments
- `categories`: Define post categories
- `post_categories`: Link posts to categories
- `likes`: Track likes on posts and comments
- `messages`: Store private messages between users

## Getting Started

1. Ensure you have Go installed
2. Clone the repository
3. Run the application:
   ```bash
   go run cmd/api/main.go
   ```
4. The server will start on port 8080

## Development

The application is structured following clean architecture principles:

- `cmd/api`: Application entry point
- `internal/config`: Configuration management
- `internal/database`: Database operations
- `internal/handlers`: HTTP request handlers
- `internal/models`: Data models and business logic
