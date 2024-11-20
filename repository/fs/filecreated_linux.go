//go:build linux

package fs

import (
	stdfs "io/fs"
	"syscall"
	"time"
)

func fileCreated(info stdfs.FileInfo) time.Time {
	created := info.ModTime()

	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		created = time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec)
	}

	return created
}
