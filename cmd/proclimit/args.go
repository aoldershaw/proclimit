package main

import (
	"flag"
	"fmt"
	"github.com/aoldershaw/proclimit"
	"github.com/friendsofgo/errors"
	"runtime"
	"strconv"
	"strings"
)

type cmdArgs struct {
	Name        string
	CPULimit    uint
	MemoryLimit proclimit.Memory

	Path string
	Args []string
}

func parseArgs() (cmdArgs, error) {
	a := cmdArgs{}
	var limiterName string
	switch runtime.GOOS {
	case "windows":
		limiterName = "job object"
	case "linux":
		limiterName = "cgroup"
	default:
		limiterName = "<<unknown>>"
	}
	flag.StringVar(&a.Name, "name", "", fmt.Sprintf("name of the %s. If not specified, a random name will be generated", limiterName))
	flag.UintVar(&a.CPULimit, "cpu", 0, "maximum CPU percentage based on a single core (100 = 1 core)")
	memoryString := flag.String("memory", "", "maximum memory usage in bytes (e.g. 1G)")
	flag.Parse()
	if memoryString != nil && *memoryString != "" {
		var err error
		a.MemoryLimit, err = parseMemory(*memoryString)
		if err != nil {
			return cmdArgs{}, err
		}
	}
	if flag.NArg() < 1 {
		flag.Usage()
		return cmdArgs{}, errors.Errorf("Usage: %s")
	}
	a.Path = flag.Arg(0)
	a.Args = flag.Args()[1:]
	return a, nil
}

// Supports 2G, 2M, 2K, or 2 (bytes)
func parseMemory(memory string) (proclimit.Memory, error) {
	var factor proclimit.Memory = 1
	hasSuffix := false
	if hasMemorySuffix(memory, "G") {
		factor = proclimit.Gigabyte
		hasSuffix = true
	} else if hasMemorySuffix(memory, "M") {
		factor = proclimit.Megabyte
		hasSuffix = true
	} else if hasMemorySuffix(memory, "K") {
		factor = proclimit.Kilobyte
		hasSuffix = true
	}
	if hasSuffix {
		memory = memory[0 : len(memory)-1]
	}
	num, err := strconv.ParseInt(memory, 10, 32)
	if err != nil {
		return 0, errors.Wrap(err, "invalid memory value")
	}
	return proclimit.Memory(num) * factor, nil
}

func hasMemorySuffix(s, letter string) bool {
	uppercase := strings.ToUpper(s)
	return strings.HasSuffix(uppercase, letter)
}
