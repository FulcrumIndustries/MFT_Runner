# 🚀 MFT Runner - Enterprise File Transfer Benchmarking Suite

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-blue)
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/yourusername/mft-runner/build.yml)

A high-performance testing framework for evaluating file transfer protocols under various network conditions and load scenarios.

![Dashboard Preview](docs/dashboard-preview.png)

## 🌟 Features

- **Multi-Protocol Support**: Test FTP, SFTP, and HTTP transfers
- **Real-Time Analytics**: Live progress tracking and metrics dashboard
- **Campaign Management**: Save and reuse test configurations
- **Smart Load Generation**:
  - Concurrent client simulations
  - Custom file size distributions
  - Automatic cleanup
- **Historical Analysis**: Compare test results over time
- **CI/CD Ready**: Integration-friendly API endpoints

## 🛠 Installation

### Prerequisites

- Go 1.21+
- Node.js 18+ (for web interface)
- Make (optional)

bash
Clone repository
git clone https://github.com/yourusername/mft-runner.git
cd mft-runner
Build backend
make build-backend
Setup frontend
make install-frontend

## ⚙️ Configuration

Create test campaigns in `Campaigns/` directory:

json
{
"name": "Production SFTP Load Test",
"protocol": "SFTP",
"hostname": "files.example.com",
"port": 22,
"path": "/uploads",
"username": "loadtester",
"password": "securepass123",
"type": "Upload",
"timeout": 30,
"filesizepolicies": [
{"size": 1, "unit": "MB", "percent": 70},
{"size": 5, "unit": "MB", "percent": 30}
]
}
bash
Start server (from project root)
./mft-server --port 8080
In separate terminal (from frontend directory)
npm start
Access the web interface at `http://localhost:3000`

**Typical Workflow**:

1. **Create Campaign** → Define protocol parameters and file distribution
2. **Configure Test** → Set clients, requests, and duration
3. **Execute Test** → Monitor real-time transfers/second and errors
4. **Analyze Results** → View throughput graphs and failure diagnostics
5. **Compare Runs** → Track performance regressions/improvements

## 📡 API Endpoints

| Method | Endpoint              | Description                         |
| ------ | --------------------- | ----------------------------------- |
| GET    | `/api/test/status`    | Current test metrics and progress   |
| POST   | `/api/test/start`     | Initiate new test with JSON payload |
| GET    | `/api/test/history`   | List all historical test runs       |
| GET    | `/api/campaigns`      | List available test configurations  |
| GET    | `/api/campaigns/{id}` | Get detailed campaign specification |

## 🧪 Development

Access the web interface at `http://localhost:3000`

**Typical Workflow**:

1. 🛠 Create a new campaign
2. 🧪 Configure test parameters:
   - Number of concurrent clients
   - Requests per client
   - Protocol-specific settings
3. ▶️ Start test execution
4. 📊 Monitor real-time statistics
5. 📄 Review detailed reports

## 🌐 API Endpoints

| Method | Endpoint              | Description              |
| ------ | --------------------- | ------------------------ |
| GET    | /api/test/status      | Current test status      |
| POST   | /api/test/start       | Start new test           |
| GET    | /api/test/history     | Get test history         |
| GET    | /api/campaigns        | List available campaigns |
| GET    | /api/campaigns/{name} | Get campaign details     |

## 🛠 Development

bash
Run test suite
make test
Build production artifacts
make release
Start dev environment
make dev

## 🛣 Roadmap

- [x] Core protocol implementations
- [x] Web dashboard
- [ ] AWS S3 protocol support (Q4 2024)
- [ ] Network condition simulation (Q1 2025)
- [ ] Automated PDF reporting (Q2 2025)
- [ ] OAuth2 authentication (Q3 2025)

## 🤝 Contributing

We welcome contributions! Please read our
[Contribution Guidelines](CONTRIBUTING.md) and
[Code of Conduct](CODE_OF_CONDUCT.md) before submitting PRs.

## 📜 License

Distributed under MIT License. See `LICENSE` for full text.

---

_MFT Runner is maintained by [Your Name] and contributors._

## Standalone CLI Usage

```bash
./mft-runner -config test_config.json
```
