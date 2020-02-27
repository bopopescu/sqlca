package sqlca

import (
	"fmt"
	"github.com/civet148/gotools/log"
	"reflect"
	"runtime"
	"strings"
)

//assert string or struct/slice/map nil (not include decimal type)
func assertNil(v interface{}, strMsg string, args ...interface{}) {
	if isNil(v) {
		panic(fmt.Sprintf(strMsg, args...))
	}
}

func isNil(v interface{}) bool {
	switch v.(type) {
	case string:
		if v.(string) != "" {
			return true
		}
	default:
		if !reflect.ValueOf(v).IsNil() {
			return true
		}
	}
	return false
}

// get function name from call stack
func getFuncName(pc uintptr) (name string) {

	n := runtime.FuncForPC(pc).Name()
	ns := strings.Split(n, ".")
	name = ns[len(ns)-1]
	return
}

func (e *Engine) setDebug(ok bool) {
	e.debug = ok
}

func (e *Engine) isDebug() bool {
	return e.debug
}

func (e *Engine) panic(strFmt string, args ...interface{}) {
	if e.isDebug() {
		panic(fmt.Sprintf(strFmt, args...))
	} else {
		pc, _, _, ok := runtime.Caller(1)
		if ok {
			strFmt = getFuncName(pc) + ": " + strFmt
		}
		log.Fatalf(strFmt, args...)
	}
}

func (e *Engine) debugf(strFmt string, args ...interface{}) {
	if e.isDebug() {
		pc, _, _, ok := runtime.Caller(1)
		if ok {
			strFmt = getFuncName(pc) + ": " + strFmt
		}
		log.Debugf(strFmt, args...)
	}
}
