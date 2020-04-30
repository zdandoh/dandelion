package compile

import (
	"dandelion/types"
	lltypes "github.com/llir/llvm/ir/types"
)

type TypeTable map[types.TypeHash]*TypePair

type TypePair struct {
	TypeNo int
	Type   lltypes.Type
}

func (t TypeTable) GetNo(ty types.Type) int {
	hash := types.HashType(ty)
	pair := t[hash]
	return pair.TypeNo
}

func (c *Compiler) SetupTypeTable() {
	typeTable := make(TypeTable)

	typeNo := 0
	for _, progType := range c.Types {
		_, isAny := progType.(*types.AnyType)
		if isAny {
			// Don't add the any type to the table
			continue
		}

		hash := types.HashType(progType)
		_, exists := typeTable[hash]
		if !exists {
			typeNo++
			typeTable[hash] = &TypePair{typeNo, c.typeToLLType(progType)}
		}
	}

	c.typeTable = typeTable
}
