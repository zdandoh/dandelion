package compile

import (
	"dandelion/types"
)

type TypeTable map[types.TypeHash]int

func (t TypeTable) GetNo(ty types.Type) int {
	hash := types.HashType(ty)
	tno := t[hash]
	return tno
}

func (t TypeTable) Add(ty types.Type) {
	_, isAny := ty.(types.AnyType)
	if isAny {
		// Don't add the any type to the table
		return
	}

	hash := types.HashType(ty)
	_, exists := t[hash]
	if !exists {
		t[hash] = len(t) + 1
	}
}

func (c *Compiler) SetupTypeTable() {
	typeTable := make(TypeTable)

	for _, progType := range c.Types {
		typeTable.Add(progType)
	}

	for _, refType := range c.prog.RefTypes {
		typeTable.Add(refType)
	}

	c.typeTable = typeTable
}
