package util

import (
	"fmt"
	"os"
	"os/user"
	"strings"
)

var Debug bool
var NotImplementedError = fmt.Errorf("Error: not implemented yet.")

// a Key for maps. It's a string (rep of a multihash).
type Key string

// Shorthand printing functions.
func PErr(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

// tilde expansion
func TildeExpansion(filename string) (string, error) {
	if strings.HasPrefix(filename, "~/") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}

		dir := usr.HomeDir + "/"
		filename = strings.Replace(filename, "~/", dir, 1)
	}
	return filename, nil
}

func POut(format string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, format, a...)
}

func DErr(format string, a ...interface{}) {
	if Debug {
		PErr(format, a...)
	}
}

func DOut(format string, a ...interface{}) {
	if Debug {
		POut(format, a...)
	}
}
