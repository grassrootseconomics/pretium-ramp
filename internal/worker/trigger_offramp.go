package worker

import (
	"github.com/grassrootseconomics/pretium-ramp/internal/store"
	"github.com/riverqueue/river"
)

type (
	OfframpArgs struct {
		InitiatorAddress string `json:"initiatorAddress"`
		TransactionHash  string `json:"transactionHash"`
		Amount           string `json:"amount"`
	}

	OfframpWorker struct {
		river.WorkerDefaults[OfframpArgs]
		wc *WorkerContainer
	}
)

func (OfframpArgs) Kind() string { return store.OFFRAMP }
