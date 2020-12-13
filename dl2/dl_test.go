package dl_test

import (
	"testing"

	dl "github.com/kitech/dl/dl2"
)

func TestDlOpenLibc(t *testing.T) {
	lib, err := dl.Open(libc_name, dl.Now)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = lib.Close()
	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestDlSymLibc(t *testing.T) {
	lib, err := dl.Open(libc_name, dl.Now)

	if err != nil {
		t.Errorf("%v", err)
	}

	_, err = lib.Symbol("puts")
	if err != nil {
		t.Errorf("%v", err)
	}

	err = lib.Close()
	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestDlOpenLibm(t *testing.T) {
	lib, err := dl.Open(libm_name, dl.Now)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = lib.Close()
	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestDlSymLibm(t *testing.T) {
	lib, err := dl.Open(libm_name, dl.Now)

	if err != nil {
		t.Errorf("%v", err)
	}

	_, err = lib.Symbol("fabs")
	if err != nil {
		t.Errorf("%v", err)
	}

	err = lib.Close()
	if err != nil {
		t.Errorf("%v", err)
	}
}

func BenchmarkDlSymLibm(b *testing.B) {
	lib, err := dl.Open(libm_name, dl.Now)

	for i := 0; i < b.N; i++ {
		_, err = lib.Symbol("fabs")
	}

	err = lib.Close()
	_ = err
}
func BenchmarkDlError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = dl.DLError()
	}
}

func BenchmarkAsmccEmpty(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dl.EmptyFunc()
	}
}

func BenchmarkCgoccEmpty(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dl.EmptyFunc2()
	}
}

// EOF
