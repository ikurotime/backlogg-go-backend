# Backlogg Go Backend

A robust backend service built with Go for managing project backlogs and tasks. This service provides a RESTful API using the Gin framework and MongoDB for data persistence.

## 🚀 Features

- RESTful API endpoints for project management
- MongoDB integration for data persistence
- Graceful shutdown handling
- Configuration management using YAML
- Modular and clean architecture

## 📋 Prerequisites

- Go 1.23.3 or higher
- MongoDB instance
- Git

## 🛠️ Tech Stack

- **Framework:** [Gin](https://github.com/gin-gonic/gin) - High-performance HTTP web framework
- **Database:** [MongoDB](https://www.mongodb.com/) - NoSQL database
- **Configuration:** YAML-based configuration management
- **Dependencies:** Managed via Go modules

## 🏗️ Project Structure

```
.
├── cmd/
│   └── main.go          # Application entry point
├── config/             # Configuration files and management
├── internal/           # Private application code
│   ├── projects/       # Project-related handlers and logic
│   └── router/         # Router setup and configuration
├── pkg/                # Public libraries that can be used by other projects
│   ├── mongodbx/      # MongoDB connection and utilities
│   ├── root/          # Root directory utilities
│   └── yamlx/         # YAML configuration utilities
├── go.mod             # Go module definition
└── go.sum             # Go module checksums
```

## ⚙️ Configuration

The application uses YAML configuration files located in the `config` directory. Create a `.env.dev` file in the config directory with the following structure:

```yaml
mongodb:
  protocol: "mongodb"
  port: "27017"
  host: "localhost"
  username: "your_username"
  password: "your_password"
  database: "your_database"
  db: 0
```

## 🚀 Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/backlogg-go-backend.git
   cd backlogg-go-backend
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up your configuration:
   - Copy the example configuration file
   - Modify the values according to your environment

4. Run the application:
   ```bash
   go run cmd/main.go
   ```

The server will start on port 8080 by default.

## 🔄 API Endpoints

### Projects

- `GET /projects` - Retrieve all projects
- More endpoints coming soon...

## 💻 Development

To contribute to the project:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🤝 Contributing

Contributions, issues, and feature requests are welcome! Feel free to check the issues page.

## 📧 Contact

Kuro - [@ikurotime](https://twitter.com/ikurotime)

Project Link: [https://github.com/ikurotime/backlogg-go-backend](https://github.com/ikurotime/backlogg-go-backend) 