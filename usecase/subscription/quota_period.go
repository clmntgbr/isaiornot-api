package subscription

import "time"

// CurrentQuotaPeriod returns the monthly quota window that contains now,
// based on an anniversary anchor (QuotaPeriodStart).
//
// Example: anchor = July 10, now = July 15 → [July 10, August 10)
// (i.e. July 10 through August 9 inclusive).
func CurrentQuotaPeriod(anchor, now time.Time) (start, end time.Time) {
	now = now.UTC()
	anchor = anchor.UTC()

	if anchor.IsZero() {
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return start, start.AddDate(0, 1, 0)
	}

	if now.Before(anchor) {
		return anchor, anchor.AddDate(0, 1, 0)
	}

	start = anchor
	for {
		next := start.AddDate(0, 1, 0)
		if next.After(now) {
			return start, next
		}
		start = next
	}
}
