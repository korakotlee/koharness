// Package credentials provides external secret provider abstractions and resolvers
// for dynamically expanding credential references (such as Proton Pass URIs)
// into Model Context Protocol (MCP) tool configuration fields during sync operations.
package credentials

import (
	"context"
)

// CredentialResolver defines the interface for evaluating and resolving sensitive credential
// reference URIs (such as pass://<vault>/<item>/<field>) into plain-text secret strings.
// Implementations handle provider binary detection, authentication status checks, and context-aware
// secret retrieval from external secret managers.
type CredentialResolver interface {
	// ProviderName returns the human-readable identifier of the credential provider (e.g., "Proton Pass").
	ProviderName() string

	// CanResolve evaluates whether the given URI string can be handled by this resolver.
	// Returns true if the scheme and structure match the provider's specification.
	CanResolve(uri string) bool

	// Resolve attempts to fetch and return the plain-text secret associated with the URI reference.
	// It accepts a context to allow execution cancellation and timeouts. If resolution fails,
	// an error is returned describing the failure mode (e.g. binary missing, unauthenticated, item not found).
	Resolve(ctx context.Context, uri string) (string, error)
}
