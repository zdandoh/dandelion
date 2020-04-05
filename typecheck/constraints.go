package typecheck

import (
	"bytes"
	"crypto/sha256"
	"dandelion/types"
	"encoding/gob"
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

func init() {
	gob.Register(TypeVar(1))
	gob.Register(BaseType{})
	gob.Register(Coroutine{})
	gob.Register(StructOptions{})
	gob.Register(Container{})
	gob.Register(Fun{})
	gob.Register(Tup{})
}

type Constrainable interface {
	ConsString() string
}

type consBox struct {
	cons []Constraint
}

type ConsHash string
type TypeVar int

func (t TypeVar) String() string {
	return fmt.Sprintf("t%d", t)
}

func (t TypeVar) ConsString() string {
	return t.String()
}

type Constraint struct {
	Left  Constrainable
	Right Constrainable
}

func (c Constraint) String() string {
	return fmt.Sprintf("%s = %s", c.Left.ConsString(), c.Right.ConsString())
}

type BaseType struct {
	types.Type
}

func (t BaseType) ConsString() string {
	return t.Type.TypeString()
}

type Coroutine struct {
	Yields Constrainable
	Reads  Constrainable
}

func (c Coroutine) ConsString() string {
	return fmt.Sprintf("<coroutine(%s) -> %s>", c.Reads.ConsString(), c.Yields.ConsString())
}

type StructOptions struct {
	Types      []types.Type
	Dependants map[TypeVar]string
}

func (o StructOptions) ConsString() string {
	typeStrings := make([]string, 0)
	depStrings := make([]string, 0)

	for _, t := range o.Types {
		typeStrings = append(typeStrings, t.TypeString())
	}
	for k, v := range o.Dependants {
		depStrings = append(depStrings, fmt.Sprintf("%v: %v", k, v))
	}

	return fmt.Sprintf("struct-options[%s]<%s>", strings.Join(typeStrings, ", "), strings.Join(depStrings, ", "))
}

type Container struct {
	Type    types.Type
	Subtype Constrainable
	Index   int
	ID      int
}

func (c Container) ConsString() string {
	return fmt.Sprintf("container<%v, id:%d>[%v]#%d", c.Type.TypeString(), c.ID, c.Subtype.ConsString(), c.Index)
}

type Tup struct {
	Subtypes []Constrainable
}

func (t Tup) ConsString() string {
	subStrings := make([]string, 0)
	for _, sub := range t.Subtypes {
		subStrings = append(subStrings, sub.ConsString())
	}
	return fmt.Sprintf("(%s)", strings.Join(subStrings, ", "))
}

type Fun struct {
	Args []Constrainable
	Ret  Constrainable
}

func (t Fun) ConsString() string {
	varStrings := make([]string, 0)
	for _, tVar := range t.Args {
		varStrings = append(varStrings, tVar.ConsString())
	}

	return fmt.Sprintf("(%s -> %s)", strings.Join(varStrings, ", "), t.Ret.ConsString())
}

func HashCons(c Constrainable) ConsHash {
	cont, isContainer := c.(Container)
	if isContainer {
		return ConsHash(fmt.Sprintf("container%d", cont.ID))
	}

	b := bytes.NewBuffer(nil)
	err := gob.NewEncoder(b).Encode(c)
	if err != nil {
		panic(errors.Wrap(err, "couldn't hash constrainable"))
	}

	sha := sha256.New()
	sha.Write(b.Bytes())
	return ConsHash(sha.Sum(nil))
}
