package differer

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/jimen0/differer/scheduler"
	"google.golang.org/protobuf/proto"
)

// Runner interface allows to run tasks.
type Runner interface {
	Run(ctx context.Context, data []byte) (*scheduler.Result, error)
	GetName() string
}

// CloudRunner is a Cloud Run runner.
type CloudRunner struct {
	Client  clienter
	Name    string
	Service string
}

// ensure CloudRunner implements Runner interface.
var _ Runner = (*CloudRunner)(nil)

// Run runs the runner with the given address.
func (cr *CloudRunner) Run(ctx context.Context, data []byte) (*scheduler.Result, error) {
	req, err := http.NewRequest(http.MethodPost, cr.Service, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("could not build %q request: %v", cr.GetName(), err)
	}
	req = req.WithContext(ctx)
	req.Header.Add("content-type", "application/protobuf")

	res, err := cr.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("runner returned bad status code: %v", res.StatusCode)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body from %s: %v", cr.GetName(), err)
	}

	var out scheduler.Result
	if err := proto.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("could not decode response from %s: %v", cr.GetName(), err)
	}

	return &out, nil
}

// GetName returns the Cloud Runner's name.
func (cr *CloudRunner) GetName() string {
	return cr.Name
}
