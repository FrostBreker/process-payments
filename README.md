# Process Payments with Stripe

This is a Go backend service that handles payment processing using Stripe. The service provides a RESTful API built with Gin framework and uses MongoDB for data storage.

This project was created as part of a YouTube tutorial series. You can follow along with the video to understand the implementation details and concepts behind the payment processing system.

🎥 [Watch the Tutorial on YouTube](#) <!-- Add your YouTube video link here -->

## Prerequisites

- Go 1.23 or later
- MongoDB installed and running
- Stripe account with API keys
- Git (for version control)

## Installation

1. Clone the repository:
```bash
git clone <your-repository-url>
cd process-payments
```

2. Install dependencies:
```bash
go mod download
```

## Environment Setup

1. Copy the example environment file:
```bash
cp .env.example .env
```

2. Fill in the environment variables in `.env`:
- `STRIPE_WEBHOOK_SECRET_KEY`: Your Stripe webhook secret key
- `STRIPE_SECRET_KEY`: Your Stripe secret key
- `PORT`: Server port (default: 8080)
- `MONGO_URI`: MongoDB connection string (default: "mongodb://127.0.0.1:27017/")
- `PRODUCTION`: Set to "true" in production environment
- `CLIENT_URL`: Frontend application URL

## Project Dependencies

Main dependencies used in this project:

- `github.com/gin-gonic/gin`: Web framework
- `github.com/stripe/stripe-go/v82`: Stripe Go client
- `go.mongodb.org/mongo-driver`: MongoDB driver
- `github.com/joho/godotenv`: Environment variable management
- `github.com/gin-contrib/cors`: CORS middleware
- `github.com/gin-contrib/secure`: Security middleware

## Running the Application

1. Make sure MongoDB is running locally or accessible via the configured URI.

2. Start the server:
```bash
go run cmd/server/main.go
```

The server will start on the configured port (default: 8080).

## Project Structure

```
.
├── cmd/server/    # Application entrypoint
├── internal/      # Private application code
├── pkg/           # Public library code
├── .env           # Environment variables
├── .env.example   # Example environment variables
├── go.mod         # Go module definition
└── go.sum         # Go module checksums
```

## API Documentation

[Add your API endpoints and documentation here]

## Development

To run the project in development mode:

1. Ensure MongoDB is running locally
2. Set up your Stripe account and get API keys
3. Configure the `.env` file with your credentials
4. Run the application using `go run cmd/server/main.go`

## Production Deployment

For production deployment:

1. Set `PRODUCTION=true` in your environment
2. Update `CLIENT_URL` to your production frontend URL
3. Use a secure MongoDB instance
4. Configure Stripe webhook endpoints
5. Use HTTPS in production