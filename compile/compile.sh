rm -f out.ll # silently remove any files left behind by a failed run
opt < llvm_ir.ll -O3 -enable-coroutines -coro-early -coro-split -coro-elide -coro-cleanup -S > out.ll &&
llc -filetype=obj out.ll
clang ../lib/headers.o ../lib/gc.a ../lib/alloc.o ../lib/exception.o out.o