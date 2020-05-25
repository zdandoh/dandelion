package errs

import (
	"dandelion/ast"
	"errors"
	"fmt"
	"os"
	"runtime"
)

var ExitFun = Exit
var sourceProg *ast.Program

var (
	ErrorType   = errors.New("Type Error")
	ErrorValue  = errors.New("Value Error")
	ErrorSyntax = errors.New("Syntax Error")
)

var errCount = 0
var sep = lineSep()

func Error(eType error, sourceNode ast.Node, format string, a ...interface{}) {
	var lineNo string
	if sourceNode != nil {
		meta := sourceProg.Meta(sourceNode)
		lineNo = Line(meta)
	} else {
		lineNo = "<unknown>"
	}

	msg := Fmt(eType, lineNo, sourceNode, format, a...)
	fmt.Fprintln(os.Stderr, msg)
	errCount++
}

func Fmt(eType error, lineNo string, sourceNode ast.Node, format string, a ...interface{}) string {
	preMsg := fmt.Sprintf(format, a...)

	msg := fmt.Sprintf("%s: line %s: %s%s", eType.Error(), lineNo, preMsg, sep)
	if sourceNode != nil {
		msg += sourceNode.String()
	}
	return msg
}

func CheckExit() {
	if errCount > 0 {
		ExitFun()
	}
}

func Exit() {
	os.Exit(1)
}

func SetProg(prog *ast.Program) {
	sourceProg = prog
}

func Line(meta *ast.Meta) string {
	var lineNo string
	if meta == nil {
		lineNo = "<unknown>"
	} else {
		lineNo = fmt.Sprintf("%d", meta.LineNo)
	}

	return lineNo
}

func lineSep() string {
	switch runtime.GOOS {
	case "windows":
		return "\r\n"
	default:
		return "\n"
	}
}
