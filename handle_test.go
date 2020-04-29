package differer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jimen0/differer/scheduler"
	"github.com/stretchr/testify/require"
)

func TestHandleInput(t *testing.T) {
	tt := []struct {
		name    string
		method  string
		input   *input
		runners []Runner

		expStatus  int
		expResults *output
	}{
		{
			name:    "valid",
			method:  http.MethodPost,
			input:   &input{Addrs: []string{"https://example.com/"}},
			runners: []Runner{stubRunner{expected: &scheduler.Result{Id: "bbac", Value: "test"}}},

			expStatus: http.StatusOK,
			expResults: &output{
				Results: []addressResult{
					{
						Input:  "https://example.com/",
						Runner: "stub",
						Output: &scheduler.Result{
							Id:    "bbac",
							Value: "test",
						},
					},
				},
			},
		},
		{
			name:    "invalid method",
			method:  http.MethodGet,
			runners: []Runner{stubRunner{expectedErr: errors.New("invalid method")}},

			expStatus: http.StatusMethodNotAllowed,
		},
		{
			name:    "empty input",
			method:  http.MethodPost,
			input:   &input{Addrs: []string{}},
			runners: []Runner{stubRunner{expectedErr: errors.New("empty input")}},

			expStatus: http.StatusBadRequest,
		},
		{
			name:    "bad runner",
			method:  http.MethodPost,
			input:   &input{Addrs: []string{"https://example.com/"}},
			runners: []Runner{stubRunner{expectedErr: errors.New("bad runner")}},

			expStatus: http.StatusOK,
			expResults: &output{
				Results: []addressResult{
					{
						Input:  "https://example.com/",
						Runner: "stub",
						Output: nil,
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.input)
			require.Nil(t, err)

			req := httptest.NewRequest(tc.method, "/differer", bytes.NewReader(b))
			w := httptest.NewRecorder()
			HandleInput(tc.runners).ServeHTTP(w, req)
			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, tc.expStatus, res.StatusCode)
			// for error requests only status must be checked.
			if tc.expStatus != http.StatusOK {
				return
			}

			var got output
			err = json.NewDecoder(res.Body).Decode(&got)
			require.Nil(t, err)
			require.Equal(t, tc.expResults, &got)
		})
	}
}

type stubRunner struct {
	expected    *scheduler.Result
	expectedErr error
}

// Run returns the configured results for the stub.
func (sr stubRunner) Run(ctx context.Context, data []byte) (*scheduler.Result, error) {
	return sr.expected, sr.expectedErr
}

// Name returns stub's name.
func (sr stubRunner) GetName() string {
	return "stub"
}

// ensure stubRunner implements the Runner interface.
var _ Runner = (*stubRunner)(nil)
