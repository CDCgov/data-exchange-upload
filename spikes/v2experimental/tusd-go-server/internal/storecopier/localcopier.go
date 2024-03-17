package storecopier

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
	tusd "github.com/tus/tusd/v2/pkg/handler"
) // .import

// localFromTusCopier based on configuration, it is used for moving files locally
// localFromTusCopier implements CopyTusSrcToDst
type localFromTusCopier struct {
	appConfig    appconfig.AppConfig
	uploadConfig metadatav1.AllUploadConfigs

	tusdEvent tusd.HookEvent
} // .StoreLocal

// CopyTusSrcToDst copies a file locally from tus upload folder to another folder
func (lc localFromTusCopier) CopyTusSrcToDst() error {

	srcFileName := lc.tusdEvent.Upload.MetaData["filename"]
	srcFolder := lc.appConfig.LocalFolderUploadsTus

	// TODO config destination per respective upload config
	// TODO: adding file ticks, change per config
	dstFolder := lc.appConfig.LocalFolderUploadsA
	dstFileName := lc.tusdEvent.Upload.MetaData["filename"] + "_" + strconv.FormatInt(time.Now().UnixNano(), 10)

	// Src
	srcPath, err := filepath.Abs(srcFolder + "/" + srcFileName)
	if err != nil {
		return err
	} // .if
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	} // .if
	defer srcFile.Close()

	// Dst
	dstPath, err := filepath.Abs(dstFolder + "/" + dstFileName)
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
