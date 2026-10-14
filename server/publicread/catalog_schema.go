package publicread

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"cloud-clicker/server/publicapi"
)

const (
	CatalogArtifactSchemaName = "PublicCatalogArtifact"
	CatalogBundleSchemaName   = "PublicCatalogBundle"
)

var ErrCatalogSchemaComposition = errors.New("invalid public catalog schema composition")

// CatalogSchemaOwner binds one served family to its owner's exported root and
// definitions. Shared definitions have one declaring owner; other owners may
// reference them, but may not redeclare them, even with identical values.
type CatalogSchemaOwner struct {
	Name        string
	Root        string
	Definitions []publicapi.NamedSchema
}

type CatalogSchemaCoverageError struct {
	Missing []string
}

func (err *CatalogSchemaCoverageError) Error() string {
	return ErrCatalogSchemaComposition.Error() + ": missing owners: " + strings.Join(err.Missing, ", ")
}

func (err *CatalogSchemaCoverageError) Unwrap() error { return ErrCatalogSchemaComposition }

// ComposeCatalogSchemas implements C12/C18's closed name-discriminated union.
// The composition root must supply the complete authoritative served-family
// set, including historical families, not a convenient subset of available
// providers. This constructor checks coverage of that set; it does not discover
// database inventory or prove that caller-supplied descriptors are owner exports.
//
// No schemas are returned on failure. This does not mount an endpoint or weaken
// existing owner semantic/byte-identity checks. A historical bundle may contain
// a subset of these arms; its exact membership/hash remains the reader's job.
func ComposeCatalogSchemas(servedFamilies []string, owners []CatalogSchemaOwner) ([]publicapi.NamedSchema, error) {
	if len(servedFamilies) == 0 {
		return nil, fmt.Errorf("%w: empty served-family set", ErrCatalogSchemaComposition)
	}
	families := append([]string(nil), servedFamilies...)
	sort.Strings(families)
	nameDefinitions := map[string]*publicapi.Schema{"CatalogFamilyName": stringSchema("mechanical-id")}
	wanted := make(map[string]bool, len(families))
	for _, name := range families {
		data, err := json.Marshal(name)
		if err != nil || publicapi.ValidateJSON("CatalogFamilyName", data, nameDefinitions) != nil || wanted[name] {
			return nil, fmt.Errorf("%w: invalid or repeated served family", ErrCatalogSchemaComposition)
		}
		wanted[name] = true
	}
	provided := make(map[string]CatalogSchemaOwner, len(owners))
	definitions := []publicapi.NamedSchema{}
	declared := map[string]bool{CatalogArtifactSchemaName: true, CatalogBundleSchemaName: true}
	for _, owner := range owners {
		if !wanted[owner.Name] {
			return nil, fmt.Errorf("%w: owner outside served set", ErrCatalogSchemaComposition)
		}
		if _, duplicate := provided[owner.Name]; duplicate {
			return nil, fmt.Errorf("%w: repeated owner", ErrCatalogSchemaComposition)
		}
		rootDeclared := false
		for _, definition := range owner.Definitions {
			if declared[definition.Name] {
				return nil, fmt.Errorf("%w: repeated or reserved definition %s", ErrCatalogSchemaComposition, definition.Name)
			}
			declared[definition.Name] = true
			if definition.Name == owner.Root && definition.Schema != nil {
				rootDeclared = true
			}
			definitions = append(definitions, definition)
		}
		if !rootDeclared {
			return nil, fmt.Errorf("%w: root is not declared by its owner", ErrCatalogSchemaComposition)
		}
		provided[owner.Name] = owner
	}
	missing := []string{}
	for _, name := range families {
		if _, present := provided[name]; !present {
			missing = append(missing, name)
		}
	}
	if len(missing) != 0 {
		return nil, &CatalogSchemaCoverageError{Missing: missing}
	}
	ref := func(name string) *publicapi.Schema { return &publicapi.Schema{Kind: publicapi.SchemaRef, Ref: name} }
	alternates := make([]*publicapi.Schema, 0, len(families))
	for _, name := range families {
		alternates = append(alternates, &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{
			field("json", ref(provided[name].Root)),
			field("name", &publicapi.Schema{Kind: publicapi.SchemaString, Enum: []string{name}}),
			field("sha256", stringSchema("sha256-prefixed")),
		}})
	}
	definitions = append(definitions,
		publicapi.NamedSchema{Name: CatalogArtifactSchemaName, Schema: &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: alternates}},
		publicapi.NamedSchema{Name: CatalogBundleSchemaName, Schema: &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{
			field("artifacts", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: ref(CatalogArtifactSchemaName)}),
			field("constants_hash", stringSchema("sha256-prefixed")),
		}}},
	)
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].Name < definitions[j].Name })
	validated, err := publicapi.ValidateSchemaDefinitions(definitions)
	if err != nil {
		return nil, errors.Join(ErrCatalogSchemaComposition, err)
	}
	// The validator detached every node/bound/enum. Return that snapshot instead
	// of aliased constructor arguments; later calls remain independently owned.
	result := make([]publicapi.NamedSchema, 0, len(definitions))
	for _, definition := range definitions {
		result = append(result, publicapi.NamedSchema{Name: definition.Name, Schema: validated[definition.Name]})
	}
	return result, nil
}
