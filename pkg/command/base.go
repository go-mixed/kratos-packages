package command

import (
	"github.com/spf13/cobra"
)

type BaseCmd struct {
	cmd *cobra.Command
}

func NewBaseCmd(cmd *cobra.Command) *BaseCmd {
	return &BaseCmd{
		cmd: cmd,
	}
}

type ICmder interface {
	GetCmd() *cobra.Command
	AddCommand(commands ...ICmder)
	Execute() error
}

type Commands []ICmder

var _ ICmder = (*BaseCmd)(nil)

func (c *BaseCmd) GetCmd() *cobra.Command {
	return c.cmd
}

func (c *BaseCmd) AddCommand(commands ...ICmder) {
	for _, command := range commands {
		c.cmd.AddCommand(command.GetCmd())
	}
}

func (c *BaseCmd) Execute() error {
	return c.cmd.Execute()
}
