package actions

import (
	"context"

	"github.com/tinker-works/donsy/netomatic"
)

// fakeClient embeds the public interface so each test overrides only the
// operation it exercises; no internal daemon or application contract leaks into
// the action tests.
type fakeClient struct {
	netomatic.Client
	log       netomatic.ReadDaemonLogResponse
	logErr    error
	logOffset int64
	logLimit  int
}

func (f *fakeClient) ReadDaemonLog(_ context.Context, offset int64, limit int) (netomatic.ReadDaemonLogResponse, error) {
	f.logOffset, f.logLimit = offset, limit
	return f.log, f.logErr
}

var _ netomatic.Client = (*fakeClient)(nil)
