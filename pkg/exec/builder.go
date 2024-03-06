package exec

import (
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"io"
)

type processBuilder struct {
	p process
}

func NewProcessBuilder(logger *log.Helper) *processBuilder {
	return &processBuilder{
		p: process{
			logger: logger,
		},
	}
}

func (pb *processBuilder) Name(name string) *processBuilder {
	pb.p.name = name
	return pb
}

func (pb *processBuilder) Executable(executable string) *processBuilder {
	pb.p.executable = executable
	return pb
}

func (pb *processBuilder) WorkDir(workDir string) *processBuilder {
	pb.p.workDir = workDir
	return pb
}

func (pb *processBuilder) Args(args ...string) *processBuilder {
	pb.p.args = args
	return pb
}

func (pb *processBuilder) Env(env map[string]string) *processBuilder {
	pb.p.env = env
	return pb
}

func (pb *processBuilder) Stdout(writer io.Writer) *processBuilder {
	pb.p.stdout = writer
	return pb
}

func (pb *processBuilder) Stderr(writer io.Writer) *processBuilder {
	pb.p.stderr = writer
	return pb
}

func (pb *processBuilder) Build() *process {
	return &pb.p
}
