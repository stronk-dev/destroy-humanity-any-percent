package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"cloud-clicker/server/multiplier"
)

type formulaArtifact struct {
	SchemaVersion       int               `json:"schema_version"`
	ProductionRate      string            `json:"production_rate"`
	MultiplierSlotOrder []multiplier.Slot `json:"multiplier_slot_order"`
	WithinSlotOrder     string            `json:"within_slot_order"`
}

func main() {
	output := flag.String("output", "", "output JSON filename")
	flag.Parse()
	if *output == "" {
		fmt.Fprintln(os.Stderr, "-output is required")
		os.Exit(2)
	}
	artifact := formulaArtifact{
		SchemaVersion:       1,
		ProductionRate:      "sum_generators(count * base_rate * product(multiplier_slots))",
		MultiplierSlotOrder: append([]multiplier.Slot(nil), multiplier.Order[:]...),
		WithinSlotOrder:     "source_id_raw_byte_ascending",
	}
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		panic(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(*output, data, 0o644); err != nil {
		panic(err)
	}
}
