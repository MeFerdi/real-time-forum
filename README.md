## Real-Time Forum

This is a Single Page Application(SPA) built to mimick a social media platform. Built using :

**JavaScript & HTML** - Client Side

**Go** - Server Side


For the current setup:
Create .env file at the root directory and add the following content
```sh # Server Configuration
SERVER_ADDRESS=:8080
READ_TIMEOUT=15s
WRITE_TIMEOUT=15s
IDLE_TIMEOUT=60s

# Database Configuration
DATABASE_URL=file:data.db?cache=shared&_fk=1
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=25
DB_CONN_MAX_LIFETIME=5m

# Auth Configuration
SESSION_SECRET=your-strong-secret-key-here
SESSION_TIMEOUT=24h

# WebSocket Configuration
WS_READ_BUFFER=1024
WS_WRITE_BUFFER=1024
WS_PING_INTERVAL=30s

# Environment
ENVIRONMENT=development
```sh
Note: Auth configuration is yet to be implemented. Do not worry about that lol
