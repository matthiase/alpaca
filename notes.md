# Building Go Bindings for llama.cpp - Lessons Learned

## Project Overview

This document captures the key learnings from creating minimal CGo bindings for llama.cpp in a Go project intended for desktop applications across Linux, Windows, and macOS.

## Key Decisions

### 1. Separate Bindings Project vs. Embedded

**Decision**: Create a separate project for bindings rather than embedding in the main application.

**Reasoning**:

- Wails framework has its own build process that adds complexity
- Bindings can be reused across multiple projects
- Cleaner separation of concerns
- Independent testing and versioning

**When to embed instead**:

- Single application with no plans for reuse
- Simpler development workflow for solo projects
- Application-specific optimizations needed

### 2. Cross-Compilation Reality

**Important**: CGo does NOT support easy cross-compilation for Windows/macOS/Linux.

**Why**:

- Requires target platform's C compiler and toolchain
- Platform-specific system libraries and frameworks
- Different shared library formats (.so, .dll, .dylib)

**Solution**: Build natively on each target platform

- Use GitHub Actions with matrix builds for CI/CD
- Each platform builds its own binaries
- Distribute platform-specific releases

## Project Structure

```
alpaca/                          # or your project name
├── bindings/
│   └── llama.go                # CGo bindings
├── examples/
│   └── main.go                 # Example usage
├── llama.cpp/                  # Git submodule (pinned to specific tag)
├── lib/                        # Compiled shared libraries (gitignored)
├── models/                     # GGUF models (gitignored)
├── go.mod
├── Makefile
└── README.md
```

## Setting Up llama.cpp Submodule

### Add and Pin to Specific Version

```bash
# Add the submodule
git submodule add https://github.com/ggerganov/llama.cpp.git

# Pin to a specific tag (recommended for stability)
cd llama.cpp
git fetch --tags
git checkout b4226  # or your chosen tag
cd ..

# Commit the pinned version
git add .gitmodules llama.cpp
git commit -m "Add llama.cpp submodule pinned to b4226"
```

### View Available Tags

```bash
cd llama.cpp
git fetch --tags
git tag -l | tail -20  # See recent tags
```

### Update to Newer Version Later

```bash
cd llama.cpp
git fetch --tags
git checkout b4300  # newer tag
cd ..
git add llama.cpp
git commit -m "Update llama.cpp to b4300"
```

## Platform-Specific Dependencies

### Pop!_OS / Ubuntu / Debian

```bash
sudo apt install build-essential cmake libcurl4-openssl-dev
```

**Optional GPU support**:

```bash
# NVIDIA CUDA
sudo apt install nvidia-cuda-toolkit

# Vulkan
sudo apt install libvulkan-dev
```

### Arch Linux

```bash
sudo pacman -S base-devel cmake curl
```

**Optional GPU support**:

```bash
# NVIDIA CUDA
sudo pacman -S cuda

# AMD ROCm
sudo pacman -S rocm-hip-sdk rocm-opencl-sdk

# Vulkan
sudo pacman -S vulkan-headers vulkan-icd-loader
```

### macOS

```bash
brew install cmake
```

(Xcode Command Line Tools provide the compiler)

### Windows

```bash
choco install cmake
```

(Or install Visual Studio with C++ tools)

## CGo Include Path Configuration

### Critical Discovery: llama.cpp Header Structure

The llama.cpp project has headers in multiple locations:

- `llama.cpp/include/llama.h` - Main API header
- `llama.cpp/ggml/include/ggml.h` - GGML backend header

**Both paths must be in CFLAGS**:

```go
// #cgo CFLAGS: -I${SRCDIR}/../llama.cpp/include -I${SRCDIR}/../llama.cpp/ggml/include
```

### Include Path Rules

1. The `#include` directive in CGo should use just the filename:

   ```go
   // #include "llama.h"  // NOT "llama.cpp/include/llama.h"
   ```

2. The `-I` flags in CFLAGS tell the compiler where to search:

   ```go
   // #cgo CFLAGS: -I${SRCDIR}/../llama.cpp/include
   ```

3. Always check header dependencies:

   ```bash
   # Find header files
   tree -L 3 llama.cpp/include
   find llama.cpp -name "*.h"
   ```

## API Version Changes

**Critical**: llama.cpp's API changes frequently. Many functions are deprecated.

### Deprecated → Current API Mapping

| Deprecated Function | Current Function |
|-------------------|------------------|
| `llama_load_model_from_file` | `llama_model_load_from_file` |
| `llama_free_model` | `llama_model_free` |
| `llama_n_ctx_train` | `llama_model_n_ctx_train` |
| `llama_n_vocab` | `llama_vocab_n_tokens` (requires `llama_model_get_vocab` first) |

**Lesson**: Always check the current llama.cpp header files for the latest API. Pinning to a specific tag helps maintain stability.

## Downloading Models

### Recommended for Testing: TinyLlama (637MB)

```bash
# Create models directory
mkdir -p models

# Download TinyLlama (small, fast for testing)
curl -L -o models/tinyllama-1.1b-chat-v1.0.Q4_K_M.gguf \
  https://huggingface.co/TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF/resolve/main/tinyllama-1.1b-chat-v1.0.Q4_K_M.gguf
```

