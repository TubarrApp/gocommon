// Package benchmark sets up and initiates benchmarking.
//
// Includes CPU profiling, memory profiling, and tracing.
package benchmark

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"time"

	"github.com/TubarrApp/gocommon/consts"
	"github.com/TubarrApp/gocommon/logging"
)

// BenchFiles contain benchmarking files written on a benchmark-enabled run.
type BenchFiles struct {
	cpuFile   *os.File
	memFile   *os.File
	traceFile *os.File
}

var (
	cpuProfPath,
	memProfPath,
	traceOutPath string
)

// SetupBenchmarking sets up and initiates benchmarking for a program run.
func SetupBenchmarking(log *logging.ProgramLogger, benchmarkDir string) (*BenchFiles, error) {
	var err error
	b := new(BenchFiles)

	startTime := time.Now().Format("2006-01-02_15-04-05")
	makeBenchFilepaths(log, benchmarkDir, startTime)

	log.I("(Benchmarking this run. Start time: %s)", startTime)

	// CPU profile
	b.cpuFile, err = os.Create(cpuProfPath)
	if err != nil {
		CloseBenchFiles(log, b, "", fmt.Errorf("could not create CPU profiling file: %w", err))
		return nil, err
	}

	if err := pprof.StartCPUProfile(b.cpuFile); err != nil {
		CloseBenchFiles(log, b, "", fmt.Errorf("could not start CPU profiling: %w", err))
		return nil, err
	}

	// Memory profile
	b.memFile, err = os.Create(memProfPath)
	if err != nil {
		CloseBenchFiles(log, b, "", fmt.Errorf("could not create memory profiling file: %w", err))
		return nil, err
	}

	// Trace
	b.traceFile, err = os.Create(traceOutPath)
	if err != nil {
		CloseBenchFiles(log, b, "", fmt.Errorf("could not create trace file: %w", err))
		return nil, err
	}
	if err := trace.Start(b.traceFile); err != nil {
		CloseBenchFiles(log, b, "", fmt.Errorf("could not start trace: %w", err))
		return nil, err
	}

	return b, nil
}

// CloseBenchFiles closes bench files on program termination.
func CloseBenchFiles(log *logging.ProgramLogger, b *BenchFiles, noErrExit string, setupErr error) {
	if b == nil {
		return
	}

	if b.cpuFile != nil {
		log.I("Stopping CPU profile...")
		pprof.StopCPUProfile()
		if err := b.cpuFile.Close(); err != nil {
			log.E("Failed to close file %q: %v", b.cpuFile.Name(), err)
		}
		b.cpuFile = nil // Prevent double-close
	}

	if b.traceFile != nil {
		log.I("Stopping trace...")
		trace.Stop()
		if err := b.traceFile.Close(); err != nil {
			log.E("Failed to close file %q: %v", b.traceFile.Name(), err)
		}
		b.traceFile = nil // Prevent double-close
	}

	if b.memFile != nil {
		log.I("Writing memory profile...")
		runtime.GC()
		if err := pprof.WriteHeapProfile(b.memFile); err != nil {
			log.E("Could not write memory profile: %v", err)
		}
		if err := b.memFile.Close(); err != nil {
			log.E("Failed to close file %q: %v", b.memFile.Name(), err)
		}
		b.memFile = nil // Prevent double-close
	}

	if setupErr != nil {
		log.E("Benchmarking failure: %v", setupErr)
	}
	if noErrExit != "" {
		log.I("%s", noErrExit)
	}
}

// makeBenchFilepaths makes paths for benchmarking files in a timestamped subdirectory.
func makeBenchFilepaths(log *logging.ProgramLogger, baseDir, timestamp string) {
	// Create timestamped subdirectory for this run
	runDir := filepath.Join(baseDir, timestamp)

	if err := os.MkdirAll(runDir, consts.PermsGenericDir); err != nil {
		log.E("Failed to create benchmark run directory: %v", err)
		return
	}

	// Simple filenames (no timestamp needed since they're in timestamped folder)
	cpuProfPath = filepath.Join(runDir, "cpu.prof")
	memProfPath = filepath.Join(runDir, "mem.prof")
	traceOutPath = filepath.Join(runDir, "trace.out")

	log.I("Created benchmark directory: %q", runDir)
}
