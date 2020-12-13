package dl

// Still use some cgo type/const/variable
// But aims avoid original cgo call to speed up call performance

/*
#cgo LDFLAGS: -ldl

#include <stdlib.h>
#include <stdio.h>
#include <dlfcn.h>

// 由于pad有问题，所有的函数都打包成一个void*结构体参数，一个返回值指针
int goasmcc_dlopen(struct {void *p0; void* r;} *ax) {
  struct {char* a0; int a1;}* argv = ax->p0;
  ax->r = dlopen(argv->a0, argv->a1);
  return 0; // just error code
}
int goasmcc_dlclose(struct {void* p0; int r; } *ax) {
  struct {void* a0;}* argv = ax->p0;
  ax->r = dlclose(argv->a0);
  return 0;
}
int goasmcc_dlsym(struct {void* p0; void* r; } *ax) {
  struct {void* a0; char* a1;}* argv = ax->p0;
  ax->r = dlsym(argv->a0, argv->a1);
  return 0;
}
int goasmcc_dlerror(char** sret) {
  *sret = dlerror();
  return 0;
}

int goasmcc_empty() { return 0; }
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/LaevusDexter/asmcgocall"
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

var openFunc = func() (result func(unsafe.Pointer) unsafe.Pointer) {
	asmcgocall.Register(C.goasmcc_dlopen, &result)
	return
}()
var closeFunc = func() (result func(unsafe.Pointer) C.int) {
	asmcgocall.Register(C.goasmcc_dlclose, &result)
	return
}()
var symFunc = func() (result func(unsafe.Pointer) unsafe.Pointer) {
	asmcgocall.Register(C.goasmcc_dlsym, &result)
	return
}()
var errorFunc = func() (result func() *C.char) {
	asmcgocall.Register(C.goasmcc_dlerror, &result)
	return
}()
var EmptyFunc = func() (result func()) {
	asmcgocall.Register(C.goasmcc_empty, &result)
	return
}()

func EmptyFunc2() { C.goasmcc_empty() }

type Handle struct {
	c unsafe.Pointer
}

func Open(fname string, flags Flags) (Handle, error) {
	c_str := (*C.char)(unsafe.Pointer(&[]byte(fname)[0]))

	var argv = struct {
		p0 *C.char
		p1 C.int
	}{c_str, C.int(flags)}
	h := openFunc(unsafe.Pointer(&argv))
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
	var argv = struct{ a0 unsafe.Pointer }{h.c}
	o := closeFunc(unsafe.Pointer(&argv))
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
		a0 unsafe.Pointer
		a1 *C.char
	}{h.c, c_sym}
	c_addr := symFunc(unsafe.Pointer(&argv))
	if c_addr == nil {
		err := fmt.Errorf("dl: %s", DLError())
		return 0, err
	}
	return uintptr(c_addr), nil
}

func (h Handle) DLError() string { return DLError() }
func DLError() string {
	c_err := errorFunc()
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
