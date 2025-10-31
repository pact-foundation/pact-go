//go:build darwin || linux

package native

import "github.com/ebitengine/purego"

func openLibrary(name string) (uintptr, error) {
	return purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
}
