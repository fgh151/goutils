package sdk

import (
	"errors"
	"os"
	"strconv"
)

func GetenvStr(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return v, errors.New("getenv: environment variable empty")
	}
	return v, nil
}

func GetenvInt(key string) (int, error) {
	s, err := GetenvStr(key)
	if err != nil {
		return 0, err
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return v, nil
}
