package transform

import (
	"ahead/parser"
	"fmt"
	"testing"
)

func TestRename(t *testing.T) {
	src := `
f{
	x = 5;
	if x > 10 {
		y = 30;
		x = x + 1;
	};
	y = y + 1;
};
`

	prog := parser.ParseProgram(src)
	RenameIdents(prog)
	fmt.Println(prog)
}
