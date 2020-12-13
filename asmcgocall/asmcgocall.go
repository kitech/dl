package asmcgocall

import (
	"unsafe"
)

//go:linkname asmcgocall runtime.asmcgocall
//go:noescape
func asmcgocall(fn unsafe.Pointer, arg unsafe.Pointer) int32

// 这个最快，直接调用好了,而且是安全的，似乎可以用
// cgo直接调用大概60ns，这种方式调用在12ns(与system stack切换时间)，go的调用2ns
// 只要 runtime.asmcgocall 还存在就可以用，与参数打包方式无关
// args &struct {p0 Type0, p1 Type1, ret Type}
func Asmcc(cfn unsafe.Pointer, args unsafe.Pointer) {
	// pargs := args
	asmcgocall(cfn, args)
}
