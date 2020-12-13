package asmcgocall

import (
	"reflect"
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

type parameter struct {
	packedOffset, offset, size uintptr
}

// 固定签名的实现方式
func Register2(cfn unsafe.Pointer) (result func(unsafe.Pointer)) {
	Register(cfn, &result)
	return
}

func Register(cfn unsafe.Pointer, caller interface{}) {
	fn := reflect.TypeOf(caller)

	if fn.Kind() != reflect.Ptr {
		panic("caller must be pointer to variable with function type")
	}

	fn = fn.Elem()

	if fn.Kind() != reflect.Func {
		panic("caller must be pointer to variable with function type")
	}

	parameters, returns := calcPadding(fn)

	var proxy func(args [0]byte) (ret [0]byte)

	if returns {
		proxy = createProxyRet(cfn, parameters)
	} else {
		proxy = createProxy(cfn, parameters)
	}

	/*
		let proxy(function above) be callable with passed function signature of call
	*/
	eface := (*[2]unsafe.Pointer)(unsafe.Pointer(&caller))[1]
	*(*unsafe.Pointer)(eface) = *(*unsafe.Pointer)(unsafe.Pointer(&proxy))
}

func calcPadding(fn reflect.Type) (parameters []parameter, returns bool) {
	if Debug {
		logf("struct argument_frame {\n")
	}

	var padoffset uintptr
	parameters = make([]parameter, 0, 1)

	off := int(0)
	for i, n := 0, fn.NumIn(); i < n; i++ {
		param := fn.In(i)
		psize := param.Size()

		palign := param.Align()
		if off%palign != 0 {
			pad := palign - off%palign

			if Debug {
				logf("\tchar __pad%d[%d]; // %d %s\n", off, pad, pad, bytes(pad))
			}

			off += pad
			padoffset += uintptr(pad)
		}

		if Debug {
			logf("\t%s p%d; // %d %s\n", param.Name(), i, psize, bytes(psize))
		}

		parameters = append(parameters, parameter{uintptr(off) - padoffset, uintptr(off), psize})

		off += int(psize)
	}

	const ptrsize = int(unsafe.Sizeof(uintptr(0)))

	if off%ptrsize != 0 {
		pad := ptrsize - off%ptrsize

		if Debug {
			logf("\tchar __pad%d[%d]; // %d %s\n", off, pad, pad, bytes(pad))
		}

		off += pad
		padoffset += uintptr(pad)
	}

	returns = false
	if fn.NumOut() > 0 {
		t := fn.Out(0)
		retalign := t.Align()
		retsize := t.Size()

		if off%retalign != 0 {
			pad := retalign - off%retalign

			if Debug {
				logf("\tchar __pad%d[%d]; // %d %s\n", off, pad, pad, bytes(pad))
			}

			off += pad
			padoffset += uintptr(pad)
		}

		if Debug {
			logf("\t%s ret;  // %d %s\n", t.Name(), retsize, bytes(retsize))
		}

		parameters = append(parameters, parameter{uintptr(off) - padoffset, uintptr(off), retsize})

		off += int(retsize)

		returns = true
	}

	if Debug && off%ptrsize != 0 {
		pad := ptrsize - off%ptrsize
		logf("\tchar __pad%d[%d]; // %d %s\n", off, pad, pad, bytes(pad))
		off += pad
	}

	if Debug && off == 0 {
		logf("\tchar unused;\n") // avoid empty struct
	}

	if Debug {
		logf("}; // %d bytes\n", off)
	}

	return
}

func createProxy(cfn unsafe.Pointer, parameters []parameter) func(args [0]byte) (ret [0]byte) {
	if len(parameters) == 0 {
		return func(args [0]byte) (ret [0]byte) {
			asmcgocall(cfn, nil)

			return
		}
	}

	return func(args [0]byte) (ret [0]byte) {
		//params := parameters

		pargs := unsafe.Pointer(&args)
		pargs = unsafe.Pointer(*(**byte)(pargs)) // 仅有一个指针参数的限定情况的转换//

		/*
			var (
				p parameter
				i uintptr
			)
		*/

		// 不需要 Pack了
		/*
			Pack parameters
		*/
		/*
			for c := 1; c < len(params); c++ {
				p = params[c]

				if p.offset == p.packedOffset {
					continue
				}

				for i = 0; i < p.size; i++ {
					*(*byte)(unsafe.Pointer(uintptr(pargs) + p.packedOffset + i)) = *(*byte)(unsafe.Pointer(uintptr(pargs) + p.offset + i))
				}
			}
		*/

		asmcgocall(cfn, pargs)

		return
	}
}

func createProxyRet(cfn unsafe.Pointer, parameters []parameter) func(args [0]byte) (ret [0]byte) {
	if len(parameters) == 1 {
		return func(args [0]byte) (ret [0]byte) {
			asmcgocall(cfn, unsafe.Pointer(&args))

			return
		}
	}

	return func(args [0]byte) (ret [0]byte) {
		params := parameters

		pargs := unsafe.Pointer(&args)

		var (
			p parameter
			i uintptr
		)

		/*
			Pack parameters
		*/
		n := len(params) - 1
		for c := 1; c < n; c++ {
			p = params[c]

			if p.offset == p.packedOffset {
				continue
			}

			for i = 0; i < p.size; i++ {
				*(*byte)(unsafe.Pointer(uintptr(pargs) + p.packedOffset + i)) = *(*byte)(unsafe.Pointer(uintptr(pargs) + p.offset + i))
			}
		}

		p = params[n]
		if p.offset == p.packedOffset {
			asmcgocall(cfn, pargs)

			return
		}

		/*
			Fill return value with nulls (taking pad's spot)
		*/
		for i = 0; i < p.size; i++ {
			*(*byte)(unsafe.Pointer(uintptr(pargs) + p.packedOffset + i)) = 0
		}

		asmcgocall(cfn, pargs)

		/*
			Reposition(unpack) return value back before exit
		*/
		for i = 0; i < p.size; i++ {
			*(*byte)(unsafe.Pointer(uintptr(pargs) + p.offset + i)) = *(*byte)(unsafe.Pointer(uintptr(pargs) + p.packedOffset + i))
		}

		return
	}
}
