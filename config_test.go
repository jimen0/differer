package differer

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestReadConfig(t *testing.T) {
	tt := []struct {
		name  string
		input string
		exp   *Config
		valid bool
	}{
		{
			name: "valid",
			input: `---
runners:
  a: b
timeout: 5s
`,
			exp: &Config{
				Timeout: 5 * time.Second,
				Runners: map[string]string{"a": "b"},
			},
			valid: true,
		},
		{
			name: "valid default timeout",
			input: `---
runners:
  a: b
`,
			exp: &Config{
				Timeout: 10 * time.Second,
				Runners: map[string]string{"a": "b"},
			},
			valid: true,
		},
		{
			name:  "unexpected YAML syntax",
			input: `@`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r := strings.NewReader(tc.input)
			got, err := ReadConfig(r)
			if !tc.valid {
				require.NotNil(t, err)
				return
			}
			require.Equal(t, tc.exp, got)
		})
	}
}
