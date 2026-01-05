<h1 align="center">
  <br>
  <img src="https://svg-banners.vercel.app/api?type=glitch&text1=Neuron&width=800&height=200" alt="Neuron">
  <br>
  Neuron
  <br>
</h1>

<h4 align="center">A blazing-fast, secure code execution engine for modern applications</h4>

<p align="center">
  <a href="#key-features">Key Features</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#api-reference">API Reference</a> •
  <a href="#pricing">Pricing</a> •
  <a href="#documentation">Documentation</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/languages-4+-blue?style=for-the-badge" alt="Languages">
  <img src="https://img.shields.io/badge/latency-200--300ms-green?style=for-the-badge" alt="Latency">
  <img src="https://img.shields.io/badge/sandbox-Docker-2496ED?style=for-the-badge&logo=docker" alt="Docker">
  <img src="https://img.shields.io/badge/status-production%20ready-success?style=for-the-badge" alt="Status">
</p>

---

## 🎯 What is Neuron?

Neuron is a **production-grade code execution platform** that enables you to run untrusted code securely at scale. Built for EdTech platforms, coding bootcamps, developer tools, and technical assessment systems.

**Why Neuron?**
- ⚡ **200-300ms average execution time** - Pre-warmed container pools eliminate cold starts
- 🔒 **Enterprise-grade security** - Docker isolation, network restrictions, resource limits
- 🌐 **Multi-language** - Python, JavaScript, Java, C++ (more coming soon)
- 🚀 **Simple integration** - REST API for easy integration

---

## ✨ Key Features

### 🔐 Security First

```
┌─────────────────────────────────────┐
│  Untrusted Code                     │
│  ↓                                  │
│  ✓ Static analysis & validation    │
│  ✓ Sandboxed Docker containers     │
│  ✓ Network isolation (no internet) │
│  ✓ Read-only filesystem            │
│  ✓ CPU & memory limits             │
│  ✓ Execution timeout (3s)          │
└─────────────────────────────────────┘
```

**Security Layers:**
- **Code Validation** - Blocks dangerous APIs (file I/O, network, process execution)
- **Container Isolation** - Each execution runs in an isolated Docker environment
- **Resource Limits** - Prevents resource exhaustion attacks
- **Automatic Cleanup** - Containers are destroyed or reset after execution

### ⚡ Performance Optimized

```
Traditional Approach:          Neuron Approach:
┌──────────────────┐          ┌──────────────────┐
│ Create Container │ 2000ms   │ Get from Pool    │ 5ms
│ Install Runtime  │ 1500ms   │ Execute Code     │ 250ms
│ Execute Code     │  250ms   │ Return to Pool   │ 2ms
│ Cleanup          │  500ms   │                  │
└──────────────────┘          └──────────────────┘
Total: ~4250ms                Total: ~257ms
```

**Performance Features:**
- **Container Pooling** - Pre-warmed containers ready to execute
- **Intelligent Scaling** - Auto-scale from 1 to N containers per language
- **Queue Management** - Redis/Kafka-powered job distribution
- **Outlier Filtering** - Accurate performance metrics with IQR-based filtering

### 📊 Built-in Analytics

Track execution metrics, performance trends, and user activity:

- **Real-time Stats** - Execution counts, success rates, response times
- **Language Analytics** - Most-used languages, execution patterns
- **Performance Insights** - Average queue time, execution time (outliers filtered)
- **User Dashboards** - Credit usage, execution history, API logs

### 🌐 Multi-Language Support

| Language | Version | Avg. Execution | Status |
|----------|---------|----------------|--------|
| **Python** | 3.12 | 150ms | ✅ Production |
| **JavaScript** | Node 22 | 200ms | ✅ Production |
| **Java** | JDK 21 | 500ms | ✅ Production |
| **C++** | GCC Latest | 280ms | ✅ Production |
| **Go** | - | - | 🚧 Coming Soon |
| **Rust** | - | - | 🔜 Planned |

---

## 🚀 Quick Start

### 1. Get Your API Key


### 2. Submit Your First Code

```bash
curl -X POST https://api.neuron-labs.xyz/api/v1/runner/submit \
  -H "Content-Type: application/json" \
  -H "X-API-KEY: nr_live_1234567890abcdef..." \
  -d '{
    "language": "python",
    "code": "print(\"Hello from Neuron!\")",
    "input": ""
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "jobId": "job_x7k9m2p4",
    "status": "queued"
  }
}
```

### 3. Get Results

```bash
curl https://api.neuron-labs.xyz/api/v1/runner/job_x7k9m2p4/result \
  -H "X-API-KEY: nr_live_1234567890abcdef..."
```

