package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

var (
	ErrCannotSaveFile            = errors.New("cannot save file")
	ErrCannotCreateTmxuConfigDir = errors.New("cannot create tmxu config dir")
	ErrCannotReadFile            = errors.New("cannot read file")
	ErrCannotDeleteFile          = errors.New("cannot delete file")
	ErrCannotCheckFile           = errors.New("cannot check file")
)

func ensureTmxuConfigDir(path string) error {
	err := os.MkdirAll(path, 0700)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotCreateTmxuConfigDir, err)
	}

	return nil
}

func tmxuConfigDirPath(path string) string {
	return filepath.Join(path, configDirName)
}

func readFile(root *os.Root, name string) ([]byte, error) {
	file, err := root.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("%w %s: %w", ErrCannotReadFile, name, err)
	}

	return file, nil
}

func hasFile(root *os.Root, name string) (bool, error) {
	_, err := root.Stat(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}

		return false, fmt.Errorf("%w %s: %w", ErrCannotCheckFile, name, err)
	}

	return true, nil
}

func saveToFile(root *os.Root, name string, data []byte) (string, error) {
	tmp := name + ".tmp"

	f, err := root.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return "", fmt.Errorf("%w %s: %w", ErrCannotSaveFile, name, err)
	}

	_, err = f.Write(data)
	if err != nil {
		return saveError(err, name, root, tmp, f)
	}

	err = f.Sync()
	if err != nil {
		return saveError(err, name, root, tmp, f)
	}

	err = f.Close()
	if err != nil {
		return saveError(err, name, root, tmp, nil)
	}

	bak := name + ".bak"

	err = root.Remove(bak)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return saveError(err, name, root, tmp, nil)
	}

	err = root.Link(name, bak)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return saveError(err, name, root, tmp, nil)
	}

	err = root.Rename(tmp, name)
	if err != nil {
		return saveError(err, name, root, tmp, nil)
	}

	err = syncDir(root)
	if err != nil {
		return "", fmt.Errorf("%w %s: %w", ErrCannotSaveFile, name, err)
	}

	return filepath.Join(root.Name(), name), nil
}

func syncDir(root *os.Root) error {
	d, err := root.Open(".")
	if err != nil {
		return err
	}

	syncErr := d.Sync()
	closeErr := d.Close()

	return errors.Join(syncErr, closeErr)
}

func deleteFile(root *os.Root, name string) (string, error) {
	err := root.Remove(name)
	if err != nil {
		return "", fmt.Errorf("%w %s: %w", ErrCannotDeleteFile, name, err)
	}

	return filepath.Join(root.Name(), name), nil
}

func saveError(originalErr error, name string, root *os.Root, tmp string, f *os.File) (string, error) {
	cleanUpErr := cleanupTmp(root, tmp, f)

	return "", fmt.Errorf("%w %s: %w", ErrCannotSaveFile, name, errors.Join(originalErr, cleanUpErr))
}

func cleanupTmp(root *os.Root, tmp string, f *os.File) error {
	var errs []error

	if f != nil {
		err := f.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	_, err := deleteFile(root, tmp)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
