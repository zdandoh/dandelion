#include <sys/mman.h>
#include <stdio.h>
#include <string.h>

int main()
{
    printf("%d\n", MAP_ANONYMOUS|MAP_PRIVATE);
    printf("%d\n", PROT_EXEC|PROT_READ|PROT_WRITE);
    void *ptr = mmap(NULL, 72, PROT_EXEC|PROT_READ|PROT_WRITE, MAP_ANONYMOUS|MAP_PRIVATE, 0, 0);
    printf("%p - %p\n", ptr, (void*)-1);
    memset(ptr, 0, 72);
}