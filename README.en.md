# dockertest Examples

Sample project demonstrating integration testing with [dockertest](https://github.com/ory/dockertest).

## Overview

This project demonstrates how to use the dockertest library to launch Docker containers and run integration tests.

## Test Cases

### 1. TestStdout
Launches a `hello-world` container and verifies that "Hello from Docker!" appears in stdout.
- Demonstrates container log retrieval

### 2. TestNginx
Launches an nginx container and verifies HTTP GET request returns status 200.
- Port binding
- Startup waiting with `pool.Retry()`
- Dynamic port allocation (`resource.GetPort()`)

### 3. TestMultipleServices
Launches two Echo applications (service-a and service-b) and tests inter-service communication.
- service-a: Calls service-b and transforms the result
- service-b: Returns a simple JSON response
- Build and run from Dockerfile (`BuildAndRunWithOptions`)
- Tests service-to-service communication

## Project Structure

```
.
├── main_test.go           # Test code
├── docker-compose.yml     # Docker Compose configuration
├── service-a/
│   ├── main.go           # Service A Echo app
│   ├── Dockerfile        # Service A Dockerfile
│   ├── go.mod
│   └── go.sum
└── service-b/
    ├── main.go           # Service B Echo app
    ├── Dockerfile        # Service B Dockerfile
    ├── go.mod
    └── go.sum
```

## Running Tests

### Run Locally

```bash
go test -v ./...
```

### Run with Docker Compose

```bash
docker compose up
```

When running with Docker Compose, the `DOCKER_HOST_ADDR` environment variable is used to connect to the Docker host via `host.docker.internal`.

## Key Points

### Dynamic Port Allocation
Tests use `resource.GetPort()` to retrieve dynamically allocated ports, avoiding port conflicts.

### Docker Compose Compatibility
By checking the `DOCKER_HOST_ADDR` environment variable, the code supports both local execution and execution within Docker Compose:

```go
dockerHost := os.Getenv("DOCKER_HOST_ADDR")
if dockerHost == "" {
    dockerHost = "localhost"
}
```

### Resource Cleanup
Uses `defer pool.Purge(resource)` to ensure containers are deleted when tests complete.

### Multi-stage Builds
Dockerfiles for service-a and service-b use multi-stage builds to reduce image size.

## Requirements

- Go 1.23.0 or later
- Docker
- Docker Compose (if running with Docker Compose)

## References

- [dockertest](https://github.com/ory/dockertest)
- [dockertest examples](https://github.com/ory/dockertest/tree/v3/examples)