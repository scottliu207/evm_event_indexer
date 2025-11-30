package utils

import (
	"reflect"
	"runtime"
)

func GetFuncName(i any) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
