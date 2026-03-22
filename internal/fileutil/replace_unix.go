//go:build !windows

package fileutil

import "os"

func replaceFile(src, dst string) error {
	return os.Rename(src, dst)
}
