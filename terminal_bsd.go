//go:build dragonfly || freebsd || netbsd || openbsd
// +build dragonfly freebsd netbsd openbsd

package uv

func supportsBackspace(uint64) bool {
	return false
}
