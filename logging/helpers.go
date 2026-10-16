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
	pl     *ProgramLogger
	writer io.Writer
}

// Write writes the current JSON log line into RAM.
func (mw *memoryWriter) Write(p []byte) (int, error) {
	// Add to RAM.
	mw.pl.addToRAMLine(p)

	// Terminate entries with newlines.
	out := p
	if len(p) > 0 && p[len(p)-1] != '\n' {
		out = append(append([]byte{}, p...), '\n')
	}

	// Write to log file.
	if _, err := mw.writer.Write(out); err != nil {
		return 0, err
	}
	return len(p), nil
}

// addToRAMLine adds the current line to the log buffer.
func (pl *ProgramLogger) addToRAMLine(p []byte) {
	// Remove ANSI from line.
	clean := ansiStripper.ReplaceAll(p, nil)

	// Update ring buffer.
	pl.mu.Lock()
	pl.logBuffer[pl.total%uint64(len(pl.logBuffer))] = append([]byte(nil), clean...)
	pl.total++
	pl.mu.Unlock()
}
