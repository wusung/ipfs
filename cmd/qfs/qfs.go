package qfs

import (
	"fmt"
	"github.com/gonuts/flag"
	"github.com/jbenet/commander"
	config "../../config"
	core "../../core"
	u "../../util"
	"os"
)

// The IPFS command tree. It is an instance of `commander.Command`.
var CmdIpfs = &commander.Command{
	UsageLine: "ipfs [<flags>] <command> [<args>]",
	Short:     "global versioned p2p merkledag file system",
	Long: `ipfs - global versioned p2p merkledag file system

Basic commands:

    add <path>    Add an object to ipfs.
    cat <ref>     Show ipfs object data.
    ls <ref>      List links from an object.
    refs <ref>    List link hashes from an object.

Tool commands:

    config        Manage configuration.
    version       Show ipfs version information.
    commands      List all available commands.

Advanced Commands:

    mount         Mount an ipfs read-only mountpoint.

Use "ipfs help <command>" for more information about a command.
`,
	Run: ipfsCmd,
	Subcommands: []*commander.Command{
		cmdIpfsAdd,
		cmdIpfsCat,
		cmdIpfsLs,
		cmdIpfsRefs,
		cmdIpfsConfig,
		cmdIpfsVersion,
		cmdIpfsCommands,
		cmdIpfsMount,
	},
	Flag: *flag.NewFlagSet("ipfs", flag.ExitOnError),
}

func ipfsCmd(c *commander.Command, args []string) error {
	u.POut(c.Long)
	return nil
}

func main() {
	err := CmdIpfs.Dispatch(os.Args[1:])
	if err != nil {
		if len(err.Error()) > 0 {
			fmt.Fprintf(os.Stderr, "ipfs %s: %v\n", os.Args[1], err)
		}
		os.Exit(1)
	}
	return
}

func localNode() (*core.IpfsNode, error) {
	//todo implement config file flag
	cfg, err := config.Load("")
	if err != nil {
		return nil, err
	}

	return core.NewIpfsNode(cfg)
}
