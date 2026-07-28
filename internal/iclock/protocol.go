package iclock

import (
	"strings"
)

// Record represents a parsed attendance record from ZKTeco device
type Record struct {
	EmployeeID string
	Timestamp  string
	Status1    string
	Status2    string
	Status3    string
	Status4    string
	Status5    string
}

// ParseRecords parses the plain-text body from a ZKTeco push POST
func ParseRecords(body string) []Record {
	lines := splitLines(body)
	var records []Record

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 2 {
			continue
		}

		r := Record{
			EmployeeID: fields[0],
			Timestamp:  fields[1],
		}
		if len(fields) > 2 {
			r.Status1 = fields[2]
		}
		if len(fields) > 3 {
			r.Status2 = fields[3]
		}
		if len(fields) > 4 {
			r.Status3 = fields[4]
		}
		if len(fields) > 5 {
			r.Status4 = fields[5]
		}
		if len(fields) > 6 {
			r.Status5 = fields[6]
		}
		records = append(records, r)
	}
	return records
}

// CountNonEmpty returns count of non-empty lines in body
func CountNonEmpty(body string) int {
	lines := splitLines(body)
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

// splitLines splits by \r\n, \r, or \n
func splitLines(s string) []string {
	// Normalize: \r\n → \n, \r → \n
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.Split(s, "\n")
}
