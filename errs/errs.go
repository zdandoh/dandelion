package errs

import (
	"dandelion/ast"
	"fmt"
	"os"
	"runtime"
)

type ErrorCat string

var sourceProg *ast.Program

const (
	ErrorType   ErrorCat = "Type Error"
	ErrorSyntax ErrorCat = "Syntax Error"
)

var errCount = 0
var sep = lineSep()

func Error(eType ErrorCat, sourceNode ast.Node, format string, a ...interface{}) {
	meta := sourceProg.Meta(sourceNode)
	lineNo := Line(meta)

	msg := Fmt(eType, lineNo, sourceNode, format, a...)
	fmt.Fprintln(os.Stderr, msg)
	errCount++
}

func Fmt(eType ErrorCat, lineNo string, sourceNode ast.Node, format string, a ...interface{}) string {
	preMsg := fmt.Sprintf(format, a...)

	msg := fmt.Sprintf("%s: line %s: %s%s%s", eType, lineNo, preMsg, sep, sourceNode)
	return msg
}

func CheckExit() {
	if errCount > 0 {
		os.Exit(1)
	}
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
