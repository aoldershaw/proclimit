// Package proclimit provides a consistent cross-platform API for limiting
// CPU and memory resources to processes. It also provides a wrapper Cmd type
// that mimics exec.Cmd, while applying resource limits upon starting the commands.
package proclimit

import (
	"bytes"
	"github.com/friendsofgo/errors"
	"os/exec"
	"strconv"
)

// Percent is a percentage value. It is used to specify CPU rate limits.
type Percent uint

// Memory represents a number of bytes
type Memory uint

const (
	Byte     Memory = 1
	Kilobyte        = 1024 * Byte
	Megabyte        = 1024 * Kilobyte
	Gigabyte        = 1024 * Megabyte
)

// Limiter allows limiting a running process' resources
type Limiter interface {
	// Limit applies limits to a running process by its pid.
	// The specific limits to apply are defined by the implementation.
	Limit(pid int) error
}

// Cmd represents an external command being prepared or run.
// This command will be limited by the provided Limiter.
//
// It can be used exactly as an exec.Cmd can be used.
type Cmd struct {
	*exec.Cmd
	Limiter Limiter
}

// Start begins the execution of a Cmd, and applies the limits defined by the
// associated Limiter. If the Limiter fails to apply limits, the process will be killed.
//
// Note that the Cmd will start before the limits are applied, so there will be a brief
// period where the limits are not enforced.
func (c *Cmd) Start() error {
	if err := c.Cmd.Start(); err != nil {
		return err
	}
	if err := c.Limiter.Limit(c.Process.Pid); err != nil {
		c.Process.Kill()
		return errors.Wrap(err, "failed to limit command")
	}
	return nil
}

// Run starts the specified command (with limits), and waits for it to complete.
func (c *Cmd) Run() error {
	if err := c.Start(); err != nil {
		return err
	}
	return c.Cmd.Wait()
}

// Output runs the command (with limits) and returns its standard output.
func (c *Cmd) Output() ([]byte, error) {
	if c.Stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	var stdout bytes.Buffer
	c.Stdout = &stdout

	captureErr := c.Stderr == nil
	if captureErr {
		c.Stderr = &prefixSuffixSaver{N: 32 << 10}
	}

	err := c.Run()
	if err != nil && captureErr {
		if ee, ok := err.(*exec.ExitError); ok {
			ee.Stderr = c.Stderr.(*prefixSuffixSaver).Bytes()
		}
	}
	return stdout.Bytes(), err
}

// CombinedOutput runs the command (with limits) and returns its combined standard output and standard error.
func (c *Cmd) CombinedOutput() ([]byte, error) {
	if c.Stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	if c.Stderr != nil {
		return nil, errors.New("exec: Stderr already set")
	}
	var b bytes.Buffer
	c.Stdout = &b
	c.Stderr = &b
	err := c.Run()
	return b.Bytes(), err
}

// prefixSuffixSaver is an io.Writer which retains the first N bytes
// and the last N bytes written to it. The Bytes() methods reconstructs
// it with a pretty error message.
//
// This was copied directly from the exec package.
type prefixSuffixSaver struct {
	N         int // max size of prefix or suffix
	prefix    []byte
	suffix    []byte // ring buffer once len(suffix) == N
	suffixOff int    // offset to write into suffix
	skipped   int64
}

func (w *prefixSuffixSaver) Write(p []byte) (n int, err error) {
	lenp := len(p)
	p = w.fill(&w.prefix, p)

	// Only keep the last w.N bytes of suffix data.
	if overage := len(p) - w.N; overage > 0 {
		p = p[overage:]
		w.skipped += int64(overage)
	}
	p = w.fill(&w.suffix, p)

	// w.suffix is full now if p is non-empty. Overwrite it in a circle.
	for len(p) > 0 { // 0, 1, or 2 iterations.
		n := copy(w.suffix[w.suffixOff:], p)
		p = p[n:]
		w.skipped += int64(n)
		w.suffixOff += n
		if w.suffixOff == w.N {
			w.suffixOff = 0
		}
	}
	return lenp, nil
}

// fill appends up to len(p) bytes of p to *dst, such that *dst does not
// grow larger than w.N. It returns the un-appended suffix of p.
func (w *prefixSuffixSaver) fill(dst *[]byte, p []byte) (pRemain []byte) {
	if remain := w.N - len(*dst); remain > 0 {
		add := minInt(len(p), remain)
		*dst = append(*dst, p[:add]...)
		p = p[add:]
	}
	return p
}

func (w *prefixSuffixSaver) Bytes() []byte {
	if w.suffix == nil {
		return w.prefix
	}
	if w.skipped == 0 {
		return append(w.prefix, w.suffix...)
	}
	var buf bytes.Buffer
	buf.Grow(len(w.prefix) + len(w.suffix) + 50)
	buf.Write(w.prefix)
	buf.WriteString("\n... omitting ")
	buf.WriteString(strconv.FormatInt(w.skipped, 10))
	buf.WriteString(" bytes ...\n")
	buf.Write(w.suffix[w.suffixOff:])
	buf.Write(w.suffix[:w.suffixOff])
	return buf.Bytes()
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
