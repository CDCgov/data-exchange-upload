package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	neturl "net/url"
	"path"

	"github.com/eventials/go-tus"
	"golang.org/x/oauth2"
)

type Result struct {
	testCase TestCase
	url      string
}

type uploadable interface {
	io.ReadSeekCloser
	Size() int64
	Metadata() map[string]string
	Fingerprint() string
}

func runTest(t TestCase, conf *config) (*Result, error) {

	var f uploadable
	if t.TemplateFile != "" {
		f = &TemplateGenerator{
			Repeats:   t.Repetitions,
			Path:      t.TemplateFile,
			Templates: t.Templates,
			Manifest:  t.Manifest,
		}
	} else {
		f = &BadFile{
			FileSize:       int(t.Size),
			Manifest:       t.Manifest,
			DummyGenerator: &RandomBytesReader{},
		}
	}

	// create the tus client.
	tusConf := tus.DefaultConfig()
	tusConf.ChunkSize = int64(t.Chunk)
	tusConf.HttpClient = &http.Client{}
	if conf.tokenSource != nil {
		tusConf.HttpClient = oauth2.NewClient(context.TODO(), conf.tokenSource)
	}
	tusConf.Header.Set("Upload-Defer-Length", "1")
	tusConf.Header.Set("Upload-Length", "")
	client, err := tus.NewClient(conf.url, tusConf)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w, %+v", err, t)
	}

	// create an upload from a file.
	upload := tus.NewUpload(f, f.Size(), f.Metadata(), f.Fingerprint())

	// create the uploader.
	uploader, err := client.CreateUpload(upload)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload: %w, %+v", err, t)
	}

	if patchURL != "" {
		p, err := neturl.JoinPath(patchURL, path.Base(uploader.Url()))
		if err != nil {
			return nil, err
		}
		uploader.SetUrl(p)
	}

	slog.Debug("UploadID", "upload_id", uploader.Url())
	c := make(chan tus.Upload)
	uploader.NotifyUploadProgress(c)
	go func(c chan tus.Upload, url string) {
		for u := range c {
			slog.Debug("Upload Progress", "url", url, "progress", u.Progress())
		}
	}(c, uploader.Url())

	for {
		if err := uploader.UploadChunck(); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
	}
	return &Result{
		testCase: t,
		url:      uploader.Url(),
	}, nil
}
