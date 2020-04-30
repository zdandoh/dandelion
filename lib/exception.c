
#include <stdlib.h>
#include <stdio.h>

#define EX_INVALID_CAST_NO 1
const char* EX_INVALID_CAST = "Fatal error: invalid assertion from any type";

void throwex(int exno) {
	const char* ex_text = NULL;

	switch(exno) {
		case EX_INVALID_CAST_NO:
			ex_text = EX_INVALID_CAST;
			break;
	}

	printf("%s\n", ex_text);
	exit(2);
}