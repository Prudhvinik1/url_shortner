# URL Shortener Service

A production-ready URL shortening web service built with Go, featuring RESTful APIs, PostgreSQL database integration, and comprehensive test coverage.

## Features

- **Cryptographically Secure Short Codes**: 6-character alphanumeric codes using `crypto/rand`
- **RESTful API**: Clean HTTP endpoints for creating and resolving short URLs
- **PostgreSQL Integration**: Reliable persistent storage with connection pooling
- **TTL Support**: Time-to-live configuration for expiring URLs
- **Custom Aliases**: Support for user-defined short codes
- **Multi-User System**: URLs associated with user accounts
- **Comprehensive Testing**: Unit and integration tests with mocking
- **Environment Configuration**: Flexible deployment with environment variables

## Tech Stack

- **Language**: Go 1.25
- **Web Framework**: Gin
- **Database**: PostgreSQL
- **Testing**: Go testing, sqlmock, httptest
- **Configuration**: godotenv

## Prerequisites

- Go 1.25 or higher
- PostgreSQL 12 or higher
- Git

## Database Setup

Create the required database and tables:

```sql
-- Create database
CREATE DATABASE urls;

-- Connect to the database
\c urls

-- Create users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create urls table
CREATE TABLE urls (
    id SERIAL PRIMARY KEY,
    short_code VARCHAR(6) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    is_alias BOOLEAN DEFAULT FALSE,
    ttl BIGINT DEFAULT 0,
    user_id INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create index for faster lookups
CREATE INDEX idx_short_code ON urls(short_code);

-- Insert a default user for testing
INSERT INTO users (name) VALUES ('default_user');
```

## Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd url_shortner
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Configure environment variables**

   Create a `.env` file in the root directory:
   ```bash
   POSTGRES_HOST=localhost
   POSTGRES_PORT=5432
   POSTGRES_USER=postgres
   POSTGRES_PASSWORD=postgres
   POSTGRES_DB=urls
   ```

4. **Set up the database**

   Run the SQL commands from the [Database Setup](#database-setup) section above.

5. **Run the application**
   ```bash
   go run src/*.go
   ```

   The server will start on `http://localhost:8080`

## API Documentation

### Create Short URL

**Endpoint**: `POST /urls`

**Request Body**:
```json
{
  "original_url": "https://example.com/very/long/url",
  "isAlias": false,
  "ttl": 0
}
```

**Response**: `202 Accepted`
```json
"abc123"
```

**Example**:
```bash
curl -X POST http://localhost:8080/urls \
  -H "Content-Type: application/json" \
  -d '{
    "original_url": "https://github.com/your-username",
    "isAlias": false,
    "ttl": 0
  }'
```

### Redirect to Original URL

**Endpoint**: `GET /:short_code`

**Response**: `307 Temporary Redirect`

Redirects to the original URL.

**Example**:
```bash
curl -L http://localhost:8080/abc123
# Redirects to https://github.com/your-username
```

### Error Responses

- `400 Bad Request`: Invalid JSON in request body
- `404 Not Found`: Short code doesn't exist
- `500 Internal Server Error`: Database or server error

## Configuration

The application uses environment variables for configuration:

| Variable | Description | Default |
|----------|-------------|---------|
| `POSTGRES_HOST` | PostgreSQL host | `localhost` |
| `POSTGRES_PORT` | PostgreSQL port | `5432` |
| `POSTGRES_USER` | Database username | (required) |
| `POSTGRES_PASSWORD` | Database password | (required) |
| `POSTGRES_DB` | Database name | `urls` |

## Running Tests

Run all tests:
```bash
go test ./src/...
```

Run tests with coverage:
```bash
go test ./src/... -cover
```

Run tests with verbose output:
```bash
go test ./src/... -v
```

## Project Structure

```
url_shortner/
├── src/
│   ├── main.go           # Entry point, API routes, handlers
│   ├── database.go       # Database operations and initialization
│   ├── shortcode.go      # Short code generation logic
│   ├── database_test.go  # Database layer tests
│   └── handlers_test.go  # HTTP handler tests
├── go.mod                # Go module dependencies
├── go.sum                # Dependency checksums
├── .gitignore           # Git ignore rules
└── README.md            # This file
```

## How It Works

1. **Short Code Generation**:
   - Uses `crypto/rand` for secure random number generation
   - Generates 6-character codes from alphanumeric charset (62 possibilities per character)
   - Total combinations: 62^6 = ~56.8 billion unique codes
   - Includes retry mechanism (up to 5 attempts) for collision handling

2. **URL Creation**:
   - Client sends POST request with original URL
   - Server generates unique short code
   - Stores mapping in PostgreSQL database
   - Returns short code to client

3. **URL Resolution**:
   - Client accesses short URL (GET request)
   - Server looks up original URL in database
   - Redirects client to original URL (HTTP 307)

## Known Limitations

- **TTL Not Enforced**: URLs with TTL are stored but not automatically expired (requires background worker)
- **Custom Aliases**: Field exists but validation/custom code support not implemented
- **User Authentication**: User system exists but no authentication endpoints
- **Rate Limiting**: No rate limiting implemented
- **Analytics**: No click tracking or usage statistics

## Future Enhancements

- [ ] Implement TTL-based URL expiration worker
- [ ] Add custom alias support with validation
- [ ] User authentication and authorization
- [ ] Analytics dashboard (click counts, geographic data)
- [ ] Rate limiting to prevent abuse
- [ ] URL validation and safety checking
- [ ] QR code generation
- [ ] Docker containerization
- [ ] API documentation (Swagger/OpenAPI)
- [ ] Bulk URL shortening
- [ ] URL preview endpoint

## Development

### Adding New Features

1. Create feature branch
2. Implement functionality
3. Write tests
4. Update documentation
5. Submit pull request

### Code Style

- Follow standard Go formatting (`gofmt`)
- Write tests for new functionality
- Use meaningful variable names
- Comment exported functions and types

## License

[Add your license here]

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Contact

[Add your contact information]

## Acknowledgments

- Built with [Gin Web Framework](https://github.com/gin-gonic/gin)
- PostgreSQL driver: [lib/pq](https://github.com/lib/pq)
- Testing: [go-sqlmock](https://github.com/DATA-DOG/go-sqlmock)
