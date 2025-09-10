#!/bin/bash

echo "Cleaning up previous containers..."
docker stop test-server 2>/dev/null || true
docker rm test-server 2>/dev/null || true

echo '{
    "urls": [
        "https://api.publicapis.org/entries",
        "https://catfact.ninja/fact",
        "https://www.boredapi.com/api/activity", 
        "https://invalid-url-that-will-fail.com"
    ]
}' > test_data.json

echo "Starting server..."
docker build -t api-fetcher .
docker run -d -p 8080:8080 --name test-server api-fetcher

echo "Waiting for server to start..."
sleep 3

echo "Test 1: Testing home page..."
response=$(curl -s http://localhost:8080/)
if [[ "$response" == "Concurrent API Fetcher Server is up and running!" ]]; then
    echo "Home page test passed"
else
    echo "Home page test failed: $response"
    docker logs test-server
    exit 1
fi

echo "Test 2: Testing API fetcher..."
response=$(curl -s -X POST http://localhost:8080/fetch \
  -H "Content-Type: application/json" \
  -d '{"urls":["https://httpbin.org/get", "https://httpbin.org/user-agent"]}')

if echo "$response" | grep -q "httpbin.org"; then
    echo "API fetcher test passed"
else
    echo "API fetcher test failed: $response"
    docker logs test-server
    exit 1
fi

echo "Test 3: Testing validation..."
response=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/fetch \
  -H "Content-Type: application/json" \
  -d '{"urls":[]}')
if [ "$response" -eq 400 ]; then
    echo "Validation test passed"
else
    echo "Validation test failed: HTTP code $response"
    docker logs test-server
    exit 1
fi

echo "Test 4: Testing with prepared test data..."
response=$(curl -s -X POST http://localhost:8080/fetch \
  -H "Content-Type: application/json" \
  -d '@test_data.json')

if echo "$response" | grep -q "catfact.ninja" && echo "$response" | grep -q "error"; then
    echo "Test data validation passed"
else
    echo "Test data validation failed"
    echo "Response was: $response"
    docker logs test-server
    exit 1
fi

echo "All tests passed!"

rm -f test_data.json
docker stop test-server 2>/dev/null || true
docker rm test-server 2>/dev/null || true