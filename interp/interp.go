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
	ValueString() string
}

type Addable interface {
	Add(Value) Value
}

type Int int

func (n Int) ValueString() string {
	return fmt.Sprintf("%d", int(n))
}

func (n Int) Add(other Value) Value {
	return n + other.(Int)
}

type String string

func (n String) ValueString() string {
	return string(n)
}

func (n String) Add(other Value) Value {
	return n + other.(String)
}

type Null struct {
}

func (n *Null) ValueString() string {
	return "null"
}

type PrimFunc struct {
	Target func([]Value) Value
}

func (n *PrimFunc) ValueString() string {
	return "<builtin>"
}

type Array struct {
	length int
	cap    int
	arr    []Value
}

func (n *Array) ValueString() string {
	strs := make([]string, 0)
	for _, item := range n.arr {
		strs = append(strs, item.ValueString())
	}
	return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
}

type Closure struct {
	Body *ast.Block
	Args []ast.Node
	Env  Env
}

func (n *Closure) ValueString() string {
	return "<closure>"
}

type Struct struct {
	Members map[string]Value
}

func (s *Struct) ValueString() string {
	memStrs := make([]string, 0)
	for _, mem := range s.Members {
		memStrs = append(memStrs, mem.ValueString())
	}
	return fmt.Sprintf("<struct>(%s)", strings.Join(memStrs, ", "))
}

type Iterator struct {
	iter chan interface{}
}

func (iter *Iterator) ValueString() string {
	return "<iterator object>"
}

var StopIteration = errors.New("Stop iteration")
var ReturnExp = errors.New("Return expression")

type Env map[string]int
type Memory struct {
	currEnv  Env
	store    map[int]Value
	envStack []Env
	ptr      int
}

func (m *Memory) Bind(ident string, val Value) {
	num, ok := m.currEnv[ident]
	if ok {
		m.store[num] = val
		return
	}

	m.ptr++
	m.currEnv[ident] = m.ptr
	m.store[m.ptr] = val
}

func (m *Memory) Get(ident string) (Value, bool) {
	num, ok := m.currEnv[ident]
	if !ok {
		return nil, false
	}
	return m.store[num], true
}

func (e Env) Copy() Env {
	newEnv := make(Env)
	for key, value := range e {
		newEnv[key] = value
	}

	return newEnv
}

func (m *Memory) PushEnv(env Env) {
	m.envStack = append(m.envStack, env)
	m.currEnv = env
}

func (m *Memory) PopEnv() {
	m.envStack = m.envStack[:len(m.envStack)-1]
	m.currEnv = m.envStack[len(m.envStack)-1]
}

type Interpreter struct {
	Env         *Memory
	CurrProgram *ast.Program
	output      string
}

func NewInterpreter() *Interpreter {
	i := &Interpreter{}
	i.Env = NewMemory(i)

	return i
}

func (i *Interpreter) Output() string {
	return i.output
}

func NewMemory(i *Interpreter) *Memory {
	mem := &Memory{}

	defaultEnv := make(Env)
	mem.PushEnv(defaultEnv)
	mem.store = make(map[int]Value)

	mem.Bind("p", &PrimFunc{func(values []Value) Value {
		for _, value := range values {
			strVal := value.ValueString()
			fmt.Print(strVal)
			i.output += strVal
		}
		i.output += "\n"
		if len(i.output) > 8192 {
			// Prevent output from getting unnecessarily huge
			i.output = ""
		}
		fmt.Println()
		return &Null{}
	}})

	return mem
}

