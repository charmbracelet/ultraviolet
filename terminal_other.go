//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris && !aix && !zos && !windows
// +build !darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris,!aix,!zos,!windows

package tv

import "fmt"

func (*Terminal) makeRaw() error {
	return fmt.Errorf("platform not supported")
}
