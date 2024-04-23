package cli

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
)

var (
	ErrNotFound = errors.New("Expected file not found")
)

type UploadInspecter interface {
	InspectInfoFile(id string) (map[string]any, error)
	InspectUploadedFile(id string) (map[string]any, error)
}

func NewFileSystemUploadInspector(baseDir string) *FileSystemUploadInspector {
	return &FileSystemUploadInspector{
		BaseDir: baseDir,
	}
}

type FileSystemUploadInspector struct {
	BaseDir string
}

type InfoFileData struct {
	MetaData map[string]any `json:"MetaData"`
}

func (fsui *FileSystemUploadInspector) InspectInfoFile(id string) (map[string]any, error) {
	// First, read in the id + .info file.
	//TODO make this stronger
	infoFilename := fsui.BaseDir + "/" + id + ".info"
	fileBytes, err := os.ReadFile(infoFilename)
	if err != nil {
		return nil, err
	}

	// Deserialize to hash map.
	jsonMap := &InfoFileData{}
	if err := json.Unmarshal(fileBytes, jsonMap); err != nil {
		return nil, err
	}

	return jsonMap.MetaData, nil
}

func (fsui *FileSystemUploadInspector) InspectUploadedFile(id string) (map[string]any, error) {
	filename := fsui.BaseDir + "/" + id
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	uploadedFileInfo := map[string]any{
		"updated_at": fi.ModTime(),
		"size_bytes": fi.Size(),
	}
	return uploadedFileInfo, nil
}

type InfoResponse struct {
	Manifest map[string]any `json:"manifest"`
	FileInfo map[string]any `json:"file_info"`
}

type InfoHandler struct {
	inspecter UploadInspecter
}

func (ih *InfoHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	//todo get the upload id
	id := r.PathValue("UploadID")

	fileInfo, err := ih.inspecter.InspectInfoFile(id)
	if err != nil {
		//todo real error handling
		status := http.StatusNotFound
		if !errors.Is(err, ErrNotFound) {
			status = http.StatusInternalServerError
		}
		http.Error(rw, err.Error(), status)
		return
	}
	uploadedFileInfo, err := ih.inspecter.InspectUploadedFile(id)
	if err != nil {
		//todo real error handling
		status := http.StatusNotFound
		if !errors.Is(err, ErrNotFound) {
			status = http.StatusInternalServerError
		}
		http.Error(rw, err.Error(), status)
		return
	}

	response := &InfoResponse{
		Manifest: fileInfo,
		FileInfo: uploadedFileInfo,
	}

	enc := json.NewEncoder(rw)
	enc.Encode(response)

}

func createInspector(appConfig *appconfig.AppConfig) (UploadInspecter, error) {
	if appConfig.TusAzStorageConfig != nil {
		//return NewAzureUploadInspector(appConfig.TusAzStorageConfig)
		return nil, errors.New("not implemented")
	}
	if appConfig.LocalFolderUploadsTus != "" {
		return NewFileSystemUploadInspector(appConfig.LocalFolderUploadsTus), nil
	}

	return nil, errors.New("oh no badly configured!")
}

func GetUploadInfoHandler(appConfig *appconfig.AppConfig) (http.Handler, error) {
	inspector, err := createInspector(appConfig)
	return &InfoHandler{
		inspector,
	}, err
}
