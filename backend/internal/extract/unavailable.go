package extract

import (
	"context"
	"errors"
	"fmt"
)

// ErrNotConfigured is returned by Unavailable. The HTTP layer maps it to
// 503 Service Unavailable rather than 502 Bad Gateway: nothing upstream
// failed, this deployment simply has no provider credentials. The
// distinction matters to an operator reading the logs and to a family
// retrying the request — one means "try again later", the other means
// "this instance will never succeed until it is configured".
var ErrNotConfigured = errors.New("extract: no API key configured for the selected provider")

// Unavailable is the Extractor used when the selected provider has no API
// key. It exists so the server can still start and serve the dataset,
// package search, and health endpoints — a deployment missing one secret
// should degrade to "intake is off", not "everything is off".
type Unavailable struct {
	Provider string
	EnvVar   string
}

// Available reports false so /api/v1/health can say so honestly.
func (Unavailable) Available() bool { return false }

func (u Unavailable) Extract(context.Context, string, []CandidatePackageInfo, []CandidateExclusionInfo) (Result, error) {
	provider := u.Provider
	if provider == "" {
		provider = "unknown"
	}
	envVar := u.EnvVar
	if envVar == "" {
		envVar = "the provider API key"
	}
	return Result{}, fmt.Errorf("%w (provider %q, set %s)", ErrNotConfigured, provider, envVar)
}

var _ Extractor = Unavailable{}
var _ Availability = Unavailable{}
