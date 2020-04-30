package compile

import (
	"github.com/llir/llvm/ir"
	lltypes "github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

var InitTrampoline value.Value
var AdjustTrampoline value.Value
var AllocClo value.Value
var Malloc value.Value
var Free value.Value
var OpenF value.Value

// Coroutine intrinsics
var CoroID value.Value
var CoroSize value.Value
var CoroBegin value.Value
var CoroFree value.Value
var CoroEnd value.Value
var CoroSuspend value.Value
var CoroSave value.Value
var CoroDone value.Value
var CoroResume value.Value
var CoroDestroy value.Value
var CoroPromise value.Value
var Print value.Value
var PrintB value.Value
var PrintP value.Value
var ThrowEx value.Value

func (c *Compiler) setupIntrinsics() {
	Print = c.mod.NewFunc(
		"print",
		lltypes.Void,
		ir.NewParam("d", lltypes.I32))
	PrintB = c.mod.NewFunc(
		"printb",
		lltypes.Void,
		ir.NewParam("b", lltypes.I1))
	PrintP = c.mod.NewFunc(
		"printp",
		lltypes.Void,
		ir.NewParam("p", lltypes.I8Ptr))
	InitTrampoline = c.mod.NewFunc(
		"llvm.init.trampoline",
		lltypes.Void,
		ir.NewParam("tramp", lltypes.I8Ptr),
		ir.NewParam("func", lltypes.I8Ptr),
		ir.NewParam("nval", lltypes.I8Ptr))
	AdjustTrampoline = c.mod.NewFunc(
		"llvm.adjust.trampoline",
		lltypes.I8Ptr,
		ir.NewParam("tramp", lltypes.I8Ptr))
	AllocClo = c.mod.NewFunc(
		"alloc_clo",
		lltypes.I8Ptr)
	ThrowEx = c.mod.NewFunc("throwex", lltypes.Void, ir.NewParam("exno", lltypes.I32))
	Malloc = c.mod.NewFunc(
		"malloc",
		lltypes.I8Ptr,
		ir.NewParam("size", lltypes.I64))
	Free = c.mod.NewFunc(
		"free",
		lltypes.Void,
		ir.NewParam("ptr", lltypes.I8Ptr))
	OpenF = c.mod.NewFunc(
		"d_open",
		lltypes.I32,
		ir.NewParam("ptr", lltypes.NewPointer(StrType)),
		ir.NewParam("ptr2", lltypes.NewPointer(StrType)))
	CoroID = c.mod.NewFunc(
		"llvm.coro.id",
		lltypes.Token,
		ir.NewParam("align", lltypes.I32),
		ir.NewParam("promise", lltypes.I8Ptr),
		ir.NewParam("coroaddr", lltypes.I8Ptr),
		ir.NewParam("fnaddrs", lltypes.I8Ptr))
	CoroSize = c.mod.NewFunc(
		"llvm.coro.size.i64",
		lltypes.I64)
	CoroBegin = c.mod.NewFunc(
		"llvm.coro.begin",
		lltypes.I8Ptr,
		ir.NewParam("id", lltypes.Token),
		ir.NewParam("mem", lltypes.I8Ptr))
	CoroFree = c.mod.NewFunc(
		"llvm.coro.free",
		lltypes.I8Ptr,
		ir.NewParam("id", lltypes.Token),
		ir.NewParam("frame", lltypes.I8Ptr))
	CoroEnd = c.mod.NewFunc(
		"llvm.coro.end",
		lltypes.I1,
		ir.NewParam("handle", lltypes.I8Ptr),
		ir.NewParam("unwind", lltypes.I1))
	CoroSuspend = c.mod.NewFunc(
		"llvm.coro.suspend",
		lltypes.I8,
		ir.NewParam("save", lltypes.Token),
		ir.NewParam("final", lltypes.I1))
	CoroSave = c.mod.NewFunc(
		"llvm.coro.save",
		lltypes.Token,
		ir.NewParam("handle", lltypes.I8Ptr))
	CoroResume = c.mod.NewFunc(
		"llvm.coro.resume",
		lltypes.Void,
		ir.NewParam("handle", lltypes.I8Ptr))
	CoroDestroy = c.mod.NewFunc(
		"llvm.coro.destroy",
		lltypes.Void,
		ir.NewParam("handle", lltypes.I8Ptr))
	CoroPromise = c.mod.NewFunc(
		"llvm.coro.promise",
		lltypes.I8Ptr,
		ir.NewParam("ptr", lltypes.I8Ptr),
		ir.NewParam("align", lltypes.I32),
		ir.NewParam("from", lltypes.I1))
	CoroDone = c.mod.NewFunc(
		"llvm.coro.done",
		lltypes.I1,
		ir.NewParam("handle", lltypes.I8Ptr),
	)
}
