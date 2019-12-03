// +build linux

package proclimit_test

import (
	"github.com/aoldershaw/proclimit"
	"os"
)

// TODO: add integration tests

func ExampleNew() {
	limiter, err := proclimit.New(
		proclimit.WithName("my-cgroup"),
		proclimit.WithCPULimit(proclimit.Percent(50)),
		proclimit.WithMemoryLimit(512*proclimit.Megabyte),
	)
	if err != nil {
		// handle err
	}
	defer limiter.Close()

	cmd := limiter.Command("stress", "--cpu", "2", "--timeout", "10")
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		// handle err
	}
}
