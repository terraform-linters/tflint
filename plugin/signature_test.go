package plugin

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-github/v81/github"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_GetSigningKey(t *testing.T) {
	cases := []struct {
		Name     string
		Config   *InstallConfig
		Envs     map[string]string
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
			Name:     "built-in signing key",
			Config:   NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SigningKey: "", SourceOwner: "terraform-linters"}),
			Expected: builtinSigningKey,
		},
		{
			Name:     "built-in signing key and configured signing key",
			Config:   NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SigningKey: testSigningKey, SourceOwner: "terraform-linters"}),
			Expected: testSigningKey,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			for k, v := range tc.Envs {
				t.Setenv(k, v)
			}

			sigchecker := NewSignatureChecker(tc.Config)

			got := sigchecker.GetSigningKey()
			if got != tc.Expected {
				t.Errorf("expected=%s, got=%s", tc.Expected, got)
			}
		})
	}
}

func Test_HasSigningKey(t *testing.T) {
	cases := []struct {
		Name     string
		Config   *InstallConfig
		Envs     map[string]string
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
			Name:     "built-in signing key",
			Config:   NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SigningKey: "", SourceOwner: "terraform-linters"}),
			Expected: true,
		},
		{
			Name:     "built-in signing key and configured signing key",
			Config:   NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SigningKey: testSigningKey, SourceOwner: "terraform-linters"}),
			Expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			for k, v := range tc.Envs {
				t.Setenv(k, v)
			}

			sigchecker := NewSignatureChecker(tc.Config)

			got := sigchecker.HasSigningKey()
			if got != tc.Expected {
				t.Errorf("expected=%t, got=%t", tc.Expected, got)
			}
		})
	}
}

func Test_SignatureChecker_VerifyPGPSignature(t *testing.T) {
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
	if err := sigchecker.VerifyPGPSignature(reader, signature); err != nil {
		t.Fatalf("Verify failed: %s", err)
	}
}

func Test_SignatureChecker_VerifyPGPSignature_errors(t *testing.T) {
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
		t.Run(tc.Name, func(t *testing.T) {
			sigchecker := NewSignatureChecker(tc.Config)
			reader := strings.NewReader(tc.Target)

			err := sigchecker.VerifyPGPSignature(reader, tc.Signature)
			if err == nil {
				t.Fatalf("expected=%s, actual=no errors", tc.Expected)
			}
			if err.Error() != tc.Expected.Error() {
				t.Errorf("expected=%s, actual=%s", tc.Expected, err)
			}
		})
	}
}

func Test_SignatureChecker_VerifyAttestations(t *testing.T) {
	// checksums.txt for tflint-ruleset-aws v0.35.0
	target := `57847831c681fcd3817945d3e4cb0ca8a72f571aa1ea91f0d0f9f19c98bf2b9f  tflint-ruleset-aws_darwin_amd64
11575e9dff6d19a91848c42f216b83d0eef788f6efd3ec07fe2dae936bade71c  tflint-ruleset-aws_darwin_amd64.zip
da3b90d2cfb91fdafeeec53e637db68691c3ac5874593c03da129121da117c3e  tflint-ruleset-aws_darwin_arm64
c156963d710e2b76be9002cc7e7eb8500928866d6622561f9d10b04d06e64985  tflint-ruleset-aws_darwin_arm64.zip
8b8b088d2d58f8735ad007d3d1240e06277966245a3ea9c0d7d81ad9f9445318  tflint-ruleset-aws_linux_386
b69f538b26a7e92f0100692d6e603eb5657172d7546b6e18888ff6f4d27f733c  tflint-ruleset-aws_linux_386.zip
c2a2e33d838cb908a393daf3d0f456fd185f997cb747980f0bf0209e5da17bd5  tflint-ruleset-aws_linux_amd64
45e409f5ce71f163f38b716a89baca3ae19d771b53e5adb4ac57120a1b714a8d  tflint-ruleset-aws_linux_amd64.zip
b59dc4cbc7883ed638ac3862540a1835662b924c327156df1cf3cf9808874d5c  tflint-ruleset-aws_linux_arm
abc4761c93fcecffd2eb273fed17596afc8f1e160553652315926b839e7246bd  tflint-ruleset-aws_linux_arm.zip
d3e80a663ef6c5ebb09b62ce9931de73edc8693d4d9d91943c15985b070122e9  tflint-ruleset-aws_linux_arm64
a5cca22160e381bbfc069358a5a229559e917a32c4d3ca9746b5218dff63e173  tflint-ruleset-aws_linux_arm64.zip
0ed17cd7e837f64e6b3708b6fcf2a3bd25b7d5d5051c3ae5da74c8a7530599e7  tflint-ruleset-aws_windows_386
d14baf6119a904a0340fd84352d6a0917cfc9bbecdff7d6981a4dcece4275d1c  tflint-ruleset-aws_windows_386.zip
475b5e6e6c569856e673195e0ce7ec81b48f9eb4b4962a02e2a969a9e7666bbb  tflint-ruleset-aws_windows_amd64
b97e20eae04a45d650886611f17020fd0aa29114b86268b71e3841195fbc55ca  tflint-ruleset-aws_windows_amd64.zip
`
	reader := strings.NewReader(target)

	attestations := []*github.Attestation{
		{
			RepositoryID: 245765716,
			Bundle:       []byte(testSigstoreBundle034), // sigstore bundle for v0.34.0 (mismatched)
		},
		{
			RepositoryID: 245765716,
			Bundle:       []byte(testSigstoreBundle035), // sigstore bundle for v0.35.0 (matched)
		},
	}

	// The first mismatched bundle is ignored without errors
	sigchecker := NewSignatureChecker(NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SourceHost: "github.com", SourceOwner: "terraform-linters", SourceRepo: "tflint-ruleset-aws"}))
	if err := sigchecker.VerifyAttestations(reader, attestations); err != nil {
		t.Fatalf("Verify failed: %s", err)
	}
}

