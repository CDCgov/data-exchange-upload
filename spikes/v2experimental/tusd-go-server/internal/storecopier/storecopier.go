package storecopier

// storeCopier interface across stores: local, azure, and aws.
type StoreCopier interface {

	// CopyTusSrcToDst copies a file with metadata from .info into the destination file
	// cli flags based for local or cloud copies
	CopyTusSrcToDst() error

	// TODO:
	// CopySrcToDst copies a file from Src location to Dst location
	CopySrcToDst() error
} // .StoreCopier
