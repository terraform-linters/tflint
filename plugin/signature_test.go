package plugin

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terraform-linters/tflint/tflint"
)

func Test_GetSigningKey(t *testing.T) {
	cases := []struct {
		Name     string
		Config   *InstallConfig
		Expected string
	}{
		{
			Name:     "no signing key",
			Config:   NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SigningKey: ""}),
			Expected: "",
		},
		{
			Name:     "configured singing key",
			Config:   NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SigningKey: testSigningKey}),
			Expected: testSigningKey,
		},
		{
			Name:     "bulit-in signing key",
			Config:   NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SigningKey: "", SourceOwner: "terraform-linters"}),
			Expected: builtinSigningKey,
		},
		{
			Name:     "bulit-in signing key and configured signing key",
			Config:   NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SigningKey: testSigningKey, SourceOwner: "terraform-linters"}),
			Expected: testSigningKey,
		},
	}

	for _, tc := range cases {
		sigchecker := NewSignatureChecker(tc.Config)

		got := sigchecker.GetSigningKey()
		if got != tc.Expected {
			t.Fatalf("Failed `%s`: expected=%s, got=%s", tc.Name, tc.Expected, got)
		}
	}
}

func Test_HasSigningKey(t *testing.T) {
	cases := []struct {
		Name     string
		Config   *InstallConfig
		Expected bool
	}{
		{
			Name:     "no signing key",
			Config:   NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SigningKey: ""}),
			Expected: false,
		},
		{
			Name:     "configured singing key",
			Config:   NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SigningKey: testSigningKey}),
			Expected: true,
		},
		{
			Name:     "bulit-in signing key",
			Config:   NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SigningKey: "", SourceOwner: "terraform-linters"}),
			Expected: true,
		},
		{
			Name:     "bulit-in signing key and configured signing key",
			Config:   NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SigningKey: testSigningKey, SourceOwner: "terraform-linters"}),
			Expected: true,
		},
	}

	for _, tc := range cases {
		sigchecker := NewSignatureChecker(tc.Config)

		got := sigchecker.HasSigningKey()
		if got != tc.Expected {
			t.Fatalf("Failed `%s`: expected=%t, got=%t", tc.Name, tc.Expected, got)
		}
	}
}

func Test_SignatureChecker_Verify(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	signature, err := os.Open(filepath.Join(cwd, "test-fixtures", "signatures", "checksums.txt.sig"))
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}
	defer signature.Close()

	target := `003432556f0380963e6802e701d40a5ea303cfd8ad0cd34719b00c796abfe90e  tflint-ruleset-aws_netbsd_amd64.zip
02caedcd1f0e9c331862b0babce83792d95021977595be9712b01dc0164c488d  tflint-ruleset-aws_netbsd_arm.zip
03ed11c1deeb4dbfc1656b030e4172d1f6b03a1257f219f6265bfc100b20c7a8  tflint-ruleset-aws_netbsd_386.zip
043416079a7ea9e0f7888915278aecda7d6268b61066deacc50feefb6e57836c  tflint-ruleset-aws_linux_arm.zip
30fb5e2d8d8ca3be7247cb0f9e644e3accf0b334d473ce6a5b83151302183e22  tflint-ruleset-aws_openbsd_amd64.zip
3a61fff3689f27c89bce22893219919c629d2e10b96e7eadd5fef9f0e90bb353  tflint-ruleset-aws_darwin_amd64.zip
482419fdeed00692304e59558b5b0d915d4727868b88a5adbbbb76f5ed1b537a  tflint-ruleset-aws_linux_amd64.zip
7147440b689291870ffafd40652a9a623df6c43557da3d2c90e712065e0d093f  tflint-ruleset-aws_freebsd_386.zip
7c7154183f3faf4c80c6703969f5f4903f795eebb881b06da40caa4827eb48e4  tflint-ruleset-aws_windows_386.zip
9df843eb85785246df1e24886b1112b324444dd02329ff45c5cffa597f674b6c  tflint-ruleset-aws_openbsd_arm.zip
b6231a2b94a71841409bb28be0f22eb6ed9de744f0c06ca3a9dd6f80cbc63956  tflint-ruleset-aws_linux_386.zip
bd804df0ff957cda8210c74017211face850cd62445a7d0102493ab5cfdc4276  tflint-ruleset-aws_freebsd_amd64.zip
bee3591d764729769fd32dae7dc8147a890aadc45c00260fa6dafe2df1585a02  tflint-ruleset-aws_openbsd_386.zip
db4eed4c0abcfb0b851da5bbfe8d0c71e1c2b6afe4fd627638a462c655045902  tflint-ruleset-aws_windows_amd64.zip
dd536fed0ebe4c1115240574c5dd7a31b563d67bfe0d1111750438718f995d43  tflint-ruleset-aws_freebsd_arm.zip
`
	reader := strings.NewReader(target)

	sigchecker := NewSignatureChecker(NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SigningKey: builtinSigningKey}))
	if err := sigchecker.Verify(reader, signature); err != nil {
		t.Fatalf("Verify failed: %s", err)
	}
}

