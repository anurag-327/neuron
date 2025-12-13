<h1 align="center" style="font-weight:700; font-size:42px;">
✨ NEURON ✨
</h1>

<div align="center">

![Languages](https://img.shields.io/badge/languages-C++%20%7C%20Python%20%7C%20JavaScript%20%7C%20Java-blue)
![Sandbox](https://img.shields.io/badge/sandbox-Docker-red)
![Status](https://img.shields.io/badge/status-production%20ready-green)

<img src="https://svg-banners.vercel.app/api?type=glitch&text1=Neuron&width=900&height=250" />

**A secure, scalable code execution platform for EdTech creators, educators, and organizations**

[Features](#-features) • [Use Cases](#-use-cases) • [API Reference](#-api-reference) • [Getting Started](#-getting-started)

</div>

---

## 🎯 What is Neuron?

Neuron is a production-ready code execution engine that lets you add interactive coding capabilities to your platform without building infrastructure from scratch. Perfect for:

- **EdTech Platforms** - Launch coding bootcamps and courses
- **Tech Educators** - Add hands-on exercises to your content
- **Organizations** - Build internal training and assessment tools
- **Developer Tools** - Create REPLs, playgrounds, and testing environments

**Skip months of development and infrastructure costs.** Integrate Neuron through a simple REST API and start accepting code submissions immediately.

---

## ✨ Features

### 🔒 **Secure Execution**
- Docker-isolated sandbox environment
- Resource limits (CPU, memory, execution time)
- Network access restrictions
- Automatic cleanup after execution

### ⚡ **High Performance**
- Redis-powered queue with microsecond latency
- Real-time execution status updates
- Handles 1M+ daily executions
- 99.9% uptime SLA

### 🌐 **Multi-Language Support**
Execute code in multiple programming languages:
- Python
- JavaScript (Node.js)
- Java
- C++

> ⚠️ **Note:** Go execution support is temporarily unavailable due to Docker sandbox limitations.

### 📈 **Scalable Architecture**
- Distributed worker pool
- Kafka support for horizontal scaling
- Auto-scaling based on demand
- Pay-per-use pricing model

---

## 💡 Use Cases

### **Online Learning Platforms**
Add interactive coding exercises to your courses. Students can write and execute code directly in your platform with instant feedback.

### **Coding Bootcamps**
Build comprehensive curriculum with hands-on practice. Track student progress and provide automated code evaluation.

### **Technical Assessments**
Create coding challenges for hiring and skill evaluation. Securely run candidate submissions without infrastructure overhead.

### **Developer Documentation**
Embed live code examples in your docs. Let users experiment with your API or SDK in real-time.

---

## 📡 API Reference

### Base URL
```
https://api.neuron.dev
```

### Authentication
Include your API key in the request header:
```
Authorization: Bearer YOUR_API_KEY
```

---

### Submit Code for Execution

Execute user-submitted code in a secure sandbox.

**Endpoint:** `POST /api/v1/runner/submit`

**Request Body:**
```json
{
  "code": "print('Hello World')",
  "language": "python",
  "input": ""
}
```

**Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `code` | string | Yes | Source code to execute |
| `language` | string | Yes | Language: `python`, `javascript`, `java`, `cpp` |
| `input` | string | No | Standard input for the program |

**Response:**
```json
{
  "jobId": "x7k9m2p4",
  "status": "queued",
  "submittedAt": "2025-12-12T10:30:00Z"
}
```

**Status Codes:**
- `200` - Code submitted successfully
- `400` - Invalid request (missing parameters, unsupported language)
- `401` - Unauthorized (invalid API key)
- `429` - Rate limit exceeded

---

### Check Execution Status

Track the progress of your code execution.

**Endpoint:** `GET /api/v1/runner/:jobId/status`

**Example:**
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
  https://api.neuron.dev/api/v1/runner/x7k9m2p4/status
```

**Response:**
```json
{
  "jobId": "x7k9m2p4",
  "status": "success",
  "output": "Hello World\n",
  "executionTime": 245,
  "completedAt": "2025-12-12T10:30:02Z"
}
```

**Status Values:**
| Status | Description |
|--------|-------------|
| `queued` | Waiting in execution queue |
| `running` | Currently executing |
| `success` | Completed successfully |
| `failed` | Execution error (syntax, runtime, timeout) |
| `cancelled` | Execution was cancelled |

**Response Fields:**
- `output` - Standard output from the program
- `error` - Error message (if status is `failed`)
- `executionTime` - Time taken in milliseconds
- `completedAt` - ISO 8601 timestamp

---

### Get Execution Result

Retrieve complete execution details including output and errors.

**Endpoint:** `GET /api/v1/runner/:jobId/result`

**Response:**
```json
{
  "jobId": "x7k9m2p4",
  "status": "success",
  "code": "print('Hello World')",
  "language": "python",
  "input": "",
  "output": "Hello World\n",
  "error": null,
  "executionTime": 245,
  "memoryUsed": 8192,
  "submittedAt": "2025-12-12T10:30:00Z",
  "completedAt": "2025-12-12T10:30:02Z"
}
```

---

## 🚀 Getting Started

### 1. Sign Up
Visit [neuron.dev](https://neuron.dev) and create an account.

### 2. Get API Key
Generate your API key from the dashboard.

### 3. Make Your First Request

```bash
curl -X POST https://api.neuron.dev/api/v1/runner/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "code": "print(\"Hello from Neuron\")",
    "language": "python"
  }'
```

### 4. Check Result

```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
  https://api.neuron.dev/api/v1/runner/JOB_ID/result
```

---

## 📊 Pricing

### Free Tier
- 1,000 executions/month
- All languages supported
- Community support

### Pro Tier
- 50,000 executions/month
- Priority execution queue
- Email support
- Custom resource limits

### Enterprise
- Unlimited executions
- Dedicated infrastructure
- SLA guarantees
- 24/7 support
- Custom integrations

[View detailed pricing →](https://neuron.dev/pricing)

---

## 🛠️ SDKs & Libraries

Official SDKs for easy integration:

- **Python** - `pip install neuron-sdk`
- **JavaScript/Node.js** - `npm install @neuron/sdk`
- **Java** - `maven: dev.neuron:sdk`
- **Go** - `go get github.com/neuron/go-sdk`

---

## 📚 Resources

- [API Documentation](https://docs.neuron.dev)
- [Integration Guides](https://docs.neuron.dev/guides)
- [Code Examples](https://github.com/neuron/examples)
- [Status Page](https://status.neuron.dev)

---

## 🔧 Self-Hosting

Want to run Neuron on your own infrastructure?

See [SETUP.md](./SETUP.md) for local development and self-hosting instructions.

---

## 🤝 Support

- **Documentation:** [docs.neuron.dev](https://docs.neuron.dev)
- **Email:** support@neuron.dev
- **Discord:** [Join our community](https://discord.gg/neuron)
- **Issues:** [GitHub Issues](https://github.com/anurag-327/neuron/issues)

---

## 📄 License

MIT License - see [LICENSE](./LICENSE) for details

---

<div align="center">

**Built with ❤️ for educators and creators worldwide**

[Website](https://neuron.dev) • [Documentation](https://docs.neuron.dev) • [GitHub](https://github.com/anurag-327/neuron)

</div>