package storelocal

import (
	"io"
	"os"
	"path/filepath"
) // .import

type CopierLocal struct {
	SrcFileName string
	SrcFolder   string
	DstFileName string
	DstFolder   string
} // .CopierLocal

// CopyTusSrcToDst copies a file locally from tus upload folder to another folder
func (cl CopierLocal) CopyTusSrcToDst() error {

	// Src
	srcPath, err := filepath.Abs(cl.SrcFolder + "/" + cl.SrcFileName)
	if err != nil {
		return err
	} // .if
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	} // .if
	defer srcFile.Close()

	// Dst
	dstPath, err := filepath.Abs(cl.DstFolder + "/" + cl.DstFileName)
	if err != nil {
		return err
	} // .if
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	} // .dstFile
	defer dstFile.Close()

	// Copy src -> dst
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	} // ._

	// Flush dst to ensure completion
	err = dstFile.Sync()
	if err != nil {
		return err
	} // .err

	return nil
} // .CopySrcToDst
