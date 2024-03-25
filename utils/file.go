package fileUtils

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const ROOT = "storage-root"

func ReplaceSpecialChars(s string) string {
	var result strings.Builder
	for _, r := range s {
			if !isASCII(r) {
					// Replace special characters
					result.WriteByte(byte(r))
			} else {
					result.WriteRune(r)
			}
	}
	return result.String()
}

func isASCII(r rune) bool {
	return r <= unicode.MaxASCII
}

// FolderStructure represents a folder structure
type FolderStructure struct {
	Name     string            `json:"name"`
	Path     string            `json:"path"`
	IsFile   bool              `json:"isFile"`
	Children []*FolderStructure `json:"children,omitempty"`
}

func FileIsFile(file string) (bool, error) {
	info, err := os.Stat(file)
	if os.IsNotExist(err) {
		return false, err
	}
	return !info.IsDir(), nil
}

func DirIsDir(dir string) (bool, error) {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false, err
	}
	
	return info.IsDir(), nil
}

func Move(oldPath string, newPath string) error {
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return err
	}
	return nil
}

func MkdirAll(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}
	return nil
}

func Remove(path string, force bool) error {
	var err error

	if force {
		err = os.RemoveAll(path)
	} else {
		err = os.Remove(path)
	}

	if err != nil {
		return err
	}

	return nil
}

func Rename(oldPath string, newPath string) error {
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return err
	}
	return nil
}

func GetFolderStructure(path string) (*FolderStructure, error) {
	formatPath := func(path string) string {
		pathWithReplacedBackslashes := strings.ReplaceAll(path, "\\", "/")
		pathWithoutRoot := strings.ReplaceAll(pathWithReplacedBackslashes, ROOT, "")

		return pathWithoutRoot
	}
	
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	root := &FolderStructure{
		Name:     filepath.Base(path),
		Path:     formatPath(path),
		IsFile:   !info.IsDir(),
		Children: []*FolderStructure{},
	}

	err = filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(path, subpath)
		if err != nil {
			return err
		}

		// Skip the root folder itself
		if relativePath == "." {
			return nil
		}

		segments := strings.Split(relativePath, string(os.PathSeparator))
		currentNode := root

		for _, segment := range segments {
			found := false

			for _, child := range currentNode.Children {
				if child.Name == segment {
					currentNode = child
					found = true
					break
				}
			}

			if !found {
				child := &FolderStructure{
					Name:     segment,
					Path:     formatPath(subpath),
					IsFile:   false,
					Children: []*FolderStructure{},
				}

				currentNode.Children = append(currentNode.Children, child)
				currentNode = child
			}
		}

		currentNode.IsFile = !info.IsDir()

		return nil
	})

	if err != nil {
		return nil, err
	}

	return root, nil
}