package nodegroup

import (
	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/app/wwctl/nodegroup/list"
)

var baseCmd = &cobra.Command{
	DisableFlagsInUseLine: true,
	Use:                   "nodegroup COMMAND [OPTIONS]",
	Short:                 "Node group management",
	Long:                  "Inspect node groups defined in nodes.conf.",
	Args:                  cobra.NoArgs,
}

func init() {
	baseCmd.AddCommand(list.GetCommand())
}

// GetCommand returns the nodegroup subcommand tree.
func GetCommand() *cobra.Command {
	return baseCmd
}