func Test_SignatureChecker_VerifyAttestations_errors(t *testing.T) {
	// checksums.txt for tflint-ruleset-aws v0.35.0
	target := `57847831c681fcd3817945d3e4cb0ca8a72f571aa1ea91f0d0f9f19c98bf2b9f  tflint-ruleset-aws_darwin_amd64
11575e9dff6d19a91848c42f216b83d0eef788f6efd3ec07fe2dae936bade71c  tflint-ruleset-aws_darwin_amd64.zip
da3b90d2cfb91fdafeeec53e637db68691c3ac5874593c03da129121da117c3e  tflint-ruleset-aws_darwin_arm64
c156963d710e2b76be9002cc7e7eb8500928866d6622561f9d10b04d06e64985  tflint-ruleset-aws_darwin_arm64.zip
8b8b088d2d58f8735ad007d3d1240e06277966245a3ea9c0d7d81ad9f9445318  tflint-ruleset-aws_linux_386
b69f538b26a7e92f0100692d6e603eb5657172d7546b6e18888ff6f4d27f733c  tflint-ruleset-aws_linux_386.zip
c2a2e33d838cb908a393daf3d0f456fd185f997cb747980f0bf0209e5da17bd5  tflint-ruleset-aws_linux_amd64
45e409f5ce71f163f38b716a89baca3ae19d771b53e5adb4ac57120a1b714a8d  tflint-ruleset-aws_linux_amd64.zip
b59dc4cbc7883ed638ac3862540a1835662b924c327156df1cf3cf9808874d5c  tflint-ruleset-aws_linux_arm
abc4761c93fcecffd2eb273fed17596afc8f1e160553652315926b839e7246bd  tflint-ruleset-aws_linux_arm.zip
d3e80a663ef6c5ebb09b62ce9931de73edc8693d4d9d91943c15985b070122e9  tflint-ruleset-aws_linux_arm64
a5cca22160e381bbfc069358a5a229559e917a32c4d3ca9746b5218dff63e173  tflint-ruleset-aws_linux_arm64.zip
0ed17cd7e837f64e6b3708b6fcf2a3bd25b7d5d5051c3ae5da74c8a7530599e7  tflint-ruleset-aws_windows_386
d14baf6119a904a0340fd84352d6a0917cfc9bbecdff7d6981a4dcece4275d1c  tflint-ruleset-aws_windows_386.zip
475b5e6e6c569856e673195e0ce7ec81b48f9eb4b4962a02e2a969a9e7666bbb  tflint-ruleset-aws_windows_amd64
b97e20eae04a45d650886611f17020fd0aa29114b86268b71e3841195fbc55ca  tflint-ruleset-aws_windows_amd64.zip
`

	cases := []struct {
		Name         string
		Config       *InstallConfig
		Attestations []*github.Attestation
		Expected     error
	}{
		{
			Name:         "no attestations",
			Config:       NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SourceHost: "github.com", SourceOwner: "terraform-linters", SourceRepo: "tflint-ruleset-aws"}),
			Attestations: []*github.Attestation{},
			Expected:     fmt.Errorf("no attestations found"),
		},
		{
			Name:   "mismatched attestations",
			Config: NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SourceHost: "github.com", SourceOwner: "terraform-linters", SourceRepo: "tflint-ruleset-aws"}),
			Attestations: []*github.Attestation{
				{
					RepositoryID: 245765716,
					Bundle:       []byte(testSigstoreBundle034), // sigstore bundle for v0.34.0 (mismatched)
				},
			},
			Expected: fmt.Errorf(`failed to verify signature: provided artifact digest does not match any digest in statement`),
		},
		{
			Name:   "invalid identity issuer",
			Config: NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SourceHost: "github.example.com", SourceOwner: "terraform-linters", SourceRepo: "tflint-ruleset-aws"}),
			Attestations: []*github.Attestation{
				{
					RepositoryID: 245765716,
					Bundle:       []byte(testSigstoreBundle035), // sigstore bundle for v0.35.0 (matched)
				},
			},
			Expected: fmt.Errorf(`failed to verify certificate identity: no matching CertificateIdentity found, last error: expected SAN value to match regex "^https://github\.example\.com/terraform-linters/tflint-ruleset-aws/", got "https://github.com/terraform-linters/tflint-ruleset-aws/.github/workflows/release.yml@refs/tags/v0.35.0"`),
		},
		{
			Name:   "invalid identity SAN",
			Config: NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SourceHost: "github.com", SourceOwner: "terraform-linters-malformed", SourceRepo: "tflint-ruleset-aws"}),
			Attestations: []*github.Attestation{
				{
					RepositoryID: 245765716,
					Bundle:       []byte(testSigstoreBundle035), // sigstore bundle for v0.35.0 (matched)
				},
			},
			Expected: fmt.Errorf(`failed to verify certificate identity: no matching CertificateIdentity found, last error: expected SAN value to match regex "^https://github\.com/terraform-linters-malformed/tflint-ruleset-aws/", got "https://github.com/terraform-linters/tflint-ruleset-aws/.github/workflows/release.yml@refs/tags/v0.35.0"`),
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			sigchecker := NewSignatureChecker(tc.Config)
			reader := strings.NewReader(target)

			err := sigchecker.VerifyAttestations(reader, tc.Attestations)
			if err == nil {
				t.Fatalf("expected=%s, actual=no errors", tc.Expected)
			}
			if err.Error() != tc.Expected.Error() {
				t.Errorf("expected=%s, actual=%s", tc.Expected, err)
			}
		})
	}
}

var testSigningKey string = `
-----BEGIN PGP PUBLIC KEY BLOCK-----

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
ZF5q4h4I33PSGDdSvGXn9UMY5Isjpg==
=7pIB
-----END PGP PUBLIC KEY BLOCK-----`

