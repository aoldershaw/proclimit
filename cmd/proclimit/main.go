package main

import (
	"github.com/aoldershaw/proclimit"
	"log"
	"os"
	"os/exec"
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
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
			os.Exit(exitCode)
		}
	}()
	defer limiter.Close()

	cmd := limiter.Command(args.Path, args.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Run()
}
