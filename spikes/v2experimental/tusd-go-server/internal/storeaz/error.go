package storeaz

type Error struct {
	Storage string
	Message string
} // .Error

func (e Error) Error() string {
	return e.Storage + ": " + e.Message
} // .Error

func (e1 Error) Is(target error) bool {
	e2, ok := target.(Error)
	return ok && e1.Storage == e2.Storage && e1.Message == e2.Message
} // .Is

func NewError(storage, message string) Error {
	return Error{
		Storage: storage,
		Message: message,
	} // .return
} // .NewError
