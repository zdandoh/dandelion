#ifndef ALLOC
#define ALLOC

#include <stdlib.h>
#include <fcntl.h>
#include <stdint.h>
#include <alloca.h>
#include <string.h>
#include <unistd.h>

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

int zero = 0;

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

typedef struct arr {
	uint32_t len;
	uint32_t cap;
	char* data;
} arr;

void prints(str* s) {
	printf("%.*s\n", (int)s->len, s->data);
}

int d_open(str* fname) {
	char* term_name = alloca(fname->len + 1);
	memcpy(term_name, fname->data, fname->len);
	term_name[fname->len] = 0;
	return open(term_name, O_RDONLY);
}

int d_read(int fd, arr* buf) {
	return read(fd, buf->data, buf->len);
}