package store

import (
	"log"
)

// ─── Device queries ──────────────────────────────────────────────

func (s *Store) GetDevices(page int) ([]Device, error) {
	offset := (page - 1) * 15
	var devices []Device
	err := s.Adms.Select(&devices,
		"SELECT id, nama, no_sn, lokasi, online, created_at, updated_at FROM devices ORDER BY online DESC LIMIT 15 OFFSET ?",
		offset)
	return devices, err
}

func (s *Store) CountDevices() (int, error) {
	var count int
	err := s.Adms.Get(&count, "SELECT COUNT(*) FROM devices")
	return count, err
}

func (s *Store) UpsertDeviceOnline(sn string) (isNew bool, err error) {
	// Check if device already exists
	var count int
	if err := s.Adms.Get(&count, "SELECT COUNT(*) FROM devices WHERE no_sn = ?", sn); err != nil {
		log.Printf("store: upsert device %q SELECT error: %v", sn, err)
		return false, err
	}
	isNew = count == 0
	log.Printf("store: upsert device %q existing=%d isNew=%v", sn, count, isNew)

	// Upsert: INSERT if new, UPDATE online if exists.
	// IMPORTANT: requires UNIQUE index on no_sn column for ON DUPLICATE KEY to work.
	// Run: ALTER TABLE devices ADD UNIQUE INDEX idx_no_sn (no_sn);
	result, err := s.Adms.Exec(
		"INSERT INTO devices (no_sn, online, created_at, updated_at) VALUES (?, NOW(), NOW(), NOW()) ON DUPLICATE KEY UPDATE online = NOW(), updated_at = NOW()",
		sn)
	if err != nil {
		log.Printf("store: upsert device %q INSERT error: %v", sn, err)
		return isNew, err
	}
	rows, _ := result.RowsAffected()
	log.Printf("store: upsert device %q rowsAffected=%d", sn, rows)
	return isNew, nil
}

// ─── Device log queries ──────────────────────────────────────────

func (s *Store) GetDeviceLogs(page int) ([]DeviceLog, error) {
	offset := (page - 1) * 15
	var logs []DeviceLog
	err := s.Adms.Select(&logs,
		"SELECT id, data, tgl, sn, `option`, url, created_at, updated_at FROM device_log ORDER BY id DESC LIMIT 15 OFFSET ?",
		offset)
	return logs, err
}

func (s *Store) CountDeviceLogs() (int, error) {
	var count int
	err := s.Adms.Get(&count, "SELECT COUNT(*) FROM device_log")
	return count, err
}

func (s *Store) InsertDeviceLog(data string, sn string, option string, url string) error {
	_, err := s.Adms.Exec(
		"INSERT INTO device_log (data, sn, `option`, url) VALUES (?, ?, ?, ?)",
		data, sn, option, url)
	if err != nil {
		log.Printf("store: insert device_log sn=%q error: %v", sn, err)
	}
	return err
}

// ─── Finger log queries ──────────────────────────────────────────

func (s *Store) GetFingerLogs(page int) ([]FingerLog, error) {
	offset := (page - 1) * 15
	var logs []FingerLog
	err := s.Adms.Select(&logs,
		"SELECT id, data, url, created_at, updated_at FROM finger_log ORDER BY id DESC LIMIT 15 OFFSET ?",
		offset)
	return logs, err
}

func (s *Store) CountFingerLogs() (int, error) {
	var count int
	err := s.Adms.Get(&count, "SELECT COUNT(*) FROM finger_log")
	return count, err
}

func (s *Store) InsertFingerLog(data string, url string) error {
	_, err := s.Adms.Exec(
		"INSERT INTO finger_log (data, url) VALUES (?, ?)",
		data, url)
	if err != nil {
		log.Printf("store: insert finger_log error: %v", err)
	}
	return err
}

// ─── Error log queries ───────────────────────────────────────────

func (s *Store) InsertErrorLog(data string) error {
	_, err := s.Adms.Exec(
		"INSERT INTO error_log (data, created_at, updated_at) VALUES (?, NOW(), NOW())",
		data)
	return err
}

// ─── Attendance queries ──────────────────────────────────────────

func (s *Store) GetAttendances(sn, start, end string, page int) ([]Attendance, error) {
	offset := (page - 1) * 15
	var atts []Attendance
	where := "1=1"
	args := make([]interface{}, 0)

	if sn != "" {
		where += " AND sn = ?"
		args = append(args, sn)
	}
	if start != "" {
		where += " AND timestamp >= ?"
		args = append(args, start)
	}
	if end != "" {
		where += " AND timestamp <= ?"
		args = append(args, end+" 23:59:59")
	}

	query := "SELECT id, sn, `table`, stamp, employee_id, timestamp, status1, status2, status3, status4, status5, created_at, updated_at FROM attendances WHERE " + where + " ORDER BY id DESC LIMIT 15 OFFSET ?"
	args = append(args, offset)

	err := s.Adms.Select(&atts, query, args...)
	return atts, err
}

func (s *Store) CountAttendances(sn, start, end string) (int, error) {
	var count int
	where := "1=1"
	args := make([]interface{}, 0)

	if sn != "" {
		where += " AND sn = ?"
		args = append(args, sn)
	}
	if start != "" {
		where += " AND timestamp >= ?"
		args = append(args, start)
	}
	if end != "" {
		where += " AND timestamp <= ?"
		args = append(args, end+" 23:59:59")
	}

	err := s.Adms.Get(&count, "SELECT COUNT(*) FROM attendances WHERE "+where, args...)
	return count, err
}

