//go:build darwin

package fs

import (
	stdfs "io/fs"
	"syscall"
	"time"
)

func fileCreated(info stdfs.FileInfo) time.Time {
	created := info.ModTime()

	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		created = time.Unix(stat.Ctimespec.Sec, stat.Ctimespec.Nsec)
	}

	return created
}
