# Smart Energy Grid Management System

A cloud-based IoT platform for monitoring and managing energy grids. Built with Go and AWS services.

## Architecture

- Backend API: Go REST API on AWS Elastic Beanstalk
- Frontend: Go web application
- Database: PostgreSQL (local) / DynamoDB (production)
- Message Queue: AWS SNS
- Serverless: AWS Lambda functions
- Storage: AWS S3
- Custom Library: energy-grid-analytics v1.0.0

## Getting Started

### Requirements

- Go 1.22 or higher
- AWS CLI
- PostgreSQL
- Docker (optional)

### Running Locally

1. Clone the repo
   ```bash
   git clone <repository-url>
   cd smart-energy-grid-management-system
   ```

2. Install dependencies
   ```bash
   go mod download
   ```

3. Setup environment variables
   ```bash
   cp .env.example .env
   # Update .env with your settings
   ```

4. Run database setup
   ```bash
   psql -U postgres -d energy_grid -f scripts/schema.sql
   psql -U postgres -d energy_grid -f scripts/seed.sql
   ```

5. Start backend
   ```bash
   go run ./cmd/api/main.go
   ```

6. Start frontend (new terminal)
   ```bash
   cd web/energy-dashboard
   go run main.go
   ```

## Project Structure

```
.
├── cmd/                      # Entry points
├── internal/                 # Internal packages
├── lambda-functions/         # Lambda code
├── web/energy-dashboard/     # Frontend
├── scripts/                  # Database scripts
└── .github/workflows/        # CI/CD
```

## Configuration

Create a `.env` file:

```env
API_ADDR=:8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=energy_grid
USE_CLOUD_SERVICES=false
AWS_REGION=eu-north-1
```

## Deployment

Using GitHub Actions for automated deployment to AWS eu-north-1 region.

Required secrets:
- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY

Manual deployment:
```bash
eb init smart-energy-grid --platform go --region eu-north-1
eb create smart-energy-grid-env --single --instance_type t3.micro
eb deploy
```

## Testing

```bash
go test ./...
go test -cover ./...
```

## API Endpoints

- GET /health - Check service status
- GET /facilities - List facilities
- GET /meters - List meters
- GET /readings/recent - Get recent readings
- POST /readings/check-anomaly - Check anomalies
- GET /alerts - List alerts
- POST /analytics/generate - Generate reports

## Troubleshooting

Database connection issues: Make sure PostgreSQL is running and credentials are correct.

AWS issues: Check credentials and permissions.

Deployment fails: Verify GitHub secrets and AWS permissions.

## Links

- Custom Library: https://github.com/ANIKETSHETTY47/energy-grid-analytics
- AWS Docs: https://docs.aws.amazon.com/elasticbeanstalk/

