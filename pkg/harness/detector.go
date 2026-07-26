package harness

import (
	"os"

	"github.com/spf13/afero"
)

// DetectorOption configures operational parameters for Detector instances.
type DetectorOption func(*Detector)

// WithFs returns a DetectorOption that sets a custom filesystem abstraction.
func WithFs(fs afero.Fs) DetectorOption {
	return func(d *Detector) {
		d.fs = fs
	}
}

// WithHomeDir returns a DetectorOption that overrides the workstation user home directory.
func WithHomeDir(dir string) DetectorOption {
	return func(d *Detector) {
		d.homeDir = dir
	}
}

// Detector manages discovery and status evaluation across supported client harnesses.
type Detector struct {
	fs       afero.Fs
	homeDir  string
	adapters []HarnessAdapter
}

// NewDetector creates and initializes a new Detector using default OS filesystem
// settings or provided functional options.
func NewDetector(opts ...DetectorOption) (*Detector, error) {
	d := &Detector{
		fs: afero.NewOsFs(),
	}

	home, err := os.UserHomeDir()
	if err == nil {
		d.homeDir = home
	}

	for _, opt := range opts {
		opt(d)
	}

	d.adapters = []HarnessAdapter{
		NewAntigravityAdapter(d.fs, d.homeDir),
		NewClaudeAdapter(d.fs, d.homeDir),
		NewCodexAdapter(d.fs, d.homeDir),
	}

	return d, nil
}

// DetectAll evaluates installation statuses for all registered client harness adapters.
func (d *Detector) DetectAll() []HarnessStatus {
	statuses := make([]HarnessStatus, 0, len(d.adapters))
	for _, adapter := range d.adapters {
		statuses = append(statuses, adapter.GetStatus())
	}
	return statuses
}

// DetectInstalled returns adapters corresponding to client harnesses detected on disk.
func (d *Detector) DetectInstalled() []HarnessAdapter {
	installed := make([]HarnessAdapter, 0)
	for _, adapter := range d.adapters {
		if adapter.IsInstalled() {
			installed = append(installed, adapter)
		}
	}
	return installed
}

// GetAdapter retrieves a registered HarnessAdapter matching the requested HarnessID.
func (d *Detector) GetAdapter(id HarnessID) (HarnessAdapter, bool) {
	for _, adapter := range d.adapters {
		if adapter.ID() == id {
			return adapter, true
		}
	}
	return nil, false
}

// GetAdapters returns all registered harness adapters regardless of installation state.
func (d *Detector) GetAdapters() []HarnessAdapter {
	result := make([]HarnessAdapter, len(d.adapters))
	copy(result, d.adapters)
	return result
}
