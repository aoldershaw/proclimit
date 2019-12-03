package main

import (
	"github.com/aoldershaw/proclimit"
	"log"
	"os"
	"os/signal"
	"runtime"
)

func main() {
	args, err := parseArgs()
	if err != nil {
		log.Fatal(err)
	}
	var opts []proclimit.Option
	if args.Name != "" {
		opts = append(opts, proclimit.WithName(args.Name))
	}
	if args.MemoryLimit > 0 {
		opts = append(opts, proclimit.WithMemoryLimit(args.MemoryLimit))
	}
	if args.CPULimit > 0 {
		cpuLimit := proclimit.Percent(args.CPULimit)
		opts = append(opts, proclimit.WithCPULimit(cpuLimit))
	}
	limiter, err := proclimit.New(opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err != nil {
			exitCode := 1
			if exitErr, ok := err.(interface{ ExitCode() int }); ok {
				exitCode = exitErr.ExitCode()
			} else {
				log.Println(err)
			}
			os.Exit(exitCode)
		}
	}()
	defer limiter.Close()

	cmd := limiter.Command(args.Path, args.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	go func() {
		sig := <-sigs
		if cmd.Process != nil {
			// https://golang.org/pkg/os/#Signal
			// "On Windows, sending os.Interrupt to a process with os.Process.Signal is not implemented"
			if runtime.GOOS == "windows" {
				cmd.Process.Kill()
			} else {
				cmd.Process.Signal(sig)
			}
		}
	}()

	err = cmd.Run()
}
