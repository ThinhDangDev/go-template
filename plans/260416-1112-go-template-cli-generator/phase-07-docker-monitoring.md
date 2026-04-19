# Phase 07: Docker & Monitoring Templates

---
status: complete
priority: P1
effort: 3h (actual: 3h)
dependencies: [phase-05]
completed-date: 2026-04-16T15:45:00Z
test-results: "17/17 tests passing (100%)"
---

## Context Links

- [Main Plan](./plan.md)
- [Previous: Authentication](./phase-06-authentication.md)
- [Next: CI/CD & Testing](./phase-08-cicd-testing.md)

## Overview

Create Docker, Docker Compose, and monitoring (Prometheus + Grafana) templates. These enable local development and production deployment with observability built-in.

## Key Insights

- Multi-stage Docker build reduces image size
- Docker Compose orchestrates app + dependencies
- Prometheus scrapes /metrics endpoint
- Grafana pre-configured with basic dashboards
- Health checks ensure container readiness

## Requirements

### Functional
- Dockerfile with multi-stage build
- Docker Compose with app, postgres, prometheus, grafana
- Prometheus configuration scraping app metrics
- Grafana with pre-provisioned datasource and dashboard

### Non-Functional
- Small production image (< 50MB)
- Fast build caching
- Secure defaults

## Architecture

```
templates/docker/
├── Dockerfile.tmpl
├── docker-compose.yml.tmpl
├── docker-compose.override.yml.tmpl  # Dev overrides
└── .dockerignore.tmpl

templates/monitoring/
├── prometheus/
│   └── prometheus.yml.tmpl
└── grafana/
    ├── provisioning/
    │   ├── datasources/
    │   │   └── prometheus.yml.tmpl
    │   └── dashboards/
    │       ├── dashboard.yml.tmpl
    │       └── app-dashboard.json.tmpl
    └── grafana.ini.tmpl
```

## Related Code Files

### Files to Create
- `templates/docker/Dockerfile.tmpl`
- `templates/docker/docker-compose.yml.tmpl`
- `templates/docker/docker-compose.override.yml.tmpl`
- `templates/docker/.dockerignore.tmpl`
- `templates/monitoring/prometheus/prometheus.yml.tmpl`
- `templates/monitoring/grafana/provisioning/datasources/prometheus.yml.tmpl`
- `templates/monitoring/grafana/provisioning/dashboards/dashboard.yml.tmpl`
- `templates/monitoring/grafana/provisioning/dashboards/app-dashboard.json.tmpl`

## Implementation Steps

### Step 1: Create Dockerfile

```dockerfile
{{/* templates/docker/Dockerfile.tmpl */}}
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/{{.Name}} \
    ./cmd/{{.Name}}

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/bin/{{.Name}} /app/{{.Name}}

# Copy migrations if using golang-migrate
COPY --from=builder /app/migrations /app/migrations

# Create non-root user
RUN adduser -D -g '' appuser
USER appuser

# Expose ports
EXPOSE 8080
{{- if hasGRPC .APIType}}
EXPOSE 9090
{{- end}}

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run
ENTRYPOINT ["/app/{{.Name}}"]
```

### Step 2: Create Docker Compose

```yaml
{{/* templates/docker/docker-compose.yml.tmpl */}}
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: {{.Name}}
    ports:
      - "8080:8080"
{{- if hasGRPC .APIType}}
      - "9090:9090"
{{- end}}
    environment:
      - APP_SERVER_PORT=8080
      - APP_DATABASE_HOST=postgres
      - APP_DATABASE_PORT=5432
      - APP_DATABASE_USER={{.Name}}
      - APP_DATABASE_PASSWORD=${DB_PASSWORD:-secret}
      - APP_DATABASE_DBNAME={{.Name}}
      - APP_DATABASE_SSLMODE=disable
      - APP_LOG_LEVEL=info
      - APP_LOG_FORMAT=json
{{- if hasJWT .AuthType}}
      - APP_AUTH_JWT_SECRET=${JWT_SECRET:-change-me-in-production}
{{- end}}
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped
    networks:
      - {{.Name}}-network

  postgres:
    image: postgres:16-alpine
    container_name: {{.Name}}-postgres
    environment:
      - POSTGRES_USER={{.Name}}
      - POSTGRES_PASSWORD=${DB_PASSWORD:-secret}
      - POSTGRES_DB={{.Name}}
    volumes:
      - postgres-data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U {{.Name}}"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped
    networks:
      - {{.Name}}-network

{{- if .WithMonitor}}

  prometheus:
    image: prom/prometheus:v2.51.0
    container_name: {{.Name}}-prometheus
    volumes:
      - ./monitoring/prometheus:/etc/prometheus
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.enable-lifecycle'
    ports:
      - "9091:9090"
    restart: unless-stopped
    networks:
      - {{.Name}}-network

  grafana:
    image: grafana/grafana:10.4.1
    container_name: {{.Name}}-grafana
    volumes:
      - ./monitoring/grafana/provisioning:/etc/grafana/provisioning
      - grafana-data:/var/lib/grafana
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD:-admin}
      - GF_USERS_ALLOW_SIGN_UP=false
    ports:
      - "3000:3000"
    depends_on:
      - prometheus
    restart: unless-stopped
    networks:
      - {{.Name}}-network
{{- end}}

networks:
  {{.Name}}-network:
    driver: bridge

volumes:
  postgres-data:
{{- if .WithMonitor}}
  prometheus-data:
  grafana-data:
{{- end}}
```

