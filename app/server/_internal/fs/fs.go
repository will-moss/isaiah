package fs

import "io/fs"

// Alias of fs.Sub, without returning any error
func Sub(fsys fs.FS, dir string) fs.FS {
	v, _ := fs.Sub(fsys, dir)
	return v
}
