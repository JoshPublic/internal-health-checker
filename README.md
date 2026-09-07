# internal-health-checker

A single Go application scaffold designed to run as multiple cloud-native services by changing environment variables.

## Features

- built-in HTTP server for self-health checks
- configurable service metadata via environment variables
- dependency monitoring via a list of target endpoints
- structured logging that includes the service name and monitored targets
- deployable as a single binary in containers or VM-based workloads

## Environment variables

- `SERVICE_NAME`: name of the service instance
- `SERVICE_VERSION`: build or application version
- `SERVICE_STATUS`: status reported in `/health`
- `PORT`: HTTP port to listen on
- `MONITORED_TARGETS`: comma-separated list of `name=url` entries, for example `alpha=http://host:8080/health,beta=http://host:8081/health`
- `TARGET_0_NAME`, `TARGET_0_URL`, etc.: alternative per-target environment convention
- `CHECK_INTERVAL_SECONDS`: how often to poll the monitored targets (default: `30`)
- `ALERT_MODE`: alert strategy for failed checks (`log` by default; `none` disables alerts; set to `webhook` to publish alerts to `ALERT_WEBHOOK_URL`)
- `ALERT_WEBHOOK_URL`: optional webhook endpoint for alert delivery; for example, a Slack, Teams, or custom HTTP endpoint
- `LOG_LEVEL`: log level

## Alerting via webhook (Future Integration)

For future possible scenario, when `ALERT_MODE=webhook`, the service sends a JSON payload to `ALERT_WEBHOOK_URL` whenever a monitored target reports as `down` or `degraded`.

Example:

```bash
ALERT_MODE=webhook \
ALERT_WEBHOOK_URL='https://hooks.example.com/notify' \
SERVICE_NAME=health-alpha \
SERVICE_VERSION=1.0.0 \
PORT=8080 \
MONITORED_TARGETS='health-beta=http://localhost:8081/health' \
go run .
```

The payload includes the service name, target name, status, URL, and any error text so that a webhook receiver can route or display the alert.

## Endpoints

- `GET /health` returns the service metadata and service status
- `GET /monitored-targets` checks each configured target and returns its availability status
- `GET /ready` same as `/health`

## Local run

To use the sample local configuration file, copy it first:

```bash
cp .env.example .env
```

Then start the app with your loaded environment:

```bash
set -a
source .env
set +a

go run .
```

Or pass the variables inline:

```bash
SERVICE_NAME=health-alpha \
SERVICE_VERSION=1.0.0 \
SERVICE_STATUS=ok \
PORT=8080 \
MONITORED_TARGETS='health-beta=http://localhost:8081/health' \
go run .
```

Then test:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/monitored-targets
```

## Docker Compose

```bash
docker compose up --build
```

This starts two app instances named `health-alpha` and `health-beta`, each monitoring the other over the internal Docker network.

## Open the services in a browser

After the stack is running, open these URLs in your browser:

- http://localhost:8080/monitored-targets
- http://localhost:8081/monitored-targets

You can also view the health responses here:

- http://localhost:8080/health
- http://localhost:8081/health