func (i *Interpreter) interpExp(astNode ast.Node) (Value, error) {
	var retVal Value
	var ctrl error

	switch node := astNode.(type) {
	case *ast.Assign:
		assignNode, ctrl := i.interpExp(node.Expr)
		if ctrl != nil {
			return assignNode, ctrl
		}

		switch target := node.Target.(type) {
		case *ast.Ident:
			i.Env.Bind(target.Value, assignNode)
		case *ast.StructAccess:
			targetStruct, ctrl := i.interpExp(target.Target)
			if ctrl != nil {
				return targetStruct, ctrl
			}
			targetStruct.(*Struct).Members[target.Field.(*ast.Ident).Value] = assignNode
		}

		retVal = &Null{}
	case *ast.Num:
		res, _ := strconv.Atoi(node.Value)
		retVal = Int(res)
	case *ast.Ident:
		val, ok := i.Env.Get(node.Value)
		if !ok {
			// Might be a function definition
			fun, ok := i.CurrProgram.Funcs[node.Value]
			if !ok {
				panic("Unbound variable " + node.Value)
			}
			val, ctrl = i.interpExp(fun)
			if ctrl != nil {
				return val, ctrl
			}
		}

		retVal = val
	case *ast.StructInstance:
		newStruct := &Struct{make(map[string]Value)}
		for key, val := range node.DefaultValues {
			interpVal, ctrl := i.interpExp(val)
			if ctrl != nil {
				return interpVal, ctrl
			}
			newStruct.Members[key.Value] = interpVal
		}

		retVal = newStruct
	case *ast.StructDef:
		panic("Struct def shouldn't be interpreted")
	case *ast.StructAccess:
		structVal, ctrl := i.interpExp(node.Target)
		if ctrl != nil {
			return structVal, ctrl
		}

		val, ok := structVal.(*Struct).Members[node.Field.(*ast.Ident).Value]
		if !ok {
			panic(fmt.Sprintf("Struct has no member '%s'", node.Field))
		}

		retVal = val
	case *ast.PipeExp:
		iter, ctrl := i.interpExp(node.Left)
		if ctrl != nil {
			return iter, ctrl
		}

		arr, isArr := iter.(*Array)
		iterator, isIter := iter.(*Iterator)

		appFun, ctrl := i.interpExp(node.Right)
		if ctrl != nil {
			return appFun, ctrl
		}

		resultArr := make([]Value, 0)
		index := 0
		for {
			var result Value
			if isArr {
				if index >= arr.length {
					break
				}
				result, ctrl = i.interpFunDef(appFun, []Value{Int(index), arr.arr[index], arr})
				if ctrl != nil {
					return result, ctrl
				}

			} else if isIter {
				iterVal := <-iterator.iter
				if iterVal == StopIteration {
					break
				}
				result, ctrl = i.interpFunDef(appFun, []Value{Int(index), iterVal.(Value), iterator})
				if ctrl != nil {
					return result, ctrl
				}
			}
			resultArr = append(resultArr, result)
			index++
		}
		retVal = &Array{len(resultArr), len(resultArr), resultArr}
	case *ast.CommandExp:
		newIter := &Iterator{}
		newIter.iter = make(chan interface{})

		cmd := exec.Command(node.Command, node.Args...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			retVal = &Array{0, 0, []Value{}}
			break
		}
		err = cmd.Start()
		if err != nil {
			retVal = &Array{0, 0, []Value{}}
			break
		}

		go func() {
			reader := bufio.NewReader(stdout)

			for {
				line, err := reader.ReadString('\n')
				line = strings.TrimRight(line, "\n")
				newIter.iter <- String(line)
				if err != nil {
					newIter.iter <- StopIteration
					break
				}
			}
		}()

		retVal = newIter
	case *ast.AddSub:
		if node.Op == "+" {
			left, ctrl := i.interpExp(node.Left)
			if ctrl != nil {
				return left, ctrl
			}
			right, ctrl := i.interpExp(node.Right)
			if ctrl != nil {
				return right, ctrl
			}
			addable, ok := left.(Addable)
			if ok {
				retVal = addable.Add(right)
			} else {
				panic("type is not addable")
			}
		} else if node.Op == "-" {
			left, ctrl := i.interpExp(node.Left)
			if ctrl != nil {
				return left, ctrl
			}
			right, ctrl := i.interpExp(node.Right)
			if ctrl != nil {
				return right, ctrl
			}
			retVal = left.(Int) - right.(Int)
		} else {
			panic("Unknown AddSub op")
		}
	case *ast.MulDiv:
		left, ctrl := i.interpExp(node.Left)
		if ctrl != nil {
			return left, ctrl
		}
		right, ctrl := i.interpExp(node.Right)
		if ctrl != nil {
			return right, ctrl
		}

		if node.Op == "*" {
			retVal = left.(Int) * right.(Int)
		} else {
			retVal = left.(Int) / right.(Int)
		}
	case *ast.Mod:
		left, ctrl := i.interpExp(node.Left)
		if ctrl != nil {
			return left, ctrl
		}
		right, ctrl := i.interpExp(node.Right)
		if ctrl != nil {
			return right, ctrl
		}

		retVal = left.(Int) % right.(Int)
	case *ast.FunDef:
		retVal = &Closure{node.Body, node.Args, i.Env.currEnv.Copy()}
	case *ast.FunApp:
		retVal, ctrl = i.interpFunApp(node)
		if ctrl != nil {
			return retVal, ctrl
		}
	case *ast.While:
		for {
			condVal, ctrl := i.interpExp(node.Cond)
			if ctrl != nil {
				return condVal, ctrl
			}
			if condVal.(Int) == 0 {
				break
			}
			for _, line := range node.Body.Lines {
				lineVal, ctrl := i.interpExp(line)
				if ctrl != nil {
					return lineVal, ctrl
				}
			}
		}

		retVal = &Null{}
	case *ast.ReturnExp:
		retVal, ctrl = i.interpExp(node.Target)
		if ctrl != nil {
			return retVal, ctrl
		}
		return retVal, ReturnExp
	case *ast.If:
		condVal, ctrl := i.interpExp(node.Cond)
		if ctrl != nil {
			return condVal, ctrl
		}

		if condVal.(Int) != 0 {
			for _, line := range node.Body.Lines {
				lineVal, ctrl := i.interpExp(line)
				if ctrl != nil {
					return lineVal, ctrl
				}
			}
		}

		retVal = &Null{}
	case *ast.CompNode:
		retVal, ctrl = i.interpComp(node)
		if ctrl != nil {
			return retVal, ctrl
		}
	case *ast.ArrayLiteral:
		newArr := &Array{}
		newArr.length = node.Length
		newArr.cap = node.Length

		for k := 0; k < node.Length; k++ {
			arrExp, ctrl := i.interpExp(node.Exprs[k])
			if ctrl != nil {
				return arrExp, ctrl
			}
			newArr.arr = append(newArr.arr, arrExp)
		}

		retVal = newArr
	case *ast.SliceNode:
		arrNode, ctrl := i.interpExp(node.Arr)
		if ctrl != nil {
			return arrNode, ctrl
		}

		indexNode, ctrl := i.interpExp(node.Index)
		if ctrl != nil {
			return indexNode, ctrl
		}

		retVal = arrNode.(*Array).arr[indexNode.(Int)]
	case *ast.StrExp:
		retVal = String(node.Value)
	case nil:
		panic("Interp on nil value")
	default:
		panic("Interp not defined for type: " + reflect.TypeOf(astNode).String())
	}

	return retVal, ctrl
}

