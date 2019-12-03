// +build windows

package proclimit

import (
	"github.com/aoldershaw/proclimit/internal/win32"
	"github.com/friendsofgo/errors"
	"os/exec"
	"runtime"
)

// TODO: godocs

type Option func(jobObject *JobObject)

func WithName(name string) Option {
	return func(jobObject *JobObject) {
		jobObject.Name = name
	}
}

func WithCPULimit(cpuLimit Percent) Option {
	return func(jobObject *JobObject) {
		if jobObject.CPULimitInformation == nil {
			jobObject.CPULimitInformation = &win32.JobObjectCPURateControlInformation{}
		}
		jobObject.CPULimitInformation.CPURate = uint32(cpuLimit*100) / uint32(runtime.NumCPU())
		jobObject.CPULimitInformation.ControlFlags |= win32.JOB_OBJECT_CPU_RATE_CONTROL_ENABLE
		jobObject.CPULimitInformation.ControlFlags |= win32.JOB_OBJECT_CPU_RATE_CONTROL_HARD_CAP
	}
}

func WithMemoryLimit(mem Memory) Option {
	return func(jobObject *JobObject) {
		if jobObject.ExtendedLimitInformation == nil {
			jobObject.ExtendedLimitInformation = &win32.JobObjectExtendedLimitInformation{}
		}
		jobObject.ExtendedLimitInformation.ProcessMemoryLimit = uintptr(mem)
		jobObject.ExtendedLimitInformation.BasicLimitInformation.LimitFlags |= win32.JOB_OBJECT_LIMIT_PROCESS_MEMORY
	}
}

type JobObject struct {
	Name                     string
	ExtendedLimitInformation *win32.JobObjectExtendedLimitInformation
	CPULimitInformation      *win32.JobObjectCPURateControlInformation
	handle                   win32.Handle
}

func New(options ...Option) (*JobObject, error) {
	j := &JobObject{}
	for _, opt := range options {
		opt(j)
	}
	var err error
	if j.Name == "" {
		j.Name, err = randomName()
		if err != nil {
			return nil, err
		}
	}
	j.handle, err = win32.CreateJobObject(nil, j.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create job object %s", j.Name)
	}
	if j.CPULimitInformation != nil {
		err = win32.SetInformationJobObject_CPURateControlInformation(j.handle, j.CPULimitInformation)
		if err != nil {
			return nil, errors.Wrap(err, "failed to set CPU rate information")
		}
	}
	if j.ExtendedLimitInformation != nil {
		err = win32.SetInformationJobObject_ExtendedLimitInformation(j.handle, j.ExtendedLimitInformation)
		if err != nil {
			return nil, errors.Wrap(err, "failed to set extended limit information")
		}
	}
	return j, nil
}

// TODO: implement Existing

func (j *JobObject) Command(name string, arg ...string) *Cmd {
	return j.Wrap(exec.Command(name, arg...))
}

func (j *JobObject) Wrap(cmd *exec.Cmd) *Cmd {
	if cmd.Process != nil {
		panic("cmd has already been started")
	}
	return &Cmd{
		Cmd:     cmd,
		Limiter: j,
	}
}

func (j *JobObject) Limit(pid int) error {
	if pid == 0 {
		return errors.New("must provide a valid pid")
	}
	handle, err := win32.OpenProcess(win32.STANDARD_RIGHTS_READ|win32.PROCESS_QUERY_INFORMATION|win32.SYNCHRONIZE|win32.PROCESS_SET_INFORMATION,
		false, uint32(pid))
	if err != nil {
		return errors.Wrap(err, "failed to open process handle")
	}
	defer win32.CloseHandle(handle)
	if err = win32.AssignProcessToJobObject(j.handle, handle); err != nil {
		return errors.Wrap(err, "failed to assign process to job object")
	}
	return nil
}

func (j *JobObject) Close() error {
	// TODO: also close all processes that have been limited?
	// https://docs.microsoft.com/en-us/windows/win32/procthread/job-objects#managing-job-objects
	// "The job is destroyed when its last handle has been closed and all associated processes have been terminated"
	return win32.CloseHandle(j.handle)
}
