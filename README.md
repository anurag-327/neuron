<h1 align="center" style="font-weight:700; font-size:42px;">
✨ NEURON ✨
</h1>
<div align="center">

![Languages](https://img.shields.io/badge/languages-C++%20%7C%20Go%20%7C%20Python%20%7C%20JavaScript%20%7C%20Java-blue)
![Sandbox](https://img.shields.io/badge/sandbox-Docker-red)


 <img src="https://svg-banners.vercel.app/api?type=glitch&text1=Neuron&width=900&height=250" />
</p>



**A powerful distributed code execution engine with multi-language support**

[Features](#-features) • [Quick Start](#-quick-start) • [API Reference](#-api-reference)

</div>

---

## 🎯 Features

- **Multi-Language Support** - Execute code in C++, Go, Python, JavaScript, and Java
- **Distributed Architecture** - Kafka-based job queue for scalable processing
- **Isolated Execution** - Docker sandbox environment for secure code running
- **Real-time Status** - Track job execution with instant status updates
- **RESTful API** - Simple HTTP interface for job submission and monitoring

---

## 📋 Prerequisites

### Required Tools

| Tool | Version | Purpose | Download |
|------|---------|---------|----------|
| **Docker** | Latest | Run Kafka & Sandbox | [Get Docker](https://www.docker.com/get-started/) |
| **Go** | 1.22+ | Run API & Worker | [Download Go](https://go.dev/dl/) |
| **MongoDB** | 5.0+ | Database (local) | [Get MongoDB](https://www.mongodb.com/try/download/community) |

> **Note:** MongoDB should be running on port **27017** (default). Update `MONGO_URI` in `.env` if using a different port.

---

# 🚀 Quick Start

## 1️⃣ Clone the Repository

```bash
git clone https://github.com/anurag-327/neuron.git
cd neuron
```

---

## 2️⃣ Configure Environment

Create `.env`:

```env
PORT=8080
MONGO_URI=mongodb://localhost:27017
MONGO_DB_NAME=neuron

# Messaging backend
QUEUE_BACKEND="redis"   # options: redis | kafka

# Kafka specific (if chosen)
KAFKA_BROKER=localhost:9092
```

### 🟩 Recommended: Redis Backend

* Extremely fast
* Near-zero queue latency (microseconds to milliseconds)
* Best for real-time code execution

### 🟦 Optional: Kafka Backend

* Distributed, partitioned queue
* Best for horizontal scaling and large clusters

---

## 3️⃣ Start Infrastructure Services

Neuron uses Docker Compose **profiles** to load only the required messaging backend.

### ✅ **To start Redis backend (recommended)**

```bash
docker compose --profile redis up -d
```

Starts:

* Redis Stack (with UI)
* Sandbox-ready environment

---

### 🟦 **To start Kafka backend**

```bash
docker compose --profile kafka up -d
```

Starts:

* Zookeeper
* Kafka broker

---

### 🔀 **To run both Redis + Kafka (for testing)**

```bash
docker compose --profile redis --profile kafka up -d
```

---

Check containers:

```bash
docker ps
```

---


### 4️⃣ Install Air (Hot Reload Tool)

```bash
go install github.com/air-verse/air@latest
```

### 5️⃣ Start Development Servers

**Terminal 1 - API Server:**
```bash
air -c .air.api.toml
```

**Terminal 2 - Worker/Consumer:**
```bash
air -c .air.worker.toml
```

Both services will automatically restart when you modify source files.

---

## 📡 API Reference

### Base URL
```
http://localhost:8080
```

### Endpoints

#### Submit Code for Execution

```http
POST /api/v1/runner/submit
```

**Request Body:**
```json
{
  "code": "print('Hello Python')",
  "language": "python",
  "input": ""
}
```

**Supported Languages:**
- `python`
- `javascript`
- `java`
- `cpp`
- `go`

**Response:**
```json
{
  "jobId": "12345abc",
  "status": "queued",
}
```

---

#### Check Job Status

```http
GET /api/v1/runner/:jobId/status
```

**Example:**
```bash
curl http://localhost:8080/api/v1/runner/12345abc/status
```

**Response:**
```json
{
  "status": "completed"
}
```

**Status Values:**
- `queued` - Job submitted, waiting for execution
- `running` - Currently executing
- `success` - Execution finished successfully
- `failed` - Execution encountered an error
- `cancelled` - Execution cancelled

---


## 🛠️ Development

### Configuration Files

The project includes pre-configured Air files for hot reload:
- `.air.api.toml` - API server configuration
- `.air.worker.toml` - Worker service configuration

Modify these files to customize build settings and watch patterns.

---



## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

<div align="center">

**Built with ❤️ using Go, Docker, and Kafka**

[Report Bug](https://github.com/anurag-327/neuron/issues) • [Request Feature](https://github.com/anurag-327/neuron/issues)

</div>

