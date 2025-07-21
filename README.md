# OpenPipe 🎷

A lightweight service that sniffs messages from a RabbitMQ exchange, extracts telemetry using OpenTelemetry, and exposes metrics for Prometheus scraping.

---

## Features

- 🎯 Listens to a RabbitMQ exchange (configurable)
- ⚙️ Parses out metrics from message content
- 📦 Gathers metrics via OpenTelemetry SDK
- 🔌 Exposes Prometheus-compatible metrics endpoint

---

## Usage

### 📏 Configure `config.yml`
```yaml
Web:
  Host:host
  Port: port

Rabbit:
  Channel: channel
  Port: port
  Host: host
  Username: username
  Password: password 
```

### 🧩 Option 1: Run from GitHub release

1. Download the latest binary from the [Releases](https://github.com/IdanKoblik/OpenPipe/releases) page
2. Make it executable:

```bash
chmod +x openpipe
./openpipe
```

### 🐳 Option 2: Run via Docker
```bash
docker run --rm \
  --name openpipe \
  -p 9464:9464 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  openpipe
```

Then point Prometheus to the `/metrics` endpoint to start scraping.

---

## Why OpenPipe?

Native RabbitMQ monitoring tools offer rich insights—but custom message-based metrics? That's where OpenPipe shines. It handles:

* Custom events inside RabbitMQ payloads
* OTLP-to-Prometheus pipelines
* Isolated, container-friendly deployment

---

## How it works

1. Consumes messages from a RabbitMQ exchange
2. Parses JSON (or other formats) for counts, durations, statuses
3. Records metrics (counters, histograms, gauges) via OTEL SDK
4. Exposes `/metrics` endpoint for Prometheus pull

---

## Contributing

Please open issues or PRs!
We welcome new metric parsers, filters, and OTEL exporter compatibility.

---

## License

MIT © 2025 Idan Koblik


