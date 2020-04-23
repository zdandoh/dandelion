package compile

import (
	"dandelion/types"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	lltypes "github.com/llir/llvm/ir/types"
)

func (c *Compiler) SetupCoro(entryBlock *ir.Block, coro *ir.Func, coroType types.CoroutineType) *ir.Block {
	newBody := coro.NewBlock("body")

	promiseType := c.PromiseType(coroType)

	// Setup promise
	promiseMem := entryBlock.NewAlloca(promiseType)
	voidPromise := entryBlock.NewBitCast(promiseMem, lltypes.I8Ptr)

	nullPtr := constant.NewNull(lltypes.I8Ptr)
	coroID := entryBlock.NewCall(CoroID, Zero, voidPromise, nullPtr, nullPtr)
	coroSize := entryBlock.NewCall(CoroSize)
	coroFrame := entryBlock.NewCall(Malloc, coroSize)

	coroHandle := entryBlock.NewCall(CoroBegin, coroID, coroFrame)
	coroHandle.ReturnAttrs = append(coroHandle.ReturnAttrs, enum.ReturnAttrNoAlias)

	suspendBlock := coro.NewBlock("suspend")
	suspendBlock.NewCall(CoroEnd, coroHandle, constant.NewBool(false))
	suspendBlock.NewRet(coroHandle)

	cleanupBlock := coro.NewBlock("cleanup")
	coroMem := cleanupBlock.NewCall(CoroFree, coroID, coroHandle)
	cleanupBlock.NewCall(Free, coroMem)
	cleanupBlock.NewBr(suspendBlock)

	// Initial suspend
	suspendRes := entryBlock.NewCall(CoroSuspend, constant.None, constant.NewBool(false))
	c.currBlock.NewSwitch(
		suspendRes,
		suspendBlock,
		ir.NewCase(constant.NewInt(lltypes.I8, 0), newBody),
		ir.NewCase(constant.NewInt(lltypes.I8, 1), cleanupBlock))

	newBody.NewBr(cleanupBlock)
	c.currCoro = &CoroState{cleanupBlock, suspendBlock, promiseMem}

	return newBody
}
