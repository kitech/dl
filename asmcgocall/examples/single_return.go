package main

/*

int ret_value = 666;

int single_ret(struct {int ret;} *ax) {

	ax->ret = ret_value;

	// error code
	return 0;
}

int regular_single_ret() {
	return ret_value;
}

*/
import "C"

import (
	"unsafe"

	"github.com/kitech/dl/asmcgocall"
)

var singleRetExpected = C.ret_value

var singleRetAsmcgocall = func() C.int {
	argv := struct{ ret C.int }{}
	asmcgocall.Asmcc(C.single_ret, unsafe.Pointer(&argv))
	return argv.ret
}

func singleRetCgocall() C.int {
	return C.regular_single_ret()
}
