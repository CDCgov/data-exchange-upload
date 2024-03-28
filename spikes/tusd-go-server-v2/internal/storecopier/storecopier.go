package storecopier

// StoreTusCopier interface across stores: local, azure, and aws.
type StoreTusCopier interface {

	// CopyTusSrcToDst copies a file with metadata from .info into the destination file
	// cli flags based for local or cloud copies
	CopyTusSrcToDst() error
} // .StoreCopier

// TODO:, this will be the router also

// StoreCopier interface across stores: local, azure, and aws.
type StoreCopier interface {

	// TODO:
	// CopySrcToDst copies a file from Src location to Dst location
	CopySrcToDst() error
} // .StoreCopier