// sigstore bundle for tflint-ruleset-aws v0.34.0
var testSigstoreBundle034 string = `{
  "mediaType": "application/vnd.dev.sigstore.bundle.v0.3+json",
  "verificationMaterial": {
    "tlogEntries": [
      {
        "logIndex": "139993835",
        "logId": {
          "keyId": "wNI9atQGlz+VWfO6LRygH4QUfY/8W4RFwiT5i5WRgB0="
        },
        "kindVersion": {
          "kind": "dsse",
          "version": "0.0.1"
        },
        "integratedTime": "1728923609",
        "inclusionPromise": {
          "signedEntryTimestamp": "MEUCIDt97uy6QimQocGsisjV28RDxktASdgz2WnwBnLeCoFMAiEA481ZtPrfDoa/WG7pE/h8Zyfa1Ba8Zj0lgPR84FJ5DRo="
        },
        "inclusionProof": {
          "logIndex": "18089573",
          "rootHash": "AHulDP6i5WnSAT+1P+OcUm/6H5mtBfRbnDIwYhBmXyI=",
          "treeSize": "18089576",
          "hashes": [
            "/K1ULbLY7g4R+s6j2/WTWfqo265JPKiQtdDJlr4bq6s=",
            "hgN5+QJdMOUGL5qssmyGs0MnibCri7KKW6JAakWBgcs=",
            "m0GIdi2NTtR+A1cxU8gP9ZDBDqGO6EymxBeek9DPkJ8=",
            "RH24reV6fKWcgzREkINGX5T7obNL0lzkYyUUBhUzZTI=",
            "GKgBr410qYdeRRBx5D/ohlvhvEKjYmwabMLz+H2tx4I=",
            "bWixxvJK+JHptI++3mo8e4g2KF75A8/SXz9Z2YKm9f8=",
            "AhKoWHIpIJ0HilsPmQKeSSJDv5EG9sxfalTpHxb78ss=",
            "7Wh91TtcDOIQLD/Q0l/LBWMTDDefwk/ZRgbmjcOEK5c=",
            "rAAfgmPXo8TJp1LmkVDhAYrf0WzE4X4/mDuW1pwVM3Y=",
            "gf+9m552B3PnkWnO0o4KdVvjcT3WVHLrCbf1DoVYKFw="
          ],
          "checkpoint": {
            "envelope": "rekor.sigstore.dev - 1193050959916656506\n18089576\nAHulDP6i5WnSAT+1P+OcUm/6H5mtBfRbnDIwYhBmXyI=\n\n— rekor.sigstore.dev wNI9ajBEAiBTh6P2VD7mx70IMv44muiLkChWgVvBjIck1phMA+5z0wIgULI4+ywdrNL/fiC7DtulKLkFVOm9mTpOlUPJVLxt1vk=\n"
          }
        },
        "canonicalizedBody": "eyJhcGlWZXJzaW9uIjoiMC4wLjEiLCJraW5kIjoiZHNzZSIsInNwZWMiOnsiZW52ZWxvcGVIYXNoIjp7ImFsZ29yaXRobSI6InNoYTI1NiIsInZhbHVlIjoiYTMzN2I0ZDNmNGFiYTIzMjZiMTI4MTE4YjUyMWMyMjkxYjhjNjg0NTRjMmNiMGE2MzliMzdhNjgzY2IzOTc1MyJ9LCJwYXlsb2FkSGFzaCI6eyJhbGdvcml0aG0iOiJzaGEyNTYiLCJ2YWx1ZSI6IjBjZTEwMGQ5NWE1NDdmNTBmN2Q1NDMyN2ZjN2U0ZTMwMDVkY2RkYzA2YmM1NDZmOWU5NmFlZjkzMGM1YjlhMTUifSwic2lnbmF0dXJlcyI6W3sic2lnbmF0dXJlIjoiTUVVQ0lRRFpqcnJtd3ZLbGFDYThwUXM5ajhhbjV4R2pBcVg5REYwMi84cUt4SmdCMEFJZ0pMN0oycHR6Uy9nUndsd2dONW9ic1hRaUJXWG5RMmJjSmdCVUkrU2Foams9IiwidmVyaWZpZXIiOiJMUzB0TFMxQ1JVZEpUaUJEUlZKVVNVWkpRMEZVUlMwdExTMHRDazFKU1VoTGFrTkRRbkVyWjBGM1NVSkJaMGxWVFdKb1NVeGpTRFZxVlRSaGFHaDZURFJsWjNaamRVOXlXREF3ZDBObldVbExiMXBKZW1vd1JVRjNUWGNLVG5wRlZrMUNUVWRCTVZWRlEyaE5UV015Ykc1ak0xSjJZMjFWZFZwSFZqSk5ValIzU0VGWlJGWlJVVVJGZUZaNllWZGtlbVJIT1hsYVV6RndZbTVTYkFwamJURnNXa2RzYUdSSFZYZElhR05PVFdwUmVFMUVSVEJOVkZsNlRYcEpORmRvWTA1TmFsRjRUVVJGTUUxVVdUQk5la2swVjJwQlFVMUdhM2RGZDFsSUNrdHZXa2w2YWpCRFFWRlpTVXR2V2tsNmFqQkVRVkZqUkZGblFVVXhjWFI1Y1ZsRFVUWmtjRWc1TTJsR1IyWXdVRXh1YkVsRmVWQmxNMjg0VUV0NFdYVUtka00yUW1oSlZtcGpUbTFRUXpSQ1QwY3JWMjEzVUhGc1pWZE1jMjFRYUN0SGVHTTVkamxUUkhwM2IzVTFUVGh2Vm5GUFEwSmpOSGRuWjFoTFRVRTBSd3BCTVZWa1JIZEZRaTkzVVVWQmQwbElaMFJCVkVKblRsWklVMVZGUkVSQlMwSm5aM0pDWjBWR1FsRmpSRUY2UVdSQ1owNVdTRkUwUlVablVWVkVXbmxtQ2tvd04wOXFiMmxvV21GcFRWZG5Zelo0TWtscE5YaG5kMGgzV1VSV1VqQnFRa0puZDBadlFWVXpPVkJ3ZWpGWmEwVmFZalZ4VG1wd1MwWlhhWGhwTkZrS1drUTRkMlJSV1VSV1VqQlNRVkZJTDBKSGMzZGhXVnB1WVVoU01HTklUVFpNZVRsdVlWaFNiMlJYU1hWWk1qbDBURE5TYkdOdVNtaGFiVGw1WWxNeGN3cGhWelV3V2xoS2Vrd3pVbTFpUjJ4MVpFTXhlV1JYZUd4ak1sWXdURmRHTTJONU9IVmFNbXd3WVVoV2FVd3paSFpqYlhSdFlrYzVNMk41T1hsYVYzaHNDbGxZVG14TWJteDBZa1ZDZVZwWFducE1NMUpvV2pOTmRtUnFRWFZOZWxGMVRVUkJOVUpuYjNKQ1owVkZRVmxQTDAxQlJVSkNRM1J2WkVoU2QyTjZiM1lLVEROU2RtRXlWblZNYlVacVpFZHNkbUp1VFhWYU1td3dZVWhXYVdSWVRteGpiVTUyWW01U2JHSnVVWFZaTWpsMFRVSkpSME5wYzBkQlVWRkNaemM0ZHdwQlVVbEZRa2hDTVdNeVozZE9aMWxMUzNkWlFrSkJSMFIyZWtGQ1FYZFJiMDFFVlhoUFIwa3lUMGRSTVU5RVdUSlpiVlpwVFhwcmVWcHFXVEJPYlZwb0NrOVVaekZaTWtVMVdrUlJlRTVVUm14T2FscG9UbnBCVmtKbmIzSkNaMFZGUVZsUEwwMUJSVVZDUVdSNVdsZDRiRmxZVG14TlJFbEhRMmx6UjBGUlVVSUtaemM0ZDBGUlZVVktTRkpzWTI1S2FGcHRPWGxpVXpGellWYzFNRnBZU25wTU0xSnRZa2RzZFdSRE1YbGtWM2hzWXpKV01FeFhSak5qZWtGbVFtZHZjZ3BDWjBWRlFWbFBMMDFCUlVkQ1FrWjVXbGRhZWt3elVtaGFNMDEyWkdwQmRVMTZVWFZOUkVFM1FtZHZja0puUlVWQldVOHZUVUZGU1VKRE1FMUxNbWd3Q21SSVFucFBhVGgyWkVjNWNscFhOSFZaVjA0d1lWYzVkV041Tlc1aFdGSnZaRmRLTVdNeVZubFpNamwxWkVkV2RXUkROV3BpTWpCM1pIZFpTMHQzV1VJS1FrRkhSSFo2UVVKRFVWSndSRWRrYjJSSVVuZGplbTkyVERKa2NHUkhhREZaYVRWcVlqSXdkbVJIVm5samJVWnRZak5LZEV4WGVIQmlibEpzWTI1TmRncGtSMXB6WVZjMU1FeFlTakZpUjFaNldsaFJkRmxZWkhwTWVUVnVZVmhTYjJSWFNYWmtNamw1WVRKYWMySXpaSHBNTTBwc1lrZFdhR015VlhWbFZ6RnpDbEZJU214YWJrMTJaRWRHYm1ONU9USk5RelI2VGtNMGQwMUVaMGREYVhOSFFWRlJRbWMzT0hkQlVXOUZTMmQzYjAxRVZYaFBSMGt5VDBkUk1VOUVXVElLV1cxV2FVMTZhM2xhYWxrd1RtMWFhRTlVWnpGWk1rVTFXa1JSZUU1VVJteE9hbHBvVG5wQlpFSm5iM0pDWjBWRlFWbFBMMDFCUlV4Q1FUaE5SRmRrY0Fwa1IyZ3hXV2t4YjJJelRqQmFWMUYzVW5kWlMwdDNXVUpDUVVkRWRucEJRa1JCVVRWRVJHUnZaRWhTZDJONmIzWk1NbVJ3WkVkb01WbHBOV3BpTWpCMkNtUkhWbmxqYlVadFlqTktkRXhYZUhCaWJsSnNZMjVOZG1SSFduTmhWelV3VEZoS01XSkhWbnBhV0ZGMFdWaGtlazFFWjBkRGFYTkhRVkZSUW1jM09IY0tRVkV3UlV0bmQyOU5SRlY0VDBkSk1rOUhVVEZQUkZreVdXMVdhVTE2YTNsYWFsa3dUbTFhYUU5VVp6RlpNa1UxV2tSUmVFNVVSbXhPYWxwb1RucEJhQXBDWjI5eVFtZEZSVUZaVHk5TlFVVlBRa0pOVFVWWVNteGFiazEyWkVkR2JtTjVPVEpOUXpSNlRrTTBkMDFDYTBkRGFYTkhRVkZSUW1jM09IZEJVVGhGQ2tOM2QwcE5hbEV4VG5wWk1VNTZSVEpOUkZGSFEybHpSMEZSVVVKbk56aDNRVkpCUlVwbmQydGhTRkl3WTBoTk5reDVPVzVoV0ZKdlpGZEpkVmt5T1hRS1RETlNiR051U21oYWJUbDVZbE14YzJGWE5UQmFXRXA2VFVKblIwTnBjMGRCVVZGQ1p6YzRkMEZTUlVWRFozZEpUbFJSZUU5VVl6Uk9WRUYzWkhkWlN3cExkMWxDUWtGSFJIWjZRVUpGWjFKd1JFZGtiMlJJVW5kamVtOTJUREprY0dSSGFERlphVFZxWWpJd2RtUkhWbmxqYlVadFlqTktkRXhYZUhCaWJsSnNDbU51VFhaa1IxcHpZVmMxTUV4WVNqRmlSMVo2V2xoUmRGbFlaSHBNZVRWdVlWaFNiMlJYU1haa01qbDVZVEphYzJJelpIcE1NMHBzWWtkV2FHTXlWWFVLWlZjeGMxRklTbXhhYmsxMlpFZEdibU41T1RKTlF6UjZUa00wZDAxRVowZERhWE5IUVZGUlFtYzNPSGRCVWsxRlMyZDNiMDFFVlhoUFIwa3lUMGRSTVFwUFJGa3lXVzFXYVUxNmEzbGFhbGt3VG0xYWFFOVVaekZaTWtVMVdrUlJlRTVVUm14T2FscG9UbnBCVlVKbmIzSkNaMFZGUVZsUEwwMUJSVlZDUVZsTkNrSklRakZqTW1kM1lYZFpTMHQzV1VKQ1FVZEVkbnBCUWtaUlVtUkVSblJ2WkVoU2QyTjZiM1pNTW1Sd1pFZG9NVmxwTldwaU1qQjJaRWRXZVdOdFJtMEtZak5LZEV4WGVIQmlibEpzWTI1TmRtUkhXbk5oVnpVd1RGaEtNV0pIVm5wYVdGRjBXVmhrZWt3eVJtcGtSMngyWW01TmRtTnVWblZqZVRoNFRWUk5lZ3BOVkZFd1RXcG5lRTFET1doa1NGSnNZbGhDTUdONU9IaE5RbGxIUTJselIwRlJVVUpuTnpoM1FWSlpSVU5CZDBkalNGWnBZa2RzYWsxSlIwdENaMjl5Q2tKblJVVkJaRm8xUVdkUlEwSklkMFZsWjBJMFFVaFpRVE5VTUhkaGMySklSVlJLYWtkU05HTnRWMk16UVhGS1MxaHlhbVZRU3pNdmFEUndlV2RET0hBS04yODBRVUZCUjFOcEswTnVRM2RCUVVKQlRVRlNla0pHUVdsRlFUSm1XWEozWWt4RU4xaFlSelJIYkV0VWJHZExUVkpuU0dSTFUzSkdaazVTZWxSdlRRbzBSbTB5VVdGSlEwbEJaWEJVVDJoVmVqRlZXRTV1TkVRMk9EVmFTV3BrU2pSU01GcExWR00xVHpWWFRURkRhbkpYYzNKeVRVRnZSME5EY1VkVFRUUTVDa0pCVFVSQk1tdEJUVWRaUTAxUlJEZDBORTl0WTFwWlEyMW5NSEJKTW14dWNIaHZTRmRPWTFkSlMyUnhha2N2V25GMVdIcG5PVVJQUTI1aFVGcGxVa29LVDJ0a2JFWnVVRUZRU2pOaFZIaDNRMDFSUkZOaU9XTmtLMUJqUVdkbFZWTk9PVFZaWm01dlYxRlpaakY2VnpKVmEwcEliRUUzUVhNNVprbzBNMWR0YUFwSGNWaHdkRXBzZEZadlZtVkRWRkJKVVV0SlBRb3RMUzB0TFVWT1JDQkRSVkpVU1VaSlEwRlVSUzB0TFMwdENnPT0ifV19fQ=="
      }
    ],
    "timestampVerificationData": {
    },
    "certificate": {
      "rawBytes": "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"
    }
  },
  "dsseEnvelope": {
    "payload": "eyJfdHlwZSI6Imh0dHBzOi8vaW4tdG90by5pby9TdGF0ZW1lbnQvdjEiLCJzdWJqZWN0IjpbeyJuYW1lIjoiY2hlY2tzdW1zLnR4dCIsImRpZ2VzdCI6eyJzaGEyNTYiOiJkMzFkNDliZmUyZjBmNjc4ZDI0OGUxZWVkYzAzNjQ2ZjQyYzA2ZmMxM2M4NWVhMzEzOGQyNjZlOTIwYWJjZjY2In19XSwicHJlZGljYXRlVHlwZSI6Imh0dHBzOi8vc2xzYS5kZXYvcHJvdmVuYW5jZS92MSIsInByZWRpY2F0ZSI6eyJidWlsZERlZmluaXRpb24iOnsiYnVpbGRUeXBlIjoiaHR0cHM6Ly9hY3Rpb25zLmdpdGh1Yi5pby9idWlsZHR5cGVzL3dvcmtmbG93L3YxIiwiZXh0ZXJuYWxQYXJhbWV0ZXJzIjp7IndvcmtmbG93Ijp7InJlZiI6InJlZnMvdGFncy92MC4zNC4wIiwicmVwb3NpdG9yeSI6Imh0dHBzOi8vZ2l0aHViLmNvbS90ZXJyYWZvcm0tbGludGVycy90ZmxpbnQtcnVsZXNldC1hd3MiLCJwYXRoIjoiLmdpdGh1Yi93b3JrZmxvd3MvcmVsZWFzZS55bWwifX0sImludGVybmFsUGFyYW1ldGVycyI6eyJnaXRodWIiOnsiZXZlbnRfbmFtZSI6InB1c2giLCJyZXBvc2l0b3J5X2lkIjoiMjQ1NzY1NzE2IiwicmVwb3NpdG9yeV9vd25lcl9pZCI6IjU0MTk3ODUwIiwicnVubmVyX2Vudmlyb25tZW50IjoiZ2l0aHViLWhvc3RlZCJ9fSwicmVzb2x2ZWREZXBlbmRlbmNpZXMiOlt7InVyaSI6ImdpdCtodHRwczovL2dpdGh1Yi5jb20vdGVycmFmb3JtLWxpbnRlcnMvdGZsaW50LXJ1bGVzZXQtYXdzQHJlZnMvdGFncy92MC4zNC4wIiwiZGlnZXN0Ijp7ImdpdENvbW1pdCI6IjA1MThiNjhkNTg2NmJlYjM5MmY2NDZmYTk4NWNhOWQ0MTUxZTY2YTcifX1dfSwicnVuRGV0YWlscyI6eyJidWlsZGVyIjp7ImlkIjoiaHR0cHM6Ly9naXRodWIuY29tL3RlcnJhZm9ybS1saW50ZXJzL3RmbGludC1ydWxlc2V0LWF3cy8uZ2l0aHViL3dvcmtmbG93cy9yZWxlYXNlLnltbEByZWZzL3RhZ3MvdjAuMzQuMCJ9LCJtZXRhZGF0YSI6eyJpbnZvY2F0aW9uSWQiOiJodHRwczovL2dpdGh1Yi5jb20vdGVycmFmb3JtLWxpbnRlcnMvdGZsaW50LXJ1bGVzZXQtYXdzL2FjdGlvbnMvcnVucy8xMTMzMTQ0MjgxMC9hdHRlbXB0cy8xIn19fX0=",
    "payloadType": "application/vnd.in-toto+json",
    "signatures": [
      {
        "sig": "MEUCIQDZjrrmwvKlaCa8pQs9j8an5xGjAqX9DF02/8qKxJgB0AIgJL7J2ptzS/gRwlwgN5obsXQiBWXnQ2bcJgBUI+Sahjk="
      }
    ]
  }
}`

