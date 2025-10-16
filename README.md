# Alpaca

The motivation for this project is learning how to create basic [LLama.cpp](https://github.com/ggerganov/llama.cpp) bindings for Go.

## Running the example

Ensure that Go is installed on your system. Installation instructions are provided [here](https://go.dev/doc/install). This project uses Go 1.25.1 but other versions should be fine.

Install the dependencies required for building llama.cpp (C++). Instructions for installing vary depending on operating system. On Arch Linux, the following command will install all of the prerequisites:

```
sudo pacman -S  --needed base-devel cmake curl
```

The Ubuntu equivalent is:

```
sudo apt install build-essential cmake libcurl4-openssl-dev
```

Clone the repository:

```
git clone --recurse-submodules https://github.com/matthiase/alpaca
```

Note: This repository uses git sub modules to track [LLama.cpp](https://github.com/ggerganov/llama.cpp).

To build the bindings locally, run:

```
cd alpaca & make build
```

Assuming the build is successful, there is one more step necessary before being able to run the example. You will need to provide llama.cpp with a model. Since this is an experiment, let's use TinyLlama:

```
curl -L -o tinyllama-1.1b-chat-v1.0.Q4_K_M.gguf \
  https://huggingface.co/TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF/resolve/main/tinyllama-1.1b-chat-v1.0.Q4_K_M.gguf
```

Once the model has been downloaded, you can run the example with:

```
make run MODEL=tinyllama-1.1b-chat-v1.0.Q4_K_M.gguf
```

That's it! For more detailed notes see [notes.md](notes.md)

## Next Steps

* Set up a Github action that builds the Go package for Linux, Windows and MacOS
