package webhook

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// retryDelays for 3 retry attempts
var retryDelays = []time.Duration{1 * time.Second, 5 * time.Second, 25 * time.Second}

// Send posts JSON payload to webhook URL with 3x retry
func Send(url string, payload interface{}) {
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

		resp, err := http.Post(url, "application/json", bytes.NewReader(body))
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
