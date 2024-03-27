package serverdex

import (
	// "sync"
	"sync/atomic"
)

type Metrics struct {
	CopiedUploadToDex    *uint64
	CopiedUploadToRouter *uint64
	CopiedUploadToEdav   *uint64
} // Metrics

func newMetricsDex() Metrics {
	return Metrics{
		CopiedUploadToDex:    new(uint64),
		CopiedUploadToRouter: new(uint64),
		CopiedUploadToEdav:   new(uint64),
	} // .return
} // .NewMetricsDex

// CopiedUploadToDex increases the counter for completed copies
func (m Metrics) IncUploadToDex() {
	atomic.AddUint64(m.CopiedUploadToDex, 1)
} // .CopiedUploadToDex

// IncUploadToRouter increases the counter for completed copies
func (m Metrics) IncUploadToRouter() {
	atomic.AddUint64(m.CopiedUploadToRouter, 1)
} // .IncUploadToRouter

// IncUploadToEdav increases the counter for completed copies
func (m Metrics) IncUploadToEdav() {
	atomic.AddUint64(m.CopiedUploadToEdav, 1)
} // .IncUploadToEdav
