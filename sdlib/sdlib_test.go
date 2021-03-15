package sdlib

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGetSdFolder(t *testing.T) {
	containingDir := t.TempDir()
	testFileName := filepath.Join(containingDir, "test.txt")

	testFile, _ := os.Create(testFileName)
	defer testFile.Close()

	fmt.Fprintf(testFile, "test\n")

	var sdDir = filepath.Join(containingDir, ".stay-deleted")

	var fetchedDir, _ = GetSdFolder(testFileName)

	if fetchedDir != sdDir {
		t.Error(`GetSdFolder(containingDir) != sdDir`)
	}
}
