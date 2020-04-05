Dandelion
---
[![Build status](https://ci.appveyor.com/api/projects/status/gw92mo0nl48cy3mm?svg=true)](https://ci.appveyor.com/project/zdandoh/dandelion)

Dandelion is a small language that intends to be useful for quickly writing terse utility/text processing programs. The goal of the project is to serve as a replacement for arcane bash incantations and shorter Python scripts.

Core Features
---
Below are some core language features that I want to implement, roughly in order of importance:

- Primitive data types: `bool`, `int`, `float`, `string`, `byte`
- Tuples
- Lists
- Structs
- Functions
- Struct methods
- Pipelines
- Type inference
- Type annotations for when type inference fails
- Closures
- Implicit return values
- Classic imperative control flow (`if`, `while`, `for`, `break`, `continue`, `return`)
- Function modifiers
- Coroutines
- GC
- Command invocation syntactic sugar
- String interpolation
- Automatic semi-colon insertion
- Cross platform (Windows, Mac, & Linux)
- No slower than 5x C (hopefully much better than that)

Nice to Have Features
---
- Basic escape analysis
- Hash tables
- Basic standard library
- JSON construction/parsing
- Python interface system
