package erring

var (
	ErrJobNotFound   = NewAppError("job-control:not-found", "job not found")
	ErrMustNotRunJob = NewAppError("job-control:must-not-run", "job must not run now")
)
