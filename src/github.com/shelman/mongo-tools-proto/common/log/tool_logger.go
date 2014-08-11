package log

import (
	"fmt"
	"github.com/shelman/mongo-tools-proto/common/options"
	"io"
	"os"
	"sync"
	"time"
)

const (
	MongoDumpLegacyDate = "Mon Jan _2 15:04:05.000"
)

type ToolLogger struct {
	m      *sync.Mutex
	w      io.Writer
	format string
	v      int
}

func (tl *ToolLogger) SetVerbosity(verbosity *options.Verbosity) {
	tl.m.Lock()
	if verbosity.Quiet {
		tl.v = -1
	} else {
		tl.v = len(verbosity.Verbose)
	}
	tl.m.Unlock()
}

func (tl *ToolLogger) SetWriter(writer io.Writer) {
	tl.m.Lock()
	tl.w = writer
	tl.m.Unlock()
}

func (tl *ToolLogger) SetDateFormat(dateFormat string) {
	tl.m.Lock()
	tl.format = dateFormat
	tl.m.Unlock()
}

func (tl *ToolLogger) Logf(minVerb int, format string, a ...interface{}) {
	if minVerb < 0 {
		panic("cannot set a minimum log verbosity that is less than 0")
	}

	if minVerb <= tl.v {
		// technically there is a race condition here wherein Logf starts,
		// and another routine changes v to a lower value than minVerb during
		// this if block. This is incredibly unlikely to be an issue, and
		// well worth the benefit of not having lock contention when
		// handling non-logged messages
		tl.m.Lock()
		tl.log(fmt.Sprintf(format, a...))
		tl.m.Unlock()
	}
}

func (tl *ToolLogger) log(msg string) {
	fmt.Fprintf(tl.w, "%v\t%v\n", time.Now().Format(tl.format), msg)
}

func NewToolLogger(verbosity *options.Verbosity) *ToolLogger {
	tl := &ToolLogger{
		m:      &sync.Mutex{},
		w:      os.Stderr,           // default to stderr
		format: MongoDumpLegacyDate, //TODO whats up with this?
	}
	tl.SetVerbosity(verbosity)
	return tl
}
