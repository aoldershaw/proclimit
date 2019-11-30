# proclimit

proclimit is a Go library for running external commands with configurable resource (CPU and memory) limits. 
Currently, only Linux (cgroups) and Windows (Job Objects) are supported.

The project also includes an application to run processes with limited resources.

## Note

proclimit is still very early in development and requires more testing (particularly on the Windows side, as I don't have easy access to a Windows machine).

## Installation

```bash
go get github.com/aoldershaw/proclimit
```

To install and run the application:

```bash
go get github.com/aoldershaw/proclimit/cmd/proclimit

proclimit -cpu=50 -memory=512M my-application arg1 arg2
```

## Usage

```go
func main() {
    limiter, _ := proclimit.New(
        proclimit.WithCPULimit(proclimit.Percent(50)),
        proclimit.WithMemoryLimit(512 * proclimit.Megabyte),
    )
    defer limiter.Close()

    // limiter.Command is nearly identical to exec.Command, and can be used in the same way
    cmd := limiter.Command("my-application", "arg1", "arg2")
    cmd.Stdout = os.Stdout
    cmd.Run()
}
```

```go
func main() {
    limiter, _ := proclimit.New(...)
    ...
    // proclimit can also limit resources of currently running processes by pid
    limiter.Limit(1234)
}
```

## License
[MIT](https://choosealicense.com/licenses/mit/)