package postprocessing

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

var targets = map[string]Deliverer{}

func RegisterTarget(name string, d Deliverer) {
	targets[name] = d
}

type Deliverer interface {
	Deliver(tuid string, metadata map[string]string) error
}

// target may end up being a type
func Deliver(tuid string, manifest map[string]string, target string) error {
	// lets make this really dumb, it should take a file uri and take the rest from there. That makes it pretty recoverable.
	// root -> ""
	// date -> default pattern
	// "" -> ""

	d, ok := targets[target]
	if !ok {
		return errors.New("not recoverable, bad target " + target)
	}
	return d.Deliver(tuid, manifest)
}

type FileDeliverer struct {
	From   fs.FS
	ToPath string
}

func (fd *FileDeliverer) Deliver(tuid string, manifest map[string]string) error {
	//dir := "./uploads/tus-prefix"
	f, err := fd.From.Open(tuid)
	if err != nil {
		return err
	}
	defer f.Close()
	os.Mkdir(fd.ToPath, 0755)
	dest, err := os.Create(filepath.Join(fd.ToPath, tuid))
	if err != nil {
		return err
	}
	defer dest.Close()
	if _, err := io.Copy(dest, f); err != nil {
		return err
	}

	m, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(fd.ToPath, tuid+".meta"), m, 0666)
	return err

}
