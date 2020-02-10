#!/usr/bin/env bash

lli -load ../lib/lib.so llvm_ir.ll
echo $?
