package install

// JAMFCertsStep is a no-op on personal computers (Jamf is Databricks MDM only).
type JAMFCertsStep struct{}

func (s *JAMFCertsStep) ID() string                 { return "jamf_certs" }
func (s *JAMFCertsStep) Name() string               { return "System Certificates" }
func (s *JAMFCertsStep) Description() string        { return "Check system certificate setup" }
func (s *JAMFCertsStep) ActiveForm() string         { return "Checking certificates" }
func (s *JAMFCertsStep) CanSkip(opts *Options) bool { return true }
func (s *JAMFCertsStep) NeedsSudo() bool            { return false }

func (s *JAMFCertsStep) Check(ctx *Context) (bool, error) {
	return true, nil // Always passes — no enterprise MDM needed
}

func (s *JAMFCertsStep) Run(ctx *Context) StepResult {
	return Skip("No enterprise MDM required on personal machines")
}
