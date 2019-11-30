# proclimit

proclimit is a Go library for running external commands with configurable resource (CPU and memory) limits. 
Currently, only Linux (cgroups) and Windows (Job Objects) are supported.

The project also includes an application to run processes with limited resources.

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
        // The CPU limit is relative to a single core. Specifying 200% on a 4 core machine
        // restricts the total CPU usage of all processes in the limiter to use up to 2 cores
        //
        // It is not guaranteed that the processes will only be scheduled on 2 physical cores -
        // in the example above, it is possible that each of the 4 cores will be at 50% utilization
        // (meaning the total CPU usage is 2 "full" cores)
        proclimit.WithCPULimit(proclimit.Percent(50)),
        // The memory limit is based on total virtual memory
        proclimit.WithMemoryLimit(512 * proclimit.Megabyte),
    )
    defer limiter.Close()

    // limiter.Command is nearly identical to exec.Command - the returned *Cmd can be treated as an *exec.Cmd
    // However, when the *Cmd is started (through Start, Run, Output, or CombinedOutput), it will run with
    // limited resources
    cmd1 := limiter.Command("application1", "arg1", "arg2")
    cmd1.Stdout = os.Stdout

    // application1 will be limited to 512M of virtual memory and 50% of a single core's compute
    cmd1.Start()

    cmd2 := limiter.Command("application2")
    
    // Since application2 is run in the same limiter, the CPU and memory limits apply to the
    // combined utilization of application1 and application2.
    // If application1 uses 40% CPU, then application2 can only use up to 10%
    out, _ := cmd2.Output()
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

## Note

* proclimit is still very early in development and requires more testing (particularly on the Windows side, as I don't have easy access to a Windows machine).
* Only Linux and Windows are supported at the moment
* Processes will first start with no limits applied. If it is important that a process start up with the limits applied (for instance, if using github.com/uber-go/automaxprocs in the application being started), proclimit is currently not the tool for the job.

## License
[MIT](https://choosealicense.com/licenses/mit/)