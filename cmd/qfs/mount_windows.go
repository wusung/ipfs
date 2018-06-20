package qfs

import (
	"fmt"
	"github.com/gonuts/flag"
	"github.com/jbenet/commander"
)

var cmdIpfsMount = &commander.Command{
	UsageLine: "mount",
	Short:     "Mount an ipfs read-only mountpoint.",
	Long:      `Not yet implemented on windows.`,
	Run:       mountCmd,
	Flag:      *flag.NewFlagSet("ipfs-mount", flag.ExitOnError),
}

func mountCmd(c *commander.Command, inp []string) error {
	fmt.Printf("not implemented yet\n")

	return nil
}
