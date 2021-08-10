package commands

import (
	cranecmd "github.com/google/go-containerregistry/cmd/crane/cmd"
	"github.com/spf13/cobra"
)

var Root = New()

func New() *cobra.Command {
	root := &cobra.Command{
		Use:          "ko",
		Short:        "Rapidly iterate with Go, Containers, and Kubernetes.",
		SilenceUsage: true, // Don't show usage on errors
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	AddKubeCommands(root)

	// Also add the auth group from crane to facilitate logging into a
	// registry.
	authCmd := cranecmd.NewCmdAuth("ko", "auth")
	// That was a mistake, but just set it to Hidden so we don't break people.
	authCmd.Hidden = true
	root.AddCommand(authCmd)

	// Just add a `ko login` command:
	root.AddCommand(cranecmd.NewCmdAuthLogin("ko"))
	return root
}
