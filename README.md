# Concurrent API Fetcher

## Overview

Concurrent API Fetcher is a lightweight Go-based HTTP server designed to fetch data from multiple API endpoints concurrently. It accepts a list of URLs via a POST request, retrieves the data from each URL in parallel, and returns the results as a JSON array. The project is containerized using Docker for easy deployment and includes comprehensive unit and integration tests to ensure reliability.

## Features

- **Concurrent Fetching**: Uses Go goroutines and a WaitGroup to fetch data from multiple URLs simultaneously, improving performance for bulk API requests.
- **Error Handling**: Gracefully handles invalid URLs, network errors, timeouts, and invalid JSON payloads, returning detailed error messages in the response.
- **Context Timeout**: Implements a 5-second timeout for all API requests to prevent hanging on slow or unresponsive endpoints.
- **Docker Support**: Includes a Dockerfile for building a minimal Alpine-based image and a Makefile for streamlined building, running, and testing.
- **Testing**: Comprehensive unit tests (`main_test.go`) cover successful requests, invalid URLs, network errors, body reading errors, and context timeouts. Integration tests (`test.sh`) verify real-world API interactions and server behavior.
- **Simple API**: Exposes two endpoints:
  - `GET /`: Returns a status message indicating the server is running.
  - `POST /fetch`: Accepts a JSON payload with a list of URLs and returns the fetched data or errors.

## Project Structure

- **Dockerfile**: Defines a multi-stage build process using `golang:1.22-alpine` for building and `alpine:latest` for the runtime, ensuring a small image size.
- **.dockerignore**: Excludes `.git` to optimize the Docker build context.
- **go.mod**: Specifies the Go module and version (`go 1.22.2`).
- **main.go**: Contains the core server logic, including the home handler, fetch handler, and concurrent URL fetching function.
- **main_test.go**: Includes unit tests for the `fetchURL` function and `fetchHandler`, covering various success and failure scenarios.
- **Makefile**: Provides commands for building, running, testing, and cleaning the project, including Docker operations.
- **test.sh**: Shell script for running integration tests against real APIs and validating server responses.

## Usage

1. **Clone the Repository**:

   ```bash
   git clone <repository-url>
   cd concurrent-api-fetcher
   ```

2. **Build and Run Locally**:

   ```bash
   make build
   ./workerpool
   ```

   The server will start on `localhost:8080`.

3. **Run with Docker**:

   ```bash
   make docker_run
   ```

   This builds and runs the Docker container, exposing the server on `localhost:8080`.

4. **Make a Request**:
   Send a POST request to `/fetch` with a JSON payload containing a list of URLs:

   ```bash
   curl -X POST http://localhost:8080/fetch \
     -H "Content-Type: application/json" \
     -d '{"urls": ["https://api.publicapis.org/entries", "https://catfact.ninja/fact"]}'
   ```

   Example response:

   ```json
   [
       {"url": "https://api.publicapis.org/entries", "data": "{\"entries\":[...]}", "error": ""},
       {"url": "https://catfact.ninja/fact", "data": "{\"fact\":\"Cats can jump...\"}", "error": ""}
   ]
   ```

5. **Run Tests**:
   - Unit tests: `make test`
   - Integration tests (requires server running): `make test_sh`
   - Integration tests in Docker: `make docker_test`
   - Code coverage: `make coverage`

## Endpoints

- **GET /**: Returns "Concurrent API Fetcher Server is up and running!" to confirm the server is operational.
- **POST /fetch**:
  - **Request Body**: JSON object with a `urls` field containing an array of URLs (e.g., `{"urls": ["url1", "url2"]}`).
  - **Response**: JSON array of objects containing `url`, `data` (response body), and `error` (if any).
  - **Errors**:
    - `400 Bad Request`: Invalid JSON or empty URL list.
    - `405 Method Not Allowed`: Non-POST requests to `/fetch`.

## Requirements

- **Go**: Version 1.22.2 or higher.
- **Docker**: For containerized deployment and testing.
- **curl**: For running integration tests in `test.sh`.

## Makefile Commands

- `make build`: Build the Go binary.
- `make docker_build`: Build the Docker image.
- `make docker_run`: Run the Docker container.
- `make docker_stop`: Stop and remove the container.
- `make docker_logs`: View container logs.
- `make docker_test`: Run integration tests in Docker.
- `make test_local`: Run integration tests on a local binary.
- `make test`: Run Go unit tests.
- `make test_sh`: Run integration tests (requires running server).
- `make coverage`: Generate a code coverage report.
- `make clean`: Remove built files.
- `make style`: Format Go code.
- `make help`: Display available commands.

## Testing

- **Unit Tests**: Cover `fetchURL` and `fetchHandler` for success cases, invalid URLs, network errors, body reading errors, context timeouts, and invalid JSON payloads.
- **Integration Tests**: Verify server behavior with real APIs (`https://api.publicapis.org/entries`, `https://catfact.ninja/fact`, etc.), validate home page responses, and test error handling for empty or invalid inputs.
- **Coverage**: Use `make coverage` to generate an HTML coverage report in the `coverage` directory.

## Limitations

- The server has a hardcoded 5-second timeout for all API requests.
- No authentication or rate-limiting is implemented.
- Only JSON responses from APIs are supported; other content types are returned as raw strings.
- The integration tests rely on external APIs, which may fail if the APIs are down or rate-limited.

## Practical Use Cases

- **Batch API Testing**: Quickly fetch data from multiple APIs for testing or monitoring purposes.
- **Data Aggregation**: Collect data from multiple sources in parallel for processing or analysis.
- **Microservices**: Serve as a lightweight component in a microservices architecture to aggregate API responses.
