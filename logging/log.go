// Package logging handles the printing and writing of debug and log messages.
package logging

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedregex"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// level is the debug level, -1 when unset.
var level atomic.Int32

func init() {
	level.Store(-1)
}

// Level returns the debug level.
func Level() int {
	return int(level.Load())
}

// SetLevel sets the debug level.
func SetLevel(l int) {
	level.Store(int32(l))
}

// Log array.
var (
	LogAccessMap sync.Map
)

// Log buffer vars.
const logBufferSize = 2500

// ProgramLogger holds logging state for a specific program instance.
type ProgramLogger struct {
	FileLogger zerolog.Logger
	Program    string
	Console    io.Writer

	mu        sync.RWMutex
	logBuffer [][]byte // Ring of the most recent lines.
	total     uint64   // Lines ever written.
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
	logDevError
	logWarn
	logInfo
	logDebug
	logSuccess
	logPrint
)

// Logging package level variables.
var ansiStripper = sharedregex.AnsiEscapeCompile()

// Config holds configuration for the logger.
type Config struct {
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
func SetupLogging(cfg Config) (*ProgramLogger, error) {
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

	// Program logger model
	pl := &ProgramLogger{
		logBuffer: make([][]byte, logBufferSize),
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
	var b strings.Builder

	// Branch with caller info.
	if caller != nil {
		endsNewline := strings.HasSuffix(msg, "\n")

		// Calc size.
		size := len(prefix) + 1 + len(msg) +
			len(tagFunc) + len(caller.funcName) +
			len(tagFile) + len(caller.file) +
			len(tagLine) + len(caller.lineStr) +
			len(tagEnd)
		if !endsNewline {
			size++
		}
		b.Grow(size)

		// Write message.
		b.WriteString(prefix)
		b.WriteByte(' ')
		b.WriteString(msg)

		if !endsNewline {
			b.WriteByte(' ')
		}

		b.WriteString(tagFunc)
		b.WriteString(caller.funcName)
		b.WriteString(tagFile)
		b.WriteString(caller.file)
		b.WriteString(tagLine)
		b.WriteString(caller.lineStr)
		b.WriteString(tagEnd)
	} else { // No caller info.
		// Calc size and write message.
		b.Grow(len(prefix) + 1 + len(msg) + 1)
		b.WriteString(prefix)
		b.WriteByte(' ')
		b.WriteString(msg)
		b.WriteByte('\n')
	}

	return b.String()
}

// GetRecentLogs returns logs from RAM for this program logger, oldest first.
func (pl *ProgramLogger) GetRecentLogs() [][]byte {
	lines, _ := pl.LogsSince(0)
	return lines
}

// LogsSince returns the lines written after the first seen lines, oldest first, and the count to pass next time.
//
// Lines which have already left the buffer are skipped.
func (pl *ProgramLogger) LogsSince(seen uint64) (lines [][]byte, next uint64) {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	size := uint64(len(pl.logBuffer))
	seen = min(max(seen, pl.total-min(pl.total, size)), pl.total)
	lines = make([][]byte, 0, pl.total-seen)
	for i := seen; i < pl.total; i++ {
		lines = append(lines, pl.logBuffer[i%size])
	}
	return lines, pl.total
}

// loadLogsFromFile reads existing log entries from the log file into the buffer.
func (pl *ProgramLogger) loadLogsFromFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		pl.W("Could not open file %q", path)
		return
	}
	defer func() { _ = file.Close() }() // Read-only, so a close error loses nothing.

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
	_, _ = fmt.Fprintf(pl.Console, "%s%s%s %s", sharedconsts.ColorBrightBlack, timestamp, sharedconsts.ColorReset, msg) // Nowhere to report a failed console write.
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
	case logDevError:
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

// DevE logs developer error messages for this program.
func (pl *ProgramLogger) DevE(msg string, args ...any) {
	msg += "\nPlease let us know about this error and how you encountered it at: " + sharedconsts.GithubIssues
	pl.log(logDevError, sharedconsts.LogTagDevError, msg, true, args...)
}

// D logs debug messages for this program.
func (pl *ProgramLogger) D(l int, msg string, args ...any) {
	if Level() < l {
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

// TestLogger checks that an externally sent in ProgramLogger is accepted.
func TestLogger(pl *ProgramLogger) {
	if pl == nil {
		log.Fatalf("%s program logger sent in nil", sharedconsts.LogTagDevError)
	}
	pl.S("%s logger is set up correctly.", pl.Program)
}
