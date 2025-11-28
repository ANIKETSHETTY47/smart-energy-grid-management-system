# Smart Energy Grid Management System

A cloud-native IoT platform for real-time energy grid monitoring, analytics, and management built with Go and AWS services.

## 🏗️ Architecture

- **Backend API**: Go-based REST API deployed on AWS Elastic Beanstalk
- **Frontend Dashboard**: Go web application with real-time monitoring
- **Database**: PostgreSQL (local) / DynamoDB (production)
- **Message Queue**: AWS SNS for notifications
- **Serverless Functions**: AWS Lambda for analytics and anomaly detection
- **Storage**: AWS S3 for data archival

## 🚀 Quick Start

### Prerequisites

- Go 1.22+
- AWS CLI configured
- PostgreSQL (for local development)
- Docker (optional)

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd smart-energy-grid-management-system
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Run database migrations**
   ```bash
   psql -U postgres -d energy_grid -f scripts/schema.sql
   psql -U postgres -d energy_grid -f scripts/seed.sql
   ```

5. **Start the backend API**
   ```bash
   go run ./cmd/api/main.go
   ```

6. **Start the frontend (in another terminal)**
   ```bash
   cd web/energy-dashboard
   go run main.go
   ```

## 📦 Project Structure

```
.
├── cmd/
│   ├── api/           # Main backend API entry point
│   ├── ingestor/      # Data ingestion service
│   └── simulator/     # Energy data simulator
├── internal/
│   ├── cloud/         # AWS service integrations
│   ├── config/        # Configuration management
│   ├── dashboard/     # Dashboard logic
│   ├── database/      # Database connections
│   ├── domain/        # Domain models
│   ├── http/          # HTTP handlers
│   ├── repository/    # Data access layer
│   └── service/       # Business logic
├── lambda-functions/
│   ├── analytics-processing/
│   └── anomaly-detection/
├── web/
│   └── energy-dashboard/  # Frontend application
├── scripts/           # SQL scripts
├── .github/workflows/ # CI/CD pipelines
├── Buildfile         # Elastic Beanstalk build config
├── Procfile          # Elastic Beanstalk process config
└── go.mod            # Go dependencies
```

## 🔧 Configuration

### Environment Variables

Create a `.env` file in the root directory:

```env
# Server Configuration
API_ADDR=:8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=energy_grid

# AWS Configuration
USE_CLOUD_SERVICES=false
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_key
AWS_SECRET_ACCESS_KEY=your_secret

# DynamoDB Tables
DYNAMODB_SENSORS_TABLE=energy-sensors
DYNAMODB_READINGS_TABLE=energy-readings

# S3
S3_BUCKET=energy-grid-data

# SNS
SNS_ALERT_TOPIC_ARN=arn:aws:sns:region:account:alerts
```

## 🚢 Deployment

### CI/CD Pipeline

The project uses GitHub Actions for automated deployments:

1. **Test & Build**: Runs tests and builds binaries
2. **Deploy Backend**: Deploys API to Elastic Beanstalk
3. **Deploy Frontend**: Deploys dashboard to Elastic Beanstalk
4. **Deploy Lambda**: Updates Lambda functions

### Manual Deployment

#### Backend to Elastic Beanstalk

```bash
# Initialize EB (first time only)
eb init smart-energy-grid --platform go --region us-east-1

# Create environment (first time only)
eb create smart-energy-grid-env --single --instance_type t3.micro

# Deploy
eb deploy
```

#### Frontend to Elastic Beanstalk

```bash
cd web/energy-dashboard

# Initialize EB (first time only)
eb init energy-dashboard-frontend --platform go --region us-east-1

# Create environment (first time only)
eb create energy-dashboard-frontend-env --single --instance_type t3.micro

# Deploy
eb deploy
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detection
go test -race ./...

# Verbose output
go test -v ./...
```

## 📊 API Endpoints

### Health Check
- `GET /health` - Service health status

### Sensors
- `GET /api/sensors` - List all sensors
- `GET /api/sensors/:id` - Get sensor details
- `POST /api/sensors` - Register new sensor
- `PUT /api/sensors/:id` - Update sensor
- `DELETE /api/sensors/:id` - Delete sensor

### Readings
- `GET /api/readings` - List readings with filters
- `POST /api/readings` - Submit new reading

### Alerts
- `GET /api/alerts` - List active alerts
- `POST /api/alerts/:id/acknowledge` - Acknowledge alert

### Dashboard
- `GET /` - Main dashboard
- `GET /dashboard` - Real-time monitoring
- `GET /alerts` - Alerts view
- `GET /equipment` - Equipment management
- `GET /analytics` - Analytics view
- `GET /ws` - WebSocket connection for real-time updates

## 🛠️ Development

### Adding New Features

1. Create feature branch: `git checkout -b feature/your-feature`
2. Implement changes
3. Write tests
4. Submit pull request

### Code Style

- Follow Go conventions and idioms
- Use `gofmt` for formatting
- Run `go vet` for static analysis
- Add comments for exported functions

## 🐛 Troubleshooting

### Common Issues

**Database Connection Failed**
- Ensure PostgreSQL is running
- Check credentials in `.env`
- Verify database exists: `psql -l`

**AWS Services Not Working**
- Verify AWS credentials are configured
- Check IAM permissions
- Ensure services are enabled in your region

**EB Deployment Fails**
- Check EB CLI version: `eb --version`
- Verify AWS permissions
- Check application logs: `eb logs`

## 📝 License

[Your License Here]

## 👥 Contributors

[Your Team Information]

## 📞 Support

For issues and questions:
- Open an issue on GitHub
- Contact: [your-email@example.com]

## 🔗 Related Documentation

- [AWS Elastic Beanstalk Docs](https://docs.aws.amazon.com/elasticbeanstalk/)
- [AWS Lambda Docs](https://docs.aws.amazon.com/lambda/)
- [Go Documentation](https://golang.org/doc/)
