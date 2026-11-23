package plugin

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"strings"
)

// hashSchemeH1 is the prefix for the sha256 hash of a plugin binary.
// Note that it is not the same as Terraform's "h1:" (dirhash).
const hashSchemeH1 = "h1:"

var (
	// platformPattern matches platform names like "linux_amd64".
	platformPattern = regexp.MustCompile(`^[a-z0-9]+_[a-z0-9]+$`)
	// h1HashPattern matches "h1:" hashes.
	h1HashPattern = regexp.MustCompile(`^h1:[0-9a-f]{64}$`)
)

// currentPlatform returns the platform name of the running TFLint, e.g. linux_amd64.
func currentPlatform() string {
	return fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
}

// Checksummer handles hashes of plugins.
// It validates checksums based on checksums.txt, and generates "h1:" hashes
// of plugin binaries recorded in the lockfile.
type Checksummer struct {
	checksums map[string][]byte
}

// NewChecksummer returns a new Checksummer from passed checksums.txt file.
// The checksums.txt must contain multiple lines containing sha256 hashes and filenames separated by spaces.
// An example is shown below:
//
// 3a61fff3689f27c89bce22893219919c629d2e10b96e7eadd5fef9f0e90bb353  tflint-ruleset-aws_darwin_amd64.zip
// 482419fdeed00692304e59558b5b0d915d4727868b88a5adbbbb76f5ed1b537a  tflint-ruleset-aws_linux_amd64.zip
// db4eed4c0abcfb0b851da5bbfe8d0c71e1c2b6afe4fd627638a462c655045902  tflint-ruleset-aws_windows_amd64.zip
func NewChecksummer(f io.Reader) (*Checksummer, error) {
	scanner := bufio.NewScanner(f)

	var line int
	checksummer := &Checksummer{checksums: map[string][]byte{}}
	for scanner.Scan() {
		line++
		fields := strings.Fields(scanner.Text())
		// checksums file should have "hash" and "filename" fields
		if len(fields) != 2 {
			return nil, fmt.Errorf("record on line %d: wrong number of fields: expected=2, actual=%d", line, len(fields))
		}
		hash := fields[0]
		filename := fields[1]

		checksum, err := hex.DecodeString(hash)
		if err != nil {
			return nil, err
		}
		checksummer.checksums[filename] = checksum
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return checksummer, nil
}

// binaryHashes returns "h1:" hashes of plugin binaries keyed by platform.
// Checksum entries for binaries must be named like tflint-ruleset-{name}_{GOOS}_{GOARCH}.
// Note that the binary for Windows does not have the ".exe" extension in the entry.
func (c *Checksummer) binaryHashes(pluginName string) map[string]string {
	prefix := fmt.Sprintf("tflint-ruleset-%s_", pluginName)

	hashes := map[string]string{}
	for filename, checksum := range c.checksums {
		platform, found := strings.CutPrefix(filename, prefix)
		if !found || !platformPattern.MatchString(platform) {
			continue
		}
		hashes[platform] = hashSchemeH1 + hex.EncodeToString(checksum)
	}
	return hashes
}

// HashFile returns the "h1:" hash of the given file.
func (c *Checksummer) HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	sum, err := sha256Sum(f)
	if err != nil {
		return "", err
	}
	return hashSchemeH1 + hex.EncodeToString(sum), nil
}

// Verify calculates the sha256 hash of the passed file and compares it to the expected hash value based on the filename.
func (c *Checksummer) Verify(filename string, f io.Reader) error {
	actual, err := sha256Sum(f)
	if err != nil {
		return err
	}

	expected := c.checksums[filename]
	if !bytes.Equal(actual, expected) {
		return fmt.Errorf("Failed to match checksums: expected=%x, actual=%x", expected, actual)
	}

	return nil
}

func sha256Sum(r io.Reader) ([]byte, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
