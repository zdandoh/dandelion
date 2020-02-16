#!/usr/bin/env bash

llc llvm_ir.ll && clang llvm_ir.s && valgrind ./a.out -all; rm ./llvm_ir.s
