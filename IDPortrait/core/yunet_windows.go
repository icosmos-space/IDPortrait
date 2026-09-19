package core

import "golang.org/x/sys/windows"

func prepareRuntimeDir(dir string) {
	_ = windows.SetDllDirectory(dir)
}
