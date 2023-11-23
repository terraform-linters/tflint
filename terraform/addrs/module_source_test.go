// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package addrs

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseModuleSource(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    ModuleSource
		wantErr string
	}{
		// Local paths
		"local in subdirectory": {
			input: "./child",
			want:  ModuleSourceLocal("./child"),
		},
		"local in subdirectory non-normalized": {
			input: "./nope/../child",
			want:  ModuleSourceLocal("./child"),
		},
		"local in sibling directory": {
			input: "../sibling",
			want:  ModuleSourceLocal("../sibling"),
		},
		"local in sibling directory non-normalized": {
			input: "./nope/../../sibling",
			want:  ModuleSourceLocal("../sibling"),
		},
		"Windows-style local in subdirectory": {
			input: `.\child`,
			want:  ModuleSourceLocal("./child"),
		},
		"Windows-style local in subdirectory non-normalized": {
			input: `.\nope\..\child`,
			want:  ModuleSourceLocal("./child"),
		},
		"Windows-style local in sibling directory": {
			input: `..\sibling`,
			want:  ModuleSourceLocal("../sibling"),
		},
		"Windows-style local in sibling directory non-normalized": {
			input: `.\nope\..\..\sibling`,
			want:  ModuleSourceLocal("../sibling"),
		},
		"an abominable mix of different slashes": {
			input: `./nope\nope/why\./please\don't`,
			want:  ModuleSourceLocal("./nope/nope/why/please/don't"),
		},
		// Registry addresses
		"main registry implied": {
			input: "hashicorp/subnets/cidr",
			want:  ModuleSourceRemote("hashicorp/subnets/cidr"),
		},
		// Remote package addresses
		"github.com shorthand": {
			input: "github.com/hashicorp/terraform-cidr-subnets",
			want:  ModuleSourceRemote("github.com/hashicorp/terraform-cidr-subnets"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			addr, err := ParseModuleSource(test.input)

			if test.wantErr != "" {
				switch {
				case err == nil:
					t.Errorf("unexpected success\nwant error: %s", test.wantErr)
				case err.Error() != test.wantErr:
					t.Errorf("wrong error messages\ngot:  %s\nwant: %s", err.Error(), test.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			if diff := cmp.Diff(addr, test.want); diff != "" {
				t.Errorf("wrong result\n%s", diff)
			}
		})
	}

}
