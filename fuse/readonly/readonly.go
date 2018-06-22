// A Go mirror of libfuse's hello.c
// +build linux darwin freebsd

package readonly

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"fmt"
	core "../../core"
	mdag "../../merkledag"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

type FileSystem struct {
	Ipfs *core.IpfsNode
}

func NewFileSystem(ipfs *core.IpfsNode) *FileSystem {
	return &FileSystem{Ipfs: ipfs}
}

func (f FileSystem) Root() (fs.Node, fuse.Error) {
	return &Root{Ipfs: f.Ipfs}, nil
}

type Root struct {
	Ipfs *core.IpfsNode
}

func (*Root) Attr() fuse.Attr {
	return fuse.Attr{Mode: os.ModeDir | 0111} // -rw+x
}

func (s *Root) Lookup(name string, intr fs.Intr) (fs.Node, fuse.Error) {

	switch name {
	case "mach_kernel", ".hidden", "._.":
		// Just quiet some log noise on OS X.
		return nil, fuse.ENOENT
	}

	nd, err := s.Ipfs.Resolver.ResolvePath(name)
	if err != nil {
		// todo: make this error more versatile.
		return nil, fuse.ENOENT
	}

	return &Node{Ipfs: s.Ipfs, Nd: nd}, nil
}

func (*Root) ReadDir(intr fs.Intr) ([]fuse.Dirent, fuse.Error) {
	return nil, fuse.EPERM
}

type Node struct {
	Ipfs *core.IpfsNode
	Nd   *mdag.Node
}

func (s *Node) Attr() fuse.Attr {
	if len(s.Nd.Links) > 0 {
		return fuse.Attr{Mode: os.ModeDir | 0555}
	}
	size, _ := s.Nd.Size()
	return fuse.Attr{Mode: 0444, Size: uint64(size)}
}

func (s *Node) Lookup(name string, intr fs.Intr) (fs.Node, fuse.Error) {

	nd, err := s.Ipfs.Resolver.ResolveLinks(s.Nd, []string{name})
	if err != nil {
		// todo: make this error more versatile.
		return nil, fuse.ENOENT
	}

	return &Node{Ipfs: s.Ipfs, Nd: nd}, nil
}

func (s *Node) ReadDir(intr fs.Intr) ([]fuse.Dirent, fuse.Error) {

	entries := make([]fuse.Dirent, len(s.Nd.Links))
	for i, link := range s.Nd.Links {
		n := link.Name
		if len(n) == 0 {
			n = link.Hash.B58String()
		}
		entries[i] = fuse.Dirent{Name: n, Type: fuse.DT_File}
	}

	if len(entries) > 0 {
		return entries, nil
	}
	return nil, fuse.ENOENT
}

func (s *Node) ReadAll(intr fs.Intr) ([]byte, fuse.Error) {
	return []byte(s.Nd.Data), nil
}

func Mount(ipfs *core.IpfsNode, fpath string) error {

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT,
		syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigc
		Unmount(fpath)
	}()

	c, err := fuse.Mount(fpath)
	if err != nil {
		return err
	}
	defer c.Close()

	err = fs.Serve(c, FileSystem{Ipfs: ipfs})
	if err != nil {
		return err
	}

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		return err
	}
	return nil
}

// Unmount attempts to unmount the provided FUSE mount point, forcibly
// if necessary.
func Unmount(point string) error {
	fmt.Printf("Unmounting %s...\n", point)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("diskutil", "umount", "force", point)
	case "linux":
		cmd = exec.Command("fusermount", "-u", point)
	default:
		return fmt.Errorf("unmount: unimplemented")
	}

	errc := make(chan error, 1)
	go func() {
		if err := exec.Command("umount", point).Run(); err == nil {
			errc <- err
		}
		// retry to unmount with the fallback cmd
		errc <- cmd.Run()
	}()

	select {
	case <-time.After(1 * time.Second):
		return fmt.Errorf("umount timeout")
	case err := <-errc:
		return err
	}
}
