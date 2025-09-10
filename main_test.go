package main

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "sync"
    "testing"
    "time"
)

// TestFetchURL_Success тестирует успешный запрос
func TestFetchURL_Success(t *testing.T) {
    mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"message": "test success"}`))
    }))
    defer mockServer.Close()

    ctx := context.Background()
    var wg sync.WaitGroup
    resultsCh := make(chan Result, 1)

    wg.Add(1)
    go fetchURL(ctx, mockServer.URL, resultsCh, &wg)

    go func() {
        wg.Wait()
        close(resultsCh)
    }()

    result := <-resultsCh

    if result.Err != "" {
        t.Errorf("Expected no error, got: %s", result.Err)
    }

    if result.URL != mockServer.URL {
        t.Errorf("Expected URL %s, got: %s", mockServer.URL, result.URL)
    }

    if !strings.Contains(result.Data, `"message": "test success"`) {
        t.Errorf("Expected response to contain test data, got: %s", result.Data)
    }
}

// TestFetchURL_RequestCreationError тестирует ошибку создания запроса
func TestFetchURL_RequestCreationError(t *testing.T) {
    ctx := context.Background()
    var wg sync.WaitGroup
    resultsCh := make(chan Result, 1)

    wg.Add(1)
    go fetchURL(ctx, "http://[invalid-url", resultsCh, &wg)

    go func() {
        wg.Wait()
        close(resultsCh)
    }()

    result := <-resultsCh

    if result.Err == "" {
        t.Error("Expected error for invalid URL, got none")
    }

    if !strings.Contains(result.Err, "invalid") && !strings.Contains(result.Err, "parse") {
        t.Errorf("Expected URL parsing error, got: %s", result.Err)
    }
}

// TestFetchURL_RequestError тестирует ошибку выполнения запроса
func TestFetchURL_RequestError(t *testing.T) {
    ctx := context.Background()
    var wg sync.WaitGroup
    resultsCh := make(chan Result, 1)

    wg.Add(1)
    go fetchURL(ctx, "http://nonexistent-domain-12345.test", resultsCh, &wg)

    go func() {
        wg.Wait()
        close(resultsCh)
    }()

    result := <-resultsCh

    if result.Err == "" {
        t.Error("Expected network error, got none")
    }
}

// TestFetchURL_ReadBodyError тестирует ошибку чтения тела ответа
func TestFetchURL_ReadBodyError(t *testing.T) {
    mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        hijacker, ok := w.(http.Hijacker)
        if ok {
            conn, _, _ := hijacker.Hijack()
            conn.Close()
        }
    }))
    defer mockServer.Close()

    ctx := context.Background()
    var wg sync.WaitGroup
    resultsCh := make(chan Result, 1)

    wg.Add(1)
    go fetchURL(ctx, mockServer.URL, resultsCh, &wg)

    go func() {
        wg.Wait()
        close(resultsCh)
    }()

    result := <-resultsCh

    if result.Err == "" {
        t.Error("Expected read error, got none")
    }
}

// TestFetchURL_ContextTimeout тестирует таймаут через контекст
func TestFetchURL_ContextTimeout(t *testing.T) {
    mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(200 * time.Millisecond)
        w.Write([]byte(`{"slow": "response"}`))
    }))
    defer mockServer.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
    defer cancel()

    var wg sync.WaitGroup
    resultsCh := make(chan Result, 1)

    wg.Add(1)
    go fetchURL(ctx, mockServer.URL, resultsCh, &wg)

    go func() {
        wg.Wait()
        close(resultsCh)
    }()

    result := <-resultsCh

    if result.Err == "" {
        t.Error("Expected timeout error, got none")
    }

    if !strings.Contains(result.Err, "context") {
        t.Errorf("Expected context error, got: %s", result.Err)
    }
}

// TestFetchHandler_Success тестирует успешный обработчик
func TestFetchHandler_Success(t *testing.T) {
    mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{"test": "data"}`))
    }))
    defer mockServer.Close()

    payload := RequestPayload{Urls: []string{mockServer.URL}}
    body, _ := json.Marshal(payload)

    req := httptest.NewRequest("POST", "/fetch", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(fetchHandler)

    handler.ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
        t.Errorf("Expected status 200, got: %d", rr.Code)
    }

    var results []Result
    if err := json.Unmarshal(rr.Body.Bytes(), &results); err != nil {
        t.Errorf("Invalid JSON response: %v", err)
    }

    if len(results) != 1 {
        t.Errorf("Expected 1 result, got: %d", len(results))
    }

    if results[0].Err != "" {
        t.Errorf("Expected no error, got: %s", results[0].Err)
    }
}

