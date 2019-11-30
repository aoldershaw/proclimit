// +build linux

package proclimit

import (
	"fmt"
	"github.com/containerd/cgroups"
	"github.com/friendsofgo/errors"
	"github.com/opencontainers/runtime-spec/specs-go"
	"os/exec"
)

// Option allows for customizing the behaviour of the Cgroup limiter.
//
// Options will typically perform operations on cgroup.LinuxResources.
type Option func(cgroup *Cgroup)

// WithName sets the name of the Cgroup. If not specified, a random UUID will
// be generated as the name
func WithName(name string) Option {
	return func(cgroup *Cgroup) {
		cgroup.Name = name
	}
}

// WithCPULimit sets the maximum CPU limit (as a percentage) allowed for all processes within the Cgroup.
// The percentage is based on a single CPU core. That is to say, 50 allows for the use of half of a core,
// 200 allows for the use of two cores, etc.
//
// `cpu.cfs_period_us` will be set to 100000 (100ms) unless it has been overridden (e.g. by an Option that
// is added before this Option. Note that no such Option has been implemented currently).
//
// `cpu.cfs_quota_us` will be set to cpuLimit percent of `cpu.cfs_period_us`.
func WithCPULimit(cpuLimit Percent) Option {
	return func(cgroup *Cgroup) {
		if cgroup.LinuxResources.CPU == nil {
			cgroup.LinuxResources.CPU = &specs.LinuxCPU{}
		}
		if cgroup.LinuxResources.CPU.Period == nil {
			cgroup.LinuxResources.CPU.Period = new(uint64)
			*cgroup.LinuxResources.CPU.Period = 100000
		}
		cgroup.LinuxResources.CPU.Quota = new(int64)
		*cgroup.LinuxResources.CPU.Quota = int64(*cgroup.LinuxResources.CPU.Period * uint64(cpuLimit) / 100)
	}
}

// WithMemoryLimit sets the maximum amount of virtual memory allowed for all processes within the Cgroup.
//
// `memory.max_usage_in_bytes` is set to memory
func WithMemoryLimit(memory Memory) Option {
	return func(cgroup *Cgroup) {
		if cgroup.LinuxResources.Memory == nil {
			cgroup.LinuxResources.Memory = &specs.LinuxMemory{}
		}
		cgroup.LinuxResources.Memory.Limit = new(int64)
		*cgroup.LinuxResources.Memory.Limit = int64(memory)
	}
}

// Cgroup represents a cgroup in a Linux system. Resource limits can be
// configured by modifying LinuxResources through Options. Modifying
// LinuxResources after calling New(...) will have no effect.
type Cgroup struct {
	Name           string
	LinuxResources *specs.LinuxResources
	cgroup         cgroups.Cgroup
}

// New creates a new Cgroup. Resource limits and the name of the Cgroup can be defined
// using Option arguments.
func New(options ...Option) (*Cgroup, error) {
	c := &Cgroup{
		LinuxResources: &specs.LinuxResources{},
	}
	for _, opt := range options {
		opt(c)
	}
	var err error
	if c.Name == "" {
		c.Name, err = randomName()
		if err != nil {
			return nil, err
		}
	}
	c.cgroup, err = cgroups.New(cgroups.V1, cgroups.StaticPath(fmt.Sprintf("/%s", c.Name)), c.LinuxResources)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cgroup")
	}
	return c, nil
}

// New loads an existing Cgroup by name
func Existing(name string) (*Cgroup, error) {
	c := &Cgroup{
		Name: name,
	}
	var err error
	c.cgroup, err = cgroups.Load(cgroups.V1, cgroups.StaticPath(fmt.Sprintf("/%s", name)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to load cgroup")
	}
	return c, nil
}

// Command constructs a wrapped Cmd struct to execute the named program with the given arguments.
// This wrapped Cmd will be added to the Cgroup when it is started.
func (c *Cgroup) Command(name string, arg ...string) *Cmd {
	return c.Wrap(exec.Command(name, arg...))
}

// Wrap takes an existing exec.Cmd and converts it into a proclimit.Cmd.
// When the returned Cmd is started, it will have the resources applied.
//
// If cmd has already been started, Wrap will panic. To limit the resources
// of a running process, use Limit instead.
func (c *Cgroup) Wrap(cmd *exec.Cmd) *Cmd {
	if cmd.Process != nil {
		panic("cmd has already been started")
	}
	return &Cmd{
		Cmd:     cmd,
		Limiter: c,
	}
}

// Limit applies Cgroup resource limits to a running process by its pid.
func (c *Cgroup) Limit(pid int) error {
	return c.cgroup.Add(cgroups.Process{Pid: pid})
}

// Close deletes the Cgroup definition from the filesystem.
func (c *Cgroup) Close() error {
	return c.cgroup.Delete()
}