**Response:**
```json
{
  "success": true,
  "data":{
        "executionTimeMs": 328,
        "finishedAt": "2025-12-21T18:46:55.84Z",
        "jobId": "6948409f41a9fe4b844d6608",
        "language": "cpp",
        "queueTimeMs": 4,
        "queuedAt": "2025-12-21T18:46:55.508Z",
        "sandboxErrorMessage": "",
        "sandboxErrorType": null,
        "startedAt": "2025-12-21T18:46:55.512Z",
        "status": "success",
        "stderr": "",
        "stdout": "Hello, World!",
        "totalTimeMs": 332
    }
}
```

---

## 📡 API Reference

### Authentication

All API requests require authentication via Bearer token or API key:

```bash
# Option 1: JWT Token
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

# Option 2: API Key
X-API-Key: nr_live_1234567890abcdef...
```

### Core Endpoints

#### `POST /api/v1/runner/submit`
Submit code for execution

**Request:**
```json
{
  "language": "python",      // python | javascript | java | cpp
  "code": "print('Hello')",  // Source code (max 256KB)
  "input": ""                // Optional stdin input
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "jobId": "job_abc123",
    "status": "queued"
  }
}
```

#### `GET /api/v1/runner/:jobId/result`
Get execution results

**Response:**
```json
{
  "success": true,
  "data": {
        "executionTimeMs": 328,
        "finishedAt": "2025-12-21T18:46:55.84Z",
        "jobId": "6948409f41a9fe4b844d6608",
        "language": "cpp",
        "queueTimeMs": 4,
        "queuedAt": "2025-12-21T18:46:55.508Z",
        "sandboxErrorMessage": "",
        "sandboxErrorType": null,
        "startedAt": "2025-12-21T18:46:55.512Z",
        "status": "success",
        "stderr": "",
        "stdout": "Hello, World!",
        "totalTimeMs": 332
    }
}
```

#### `GET /status`
Check system health

**Response:**
```json
{
  "publisher": "up",
  "subscriber": "up",
  "runner": "up",
  "updatedAt": "2025-12-21T18:00:00Z"
}
```

---

## � Pricing

### 🎁 Trial Phase
Get started for free

- **1,000 credits** to explore the platform
- All languages supported
- Full API access
- Community support



---

## 📚 Documentation

- 📖 [**API Documentation**](./README.md) - Complete API reference
- 🤝 [**Contributing**](./CONTRIBUTING.md) - How to contribute
- 🌐 [**Language Support**](./LANGUAGE_SUPPORT.md) - Add new languages
---

## 🏗️ Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTPS
       ▼
┌─────────────────────────────┐
│      API Server (Go)        │
│  • Authentication           │
│  • Code Validation          │
│  • Job Queueing            │
└──────────┬──────────────────┘
           │
           ▼
    ┌──────────────┐
    │ Redis/Kafka  │ (Message Queue)
    └──────┬───────┘
           │
           ▼
┌──────────────────────────────┐
│     Worker (Go)              │
│  • Container Pool Manager    │
│  • Code Execution Engine     │
│  • Result Processing         │
└──────────┬───────────────────┘
           │
           ▼
    ┌──────────────┐
    │   MongoDB    │ (Results & Logs)
    └──────────────┘
```

**Key Components:**
- **API Server** - Handles requests, validates code, manages queue
- **Worker** - Executes code in Docker containers
- **Container Pools** - Pre-warmed containers for each language
- **Message Queue** - Distributes jobs (Redis Streams or Kafka)
- **MongoDB** - Stores jobs, users, analytics

---

## 🔒 Security

Neuron implements defense-in-depth security:

### Code Validation
- Static analysis for dangerous patterns
- Size limits (256KB per submission)
- Blocked APIs: file I/O, network, process execution

### Container Isolation
- Network disabled (`--network=none`)
- Read-only root filesystem
- Temporary writable `/tmp` (64MB limit)
- No privileged access

### Resource Limits
- **CPU**: Shared (Docker host limits)
- **Memory**: 256MB per container
- **Execution Time**: 3 seconds timeout
- **Disk**: Read-only + 64MB temp

### Monitoring
- Real-time health checks
- Automatic container replacement
- Execution logging and audit trails

---

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](./CONTRIBUTING.md) for:

- Development setup
- Code architecture
- Pull request process
- Adding new languages

---

## 📄 License

MIT License - see [LICENSE](./LICENSE) for details.

---

## 🌟 Community & Support

- **Documentation**: [docs.neuron.dev](https://docs.neuron.dev)
- **GitHub Issues**: [Report bugs](https://github.com/anurag-327/neuron/issues)
- **Discord**: [Join community](https://discord.gg/neuron)
- **Email**: support@neuron.dev

---

<div align="center">

**Built with ❤️ for developers, educators, and creators worldwide**

[Website](https://neuron.dev) • [Documentation](https://docs.neuron.dev) • [GitHub](https://github.com/anurag-327/neuron) • [Discord](https://discord.gg/neuron)

⭐ **Star us on GitHub** — it helps!

</div>