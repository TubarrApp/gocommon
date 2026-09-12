package sharedmodels

import (
	"io"
	"sync"

	"github.com/rs/zerolog"
)

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
