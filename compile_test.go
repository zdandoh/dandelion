package main

import (
	"fmt"
	"testing"
	"time"
)

func TestCompileFunction(t *testing.T) {
	tokens := tokenizeCode("45 + 36")
	main := compileTokens(tokens)
	main.Run()
	fmt.Println(main.registers)
}

func TestIterationSpeed(t *testing.T) {
	tm := time.Now()

	iter := 100000000
	b := 0
	for i := 0; i < iter; i++ {
		b = 5848 + 382474
	}

	fmt.Println(b)
	fmt.Println(float64(time.Since(tm).Nanoseconds()) / float64(iter))
}
