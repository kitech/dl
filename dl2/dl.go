package dl

// Still use some cgo type/const/variable
// But aims avoid original cgo call to speed up call performance

/*
#cgo LDFLAGS: -ldl

#include <stdlib.h>
#include <stdio.h>
#include <stdint.h>
#include <dlfcn.h>

// 由于pad有问题，所有的函数都打包成一个void*结构体参数，返回值也打包在这个结构体中
int goasmcc_dlopen(struct {char *a0; int a1; void* ret;} *ax) {
   ax->ret = dlopen(ax->a0, ax->a1);
   return 0; // just error code
}
int goasmcc_dlclose(struct {void* a0; int ret; } *ax) {
    ax->ret = dlclose(ax->a0);
  return 0;
}
int goasmcc_dlsym(struct {void* a0; char* a1; void* ret; } *ax) {
  ax->ret = dlsym(ax->a0, ax->a1);
  return 0;
}
int goasmcc_dlerror(struct {char* ret;} *ax) {
  ax->ret = dlerror();
  return 0;
}

int goasmcc_empty() {
  char buf[100];
  buf[0] = 12;
  //sprintf(buf, "ttt %d", 999);
  return 0;
}
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/kitech/dl/asmcgocall"
)

//
type Flags int

const (
	Lazy   Flags = C.RTLD_LAZY
	Now    Flags = C.RTLD_NOW
	Global Flags = C.RTLD_GLOBAL
	Local  Flags = C.RTLD_LOCAL
	//NoLoad   Flags = C.RTLD_NOLOAD
	//NoDelete Flags = C.RTLD_NODELETE
	// First Flags = C.RTLD_FIRST
)

// return func(unsafe.Pointer)
var openFunc = asmcgocall.Register2(C.goasmcc_dlopen)
var closeFunc = asmcgocall.Register2(C.goasmcc_dlclose)
var symFunc = asmcgocall.Register2(C.goasmcc_dlsym)
var errorFunc = asmcgocall.Register2(C.goasmcc_dlerror)
var EmptyAsmcc = asmcgocall.Register2(C.goasmcc_empty)

func EmptyCgocc() { C.goasmcc_empty() }

type Handle struct {
	c unsafe.Pointer
}

func Open(fname string, flags Flags) (Handle, error) {
	c_str := (*C.char)(unsafe.Pointer(&[]byte(fname)[0]))

	var argv = struct {
		p0  *C.char
		p1  C.int
		ret unsafe.Pointer
	}{c_str, C.int(flags), nil}
	openFunc((unsafe.Pointer(&argv)))
	h := argv.ret
	if h == nil {
		err := fmt.Errorf("dl: %s", DLError())
		return Handle{}, err
	}
	h2 := unsafe.Pointer(uintptr(h))
	return Handle{h2}, nil
}

func (h Handle) Close() error {
	if h.c == nil {
		return nil
	}
	var argv = struct {
		a0  unsafe.Pointer
		ret C.int
	}{h.c, 0}
	closeFunc(unsafe.Pointer(&argv))
	o := argv.ret
	if o != C.int(0) {
		err := fmt.Errorf("dl: %s", DLError())
		return err
	}
	h.c = nil
	return nil
}

func (h Handle) Addr() uintptr {
	return uintptr(h.c)
}

func (h Handle) Symbol(symbol string) (uintptr, error) {
	c_sym := (*C.char)(unsafe.Pointer(&[]byte(symbol)[0]))

	var argv = struct {
		a0  unsafe.Pointer
		a1  *C.char
		ret unsafe.Pointer
	}{h.c, c_sym, nil}

	symFunc(unsafe.Pointer(&argv))
	c_addr := argv.ret
	if c_addr == nil {
		err := fmt.Errorf("dl: %s", DLError())
		return 0, err
	}
	return uintptr(c_addr), nil
}

func (h Handle) DLError() string { return DLError() }
func DLError() string {
	var argv = struct{ ret *C.char }{nil}
	errorFunc(unsafe.Pointer(&argv))
	c_err := argv.ret
	if c_err == nil {
		return ""
	}
	return C.GoString(c_err)
}

// /* Portable libltdl versions of the system dlopen() API. */
// LT_SCOPE lt_dlhandle lt_dlopen          (const char *filename);
// LT_SCOPE lt_dlhandle lt_dlopenext       (const char *filename);
// LT_SCOPE lt_dlhandle lt_dlopenadvise    (const char *filename,
//                                          lt_dladvise advise);
// LT_SCOPE void *     lt_dlsym            (lt_dlhandle handle, const char *name);
// LT_SCOPE const char *lt_dlerror         (void);
// LT_SCOPE int        lt_dlclose          (lt_dlhandle handle);

// EOF
