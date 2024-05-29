package postprocessing

import (
	"encoding/json"
	"io"
	"os"

	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
)

func Merge(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	id := event.Upload.ID
	dir := "./uploads/tus-prefix"
	f, err := os.Open(dir + "/" + id)
	if err != nil {
		return resp, err
	}
	os.Mkdir("./uploads/dex", 0750)
	dest, err := os.Create("./uploads/dex/" + id)
	if err != nil {
		return resp, err
	}
	if _, err := io.Copy(dest, f); err != nil {
		return resp, err
	}

	meta := event.Upload.MetaData
	if resp.ChangeFileInfo.MetaData != nil {
		meta = resp.ChangeFileInfo.MetaData
	}

	m, err := json.Marshal(meta)
	if err != nil {
		return resp, err
	}
	err = os.WriteFile("./uploads/dex/"+id+".meta", m, 0666)
	return resp, err
}
