package context

import (
	"github.com/yosssi/gb/options"
	"github.com/yosssi/gb/result"
	"log"
	"sync"
)

// A Context represents a context of the main process.
type Context struct {
	Options options.Options
	Url string
	Debug bool
	Results []result.Result
	ResultsMutex sync.Mutex
}

// Dprintf executes log.Printf if receiver's debug property is true.
func (ctx *Context) Dprintf(format string, v ...interface{}) {
	if ctx.Debug {
		log.Printf(format, v...)
	}
}

// AppendResult appends a result to the context's results.
func (ctx *Context) AppendResult(r result.Result) {
	ctx.LockResults()
	ctx.Results = append(ctx.Results, r)
	ctx.UnlockResults()
}

// LockResults locks the context's results.
func (ctx *Context) LockResults() {
	ctx.ResultsMutex.Lock()
}

// UnlockResults unlocks the context's results.
func (ctx *Context) UnlockResults() {
	ctx.ResultsMutex.Unlock()
}
