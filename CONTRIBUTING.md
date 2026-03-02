# Contributing to Neuron

Thank you for your interest in contributing to **Neuron** - a high-performance code execution platform! This guide will help you understand our architecture, development workflow, and contribution process.

## 📋 Table of Contents

- [Tech Stack](#tech-stack)
- [Architecture Overview](#architecture-overview)
- [Development Setup](#development-setup)
- [Code Execution Flow](#code-execution-flow)
- [Project Structure](#project-structure)
- [Contributing Guidelines](#contributing-guidelines)
- [Adding Language Support](#adding-language-support)

---

## 🛠 Tech Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: Gin (HTTP router)
- **Database**: MongoDB (with MGM ODM)
- **Message Queue**: Redis Streams / Apache Kafka (pluggable)
- **Container Runtime**: Docker
- **Authentication**: JWT + API Keys

### Infrastructure
- **Containerization**: Docker
- **Hot Reload**: Air (development)
- **Process Management**: Two separate services (API + Worker)

---

## 🏗 Architecture Overview

Neuron follows a **microservices-inspired architecture** with two main components:

### 1. **API Server** (`cmd/api/main.go`)
- Handles HTTP requests
- Validates user code
- Publishes jobs to message queue
- Manages user authentication & credits
- Serves analytics and logs

### 2. **Worker** (`cmd/worker/main.go`)
- Consumes jobs from message queue
- Forwards execution to the Core Engine
- Updates job status and results

### Key Design Patterns

- **Factory Pattern**: For creating publishers, subscribers
- **Repository Pattern**: Database access layer

---

## 🚀 Development Setup

### Prerequisites

```bash
# Required
- Go 1.21+
- Docker Desktop
- MongoDB
- Redis (or Kafka)

# Optional
- Air (for hot reload)
```

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/anurag-327/neuron.git
cd neuron/backend
```

2. **Install dependencies**
```bash
go mod download
```

3. **Set up environment variables**
```bash
cp .env.sample .env
# Edit .env with your configuration
```

4. **Start MongoDB and Redis**
```bash
docker-compose up -d mongodb redis
```

5. **Run the services**

```bash
# Terminal 1: API Server
air -c .air.api.toml

# Terminal 2: Worker
air -c .air.worker.toml
```

The API will be available at `http://localhost:8080`

---

## 🔄 Code Execution Flow

Here's how code execution works in Neuron:

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ POST /api/v1/runner/submit
       ▼
┌─────────────────────────────────────────┐
│          API Server                      │
│  1. Authenticate user                    │
│  2. Validate code (security checks)      │
│  3. Create Job document (status: queued) │
│  4. Publish to message queue             │
│  5. Return job ID to client              │
└──────────────┬──────────────────────────┘
               │
               ▼
        ┌──────────────┐
        │ Message Queue│ (Redis/Kafka)
        └──────┬───────┘
               │
               ▼
┌──────────────────────────────────────────┐
│           Worker                          │
│  1. Consume job from queue                │
│  2. Forward job to Core Engine (HTTP)     │
│  3. Wait for execution results            │
│  4. Compile execution logs (stdout/stderr)│
│  5. Update job status (success/failed)    │
│  6. Deduct user credits                   │
└───────────────────────────────────────────┘
```

---

## 📁 Project Structure

```
backend/
├── cmd/
│   ├── api/          # API server entrypoint
│   └── worker/       # Worker entrypoint
├── config/           # Configuration (pools, credits)
├── conn/             # Database & Docker connections
├── internal/
│   ├── factory/      # Factory patterns (pubsub, sandbox)
│   ├── handler/      # HTTP handlers
│   ├── middleware/   # Auth, CORS, etc.
│   ├── models/       # Database models
│   ├── registry/     # Language configs & validators
│   ├── repository/   # Database access layer
│   ├── routes/       # Route registration
│   ├── services/     # Business logic
│   └── util/         # Helper functions
├── pkg/
│   ├── messaging/    # Queue abstraction (Redis/Kafka)
│   └── sandbox/      # Code execution engine
│       └── docker/   # Docker-specific implementation
│           └── pool/ # Container pool management
└── scripts/          # Utility scripts
```

---

## 📝 Contributing Guidelines

> **🚨 Important Note:** Code execution logic, container pooling, and language support has been moved to the isolated **Core Engine**.
> 
> If you would like to contribute to the code execution capabilities, sandboxing, or add a new programming language, please head over to the **[Core Engine directory](../core)**.
> 
> This directory (`backend`) is focused purely on the API Gateway, Authentication, Billing, and the Code Queue Worker!

### Code Style

- Follow **Go conventions** (gofmt, golint)
- Use **meaningful variable names**
- Add **comments for complex logic**
- Keep functions **small and focused**

### Commit Messages

Use conventional commits:

```
feat: add support for Rust language
fix: resolve memory leak in container pool
docs: update API documentation
refactor: simplify job validation logic
test: add unit tests for credit service
```

### Pull Request Process

1. **Fork** the repository
2. Create a **feature branch** (`git checkout -b feat/amazing-feature`)
3. **Commit** your changes (`git commit -m 'feat: add amazing feature'`)
4. **Push** to the branch (`git push origin feat/amazing-feature`)
5. Open a **Pull Request**

### PR Checklist

- [ ] Code builds without errors
- [ ] All tests pass
- [ ] Added/updated documentation
- [ ] Followed code style guidelines
- [ ] Updated CHANGELOG.md (if applicable)

---

## 🌐 Adding Language Support

See the Core Engine's `ADD_LANGUAGE.md` for detailed instructions on adding new programming languages to the runtime environment.

---

## 📚 Additional Resources

- [API Documentation](./README.md)
- [Setup Guide](./SETUP.md)
- [Stats API Documentation](./STATS_API.md)
- [Core Engine Guide](../core/README.md)

---

## 🤝 Community

- **Issues**: Report bugs or request features via [GitHub Issues](https://github.com/anurag-327/neuron/issues)
- **Discussions**: Join conversations in [GitHub Discussions](https://github.com/anurag-327/neuron/discussions)

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.

---

**Happy Contributing! 🚀**
