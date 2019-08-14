package mint

// inspired from gin logger (both are same , but added some code)
import (
	"fmt"
	"net/http"
	"time"
)

//ANSI color for appropriately logging
var (
	green        = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white        = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow       = string([]byte{27, 91, 57, 48, 59, 52, 51, 109})
	red          = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue         = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta      = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan         = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	reset        = string([]byte{27, 91, 48, 109})
	disableColor = false
	forceColor   = false
)

//Logger logger structure
type Logger struct {
	TimeStamp  time.Time
	StatusCode int
	Latency    time.Duration
	Method     string
	Path       string
	BodySize   int
	ClientIP   string
	UserName   string
	Errors     []error
}

//NewLogger creates a new logger
func NewLogger() *Logger {
	return new(Logger)
}

func (l *Logger) getStatusCodeColor() string {
	code := l.StatusCode
	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return white
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return yellow
	default:
		return red
	}
}

func (l *Logger) getMethodColor() string {
	method := l.Method
	switch method {
	case "GET":
		return blue
	case "POST":
		return cyan
	case "PUT":
		return yellow
	case "DELETE":
		return red
	case "PATCH":
		return green
	case "HEAD":
		return magenta
	case "OPTIONS":
		return white
	default:
		return reset
	}
}

func (l *Logger) getResetColor() string {
	return reset
}

//Print prints log
func (l *Logger) Print() {
	statusColor := l.getStatusCodeColor()
	methodColor := l.getMethodColor()
	resetColor := l.getResetColor()
	fmt.Println(fmt.Sprintf("[Service] %v |%s %3d %s| %13v | %15s |%s %-7s %s| %s > %v Bytes",
		l.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, l.StatusCode, resetColor,
		l.Latency,
		l.ClientIP,
		methodColor, l.Method, resetColor,
		l.Path,
		l.BodySize,
	))
	for _, err := range l.Errors {
		switch err.(type) {
		case Error:
			err1 := err.(Error)
			fmt.Println(fmt.Errorf("%s %s %v %s", err1.file, err1.funcName, err1.line, err1.error.Error()))
		case error:
			fmt.Println(fmt.Errorf("%s", err.Error()))
		}
	}
}
