.PHONY: build run clean
MODEL ?= models/

build:
	cd llama.cpp && mkdir -p build && cd build && \
	cmake .. -DBUILD_SHARED_LIBS=ON && \
	cmake --build . --config Release

run: build
	LD_LIBRARY_PATH=$(PWD)/llama.cpp/build/bin go run ./examples/main.go -model $(MODEL)

clean:
	rm -rf llama.cpp/build