// TestFetchHandler_WrongMethod тестирует неправильный метод
func TestFetchHandler_WrongMethod(t *testing.T) {
    req := httptest.NewRequest("GET", "/fetch", nil)
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(fetchHandler)

    handler.ServeHTTP(rr, req)

    if rr.Code != http.StatusMethodNotAllowed {
        t.Errorf("Expected status 405, got: %d", rr.Code)
    }

    expected := "Only POST requests are allowed"
    if !strings.Contains(rr.Body.String(), expected) {
        t.Errorf("Expected error message '%s', got: %s", expected, rr.Body.String())
    }
}

// TestFetchHandler_InvalidJSON тестирует невалидный JSON
func TestFetchHandler_InvalidJSON(t *testing.T) {
    req := httptest.NewRequest("POST", "/fetch", bytes.NewReader([]byte(`invalid json`)))
    req.Header.Set("Content-Type", "application/json")
    
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(fetchHandler)

    handler.ServeHTTP(rr, req)

    if rr.Code != http.StatusBadRequest {
        t.Errorf("Expected status 400, got: %d", rr.Code)
    }

    if !strings.Contains(rr.Body.String(), "Invalid JSON format") {
        t.Errorf("Expected JSON error message, got: %s", rr.Body.String())
    }
}

// TestFetchHandler_EmptyURLs тестирует пустой список URL
func TestFetchHandler_EmptyURLs(t *testing.T) {
    payload := RequestPayload{Urls: []string{}}
    body, _ := json.Marshal(payload)

    req := httptest.NewRequest("POST", "/fetch", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(fetchHandler)

    handler.ServeHTTP(rr, req)

    if rr.Code != http.StatusBadRequest {
        t.Errorf("Expected status 400, got: %d", rr.Code)
    }

    if !strings.Contains(rr.Body.String(), "No URLs provided") {
        t.Errorf("Expected empty URLs error message, got: %s", rr.Body.String())
    }
}

// TestFetchHandler_MultipleURLs тестирует несколько URL
func TestFetchHandler_MultipleURLs(t *testing.T) {
    mockServer1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{"server": "1"}`))
    }))
    defer mockServer1.Close()

    mockServer2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{"server": "2"}`))
    }))
    defer mockServer2.Close()

    payload := RequestPayload{Urls: []string{mockServer1.URL, mockServer2.URL}}
    body, _ := json.Marshal(payload)

    req := httptest.NewRequest("POST", "/fetch", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(fetchHandler)

    handler.ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
        t.Errorf("Expected status 200, got: %d", rr.Code)
    }

    var results []Result
    if err := json.Unmarshal(rr.Body.Bytes(), &results); err != nil {
        t.Errorf("Invalid JSON response: %v", err)
    }

    if len(results) != 2 {
        t.Errorf("Expected 2 results, got: %d", len(results))
    }
}

// TestFetchHandler_ResponseHeaders тестирует заголовки ответа
func TestFetchHandler_ResponseHeaders(t *testing.T) {
    mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{"test": "data"}`))
    }))
    defer mockServer.Close()

    payload := RequestPayload{Urls: []string{mockServer.URL}}
    body, _ := json.Marshal(payload)

    req := httptest.NewRequest("POST", "/fetch", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(fetchHandler)

    handler.ServeHTTP(rr, req)

    contentType := rr.Header().Get("Content-Type")
    if contentType != "application/json" {
        t.Errorf("Expected Content-Type application/json, got: %s", contentType)
    }
}