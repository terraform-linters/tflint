package models

import "testing"

func Test_truncateLongMessage(t *testing.T) {
	cases := []struct {
		Name     string
		Text     string
		Expected string
	}{
		{
			Name:     "short text",
			Text:     "foo",
			Expected: "foo",
		},
		{
			Name:     "long text",
			Text:     "looooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong text",
			Expected: "looooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong...",
		},
		{
			Name: "with newlines",
			Text: `foo
bar`,
			Expected: "foo\\nbar",
		},
	}

	for _, tc := range cases {
		ret := truncateLongMessage(tc.Text)
		if ret != tc.Expected {
			t.Fatalf("Fail `%s`: expected=%s got=%s", tc.Name, tc.Expected, ret)
		}
	}
}
