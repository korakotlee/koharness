package credentials

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

var (
	// ErrBinaryNotFound is returned when the pass-cli executable is missing from system $PATH.
	ErrBinaryNotFound = errors.New("pass-cli binary not found in system $PATH")
	// ErrInvalidURI is returned when the provided secret reference URI does not follow pass:// schema.
	ErrInvalidURI = errors.New("invalid secret URI: scheme must start with pass://")
	// ErrResolutionFailed is returned when pass-cli returns a non-zero exit code or fails to resolve the item.
	ErrResolutionFailed = errors.New("pass-cli secret resolution failed")
)

// CommandExecutor abstracts OS command execution to enable clean unit test mocking.
type CommandExecutor func(ctx context.Context, name string, args ...string) ([]byte, error)

// DefaultCommandExecutor executes commands via os/exec.CommandContext.
func DefaultCommandExecutor(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return nil, fmt.Errorf("%w: %s", err, stderrStr)
		}
		return nil, err
	}

	return stdout.Bytes(), nil
}

// ProtonPassResolver resolves secrets stored in Proton Pass using the pass-cli command line utility.
type ProtonPassResolver struct {
	// Exec is the command execution function used to run pass-cli subcommands.
	Exec CommandExecutor
	// LookPath is the function used to locate the pass-cli binary in system $PATH.
	LookPath func(file string) (string, error)
	// Timeout specifies the maximum context timeout for pass-cli subprocess calls.
	Timeout time.Duration
}

// NewProtonPassResolver constructs a ProtonPassResolver initialized with standard system execution primitives.
func NewProtonPassResolver() *ProtonPassResolver {
	return &ProtonPassResolver{
		Exec:     DefaultCommandExecutor,
		LookPath: exec.LookPath,
		Timeout:  5 * time.Second,
	}
}

// ProviderName returns the human-readable identifier of the credential provider.
func (p *ProtonPassResolver) ProviderName() string {
	return "Proton Pass"
}

// CanResolve checks whether the given URI uses the pass:// scheme.
func (p *ProtonPassResolver) CanResolve(uri string) bool {
	return strings.HasPrefix(strings.TrimSpace(uri), "pass://")
}

// IsAvailable checks if pass-cli is available in system $PATH.
func (p *ProtonPassResolver) IsAvailable() bool {
	lookPath := p.LookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	_, err := lookPath("pass-cli")
	return err == nil
}

// Resolve retrieves the secret value associated with the pass:// URI using pass-cli item view.
func (p *ProtonPassResolver) Resolve(ctx context.Context, uri string) (string, error) {
	uri = strings.TrimSpace(uri)
	if !p.CanResolve(uri) {
		return "", ErrInvalidURI
	}

	if !p.IsAvailable() {
		return "", ErrBinaryNotFound
	}

	execFn := p.Exec
	if execFn == nil {
		execFn = DefaultCommandExecutor
	}

	timeout := p.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute pass-cli item view <uri>
	output, err := execFn(reqCtx, "pass-cli", "item", "view", uri)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrResolutionFailed, err)
	}

	secret := strings.TrimRight(string(output), "\r\n")
	return secret, nil
}
