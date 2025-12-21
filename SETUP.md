# 🛠️ Local Development Setup

This guide is for developers who want to run Neuron locally for development.

---

## 📋 Prerequisites

### Required Tools

| Tool | Version | Purpose | Download |
|------|---------|---------|----------|
| **Docker** | Latest | Run infrastructure & sandbox | [Get Docker](https://www.docker.com/get-started/) |
| **Go** | 1.22+ | Run API & Worker | [Download Go](https://go.dev/dl/) |
| **MongoDB** | 5.0+ | Database | [Get MongoDB](https://www.mongodb.com/try/download/community) |

> **Note:** MongoDB should be running on port **27017** (default).

---

## 🚀 Quick Start

### 1️⃣ Clone the Repository

```bash
git clone https://github.com/anurag-327/neuron.git
cd neuron
```

---

### 2️⃣ Configure Environment

Create a `.env` file in the project root:

```env
# Server Configuration
PORT=8080

# Database
MONGO_URI=mongodb://localhost:27017
MONGO_DB_NAME=neuron

# Messaging Backend
QUEUE_BACKEND="redis"   # options: redis | kafka

# Redis Configuration (if using Redis)
REDIS_HOST=localhost
REDIS_PORT=6379

# Kafka Configuration (if using Kafka)
KAFKA_BROKER=localhost:9092
KAFKA_TOPIC=neuron-jobs


```

---

### 3️⃣ Choose Your Message Queue Backend

Neuron supports two messaging backends. Choose based on your needs:

#### 🟩 **Redis (Recommended)**

**Best for:**
- Local development
- Single-server deployments
- Real-time execution with minimal latency

**Pros:**
- Extremely fast (microseconds to milliseconds latency)
- Simple setup
- Built-in UI for monitoring

**Start Redis:**
```bash
docker compose --profile redis up -d
```

This starts:
- Redis Stack (with RedisInsight UI at http://localhost:8001)
- Sandbox-ready Docker environment

---

#### 🟦 **Kafka (Optional)**

**Best for:**
- Distributed systems
- High-throughput scenarios
- Multi-datacenter deployments

**Pros:**
- Horizontal scaling
- Partitioned queues
- Message persistence

**Start Kafka:**
```bash
docker compose --profile kafka up -d
```

This starts:
- Zookeeper
- Kafka broker (accessible at localhost:9092)

---

#### 🔀 **Run Both (For Testing)**

```bash
docker compose --profile redis --profile kafka up -d
```

---

### 4️⃣ Verify Infrastructure

Check that all containers are running:

```bash
docker ps
```

You should see containers for:
- `redis` (if using Redis backend)
- `zookeeper` and `kafka` (if using Kafka backend)

---

### 5️⃣ Install Go Dependencies

```bash
go mod download
```

---

### 6️⃣ Install Air (Hot Reload Tool)

Air provides hot reload during development:

```bash
go install github.com/air-verse/air@latest
```

Make sure `$GOPATH/bin` is in your `PATH`.

---

### 7️⃣ Start Development Servers

Open two terminal windows:

**Terminal 1 - API Server:**
```bash
air -c .air.api.toml
```

The API server will start on `http://localhost:8080`

**Terminal 2 - Worker/Consumer:**
```bash
air -c .air.worker.toml
```

The worker will start processing jobs from the queue.

Both services will automatically restart when you modify source files.

---

## 🧪 Testing the Setup

### Submit a Test Job

```bash
curl -X POST http://localhost:8080/api/v1/runner/submit \
  -H "Content-Type: application/json" \
  -d '{
    "code": "print(\"Hello Neuron\")",
    "language": "python",
    "input": ""
  }'
```

**Expected Response:**
```json
{
  "jobId": "abc123xyz",
  "status": "queued"
}
```

### Check Job Status

```bash
curl http://localhost:8080/api/v1/runner/abc123xyz/result
```

**Expected Response:**
```json
{
  "jobId": "abc123xyz",
  "status": "success",
  "output": "Hello Neuron\n",
  "executionTime": 245
}
```

---

## 📁 Project Structure

```
neuron/
├── cmd/
│   ├── api/          # API server entry point
│   └── worker/       # Worker service entry point
├── internal/
│   ├── api/          # HTTP handlers and routes
│   ├── executor/     # Code execution logic
│   ├── queue/        # Message queue abstractions
│   ├── sandbox/      # Docker sandbox management
│   └── store/        # Database operations
├── docker/
│   └── sandbox/      # Docker images for each language
├── .air.api.toml     # Hot reload config for API
├── .air.worker.toml  # Hot reload config for Worker
├── docker-compose.yml
├── go.mod
└── README.md
```

---

#

### Worker Configuration (`.air.worker.toml`)

Similar to API config but builds `./cmd/worker` instead.

---

## 🐳 Docker Compose Profiles

### Available Profiles

| Profile | Services | Use Case |
|---------|----------|----------|
| `redis` | Redis Stack | Development, single-server |
| `kafka` | Zookeeper, Kafka | Production, distributed |

### Useful Commands

```bash
# Start with Redis
docker compose --profile redis up -d

# Start with Kafka
docker compose --profile kafka up -d

# Stop all services
docker compose down

# View logs
docker compose logs -f

# Remove volumes (clean slate)
docker compose down -v
```

---

## 🧹 Cleanup

### Stop All Services

```bash
# Stop Docker services
docker compose down

# Stop API and Worker (Ctrl+C in each terminal)
```

### Remove All Data

```bash
# Remove Docker volumes
docker compose down -v

# Clean MongoDB
mongosh neuron --eval "db.dropDatabase()"
```

---

## 🐛 Debugging

### Check API Server Logs

```bash
# If using Air
# Logs appear in the terminal where you ran air

# If running directly
go run ./cmd/api
```

### Check Worker Logs

```bash
# If using Air
# Logs appear in the terminal where you ran air

# If running directly
go run ./cmd/worker
```

### Monitor Redis Queue

Access RedisInsight UI at http://localhost:8001 to view:
- Queue depth
- Message contents
- Connection status

### Monitor Kafka

Use Kafka CLI tools:

```bash
# List topics
docker exec -it kafka kafka-topics --list --bootstrap-server localhost:9092

# View messages
docker exec -it kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic neuron-jobs \
  --from-beginning
```



## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

See [CONTRIBUTING.md](./CONTRIBUTING.md) for detailed guidelines.

---

## 📚 Additional Resources

- [API Documentation](https://docs.neuron.dev/api)
- [Architecture Overview](./docs/ARCHITECTURE.md)
- [Language Executors](./docs/EXECUTORS.md)
- [Troubleshooting Guide](./docs/TROUBLESHOOTING.md)

---

## 💬 Support

- **GitHub Issues:** [Report bugs or request features](https://github.com/anurag-327/neuron/issues)
- **Discussions:** [Ask questions](https://github.com/anurag-327/neuron/discussions)
- **Discord:** [Join our community](https://discord.gg/neuron)

---

## 📄 License

MIT License - see [LICENSE](./LICENSE) for details