package compile

import (
	"bytes"
	"dandelion/errs"
	"dandelion/parser"
	"dandelion/transform"
	"dandelion/typecheck"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func RunProg(progText string) (string, int) {
	prog := parser.ParseProgram(progText)
	errs.SetProg(prog)
	fmt.Println(prog)
	transform.TransformAst(prog)

	progTypes := typecheck.Infer(prog)
	typecheck.ValidateProg(prog, progTypes)
	errs.CheckExit()

	llvm_ir := Compile(prog, progTypes)

	fmt.Println(llvm_ir)
	err := ioutil.WriteFile("llvm_ir.ll", []byte(llvm_ir), os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	output, err := exec.Command("bash", "-i", "tester.sh").Output()
	if err != nil {
		log.Println(string(output))
		log.Fatalf(err.Error())
	}

	outputStr := strings.TrimSpace(string(output))
	outLines := strings.Split(outputStr, "\n")
	lastLine := outLines[len(outLines)-1]
	outLines = outLines[0 : len(outLines)-1]

	exitCode, err := strconv.Atoi(lastLine)
	if err != nil {
		log.Fatalln(outputStr, err)
	}

	return strings.Join(outLines, "\n"), exitCode
}

func CompileCheckExit(progText string, code int) bool {
	outputStr, exitCode := RunProg(progText)

	if exitCode != code {
		log.Println(outputStr)
		return false
	}

	return true
}

func CompileCheckOutput(progText string, output string) bool {
	outputStr, _ := RunProg(progText)

	if outputStr != strings.TrimSpace(output) {
		fmt.Println("Output doesn't match:")
		fmt.Println(outputStr)
		return false
	}

	return true
}

func ExecIR(llvmIr string) error {
	cmd := exec.Command("lli", "-load", "lib/lib.so")
	buffer := bytes.NewBufferString(llvmIr)

	cmd.Stdin = buffer
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Start()
	if err != nil {
		log.Fatalf(err.Error())
	}

	exitStatus := 0
	err = cmd.Wait()
	if err != nil {
		exitCode, ok := err.(*exec.ExitError)
		if ok {
			status, ok := exitCode.Sys().(syscall.WaitStatus)
			if ok {
				exitStatus = status.ExitStatus()
			}
		}
	}

	os.Exit(exitStatus)
	return nil
}

func CompileSource(progText string, optLevel int) string {
	prog := parser.ParseProgram(progText)
	transform.TransformAst(prog)

	progTypes := typecheck.Infer(prog)
	llvmIr := Compile(prog, progTypes)

	optLevelArg := fmt.Sprintf("-O%d", optLevel)
	cmd := exec.Command("opt", optLevelArg, "-enable-coroutines", "-coro-early", "-coro-split", "-coro-elide", "-coro-cleanup", "-S")

	inputIR := bytes.NewBufferString(llvmIr)
	cmd.Stdin = inputIR

	optIR, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}

	return string(optIR)
}
