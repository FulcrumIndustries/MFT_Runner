# Contributing Guidelines

We welcome contributions from the community! Here's how to get started:

## 🛠 Development Setup

1. **Fork & Clone**

   ```bash
   git clone https://github.com/your-fork/MFT_Runner.git
   cd MFT_Runner
   ```

2. **Install Dependencies**

   ```bash
   # Backend (Go)
   go get ./...

   # Frontend (Node.js)
   cd frontend
   npm install
   ```

3. **Build System**

   ```bash
   # Backend
   make build-backend

   # Frontend
   make build
   ```

## 🔄 Workflow

1. Create a feature branch:

   ```bash
   git checkout -b feat/my-feature
   ```

2. Follow our coding standards:

   - Go: `gofmt` style, 120 char line length
   - JavaScript: Prettier + ESLint enforced
   - 2-space indentation for all files

3. Write tests for new features:

   ```bash
   # Run backend tests
   go test -v ./...

   # Run frontend tests
   cd frontend
   npm test
   ```

4. Commit with semantic messages:

   ```bash
   git commit -m "feat: add sftp connection pooling"
   git commit -m "fix: resolve race condition in worker pool"
   ```

5. Push and open a Pull Request

## 🚦 Quality Gates

- All tests must pass
- 90%+ test coverage for new code
- Documentation updates required
- No sensitive data in commits
- Signed-off-by line (Developer Certificate of Origin)

## 🐛 Reporting Issues

Use our issue template:

```markdown
**Description**
Clear explanation of the problem

**Steps to Reproduce**

1. ...
2. ...

**Expected Behavior**
What should happen

**Environment**

- OS: [e.g. Ubuntu 22.04]
- Go version:
- Node version:
```

## 💡 Feature Requests

1. Check existing issues
2. Use feature request template
3. Include:
   - Use case description
   - Proposed implementation
   - Alternatives considered

## 📚 Documentation

Help improve our docs:

- Fix typos in README
- Add code examples
- Translate documentation
- Improve test server setup guides

## ❓ Need Help?

Join our [Discord Server] or email maintainers@mftrunner.org
