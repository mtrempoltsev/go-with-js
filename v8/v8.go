package v8

// #cgo CFLAGS: -I../thirdparty/v8capi/include -O0 -g
// #cgo LDFLAGS: -L../out -lv8capi -lv8_monolith
// #cgo LDFLAGS: -lstdc++ -lm
// #include <stdlib.h>
// #include <v8capi.h>
import "C"

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"unsafe"
)

func makeError(err C.struct_v8_error) string {
	if err.location == nil {
		return C.GoString(err.message)
	}

	buf := strings.Builder{}

	fmt.Fprintf(&buf,
		"%s:%d: %s\n%s",
		C.GoString(err.location),
		err.line_number,
		C.GoString(err.message),
		C.GoString(err.wavy_underline))

	if err.stack_trace != nil {
		fmt.Fprintf(&buf,
			"\nstack trace:\n%s",
			C.GoString(err.stack_trace))
	}

	return buf.String()
}

type Isolate struct {
	ptr *C.struct_v8_isolate
}

func (isolate *Isolate) Compile(code, location string) (*Script, error) {
	codePtr := C.CString(code)
	defer C.free(unsafe.Pointer(codePtr))

	locationPtr := C.CString(location)
	defer C.free(unsafe.Pointer(locationPtr))

	var err C.struct_v8_error
	defer C.v8_delete_error(&err)

	script := C.v8_compile_script(isolate.ptr, codePtr, locationPtr, &err)

	if script == nil {
		return nil, errors.New(makeError(err))
	}

	return &Script{script}, nil
}

func (isolate *Isolate) Dispose() {
	C.v8_delete_isolate(isolate.ptr)
}

type Script struct {
	ptr *C.struct_v8_script
}

func (script *Script) Run() (*Value, error) {
	var res C.struct_v8_value
	defer C.v8_delete_value(&res)

	var err C.struct_v8_error
	defer C.v8_delete_error(&err)

	if !C.v8_run_script(script.ptr, &res, &err) {
		return nil, errors.New(makeError(err))
	}

	return &Value{data: res}, nil
}

func (script *Script) Dispose() {
	C.v8_delete_script(script.ptr)
}

type Function struct {
	ptr *C.struct_v8_callable
}

func (function *Function) Dispose() {
	C.v8_delete_function(function.ptr)
}

type Instance struct {
	ptr *C.struct_v8_instance
}

func New() *Instance {
	path := C.CString(os.Args[0])
	defer C.free(unsafe.Pointer(path))

	return &Instance{
		ptr: C.v8_new_instance(path),
	}
}

func (*Instance) NewIsolate() *Isolate {
	return &Isolate{
		ptr: C.v8_new_isolate(),
	}
}

func (instance *Instance) Dispose() {
	C.v8_delete_instance(instance.ptr)
}
