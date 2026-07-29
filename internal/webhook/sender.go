package webhook

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

var retryDelays = []time.Duration{1 * time.Second, 5 * time.Second, 25 * time.Second}

// Send posts JSON payload to webhook URL with optional custom headers and 3x retry
func Send(url string, payload interface{}, headers string) {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("webhook marshal: %v", err)
		return
	}

	for attempt := 0; attempt <= len(retryDelays); attempt++ {
		if attempt > 0 {
			delay := retryDelays[attempt-1]
			log.Printf("webhook retry %d/%d for %s (waiting %v)", attempt, len(retryDelays), url, delay)
			time.Sleep(delay)
		}

		req, err := http.NewRequest("POST", url, bytes.NewReader(body))
		if err != nil {
			log.Printf("webhook attempt %d failed (request): %v", attempt+1, err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		// Parse custom headers (key:value per line, or key=value)
		for _, line := range strings.Split(headers, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var k, v string
			if idx := strings.Index(line, ":"); idx >= 0 {
				k = strings.TrimSpace(line[:idx])
				v = strings.TrimSpace(line[idx+1:])
			} else if idx := strings.Index(line, "="); idx >= 0 {
				k = strings.TrimSpace(line[:idx])
				v = strings.TrimSpace(line[idx+1:])
			} else {
				continue
			}
			if k != "" && v != "" {
				req.Header.Set(k, v)
			}
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("webhook attempt %d failed (http): %v", attempt+1, err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Printf("webhook OK: %s → %s (%d)", url, http.StatusText(resp.StatusCode), resp.StatusCode)
			return
		}

		log.Printf("webhook attempt %d returned %d for %s", attempt+1, resp.StatusCode, url)
	}

	log.Printf("webhook FAILED after %d retries: %s", len(retryDelays)+1, url)
}
