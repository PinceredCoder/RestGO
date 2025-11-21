# RestGo - Task Management API

A RESTful API service for task management built with Go, featuring Protocol Buffers for data validation and serialization.

## Features

- **Complete CRUD Operations** - Create, read, update, and delete tasks
- **Protobuf Validation** - Schema-based data validation using Protocol Buffers
- **RESTful Design** - Clean REST API with proper HTTP methods and status codes
- **UUID-Based IDs** - Unique identifiers for all tasks
- **Structured Error Handling** - Comprehensive error responses with validation details
- **Health Check Endpoint** - Service monitoring and status verification

## Technology Stack

- **Go 1.25.4** - Core language
- **Chi Router** - Lightweight HTTP routing
- **Protocol Buffers** - Data serialization and validation
- **Google UUID** - Unique identifier generation

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Service health check |
| GET | `/api/v1/tasks` | List all tasks |
| POST | `/api/v1/tasks` | Create a new task |
| GET | `/api/v1/tasks/{id}` | Get task by ID |
| PUT | `/api/v1/tasks/{id}` | Update a task |
| DELETE | `/api/v1/tasks/{id}` | Delete a task |

## Task Object Structure

```json
{
  "id": "uuid-string",
  "title": "string (1-100 chars)",
  "description": "string (max 500 chars)",
  "completed": false,
  "createdAt": "2025-11-13T10:00:00Z",
  "updatedAt": "2025-11-13T10:00:00Z"
}
```

### Validation Rules

- **Title**: Required, 1-100 characters
- **Description**: Optional, maximum 500 characters
- **Completed**: Optional boolean flag

## Getting Started

### Prerequisites

- Go 1.25.4 or higher
- Protocol Buffers compiler (`protoc`)
- Make (optional, for using Makefile commands)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd go_service
```

2. Install dependencies:
```bash
make deps
```

3. Install required protobuf tools:
```bash
make tools
```

4. Generate protobuf code:
```bash
make proto
```

### Running the Application

Start the server on `localhost:8080`:

```bash
make run
```

Or build and run the binary:

```bash
make build
./bin/api
```

### Testing the API

A [requests.rest](requests.rest) file is included for testing with REST client extensions. It contains example requests for all endpoints.

Example using curl:

```bash
# Health check
curl http://localhost:8080/health

# Create a task
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"My Task","description":"Task description"}'

# Get all tasks
curl http://localhost:8080/api/v1/tasks

# Get task by ID
curl http://localhost:8080/api/v1/tasks/{task-id}

# Update a task
curl -X PUT http://localhost:8080/api/v1/tasks/{task-id} \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated Task","completed":true}'

# Delete a task
curl -X DELETE http://localhost:8080/api/v1/tasks/{task-id}
```

## Project Structure

```
.
├── cmd/api/              # Application entry point
│   └── main.go           # Server initialization and routing
├── api/proto/v1/         # Protocol Buffer definitions
│   └── tasks.proto       # Task schema and validation rules
├── internal/
│   ├── handlers/         # HTTP request handlers
│   │   ├── tasks.go      # Task CRUD operations
│   │   └── validation.go # Validation error handling
│   └── errors/           # Error handling utilities
│       └── errors.go     # Custom error types
├── Makefile              # Build automation
├── build_proto.sh        # Protobuf generation script
└── requests.rest         # API testing examples
```

## Available Make Commands

```bash
make proto       # Generate protobuf and validation code
make proto-clean # Remove generated protobuf files
make tools       # Install required protobuf tools
make run         # Run the application
make build       # Build the application
make test        # Run tests
make deps        # Update Go dependencies
make help        # Show available commands
```

## Error Handling

The API returns structured error responses:

```json
{
  "type": "VALIDATION_ERROR",
  "message": "Validation failed",
  "details": {
    "title": ["value length must be at least 1 characters"]
  }
}
```

Error types:
- `VALIDATION_ERROR` - Invalid input data
- `NOT_FOUND` - Resource not found
- `BAD_REQUEST` - Malformed request
- `INTERNAL_ERROR` - Server error

## Development

### Adding New Endpoints

1. Define the data structure in [api/proto/v1/tasks.proto](api/proto/v1/tasks.proto)
2. Generate protobuf code: `make proto`
3. Implement the handler in [internal/handlers/](internal/handlers/)
4. Register the route in [cmd/api/main.go](cmd/api/main.go)

### Modifying Validation Rules

Edit the protobuf definitions in [api/proto/v1/tasks.proto](api/proto/v1/tasks.proto) and regenerate code with `make proto`.

## License

[TBD]

## Contributing

[TBD]
