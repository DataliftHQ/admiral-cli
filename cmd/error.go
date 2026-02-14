package cmd

type exitError struct {
	err  error
	code int
}

func (e *exitError) Error() string {
	return e.err.Error()
}
