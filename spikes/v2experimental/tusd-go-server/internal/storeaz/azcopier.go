package storeaz

import (
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
	tusd "github.com/tus/tusd/v2/pkg/handler"
)

type CopierAz struct {
	EventUploadComplete tusd.HookEvent
	UploadConfig        metadatav1.UploadConfig
	CopyTargets         []metadatav1.CopyTarget
	// SrcFileName string
	// SrcFolder   string
	// DstFileName string
	// DstFolder   string
} // .CopierLocal

// CopyTusSrcToDst copies a file locally from tus upload folder to another folder
func (caz CopierAz) CopyTusSrcToDst() error {

	// TODO

	return nil
} // .CopySrcToDst
