# Chat Service Backend

A backend service for a real-time chat application built with Go.

## Features
- [x] User registration and login
- [x] JWT-based authentication  
- [x] WebSocket support for real-time updates
- [x] Direct Messages (DM)
- [ ] Group Chat
- [ ] Reply to messages
## Prerequisites

- Go 1.24 or higher
- PostgreSQL 16.x
- Redis (for real-time features)
- Nats (for pubsub)

## Configuration

Create a `config.yaml` file in the root directory with the following variables:

```yaml
server:
    port: 8080
    origin: "http://localhost:8080"
postgres:
    host: localhost
    port: 5432
    username: postgres
    password: password
    database: chat_socio
    sslmode: disable
redis:
    host: localhost
    port: 6379
    password: ""
    database: 0
nats:
    address: "nats://localhost:4222"
jwt:
    secret: "ASDWRTGHJKLQWERTYUIOPZXCVBNMASDFGHJKLQWERTYUIOPZXCVBNM"
    issuer: "chat_socio"
    expiration: 3600
```

## Run


## Technology Stack
- Go 1.24
- PostgreSQL 16.x
- Redis 7.x
- Nats 2.9.x
- HTTP Server: Hertz (https://github.com/cloudwego/hertz)
- Database: PostgreSQL
- Cache: Redis
- Pubsub: Nats
- Authentication: JWT