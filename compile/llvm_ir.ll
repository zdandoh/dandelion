define i32 @main() {
; <label>:0
	%1 = alloca i64
	store i64 6, i64* %1
	%2 = alloca i64
	store i64 7, i64* %2
	%3 = alloca i64
	%4 = load i64, i64* %1
	%5 = load i64, i64* %2
	%6 = add i64 %4, %5
	%7 = add i64 %6, 78
	store i64 %7, i64* %3
	ret i32 0
}
