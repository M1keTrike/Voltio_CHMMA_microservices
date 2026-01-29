# Contributing to Voltio CHMMA Microservices

Thank you for your interest in contributing to Voltio CHMMA! This document provides guidelines and instructions for contributing to the project.

## 🌟 How to Contribute

We welcome contributions in many forms:
- Bug fixes
- New features
- Documentation improvements
- Code refactoring
- Test coverage improvements
- Performance optimizations

## 🚀 Getting Started

### 1. Set Up Your Development Environment

Before contributing, make sure you have:

1. **Installed Prerequisites**:
   - Go 1.19 or higher
   - RabbitMQ
   - InfluxDB 2.x
   - PostgreSQL
   - Git

2. **Forked and Cloned the Repository**:
   ```bash
   # Fork the repository on GitHub
   # Then clone your fork
   git clone https://github.com/YOUR_USERNAME/Voltio_CHMMA_microservices.git
   cd Voltio_CHMMA_microservices
   ```

3. **Set Up Environment Variables**:
   ```bash
   cp .env.example .env
   cp Backend/automation-engine/.env.example Backend/automation-engine/.env
   # Edit both .env files with your configuration
   ```

4. **Install Dependencies**:
   ```bash
   # Install Go modules for each service
   cd Backend/test_producers && go mod tidy
   # Repeat for other services as needed
   ```

5. **Verify Everything Works**:
   ```bash
   # Start test producers and verify they connect
   cd Backend/test_producers/dht22
   go run main.go
   ```

### 2. Create a Branch

Create a feature branch for your changes:

```bash
git checkout -b feature/your-feature-name
```

Use descriptive branch names:
- `feature/add-mqtt-support` - for new features
- `fix/rabbitmq-connection-timeout` - for bug fixes
- `docs/improve-setup-guide` - for documentation
- `refactor/consumer-error-handling` - for refactoring

## 📝 Contribution Guidelines

### Code Style