// GetDeviceList returns distinct device SNs from attendance records for filter dropdown
func (s *Store) GetDeviceList() ([]Device, error) {
	var devices []Device
	err := s.Adms.Select(&devices, "SELECT DISTINCT sn AS no_sn FROM attendances ORDER BY sn")
	return devices, err
}

func (s *Store) InsertAttendance(att *Attendance) error {
	result, err := s.Adms.Exec(
		"INSERT IGNORE INTO attendances (sn, `table`, stamp, employee_id, timestamp, status1, status2, status3, status4, status5, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())",
		att.SN, att.TableName, att.Stamp, att.EmployeeID, att.Timestamp, att.Status1, att.Status2, att.Status3, att.Status4, att.Status5)
	if err != nil {
		log.Printf("store: insert attendance sn=%q emp=%d error: %v", att.SN, att.EmployeeID, err)
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		log.Printf("store: insert attendance skipped (duplicate) sn=%q emp=%d ts=%s", att.SN, att.EmployeeID, att.Timestamp.Format("2006-01-02 15:04:05"))
		return nil
	}
	id, _ := result.LastInsertId()
	att.ID = id
	log.Printf("store: insert attendance ok sn=%q emp=%d id=%d", att.SN, att.EmployeeID, id)
	return nil
}

// Checklog determines attendance type from status values
func ChecklogFromStatus(status1, status2 *int) *string {
	if status1 == nil {
		return nil
	}
	status2OK := status2 != nil && (*status2 == 0 || *status2 == 1)

	switch *status1 {
	case 0:
		if status2OK {
			s := "IN"
			return &s
		}
	case 1:
		if status2OK {
			s := "OUT"
			return &s
		}
	case 4:
		if status2OK {
			s := "LEMBUR IN"
			return &s
		}
	case 5:
		if status2OK {
			s := "LEMBUR OUT"
			return &s
		}
	}
	return nil
}

// ─── Webhook queries ─────────────────────────────────────────────

func (s *Store) GetWebhooks(event, deviceSN string) ([]Webhook, error) {
	var whs []Webhook
	// Match: (global webhooks with device_sn='') OR (device-specific webhooks)
	err := s.Adms.Select(&whs,
		"SELECT id, device_sn, name, url, COALESCE(headers,'') as headers, event, is_active, created_at, updated_at FROM webhooks WHERE event = ? AND (device_sn = ? OR device_sn = '') AND is_active = 1",
		event, deviceSN)
	return whs, err
}

func (s *Store) GetAllWebhooks(deviceSN string) ([]Webhook, error) {
	var whs []Webhook
	var err error
	if deviceSN == "" {
		err = s.Adms.Select(&whs, "SELECT id, device_sn, name, url, COALESCE(headers,'') as headers, event, is_active, created_at, updated_at FROM webhooks ORDER BY device_sn, event")
	} else {
		err = s.Adms.Select(&whs, "SELECT id, device_sn, name, url, COALESCE(headers,'') as headers, event, is_active, created_at, updated_at FROM webhooks WHERE device_sn = ? OR device_sn = '' ORDER BY device_sn, event", deviceSN)
	}
	return whs, err
}

func (s *Store) CreateWebhook(w *Webhook) error {
	result, err := s.Adms.Exec(
		"INSERT INTO webhooks (device_sn, name, url, headers, event, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())",
		w.DeviceSN, w.Name, w.URL, w.Headers, w.Event, w.IsActive)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	w.ID = id
	return nil
}

func (s *Store) DeleteWebhook(id int64) error {
	_, err := s.Adms.Exec("DELETE FROM webhooks WHERE id = ?", id)
	return err
}

// ─── Handshake config queries ────────────────────────────────────

func (s *Store) GetHandshakeConfig(deviceType string) (*HandshakeConfig, error) {
	var cfg HandshakeConfig
	err := s.Adms.Get(&cfg,
		"SELECT id, device_type, stamp, error_delay, delay, res_log_day, res_log_del_count, res_log_count, trans_times, trans_interval, trans_flag, time_zone, realtime, encrypt FROM device_handshake_configs WHERE device_type = ? LIMIT 1",
		deviceType)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ─── Pagination helpers ──────────────────────────────────────────

type PageInfo struct {
	CurrentPage int
	TotalPages  int
	HasPrev     bool
	HasNext     bool
	PrevPage    int
	NextPage    int
	Pages       []PageNum
}

type PageNum struct {
	Num     int
	Current bool
}

func CalcPage(page, total int) PageInfo {
	totalPages := total / 15
	if total%15 > 0 {
		totalPages++
	}
	if totalPages < 1 {
		totalPages = 1
	}

	var pages []PageNum
	add := func(n int, cur bool) {
		pages = append(pages, PageNum{Num: n, Current: cur})
	}
	addDots := func() { pages = append(pages, PageNum{Num: -1}) }

	// Build window: first, last, +-2 around current
	windowStart := page - 2
	if windowStart < 1 {
		windowStart = 1
	}
	windowEnd := page + 2
	if windowEnd > totalPages {
		windowEnd = totalPages
	}

	shown := make(map[int]bool)
	shown[1] = true
	shown[totalPages] = true
	for i := windowStart; i <= windowEnd; i++ {
		shown[i] = true
	}

	prev := 0
	for i := 1; i <= totalPages; i++ {
		if !shown[i] {
			continue
		}
		if prev > 0 && i > prev+1 {
			addDots()
		}
		add(i, i == page)
		prev = i
	}

	return PageInfo{
		CurrentPage: page,
		TotalPages:  totalPages,
		HasPrev:     page > 1,
		HasNext:     page < totalPages,
		PrevPage:    page - 1,
		NextPage:    page + 1,
		Pages:       pages,
	}
}
