package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetWorkDir() string {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		//
		return ""
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return ""
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return ""
	}
	return string(path[:i+1])
}