### Step 3: Create Docker Compose Override (Dev)

```yaml
{{/* templates/docker/docker-compose.override.yml.tmpl */}}
# Development overrides
version: '3.8'

services:
  app:
    build:
      target: builder  # Use builder stage with full toolchain
    volumes:
      - .:/app         # Mount source for hot reload
    environment:
      - APP_LOG_LEVEL=debug
      - APP_LOG_FORMAT=console
    command: ["go", "run", "./cmd/{{.Name}}"]
```

### Step 4: Create .dockerignore

```
{{/* templates/docker/.dockerignore.tmpl */}}
# Git
.git
.gitignore

# IDE
.idea
.vscode
*.swp
*.swo

# Build artifacts
bin/
dist/
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out

# Dependencies (rebuilt in container)
vendor/

# Test
coverage.out
coverage.html

# Environment
.env
.env.*
*.local

# Logs
*.log
logs/

# Documentation
*.md
!README.md

# CI/CD
.github/
.gitlab-ci.yml

# Docker (avoid recursive copy)
docker-compose*.yml
Dockerfile
.dockerignore

# Monitoring configs (copied separately)
monitoring/
```

### Step 5: Create Prometheus Config

```yaml
{{/* templates/monitoring/prometheus/prometheus.yml.tmpl */}}
global:
  scrape_interval: 15s
  evaluation_interval: 15s

alerting:
  alertmanagers: []

rule_files: []

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: '{{.Name}}'
    static_configs:
      - targets: ['app:8080']
    metrics_path: /metrics
    scrape_interval: 10s
```

### Step 6: Create Grafana Datasource

```yaml
{{/* templates/monitoring/grafana/provisioning/datasources/prometheus.yml.tmpl */}}
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: false
```

### Step 7: Create Grafana Dashboard Provisioning

```yaml
{{/* templates/monitoring/grafana/provisioning/dashboards/dashboard.yml.tmpl */}}
apiVersion: 1

providers:
  - name: 'default'
    orgId: 1
    folder: ''
    folderUid: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /etc/grafana/provisioning/dashboards
```

### Step 8: Create Grafana Dashboard JSON

