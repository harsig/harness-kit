// Package yamlconfig loads YAML configuration files strictly and resolves the
// paths in them relative to the file.
package yamlconfig

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Load decodes the file at path into v (see Decode) and returns the file's
// absolute directory, the base for relative paths in it.
func Load(path string, v any) (dir string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if err := Decode(data, v); err != nil {
		return "", fmt.Errorf("%s: %w", path, err)
	}
	return Dir(path)
}

// Decode decodes data into v. Unknown keys are an error; empty data leaves v unchanged.
func Decode(data []byte, v any) error {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(v); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

// Dir returns the absolute directory of the file at path.
func Dir(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Dir(abs), nil
}

// ResolvePath expands a leading ~ to the home directory and makes p absolute
// relative to baseDir. Without a known home directory, ~ is kept as is.
func ResolvePath(baseDir, p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			p = filepath.Join(home, p[1:])
		}
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(baseDir, p)
	}
	return filepath.Clean(p)
}
