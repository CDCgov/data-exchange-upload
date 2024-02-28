package processingstatus

type UploadError struct {
	// have some pre-defined values, e.g. Upload API, version, etc...

	// TODO

} // .reportingError

func Send(ue UploadError) error { // TODO: probably not return if an error

	
 // TODO: what happens when we can't send to processing status and error? 
 // should it even be returned? probably not so this can be called on a goroutine 


 // TODO send to processing status API

	return nil // all good no errors
} // .Send