package handler

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"adms-go/internal/store"
)

type DashboardHandler struct {
	Store    *store.Store
	Template *template.Template
}

func NewDashboardHandler(s *store.Store, t *template.Template) *DashboardHandler {
	return &DashboardHandler{Store: s, Template: t}
}

func (h *DashboardHandler) Devices(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1)
	devices, err := h.Store.GetDevices(page)
	if err != nil {
		log.Printf("devices query: %v", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	total, _ := h.Store.CountDevices()

	data := map[string]interface{}{
		"Title":   "Devices",
		"Active":  "devices",
		"Devices": devices,
		"Page":    store.CalcPage(page, total),
		"IsOnline": func(d store.Device) string {
			if d.IsOnline() {
				return "Online"
			}
			return "Offline"
		},
		"IsOnlineBadge": func(d store.Device) template.HTML {
			if d.IsOnline() {
				return `<span class="badge bg-success">Online</span>`
			}
			return `<span class="badge bg-secondary">Offline</span>`
		},
	}
	h.Template.ExecuteTemplate(w, "devices_index", data)
}

func (h *DashboardHandler) DeviceLog(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1)
	logs, err := h.Store.GetDeviceLogs(page)
	if err != nil {
		log.Printf("device_log query: %v", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	total, _ := h.Store.CountDeviceLogs()

	data := map[string]interface{}{
		"Title":  "Device Log",
		"Active": "devices-log",
		"Logs":   logs,
		"Page":   store.CalcPage(page, total),
	}
	h.Template.ExecuteTemplate(w, "devices_log", data)
}

func (h *DashboardHandler) FingerLog(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1)
	logs, err := h.Store.GetFingerLogs(page)
	if err != nil {
		log.Printf("finger_log query: %v", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	total, _ := h.Store.CountFingerLogs()

	data := map[string]interface{}{
		"Title":  "Finger Log",
		"Active": "finger-log",
		"Logs":   logs,
		"Page":   store.CalcPage(page, total),
	}
	h.Template.ExecuteTemplate(w, "finger_log", data)
}

func (h *DashboardHandler) Attendance(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1)
	sn := r.URL.Query().Get("sn")
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	atts, err := h.Store.GetAttendances(sn, start, end, page)
	if err != nil {
		log.Printf("attendances query: %v", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	total, _ := h.Store.CountAttendances(sn, start, end)
	log.Printf("attendance: sn=%q start=%q end=%q page=%d rows=%d total=%d", sn, start, end, page, len(atts), total)

	devices, err := h.Store.GetDeviceList()
	if err != nil {
		log.Printf("device list query: %v", err)
		devices = []store.Device{}
	}

	data := map[string]interface{}{
		"Title":       "Attendance",
		"Active":      "attendance",
		"Attendances": atts,
		"Devices":     devices,
		"Page":        store.CalcPage(page, total),
		"Filters": map[string]string{
			"sn":    sn,
			"start": start,
			"end":   end,
		},
	}
	if err := h.Template.ExecuteTemplate(w, "devices_attendance", data); err != nil {
		log.Printf("template devices_attendance: %v", err)
	}
}

func (h *DashboardHandler) Webhooks(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":  "Webhooks",
		"Active": "webhooks",
	}
	h.Template.ExecuteTemplate(w, "webhooks", data)
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return defaultVal
	}
	return n
}
