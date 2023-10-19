package erring

type SQSEventError struct {
	NewVisibilityTimeout int32
	Message              string
}

func (e *SQSEventError) Error() string { return e.Message }

func NewSQSEventError(newVisibilityTimeout int32, message string) *SQSEventError {
	return &SQSEventError{
		NewVisibilityTimeout: newVisibilityTimeout,
		Message:              message,
	}
}
