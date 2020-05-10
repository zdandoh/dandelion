#ifndef ALLOC
#define ALLOC

#include <stdlib.h>
#include <fcntl.h>
#include <stdint.h>

#define MEM_SIZE 72

#ifdef _WIN32
#include <stdio.h>
#include <windows.h>

__declspec(dllexport)
void* __cdecl alloc_clo() {
    void* adr = VirtualAlloc(NULL, MEM_SIZE, MEM_COMMIT | MEM_RESERVE, PAGE_READWRITE);
    return adr;
}

#else
#include <stdio.h>
#include <sys/mman.h>

void* alloc_clo() {
    void *ptr = mmap(NULL, MEM_SIZE, PROT_EXEC|PROT_READ|PROT_WRITE, MAP_ANONYMOUS|MAP_PRIVATE, 0, 0);
    return ptr;
}

#endif
#endif

void print(int d) {
	printf("%d\n", d);
}

void printb(char b) {
	if(b == 1) {
		printf("true\n");
	}
	else if(b == 0) {
		printf("false\n");
	} else {
		printf("BADVAL\n");
	}
}

void printp(void* p) {
	printf("%p\n", p);
}

typedef struct str {
	uint64_t len;
	char* data;
} str;

void prints(str* s) {
	printf("%.*s\n", s->len, s->data);
};

int d_open(void* fname, void* mode) {
	open(fname, O_CREAT);
	printf("%p %p\n", fname, mode);
	return 5;
}
