package main

import (
	"bufio"
	"errors"
	"io"
	"log/slog"
	"math/rand"
	"text/template"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func repeat(n int) []int {
	return make([]int, n)
}

var funcs = template.FuncMap{
	"RandomString": randSeq,
	"RandomInt":    rand.Intn,
	"Repeat":       repeat,
}

type TemplateGenerator struct {
	t         *template.Template
	Path      string
	Repeats   int
	Templates []SubTemplate
	Manifest  map[string]string
	r         *bufio.Reader
	w         io.WriteCloser
}

func (tg *TemplateGenerator) Size() int64 {
	return 1
}

func (tg *TemplateGenerator) next() (err error) {
	if tg.t == nil {
		tg.t, err = template.New(tg.Path).Funcs(funcs).ParseFiles(tg.Path)
		if err != nil {
			return err
		}
	}

	var r io.Reader
	r, tg.w = io.Pipe()
	tg.r = bufio.NewReader(r)

	go func() {
		templates := tg.Templates
		for range tg.Repeats {
			for _, t := range templates {
				if t.Args == nil {
					t.Args = map[string]any{}
				}
				for i := range t.Repetitions {
					t.Args["Index"] = i
					if err := tg.t.ExecuteTemplate(tg.w, t.Name, t.Args); err != nil {
						slog.Error("failed to execute template", "error", err)
					}
				}
			}
		}
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

	_, peakErr := tg.r.Peek(1)

	n, err := io.ReadFull(tg.r, p)

	//todo only swallow unexpected eof errors
	slog.Debug("read template", "n", n, "p", len(p))

	if err != nil || peakErr != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(peakErr, io.EOF) {
			return n, io.EOF
		}
		slog.Error("template read error", "error", err, "peak error", peakErr)
	}
	return n, nil
}

func (tg *TemplateGenerator) Seek(offset int64, whence int) (int64, error) {
	return offset, nil
}

func (tg *TemplateGenerator) Close() error {
	return nil
}
