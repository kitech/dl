package main

/*

int empty() {
	// error code
	return 0;
}

void regular_empty() {}

*/
import "C"

import "github.com/kitech/dl/asmcgocall"

var emptyAsmcgocall = func() { asmcgocall.Asmcc(C.empty, nil) }

func emptyCgocall() {
	C.regular_empty()
}
