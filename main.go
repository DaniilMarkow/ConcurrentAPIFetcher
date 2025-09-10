package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type RequestPayload struct {
	Urls []string `json:"urls"`
}

type Result struct {
	URL  string `json:"url"`
	Data string `json:"data"`
	Err  string `json:"error"`
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Concurrent API Fetcher Server is up and running!")
}

func fetchURL(ctx context.Context, url string, ch chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		ch <- Result{URL: url, Err: err.Error()}
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ch <- Result{URL: url, Err: err.Error()}
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- Result{URL: url, Err: err.Error()}
		return
	}
	ch <- Result{URL: url, Data: string(body)}
}

func fetchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	var request RequestPayload
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(request.Urls) == 0 {
		http.Error(w, "No URLs provided", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	resultsCh := make(chan Result, len(request.Urls))

	for _, url := range request.Urls {
		wg.Add(1)
		go fetchURL(ctx, url, resultsCh, &wg)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var results []Result
	for result := range resultsCh {
		results = append(results, result)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/fetch", fetchHandler)

	port := ":8080"
	fmt.Println("Server starting on :8080")
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
