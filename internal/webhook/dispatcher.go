package webhook

import (
	"log"
	"time"

	"adms-go/internal/store"
)

// Payload is the JSON body sent to webhook endpoints
type Payload struct {
	Event     string      `json:"event"`
	DeviceSN  string      `json:"device_sn"`
	Timestamp string      `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// Dispatcher handles webhook dispatch
type Dispatcher struct {
	Store *store.Store
}

func NewDispatcher(s *store.Store) *Dispatcher {
	return &Dispatcher{Store: s}
}

// Dispatch sends webhook payloads for a given event+device. Non-blocking.
func (d *Dispatcher) Dispatch(event, deviceSN string, data interface{}) {
	whs, err := d.Store.GetWebhooks(event, deviceSN)
	if err != nil {
		log.Printf("webhook dispatch: query error: %v", err)
		return
	}

	if len(whs) == 0 {
		return
	}

	payload := Payload{
		Event:     event,
		DeviceSN:  deviceSN,
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      data,
	}

	for _, wh := range whs {
		wh := wh // capture
		go func() {
			Send(wh.URL, payload, wh.Headers)
		}()
	}
}
