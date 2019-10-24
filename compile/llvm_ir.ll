define i64 @main() {
; <label>:0
	%1 = alloca i64 (i64, i64)*
	store i64 (i64, i64)* @fun_0, i64 (i64, i64)** %1
	%2 = alloca i64
	%3 = load i64 (i64, i64)*, i64 (i64, i64)** %1
	%4 = call i64 %3(i64 21, i64 32)
	store i64 %4, i64* %2
	ret i64 0
}

define i64 @fun_0(i64 %a_1, i64 %b_1) {
; <label>:0
	%1 = alloca i64
	store i64 %a_1, i64* %1
	%2 = alloca i64
	store i64 %b_1, i64* %2
	%3 = load i64 (i64, i64)*, i64 (i64, i64)** %1
	%4 = call i64 %3(i64 3, i64 4)
	ret i64 %4
}
