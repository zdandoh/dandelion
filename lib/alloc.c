#ifndef ALLOC
#define ALLOC

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