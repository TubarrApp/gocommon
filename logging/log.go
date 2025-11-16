// Package logging handles the printing and writing of debug and log messages.
package logging

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedregex"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Global logging variables.
var (
	Level    = -1
	Loggable = false
)

// Log array.
var (
	LogAccessMap sync.Map
)

// Log buffer vars.
const logBufferSize = 5000

// ProgramLogger holds logging state for a specific program instance.
type ProgramLogger struct {
	FileLogger    zerolog.Logger
	LogBuffer     [][]byte
	LogBufferLock sync.RWMutex
	LogBufferPos  int
	LogBufferFull bool
	Program       string
	Console       io.Writer
}

// Log entry constants.
const (
	timeFormat = "01/02 15:04:05"

	tagFunc = "[" + sharedconsts.ColorDimCyan + "Function:" + sharedconsts.ColorReset + " "
	tagFile = " - " + sharedconsts.ColorDimCyan + "File:" + sharedconsts.ColorReset + " "
	tagLine = " : " + sharedconsts.ColorDimCyan + "Line:" + sharedconsts.ColorReset + " "
	tagEnd  = "]\n"

	jFunction = "function"
	jFile     = "file"
	jLine     = "line"
)

// logLevel represents different logging levels.
type logType int

const (
	logError logType = iota
	logWarn
	logInfo
	logDebug
	logSuccess
	logPrint
)

// Logging package level variables.
var ansiStripper = sharedregex.AnsiEscapeCompile()

// init zerolog time format.
func init() {
	zerolog.TimeFieldFormat = time.RFC3339
}

// LogBuilder wraps strings.Builder for logging with automatic pooling.
type logBuilder struct {
	*strings.Builder
}

var logBuilderPool = sync.Pool{
	New: func() any {
		return &logBuilder{
			Builder: &strings.Builder{},
		}
	},
}

// getLogBuilder retrieves a builder from the pool.
func getLogBuilder() *logBuilder {
	lb := logBuilderPool.Get().(*logBuilder)
	lb.Reset()
	return lb
}

// Release returns the builder to the pool.
func (lb *logBuilder) Release() {
	if lb == nil || lb.Builder == nil {
		return
	}

	const maxPooledSize = 4096
	if lb.Cap() <= maxPooledSize {
		lb.Reset()
		logBuilderPool.Put(lb)
	}
}

// LoggingConfig holds configuration for the logger.
type LoggingConfig struct {
	LogFilePath string    // Full path to the log file
	MaxSizeMB   int       // Max size of log file in MB before rotation
	MaxBackups  int       // Number of old log files to keep
	Console     io.Writer // Where to write console output (os.Stdout or os.Stderr)
	Program     string    // Tubarr or Metarr
}

// SetupLogging sets up logging for the application.
func SetupLogging(cfg LoggingConfig) (*ProgramLogger, error) {
	if cfg.Program == "" {
		return nil, fmt.Errorf("program name is required")
	}

	if cfg.Console == nil {
		return nil, fmt.Errorf("console writer is required")
	}

	if cfg.MaxSizeMB == 0 {
		cfg.MaxSizeMB = 1
	}
	if cfg.MaxBackups == 0 {
		cfg.MaxBackups = 5
	}

	// Set up zerolog
	fileLogger := zerolog.New(&lumberjack.Logger{
		Filename:   cfg.LogFilePath,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		LocalTime:  true,
	}).
		With().
		Timestamp().
		Logger()

	Loggable = true

	// Program logger model
	pl := &ProgramLogger{
		FileLogger: fileLogger,
		LogBuffer:  make([][]byte, logBufferSize),
		Program:    cfg.Program,
		Console:    cfg.Console,
	}

	pl.D(2, "Loading log file from %q", cfg.LogFilePath)
	pl.loadLogsFromFile(cfg.LogFilePath)

	LogAccessMap.Store(cfg.Program, pl)

	b := getLogBuilder()
	defer b.Release()

	b.WriteString("=========== ")
	b.WriteString(time.Now().Format(time.RFC1123Z))
	b.WriteString(" ===========")
	b.WriteByte('\n')

	startMsg := b.String()
	fileLogger.Log().Msg(sharedregex.AnsiEscapeCompile().ReplaceAllString(startMsg, ""))

	return pl, nil
}

