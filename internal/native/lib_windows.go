//go:build windows

package native

import "golang.org/x/sys/windows"

func openLibrary(name string) (uintptr, error) {
	handle, err := windows.LoadLibrary(name)
	return uintptr(handle), err
}
