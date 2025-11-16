package logging

import (
	"path/filepath"
	"runtime"
	"strconv"
)

// callerInfo retrieves caller information for logging.
type callerInfo struct {
	funcName string
	file     string
	line     int
	lineStr  string
}

// getCaller gets caller information from the call stack.
func getCaller(skip int) callerInfo {
	pc, file, line, _ := runtime.Caller(skip)
	return callerInfo{
		funcName: filepath.Base(runtime.FuncForPC(pc).Name()),
		file:     filepath.Base(file),
		line:     line,
		lineStr:  strconv.Itoa(line),
	}
}