// loadLogsFromFile reads existing log entries from the log file into the buffer.
func (pl *ProgramLogger) loadLogsFromFile(logFilePath string) {
	file, err := os.Open(logFilePath)
	if err != nil {
		pl.W("Could not open file from path %q", logFilePath)
		return
	}
	defer file.Close()

	// Read all lines from file
	var lines [][]byte
	scanner := bufio.NewScanner(file)
	// Increase buffer size for potentially long log lines
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		// Make a copy of the line since scanner reuses buffer
		line := append([]byte(nil), scanner.Bytes()...)
		lines = append(lines, line)
	}

	if scanner.Err() != nil || len(lines) == 0 {
		pl.W("Log file %q, is empty or got error: %v", logFilePath, scanner.Err())
		return
	}

	// Take the last logBufferSize lines (or all if fewer)
	startIdx := 0
	if len(lines) > logBufferSize {
		startIdx = len(lines) - logBufferSize
	}

	pl.LogBufferLock.Lock()
	defer pl.LogBufferLock.Unlock()

	// Load lines into buffer
	for i := startIdx; i < len(lines); i++ {
		pl.LogBuffer[pl.LogBufferPos] = lines[i]
		pl.LogBufferPos++

		if pl.LogBufferPos >= logBufferSize {
			pl.LogBufferPos = 0
			pl.LogBufferFull = true
		}
	}
}

// writeToConsole writes messages to console without using zerolog.
func (pl *ProgramLogger) writeToConsole(msg string) {
	timestamp := time.Now().Format(timeFormat)
	fmt.Fprintf(pl.Console, "%s%s%s %s", sharedconsts.ColorBrightBlack, timestamp, sharedconsts.ColorReset, msg)
}

// buildLogMessage constructs a log message with optional caller info.
func buildLogMessage(prefix, msg string, caller *callerInfo) string {
	b := getLogBuilder()
	defer b.Release()

	if caller != nil {
		estimatedSize := len(prefix) + len(msg) +
			len(tagFunc) + len(caller.funcName) +
			len(tagFile) + len(caller.file) +
			len(tagLine) + len(caller.lineStr) +
			len(tagEnd) + 10

		if b.Cap() < estimatedSize {
			b.Grow(estimatedSize - b.Len())
		}

		b.WriteString(prefix)
		b.WriteString(msg)

		if !strings.HasSuffix(msg, "\n") {
			b.WriteByte(' ')
		}

		b.WriteString(tagFunc)
		b.WriteString(caller.funcName)
		b.WriteString(tagFile)
		b.WriteString(caller.file)
		b.WriteString(tagLine)
		b.WriteString(caller.lineStr)
		b.WriteString(tagEnd)
	} else {
		estimatedSize := len(prefix) + len(msg) + 1

		if b.Cap() < estimatedSize {
			b.Grow(estimatedSize - b.Len())
		}

		b.WriteString(prefix)
		b.WriteString(msg)
		b.WriteByte('\n')
	}

	return b.String()
}

// GetProgramLogger retrieves a program-specific logger from LogAccessMap.
func GetProgramLogger(program string) (*ProgramLogger, bool) {
	val, ok := LogAccessMap.Load(program)
	if !ok {
		return nil, false
	}
	pl, ok := val.(*ProgramLogger)
	return pl, ok
}

// GetRecentLogsForProgram returns logs from RAM for a specific program.
func GetRecentLogsForProgram(program string) [][]byte {
	pl, ok := GetProgramLogger(program)
	if !ok {
		return nil
	}
	return pl.GetRecentLogs()
}

