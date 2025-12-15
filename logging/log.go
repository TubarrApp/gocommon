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
const logBufferSize = 2500

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
	LogFilePath string    // Full path to the log file.
	MaxSizeMB   int       // Max size of log file in MB before rotation.
	MaxBackups  int       // Number of old log files to keep.
	Console     io.Writer // Where to write console output (os.Stdout or os.Stderr).
	Program     string    // Tubarr or Metarr.
}

// init runs before other functions.
func init() {
	zerolog.TimeFieldFormat = time.RFC3339
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
	fileWriter := &lumberjack.Logger{
		Filename:   cfg.LogFilePath,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		LocalTime:  true,
	}

	Loggable = true

	// Program logger model
	pl := &ProgramLogger{
		LogBuffer: make([][]byte, logBufferSize),
		Program:   cfg.Program,
		Console:   cfg.Console,
	}

	// Write to file + RAM
	mw := &memoryWriter{
		pl:     pl,
		writer: fileWriter,
	}

	pl.FileLogger = zerolog.New(mw).With().Timestamp().Logger()

	// Only load in file from Tubarr on start, Metarr load in should
	// be handled by the caller (usually in server handlers).
	if cfg.Program == "Tubarr" {
		pl.D(2, "Loading log file from %q", cfg.LogFilePath)
		pl.loadLogsFromFile(cfg.LogFilePath)
	}

	LogAccessMap.Store(cfg.Program, pl)

	startMsg := fmt.Sprintf("=========== %s ===========", time.Now().Format(time.RFC1123Z))
	pl.FileLogger.Log().Msg(startMsg)

	return pl, nil
}

// GetRecentLogsForProgram returns logs from RAM for a specific program.
// Usually used by server handlers to fill display views.
func GetRecentLogsForProgram(program string) [][]byte {
	// Load logs for program.
	val, ok := LogAccessMap.Load(program)
	if !ok {
		return nil
	}

	// Ensure type correctness.
	pl, ok := val.(*ProgramLogger)
	if !ok {
		return nil
	}

	// Return recent logs.
	return pl.GetRecentLogs()
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

// GetRecentLogs returns logs from RAM for this program logger.
func (pl *ProgramLogger) GetRecentLogs() [][]byte {
	pl.LogBufferLock.RLock()
	defer pl.LogBufferLock.RUnlock()

	// BUFFER FULL:
	if !pl.LogBufferFull {
		return append([][]byte(nil), pl.LogBuffer[:pl.LogBufferPos]...)
	}

	// BUFFER NOT FULL:
	out := make([][]byte, 0, logBufferSize)

	// From current write position to end.
	out = append(out, pl.LogBuffer[pl.LogBufferPos:]...)

	// From start to current write position.
	out = append(out, pl.LogBuffer[:pl.LogBufferPos]...)

	return out
}

// GetLogsSincePosition returns only the logs added since a specific buffer position.
func (pl *ProgramLogger) GetLogsSincePosition(lastPos int, wasWrapped bool) [][]byte {
	pl.LogBufferLock.RLock()
	defer pl.LogBufferLock.RUnlock()

	currentPos := pl.LogBufferPos
	currentWrapped := pl.LogBufferFull

	// Case #1: No new logs (position unchanged).
	if currentPos == lastPos && currentWrapped == wasWrapped {
		return nil
	}

	// Case #2: Buffer was not wrapped before and still isn't.
	if !wasWrapped && !currentWrapped {
		if currentPos > lastPos {
			return append([][]byte(nil), pl.LogBuffer[lastPos:currentPos]...)
		}
		return nil
	}

	// Case #3: Buffer was not wrapped before but is now.
	if !wasWrapped && currentWrapped {
		// Return from 'lastPos' to end of buffer, then 0 to 'currentPos'.
		out := make([][]byte, 0, logBufferSize)
		out = append(out, pl.LogBuffer[lastPos:]...)
		out = append(out, pl.LogBuffer[:currentPos]...)
		return out
	}

	// Case #4: Buffer was wrapped before and still is.
	if wasWrapped && currentWrapped {
		if currentPos > lastPos {
			// No wrap-around occurred between checks
			return append([][]byte(nil), pl.LogBuffer[lastPos:currentPos]...)
		}
		// Wrap-around occurred
		out := make([][]byte, 0, logBufferSize-lastPos+currentPos)
		out = append(out, pl.LogBuffer[lastPos:]...)
		out = append(out, pl.LogBuffer[:currentPos]...)
		return out
	}

	// Case #5: Buffer was wrapped but now isn't (shouldn't happen)
	fmt.Fprintf(os.Stderr, "Dev Error: Buffer unwrapped.")
	return pl.GetRecentLogs()
}

// GetBufferPosition returns the current write position in the log buffer.
func (pl *ProgramLogger) GetBufferPosition() int {
	pl.LogBufferLock.RLock()
	defer pl.LogBufferLock.RUnlock()
	return pl.LogBufferPos
}

// IsBufferFull returns whether the log buffer is full.
func (pl *ProgramLogger) IsBufferFull() bool {
	pl.LogBufferLock.RLock()
	defer pl.LogBufferLock.RUnlock()
	return pl.LogBufferFull
}

// loadLogsFromFile reads existing log entries from the log file into the buffer.
func (pl *ProgramLogger) loadLogsFromFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		pl.W("Could not open file %q", path)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		raw := scanner.Bytes()
		line := make([]byte, len(raw)+1)
		copy(line, raw)
		line[len(raw)] = '\n'

		pl.addToRAMLine(line)
	}

	if err := scanner.Err(); err != nil {
		pl.W("Error scanning log file %q: %v", path, err)
	}
}

