package fs

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/pkg/errors"
)

var (
	errSrcNotDir = errors.New("source is not a directory")
)

// SyncDir recursively copies a directory tree, but tries to do nothing when not
// required.
func SyncDir(src, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	// We use os.Lstat() here to ensure we don't fall in a loop where a symlink
	// actually links to a one of its parent directories.
	fi, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return errSrcNotDir
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if err = os.MkdirAll(dst, fi.Mode()); err != nil {
		return errors.Wrapf(err, "cannot mkdir %s", dst)
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return errors.Wrapf(err, "cannot read directory %s", dst)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err = SyncDir(srcPath, dstPath); err != nil {
				return errors.Wrap(err, "copying directory failed")
			}
		} else {
			// This will include symlinks, which is what we want when
			// copying things.
			if err = syncFile(srcPath, dstPath); err != nil {
				return errors.Wrap(err, "copying file failed")
			}
		}
	}

	// Remove files in dst that aren't in src
	entries, err = ioutil.ReadDir(dst)
	if err != nil {
		return errors.Wrapf(err, "cannot read directory %s", dst)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			if err := os.RemoveAll(dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// syncFile copies the contents of the file named src to the file named by dst.
// The file will be created if it does not already exist. If the destination
// file exists, the file will be updated only where it differs. The file mode
// will be copied from the source.
func syncFile(src, dst string) error {
	if sym, err := IsSymlink(src); err != nil {
		return errors.Wrap(err, "symlink check failed")
	} else if sym {
		if err := cloneSymlink(src, dst); err != nil {
			if runtime.GOOS == "windows" {
				// If cloning the symlink fails on Windows because the user
				// does not have the required privileges, ignore the error and
				// fall back to copying the file contents.
				//
				// ERROR_PRIVILEGE_NOT_HELD is 1314 (0x522):
				// https://msdn.microsoft.com/en-us/library/windows/desktop/ms681385(v=vs.85).aspx
				if lerr, ok := err.(*os.LinkError); ok && lerr.Err != syscall.Errno(1314) {
					return err
				}
			} else {
				return err
			}
		} else {
			return nil
		}
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	inS, err := in.Stat()
	if err != nil {
		return err
	}
	outS, err := out.Stat()
	if err != nil {
		return err
	}

	var diff int
	// byte by byte, diff the file
	i := int64(0)
	for ; i < inS.Size() && i < outS.Size(); i++ {
		inCtt, outCtt := make([]byte, 1), make([]byte, 1)
		if _, err = in.ReadAt(inCtt, i); err != nil {
			return err
		}
		if _, err = out.ReadAt(outCtt, i); err != nil {
			return err
		}

		if diff = bytes.Compare(inCtt, outCtt); diff != 0 {
			break
		}
	}
	if i == inS.Size() {
		// We have reached the end of the src file, truncating the dst file in
		// case.
		if err := out.Truncate(i); err != nil {
			return err
		}
	} else if i < inS.Size() {
		// Files differ from i, lets copy src onto dst from i
		if _, err := in.Seek(i, 0); err != nil {
			return err
		}
		if _, err := out.Seek(i, 0); err != nil {
			return err
		}

		if _, err := io.Copy(out, in); err != nil {
			return err
		}
	}

	// Check for write errors on Close
	if err = out.Close(); err != nil {
		return err
	}

	si, err := os.Stat(src)
	if err != nil {
		return err
	}

	err = os.Chmod(dst, si.Mode())

	return err
}

// IsSymlink determines if the given path is a symbolic link.
func IsSymlink(path string) (bool, error) {
	l, err := os.Lstat(path)
	if err != nil {
		return false, err
	}

	return l.Mode()&os.ModeSymlink == os.ModeSymlink, nil
}

// cloneSymlink will create a new symlink that points to the resolved path of sl.
// If sl is a relative symlink, dst will also be a relative symlink.
func cloneSymlink(sl, dst string) error {
	resolved, err := os.Readlink(sl)
	if err != nil {
		return err
	}

	return os.Symlink(resolved, dst)
}
