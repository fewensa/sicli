package hastrt

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// MigrateLegacyConfig copies Hast state from the aicli layout to sicli when the
// destination is still empty and the legacy tree has content.
func MigrateLegacyConfig() error {
	manager := ContextManager()
	selection := manager.DefaultSelection()
	if err := migrateTreeIfNeeded(legacyPath(selection.ConfigDir), selection.ConfigDir); err != nil {
		return err
	}
	return migrateTreeIfNeeded(legacyPath(selection.CacheDir), selection.CacheDir)
}

func legacyPath(modern string) string {
	if modern == "" {
		return ""
	}
	sep := string(os.PathSeparator)
	from := sep + "sicli" + sep
	to := sep + "aicli" + sep
	if !strings.Contains(modern, from) {
		return ""
	}
	return strings.Replace(modern, from, to, 1)
}

func migrateTreeIfNeeded(src, dst string) error {
	if src == "" || dst == "" || src == dst {
		return nil
	}
	srcInfo, err := os.Lstat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat legacy HAST state %s: %w", src, err)
	}
	if srcInfo.Mode()&os.ModeSymlink != 0 || !srcInfo.IsDir() {
		return fmt.Errorf("legacy HAST state must be a directory: %s", src)
	}
	if !legacyTreeHasContent(src) {
		return nil
	}
	if destinationConfigured(dst) {
		return nil
	}
	if err := copyTree(src, dst); err != nil {
		return fmt.Errorf("migrate legacy HAST state from %s to %s: %w", src, dst, err)
	}
	return nil
}

func legacyTreeHasContent(root string) bool {
	entries, err := os.ReadDir(root)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		name := entry.Name()
		path := filepath.Join(root, name)
		if entry.IsDir() {
			if legacyTreeHasContent(path) {
				return true
			}
			continue
		}
		if strings.HasPrefix(name, ".") {
			continue
		}
		return true
	}
	return false
}

func destinationConfigured(root string) bool {
	path := filepath.Join(root, configFileName)
	info, err := os.Lstat(path)
	if err != nil {
		return legacyTreeHasContent(root)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return true
	}
	file, err := loadFileConfig(path, false)
	if err != nil {
		return true
	}
	if strings.TrimSpace(file.BaseURL) != "" {
		return true
	}
	if file.Auth != nil && strings.TrimSpace(file.Auth.APIKey) != "" {
		return true
	}
	return false
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, dirPerm)
		}
		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, dirPerm)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing to migrate symlink %s", path)
		}
		if err := os.MkdirAll(filepath.Dir(target), dirPerm); err != nil {
			return err
		}
		return copyRegularFile(path, target, entry)
	})
}

func copyRegularFile(src, dst string, entry os.DirEntry) error {
	info, err := entry.Info()
	if err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
