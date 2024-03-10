package storecopier

import (
	"io"
	"os"
	"path/filepath"
) // .import

type StoreLocal struct {
	FileLocalFolder string
	FileName        string
} // .StoreLocal

func CopySrcToDst(src StoreLocal, dst StoreLocal) error {
	// Src
	srcPath, err := filepath.Abs(src.FileLocalFolder + "/" + src.FileName)
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	} // .if
	defer srcFile.Close()

	// Dst
	dstPath, err := filepath.Abs(dst.FileLocalFolder + "/" + dst.FileName)
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
