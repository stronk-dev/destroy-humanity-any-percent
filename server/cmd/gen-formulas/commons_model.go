package main

import (
	"cloud-clicker/server/commons"
	"cloud-clicker/server/formulas"
)

func commonsSourceWeights(source []commons.SourceWeight) []formulas.CommonsSourceWeight {
	result := make([]formulas.CommonsSourceWeight, len(source))
	for index, row := range source {
		result[index] = formulas.CommonsSourceWeight{
			SourceID: row.SourceID, Slot: row.Slot, WeightPPM: row.WeightPPM, Forsworn: row.Forsworn,
		}
	}
	return result
}
