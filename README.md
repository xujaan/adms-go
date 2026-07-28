# ADMS Go

Attendance Device Management System — Go rewrite of Laravel ADMS server.

Receives biometric attendance data from ZKTeco devices via push protocol, stores records, displays web dashboard, and fires configurable webhooks on events.

## Features

- **ZKTeco Push Protocol** — Handshake + receive attendance records (`/iclock/cdata`)
- **Web Dashboard** — Device list, attendance, device log, finger log (Bootstrap 5.3 dark theme)
- **Webhook System** — Configurable per-device webhooks for `attendance`, `device_register`, `device_online` events with 3x retry
- **Single Binary** — ~9.6MB, runs anywhere

## Quick Start

```bash
# Set environment
export DB_ADMS_DSN="root:password@tcp(127.0.0.1:3306)/java_adms?parseTime=true&loc=Asia%2FJakarta"
export PORT=8000

# Run
./server
```

Or with Docker:

```bash
docker compose up -d
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8000` | HTTP listen port |
| `DB_ADMS_DSN` | `root:@tcp(127.0.0.1:3306)/...` | MySQL DSN for java_adms database |

## Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Redirect to /devices |
| GET | `/devices` | Device list with online status |
| GET | `/attendance` | Attendance records |
| GET | `/devices-log` | Raw device handshake logs |
| GET | `/finger-log` | Raw attendance payload logs |
| GET | `/health` | Health check |
| GET | `/iclock/cdata` | Device handshake (ZKTeco protocol) |
| POST | `/iclock/cdata` | Receive attendance records from device |
| GET | `/iclock/test` | Test endpoint |
| GET | `/iclock/getrequest` | Returns OK |
| GET | `/api/webhooks` | List webhooks (?device_sn=X) |
| POST | `/api/webhooks` | Create webhook |
| DELETE | `/api/webhooks/{id}` | Delete webhook |

## Webhooks

### Events

| Event | Trigger |
|-------|---------|
| `attendance` | After attendance record inserted |
| `device_register` | First time device handshakes |
| `device_online` | Every device handshake |

### Payload

```json
{
  "event": "attendance",
  "device_sn": "ABC123",
  "timestamp": "2026-07-28T10:30:00+07:00",
  "data": {
    "employee_id": 123,
    "attendance_timestamp": "2026-07-28 10:30:00",
    "status1": 0,
    "status2": 1
  }
}
```

### Manage via API

```bash
# Create webhook
curl -X POST http://localhost:8000/api/webhooks \
  -H "Content-Type: application/json" \
  -d '{"device_sn":"", "name":"sync-prod", "url":"https://example.com/hook", "event":"attendance"}'

# List webhooks
curl http://localhost:8000/api/webhooks?device_sn=ABC123

# Delete webhook
curl -X DELETE http://localhost:8000/api/webhooks/1
```

`device_sn=""` = global webhook (fires for all devices). `device_sn="ABC123"` = device-specific.

### Retry

Failed webhooks retry 3 times: 1s → 5s → 25s. Logged to stdout.

## Database

Requires MySQL/MariaDB with `java_adms` database. Run `adms.sql` to create tables.

Create webhooks table:

```sql
CREATE TABLE IF NOT EXISTS `webhooks` (
    `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
    `device_sn` varchar(255) NOT NULL DEFAULT '',
    `name` varchar(255) NOT NULL DEFAULT '',
    `url` varchar(1024) NOT NULL,
    `event` varchar(50) NOT NULL,
    `is_active` tinyint(1) NOT NULL DEFAULT 1,
    `created_at` timestamp NULL DEFAULT current_timestamp(),
    `updated_at` timestamp NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
    PRIMARY KEY (`id`),
    INDEX `idx_device_sn` (`device_sn`),
    INDEX `idx_event` (`event`)
);
```

## Build

```bash
go build -ldflags="-s -w" -o server ./cmd/server/
```

## Docker

```bash
docker build -t adms-go .
docker run -p 8000:8000 -e DB_ADMS_DSN="..." adms-go
```

## Project Structure

```
adms-go/
├── cmd/server/main.go           # Entry point + router
├── internal/
│   ├── handler/
│   │   ├── iclock.go            # ZKTeco protocol handlers
│   │   ├── dashboard.go         # Web UI handlers
│   │   └── webhook.go           # Webhook CRUD API
│   ├── iclock/protocol.go       # ZKTeco text protocol parser
│   ├── store/
│   │   ├── db.go                # DB connection
│   │   ├── model.go             # Data structs
│   │   └── query.go             # SQL queries
│   └── webhook/
│       ├── dispatcher.go        # Dispatch to matching webhooks
│       └── sender.go            # HTTP POST + retry
├── templates/                   # Go html/template
├── Dockerfile
├── docker-compose.yml
└── adms.sql                     # Database schema
```

## License

MIT
