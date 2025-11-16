package logging

import (
	"io"
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

// memoryWriter writes JSON log output into RAM then forwards to the real writer.
type memoryWriter struct {
	pl   *ProgramLogger
	next io.Writer
}

func (mw *memoryWriter) Write(p []byte) (int, error) {
	// p contains the FINAL zerolog JSON log entry
	mw.pl.AddToMemoryLog(p)

	if mw.next != nil {
		return mw.next.Write(p)
	}
	return len(p), nil
}
