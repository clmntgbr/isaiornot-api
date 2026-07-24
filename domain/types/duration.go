package types

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type Duration time.Duration

func NewDuration(d time.Duration) Duration {
	return Duration(d)
}

func (d Duration) ToDuration() time.Duration {
	return time.Duration(d)
}

func (d Duration) Value() (driver.Value, error) {
	return int64(d), nil
}

func (d *Duration) Scan(value interface{}) error {
	if value == nil {
		*d = 0
		return nil
	}
	switch v := value.(type) {
	case int64:
		*d = Duration(v)
	case int32:
		*d = Duration(int64(v))
	case float64:
		*d = Duration(int64(v))
	default:
		return fmt.Errorf("cannot scan type %T into types.Duration", value)
	}
	return nil
}
