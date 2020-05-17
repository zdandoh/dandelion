#include <stdlib.h>
#include <stdio.h>

#define EX_INVALID_CAST_NO 1
const char* EX_INVALID_CAST = "Fatal error: invalid assertion from any type";
const char* EX_INDEX_OOB = "Fatal error: index %d out of bounds\n";

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

void indexoob(int index) {
	printf(EX_INDEX_OOB, index);
	exit(2);
}