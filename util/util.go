package util

import (
	"bytes"
	"fmt"
	"errors"
	b58 "github.com/jbenet/go-base58"
	mh "github.com/jbenet/go-multihash"
	"os"
	"os/user"
	"runtime"
	"strings"
)

// Debug is a global flag for debugging.
var Debug bool

// ErrNotImplemented signifies a function has not been implemented yet.
var ErrNotImplemented = errors.New("Error: not implemented yet.")

// ErrTimeout implies that a timeout has been triggered
var ErrTimeout = errors.New("Error: Call timed out.")

// ErrSeErrSearchIncomplete implies that a search type operation didnt
// find the expected node, but did find 'a' node.
var ErrSearchIncomplete = errors.New("Error: Search Incomplete.")

// ErrNotFound is returned when a search fails to find anything
var ErrNotFound = errors.New("Error: Not Found.")

// Key is a string representation of multihash for use with maps.
type Key string

func (k Key) Pretty() string {
	return b58.Encode([]byte(k))
}

type IpfsError struct {
	Inner error
	Note  string
	Stack string
}

func (ie *IpfsError) Error() string {
	buf := new(bytes.Buffer)
	fmt.Fprintln(buf, ie.Inner)
	fmt.Fprintln(buf, ie.Note)
	fmt.Fprintln(buf, ie.Stack)
	return buf.String()
}

func WrapError(err error, note string) error {
	ie := new(IpfsError)
	ie.Inner = err
	ie.Note = note
	stack := make([]byte, 2048)
	n := runtime.Stack(stack, false)
	ie.Stack = string(stack[:n])
	return ie
}

// Hash is the global IPFS hash function. uses multihash SHA2_256, 256 bits
func Hash(data []byte) (mh.Multihash, error) {
	return mh.Sum(data, mh.SHA2_256, -1)
}

// TildeExpansion expands a filename, which may begin with a tilde.
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

// PErr is a shorthand printing function to output to Stderr.
func PErr(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
}

// POut is a shorthand printing function to output to Stdout.
func POut(format string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, format+"\n", a...)
}

// DErr is a shorthand debug printing function to output to Stderr.
// Will only print if Debug is true.
func DErr(format string, a ...interface{}) {
	if Debug {
		PErr(format, a...)
	}
}

// DOut is a shorthand debug printing function to output to Stdout.
// Will only print if Debug is true.
func DOut(format string, a ...interface{}) {
	if Debug {
		POut(format, a...)
	}
}