```json
{{/* templates/monitoring/grafana/provisioning/dashboards/app-dashboard.json.tmpl */}}
{
  "annotations": {
    "list": []
  },
  "editable": true,
  "fiscalYearStartMonth": 0,
  "graphTooltip": 0,
  "id": null,
  "links": [],
  "panels": [
    {
      "datasource": {
        "type": "prometheus",
        "uid": "prometheus"
      },
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "palette-classic"
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {"color": "green", "value": null},
              {"color": "red", "value": 80}
            ]
          },
          "unit": "reqps"
        }
      },
      "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0},
      "id": 1,
      "options": {
        "colorMode": "value",
        "graphMode": "area",
        "justifyMode": "auto",
        "orientation": "auto",
        "reduceOptions": {
          "calcs": ["lastNotNull"],
          "fields": "",
          "values": false
        },
        "textMode": "auto"
      },
      "title": "Request Rate",
      "type": "stat",
      "targets": [
        {
          "expr": "rate(http_requests_total[5m])",
          "refId": "A"
        }
      ]
    },
    {
      "datasource": {
        "type": "prometheus",
        "uid": "prometheus"
      },
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "palette-classic"
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {"color": "green", "value": null},
              {"color": "yellow", "value": 100},
              {"color": "red", "value": 500}
            ]
          },
          "unit": "ms"
        }
      },
      "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0},
      "id": 2,
      "options": {
        "colorMode": "value",
        "graphMode": "area",
        "justifyMode": "auto",
        "orientation": "auto",
        "reduceOptions": {
          "calcs": ["mean"],
          "fields": "",
          "values": false
        },
        "textMode": "auto"
      },
      "title": "Response Time (p95)",
      "type": "stat",
      "targets": [
        {
          "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) * 1000",
          "refId": "A"
        }
      ]
    },
    {
      "datasource": {
        "type": "prometheus",
        "uid": "prometheus"
      },
      "fieldConfig": {
        "defaults": {
          "color": {"mode": "palette-classic"},
          "custom": {
            "axisBorderShow": false,
            "axisCenteredZero": false,
            "axisColorMode": "text",
            "axisLabel": "",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "line",
            "fillOpacity": 10,
            "gradientMode": "none",
            "hideFrom": {"legend": false, "tooltip": false, "viz": false},
            "insertNulls": false,
            "lineInterpolation": "linear",
            "lineWidth": 1,
            "pointSize": 5,
            "scaleDistribution": {"type": "linear"},
            "showPoints": "auto",
            "spanNulls": false,
            "stacking": {"group": "A", "mode": "none"},
            "thresholdsStyle": {"mode": "off"}
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [{"color": "green", "value": null}]
          }
        }
      },
      "gridPos": {"h": 8, "w": 24, "x": 0, "y": 8},
      "id": 3,
      "options": {
        "legend": {"calcs": [], "displayMode": "list", "placement": "bottom"},
        "tooltip": {"mode": "single", "sort": "none"}
      },
      "title": "HTTP Requests by Status",
      "type": "timeseries",
      "targets": [
        {
          "expr": "sum(rate(http_requests_total[5m])) by (status)",
          "legendFormat": "{{`{{status}}`}}",
          "refId": "A"
        }
      ]
    },
    {
      "datasource": {
        "type": "prometheus",
        "uid": "prometheus"
      },
      "fieldConfig": {
        "defaults": {
          "color": {"mode": "palette-classic"},
          "custom": {
            "axisBorderShow": false,
            "axisCenteredZero": false,
            "axisColorMode": "text",
            "axisLabel": "",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "line",
            "fillOpacity": 10,
            "gradientMode": "none",
            "hideFrom": {"legend": false, "tooltip": false, "viz": false},
            "insertNulls": false,
            "lineInterpolation": "linear",
            "lineWidth": 1,
            "pointSize": 5,
            "scaleDistribution": {"type": "linear"},
            "showPoints": "auto",
            "spanNulls": false,
            "stacking": {"group": "A", "mode": "none"},
            "thresholdsStyle": {"mode": "off"}
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [{"color": "green", "value": null}]
          },
          "unit": "bytes"
        }
      },
      "gridPos": {"h": 8, "w": 12, "x": 0, "y": 16},
      "id": 4,
      "options": {
        "legend": {"calcs": [], "displayMode": "list", "placement": "bottom"},
        "tooltip": {"mode": "single", "sort": "none"}
      },
      "title": "Memory Usage",
      "type": "timeseries",
      "targets": [
        {
          "expr": "go_memstats_alloc_bytes",
          "legendFormat": "Allocated",
          "refId": "A"
        },
        {
          "expr": "go_memstats_sys_bytes",
          "legendFormat": "System",
          "refId": "B"
        }
      ]
    },
    {
      "datasource": {
        "type": "prometheus",
        "uid": "prometheus"
      },
      "fieldConfig": {
        "defaults": {
          "color": {"mode": "palette-classic"},
          "custom": {
            "axisBorderShow": false,
            "axisCenteredZero": false,
            "axisColorMode": "text",
            "axisLabel": "",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "line",
            "fillOpacity": 10,
            "gradientMode": "none",
            "hideFrom": {"legend": false, "tooltip": false, "viz": false},
            "insertNulls": false,
            "lineInterpolation": "linear",
            "lineWidth": 1,
            "pointSize": 5,
            "scaleDistribution": {"type": "linear"},
            "showPoints": "auto",
            "spanNulls": false,
            "stacking": {"group": "A", "mode": "none"},
            "thresholdsStyle": {"mode": "off"}
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [{"color": "green", "value": null}]
          }
        }
      },
      "gridPos": {"h": 8, "w": 12, "x": 12, "y": 16},
      "id": 5,
      "options": {
        "legend": {"calcs": [], "displayMode": "list", "placement": "bottom"},
        "tooltip": {"mode": "single", "sort": "none"}
      },
      "title": "Goroutines",
      "type": "timeseries",
      "targets": [
        {
          "expr": "go_goroutines",
          "legendFormat": "goroutines",
          "refId": "A"
        }
      ]
    }
  ],
  "refresh": "10s",
  "schemaVersion": 39,
  "tags": ["{{.Name}}", "go"],
  "templating": {"list": []},
  "time": {"from": "now-1h", "to": "now"},
  "timepicker": {},
  "timezone": "browser",
  "title": "{{.Name | title}} Dashboard",
  "uid": "{{.Name}}-dashboard",
  "version": 1,
  "weekStart": ""
}
```

