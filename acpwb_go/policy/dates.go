package policy

import "time"

// ymd is a plain (year, month, day) tuple used for the date arithmetic in
// generate_related_links, avoiding any dependency on Go's time.Time zero-
// value/range quirks for very old/future years.
type ymd struct {
	y, m, d int
}

// mkTime builds a UTC time.Time at midnight for (y, m, d), used only as an
// intermediate for AddDate's calendar-correct day arithmetic (matching
// Python's datetime.date + timedelta(days=...) semantics).
func mkTime(y, m, d int) time.Time {
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

// addDays adds (possibly negative) days to a base time.Time and returns the
// resulting (year, month, day), mirroring Python's date + timedelta(days=n).
func addDays(t time.Time, days int64) ymd {
	t2 := t.AddDate(0, 0, int(days))
	y, m, d := t2.Date()
	return ymd{y, int(m), d}
}
