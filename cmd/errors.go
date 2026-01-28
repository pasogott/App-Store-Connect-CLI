package cmd

// ReportedError marks an error as already reported to the user.
// The main entrypoint should exit non-zero without duplicating output.
type ReportedError interface {
	error
	Reported() bool
}

type reportedError struct {
	err error
}

func (e reportedError) Error() string {
	return e.err.Error()
}

func (e reportedError) Unwrap() error {
	return e.err
}

func (e reportedError) Reported() bool {
	return true
}

// NewReportedError wraps an error that has already been printed.
func NewReportedError(err error) error {
	if err == nil {
		return nil
	}
	return reportedError{err: err}
}
