# oli

a openrouter cli terminal application.

## Install

```sh
go install github.com/janstuemmel/oli@latest
```

## Usage

```sh
# repl
oli 

# prompt as argument
oli hello world

# from stdin
echo "hello world" | oli

# configure
oli --model google/gemini-2.5-flash

# intercept answer
OLI_PIPE=cat oli
```
