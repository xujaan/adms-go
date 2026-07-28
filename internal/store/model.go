package store

import "time"

// NullTime handles MySQL nullable datetime. Use it for columns that may be NULL
// and may arrive as []uint8 when parseTime is disabled in DSN.
type NullTime struct {
	Time  time.Time
	Valid bool
}

// Scan implements sql.Scanner — handles nil, time.Time, []uint8, string
func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Valid = false
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		nt.Time = v
		nt.Valid = true
	case []uint8:
		t, err := time.ParseInLocation("2006-01-02 15:04:05", string(v), time.Local)
		if err != nil {
			return err
		}
		nt.Time = t
		nt.Valid = true
	case string:
		t, err := time.ParseInLocation("2006-01-02 15:04:05", v, time.Local)
		if err != nil {
			return err
		}
		nt.Time = t
		nt.Valid = true
	}
	return nil
}

// Device represents a ZKTeco device
type Device struct {
	ID        int64    `db:"id"`
	Nama      *string  `db:"nama"`
	NoSN      string   `db:"no_sn"`
	Lokasi    *string  `db:"lokasi"`
	Online    NullTime `db:"online"`
	CreatedAt NullTime `db:"created_at"`
	UpdatedAt NullTime `db:"updated_at"`
}

// IsOnline checks if device was active within last 5 minutes
func (d Device) IsOnline() bool {
	if !d.Online.Valid || d.Online.Time.IsZero() {
		return false
	}
	return d.Online.Time.After(time.Now().Add(-5 * time.Minute))
}

// DeviceLog represents a raw handshake request log
type DeviceLog struct {
	ID        int64    `db:"id"`
	Data      string   `db:"data"`
	Tgl       *string  `db:"tgl"`
	SN        string   `db:"sn"`
	Option    *string  `db:"option"`
	URL       *string  `db:"url"`
	CreatedAt NullTime `db:"created_at"`
	UpdatedAt NullTime `db:"updated_at"`
}

// FingerLog represents a raw attendance payload log
type FingerLog struct {
	ID        int64    `db:"id"`
	Data      string   `db:"data"`
	URL       string   `db:"url"`
	CreatedAt NullTime `db:"created_at"`
	UpdatedAt NullTime `db:"updated_at"`
}

// ErrorLog represents a processing error
type ErrorLog struct {
	ID        int64     `db:"id"`
	Data      string    `db:"data"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Attendance represents a parsed attendance record
type Attendance struct {
	ID         int64    `db:"id"`
	SN         string   `db:"sn"`
	TableName  string   `db:"table"`
	Stamp      string   `db:"stamp"`
	EmployeeID int      `db:"employee_id"`
	Timestamp  NullTime `db:"timestamp"`
	Status1    *int     `db:"status1"`
	Status2    *int     `db:"status2"`
	Status3    *int     `db:"status3"`
	Status4    *int     `db:"status4"`
	Status5    *int     `db:"status5"`
	CreatedAt  NullTime `db:"created_at"`
	UpdatedAt  NullTime `db:"updated_at"`
}

// Webhook represents a webhook subscription
type Webhook struct {
	ID        int64     `db:"id"`
	DeviceSN  string    `db:"device_sn"`
	Name      string    `db:"name"`
	URL       string    `db:"url"`
	Event     string    `db:"event"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// HandshakeConfig represents per-device handshake configuration
type HandshakeConfig struct {
	ID             int64  `db:"id"`
	DeviceType     string `db:"device_type"`
	Stamp          int    `db:"stamp"`
	ErrorDelay     int    `db:"error_delay"`
	Delay          int    `db:"delay"`
	ResLogDay      int    `db:"res_log_day"`
	ResLogDelCount int    `db:"res_log_del_count"`
	ResLogCount    int    `db:"res_log_count"`
	TransTimes     string `db:"trans_times"`
	TransInterval  int    `db:"trans_interval"`
	TransFlag      string `db:"trans_flag"`
	TimeZone       int    `db:"time_zone"`
	Realtime       bool   `db:"realtime"`
	Encrypt        bool   `db:"encrypt"`
}
