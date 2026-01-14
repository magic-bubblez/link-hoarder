package main

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	url := "http://localhost:8080/bubbles"
	concurrency := 1000

	var wg sync.WaitGroup
	wg.Add(concurrency)

	fmt.Printf("Launching %d concurrent requests against %s...\n", concurrency, url)
	start := time.Now()

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			jsonBody := []byte(`{"name":"Chaos Bubble"}`)

			resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
			if err != nil {
				fmt.Printf("💥 Attacker %d failed: %v\n", id, err)
				return
			}
			defer resp.Body.Close()

			//everything good
			if resp.StatusCode != 200 {
				fmt.Printf("Server Error: %d\n", resp.StatusCode)
			}
		}(i)
	}

	wg.Wait()
	fmt.Printf("🏁 Attack finished in %v\n", time.Since(start))
}
