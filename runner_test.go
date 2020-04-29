package differer

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/jimen0/differer/scheduler"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestRun(t *testing.T) {
	tt := []struct {
		name   string
		input  string
		runner *CloudRunner
		exp    *scheduler.Result
		valid  bool
	}{
		{
			name: "valid",
			runner: &CloudRunner{
				Name:    "test",
				Service: "https://example.test",
				Client: doFunc(func(req *http.Request) (*http.Response, error) {
					var resp http.Response

					resp.StatusCode = http.StatusOK
					result := &scheduler.Result{Id: "foo", Value: "bbac"}
					b, err := proto.Marshal(result)
					require.Nil(t, err)

					resp.Body = ioutil.NopCloser(bytes.NewReader(b))
					return &resp, nil
				}),
			},
			exp:   &scheduler.Result{Id: "foo", Value: "bbac"},
			valid: true,
		},
		{
			name: "error response",
			runner: &CloudRunner{
				Name:    "errorer",
				Service: "https://example.test",
				Client: doFunc(func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("error")
				}),
			},
		},
		{
			name: "unexpected status code",
			runner: &CloudRunner{
				Name:    "internal errorer",
				Service: "https://example.test",
				Client: doFunc(func(req *http.Request) (*http.Response, error) {
					var resp http.Response
					resp.StatusCode = http.StatusInternalServerError
					resp.Body = ioutil.NopCloser(strings.NewReader(``))
					return &resp, nil
				}),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.runner.Run(context.Background(), []byte(tc.input))
			if !tc.valid {
				require.NotNil(t, err)
				return
			}

			require.Equal(t, tc.exp.Id, res.Id)
			require.Equal(t, tc.exp.Value, res.Value)
			require.Equal(t, tc.exp.Error, res.Error)
		})
	}
}

// doFunc is a testing helper that builds HTTP responses.
type doFunc func(req *http.Request) (*http.Response, error)

func (fn doFunc) Do(req *http.Request) (*http.Response, error) {
	return fn(req)
}

// ensure doFunc implements clienter interface.
var _ clienter = (doFunc)(nil)
