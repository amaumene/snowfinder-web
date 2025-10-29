# Snowfinder Web

Web interface for querying Japanese ski resort snow data.

## Overview

The web interface provides an interactive search for finding the snowiest Japanese ski resorts based on historical data. It uses:
- Read-only database access (enforced at compile-time via `Reader` interface)
- Server-side rendering with Go templates
- Client-side sorting and filtering with vanilla JavaScript
- Responsive design with custom CSS

## Prerequisites

- Go 1.24+
- PostgreSQL (populated by scraper)

## Setup

### Using Scraper's Database

The recommended approach is to use the database from the scraper project:

```bash
# No setup needed - just point to scraper's DATABASE_URL
export DATABASE_URL="postgres://snowfinder:snowfinder@localhost:5432/snowfinder?sslmode=disable"
```

### Using Own Database (Optional)

If you want a separate database instance:

```bash
make docker-up
```

## Usage

### Run the Server

```bash
# Run directly
make run

# Or build and run
make build
./bin/snowfinder-web
```

The server starts on http://localhost:8080 (configurable via PORT env var).

### Search for Snowiest Resorts

1. Open http://localhost:8080 in your browser
2. Select date range (e.g., February 8-14)
3. Optionally filter by prefecture
4. Click "Search"
5. Results show average snowfall across all recorded years

### API Endpoint

Query data programmatically:

```bash
# Search by date range
curl "http://localhost:8080/api/search?start_date=02-08&end_date=02-14&limit=10"

# Filter by prefecture
curl "http://localhost:8080/api/search?start_date=02-08&end_date=02-14&prefecture=hokkaido&limit=10"
```

Response format:
```json
[
  {
    "Rank": 1,
    "Name": "Niseko United",
    "Prefecture": "Hokkaido",
    "AvgSnowfall": 145,
    "YearsWithData": 5,
    "TopElevation": 1308,
    "BaseElevation": 308,
    "VerticalDrop": 1000,
    "NumCourses": 38,
    "LongestCourseKM": 5.5
  }
]
```

## Configuration

Configuration via environment variables:

```bash
export DATABASE_URL="postgres://snowfinder:snowfinder@localhost:5432/snowfinder?sslmode=disable"
export PORT=8080
```

## Project Structure

```
snowfinder-web/
├── cmd/server/main.go            # Server entry point
├── internal/
│   └── handlers/                  # HTTP handlers
│       └── handlers.go           # Request handling logic
├── web/
│   ├── static/
│   │   ├── app.js                # Frontend JavaScript
│   │   └── style.css             # Styles
│   └── templates/
│       └── index.html            # HTML template
├── docker-compose.yml            # Optional PostgreSQL
└── Makefile                      # Build commands
```

## Architecture

The web interface uses `snowfinder-common` as a shared Go module containing:
- Data models (Resort, WeeklyResortStats, etc.)
- Repository layer (`Reader` interface for read-only database access)
- Common configuration

The use of the `Reader` interface ensures the web application **cannot** accidentally write to the database - this is enforced at compile-time.

## Development

### Build

```bash
make build
```

### Run Tests

```bash
go test ./...
```

### Clean

```bash
make clean
```

## Features

- Date range search (MM-DD format)
- Prefecture filtering (24 prefectures)
- Sortable results table (click column headers)
- Top 3 resorts highlighted (gold/silver/bronze)
- Responsive design
- No JavaScript frameworks required

