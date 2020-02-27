package main

import (
	"os"
	"strings"
)

func pathExists(path string) bool {
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false
	}

	return true
}

func isFile(path string) bool {
	if !pathExists(path) {
		return false
	}

	info, _ := os.Stat(path)
	return !info.IsDir()
}

func isDir(path string) bool {
	if !pathExists(path) {
		return false
	}

	info, _ := os.Stat(path)
	return info.IsDir()
}

func alignCenter(s string, fullSize int) string {
	strSize := len(s)
	if strSize >= fullSize {
		return s
	}

	emptySize := fullSize - strSize
	div := emptySize / 2
	return strings.Repeat(" ", div) + s + strings.Repeat(" ", emptySize-div)
}

// selected will passed (N+1)
func isValidIndex(selected int, arrayLen int) bool {
	if selected < 1 || selected > arrayLen {
		return false
	}
	return true
}