// writeToConsole writes messages to console without using zerolog.
func (pl *ProgramLogger) writeToConsole(msg string) {
	timestamp := time.Now().Format(timeFormat)
	fmt.Fprintf(pl.Console, "%s%s%s %s", sharedconsts.ColorBrightBlack, timestamp, sharedconsts.ColorReset, msg)
}

// Log logs a message to the program-specific logger.
func (pl *ProgramLogger) log(level logType, prefix, msg string, withCaller bool, args ...any) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	var caller *callerInfo
	if withCaller {
		c := getCaller(3) // skip lines: getCaller -> log -> D/E/W/I/P (etc.) -> [ DESIRED FUNCTION ]
		caller = &c
	}

	// Build human-readable console message.
	logMsg := buildLogMessage(prefix, msg, caller)

	// Write to console.
	pl.writeToConsole(logMsg)

	// Call zerolog event.
	clean := ansiStripper.ReplaceAllString(msg, "")
	if caller != nil {
		pl.getZerologEvent(level).
			Str(jFunction, caller.funcName).
			Str(jFile, caller.file).
			Int(jLine, caller.line).
			Msg(clean)
	} else {
		pl.getZerologEvent(level).Msg(clean)
	}
}

// getZerologEvent returns the appropriate zerolog event for the level.
func (pl *ProgramLogger) getZerologEvent(level logType) *zerolog.Event {
	switch level {
	case logDebug:
		return pl.FileLogger.Debug()
	case logError:
		return pl.FileLogger.Error()
	case logInfo:
		return pl.FileLogger.Info()
	case logSuccess:
		return pl.FileLogger.Log().Str("level", "success") // Zerolog doesn't have a built-in success level, make custom.
	case logWarn:
		return pl.FileLogger.Warn()
	default:
		return pl.FileLogger.Log()
	}
}

// ---- OUTER PROGRAM LOG CALLS ----

// D logs debug messages for this program.
func (pl *ProgramLogger) D(l int, msg string, args ...any) {
	if Level < l {
		return
	}
	pl.log(logDebug, sharedconsts.LogTagDebug, msg, true, args...)
}

// E logs error messages for this program.
func (pl *ProgramLogger) E(msg string, args ...any) {
	pl.log(logError, sharedconsts.LogTagError, msg, true, args...)
}

// I logs info messages for this program.
func (pl *ProgramLogger) I(msg string, args ...any) {
	pl.log(logInfo, sharedconsts.LogTagInfo, msg, false, args...)
}

// P logs plain messages for this program.
func (pl *ProgramLogger) P(msg string, args ...any) {
	pl.log(logPrint, "", msg, false, args...)
}

// S logs success messages for this program.
func (pl *ProgramLogger) S(msg string, args ...any) {
	pl.log(logSuccess, sharedconsts.LogTagSuccess, msg, false, args...)
}

// W logs warning messages for this program.
func (pl *ProgramLogger) W(msg string, args ...any) {
	pl.log(logWarn, sharedconsts.LogTagWarning, msg, false, args...)
}
