# DisneyTables

DisneyTables is a Go backend service that tracks dining reservation availability at Disneyland Paris. When an open table matching a user's criteria is detected, the service publishes an event to Redis Pub/Sub so that consumer applications—such as Discord bots like [Remy](https://github.com/Romitou/Remy)—can notify guests in real time.

## Background

Booking a table-service restaurant at Disneyland Paris (like Bistrot Chez Rémy or Captain Jack's) can be difficult due to high demand. Slots frequently sell out months ahead, and cancellations appear unpredictably.

Manually refreshing the mobile app throughout the day is tedious. DisneyTables was built to automate this lookup reliably while keeping request rates low and predictable, avoiding unnecessary load on upstream Disney services.

## How it works

1. **Scheduled checks**: A background worker periodically queries the Disneyland Paris API for dates and party sizes requested by active alerts.
2. **Rate throttling**: Requests are distributed across 60-second windows using delayed execution, preventing burst traffic.
3. **Hourly modifiers**: Request limits can be adjusted per hour of the day (e.g. reduced during the night and increased during typical booking release windows).
4. **State persistence**: Restaurants, user alerts, and historical slots are stored in MySQL through GORM.
5. **Notification batching**: When multiple times become available at once for the same alert (e.g. 19:00 and 19:30), they are grouped into a single Redis notification to avoid spam.
6. **Token rotation**: Disney OAuth credentials are automatically refreshed every 6 hours.

## Project layout

```
├── api/             # HTTP and GraphQL client for Disneyland Paris APIs
├── core/            # Business logic and notification batching
├── database/        # GORM connection and MySQL models
│   └── models/      # Alerts, restaurants, slots, notifications, auth
├── redis/           # Redis client and Pub/Sub publisher
├── tasker/          # Cron scheduler and periodic background jobs
│   └── tasks/       # Availability polling, auth refresh, catalog sync
└── webserver/       # HTTP server (Gin) and authentication middleware
```

## Configuration

Configuration is handled through environment variables. A template is provided in `.env.example`:

| Variable | Description |
|---|---|
| `API_KEY` | Disney mobile API key for request headers |
| `AVAILABILITIES_ENDPOINT` | URL of the Disney dining availability endpoint |
| `GRAPHQL_ENDPOINT` | URL of the Disneyland Paris GraphQL gateway |
| `REFRESH_AUTH_ENDPOINT` | URL used to refresh Disney OAuth access tokens |
| `CUSTOM_HEADERS` | JSON map of extra headers to mirror official mobile app requests |
| `RESTAURANTS_QUERY` | GraphQL query string to fetch restaurant details |
| `MYSQL_DSN` | MySQL connection string (e.g. `user:pass@tcp(localhost:3306)/db?...`) |
| `REDIS_HOST` | Redis server address (`host:port`) |
| `REDIS_PASSWORD` | Redis password (leave empty if none) |
| `WEBSERVER_TOKEN` | Bearer token required to call the DisneyTables REST API |
| `MAX_REQUESTS_PER_MINUTE` | Base request rate limit (defaults to `5`) |
| `REQUEST_MODIFIERS` | Optional JSON map of hourly multipliers (e.g. `{"3": 0.2, "11": 1.5}`) |
| `SENTRY_DSN` | Optional Sentry DSN for error monitoring |

## API

All endpoints require a Bearer token:

```
Authorization: Bearer <WEBSERVER_TOKEN>
```

### Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/restaurants` | List all synced restaurants |
| `POST` | `/restaurantAvailabilities` | Live lookup of current dining slots for a given date/party |
| `GET` | `/bookAlerts` | List all active booking alerts |
| `POST` | `/bookAlerts` | Create a new alert |
| `POST` | `/completeBookAlert` | Mark an alert as completed |
| `GET` | `/statistics` | Return active alerts and future slot counts |
| `GET` | `/dailyReport` | Return a 24-hour summary of checks and discoveries |

### Creating an alert

```http
POST /bookAlerts HTTP/1.1
Content-Type: application/json
Authorization: Bearer your-token

{
  "discordId": "123456789012345678",
  "restaurantDisneyId": "P1NA100_bistrot-chez-remy",
  "date": "2024-06-15",
  "mealPeriod": "Dinner",
  "partyMix": 4
}
```

## Setup and development

### Prerequisites

- Go 1.19 or later
- MySQL 8.0+
- Redis 6.0+

### Local run

```bash
# Clone the repository
git clone https://github.com/romitou/disneytables.git
cd disneytables

# Create configuration
cp .env.example .env

# Run tests and linter
make test
make vet

# Start the service
make run
```

### Docker

```bash
# Build image
docker build -t disneytables .

# Run with environment file
docker run -d \
  --name disneytables \
  --restart unless-stopped \
  --env-file .env \
  -p 8080:8080 \
  disneytables
```

## License

This project is licensed under the [MIT License](LICENSE).
