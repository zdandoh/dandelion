#include <stdio.h>

int func(int a, int b) {
    return a + b;
}

int main() {
    int x = func(4, 7);
    int y = func(21, 11);

    printf("%d %d", x, y);
    return 0;
}