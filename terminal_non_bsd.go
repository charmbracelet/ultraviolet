//go:build darwin || linux || solaris || aix || zos
// +build darwin linux solaris aix zos

package uv

import "golang.org/x/sys/unix"

func supportsBackspace(lflag uint64) bool {
	return lflag&unix.BSDLY == unix.BS0
}
