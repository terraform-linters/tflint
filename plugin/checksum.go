package plugin

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

// Checksummer validates checksums
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
//
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

// Verify calculates the sha256 hash of the passed file and compares it to the expected hash value based on the filename.
func (c *Checksummer) Verify(filename string, f io.Reader) error {
	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return err
	}

	expected := c.checksums[filename]
	actual := hash.Sum(nil)
	if !bytes.Equal(actual, expected) {
		return fmt.Errorf("Failed to match checksums: expected=%x, actual=%x", expected, actual)
	}

	return nil
}
