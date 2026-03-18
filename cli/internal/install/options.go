package install

// Options contains configuration for the installation process
type Options struct {
	// ForceReinstall triggers a clean-slate installation
	ForceReinstall bool

	// CleanOnly only runs cleanup without reinstalling (used with ForceReinstall)
	CleanOnly bool

	// SkipJAMF skips JAMF certificate provisioning
	SkipJAMF bool

	// SkipPlugins skips plugin installation
	SkipPlugins bool

	// NoInteractive disables TUI for CI/scripts
	NoInteractive bool

	// Resume continues from a failed installation
	Resume bool

	// NoBrew skips Homebrew and all brew-based installations.
	// Instead of installing, missing tools are reported to the user.
	NoBrew bool

	// Verbose shows detailed output
	Verbose bool

	// TargetVersion specifies a specific version to install (empty = latest)
	TargetVersion string

	// ExtraPlugins are additional plugins to install alongside defaults
	ExtraPlugins []string
}

// DefaultOptions returns the default installation options
func DefaultOptions() *Options {
	return &Options{
		ForceReinstall: false,
		CleanOnly:      false,
		SkipJAMF:       false,
		SkipPlugins:    false,
		NoBrew:         false,
		NoInteractive:  false,
		Resume:         false,
		Verbose:        false,
		TargetVersion:  "",
	}
}
