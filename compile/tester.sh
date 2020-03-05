#!/usr/bin/env bash

opt < llvm_ir.ll -O0 -enable-coroutines -S > out.ll &&
llc -load ../lib/lib.so out.ll &&
clang ../lib/lib.so out.s &&
>&2 ./a.out
echo $?
rm out.s out.ll
