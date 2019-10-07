package interp

import (
	"ahead/ast"
	"ahead/parser"
	"ahead/transform"
	"bufio"
	"errors"
	"fmt"
	"os/exec"
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

type Iterator struct {
	iter chan Value
}

func (iter *Iterator) String() string {
	return "<iterator object>"
}

var StopIteration = errors.New("Stop iteration")

type Env map[string]Value

type PrimFunc = func([]Value) Value

type Interpreter struct {
	Environ     Env
	CurrProgram *ast.Program
	output      string
}

func NewInterpreter() *Interpreter {
	i := &Interpreter{}
	i.Environ = NewEnv(i)

	return i
}

func (i *Interpreter) Output() string {
	return i.output
}

func NewEnv(i *Interpreter) Env {
	newEnv := make(Env)
	newEnv["p"] = func(values []Value) Value {
		for _, value := range values {
			fmt.Print(value)
			i.output += fmt.Sprintf("%v", value)
		}
		i.output += "\n"
		fmt.Println()
		return &Null{}
	}

	return newEnv
}

func (i *Interpreter) interpExp(astNode ast.Node) Value {
	var retVal Value

	switch node := astNode.(type) {
	case *ast.Assign:
		i.Environ[node.Ident] = i.interpExp(node.Expr)
		retVal = &Null{}
	case *ast.Num:
		res, _ := strconv.Atoi(node.Value)
		retVal = res
	case *ast.Ident:
		var ok bool
		retVal, ok = i.Environ[node.Value]
		if !ok {
			funcVal, ok := i.CurrProgram.Funcs[node.Value]
			if !ok {
				panic("Unbound variable: " + node.Value)
			}
			retVal = funcVal
		}
	case *ast.PipeExp:
		iter := i.interpExp(node.Left)
		arr, isArr := iter.(*Array)
		iterator, isIter := iter.(*Iterator)

		appFun := i.interpExp(node.Right)
		resultArr := make([]interface{}, 0)
		index := 0
		for {
			var result Value
			if isArr {
				if index >= arr.length {
					break
				}
				result = i.interpFunDef(appFun, []Value{index, arr.arr[index], arr})
			} else if isIter {
				iterVal := <- iterator.iter
				if iterVal == StopIteration {
					break
				}
				result = i.interpFunDef(appFun, []Value{index, iterVal, iterator})
			}
			resultArr = append(resultArr, result)
			index++
		}
		retVal = &Array{len(resultArr), len(resultArr), resultArr}
	case *ast.CommandExp:
		newIter := &Iterator{}
		newIter.iter = make(chan Value)

		fmt.Println(node.Command, node.Args)
		cmd := exec.Command(node.Command, node.Args...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			retVal = Array{0, 0, []interface{}{}}
			break
		}
		err = cmd.Start()
		if err != nil {
			retVal = Array{0,0,[]interface{}{}}
			break
		}

		go func() {
			reader := bufio.NewReader(stdout)

			for {
				fmt.Println("reading")
				line, err := reader.ReadString('\n')
				fmt.Println("never read")
				line = strings.TrimRight(line, "\n")
				newIter.iter <- line
				if err != nil {
					newIter.iter <- StopIteration
					break
				}
			}
		}()

		retVal = newIter
	case *ast.AddSub:
		if node.Op == "+" {
			retVal = i.interpExp(node.Left).(int) + i.interpExp(node.Right).(int)
		} else if node.Op == "-" {
			retVal = i.interpExp(node.Left).(int) - i.interpExp(node.Right).(int)
		} else {
			panic("Unknown AddSub op")
		}
	case *ast.MulDiv:
		retVal = i.interpExp(node.Left).(int) * i.interpExp(node.Right).(int)
	case *ast.FunDef:
		retVal = node
	case *ast.FunApp:
		retVal = i.interpFunApp(node)
	case *ast.While:
		for i.interpExp(node.Cond).(int) != 0 {
			for _, line := range node.Body.Lines {
				i.interpExp(line)
			}
		}

		retVal = &Null{}
	case *ast.ReturnExp:
		retVal = i.interpExp(node.Target)
	case *ast.If:
		if i.interpExp(node.Cond).(int) != 0 {
			for _, line := range node.Body.Lines {
				i.interpExp(line)
			}
		}

		retVal = &Null{}
	case *ast.CompNode:
		retVal = i.interpComp(node)
	case *ast.ArrayLiteral:
		newArr := &Array{}
		newArr.length = node.Length
		newArr.cap = node.Length

		for k := 0; k < node.Length; k++ {
			newArr.arr = append(newArr.arr, i.interpExp(node.Exprs[k]))
		}

		retVal = newArr
	case *ast.SliceNode:
		arr := i.interpExp(node.Arr).(*Array)
		index := i.interpExp(node.Index).(int)
		retVal = arr.arr[index]
	case *ast.StrExp:
		retVal = node.Value
	case nil:
		panic("Interp on nil value")
	default:
		panic("Interp not defined for type: " + reflect.TypeOf(astNode).String())
	}

	return retVal
}

func (i *Interpreter) interpComp(comp *ast.CompNode) Value {
	left := i.interpExp(comp.Left).(int)
	right := i.interpExp(comp.Right).(int)

	var retVal bool

	switch comp.Op {
	case ">":
		retVal = left > right
	case "<":
		retVal = left < right
	case ">=":
		retVal = left >= right
	case "<=":
		retVal = left <= right
	case "==":
		retVal = left == right
	default:
		panic("Invalid comp op")
	}

	if retVal {
		return 1
	}
	return 0
}

func (i *Interpreter) interpFunDef(funExp ast.Node, args []Value) Value {
	primFunc, isPrimFunc := funExp.(PrimFunc)
	if isPrimFunc {
		return primFunc(args)
	}

	funVal := funExp.(*ast.FunDef)
	for k := 0; k < len(funVal.Args); k++ {
		argName := funVal.Args[k]
		i.Environ[argName.(*ast.Ident).Value] = args[k]
	}

	var lastVal Value
	for _, line := range funVal.Body.Lines {
		lastVal = i.interpExp(line)
	}

	return lastVal
}

func (i *Interpreter) interpFunApp(funApp *ast.FunApp) Value {
	funExp := i.interpExp(funApp.Fun)

	args := make([]Value, 0)
	for _, arg := range funApp.Args {
		args = append(args, i.interpExp(arg))
	}

	return i.interpFunDef(funExp, args)
}

func (i *Interpreter) Interp(p *ast.Program) {
	mainApp := &ast.FunApp{}
	mainApp.Fun = p.MainFunc
	mainApp.Args = make([]ast.Node, 0)
	i.CurrProgram = p
	i.interpFunApp(mainApp)
}

func CompareOutput(progText string, output string) bool {
	prog := parser.ParseProgram(progText)
	transform.TransformAst(prog)

	//_, err := TypeCheck(prog)
	// if err != nil {
	// 	log.Fatal("Program doesn't type check: " + err.Error())
	// 	return false
	// }
	fmt.Println(prog)
	i := NewInterpreter()
	i.Interp(prog)

	reference := strings.Trim(output, "\r\n")
	produced := strings.Trim(i.Output(), "\r\n")
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
	strs := make([]string, 0)

	for _, item := range items {
		strs = append(strs, fmt.Sprintf("%v", item))
	}

	return strs
}
