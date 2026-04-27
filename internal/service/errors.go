package service

type TooManyRequestsError struct {
	Message string
	Active  int
	Limit   int
}

func (e *TooManyRequestsError) Error() string {
	return e.Message
}
