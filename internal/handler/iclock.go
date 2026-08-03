package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"adms-go/internal/iclock"
	"adms-go/internal/store"
	"adms-go/internal/webhook"
)

type IClockHandler struct {
	Store      *store.Store
	Dispatcher *webhook.Dispatcher
}

// Handshake handles GET /iclock/cdata
func (h *IClockHandler) Handshake(w http.ResponseWriter, r *http.Request) {
	sn := r.URL.Query().Get("SN")
	option := r.URL.Query().Get("option")

	logData := fmt.Sprintf("%s %s", r.Method, r.URL.String())
	_ = h.Store.InsertDeviceLog(logData, sn, option, r.URL.String())

	isNew, err := h.Store.UpsertDeviceOnline(sn)
	if err != nil {
		log.Printf("upsert device %q: %v", sn, err)
		_ = h.Store.InsertErrorLog(fmt.Sprintf("upsert device: %v", err))
	}

	// Device events
	if isNew {
		h.Dispatcher.Dispatch("device_register", sn, map[string]interface{}{
			"device_sn": sn,
		})
	}
	h.Dispatcher.Dispatch("device_online", sn, map[string]interface{}{
		"device_sn": sn,
	})

	opStamp := int(time.Now().Unix())
	cfg := handshakeDefaults()

	if config, err := h.Store.GetHandshakeConfig(sn); err == nil && config != nil {
		if config.Stamp > 0 {
			cfg["Stamp"] = strconv.Itoa(config.Stamp)
		}
		cfg["ErrorDelay"] = strconv.Itoa(config.ErrorDelay)
		cfg["Delay"] = strconv.Itoa(config.Delay)
		cfg["ResLogDay"] = strconv.Itoa(config.ResLogDay)
		cfg["ResLogDelCount"] = strconv.Itoa(config.ResLogDelCount)
		cfg["ResLogCount"] = strconv.Itoa(config.ResLogCount)
		cfg["TransTimes"] = config.TransTimes
		cfg["TransInterval"] = strconv.Itoa(config.TransInterval)
		cfg["TransFlag"] = config.TransFlag
		cfg["TimeZone"] = strconv.Itoa(config.TimeZone)
		cfg["Realtime"] = boolToInt(config.Realtime)
		cfg["Encrypt"] = boolToInt(config.Encrypt)
	}

	resp := fmt.Sprintf(
		"GET OPTION FROM: %s\r\n"+
			"Stamp=%s\r\n"+
			"OpStamp=%d\r\n"+
			"ErrorDelay=%s\r\n"+
			"Delay=%s\r\n"+
			"ResLogDay=%s\r\n"+
			"ResLogDelCount=%s\r\n"+
			"ResLogCount=%s\r\n"+
			"TransTimes=%s\r\n"+
			"TransInterval=%s\r\n"+
			"TransFlag=%s\r\n"+
			"TimeZone=%s\r\n"+
			"Realtime=%s\r\n"+
			"Encrypt=%s",
		sn,
		cfg["Stamp"],
		opStamp,
		cfg["ErrorDelay"],
		cfg["Delay"],
		cfg["ResLogDay"],
		cfg["ResLogDelCount"],
		cfg["ResLogCount"],
		cfg["TransTimes"],
		cfg["TransInterval"],
		cfg["TransFlag"],
		cfg["TimeZone"],
		cfg["Realtime"],
		cfg["Encrypt"],
	)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(resp))
}

