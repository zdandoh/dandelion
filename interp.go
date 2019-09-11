package main

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type Value interface {
}

type Null struct {
}

type Array struct {
	length int
	cap    int
	arr    []interface{}
}

func (a *Array) String() string {
	return fmt.Sprintf("[%v]", strings.Join(Strings(a.arr), ", "))
}

type Env map[string]Value
type PrimFunc = func([]Value) Value

func (n *Null) String() string {
	return "null"
}

type Program struct {
	mainFunc *FunDef
	Environ  Env
	output   string
}

func (p *Program) Output() string {
	return p.output
}

func NewProgram(mainFunc *FunDef) *Program {
	newProg := &Program{}
	newProg.Environ = NewEnv(newProg)
	newProg.mainFunc = mainFunc

	return newProg
}

func NewEnv(prog *Program) Env {
	newEnv := make(Env)
	newEnv["p"] = func(values []Value) Value {
		for _, value := range values {
			fmt.Print(value)
			prog.output += fmt.Sprintf("%v", value)
		}
		prog.output += "\n"
		fmt.Println()
		return &Null{}
	}

	return newEnv
}

func (p *Program) interpExp(astNode AstNode) Value {
	var retVal Value

	switch node := astNode.(type) {
	case *Assign:
		p.Environ[node.ident] = p.interpExp(node.expr)
		retVal = &Null{}
	case *Num:
		res, _ := strconv.Atoi(node.value)
		retVal = res
	case *Ident:
		var ok bool
		retVal, ok = p.Environ[node.value]
		if !ok {
			panic("Unbound variable: " + node.value)
		}
	case *AddSub:
		if node.op == "+" {
			retVal = p.interpExp(node.left).(int) + p.interpExp(node.right).(int)
		} else if node.op == "-" {
			retVal = p.interpExp(node.left).(int) - p.interpExp(node.right).(int)
		} else {
			panic("Unknown AddSub op")
		}
	case *MulDiv:
		retVal = p.interpExp(node.left).(int) * p.interpExp(node.right).(int)
	case *FunDef:
		retVal = node
	case *FunApp:
		retVal = p.interpFunApp(node)
	case *While:
		for p.interpExp(node.cond).(int) != 0 {
			for _, line := range node.body.lines {
				p.interpExp(line)
			}
		}

		retVal = &Null{}
	case *If:
		if p.interpExp(node.cond).(int) != 0 {
			for _, line := range node.body.lines {
				p.interpExp(line)
			}
		}

		retVal = &Null{}
	case *CompNode:
		retVal = p.interpComp(node)
	case *ArrayLiteral:
		newArr := &Array{}
		newArr.length = node.length
		newArr.cap = node.length

		for i := 0; i < node.length; i++ {
			newArr.arr = append(newArr.arr, p.interpExp(node.exprs[i]))
		}

		retVal = newArr
	case *SliceNode:
		arr := p.interpExp(node.arr).(*Array)
		index := p.interpExp(node.index).(int)
		retVal = arr.arr[index]
	case *StrExp:
		retVal = node.value
	default:
		panic("Interp not defined for type: " + reflect.TypeOf(astNode).String())
	}

	return retVal
}

func (p *Program) interpComp(comp *CompNode) Value {
	left := p.interpExp(comp.left).(int)
	right := p.interpExp(comp.right).(int)

	var retVal bool

	switch comp.op {
	case ">":
		retVal = left > right
	case "<":
		retVal = left < right
	case ">=":
		retVal = left >= right
	case "<=":
		retVal = left <= right
	case "==":
		retVal = left <= right
	default:
		panic("Invalid comp op")
	}

	if retVal {
		return 1
	}
	return 0
}

func (p *Program) interpFunApp(funApp *FunApp) Value {
	funExp := p.interpExp(funApp.fun)

	primFunc, isPrimFunc := funExp.(PrimFunc)
	if isPrimFunc {
		args := make([]Value, 0)
		for _, arg := range funApp.args {
			args = append(args, p.interpExp(arg))
		}

		return primFunc(args)
	}

	funVal := funExp.(*FunDef)
	for i := 0; i < len(funVal.args); i++ {
		argName := funVal.args[i]
		argValue := p.interpExp(funApp.args[i])
		p.Environ[argName] = argValue
	}

	var lastVal Value
	for _, line := range funVal.body.lines {
		lastVal = p.interpExp(line)
	}

	return lastVal
}

func (p *Program) interp() {
	mainApp := &FunApp{}
	mainApp.fun = p.mainFunc
	mainApp.args = make([]AstNode, 0)
	p.interpFunApp(mainApp)
}

func CompareOutput(progText string, output string) bool {
	prog := ParseProgram(progText)
	_, err := TypeCheck(prog)
	if err != nil {
		log.Fatal("Program doesn't type check: " + err.Error())
		return false
	}

	prog.interp()

	reference := strings.Trim(output, "\r\n")
	produced := strings.Trim(prog.Output(), "\r\n")
	if reference != produced {
		DiffOutput(reference, produced)
		return false
	}

	return true
}

func DiffOutput(reference string, produced string) {
	if reference == produced {
		fmt.Println("Strings do not differ")
		return
	}

	line := 1
	char := 0
	for i := 0; i < len(reference); i++ {
		if char == '\n' {
			line++
			char = 0
		}
		if reference[i] != produced[i] {
			break
		}
		char++
	}

	fmt.Printf("First differing character at line %d char %d\n", line, char)
	fmt.Printf("-- Reference --\n%s\n-- Produced --\n%s\n", reference, produced)
}

func Strings(items []interface{}) []string {
	strings := make([]string, 0)

	for _, item := range items {
		strings = append(strings, fmt.Sprintf("%v", item))
	}

	return strings
}
