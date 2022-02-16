# On MacOS:
# brew install libgc

UNAME := $(shell uname)

build: runtime
	java -jar antlr.jar -Dlanguage=Go -o aparser Dandelion.g4 DandelionLex.g4
	go build

runtime: lib/alloc.c lib/exception.c
ifeq ($(UNAME), Linux)
	clang -shared -Wall -fPIC -o lib/lib.so lib/alloc.c lib/exception.c
	clang -Wall -o lib/linux/alloc.o -c lib/alloc.c
	clang -Wall -o lib/linux/exception.o -c lib/exception.c
endif

ifeq ($(UNAME), Darwin)
	mkdir -p lib/darwin
	clang -shared -Wall -fPIC -o lib/lib.dylib lib/alloc.c lib/exception.c
	clang -Wall -o lib/darwin/alloc.o -c lib/alloc.c
	clang -Wall -o lib/darwin/exception.o -c lib/exception.c
endif

ifeq ($(UNAME), windows32)
	gcc -c -v -m32 -DBUILD_SHARED lib/alloc.c lib/exception.c -o lib/alloc.o
	gcc -shared -m32 -v -Wl,--out-implib,libtstdll.a lib/alloc.o -o lib/lib.dll
endif
