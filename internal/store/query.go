package store

import (
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
	// Check if device exists first
	var count int
	if err := s.Adms.Get(&count, "SELECT COUNT(*) FROM devices WHERE no_sn = ?", sn); err != nil {
		return false, err
	}
	isNew = count == 0

	_, err = s.Adms.Exec(
		"INSERT INTO devices (no_sn, online, created_at, updated_at) VALUES (?, NOW(), NOW(), NOW()) ON DUPLICATE KEY UPDATE online = NOW(), updated_at = NOW()",
		sn)
	return isNew, err
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

func (s *Store) GetAttendances(page int) ([]Attendance, error) {
	offset := (page - 1) * 15
	var atts []Attendance
	err := s.Adms.Select(&atts,
		"SELECT id, sn, `table`, stamp, employee_id, timestamp, status1, status2, status3, status4, status5, created_at, updated_at FROM attendances ORDER BY id DESC LIMIT 15 OFFSET ?",
		offset)
	return atts, err
}

func (s *Store) CountAttendances() (int, error) {
	var count int
	err := s.Adms.Get(&count, "SELECT COUNT(*) FROM attendances")
	return count, err
}

func (s *Store) InsertAttendance(att *Attendance) error {
	result, err := s.Adms.Exec(
		"INSERT INTO attendances (sn, `table`, stamp, employee_id, timestamp, status1, status2, status3, status4, status5, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())",
		att.SN, att.TableName, att.Stamp, att.EmployeeID, att.Timestamp.Time, att.Status1, att.Status2, att.Status3, att.Status4, att.Status5)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	att.ID = id
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
		"SELECT id, device_sn, name, url, event, is_active, created_at, updated_at FROM webhooks WHERE event = ? AND (device_sn = ? OR device_sn = '') AND is_active = 1",
		event, deviceSN)
	return whs, err
}

func (s *Store) GetAllWebhooks(deviceSN string) ([]Webhook, error) {
	var whs []Webhook
	var err error
	if deviceSN == "" {
		err = s.Adms.Select(&whs, "SELECT id, device_sn, name, url, event, is_active, created_at, updated_at FROM webhooks ORDER BY device_sn, event")
	} else {
		err = s.Adms.Select(&whs, "SELECT id, device_sn, name, url, event, is_active, created_at, updated_at FROM webhooks WHERE device_sn = ? OR device_sn = '' ORDER BY device_sn, event", deviceSN)
	}
	return whs, err
}

func (s *Store) CreateWebhook(w *Webhook) error {
	result, err := s.Adms.Exec(
		"INSERT INTO webhooks (device_sn, name, url, event, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, NOW(), NOW())",
		w.DeviceSN, w.Name, w.URL, w.Event, w.IsActive)
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
}

func CalcPage(page, total int) PageInfo {
	totalPages := total / 15
	if total%15 > 0 {
		totalPages++
	}
	if totalPages < 1 {
		totalPages = 1
	}
	return PageInfo{
		CurrentPage: page,
		TotalPages:  totalPages,
		HasPrev:     page > 1,
		HasNext:     page < totalPages,
		PrevPage:    page - 1,
		NextPage:    page + 1,
	}
}
