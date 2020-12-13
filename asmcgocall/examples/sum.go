package main

/*
int sum(struct {
		int p0;
		int p1;
		int p2;
		int p3;
		int p4;
		int r;
	} *a) {

	//printf("p1 = %d, p2 = %d, p3 = %d, p4 = %d, p5 = %d, r = %d\n", a->p0, a->p1, a->p2, a->p3, a->p4, a->r);

	a->r = a->p0 + a->p1 + a->p2 + a->p3 + a->p4;

	// error code
	return 0;
}

int regular_sum(int p0, int p1, int p2, int p3, int p4) {
	return p0 + p1 + p2 + p3 + p4;
}


*/
import "C"

import (
	"unsafe"

	"github.com/kitech/dl/asmcgocall"
)

var sumAsmcgocall = func(a0, a1, a2, a3, a4 int) int {
	argv := struct {
		a0, a1, a2, a3, a4 C.int
		r                  C.int
	}{C.int(a0), C.int(a1), C.int(a2), C.int(a3), C.int(a4), 0}
	asmcgocall.Asmcc(C.sum, unsafe.Pointer(&argv))
	return int(argv.r)
}

func sumCgocall(p0, p1, p2, p3, p4 C.int) int {
	return int(C.regular_sum(p0, p1, p2, p3, p4))
}
