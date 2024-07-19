package db

import (
	"database/sql"
	"database/sql/driver"
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

func (v *NullString) Scan(value any) error {
	v.Valid = true
	if value == nil {
		v.String = ""
		return nil
	}
	v.String = value.(string)
	return nil
}

func (v *NullString) Value() (driver.Value, error) {
	if !v.Valid {
		return nil, nil
	}
	return v.String, nil
}

type JsonNullInt64 sql.NullInt64

func (v *JsonNullInt64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Int64)
	} else {
		return json.Marshal(nil)
	}
}

func (v *JsonNullInt64) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *int64
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Int64 = *x
	} else {
		v.Valid = false
	}
	return nil
}

type JsonNullInt32 sql.NullInt32

func (v *JsonNullInt32) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Int32)
	} else {
		return json.Marshal(nil)
	}
}

func (v *JsonNullInt32) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *int32
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Int32 = *x
	} else {
		v.Valid = false
	}
	return nil
}
