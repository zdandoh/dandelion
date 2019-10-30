package compile

import (
	"testing"
)

func TestCompileBasic(t *testing.T) {
	src := `
a = 6;
b = 7;
d = a + b + 78;
return d;
`
	if !CompileCheckExit(src, 91) {
		t.FailNow()
	}
}

func TestCompileFunc(t *testing.T) {
	src := `
my_func = f(int a, int b) int {
	a + b;
};

d = my_func(21, 32);
return d;
`

	if !CompileCheckExit(src, 53) {
		t.FailNow()
	}
}

func TestCompileConditional(t *testing.T) {
	src := `
x = 100;
while x > 36 {
	x = x - 1;
};
return x;
`

	if !CompileCheckExit(src, 36) {
		t.FailNow()
	}
}

func TestArrayCompile(t *testing.T) {
	src := `
arr = [5, 6, 7];
return arr[1];
`

	if !CompileCheckExit(src, 6) {
		t.FailNow()
	}
}

func TestNestedIf(t *testing.T) {
	src := `
x = 7;
if x > 5 {
	if x > 6 {
		x = 20;
	};
};

return x;
`

	if !CompileCheckExit(src, 20) {
		t.FailNow()
	}
}

func TestControlFlowArray(t *testing.T) {
	src := `
arr = [5, 6, 7, 8, 9];
x = 0;
index = 0;
sum = 0;
while x < 100 {
	sum = sum + arr[index];
	index = index + 1;
	if index > 4 {
		index = 0;
	};
	x = x + 1;
};

return sum;
`

	if !CompileCheckExit(src, 700) {
		t.FailNow()
	}
}
