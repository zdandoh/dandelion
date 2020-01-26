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
		t.Fail()
	}
}

func TestFloatOps(t *testing.T) {
	src := `
data = 3.5;
data + 32.7;
33.3 / 4.2;
21.3 * 7.0;
return 0;
`

	if !CompileCheckExit(src, 0) {
		t.Fail()
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
		t.Fail()
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
		t.Fail()
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
		t.Fail()
	}
}

func TestArrayCompile(t *testing.T) {
	src := `
arr = [5, 6, 7];
return arr[1];
`

	if !CompileCheckExit(src, 6) {
		t.Fail()
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
		t.Fail()
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
		t.Fail()
	}
}

func TestArrayBasics(t *testing.T) {
	src := `
arr = [3, 4, 5, 6, 7];
arr[2] = 21;
return arr[2];
`

	if !CompileCheckExit(src, 21) {
		t.Fail()
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
		t.Fail()
	}
}

func TestString(t *testing.T) {
	src := `
string = "Look ma, a string!";
`

	if !CompileCheckExit(src, 0) {
		t.Fail()
	}
}

func TestNestedArray(t *testing.T) {
	src := `
[11, 12, 13][1];
arr = [[1, 2, 3], [4, 5, 6]];
return arr[1][1];
`

	if !CompileCheckExit(src, 5) {
		t.Fail()
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
		t.Fail()
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
		t.Fail()
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
		t.Fail()
	}
}

func TestCompileParens(t *testing.T) {
	src := `
x = (3 + 4) * 7;
return x;
`

	if !CompileCheckExit(src, 49) {
		t.Fail()
	}
}

func TestCompileTuple(t *testing.T) {
	src := `
x = ("a string", "another string", 4);
return x[2];
`

	if !CompileCheckExit(src, 4) {
		t.Fail()
	}
}

func TestReturnArr(t *testing.T) {
	src := `
fun = f() {
	[1, 2, 3, 4, 5];
};
arr = fun();
return arr[3];
`

	if !CompileCheckExit(src, 4) {
		t.Fail()
	}
}

func TestTupleFunc(t *testing.T) {
	src := `
fun = f() {
	(3, 4, "data");
};

return fun()[1];
`

	if !CompileCheckExit(src, 4) {
		t.Fail()
	}
}

// This doesn't work because main has an int return type, not i8
//func TestByteLiteral(t *testing.T) {
//	src := `
//x = 'h';
//y = 'g';
//
//return y;
//`
//
//	CompileCheckExit(src, 103)
//}

func TestBooleanLiteral(t *testing.T) {
	src := `
x = true;
y = false;
if x {
	return 43;
};
return 21;
`

	if !CompileCheckExit(src, 43) {
		t.Fail()
	}
}

func TestInfer(t *testing.T) {
	src := `
add = f(x, y) {
	x + y;
};

bob = f() {
	add;
};

unrelated = f(x, y, z) {
	p = z + 1;
	x + y + 3;
};

cob = bob;
return add(3, 4);
`
	if !CompileCheckExit(src, 7) {
		t.Fail()
	}
}

func TestInferAddable(t *testing.T) {
	src := `
flo = 4.6;
res = 3.0 + flo;
`

	if !CompileCheckExit(src, 0) {
		t.Fail()
	}
}

func TestInferSimpleFunc(t *testing.T) {
	src := `
fun = f(a, b) {
	a + b + 1;
};

return fun(1, 4);
`

	if !CompileCheckExit(src, 6) {
		t.Fail()
	}
}

func TestInferFunc(t *testing.T) {
	src := `
fun = f(a, b) {
	a + b + 1;
};

bun = f(c, d) {
	c + "string";
	d + 1;
};

gun = fun;

return gun(1, 4);
`

	if !CompileCheckExit(src, 6) {
		t.Fail()
	}
}

func TestInferMain(t *testing.T) {
	src := `
return 3 + 4;
`

	if !CompileCheckExit(src, 7) {
		t.Fail()
	}
}

func TestFuncTupleInfer(t *testing.T) {
	src := `
fun = f(x) {
	(x, 4, "string");
};

return fun("string")[1];
`

	if !CompileCheckExit(src, 4) {
		t.Fail()
	}
}

func TestInferArrSlice(t *testing.T) {
	src := `
arr = [1, 2, 3, 4, 5];
arr[2] = 90;
ret = arr[2] + arr[3];
return ret;
`

	if !CompileCheckExit(src, 94) {
		t.Fail()
	}
}

func TestAmbiguousInfer(t *testing.T) {
	src := `
fun = f(x) {
	x[3] = 7;
	x;
};

arr = [1, 2, 3, 4, 5];
return fun(arr)[3];
`

	if !CompileCheckExit(src, 7) {
		t.Fail()
	}
}

func TestClosure(t *testing.T) {
	src := `
x = 56;
fun = f(int one, string two) int {
	if 1 == 1 {
		return x;
	};
	one;
};

return fun(1, "two");
`

	if !CompileCheckExit(src, 56) {
		t.Fail()
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
//		t.Fail()
//	}
//}
