package main

import (
	"io"
	"log"
	"os"
	"text/template"
)

type TemplateGenerator struct {
	t        *template.Template
	Path     string
	Repeats  int
	Manifest map[string]string
	r        io.Reader
	w        io.Writer
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

	if tg.r == nil || tg.w == nil {
		tg.r, tg.w = io.Pipe()
	}

	return tg.t.Execute(tg.w, nil)
}

func (tg *TemplateGenerator) Metadata() map[string]string {
	return tg.Manifest
}

func (tg *TemplateGenerator) Fingerprint() string {
	return ""
}

func (tg *TemplateGenerator) Read(p []byte) (int, error) {
	n, err := tg.r.Read(p)
	if err == io.EOF {
		tg.next()
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
