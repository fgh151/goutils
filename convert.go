package sdk

import (
	"strconv"
)

func String2Int(s string) (int, error) {
	return strconv.Atoi(s)
}

func String2Int64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
