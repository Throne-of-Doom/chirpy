# Chirpy

> **Note**: This project was created as part of the [Boot.dev](https://boot.dev) curriculum, following their guided project tutorial.

A social media backend API built with Go that allows users to post short messages (chirps), manage accounts, and interact with content.

## Features

- **User Management**: Register, login, and update user accounts
- **Authentication**: JWT-based authentication with access and refresh tokens
- **Chirps**: Create, retrieve, and delete short messages (140 character limit)
- **Content Moderation**: Automatic profanity filtering
- **Premium Subscriptions**: Webhook integration for upgrading users to Chirpy Red
- **Query & Filtering**: Filter chirps by author and sort by date
- **Admin Panel**: Metrics tracking and database reset functionality

## Tech Stack

- **Language**: Go 1.25.3
- **Database**: PostgreSQL
- **Authentication**: JWT tokens (golang-jwt)
- **Password Hashing**: Argon2id
- **Database Access**: sqlc for type-safe SQL queries
- **Environment Configuration**: godotenv

## API Endpoints

### Health & Admin
- `GET /api/healthz` - Health check endpoint
- `GET /admin/metrics` - View server metrics
- `POST /admin/reset` - Reset database (dev only)

### Authentication
- `POST /api/users` - Register a new user
- `POST /api/login` - Login and receive JWT tokens
- `POST /api/refresh` - Refresh access token
- `POST /api/revoke` - Revoke refresh token
- `PUT /api/users` - Update user information

### Chirps
- `GET /api/chirps` - Get all chirps (supports `?author_id=<uuid>` and `?sort=asc|desc`)
- `POST /api/chirps` - Create a new chirp (requires authentication)
- `GET /api/chirps/{chirpID}` - Get a specific chirp
- `DELETE /api/chirps/{chirpID}` - Delete a chirp (author only)

### Webhooks
- `POST /api/polka/webhooks` - Polka webhook for user upgrades

## Setup

### Prerequisites

- Go 1.25.3 or higher
- PostgreSQL database
- [sqlc](https://sqlc.dev/) for generating database code

### Environment Variables

Create a `.env` file in the project root:

```env
DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
PLATFORM=dev
SECRET=your-jwt-secret-key
POLKA_KEY=your-polka-webhook-api-key
```

### Installation

1. Clone the repository:
```bash
git clone https://github.com/Throne-of-Doom/chirpy.git
cd chirpy
```

2. Install dependencies:
```bash
go mod download
```

3. Run database migrations:
```bash
# Apply schema migrations from sql/schema/
```

4. Generate database code with sqlc:
```bash
sqlc generate
```

5. Run the server:
```bash
go run .
```

The server will start on `http://localhost:8080`.

## Project Structure

```
chirpy/
├── internal/
│   ├── auth/          # Authentication logic (JWT, password hashing)
│   └── database/      # Generated sqlc database code
├── sql/
│   ├── queries/       # SQL queries for sqlc
│   └── schema/        # Database schema migrations
├── assets/            # Static assets
├── handler_*.go       # HTTP request handlers
├── main.go            # Application entry point
├── middleware.go      # HTTP middleware
├── types.go           # Type definitions
└── response_helpers.go # HTTP response utilities
```

## Database Schema

The application uses PostgreSQL with the following main tables:
- **users**: User accounts with authentication details
- **chirps**: Short messages posted by users
- **refresh_tokens**: JWT refresh token management

## Development

### Generating Database Code

After modifying SQL queries or schema:

```bash
sqlc generate
```

### Profanity Filter

The following words are automatically replaced with `****`:
- kerfuffle
- sharbert
- fornax

## License

This project is part of the Boot.dev curriculum.

## Author

Throne-of-Doom
