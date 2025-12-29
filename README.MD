# URL Shortener API

A RESTful API for shortening URLs, handling redirections, and tracking access statistics.

This is an educational project based on the roadmap.sh backend challenge:
https://roadmap.sh/projects/url-shortening-service

The main goal of this project is to practice backend fundamentals using Go (Golang), MongoDB, and a clean, layered architecture.

---

## Overview

The API allows users to:

- Create short URLs from long URLs
- Redirect users using a short code
- Retrieve URL details
- Track how many times a short URL has been accessed
- Update or delete existing short URLs

The project is structured following Clean Architecture principles to ensure separation of concerns, testability, and maintainability.

---

## Technology Stack

- Go (Golang)
- MongoDB (official Go driver)
- net/http

---

## Project Structure

```
url-shortener/
├── cmd/
│   └── server/
│       └── main.go        # Application entry point
├── internal/
│   ├── config/            # Application and database configuration
│   ├── handlers/          # HTTP handlers
│   ├── models/            # Domain models and entities
│   ├── repository/        # Data access layer (MongoDB)
│   └── utils/             # Utility functions (code generation, helpers)
```

---

## API Endpoints

### Create Short URL

**POST** `/shorten`

Request body:
```json
{
  "url": "https://example.com"
}
```

---

### Redirect to Original URL

**GET** `/{code}`

- Redirects to the original URL
- Increments the access counter

---

### Get Short URL Details

**GET** `/shorten/{code}`

Response example:
```json
{
  "id": 1,
  "url": "https://example.com",
  "short_code": "aB3x9Q",
  "access_count": 12,
  "created_at": "2025-01-01T10:00:00Z",
  "updated_at": "2025-01-01T10:00:00Z"
}
```

---

### Get URL Statistics

**GET** `/shorten/{code}/stats`

Response example:
```json
{
  "short_code": "aB3x9Q",
  "access_count": 12
}
```

---

### Update Short URL

**PUT** `/shorten/{code}`

Request body:
```json
{
  "url": "https://new-url.com"
}
```

---

### Delete Short URL

**DELETE** `/shorten/{code}`

---

## Requirements

- Go 1.21 or higher
- MongoDB running locally on port `27017`

Connection string used by default:
```
mongodb://localhost:27017
```

---

## Installation and Execution

### Clone the repository
```bash
git clone https://github.com/your-username/url-shortener.git
cd url-shortener
```

### Install dependencies
```bash
go mod tidy
```

### Run the application
```bash
go run cmd/server/main.go
```

The API will be available at:
```
http://localhost:8080
```

---



## Notes

- Short codes are generated randomly
- Collisions are handled internally by retrying code generation


---

## License

MIT License
