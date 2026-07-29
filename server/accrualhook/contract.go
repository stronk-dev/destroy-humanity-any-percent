// Package accrualhook defines the neutral post-accrual extension boundary.
package accrualhook

import (
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

type Result struct {
	Receipt        economy.Receipt
	ElapsedMS      int64
	ProductionMS   int64
	BankedCreditMS int64
}

type Hook interface {
	AfterAccrual(*save.State, *economy.Catalog, save.Revision, Result, []multiplier.Contribution) ([]save.EventWrite, error)
}