// ReceiveRecords handles POST /iclock/cdata
func (h *IClockHandler) ReceiveRecords(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := readBody(r, 1<<20)
	if err != nil {
		w.Write([]byte("ERROR: 0\n"))
		return
	}
	bodyStr := string(bodyBytes)

	sn := r.URL.Query().Get("SN")
	table := r.URL.Query().Get("table")
	stamp := r.URL.Query().Get("Stamp")

	allParams := fmt.Sprintf("SN=%s&table=%s&stamp=%s", sn, table, stamp)
	_ = h.Store.InsertFingerLog(allParams+"\n"+bodyStr, r.URL.String())

	if strings.EqualFold(table, "OPERLOG") {
		count := iclock.CountNonEmpty(bodyStr)
		w.Write([]byte(fmt.Sprintf("OK: %d", count)))
		return
	}

	records := iclock.ParseRecords(bodyStr)
	log.Printf("iclock POST: sn=%s table=%s records=%d", sn, table, len(records))
	total := 0

	for _, rec := range records {
		att, err := parseAttendanceRecord(rec, sn, table, stamp)
		if err != nil {
			log.Printf("iclock parse error: %v | rec: %+v", err, rec)
			_ = h.Store.InsertErrorLog(fmt.Sprintf("parse: %v | rec: %+v", err, rec))
			continue
		}

		if err := h.Store.InsertAttendance(att); err != nil {
			log.Printf("iclock insert error: %v | att: %+v", err, att)
			_ = h.Store.InsertErrorLog(fmt.Sprintf("insert: %v | att: %+v", err, att))
			continue
		}
		if att.ID == 0 {
			continue // duplicate, skipped by INSERT IGNORE
		}
		total++

		// Webhook: attendance event
		h.Dispatcher.Dispatch("attendance", sn, map[string]interface{}{
			"employee_id":          att.EmployeeID,
			"attendance_timestamp": att.Timestamp.Format("2006-01-02 15:04:05"),
			"status1":              att.Status1,
			"status2":              att.Status2,
			"status3":              att.Status3,
			"status4":              att.Status4,
			"status5":              att.Status5,
		})
	}

	log.Printf("iclock POST done: sn=%s inserted=%d", sn, total)
	w.Write([]byte(fmt.Sprintf("OK: %d", total)))
}

func (h *IClockHandler) TestHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := readBody(r, 1<<20)
	_ = h.Store.InsertFingerLog(string(body), r.URL.String())
	w.Write([]byte("OK"))
}

func (h *IClockHandler) GetRequestHandler(w http.ResponseWriter, r *http.Request) {
	sn := r.URL.Query().Get("SN")
	if sn != "" {
		if _, err := h.Store.UpsertDeviceOnline(sn); err != nil {
			log.Printf("getrequest upsert %q: %v", sn, err)
		}
	}
	w.Write([]byte("OK"))
}

// ─── helpers ─────────────────────────────────────────────────────

func handshakeDefaults() map[string]string {
	return map[string]string{
		"Stamp":          "9999",
		"ErrorDelay":     "60",
		"Delay":          "30",
		"ResLogDay":      "18250",
		"ResLogDelCount": "10000",
		"ResLogCount":    "50000",
		"TransTimes":     "00:00;14:05",
		"TransInterval":  "1",
		"TransFlag":      "1111000000",
		"TimeZone":       "7",
		"Realtime":       "1",
		"Encrypt":        "0",
	}
}

func boolToInt(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func parseAttendanceRecord(rec iclock.Record, sn, table, stamp string) (*store.Attendance, error) {
	employeeID, err := strconv.Atoi(rec.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("employee_id %q: %w", rec.EmployeeID, err)
	}

	// Try standard format, then without space (some devices omit it)
	ts, err := time.Parse("2006-01-02 15:04:05", rec.Timestamp)
	if err != nil {
		ts, err = time.Parse("2006-01-0215:04:05", rec.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("timestamp %q: %w", rec.Timestamp, err)
		}
	}

	return &store.Attendance{
		SN:         sn,
		TableName:  table,
		Stamp:      stamp,
		EmployeeID: employeeID,
		Timestamp:  ts,
		Status1:    parseIntOrNil(rec.Status1),
		Status2:    parseIntOrNil(rec.Status2),
		Status3:    parseIntOrNil(rec.Status3),
		Status4:    parseIntOrNil(rec.Status4),
		Status5:    parseIntOrNil(rec.Status5),
	}, nil
}

func parseIntOrNil(s string) *int {
	if s == "" {
		return nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &v
}

func readBody(r *http.Request, maxBytes int64) ([]byte, error) {
	r.Body = http.MaxBytesReader(nil, r.Body, maxBytes)
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}