func (i *Interpreter) interpComp(comp *ast.CompNode) (Value, error) {
	leftNode, ctrl := i.interpExp(comp.Left)
	if ctrl != nil {
		return leftNode, ctrl
	}

	rightNode, ctrl := i.interpExp(comp.Right)
	if ctrl != nil {
		return rightNode, ctrl
	}

	left := leftNode.(Int)
	right := rightNode.(Int)
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
		return Int(1), nil
	}
	return Int(0), nil
}

func (i *Interpreter) interpFunDef(funExp ast.Node, args []Value) (Value, error) {
	primFunc, isPrimFunc := funExp.(*PrimFunc)
	if isPrimFunc {
		return primFunc.Target(args), nil
	}

	cloVal := funExp.(*Closure)
	i.Env.PushEnv(cloVal.Env.Copy())
	for k := 0; k < len(cloVal.Args); k++ {
		argName := cloVal.Args[k]
		i.Env.Bind(argName.(*ast.Ident).Value, args[k])
	}

	var lastVal Value
	var ctrl error
	for _, line := range cloVal.Body.Lines {
		lastVal, ctrl = i.interpExp(line)
		if ctrl == ReturnExp {
			return lastVal, nil
		} else if ctrl != nil {
			return lastVal, ctrl
		}
	}
	i.Env.PopEnv()

	return lastVal, nil
}

func (i *Interpreter) interpFunApp(funApp *ast.FunApp) (Value, error) {
	funExp, ctrl := i.interpExp(funApp.Fun)
	if ctrl != nil {
		return funExp, ctrl
	}

	args := make([]Value, 0)
	for _, arg := range funApp.Args {
		argVal, ctrl := i.interpExp(arg)
		if ctrl != nil {
			return argVal, ctrl
		}
		args = append(args, argVal)
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
