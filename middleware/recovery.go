package middleware

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strings"

	"github.com/appist/appy/support"
	"github.com/gin-gonic/gin"
)

var (
	recoveryDunno     = []byte("???")
	recoveryCenterDot = []byte("·")
	recoveryDot       = []byte(".")
	recoverySlash     = []byte("/")
)

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				recoveryErrorHandler(c, err)
			}
		}()

		c.Next()
	}
}

func recoveryErrorHandler(c *gin.Context, err interface{}) {
	// Check for a broken connection, as it is not really a condition that warrants a panic
	// stack trace.
	var brokenPipe bool
	if ne, ok := err.(*net.OpError); ok {
		if se, ok := ne.Err.(*os.SyscallError); ok {
			if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
				brokenPipe = true
			}
		}
	}

	stack := recoveryStack(3)
	httpRequest, _ := httputil.DumpRequest(c.Request, false)
	if brokenPipe {
		c.Error(fmt.Errorf("panic recovered:\n%s\n%s", err, string(httpRequest)))
	} else {
		c.Error(fmt.Errorf("panic recovered:\n%s\n%s", err, stack))
	}

	renderErrors(c)
}

func renderErrors(c *gin.Context) {
	session := DefaultSession(c)
	sessionVars := ""
	if session != nil {
		for key, val := range session.Values() {
			sessionVars = sessionVars + fmt.Sprintf("%s: %+v<br>", key, val)
		}
	}

	if sessionVars == "" {
		sessionVars = "None"
	}

	tplErrors := []template.HTML{}
	for _, err := range c.Errors {
		support.Logger.Error(err.Error())
		tplErrors = append(tplErrors, template.HTML(err.Error()))
	}

	headers := ""
	for key, val := range c.Request.Header {
		headers = headers + fmt.Sprintf("%s: %s<br>", key, strings.Join(val, ", "))
	}

	qsParams := ""
	for key, val := range c.Request.URL.Query() {
		qsParams = qsParams + fmt.Sprintf("%s: %s<br>", key, strings.Join(val, ", "))
	}

	if qsParams == "" {
		qsParams = "None"
	}

	defer func(c *gin.Context) {
		if r := recover(); r != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}(c)

	c.HTML(http.StatusInternalServerError, "error/500", gin.H{
		"errors":      tplErrors,
		"headers":     template.HTML(headers),
		"qsParams":    template.HTML(qsParams),
		"sessionVars": template.HTML(sessionVars),
		"title":       "500 Internal Server Error",
	})
}

// recoveryStack returns a nicely formatted stack frame, skipping skip frames.
func recoveryStack(skip int) []byte {
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
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", recoveryFunction(pc), recoverySource(lines, line))
	}
	return buf.Bytes()
}

// recoverySource returns a space-trimmed slice of the n'th line.
func recoverySource(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return recoveryDunno
	}
	return bytes.TrimSpace(lines[n])
}

// recoveryFunction returns, if possible, the name of the function containing the PC.
func recoveryFunction(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return recoveryDunno
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
	if lastSlash := bytes.LastIndex(name, recoverySlash); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if period := bytes.Index(name, recoveryDot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, recoveryCenterDot, recoveryDot, -1)
	return name
}
