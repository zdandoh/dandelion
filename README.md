Dandelion
---
[![Build status](https://ci.appveyor.com/api/projects/status/gw92mo0nl48cy3mm?svg=true)](https://ci.appveyor.com/project/zdandoh/dandelion)
[![codecov](https://codecov.io/gh/zdandoh/dandelion/branch/master/graph/badge.svg)](https://codecov.io/gh/zdandoh/dandelion)

Dandelion is a small language that intends to be useful for quickly writing terse utility/text processing programs. The goal of the project is to serve as a replacement for arcane bash incantations and some Python scripts.

Dandelion is statically typed, but uses global type inference, so type annotations generally don't need to be provided. The compiler leverages LLVM to provide native binaries and an interpeted mode (with JIT). Dandelion was designed to provide the programmer the ability to write terse programs that are comparable in performance to C, while exceeding the expressiveness of Python. Below is an example of a simple `grep` program written in Dandelion:
```
collect("*.txt") -> lines ->> fi{ "substr" in e } -> p
```
(This utilizes one of the headline features of Dandelion - pipelines. Imagine classic shell pipelines but connecting functions/coroutines instead of independent executables. I will write more about this later.)

I am currently working on implementing core language features (rough checklist below). To get an idea of where the language is at, you can check out the tests in [compile_test.go](compile/compile_test.go)

Language Features
---

- [x] Primitive data types: `bool`, `int`, `float`, `string`, `byte`
- [x] Tuples
- [x] Lists
- [x] Structs
- [x] Functions
- [x] Struct methods
- [ ] Pipelines
- [x] Global type inference
- [x] Type annotations (for when types can't be inferred)
- [x] Closures
- [x] Implicit return values
- [x] Classic imperative control flow (`if`, `while`, `for`, `break`, `continue`, `return`)
- [ ] Function modifiers
- [x] Coroutines
- [x] GC
- [ ] Command invocation syntactic sugar
- [ ] String interpolation
- [ ] Automatic semi-colon insertion
- [ ] Cross platform (Windows, Mac, & Linux)

Planned Features
---
- [ ] Regex support
- [ ] Escape analysis
- [ ] Hash tables
- [ ] Basic standard library
- [ ] JSON construction/parsing
- [ ] Python interface system
