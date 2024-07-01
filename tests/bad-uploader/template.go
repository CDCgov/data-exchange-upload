package main

import (
	"io"
	"log"
	"log/slog"
	"os"
	"text/template"
)

type TemplateGenerator struct {
	t        *template.Template
	Path     string
	Repeats  int
	Manifest map[string]string
	r        io.Reader
	w        io.WriteCloser
}

func (tg *TemplateGenerator) Size() int64 {
	fi, err := os.Stat(tg.Path)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return fi.Size() * int64(tg.Repeats)
}

func (tg *TemplateGenerator) next() (err error) {
	if tg.t == nil {
		tg.t, err = template.New(tg.Path).ParseFiles(tg.Path)
		if err != nil {
			return err
		}
	}

	tg.r, tg.w = io.Pipe()

	go func() {
		slog.Debug("writing template")
		tg.t.Execute(tg.w, nil)
		tg.w.Close()
	}()
	return nil
}

func (tg *TemplateGenerator) Metadata() map[string]string {
	return tg.Manifest
}

func (tg *TemplateGenerator) Fingerprint() string {
	return ""
}

func (tg *TemplateGenerator) Read(p []byte) (int, error) {
	if tg.t == nil {
		if err := tg.next(); err != nil {
			return 0, err
		}
	}
	slog.Debug("reading template")
	n, err := tg.r.Read(p)
	slog.Debug("read template")
	if err == io.EOF {
		slog.Debug("hit eof")
		//TODO if we should stop return the EOF?
		if err := tg.next(); err != nil {
			return n, err
		}
		// do we need to re-read here or will it just try again?
		return n, nil
	}
	return n, err
}

func (tg *TemplateGenerator) Seek(offset int64, whence int) (int64, error) {
	return offset, nil
}

func (tg *TemplateGenerator) Close() error {
	return nil
}
