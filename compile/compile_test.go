package compile

import (
	"fmt"
	"sort"
	"strings"
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

func TestCallGlobal(t *testing.T) {
	src := `
other = f(a) {
	a + 5;
};

my_func = f(a, b) {
	other(a * b);
};

d = my_func(3, 8);
return d;
`

	if !CompileCheckExit(src, 29) {
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
arr = [5, 6, 7, 10, 20, 30];
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

func TestModulo(t *testing.T) {
	src := `
x = 12 % 3;
y = 14 % 3;
return x + y;
`

	if !CompileCheckExit(src, 2) {
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

func TestInferStruct(t *testing.T) {
	src := `
struct Line {
	string value;
	int num;
};

fun = f(x) {
	x.num;
};

return fun(Line("some data", 7));
`

	if !CompileCheckExit(src, 7) {
		t.Fail()
	}
}

func TestInferStructChoice(t *testing.T) {
	src := `
struct Line {
	string value;
	int num;
};

struct Pie {
	string value;
	int flavor;
};

fun = f(x) {
	x.value;
	x.flavor;
};
return 3;
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

func TestNestedType(t *testing.T) {
	src := `
x = 5;
fun = f() {
	fun2 = f() {
		fun3 = f() {
			return x;
		};
		return fun3;
	};
	return fun2;
};

return fun()()();
`

	if !CompileCheckExit(src, 5) {
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

func TestInferArray2(t *testing.T) {
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

func TestInferEmptyArray(t *testing.T) {
	src := `
empty = [];
empty[0] = 7;

empty2 = [];
empty2[0] = "string";

return 3;
`

	if !CompileCheckExit(src, 3) {
		t.Fail()
	}
}

func TestInferTupleArray(t *testing.T) {
	src := `
tup_arr = [(3, 4), (6, 7), (8, 9)];
return tup_arr[2][0];
`

	if !CompileCheckExit(src, 8) {
		t.Fail()
	}
}

func TestMutableStruct(t *testing.T) {
	src := `
struct Line {
	int num;
};

l = Line(45);
fun = f(stru) {
	stru.num = 33;
};

fun(l);

return l.num;
`

	if !CompileCheckExit(src, 33) {
		t.Fail()
	}
}

func TestClosure(t *testing.T) {
	src := `
x = 22;
y = 6;
fun = f() {
	x + y + 1;
};

return fun();
`

	if !CompileCheckExit(src, 29) {
		t.Fail()
	}
}

func TestPassClosure(t *testing.T) {
	src := `
x = 56;
fun = f() {
	x + 1;
};

other = f(take) {
	take();
};
return other(fun);
`

	if !CompileCheckExit(src, 57) {
		t.Fail()
	}
}

func TestArrClosure(t *testing.T) {
	src := `
arr = [1, 2, 3, 4];
fun = f() {
	arr[2] = 9;
};

fun();
return arr[2];
`

	if !CompileCheckExit(src, 9) {
		t.Fail()
	}
}

func TestMutableStructClosure(t *testing.T) {
	src := `

struct Point {
	int x;
	int y;
};

p = Point(45, 21);

fun = f() {
	p.x = 12;
};

fun();
return p.x;
`

	if !CompileCheckExit(src, 12) {
		t.Fail()
	}
}

func TestNestedClosure(t *testing.T) {
	src := `
x = 34;
fun = f() {
	inner = f() {
		more = f() {
			76;
		};

		hoho = f() {
			32;
		};

		x + 1;
	};
	inner;
};

capture = fun();
return capture();
`

	if !CompileCheckExit(src, 35) {
		t.Fail()
	}
}

func TestNestedClosure2(t *testing.T) {
	src := `
fun = f() {
	x = 34;
	inner = f() {
		x + 1;
	};
	inner;
};

return fun()();
`

	if !CompileCheckExit(src, 35) {
		t.Fail()
	}
}

func TestNestedClosure3(t *testing.T) {
	src := `
fun = f(x) {
	inner = f() {
		x + 1;
	};
	inner;
};

return fun(21)() + fun(5)();
`

	if !CompileCheckExit(src, 28) {
		t.Fail()
	}
}

func TestMethod(t *testing.T) {
	src := `
struct Point {
	int x;
	int y;
};

Point.mul = f() {
	x = 11;
	x * y;
};

p = Point(4, 5);
lol = p.mul();
return lol + p.x;
`

	if !CompileCheckExit(src, 66) {
		t.Fail()
	}
}

func TestMethod2(t *testing.T) {
	src := `
struct Point {
	int x;
	int y;
};

Point.dot = f(other) {
	x * other.x + y * other.y;
};

p = Point(3, 5);
p2 = Point(2, 7);
lol = p.dot(p2);
return lol;
`

	if !CompileCheckExit(src, 41) {
		t.Fail()
	}
}

func TestBoundMethod(t *testing.T) {
	src := `
struct Point {
	int x;
	int y;
};

Point.dot = f(other) {
	x * other.x + y * other.y;
};

p = Point(3, 5);
p2 = Point(2, 7);
bound = p.dot;

return bound(p2);
`

	if !CompileCheckExit(src, 41) {
		t.Fail()
	}
}

func TestMultipleReturn(t *testing.T) {
	src := `
fun = f() {
	num = 65;
	return num;
	num = 32;
	return 21;
};

return fun();
`

	if !CompileCheckExit(src, 65) {
		t.Fail()
	}
}

func TestNestedStruct(t *testing.T) {
	src := `
struct Point {
	int x;
	int y;
};

struct Triangle {
	Point v1;
	Point v2;
	Point v3;
};

tri = Triangle(Point(1, 2), Point(3, 4), Point(5, 6));

return tri.v3.x;
`

	if !CompileCheckExit(src, 5) {
		t.Fail()
	}
}

func TestSimpleCoro(t *testing.T) {
	src := `
gen = f() {
	yield 5;
	yield 2;
};
g = gen();
return next(g) + next(g);
`

	if !CompileCheckExit(src, 7) {
		t.Fail()
	}
}

func TestCoroLoop(t *testing.T) {
	src := `
gen = f() {
	x = 1;
	while true {
		yield x;
		x = x + 1;
	};
};

g = gen();
sum = 0;
sum = sum + next(g);
sum = sum + next(g);
sum = sum + next(g);
sum = sum + next(g);
sum = sum + next(g);
sum = sum + next(g);

return sum;
`

	if !CompileCheckExit(src, 21) {
		t.Fail()
	}
}

func TestCoroutine(t *testing.T) {
	src := `
struct Box {
	int num;
};

b = Box(5);

fun = f() {
	yield 1;
	b.num = 3;
	yield 3;
};

final = 0;
co = fun();
final = final + b.num;
next(co);
final = final + b.num;
next(co);
final = final + b.num;
return final;
`

	if !CompileCheckExit(src, 13) {
		t.Fail()
	}
}

func TestCloArg(t *testing.T) {
	src := `
arr = [5, 7];
arr_len = 2;

clo = f(a) {
	a[arr_len - 1] = 11;
	return 0;
};

clo(arr);
return arr[arr_len - 1];
`

	if !CompileCheckExit(src, 11) {
		t.Fail()
	}
}

func TestBubbleSort(t *testing.T) {
	sortArr := []int{32, 12, 65, 2, 11}
	strArr := make([]string, len(sortArr))
	for i, elem := range sortArr {
		strArr[i] = fmt.Sprintf("%d", elem)
	}
	formattedArr := fmt.Sprintf("[%s]", strings.Join(strArr, ", "))

	src := `
arr = %s;
leng = %d;
i = 0;
while i < leng - 1 {
	j = 0;
	while j < leng - i - 1 {
		if arr[j] > arr[j + 1] {
			tmp = arr[j];
			arr[j] = arr[j + 1];
			arr[j + 1] = tmp;
		};
		j = j + 1;
	};
	i = i + 1;
};

return arr[%d];
`

	sort.Ints(sortArr)
	for i, elem := range sortArr {
		newSrc := fmt.Sprintf(src, formattedArr, len(sortArr), i)
		if !CompileCheckExit(newSrc, elem) {
			t.Fail()
		}
	}
}

func TestRecursiveStruct(t *testing.T) {
	src := `
struct Node {
	int val;
	Node left;
	Node right;
};

n = Node(5, Node(7, null, null), null);

return n.left.val;
`

	if !CompileCheckExit(src, 7) {
		t.Fail()
	}
}

func TestMaxLinkedList(t *testing.T) {
	src := `
struct Node {
	int val;
	Node nex;
};

list = Node(5, Node(10, Node(1, Node(45, Node(9, null)))));

max = f(l) {
	curr = l;
	max_val = 0;
	while curr != null {
		if curr.val > max_val {
			max_val = curr.val;
		};
		curr = curr.nex;
	};

	max_val;
};

return max(list);
`

	if !CompileCheckExit(src, 45) {
		t.Fail()
	}
}

func TestTypeHint(t *testing.T) {
	src := `
f(int, byte)[]string splitter = null;
splitter = f(n, n2) {
	return ["hello", "world"];
};
`

	if !CompileCheckExit(src, 0) {
		t.Fail()
	}
}

func TestExtern(t *testing.T) {
	src := `
f(int)void __extern_print(5);
return 0;
`

	if !CompileCheckExit(src, 0) {
		t.Fail()
	}
}

func TestNestedCoro(t *testing.T) {
	src := `
getco = f(a, b) {
	gen2 = f() {
		while true {
			yield a + b;
		};
	};

	return gen2;
};

gen = getco(3, 4);
co1 = gen();
sum = next(co1) + next(co1);
return sum;
`

	if !CompileCheckExit(src, 14) {
		t.Fail()
	}
}

func TestStructWithFunc(t *testing.T) {
	src := `
struct Point {
	int x;
	int y;
	f()int thing;
};

p = Point(4, 5, f(){5;});
val = p.thing();
return val;
`

	if !CompileCheckExit(src, 5) {
		t.Fail()
	}
}

func TestAssignTup(t *testing.T) {
	src := `
t = (1, 2, 3);
t[1] = 1;
return t[1];
`

	if !CompileCheckExit(src, 1) {
		t.Fail()
	}
}

func TestRecursion(t *testing.T) {
	src := `
fun = f(num) {
	if num == 0 {
		return 0;
	};
	return num + fun(num - 1);
};

return fun(10);
`

	if !CompileCheckExit(src, 55) {
		t.Fail()
	}
}

func TestNakedBlock(t *testing.T) {
	src := `
i = 5;
{
	b = 7;
	return b;
};

return i;
`

	if !CompileCheckExit(src, 7) {
		t.Fail()
	}
}

func TestFor(t *testing.T) {
	src := `
sum = 0;
for i = 0; i < 5; i = i + 1 {
	sum = sum + i;
};

return sum;
`

	if !CompileCheckExit(src, 10) {
		t.Fail()
	}
}

func TestWhileBreak(t *testing.T) {
	src := `
c = 0;
while true {
	if c >= 5 {
		break;
	};
	c = c + 1;
};

return 5;
`

	if !CompileCheckExit(src, 5) {
		t.Fail()
	}
}

func TestWhileContinue(t *testing.T) {
	src := `
c = 0;
while c < 5 {
	c = c + 1;
	continue;
	c = c + 30;
};

return c;
`

	if !CompileCheckExit(src, 5) {
		t.Fail()
	}
}

func TestForContinue(t *testing.T) {
	src := `
sum = 0;
for i = 0; i < 10; i = i + 1 {
	continue;
	sum = sum + i;
};

return sum;
`

	if !CompileCheckExit(src, 0) {
		t.Fail()
	}
}

func TestPostReturn(t *testing.T) {
	src := `
fun = f() {
	c = 5;
	return c;
	c + 1;
};

return fun();
`
	if !CompileCheckExit(src, 5) {
		t.Fail()
	}
}

func TestArrayLen(t *testing.T) {
	src := `
arr = [1, 2, 3, 2];
return len(arr);
`

	if !CompileCheckExit(src, 4) {
		t.Fail()
	}
}

func TestTupleLen(t *testing.T) {
	src := `
tup = (1, "hello", 2, "world");
return len(tup);
`

	if !CompileCheckExit(src, 4) {
		t.Fail()
	}
}

func TestStringLen(t *testing.T) {
	src := `
str = "hello, world!";
return len(str);
`

	if !CompileCheckExit(src, 13) {
		t.Fail()
	}
}

func TestDone(t *testing.T) {
	src := `
iter = f() {
	for i = 0; i < 10; i = i + 1 {
		yield i;
	};
};

count = iter();
sum = 0;

while true {
	n = next(count);
	if done(count) == true {
		break;
	};
	sum = sum + n;
};

return sum;
`

	if !CompileCheckExit(src, 45) {
		t.Fail()
	}
}

func TestForGen(t *testing.T) {
	src := `
iter = f() {
	for i = 0; i < 10; i = i + 1 {
		yield i;
	};
};

sum = 0;
for item in iter() {
	sum = sum + item;
};
return sum;
`
	if !CompileCheckExit(src, 45) {
		t.Fail()
	}
}

func TestTypeBuiltin(t *testing.T) {
	src := `
x = 5;
y = 23;

if type(x) == type(y) {
	return 5;
};
return 3;
`

	if !CompileCheckExit(src, 5) {
		t.Fail()
	}
}

func TestAnyType(t *testing.T) {
	src := `
y = any(5);

struct Box {
	int num;
};

b = Box(7);

anybox = any(b);
box2 = anybox.(Box);

return box2.num + y.(int);
`

	if !CompileCheckExit(src, 12) {
		t.Fail()
	}
}

func TestIsExp(t *testing.T) {
	src := `
sum = 0;

x = 5;
if x is int {
	sum = sum + 1;
};

y = any(5);
if x is string {
	sum = sum + 2;
};

z = "string";
if z is int {
	sum = sum + 4;
};

foo = any(8);
if foo is int {
	sum = sum + foo.(int);
};

return sum;
`

	if !CompileCheckExit(src, 9) {
		t.Fail()
	}
}

func TestForIter(t *testing.T) {
	src := `
a = [1, 2, 4, 8];

sum = 0;
for item in a {
	sum = sum + item;
};

range = f() {
	for i = 1; i < 10; i = i * 2 {
		yield i;
	};
};

sum2 = 0;
for x in range() {
	sum2 = sum2 + x;
};

return sum + sum2;
`

	if !CompileCheckExit(src, 30) {
		t.Fail()
	}
}

//func TestManyClosures(t *testing.T) {
//	src := `
//x = 16;
//
//fun = f(arg) {
//	f(arg2) {
//		f(arg3) {
//			f(arg4) {
//				return x + arg + arg2 + arg3 + arg4;
//			};
//		};
//	};
//};
//
//fun(1)(2)(4)(8);
//`
//
//	if !CompileCheckExit(src, 1+2+4+8+16) {
//		t.Fail()
//	}
//}

//func TestMutableNumClosure(t *testing.T) {
//	src := `
//x = 22;
//fun = f() {
//	x + 1;
//	x = 33;
//};
//
//fun();
//return x;
//`
//
//	if !CompileCheckExit(src, 33) {
//		t.Fail()
//	}
//}

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
