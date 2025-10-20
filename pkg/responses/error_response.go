package responses

import "fmt"

type ErrorResponse struct {
	MessageID string
	Err       error
}

// Error implements error.
func (e *ErrorResponse) Error() string {
	panic("unimplemented")
}

func (e *ErrorResponse) ErrorString() string {
	return fmt.Sprintf("MessageID: %s, Error: %v",
		e.MessageID, e.Err)
}

func (e *ErrorResponse) NewErrorResponse(messageId string, err error) *ErrorResponse {
	return &ErrorResponse{
		MessageID: messageId,
		Err:       err,
	}
}
