package log

import (
	"fmt"
)

type LogLevel int

const (
	DebugLevel LogLevel = iota + 1
	PrintLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)
const (
	colorRed = uint8(iota + 91)
	colorGreen
	colorYellow
	colorBlue
	//colorPurple
)

var (
	red    = fmt.Sprintf("\x1b[%dm", colorRed)
	green  = fmt.Sprintf("\x1b[%dm", colorGreen)
	yellow = fmt.Sprintf("\x1b[%dm", colorYellow)
	blue   = fmt.Sprintf("\x1b[%dm", colorBlue)
	cyan   = fmt.Sprintf("\x1b[%dm", 36)
	//purple   = fmt.Sprintf("\x1b[%dm", colorPurple)
	//colorEnd = "\x1b[0m"
)

func ColorLevel(lvl LogLevel) string {
	switch lvl {
	case DebugLevel:
		return fmt.Sprintf("%s[DEG]\x1b[0m ", blue)
	case PrintLevel:
		return fmt.Sprintf("%s[PRT]\x1b[0m ", cyan)
	case InfoLevel:
		return fmt.Sprintf("%s[INF]\x1b[0m ", green)
	case WarnLevel:
		return fmt.Sprintf("%s[WAR]\x1b[0m ", yellow)
	case ErrorLevel:
		return fmt.Sprintf("%s[ERR]\x1b[0m ", red)
	case FatalLevel:
		return fmt.Sprintf("%s[FATAL]\x1b[0m ", red)
	case PanicLevel:
		return fmt.Sprintf("%s[PANIC]\x1b[0m ", red)
	default:
		return fmt.Sprintf("%s[L%d]\x1b[0m ", red, lvl)
	}
}
