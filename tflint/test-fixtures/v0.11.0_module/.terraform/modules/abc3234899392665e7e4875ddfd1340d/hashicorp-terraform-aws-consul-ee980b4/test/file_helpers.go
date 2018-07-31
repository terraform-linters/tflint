package test

import (
	"os"
	"path/filepath"
	"testing"
	"io/ioutil"
	"strings"
)

// Copy the files in the given path to a temp folder and return the path to that temp folder. We do this so
// we can run tests in parallel on the same Terraform code without their state files overwriting each other.
func copyRepoToTempFolder(t *testing.T, path string) string {
	tmpPath, err := ioutil.TempDir("", "terraform-aws-consul-test")
	if err != nil {
		t.Fatalf("Failed to create temp folder due to error: %v", err)
	}

	copyFolderContents(t, path, tmpPath)
	return tmpPath
}

// Copy the files and folders within the source folder into the destination folder. Note that this method skips hidden
// files and folders (those that have names starting with a dot).
func copyFolderContents(t *testing.T, source string, destination string) {
	files, err := ioutil.ReadDir(source)
	if err != nil {
		t.Fatalf("Unable to read source folder %s due to error: %v", source, err)
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			// Don't copy any hidden files and folders, such as a local .terraform folder
			continue
		}

		src := filepath.Join(source, file.Name())
		dest := filepath.Join(destination, file.Name())

		if file.IsDir() {
			if err := os.MkdirAll(dest, file.Mode()); err != nil {
				t.Fatalf("Unable to create folder %s due to error: %v", dest, err)
			}

			copyFolderContents(t, src, dest)
		} else {
			copyFile(t, src, dest)
		}
	}
}

// Copy a file from source to destination
func copyFile(t *testing.T, source string, destination string) {
	contents, err := ioutil.ReadFile(source)
	if err != nil {
		t.Fatalf("Failed to read file %s due to error: %v", source, err)
	}

	writeFileWithSamePermissions(t, source, destination, contents)
}

// Write a file to the given destination with the given contents using the same permissions as the file at source
func writeFileWithSamePermissions(t *testing.T, source string, destination string, contents []byte) {
	fileInfo, err := os.Stat(source)
	if err != nil {
		t.Fatalf("Failed to stat file %s due to error: %v", source, err)
	}

	err = ioutil.WriteFile(destination, contents, fileInfo.Mode())
	if err != nil {
		t.Fatalf("Failed to write file to %s due to error: %v", destination, err)
	}
}
