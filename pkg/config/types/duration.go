package types

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}

func (d *Duration) Decode(value string) error {
	var err error
	d.Duration, err = time.ParseDuration(value)
	if err != nil {
		return err
	}
	return nil
}

func (d *Duration) Type() string { return "duration" }

func (d *Duration) Set(s string) error {
	return d.UnmarshalJSON([]byte(s))
}
