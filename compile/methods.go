package compile

import (
	"dandelion/ast"
	"dandelion/errs"
	"dandelion/types"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	lltypes "github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
	"reflect"
)

func (c *Compiler) strConcat(leftNode value.Value, rightNode value.Value) value.Value {
	// Load and calculate new length
	rightLenPtr := NewGetElementPtr(c.currBlock, rightNode, Zero, Zero)
	leftLenPtr := NewGetElementPtr(c.currBlock, leftNode, Zero, Zero)
	rightLen := NewLoad(c.currBlock, rightLenPtr)
	leftLen := NewLoad(c.currBlock, leftLenPtr)
	newLen := c.currBlock.NewAdd(rightLen, leftLen)

	strSize := GetSize(c.currBlock, StrType)
	totalLen := c.currBlock.NewAdd(newLen, strSize)

	newStrMem := c.currBlock.NewCall(Malloc, totalLen)
	newStr := c.currBlock.NewBitCast(newStrMem, lltypes.NewPointer(StrType))

	// Store new length
	newLenPtr := NewGetElementPtr(c.currBlock, newStr, Zero, Zero)
	c.currBlock.NewStore(newLen, newLenPtr)

	// Calculate str data pointer
	newStrDataPtr := NewGetElementPtr(c.currBlock, newStr, One)
	newStrDataPtr = c.currBlock.NewBitCast(newStrDataPtr, lltypes.I8Ptr)

	// Store new data pointer
	newDataPtr := NewGetElementPtr(c.currBlock, newStr, Zero, One)
	c.currBlock.NewStore(newStrDataPtr, newDataPtr)

	// Load old data pointers
	rightDataPtr := NewGetElementPtr(c.currBlock, rightNode, Zero, One)
	leftDataPtr := NewGetElementPtr(c.currBlock, leftNode, Zero, One)
	rightData := NewLoad(c.currBlock, rightDataPtr)
	leftData := NewLoad(c.currBlock, leftDataPtr)

	// Calculate offset
	offPtr := NewGetElementPtr(c.currBlock, newStrDataPtr, leftLen)

	// Memcpy the data
	c.currBlock.NewCall(MemCopy, newStrDataPtr, leftData, leftLen, constant.False)
	c.currBlock.NewCall(MemCopy, offPtr, rightData, rightLen, constant.False)

	return newStr
}

func (c *Compiler) listPush(list ast.Node, value ast.Node) value.Value {
	listVal := c.CompileNode(list)
	newVal := c.CompileNode(value)

	len := c.arrLen(listVal)
	cap := c.arrCap(listVal)

	newLen := c.currBlock.NewAdd(len, One)
	c.setArrLen(listVal, newLen)
	shouldResize := c.currBlock.NewICmp(enum.IPredUGT, newLen, cap)

	resizeBlock := c.currFun.NewBlock(c.getLabel("push_resize"))
	contBlock := c.currFun.NewBlock(c.getLabel("push_cont"))

	resizeBlock.NewBr(contBlock)
	contBlock.Term = c.currBlock.Term

	c.currBlock.NewCondBr(shouldResize, resizeBlock, contBlock)
	c.currBlock = resizeBlock

	// Setup resize block
	newCap := c.currBlock.NewAdd(cap, cap)
	c.setArrCap(listVal, newCap)

	// Calculate new arr size in bytes
	listType := c.GetType(list).(types.ArrayType)
	llSubtype := c.typeToLLType(listType.Subtype)
	newByteSize := c.currBlock.NewMul(GetSize(c.currBlock, llSubtype), c.currBlock.NewSExt(newCap, lltypes.I64))

	dataPtr := c.arrData(listVal)
	voidDataPtr := c.currBlock.NewBitCast(dataPtr, lltypes.I8Ptr)
	newDataPtr := c.currBlock.NewCall(Realloc, voidDataPtr, newByteSize)
	castDataPtr := c.currBlock.NewBitCast(newDataPtr, dataPtr.Type())
	c.setArrData(listVal, castDataPtr)

	c.currBlock = contBlock
	dataPtr = c.arrData(listVal)
	elemPtr := NewGetElementPtr(c.currBlock, dataPtr, len)
	c.currBlock.NewStore(newVal, elemPtr)

	return nil
}

func (c *Compiler) listPop(list ast.Node) value.Value {
	return nil
}

func (c *Compiler) compileBaseMethod(baseType ast.Node, methodName string, args []ast.Node) value.Value {
	ty := c.GetType(baseType)

	switch ty.(type) {
	case types.ArrayType:
		switch methodName {
		case "push":
			return c.listPush(baseType, args[0])
		case "pop":
			return c.listPop(baseType)
		}
	}
	return nil
}

func (c *Compiler) checkBaseMethod(node ast.Node) (ast.Node, string, bool) {
	structAccess, isStructAccess := node.(*ast.StructAccess)
	if !isStructAccess {
		return nil, "", false
	}

	targType := c.GetType(structAccess.Target)
	fieldName := structAccess.Field.(*ast.Ident).Value

	switch targType.(type) {
	case types.StructType:
		return nil, "", false
	case types.ArrayType:
		if types.HasMethod(types.ListMethods, fieldName) {
			return structAccess.Target, fieldName, true
		}
	}

	errs.Error(errs.ErrorValue, node, "base type '%s' doesn't have method '%s'", reflect.TypeOf(targType).Name(), fieldName)
	errs.CheckExit()

	return nil, "", false
}