// sigstore bundle for tflint-ruleset-aws v0.35.0
var testSigstoreBundle035 string = `{
  "mediaType": "application/vnd.dev.sigstore.bundle.v0.3+json",
  "verificationMaterial": {
    "tlogEntries": [
      {
        "logIndex": "149329939",
        "logId": {
          "keyId": "wNI9atQGlz+VWfO6LRygH4QUfY/8W4RFwiT5i5WRgB0="
        },
        "kindVersion": {
          "kind": "dsse",
          "version": "0.0.1"
        },
        "integratedTime": "1731835307",
        "inclusionPromise": {
          "signedEntryTimestamp": "MEYCIQCrvIj/L+4Wjvh/rYr+QIJl7mfKGkOO7jE7ifPYuA8fRgIhAJJIBRWISMEOSS9ecShkkJfJtORpFLzKhAGobaKmeIRQ"
        },
        "inclusionProof": {
          "logIndex": "27425677",
          "rootHash": "nUaTZsbLPICWzJo57PNhz/fCrYxS99xfOp21OhUeDWo=",
          "treeSize": "27425678",
          "hashes": [
            "bljRkivBVGunbGbjjvuDEjTlQ6yHxWYIZI+kABKzLQM=",
            "sfGRa6EMAzULIRUobf1CYHvSwN2F+Oi5POY1s6gyvQE=",
            "h9fJYidGkGKHfHbCkZ19bZM8aeLfjzzu1xLTAwQCK4Y=",
            "RPxyvyvtPZaNEZ1SGfTA5jnClld84kshxctPuQAc9HU=",
            "0xpBX8D1FxB3jGFWcP44QeJ1i+3onFgj7pRe6RJPPdk=",
            "RGBlI7EA3a8lXH+EeiKdiPHid3xIgBDmgf70U6/JPhk=",
            "twlY0GMAe1WGbFsmvenvcVDRhCYSWL8BzlFaVZS1kIo=",
            "1uWLSTsQSxZvL3/3Fd0cx09O3G+tM34u2xiZ2ajxhEE=",
            "e9E4YrQeqXnsscNChrMoMgyaRdFogVkh0T0azIpcwyI=",
            "vH2a7kQ+SRIHTva7hHBoGu9AX70jls61uqRg/BprNAU=",
            "X+WKzna8ARHxD0HZdOLUPAMSYaEIIMMtWS7Hxkf6TJg=",
            "E2rLOYPJFKiizYiyu07QLqkMVTVL7i2ZgXiQywdI9KQ=",
            "4lUF0YOu9XkIDXKXA0wMSzd6VeDY3TZAgmoOeWmS2+Y=",
            "gf+9m552B3PnkWnO0o4KdVvjcT3WVHLrCbf1DoVYKFw="
          ],
          "checkpoint": {
            "envelope": "rekor.sigstore.dev - 1193050959916656506\n27425678\nnUaTZsbLPICWzJo57PNhz/fCrYxS99xfOp21OhUeDWo=\n\n— rekor.sigstore.dev wNI9ajBGAiEA5ndAgomrOduT43uVDDLygAf5VgBsEHqOA1u27kkmsxUCIQCXQUYwKhrLEnpvxHgsf1dV5D37m0CEcZiIQLqIqnpgcw==\n"
          }
        },
        "canonicalizedBody": "eyJhcGlWZXJzaW9uIjoiMC4wLjEiLCJraW5kIjoiZHNzZSIsInNwZWMiOnsiZW52ZWxvcGVIYXNoIjp7ImFsZ29yaXRobSI6InNoYTI1NiIsInZhbHVlIjoiNzIwZGVhYjVhYjRiNTkzODUyMDg4ZjdmMTc5ZDUwOTZiNTdjMWY3YjgyYjQ0NGFkZDI2N2RmOWI5NmRmNDgzOSJ9LCJwYXlsb2FkSGFzaCI6eyJhbGdvcml0aG0iOiJzaGEyNTYiLCJ2YWx1ZSI6IjNmMTUyNmYwNGFmMDk4ZTI4ZGZhYTZhMzM1MzczZTU3YjY3MmYxZjRlNDM4NTUxN2U2NmQ0NzkxNDJmMTJhNzIifSwic2lnbmF0dXJlcyI6W3sic2lnbmF0dXJlIjoiTUVVQ0lGMFBmM00vQzA1cnZJNG5TNHBRYUJXcWxtOGo1ZnoxZFp3akZBWldlUTZCQWlFQWk1NDJVTkJlUldvK3F3dlk4aTFOd2E0TTlmbzJMUlA5eXF1enZ3UlN0b009IiwidmVyaWZpZXIiOiJMUzB0TFMxQ1JVZEpUaUJEUlZKVVNVWkpRMEZVUlMwdExTMHRDazFKU1VoTFZFTkRRbkVyWjBGM1NVSkJaMGxWUkhGd2N5dEZNMHhsZFVaNE5UaE1kM1JPTkhOeWNsaDNVbEV3ZDBObldVbExiMXBKZW1vd1JVRjNUWGNLVG5wRlZrMUNUVWRCTVZWRlEyaE5UV015Ykc1ak0xSjJZMjFWZFZwSFZqSk5ValIzU0VGWlJGWlJVVVJGZUZaNllWZGtlbVJIT1hsYVV6RndZbTVTYkFwamJURnNXa2RzYUdSSFZYZElhR05PVFdwUmVFMVVSVE5OUkd0NVRWUlJNbGRvWTA1TmFsRjRUVlJGTTAxRWEzcE5WRkV5VjJwQlFVMUdhM2RGZDFsSUNrdHZXa2w2YWpCRFFWRlpTVXR2V2tsNmFqQkVRVkZqUkZGblFVVXJaR0p2VjNCb1VHSlFjRTVsY0VaWVVHRlVhVnBJZWpWc1oxVjJUVVl2ZUdKU1ZtRUtVWGg2TlZsT1lUUnRValJ2UzNGeVUyWlVVMHhMZFZaNGFEUnVUbkp3UlVzd2JrZzVNalEzYUhaNGJHdzNkMGcxTW1GUFEwSmpOSGRuWjFoTFRVRTBSd3BCTVZWa1JIZEZRaTkzVVVWQmQwbElaMFJCVkVKblRsWklVMVZGUkVSQlMwSm5aM0pDWjBWR1FsRmpSRUY2UVdSQ1owNVdTRkUwUlVablVWVk1NRFl3Q25SV04xRTVZamxRU0RoUGMxcGhhMFJSVW1KMVVWQTBkMGgzV1VSV1VqQnFRa0puZDBadlFWVXpPVkJ3ZWpGWmEwVmFZalZ4VG1wd1MwWlhhWGhwTkZrS1drUTRkMlJSV1VSV1VqQlNRVkZJTDBKSGMzZGhXVnB1WVVoU01HTklUVFpNZVRsdVlWaFNiMlJYU1hWWk1qbDBURE5TYkdOdVNtaGFiVGw1WWxNeGN3cGhWelV3V2xoS2Vrd3pVbTFpUjJ4MVpFTXhlV1JYZUd4ak1sWXdURmRHTTJONU9IVmFNbXd3WVVoV2FVd3paSFpqYlhSdFlrYzVNMk41T1hsYVYzaHNDbGxZVG14TWJteDBZa1ZDZVZwWFducE1NMUpvV2pOTmRtUnFRWFZOZWxWMVRVUkJOVUpuYjNKQ1owVkZRVmxQTDAxQlJVSkNRM1J2WkVoU2QyTjZiM1lLVEROU2RtRXlWblZNYlVacVpFZHNkbUp1VFhWYU1td3dZVWhXYVdSWVRteGpiVTUyWW01U2JHSnVVWFZaTWpsMFRVSkpSME5wYzBkQlVWRkNaemM0ZHdwQlVVbEZRa2hDTVdNeVozZE9aMWxMUzNkWlFrSkJSMFIyZWtGQ1FYZFJiMDVIVm14T01sWm9UMVJyTkZreVVYbE5SRlpvV20xT2JVMXFUbXhPUkVadENrOVVUVEZhVjBacVdtMU5lVnBFVlRGUFZGcHFXbFJCVmtKbmIzSkNaMFZGUVZsUEwwMUJSVVZDUVdSNVdsZDRiRmxZVG14TlJFbEhRMmx6UjBGUlVVSUtaemM0ZDBGUlZVVktTRkpzWTI1S2FGcHRPWGxpVXpGellWYzFNRnBZU25wTU0xSnRZa2RzZFdSRE1YbGtWM2hzWXpKV01FeFhSak5qZWtGbVFtZHZjZ3BDWjBWRlFWbFBMMDFCUlVkQ1FrWjVXbGRhZWt3elVtaGFNMDEyWkdwQmRVMTZWWFZOUkVFM1FtZHZja0puUlVWQldVOHZUVUZGU1VKRE1FMUxNbWd3Q21SSVFucFBhVGgyWkVjNWNscFhOSFZaVjA0d1lWYzVkV041Tlc1aFdGSnZaRmRLTVdNeVZubFpNamwxWkVkV2RXUkROV3BpTWpCM1pIZFpTMHQzV1VJS1FrRkhSSFo2UVVKRFVWSndSRWRrYjJSSVVuZGplbTkyVERKa2NHUkhhREZaYVRWcVlqSXdkbVJIVm5samJVWnRZak5LZEV4WGVIQmlibEpzWTI1TmRncGtSMXB6WVZjMU1FeFlTakZpUjFaNldsaFJkRmxZWkhwTWVUVnVZVmhTYjJSWFNYWmtNamw1WVRKYWMySXpaSHBNTTBwc1lrZFdhR015VlhWbFZ6RnpDbEZJU214YWJrMTJaRWRHYm1ONU9USk5RelI2VGxNMGQwMUVaMGREYVhOSFFWRlJRbWMzT0hkQlVXOUZTMmQzYjA1SFZteE9NbFpvVDFSck5Ga3lVWGtLVFVSV2FGcHRUbTFOYWs1c1RrUkdiVTlVVFRGYVYwWnFXbTFOZVZwRVZURlBWRnBxV2xSQlpFSm5iM0pDWjBWRlFWbFBMMDFCUlV4Q1FUaE5SRmRrY0Fwa1IyZ3hXV2t4YjJJelRqQmFWMUYzVW5kWlMwdDNXVUpDUVVkRWRucEJRa1JCVVRWRVJHUnZaRWhTZDJONmIzWk1NbVJ3WkVkb01WbHBOV3BpTWpCMkNtUkhWbmxqYlVadFlqTktkRXhYZUhCaWJsSnNZMjVOZG1SSFduTmhWelV3VEZoS01XSkhWbnBhV0ZGMFdWaGtlazFFWjBkRGFYTkhRVkZSUW1jM09IY0tRVkV3UlV0bmQyOU9SMVpzVGpKV2FFOVVhelJaTWxGNVRVUldhRnB0VG0xTmFrNXNUa1JHYlU5VVRURmFWMFpxV20xTmVWcEVWVEZQVkZwcVdsUkJhQXBDWjI5eVFtZEZSVUZaVHk5TlFVVlBRa0pOVFVWWVNteGFiazEyWkVkR2JtTjVPVEpOUXpSNlRsTTBkMDFDYTBkRGFYTkhRVkZSUW1jM09IZEJVVGhGQ2tOM2QwcE5hbEV4VG5wWk1VNTZSVEpOUkZGSFEybHpSMEZSVVVKbk56aDNRVkpCUlVwbmQydGhTRkl3WTBoTk5reDVPVzVoV0ZKdlpGZEpkVmt5T1hRS1RETlNiR051U21oYWJUbDVZbE14YzJGWE5UQmFXRXA2VFVKblIwTnBjMGRCVVZGQ1p6YzRkMEZTUlVWRFozZEpUbFJSZUU5VVl6Uk9WRUYzWkhkWlN3cExkMWxDUWtGSFJIWjZRVUpGWjFKd1JFZGtiMlJJVW5kamVtOTJUREprY0dSSGFERlphVFZxWWpJd2RtUkhWbmxqYlVadFlqTktkRXhYZUhCaWJsSnNDbU51VFhaa1IxcHpZVmMxTUV4WVNqRmlSMVo2V2xoUmRGbFlaSHBNZVRWdVlWaFNiMlJYU1haa01qbDVZVEphYzJJelpIcE1NMHBzWWtkV2FHTXlWWFVLWlZjeGMxRklTbXhhYmsxMlpFZEdibU41T1RKTlF6UjZUbE0wZDAxRVowZERhWE5IUVZGUlFtYzNPSGRCVWsxRlMyZDNiMDVIVm14T01sWm9UMVJyTkFwWk1sRjVUVVJXYUZwdFRtMU5hazVzVGtSR2JVOVVUVEZhVjBacVdtMU5lVnBFVlRGUFZGcHFXbFJCVlVKbmIzSkNaMFZGUVZsUEwwMUJSVlZDUVZsTkNrSklRakZqTW1kM1lYZFpTMHQzV1VKQ1FVZEVkbnBCUWtaUlVtUkVSblJ2WkVoU2QyTjZiM1pNTW1Sd1pFZG9NVmxwTldwaU1qQjJaRWRXZVdOdFJtMEtZak5LZEV4WGVIQmlibEpzWTI1TmRtUkhXbk5oVnpVd1RGaEtNV0pIVm5wYVdGRjBXVmhrZWt3eVJtcGtSMngyWW01TmRtTnVWblZqZVRoNFRWUm5Nd3BPZWxsM1RucE5ORTU1T1doa1NGSnNZbGhDTUdONU9IaE5RbGxIUTJselIwRlJVVUpuTnpoM1FWSlpSVU5CZDBkalNGWnBZa2RzYWsxSlIwdENaMjl5Q2tKblJVVkJaRm8xUVdkUlEwSklkMFZsWjBJMFFVaFpRVE5VTUhkaGMySklSVlJLYWtkU05HTnRWMk16UVhGS1MxaHlhbVZRU3pNdmFEUndlV2RET0hBS04yODBRVUZCUjFSUFZ6SnFjVkZCUVVKQlRVRlNla0pHUVdsQ2RtZERObGRJY0VrNFVtRnRlV1JtWkhKWE9UaFZiblF6WmxGMVpVTnJNekZ2VDFOQlFRcFllRFZ2ZEZGSmFFRkpWekprTTIxQlVXNXhkVFl5VDBkWVJYTnpOa3RaT0ZCd2R6SjJWRVZTUWxsa2NtRktkU3RqYldGS1RVRnZSME5EY1VkVFRUUTVDa0pCVFVSQk1tZEJUVWRWUTAxUlJIVTFjVFJQT1hFd1NETTBXRFU1TlVORmFHVkVjV1ZxVDNkSGJHUkVjVXhCYUM5WFRXY3ZOMGhVUVRkamVVUjBPRFFLUkhZMGJqQmpVeXMxV2poRWR6TlZRMDFFWm5KSWNHWlRNV1Z5UjBwUWMxTmFWR1pKVkc1aU9FRnNUMkpRZHpSSU0zRm1hR3MzZGxjNFUwZHFNaXROTlFwcE5VRlhXWGRHVkZaRGVqUjBNV0kzWW1jOVBRb3RMUzB0TFVWT1JDQkRSVkpVU1VaSlEwRlVSUzB0TFMwdENnPT0ifV19fQ=="
      }
    ],
    "timestampVerificationData": {
    },
    "certificate": {
      "rawBytes": "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"
    }
  },
  "dsseEnvelope": {
    "payload": "eyJfdHlwZSI6Imh0dHBzOi8vaW4tdG90by5pby9TdGF0ZW1lbnQvdjEiLCJzdWJqZWN0IjpbeyJuYW1lIjoiY2hlY2tzdW1zLnR4dCIsImRpZ2VzdCI6eyJzaGEyNTYiOiI2Y2UyYzIwNjUyNjJjZGQ0Njg2NjlkNmU4MzEyMTU0NTViYmFjNTY5Y2I5YjFjOGQyNzI3NWM1ZjQ0NDdkMTY5In19XSwicHJlZGljYXRlVHlwZSI6Imh0dHBzOi8vc2xzYS5kZXYvcHJvdmVuYW5jZS92MSIsInByZWRpY2F0ZSI6eyJidWlsZERlZmluaXRpb24iOnsiYnVpbGRUeXBlIjoiaHR0cHM6Ly9hY3Rpb25zLmdpdGh1Yi5pby9idWlsZHR5cGVzL3dvcmtmbG93L3YxIiwiZXh0ZXJuYWxQYXJhbWV0ZXJzIjp7IndvcmtmbG93Ijp7InJlZiI6InJlZnMvdGFncy92MC4zNS4wIiwicmVwb3NpdG9yeSI6Imh0dHBzOi8vZ2l0aHViLmNvbS90ZXJyYWZvcm0tbGludGVycy90ZmxpbnQtcnVsZXNldC1hd3MiLCJwYXRoIjoiLmdpdGh1Yi93b3JrZmxvd3MvcmVsZWFzZS55bWwifX0sImludGVybmFsUGFyYW1ldGVycyI6eyJnaXRodWIiOnsiZXZlbnRfbmFtZSI6InB1c2giLCJyZXBvc2l0b3J5X2lkIjoiMjQ1NzY1NzE2IiwicmVwb3NpdG9yeV9vd25lcl9pZCI6IjU0MTk3ODUwIiwicnVubmVyX2Vudmlyb25tZW50IjoiZ2l0aHViLWhvc3RlZCJ9fSwicmVzb2x2ZWREZXBlbmRlbmNpZXMiOlt7InVyaSI6ImdpdCtodHRwczovL2dpdGh1Yi5jb20vdGVycmFmb3JtLWxpbnRlcnMvdGZsaW50LXJ1bGVzZXQtYXdzQHJlZnMvdGFncy92MC4zNS4wIiwiZGlnZXN0Ijp7ImdpdENvbW1pdCI6IjRlZTdlYTk5OGNkMjA1YWZjZjIzZTQxZjkzNWVhY2ZjMmQ1NTk2Y2UifX1dfSwicnVuRGV0YWlscyI6eyJidWlsZGVyIjp7ImlkIjoiaHR0cHM6Ly9naXRodWIuY29tL3RlcnJhZm9ybS1saW50ZXJzL3RmbGludC1ydWxlc2V0LWF3cy8uZ2l0aHViL3dvcmtmbG93cy9yZWxlYXNlLnltbEByZWZzL3RhZ3MvdjAuMzUuMCJ9LCJtZXRhZGF0YSI6eyJpbnZvY2F0aW9uSWQiOiJodHRwczovL2dpdGh1Yi5jb20vdGVycmFmb3JtLWxpbnRlcnMvdGZsaW50LXJ1bGVzZXQtYXdzL2FjdGlvbnMvcnVucy8xMTg3NzYwNzM4Ny9hdHRlbXB0cy8xIn19fX0=",
    "payloadType": "application/vnd.in-toto+json",
    "signatures": [
      {
        "sig": "MEUCIF0Pf3M/C05rvI4nS4pQaBWqlm8j5fz1dZwjFAZWeQ6BAiEAi542UNBeRWo+qwvY8i1Nwa4M9fo2LRP9yquzvwRStoM="
      }
    ]
  }
}`
