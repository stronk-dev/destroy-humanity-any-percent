package harness

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/relevancepolicy"
	"cloud-clicker/server/routes"
)

const RelevancePolicySchemaVersion = relevancepolicy.RelevancePolicySchemaVersion
const relevanceMaxSafeInteger = int64(9_007_199_254_740_991)

var relevanceIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
var relevanceHashPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

type RelevanceWindow = relevancepolicy.RelevanceWindow
type RelevancePolicyItem = relevancepolicy.RelevancePolicyItem
type RelevancePolicyGroup = relevancepolicy.RelevancePolicyGroup
type RelevancePolicy = relevancepolicy.RelevancePolicy

func LoadRelevancePolicy(data []byte, catalog *economy.Catalog, routeCatalog *routes.Catalog) (*RelevancePolicy, error) {
	return relevancepolicy.Load(data, catalog, routeCatalog, true)
}

func loadRelevancePolicy(data []byte, catalog *economy.Catalog, routeCatalog *routes.Catalog, requireComplete bool) (*RelevancePolicy, error) {
	return relevancepolicy.Load(data, catalog, routeCatalog, requireComplete)
}

func relevanceSafeInteger(value *json.Number) (int64, error) {
	if value == nil {
		return 0, errors.New("missing integer")
	}
	lexical := value.String()
	if len(lexical) > 64 {
		return 0, errors.New("integer lexical form is too long")
	}
	if exponentAt := strings.IndexAny(lexical, "eE"); exponentAt >= 0 {
		exponentLexical := lexical[exponentAt+1:]
		if len(strings.TrimLeft(exponentLexical, "+-")) > 2 {
			return 0, errors.New("integer exponent is too large")
		}
		exponent, err := strconv.Atoi(exponentLexical)
		if err != nil || exponent < -32 || exponent > 32 {
			return 0, errors.New("integer exponent is too large")
		}
	}
	rational, ok := new(big.Rat).SetString(lexical)
	if !ok || !rational.IsInt() || !rational.Num().IsInt64() {
		return 0, errors.New("expected exact integer")
	}
	result := rational.Num().Int64()
	if result < 0 || result > relevanceMaxSafeInteger {
		return 0, errors.New("integer outside exact range")
	}
	return result, nil
}

func relevanceBoundaryPosition(gate *string, order map[string]int) (int, bool) {
	if gate == nil {
		return -1, true
	}
	position, ok := order[*gate]
	return position, ok
}

func sortedUniqueIDs(values []string) error {
	for index, value := range values {
		if !relevanceIDPattern.MatchString(value) || index > 0 && values[index-1] >= value {
			return errors.New("values must be raw-byte sorted unique ids")
		}
	}
	return nil
}

func containsSorted(values []string, target string) bool {
	index := sort.SearchStrings(values, target)
	return index < len(values) && values[index] == target
}

func decodeRelevanceStrict(data []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}