1. **Follow Go Best Practices**:
   - Use `gofmt` to format your code
   - Run `go vet` to check for common mistakes
   - Follow the [Effective Go](https://golang.org/doc/effective_go) guide

2. **Naming Conventions**:
   - Use descriptive variable names
   - Follow Go naming conventions (camelCase for local variables, PascalCase for exported items)
   - Package names should be short, lowercase, single-word

3. **Code Organization**:
   - Keep functions focused and small
   - Group related functionality
   - Separate business logic from infrastructure code

4. **Comments**:
   - Add comments for complex logic
   - Document exported functions and types
   - Use complete sentences in comments

### Code Quality

1. **Error Handling**:
   ```go
   // Good - explicit error handling
   conn, err := amqp091.Dial(amqpURI)
   if err != nil {
       log.Fatalf("Failed to connect to RabbitMQ: %v", err)
   }
   defer conn.Close()
   
   // Bad - ignoring errors
   conn, _ := amqp091.Dial(amqpURI)
   ```

2. **Environment Variables**:
   ```go
   // Good - use environment variables for configuration
   amqpURI := getEnv("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/")
   
   // Bad - hardcoded credentials
   amqpURI := "amqp://admin:password@52.73.74.139:5672/"
   ```

3. **Logging**:
   ```go
   // Good - informative logging
   log.Printf("Connected to RabbitMQ - Queue: %s", queueName)
   log.Printf("Error publishing message: %v", err)
   
   // Bad - vague logging
   log.Println("Connected")
   log.Println("Error")
   ```

### Documentation

1. **Update README.md** if you:
   - Add new services or components
   - Change the architecture
   - Add new dependencies
   - Change setup procedures

2. **Update ENVIRONMENT_SETUP.md** if you:
   - Add new environment variables
   - Change configuration options
   - Add new services requiring configuration

3. **Add Inline Documentation** for:
   - Complex algorithms
   - Non-obvious decisions
   - Important constraints or assumptions

### Security

1. **Never Commit**:
   - Credentials or secrets
   - `.env` files
   - API keys or tokens
   - Private keys

2. **Always Use**:
   - Environment variables for sensitive data
   - Secure defaults
   - Input validation
   - Proper error messages (don't leak sensitive info)

3. **Review**:
   - Check for exposed credentials before committing
   - Verify `.gitignore` is working properly
   - Use `git status` before commits

## 🧪 Testing Your Changes

### 1. Build Your Service

```bash
cd Backend/YourService
go build -o service_name main.go
```

### 2. Test with Test Producers

```bash
# Start your consumer service
cd Backend/YourService
go run main.go

# In another terminal, start a test producer
cd Backend/test_producers/dht22
go run main.go
```

### 3. Verify Data Flow

1. Check RabbitMQ management UI
2. Query InfluxDB for stored data
3. Connect to WebSocket server and verify streaming
4. Monitor logs for errors

### 4. Test Edge Cases

- Service restart behavior
- Connection loss and reconnection
- Invalid message format handling
- Timeout conditions
- High message throughput

## 📤 Submitting Your Contribution

### 1. Commit Your Changes

Write clear, descriptive commit messages:

```bash
# Good commit messages
git commit -m "Add MQTT consumer service for ESP32 devices"
git commit -m "Fix WebSocket reconnection timeout issue"
git commit -m "Update setup documentation with Docker instructions"

# Bad commit messages
git commit -m "Fixed stuff"
git commit -m "Update"
git commit -m "WIP"
```

**Commit Message Guidelines**:
- Use present tense ("Add feature" not "Added feature")
- Use imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit first line to 72 characters
- Reference issues and pull requests when relevant

### 2. Push to Your Fork

```bash
git push origin feature/your-feature-name
```

### 3. Create a Pull Request

1. Go to the [repository on GitHub](https://github.com/M1keTrike/Voltio_CHMMA_microservices)
2. Click "New Pull Request"
3. Select your fork and branch
4. Fill out the pull request template:

   **Title**: Brief, descriptive summary
   
   **Description**:
   - What does this PR do?
   - Why is this change needed?
   - How has this been tested?
   - Related issues or PRs
   - Screenshots (if applicable)

5. Request review from maintainers

### 4. Respond to Feedback

- Be responsive to code review comments
- Make requested changes promptly
- Update your PR with fixes
- Discuss any disagreements respectfully

## 🏗️ Project-Specific Guidelines

### Adding a New Sensor Type

1. **Create Producer** (`Backend/test_producers/newsensor/main.go`):
   ```go
   package main
   
   import (
       "encoding/json"
       "log"
       "os"
       "time"
       "github.com/rabbitmq/amqp091-go"
   )
   
   const queueName = "NewSensor_queue"
   
   type NewSensorMessage struct {
       DeviceID string `json:"deviceId"`
       Payload  struct {
           MAC   string  `json:"mac"`
           Value float64 `json:"value"`
       } `json:"payload"`
   }
   
   func main() {
       // Implementation
   }
   ```

2. **Create Consumer** (`Backend/NewSensor_ConsumerSender/middleware/`):
   - Follow pattern from existing consumers
   - Implement RabbitMQ subscription
   - Add InfluxDB write logic
   - Add WebSocket forwarding
   - Implement timeout monitoring

3. **Update Configuration**:
   - Add `NEWSENSOR_QUEUE_NAME` to `.env.example`
   - Add `NEWSENSOR_WEBSOCKET_URI` to `.env.example`
   - Document in `ENVIRONMENT_SETUP.md`

4. **Update Scripts**:
   - Add to `start_all_voltio_services.ps1`
   - Add to test producer start scripts

5. **Update Documentation**:
   - Add to README.md services section
   - Add to architecture diagram
   - Document message format

### Modifying Existing Services

1. **Maintain Backward Compatibility**:
   - Don't change message formats without versioning
   - Keep existing environment variables working
   - Deprecate features gradually

2. **Update Tests**:
   - Test with existing test producers
   - Verify data still flows correctly
   - Check alert system still works

3. **Update Documentation**:
   - Reflect changes in README.md
   - Update configuration docs if needed

## 🐛 Reporting Bugs

When reporting bugs, please include:

### Bug Report Template

```markdown
**Description**
A clear description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Start service X
2. Send message Y
3. See error

**Expected Behavior**
What you expected to happen.

**Actual Behavior**
What actually happened.

**Environment**
- OS: [e.g., Windows 10, Ubuntu 20.04]
- Go Version: [e.g., 1.20.5]
- RabbitMQ Version: [e.g., 3.12.0]
- InfluxDB Version: [e.g., 2.7.0]

**Logs**
```
Paste relevant log output here
```

**Additional Context**
Any other context about the problem.
```

## 💡 Feature Requests

When requesting features, please include:

1. **Use Case**: Why is this feature needed?
2. **Proposed Solution**: How should it work?
3. **Alternatives**: What alternatives have you considered?
4. **Additional Context**: Any other relevant information

## 📚 Resources

- [Go Documentation](https://golang.org/doc/)
- [RabbitMQ Tutorials](https://www.rabbitmq.com/getstarted.html)
- [InfluxDB Documentation](https://docs.influxdata.com/influxdb/v2/)
- [WebSocket Protocol](https://datatracker.ietf.org/doc/html/rfc6455)

## ❓ Questions?

If you have questions about contributing:

1. Check the [README.md](./README.md)
2. Review [ENVIRONMENT_SETUP.md](./ENVIRONMENT_SETUP.md)
3. Search existing [issues](https://github.com/M1keTrike/Voltio_CHMMA_microservices/issues)
4. Ask in a new issue with the "question" label

## 🎉 Recognition

Contributors will be recognized in:
- The project's contributor list
- Release notes for significant contributions
- GitHub contributor insights

Thank you for helping make Voltio CHMMA better!

---

**Code of Conduct**: Be respectful, inclusive, and collaborative. We're all here to learn and build great software together.
