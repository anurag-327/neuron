
# 🚀 Getting Started

This guide walks you through **cloning**, **environment setup**, **running with Docker**, and **running API + Worker using Air**.

---

# 📋 Prerequisites

Before running this project, make sure you have the following installed:

### **1. Docker**

Required to run:

* Kafka
* Sandbox / Code Runner

Install from:
[https://www.docker.com/get-started/](https://www.docker.com/get-started/)

---

### **2. Go (Golang)**

Needed only for running API and Worker in development mode.
Version: **Go 1.22+**

Install from:
[https://go.dev/dl/](https://go.dev/dl/)

---

### **3. MongoDB (Local Installation)**

We assume MongoDB is installed locally and running on **port 27017**.

Download from:
[https://www.mongodb.com/try/download/community](https://www.mongodb.com/try/download/community)

> ⚠️ If your MongoDB runs on a different port, update the `MONGO_URI` in your `.env`.

---

# 1. Clone the Repository

```bash
git clone https://github.com/anurag-327/neuron.git
cd neuron
```

---

# 2. Environment Setup

Create a `.env` file in the project root:

```env
PORT=8080
KAFKA_BROKER=localhost:9092
MONGO_URI=mongodb://localhost:27017
MONGO_DB_NAME=neuron
```

These variables are required by both **API** and **Worker** services.

---

# 3. Running Services with Docker Compose

Start required infrastructure:

```bash
docker compose up -d
```

This launches:

* Kafka
* Sandbox Docker environment

Check if containers are running:

```bash
docker ps
```

---

# 4. Running API & Worker in Development Mode (Hot Reload with Air)

Install **Air**:

```bash
go install github.com/air-verse/air@latest
```

The repository already contains:

* `.air.api.toml`
* `.air.worker.toml`

### Run API (live reload)

```bash
air -c .air.api.toml
```

### Run Worker / Consumer (live reload)

```bash
air -c .air.worker.toml
```

Both services restart automatically on file changes.

---

# 5. Accessing the API

The API server runs at:

```
http://localhost:8080
```

### 1. Submit Code

```
POST /api/v1/runner/submit
```

**Sample Request:**

```json
{
  "code": "print('Hello Python')",
  "language": "python",
  "input": ""
}
```

---

### 2. Check Job Status

```
GET /api/v1/runner/:jobId/status
```

Example:

```
GET http://localhost:8080/api/v1/runner/12345/status
```
