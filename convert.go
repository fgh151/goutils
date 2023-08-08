package goutils

import (
	"strconv"
)

func String2int(s string) (int, error) {
	return strconv.Atoi(s)
}

func String2Int64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
