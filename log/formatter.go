package log

import (
	"fmt"
)


const (
	colorRed = uint8(iota + 91)
	colorGreen
	colorYellow
	colorBlue
	//colorPurple
)

var (
	// red      = fmt.Sprintf("\x1b[%dm", colorRed)
	// green    = fmt.Sprintf("\x1b[%dm", colorGreen)
	// yellow   = fmt.Sprintf("\x1b[%dm", colorYellow)
	// blue     = fmt.Sprintf("\x1b[%dm", colorBlue)
	// cyan     = fmt.Sprintf("\x1b[%dm", 36)
	// purple   = fmt.Sprintf("\x1b[%dm", colorPurple)
	// colorEnd = "\x1b[0m"

	debugColor = fmt.Sprintf("\x1b[%dm[DEG]\x1b[0m ", colorBlue)
	printColor = fmt.Sprintf("\x1b[%dm[PRT]\x1b[0m ", 36)
	infoColor = fmt.Sprintf("\x1b[%dm[INF]\x1b[0m ", colorGreen)
	warnColor = fmt.Sprintf("\x1b[%dm[WAR]\x1b[0m ", colorYellow)
	errorColor = fmt.Sprintf("\x1b[%dm[ERR]\x1b[0m ", colorRed)
	fatalColor = fmt.Sprintf("\x1b[%dm[FATAL]\x1b[0m ", colorRed)
	panicColor = fmt.Sprintf("\x1b[%dm[PANIC]\x1b[0m ", colorRed)

)

func ColorLevel(lvl LogLevel) string {
	switch lvl {
	case DebugLevel:
		return debugColor
	case PrintLevel:
		return printColor
	case InfoLevel:
		return infoColor
	case WarnLevel:
		return warnColor
	case ErrorLevel:
		return errorColor
	case FatalLevel:
		return fatalColor
	case PanicLevel:
		return panicColor
	default:
		return fmt.Sprintf("\x1b[%dm[L%d]\x1b[0m ", colorRed, lvl)
	}
}
