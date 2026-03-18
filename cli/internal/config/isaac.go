package config

// Mode names used in the settings TUI.
const (
	ModeDirect       = "direct"
	ModePowerUser    = "power_user"
	ModeModelServing = "model_serving"
)

// IsaacConfig is a stub retained for source compatibility.
// On personal machines there is no isaac/dbexec launcher;
// this type carries no real state.
type IsaacConfig struct {
	mode string
}

// LoadIsaacConfig returns a default config (no file I/O).
func LoadIsaacConfig() (*IsaacConfig, error) {
	return &IsaacConfig{mode: ModeDirect}, nil
}

// GetMode returns the current named mode.
func (c *IsaacConfig) GetMode() string {
	if c.mode == "" {
		return ModeDirect
	}
	return c.mode
}

// SetMode updates the mode.
func (c *IsaacConfig) SetMode(mode string) {
	c.mode = mode
}

// SetModeViaIsaac is a no-op on personal machines (no dbexec available).
func SetModeViaIsaac(mode string) error {
	return nil
}
