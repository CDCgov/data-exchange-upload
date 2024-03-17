package storecopier

// storeCopier interface across stores: local, azure, and aws.
type storeCopier interface {

	// CopyTusSrcToDst copies a file with metadata from .info into the destination file
	// cli flags based for local or cloud copies
	copyTusSrcToDst() error

	// TODO:
	// CopySrcToDst copies a file from Src location to Dst location
	copySrcToDst() error
} // .StoreCopier
