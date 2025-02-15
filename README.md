# 🚀 MFT Runner - Enterprise File Transfer Benchmarking Suite

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-blue)
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/FulcrumIndustries/MFT_Runner/build.yml)

A high-performance testing framework for evaluating file transfer protocols under various network conditions and load scenarios.

![Dashboard Preview](docs/dashboard-preview.png)

## 🌟 Key Features

- **Multi-Protocol Testing**: FTP, SFTP, HTTP/HTTPS support
- **Visual Analytics**: Illustrated dashboard after test execution
- **Campaign System**: Save and reuse test configurations
- **Smart Load Generation**:
  - Custom file size distributions
  - Concurrent client simulations
  - Automated cleanup
- **Historical Comparison**: Track performance changes over time

## 🛠 Getting Started

### Prerequisites

- Go 1.21+ [only for building the .exe if needed]
- Node.js 18+ (for web interface) [frontend directory]
- Test server running (FTP/SFTP/HTTP) [one is provided in the repo: testserver.py]

```bash
# Clone & Build
git clone https://github.com/yourusername/mft-runner.git
cd mft-runner
make build-backend
make install-frontend
```

## 🧪 Basic Workflow

### 1. Run Upload Test

```bash
./mft-runner Campaigns/UPLOAD_FTP_1KB.json 10 100
```

- `10` concurrent clients
- `100` total files
- Outputs `UploadTestID` for download tests

### 2. Run Download Test

- Set the `UploadTestID` in the download campaign file

```bash
./mft-runner Campaigns/DOWNLOAD_FTP_1KB.json 10 100
```

- Downloads same files uploaded in first test

### 3. View Results

1. Launch web interface from `frontend` directory:

```bash
npm start
```

2. Open `http://localhost:3000`
3. Upload the generated reports in the `TestReports` directory in the UI

## 📂 Campaign Management

### Example Upload Campaign

```json:Campaigns/UPLOAD_SFTP_1KB.json
{
  "FilesizePolicies": [{ "Size": 1, "Unit": "K", "Percent": 100 }],
  "Protocol": "SFTP",
  "Type": "UPLOAD",
  "Host": "localhost",
  "Port": 2222,
  "Username": "user",
  "Password": "pass",
  "RemotePath": "/uploads/"
}
```

### Example Download Campaign

```json:Campaigns/DOWNLOAD_SFTP_1KB.json
{
  "Protocol": "SFTP",
  "Type": "DOWNLOAD",
  "UploadTestID": "test_123456789",  // From upload test
  "Host": "localhost",
  "Port": 2222,
  "RemotePath": "/uploads/"
}
```

## 📊 Web Interface Features

- Interactive performance dashboards
- Comparative analysis of test runs
- Error distribution breakdowns
- Latency percentile charts
- Throughput over time visualizations

## 📡 CLI Commands

```bash
# List available campaigns
./mft-runner -lc

# View campaign details
./mft-runner -vc DOWNLOAD_FTP_1KB

# Run test with custom parameters
./mft-runner <campaign> <clients> <requests>
```

## 🔄 Workflow Diagram

```mermaid
graph TD
    A[Create Upload Campaign] --> B[Run Upload Test]
    B --> C{Get UploadTestID}
    C --> D[Create Download Campaign]
    D --> E[Run Download Test]
    E --> F[Analyze Results in UI]
```

## 🧩 Core Components

| Component         | Description                               |
| ----------------- | ----------------------------------------- |
| Test Orchestrator | Manages client workers and test lifecycle |
| Protocol Handlers | FTP/SFTP/HTTP implementation modules      |
| Result Analyzer   | Processes metrics and generates reports   |
| Web Dashboard     | React-based visualization interface       |

## 📈 Key Metrics Tracked

- Throughput (requests/sec)
- Data transfer rates (MB/s)
- Latency distributions
- Error rates
- Resource utilization
- Protocol-specific metrics

## 🛠 Installation

### Prerequisites

- Go 1.21+
- Node.js 18+ (for web interface)
- Make (optional)

bash
Clone repository
git clone https://github.com/FulcrumIndustries/MFT_Runner.git
cd MFT_Runner
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
Access the web interface at `http://localhost:5173`

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

Access the web interface at `http://localhost:5173`

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

- [x] Core protocol implementations : FTP, SFTP, HTTP
- [x] Handling upload/download
- [x] Web dashboard using Test reports

## 🤝 Contributing

We welcome contributions! Please read our
[Contribution Guidelines](CONTRIBUTING.md) and
[Code of Conduct](CODE_OF_CONDUCT.md) before submitting PRs.

## 📜 License

Distributed under MIT License. See `LICENSE` for full text.

---

_MFT Runner is maintained by [Aionyx] and contributors._