// AddToMemoryLog adds an entry to the program's memory log.
func (pl *ProgramLogger) AddToMemoryLog(p []byte) {
	pl.LogBufferLock.Lock()
	defer pl.LogBufferLock.Unlock()

	// Strip ANSI sequences
	clean := ansiStripper.ReplaceAll(p, nil)

	pl.LogBuffer[pl.LogBufferPos] = append([]byte(nil), clean...)
	pl.LogBufferPos++

	if pl.LogBufferPos >= logBufferSize {
		pl.LogBufferPos = 0
		pl.LogBufferFull = true
	}
}

// GetRecentLogs returns logs from RAM for this program logger.
func (pl *ProgramLogger) GetRecentLogs() [][]byte {
	pl.LogBufferLock.RLock()
	defer pl.LogBufferLock.RUnlock()

	// Buffer not full:
	if !pl.LogBufferFull {
		return append([][]byte(nil), pl.LogBuffer[:pl.LogBufferPos]...)
	}

	// Buffer is full:
	// Build output with correct ordering and count
	out := make([][]byte, 0, logBufferSize)

	// From current write position to end
	out = append(out, pl.LogBuffer[pl.LogBufferPos:]...)

	// From start to current write position
	out = append(out, pl.LogBuffer[:pl.LogBufferPos]...)

	return out
}

// Log logs a message to the program-specific logger.
func (pl *ProgramLogger) Log(level logType, prefix, msg string, withCaller bool, args ...any) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	var caller *callerInfo
	if withCaller {
		c := getCaller(3)
		caller = &c
	}

	logMsg := buildLogMessage(prefix, msg, caller)

	// Write to console
	pl.writeToConsole(logMsg)

	// Add to program-specific memory buffer
	pl.AddToMemoryLog([]byte(logMsg))

	// Log to file
	if !Loggable {
		return
	}
	cleanMsg := sharedregex.AnsiEscapeCompile().ReplaceAllString(msg, "")
	if caller != nil {
		event := pl.getZerologEvent(level).
			Str(jFunction, caller.funcName).
			Str(jFile, caller.file).
			Int(jLine, caller.line)
		event.Msg(cleanMsg)
	} else {
		pl.getZerologEvent(level).Msg(cleanMsg)
	}
}

// getZerologEvent returns the appropriate zerolog event for the level.
func (pl *ProgramLogger) getZerologEvent(level logType) *zerolog.Event {
	switch level {
	case logError:
		return pl.FileLogger.Error()
	case logWarn:
		return pl.FileLogger.Warn()
	case logDebug:
		return pl.FileLogger.Debug()
	case logInfo:
		return pl.FileLogger.Info()
	default:
		return pl.FileLogger.Log()
	}
}

// E logs error messages for this program.
func (pl *ProgramLogger) E(msg string, args ...any) {
	pl.Log(logError, sharedconsts.LogTagError, msg, true, args...)
}

// S logs success messages for this program.
func (pl *ProgramLogger) S(msg string, args ...any) {
	pl.Log(logSuccess, sharedconsts.LogTagSuccess, msg, false, args...)
}

// D logs debug messages for this program.
func (pl *ProgramLogger) D(l int, msg string, args ...any) {
	if l < Level {
		return
	}
	pl.Log(logDebug, sharedconsts.LogTagDebug, msg, true, args...)
}

// W logs warning messages for this program.
func (pl *ProgramLogger) W(msg string, args ...any) {
	pl.Log(logWarn, sharedconsts.LogTagWarning, msg, false, args...)
}

// I logs info messages for this program.
func (pl *ProgramLogger) I(msg string, args ...any) {
	pl.Log(logInfo, sharedconsts.LogTagInfo, msg, false, args...)
}

// P logs plain messages for this program.
func (pl *ProgramLogger) P(msg string, args ...any) {
	pl.Log(logPrint, "", msg, false, args...)
}
