// Package bindings provides Go bundings for llama.cpp
package bindings

// #cgo CFLAGS: -I${SRCDIR}/../llama.cpp/include -I${SRCDIR}/../llama.cpp/ggml/include
// #cgo LDFLAGS: -L${SRCDIR}/../llama.cpp/build/bin -lllama -lstdc++ -lm
// #cgo darwin LDFLAGS: -framework Accelerate -framework Foundation -framework Metal -framework MetalKit
// #include <stdlib.h>
// #include "llama.h"
import "C"

import (
	"fmt"
	"unsafe"
)

type Model struct {
	ptr *C.struct_llama_model
}

// Init initializes the llama backend
func Init() {
	C.llama_backend_init()
}

// Free frees the llama backend
func Free() {
	C.llama_backend_free()
}

// LoadModel loads a GGUF model from the given path
func LoadModel(path string) (*Model, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	params := C.llama_model_default_params()
	modelPtr := C.llama_model_load_from_file(cPath, params)

	if modelPtr == nil {
		return nil, fmt.Errorf("failed to load model: %s", path)
	}

	return &Model{ptr: modelPtr}, nil
}

// Free frees the model
func (m *Model) Free() {
	if m.ptr != nil {
		C.llama_model_free(m.ptr)
		m.ptr = nil
	}
}

// VocabSize returns the vocabulary size
func (m *Model) VocabSize() int {
	vocab := C.llama_model_get_vocab(m.ptr)
	return int(C.llama_vocab_n_tokens(vocab))
}

// ContextSize returns the context size
func (m *Model) ContextSize() int {
	return int(C.llama_model_n_ctx_train(m.ptr))
}
