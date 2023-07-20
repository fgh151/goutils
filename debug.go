package goutils

import "fmt"

func DumpObject(obj interface{}) string {
	return fmt.Sprintf("%+v\n", obj)
}
