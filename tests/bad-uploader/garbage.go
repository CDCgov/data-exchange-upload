package main

import (
	"io"
	"math/rand"
)

type Generator interface {
	io.Reader
}

type RandomBytesReader struct{}

func (rb *RandomBytesReader) Read(b []byte) (int, error) {
	return rand.Read(b)
}

type BadFile struct {
	FileSize       int
	offset         int
	Manifest       map[string]string
	DummyGenerator Generator
}

func (b *BadFile) Size() int64 {
	return int64(b.FileSize)
}

func (b *BadFile) Metadata() map[string]string {
	return b.Manifest
}

func (b *BadFile) Fingerprint() string {
	return ""
}

func (b *BadFile) Read(p []byte) (int, error) {

	// needs to limit size read to size eventually
	i, err := b.DummyGenerator.Read(p)
	if err != nil {
		return i, err
	}

	if b.offset+i > b.FileSize {
		i = b.FileSize - b.offset
	}

	b.offset += i

	if b.offset >= b.FileSize {
		return i, io.EOF
	}
	return i, nil
}

func (b *BadFile) Seek(offset int64, whence int) (int64, error) {
	return offset, nil
}

func (b *BadFile) Close() error {
	return nil
}
