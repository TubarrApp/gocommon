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

// Write writes the current JSON log line into RAM.
func (mw *memoryWriter) Write(p []byte) (int, error) {
	mw.pl.addToRAMLine(p)

	out := p
	if len(p) > 0 && p[len(p)-1] != '\n' {
		out = append(append([]byte{}, p...), '\n')
	}

	// Write to actual file
	return mw.next.Write(out)
}

// addToRAMLine adds the current line to the log buffer.
func (pl *ProgramLogger) addToRAMLine(p []byte) {
	clean := ansiStripper.ReplaceAll(p, nil)

	pl.LogBufferLock.Lock()
	pl.LogBuffer[pl.LogBufferPos] = append([]byte(nil), clean...)
	pl.LogBufferPos = (pl.LogBufferPos + 1) % len(pl.LogBuffer)
	if pl.LogBufferPos == 0 {
		pl.LogBufferFull = true
	}
	pl.LogBufferLock.Unlock()
}
