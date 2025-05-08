//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris && !aix && !zos && !windows
// +build !darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris,!aix,!zos,!windows

package tv

func (*Terminal) makeRaw() error {
	return ErrPlatformNotSupported
}

func (t *Terminal) optimizeMovements() {}

func (*Terminal) enableWindowsMouse() error  { return ErrPlatformNotSupported }
func (*Terminal) disableWindowsMouse() error { return ErrPlatformNotSupported }
