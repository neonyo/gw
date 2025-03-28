package middleware

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/neonyo/gw/pkg/util"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"os"
	"runtime"
	"time"
)

type RecoverMw struct {
	Baser
}

func (mw *RecoverMw) Init() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			defer func() {
				gw := rw.(gatewayResponseWriter)
				ctx := r.Context()
				span := trace.SpanFromContext(ctx)
				if span.IsRecording() {
					var description string
					var code codes.Code
					if err := recover(); err != nil {
						code = codes.Error
						description = fmt.Sprint(description)
						span.RecordError(errors.New(description), trace.WithTimestamp(time.Now()), trace.WithStackTrace(true))
						rw.WriteHeader(http.StatusInternalServerError)
					} else {
						code = codes.Ok
						description = "ok"
						var rspBody string
						if util.IsGzip(rw.Header()) {
							rspBody = util.GzipByteToString(gw.RspBody())
						} else {
							rspBody = string(gw.RspBody())
						}
						span.SetAttributes(attribute.String("http.request.response", rspBody))
					}
					span.SetStatus(code, description)
					return
				}
				if err := recover(); err != nil {
					cover := stack(3)
					fmt.Printf("[Recovery] panic recovered:\n%s\n%s\n", err, cover)
					rw.WriteHeader(http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(rw, r)
		})
	}
}

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

// stack returns a nicely formated stack frame, skipping skip frames
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}
