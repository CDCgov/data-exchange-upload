package serverdex

import (
	// "sync"
	"sync/atomic"
)

type Metrics struct {
	CopiedAToB *uint64
	CopiedBToC *uint64
} // Metrics

func newMetricsDex() Metrics {
	return Metrics{
		CopiedAToB: new(uint64),
		CopiedBToC: new(uint64),
	} // .return
} // .NewMetricsDex

// incCopiedAToB increases the counter for completed copies
func (m Metrics) incCopiedAToB() {
	atomic.AddUint64(m.CopiedAToB, 1)
} // .incCopiedAToB

// incCopiedBToC increases the counter for completed copies
func (m Metrics) incCopiedBToC() {
	atomic.AddUint64(m.CopiedBToC, 1)
} // .incCopiedBToC
