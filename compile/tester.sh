#!/usr/bin/env bash

rm -f out.ll # silently remove any files left behind by a failed run
opt < llvm_ir.ll -O0 -enable-coroutines -coro-early -coro-split -coro-elide -coro-cleanup -S > out.ll &&
lli -load ../lib/lib.so out.ll
echo $?
rm out.ll
