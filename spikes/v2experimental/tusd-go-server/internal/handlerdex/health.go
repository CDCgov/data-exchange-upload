package handlerdex

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/cliflags"
) // .import

type HealthResp struct { // TODO: line up with DEX other products and apps

	RootResp        // Embedding rootResp, TODO: maybe this is not needed
	Health   string `json:"health"` // TODO: line-up
	Status   string `json:"status"`

	// TODO: temp for dev
	ErrAzCred string `json:"err_az_credential"`
	ErrAzBlob string `json:"err_az_blob"`
} // .Health

// health responds to /health endpoint with the health of the app
// TODO: line-up with DEX standards
// TODO: check the dependencies such as storages
func (hd *HandlerDex) health(w http.ResponseWriter, r *http.Request) {

	status := "ok"
	errAzCredStr := "nil"
	errAzBlobStr := "nil"

	// check storage dependency:
	var errAzCred, errAzBlob error

	if hd.cliFlags.Environment == cliflags.ENV_AZURE || hd.cliFlags.Environment == cliflags.ENV_LOCAL_TO_AZURE {

		credential, err := azblob.NewSharedKeyCredential(hd.appConfig.AzStorageName, hd.appConfig.AzStorageKey)
		errAzCred = err

		_, errAzBlob = azblob.NewClientWithSharedKeyCredential(hd.appConfig.AzContainerEndpoint, credential, nil)
	} // .if

	if errAzCred != nil {
		errAzCredStr = errAzCred.Error()
		status = "bad"
	} // .if
	if errAzBlob != nil {
		errAzBlobStr = errAzBlob.Error()
		status = "bad"
	} // .if

	jsonResp, err := json.Marshal(HealthResp{
		RootResp: RootResp{
			System:      hd.appConfig.System,
			DexProduct:  hd.appConfig.DexProduct,
			DexApp:      hd.appConfig.DexApp,
			ServerTime:  time.Now().Format(time.RFC3339),
			Environment: hd.cliFlags.Environment,
		},
		Health: "All Good",
		Status: status,

		ErrAzCred: errAzCredStr,
		ErrAzBlob: errAzBlobStr,
	}) // .jsonResp
	if err != nil {
		errMsg := "error marshal json for health response"
		hd.logger.Error(errMsg, "error", err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	} // .if

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
} // .health
