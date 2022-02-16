#!/usr/bin/env bash

libext=so
if [[ $(uname) == "Darwin" ]]
then
	libext=dylib
fi

rm -f out.ll # silently remove any files left behind by a failed run
opt < llvm_ir.ll -O1 -enable-coroutines -coro-early -coro-split -coro-elide -coro-cleanup -S > out.ll &&
lli -load ../lib/lib.$libext -load ../lib/libgc.$libext out.ll
echo $?
rm out.ll
