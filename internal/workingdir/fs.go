package workingdir

import (
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// Resolve preserves absolute paths and resolves relative paths against dir.
func Resolve(dir, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(dir, path)
}

// FS reads relative paths against Dir without changing the process directory.
// Absolute paths and parent references retain their normal filesystem meaning.
type FS struct {
	afero.Fs
	Dir string
}

func New(fs afero.Fs, dir string) *FS {
	return &FS{Fs: afero.NewReadOnlyFs(fs), Dir: dir}
}

func (fs *FS) Open(name string) (afero.File, error) {
	f, err := fs.Fs.Open(Resolve(fs.Dir, name))
	if err != nil {
		return nil, err
	}
	return namedFile{File: f, name: name}, nil
}

func (fs *FS) Stat(name string) (os.FileInfo, error) {
	return fs.Fs.Stat(Resolve(fs.Dir, name))
}

func (fs *FS) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if flag&(os.O_WRONLY|os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC) != 0 {
		return nil, os.ErrPermission
	}
	return fs.Open(name)
}

type namedFile struct {
	afero.File
	name string
}

func (f namedFile) Name() string { return f.name }