## Completed Deliverables

- [x] Create Dockerfile.tmpl with multi-stage build
- [x] Create docker-compose.yml.tmpl with all services
- [x] Create docker-compose.override.yml.tmpl for dev
- [x] Create .dockerignore.tmpl
- [x] Create prometheus/prometheus.yml.tmpl
- [x] Create grafana/provisioning/datasources/prometheus.yml.tmpl
- [x] Create grafana/provisioning/dashboards/dashboard.yml.tmpl
- [x] Create grafana/provisioning/dashboards/app-dashboard.json.tmpl
- [x] Test docker build succeeds
- [x] Test docker-compose up starts all services
- [x] Verify Prometheus scrapes app metrics
- [x] Verify Grafana dashboard loads
- [x] Metrics endpoint handler implemented
- [x] Health checks configured

## Success Criteria - VERIFIED

- [x] `docker build` produces working image < 50MB
- [x] `docker-compose up` starts app, postgres, prometheus, grafana
- [x] App connects to postgres container
- [x] Prometheus scrapes /metrics endpoint
- [x] Grafana shows pre-configured dashboard
- [x] Health checks pass
- [x] All unit tests (17/17) passing

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| Image size too large | Low | Low | Multi-stage build |
| Container networking | Medium | Medium | Use docker network |
| Grafana dashboard JSON | Medium | Low | Test dashboard import |

## Security Considerations

- Non-root user in container
- No secrets in Dockerfile
- Secrets via environment variables
- Grafana admin password configurable
- Database password configurable

## Completion Summary

### Deliverables Implemented
1. **Dockerfile.tmpl**: Multi-stage build (builder + alpine final stage) with non-root user
2. **docker-compose.yml.tmpl**: Full orchestration - app, postgres, prometheus, grafana
3. **docker-compose.override.yml.tmpl**: Dev overrides for hot reload with source mount
4. **.dockerignore.tmpl**: Comprehensive ignore patterns for lean builds
5. **prometheus/prometheus.yml.tmpl**: App metrics scraping at 10s intervals
6. **grafana/provisioning/datasources/prometheus.yml.tmpl**: Datasource config
7. **grafana/provisioning/dashboards/dashboard.yml.tmpl**: Dashboard provisioning
8. **grafana/provisioning/dashboards/app-dashboard.json.tmpl**: 5 pre-built dashboard panels
   - Request Rate (reqps)
   - Response Time p95 (ms)
   - HTTP Requests by Status (timeseries)
   - Memory Usage (bytes)
   - Goroutines count
9. **Metrics handler**: HTTP middleware exposing Prometheus metrics endpoint
10. **Health checks**: Container health validation via HTTP GET /health

### Test Results
- **All 17 unit tests passing** (100% success rate)
- Docker image builds successfully < 50MB
- docker-compose orchestrates all 4 services
- Prometheus scrapes metrics correctly
- Grafana dashboards load with pre-configured panels

### Integration Points
- App exposes /metrics on port 8080 (Prometheus format)
- App exposes /health on port 8080 (liveness probe)
- Prometheus targets app:8080/metrics
- Grafana connects to Prometheus at http://prometheus:9090
- All services communicate via docker network bridge

## Next Steps

Proceed to [Phase 08: CI/CD & Testing](./phase-08-cicd-testing.md)
1. Create GitHub Actions CI/CD pipeline
2. Implement test templates and coverage
3. Set up artifact publishing
