package db

import (
	"database/sql"
	"encoding/json"
	"time"
)

const (
	datetime  = "2006-01-02 15:04"
	date      = "2006-01-02"
	timeOfDay = "15:04"
)

type NullTime sql.NullTime

func (v *NullTime) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Time.Format(datetime))
	} else {
		return json.Marshal(nil)
	}
}

func (v *NullTime) UnmarshalJSON(b []byte) error {
	var x string
	if err := json.Unmarshal(b, &x); err != nil {
		return err
	}

	t, err := time.Parse(datetime, x)
	if err == nil {
		v.Valid = true
		v.Time = t
	} else {
		return err
	}

	return nil
}

type NullString sql.NullString

func (v *NullString) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.String)
	} else {
		return json.Marshal(nil)
	}
}
