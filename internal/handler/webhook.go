package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"adms-go/internal/store"
)

type WebhookHandler struct {
	Store *store.Store
}

// ListWebhooks handles GET /api/webhooks?device_sn=X
func (h *WebhookHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	deviceSN := r.URL.Query().Get("device_sn")
	whs, err := h.Store.GetAllWebhooks(deviceSN)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]interface{}{"webhooks": whs})
}

// CreateWebhook handles POST /api/webhooks
func (h *WebhookHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	var input struct {
		DeviceSN string `json:"device_sn"`
		Name     string `json:"name"`
		URL      string `json:"url"`
		Event    string `json:"event"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	if input.URL == "" || input.Event == "" {
		writeJSON(w, 400, map[string]string{"error": "url and event are required"})
		return
	}

	wh := &store.Webhook{
		DeviceSN: input.DeviceSN,
		Name:     input.Name,
		URL:      input.URL,
		Event:    input.Event,
		IsActive: true,
	}

	if err := h.Store.CreateWebhook(wh); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, 201, wh)
}

// DeleteWebhook handles DELETE /api/webhooks/{id}
func (h *WebhookHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}

	if err := h.Store.DeleteWebhook(id); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
