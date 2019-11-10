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

func TestCompileFunc2(t *testing.T) {
	src := `
fun = f(int a, string b) int {
	a;
};

return fun(4, "data");
`

	if !CompileCheckExit(src, 4) {
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

condfun = f(int x) int {
	if x < 6 {
		x = 7;
		if x < 5 {
			x = 5;
			if x < 3 {
				x = 3;
			};
		};
	};

	return x;
};
return condfun(5) + condfun(2) + condfun(6);
`

	if !CompileCheckExit(src, 20) {
		t.FailNow()
	}
}

func TestNestedWhile(t *testing.T) {
	src := `
x = 0;
y = 0;
while x < 100 {
	y = 0;
	while y < 100 {
		y = y + 1;
	};
	x = x + 1;
};

return x + y;
`

	if !CompileCheckExit(src, 200) {
		t.FailNow()
	}
}

func TestArrayBasics(t *testing.T) {
	src := `
arr = [3, 4, 5, 6, 7];
arr[2] = 21;
return arr[2];
`

	if !CompileCheckExit(src, 21) {
		t.FailNow()
	}
}

func TestControlFlowArray(t *testing.T) {
	src := `
arr = [5, 6, 7, 8, 9];
x = 0;
index = 0;
sum = 0;
while x < 20 {
	sum = sum + arr[index];
	index = index + 1;
	if index > 4 {
		index = 0;
	};
	x = x + 1;
};

return sum;
`

	if !CompileCheckExit(src, 140) {
		t.FailNow()
	}
}

func TestString(t *testing.T) {
	src := `
string = "Look ma, a string!";
`

	if !CompileCheckExit(src, 0) {
		t.FailNow()
	}
}

func TestNestedArray(t *testing.T) {
	src := `
arr = [[1, 2, 3], [4, 5, 6]];
return arr[1][1];
`

	if !CompileCheckExit(src, 5) {
		t.FailNow()
	}
}

func TestCompileStruct(t *testing.T) {
	src := `
struct Line {
	string value;
	int num;
};

l1 = Line("this is the contents of my line", 3);
return l1.num;
`

	if !CompileCheckExit(src, 3) {
		t.FailNow()
	}
}

func TestAnonStruct(t *testing.T) {
	src := `
l1 = struct {
	string value;
	int num;
}("very anonymous, very cool", 54);
return l1.num;
`

	if !CompileCheckExit(src, 54) {
		t.FailNow()
	}
}

func TestStructAssign(t *testing.T) {
	src := `
struct Line {
	string value;
	int num;
};

l1 = Line("hi mr. line", 5);
l1.num = 71;
l1.value = "a new string, intense.";

return l1.num;
`

	if !CompileCheckExit(src, 71) {
		t.FailNow()
	}
}

func TestCompileParens(t *testing.T) {
	src := `
x = (3 + 4) * 7;
return x;
`

	if !CompileCheckExit(src, 49) {
		t.FailNow()
	}
}

func TestCompileTuple(t *testing.T) {
	src := `
x = (3, "string", 4);
return x[2];
`

	if !CompileCheckExit(src, 4) {
		t.FailNow()
	}
}

func TestTupleFunc(t *testing.T) {
	src := `
fun = f(int x) (int, int, string) {
	(3, 4, "data");
};

return fun(111)[0];
`

	if !CompileCheckExit(src, 3) {
		t.FailNow()
	}
}

//func TestPipeline(t *testing.T) {
//	src := `
//adder = f(int e, int i, int[] a) int {
//	e * 2;
//};
//newarr = [1, 2, 3] -> adder;
//return newarr[2];
//`
//
//	if !CompileCheckExit(src, 6) {
//		t.FailNow()
//	}
//}
