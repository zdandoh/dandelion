package infer

import (
	"dandelion/types"
	"encoding/gob"
	"fmt"
	"strings"
)

func init() {
	gob.Register(TypeFunc{})
}

type TypeExpr interface {
	ExprString() string
}

type StorableType interface {
	Key() StoreKey
	StorableString() string
}

type StoreKey interface {
	KeyString() string
}

type TypeRef int
type TypeVar int

func (t TypeVar) Key() StoreKey {
	return t
}

func (t TypeVar) String() string {
	return fmt.Sprintf("t%d", t)
}

func (t TypeVar) KeyString() string {
	return t.String()
}

func (t TypeVar) ExprString() string {
	return t.String()
}

func (t TypeVar) StorableString() string {
	return t.String()
}

func (t TypeRef) String() string {
	return fmt.Sprintf("<type ref #%d>", t)
}

func (t TypeRef) ExprString() string {
	return t.String()
}

type FuncKind string
const (
	KindFunc FuncKind = "f"
	KindArray FuncKind = "a"
	KindTuple FuncKind = "tup"
	KindTupleAccess FuncKind = "ta"
	KindPropAccess FuncKind = "prop"
	KindPropOwner FuncKind = "own"
	KindStructInstance FuncKind = "struct"
	KindCoro FuncKind = "coro"
	KindContainer FuncKind = "cont"
)

// ReducibleKinds are types of functions that can be replaced with simpler types
// ie; a type var or base type. Certain kinds cannot be reducible or type inference
// data will be lost.
var ReducibleKinds = map[FuncKind]bool {
	KindTupleAccess: true,
	KindPropAccess: true,
	KindContainer: true, // TODO think about if this might be a bug
	KindPropOwner: true,
}

type TypeFunc struct {
	Args []TypeRef
	Ret TypeRef
	Kind FuncKind
	ID int
}

func (i TypeFunc) Reducible() bool {
	return ReducibleKinds[i.Kind]
}

func (i TypeFunc) String() string {
	argStrs := make([]string, 0)
	for _, arg := range i.Args {
		argStrs = append(argStrs, arg.ExprString())
	}

	return fmt.Sprintf("(%s) -> %s", strings.Join(argStrs, ", "), i.Ret)
}

func (i TypeFunc) ExprString() string {
	return i.String()
}

func (i TypeFunc) StorableString() string {
	return i.String()
}

type FuncMeta struct {
	ID int
	data interface{}
}

type FuncMetaKey string

func (f FuncMeta) String() string {
	return fmt.Sprintf("meta(%v)", f.data)
}

func (f FuncMeta) ExprString() string {
	return f.String()
}

func (f FuncMeta) StorableString() string {
	return f.String()
}

func (f FuncMeta) Key() StoreKey {
	return FuncMetaKey(f.String())
}

func (f FuncMetaKey) KeyString() string {
	return fmt.Sprintf("meta#%s", f)
}

type FuncKey string

func (i TypeFunc) Key() StoreKey {
	return FuncKey(fmt.Sprintf("func#%d", i.ID))
}

func (k FuncKey) KeyString() string {
	return fmt.Sprintf("func#%s", k)
}

type TypeBase struct {
	Type types.Type
}

func (b TypeBase) Key() StoreKey {
	return b
}

func (b TypeBase) String() string {
	return fmt.Sprintf("<%s>", b.Type.TypeString())
}

func (b TypeBase) KeyString() string {
	return b.String()
}

func (b TypeBase) StorableString() string {
	return b.String()
}

func (b TypeBase) ExprString() string {
	return b.String()
}

type TCons struct {
	Left TypeRef
	Right TypeRef
}

//func (t *TCons) String() string {
//	return fmt.Sprintf("%s = %s", t.Left.ExprString(), t.Right.ExprString())
//}