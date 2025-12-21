# Documentation Summary

This directory contains comprehensive documentation for the Neuron code execution platform.

## 📚 Available Documentation

### [CONTRIBUTING.md](./CONTRIBUTING.md)
**Complete contribution guide** covering:
- 🛠 Tech stack (Go, MongoDB, Redis/Kafka, Docker)
- 🏗 Architecture overview (API + Worker microservices)
- 🚀 Development setup instructions
- 🔄 Code execution flow diagram
- 📁 Project structure
- 📝 Code style and PR guidelines

### [LANGUAGE_SUPPORT.md](./LANGUAGE_SUPPORT.md)
**Step-by-step guide for adding new programming languages**:
- ✅ 4-step process (Registry → Validator → Pool → Error Detection)
- 🔒 Security validation examples
- 🐳 Docker pool configuration
- 🧪 Testing guidelines
- 🎯 Complete Rust implementation example

### [SETUP.md](./SETUP.md)
**Initial setup and deployment guide** (existing)

### [README.md](./README.md)
**Project overview and API documentation** (existing)

### [STATS_API.md](./STATS_API.md)
**Statistics API documentation** with response formats

---

## 🚀 Quick Start for Contributors

1. **Read** [CONTRIBUTING.md](./CONTRIBUTING.md) to understand the architecture
2. **Follow** [SETUP.md](./SETUP.md) to set up your development environment
3. **Add languages** using [LANGUAGE_SUPPORT.md](./LANGUAGE_SUPPORT.md)
4. **Submit** a pull request following the guidelines

---

## 🏗 Architecture at a Glance

```
Client Request
     ↓
API Server (validates, queues)
     ↓
Message Queue (Redis/Kafka)
     ↓
Worker (executes in Docker)
     ↓
Results stored in MongoDB
```

---

## 🤝 Contributing

We welcome contributions! Please read [CONTRIBUTING.md](./CONTRIBUTING.md) for:
- Development workflow
- Code style guidelines
- Testing requirements
- PR process

---

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/anurag-327/neuron/issues)
- **Discussions**: [GitHub Discussions](https://github.com/anurag-327/neuron/discussions)

---

**Built with ❤️ by the Anurag 🙏**