### Larger Models: Llama-2-7B

```bash
# Q4_K_M quantization (~4GB) - good balance
curl -L -o models/llama-2-7b-chat.Q4_K_M.gguf \
  https://huggingface.co/TheBloke/Llama-2-7B-Chat-GGUF/resolve/main/llama-2-7b-chat.Q4_K_M.gguf
```

### Model Quantization Sizes (Llama-2-7B)

| Quantization | Size | Quality | Use Case |
|-------------|------|---------|----------|
| Q2_K | ~2.8GB | Lowest | Testing only |
| Q4_K_M | ~4.1GB | Good | **Recommended for most use** |
| Q5_K_M | ~4.8GB | Better | Higher quality needs |
| Q8_0 | ~7.2GB | Highest | Maximum quality |

## Minimal Working Makefile

```makefile
.PHONY: build-llama run clean

MODEL ?= models/tinyllama-1.1b-chat-v1.0.Q4_K_M.gguf

build-llama:
 cd llama.cpp && mkdir -p build && cd build && \
 cmake .. -DBUILD_SHARED_LIBS=ON && \
 cmake --build . --config Release
 mkdir -p lib
 cp llama.cpp/build/bin/libllama.* lib/

run: build-llama
 LD_LIBRARY_PATH=$(PWD)/lib go run ./examples/main.go -model $(MODEL)

clean:
 rm -rf llama.cpp/build lib/
```

### Usage

```bash
# Build and run with defaults
make run

# Run with custom model
make run MODEL=models/llama-2-7b-chat.Q4_K_M.gguf
```

## Minimal Working Bindings

### bindings/llama.go

```go
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
```

### examples/main.go

```go
package main

import (
 "flag"
 "fmt"
 "log"

 "github.com/matthiase/alpaca/bindings"  // Update with your module path
)

func main() {
 modelPath := flag.String("model", "", "Path to GGUF model")
 flag.Parse()

 if *modelPath == "" {
  log.Fatal("Please provide -model flag")
 }

 // Initialize backend
 bindings.Init()
 defer bindings.Free()

 // Load model
 fmt.Println("Loading model...")
 model, err := bindings.LoadModel(*modelPath)
 if err != nil {
  log.Fatal(err)
 }
 defer model.Free()

 // Print basic info
 fmt.Printf("✓ Model loaded successfully!\n")
 fmt.Printf("  Vocabulary size: %d\n", model.VocabSize())
 fmt.Printf("  Context size: %d\n", model.ContextSize())
}
```

## Common Issues and Solutions

### Issue: `fatal error: llama.h: No such file or directory`

**Solution**: Include path is wrong. Make sure CFLAGS includes both:

- `-I${SRCDIR}/../llama.cpp/include`
- `-I${SRCDIR}/../llama.cpp/ggml/include`

### Issue: `fatal error: ggml.h: No such file or directory`

**Solution**: Missing the ggml include path in CFLAGS.

### Issue: `could not determine what C.llama_batch_add refers to`

**Solution**: The function is likely a macro or inline function. You need to create a C wrapper function (bridge.c) to call it. For minimal bindings, avoid these complex functions initially.

### Issue: Deprecated function warnings

**Solution**: Use the new API functions as shown in the API mapping table above.

### Issue: `cannot use ... as ... value in argument`

**Solution**: API signature changed. Check the current llama.cpp headers for the correct function signature.

## Development Strategy

### Start Minimal

1. **First**: Just load a model and get basic info (vocabulary size, context size)
2. **Second**: Add tokenization
3. **Third**: Add context creation
4. **Fourth**: Add simple generation
5. **Finally**: Add advanced features (sampling, streaming, etc.)

### Incremental Complexity

Don't try to implement everything at once. Each step should:

- Compile successfully
- Run successfully
- Be tested before moving to the next feature

### When You Hit a Wall

If you encounter complex CGo issues:

1. Check if there's a simpler API function
2. Consider creating a C wrapper/bridge
3. Look at the llama.cpp examples directory for usage patterns
4. Pin to a specific llama.cpp version and stick with it

## .gitignore Recommendations

```
# Build artifacts
lib/
llama.cpp/build/

# Models (too large for git)
models/
*.gguf

# Go
*.exe
*.exe~
*.dll
*.so
*.dylib
```

## Next Steps

From this minimal foundation, you can add:

- Tokenization and detokenization
- Context management
- Text generation with sampling
- Streaming output
- Thread safety
- Better error handling
- Performance optimizations

## Resources

- llama.cpp GitHub: <https://github.com/ggerganov/llama.cpp>
- Hugging Face Models: <https://huggingface.co/models?library=gguf>
- CGo Documentation: <https://pkg.go.dev/cmd/cgo>
- Wails Framework: <https://wails.io>

## Conclusion

Building CGo bindings for llama.cpp is achievable but requires:

- Understanding of C/Go interop
- Awareness of API changes in llama.cpp
- Platform-specific build considerations
- Starting minimal and building incrementally

The payoff is native performance with full control over the integration.
