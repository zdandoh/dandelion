#include <windows.h>
#include <stdio.h>

int main() {
    void* adr = VirtualAlloc(NULL, 72, MEM_COMMIT | MEM_RESERVE, PAGE_READWRITE);
    printf("%p\n", adr);
}