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

func TestSetGetAction(t *testing.T) {
	dir := t.TempDir()

	tfn := "test.txt"
	tfp := filepath.Join(dir, tfn)

	tf, _ := os.Create(tfp)
	defer tf.Close()

	action := "keep"
	err := SetActionForFile(tfp, action)
	if err != nil {
		t.Error(err)
	}

	sdfp, err := GetSdFile(tfp)
	if err != nil {
		t.Error(err)
	}

	gotAction, err := GetActionForFile(sdfp, dir, os.Stderr)
	if err != nil {
		t.Error(err)
	}

	if gotAction.File != tfp {
		t.Error(fmt.Sprintf("gotAction.File is '%s', expecting '%s'!", gotAction.File, tfp))
	}

	if gotAction.Action != action {
		t.Error(fmt.Sprintf("gotAction.Action: %s!", gotAction.Action))
	}
}