func Test_SignatureChecker_Verify_errors(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	signature, err := os.Open(filepath.Join(cwd, "test-fixtures", "signatures", "checksums.txt.sig"))
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}
	defer signature.Close()
	brokenSignature := strings.NewReader("broken")

	target := `003432556f0380963e6802e701d40a5ea303cfd8ad0cd34719b00c796abfe90e  tflint-ruleset-aws_netbsd_amd64.zip
02caedcd1f0e9c331862b0babce83792d95021977595be9712b01dc0164c488d  tflint-ruleset-aws_netbsd_arm.zip
03ed11c1deeb4dbfc1656b030e4172d1f6b03a1257f219f6265bfc100b20c7a8  tflint-ruleset-aws_netbsd_386.zip
043416079a7ea9e0f7888915278aecda7d6268b61066deacc50feefb6e57836c  tflint-ruleset-aws_linux_arm.zip
30fb5e2d8d8ca3be7247cb0f9e644e3accf0b334d473ce6a5b83151302183e22  tflint-ruleset-aws_openbsd_amd64.zip
3a61fff3689f27c89bce22893219919c629d2e10b96e7eadd5fef9f0e90bb353  tflint-ruleset-aws_darwin_amd64.zip
482419fdeed00692304e59558b5b0d915d4727868b88a5adbbbb76f5ed1b537a  tflint-ruleset-aws_linux_amd64.zip
7147440b689291870ffafd40652a9a623df6c43557da3d2c90e712065e0d093f  tflint-ruleset-aws_freebsd_386.zip
7c7154183f3faf4c80c6703969f5f4903f795eebb881b06da40caa4827eb48e4  tflint-ruleset-aws_windows_386.zip
9df843eb85785246df1e24886b1112b324444dd02329ff45c5cffa597f674b6c  tflint-ruleset-aws_openbsd_arm.zip
b6231a2b94a71841409bb28be0f22eb6ed9de744f0c06ca3a9dd6f80cbc63956  tflint-ruleset-aws_linux_386.zip
bd804df0ff957cda8210c74017211face850cd62445a7d0102493ab5cfdc4276  tflint-ruleset-aws_freebsd_amd64.zip
bee3591d764729769fd32dae7dc8147a890aadc45c00260fa6dafe2df1585a02  tflint-ruleset-aws_openbsd_386.zip
db4eed4c0abcfb0b851da5bbfe8d0c71e1c2b6afe4fd627638a462c655045902  tflint-ruleset-aws_windows_amd64.zip
dd536fed0ebe4c1115240574c5dd7a31b563d67bfe0d1111750438718f995d43  tflint-ruleset-aws_freebsd_arm.zip
`

	cases := []struct {
		Name      string
		Config    *InstallConfig
		Target    string
		Signature io.Reader
		Expected  error
	}{
		{
			Name:      "invalid signature",
			Config:    NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SigningKey: builtinSigningKey}),
			Target:    "broken",
			Signature: signature,
			Expected:  fmt.Errorf("openpgp: invalid signature: hash tag doesn't match"),
		},
		{
			Name:      "no signing key",
			Config:    NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SigningKey: ""}),
			Target:    target,
			Signature: signature,
			Expected:  fmt.Errorf("No signing key configured"),
		},
		{
			Name:      "broken signing key",
			Config:    NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SigningKey: "broken"}),
			Target:    target,
			Signature: signature,
			Expected:  fmt.Errorf("openpgp: invalid argument: no armored data found"),
		},
		{
			Name:      "broken signature",
			Config:    NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SigningKey: builtinSigningKey}),
			Target:    target,
			Signature: brokenSignature,
			Expected:  fmt.Errorf("openpgp: invalid data: tag byte does not have MSB set"),
		},
	}

	for _, tc := range cases {
		sigchecker := NewSignatureChecker(tc.Config)
		reader := strings.NewReader(tc.Target)

		err := sigchecker.Verify(reader, tc.Signature)
		if err == nil {
			t.Fatalf("Failed `%s`: expected=%s, actual=no errors", tc.Name, tc.Expected)
		}
		if err.Error() != tc.Expected.Error() {
			t.Fatalf("Failed `%s`: expected=%s, actual=%s", tc.Name, tc.Expected, err)
		}
	}
}

var testSigningKey = `
-----BEGIN PGP PUBLIC KEY BLOCK-----

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
ZF5q4h4I33PSGDdSvGXn9UMY5Isjpg==
=7pIB
-----END PGP PUBLIC KEY BLOCK-----`
