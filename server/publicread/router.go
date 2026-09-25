package publicread

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"cloud-clicker/server/httpapi"
	"cloud-clicker/server/publicapi"
)

// ErrComposition reports an invalid public-router composition. Startup fails
// rather than serving a partially configured public surface.
var ErrComposition = errors.New("invalid public read composition")

// CursorKeys is the deployment cursor pair (Deployment Foundation). Previous is
// optional as a pair.
type CursorKeys struct {
	CurrentID  string
	Current    []byte
	PreviousID string
	Previous   []byte
}

// CursorSecretResolver maps the policy's named cursor keys (C20) onto the
// deployment pair by ID. When the deployment carries no previous pair, the
// policy's previous name resolves to the current secret: C20 permits both
// names to share one value at first deployment, while the deployment decoder
// forbids a previous value equal to the current one. Any other name is
// unresolvable, so ResolveCursorCodec fails startup.
func CursorSecretResolver(policy publicapi.Policy, keys CursorKeys) publicapi.SecretResolver {
	return func(id string) ([]byte, bool) {
		switch {
		case id == "":
			return nil, false
		case id == keys.CurrentID && len(keys.Current) != 0:
			return bytes.Clone(keys.Current), true
		case keys.PreviousID != "" && id == keys.PreviousID && len(keys.Previous) != 0:
			return bytes.Clone(keys.Previous), true
		case keys.PreviousID == "" && id == policy.CursorKeyIDs.Previous && keys.CurrentID == policy.CursorKeyIDs.Current && len(keys.Current) != 0:
			return bytes.Clone(keys.Current), true
		}
		return nil, false
	}
}

// Dependencies are the narrow sources the public router reads (C10).
type Dependencies struct {
	PolicyJSON []byte
	CursorKeys CursorKeys
	Epochs     EpochReader
	Boards     BoardReader
	Clock      func() time.Time
	Random     io.Reader
}

var notFound = []byte(`{"category":"unknown_id","detail":"route"}` + "\n")

// NewRouter mounts every public operation from the public registry, and only
// those, behind the shared request-ID middleware.
func NewRouter(dependencies Dependencies) (http.Handler, error) {
	if dependencies.Epochs == nil || dependencies.Boards == nil || dependencies.Clock == nil || dependencies.Random == nil {
		return nil, ErrComposition
	}
	policy, err := publicapi.LoadPolicy(dependencies.PolicyJSON)
	if err != nil {
		return nil, errors.Join(ErrComposition, err)
	}
	registry, err := Registry()
	if err != nil {
		return nil, errors.Join(ErrComposition, err)
	}
	cursors, err := publicapi.ResolveCursorCodec(policy, registry, CursorSecretResolver(policy, dependencies.CursorKeys))
	if err != nil {
		return nil, errors.Join(ErrComposition, err)
	}
	requestIDs, err := httpapi.NewRequestIDs(policy.RequestID.Pattern, policy.RequestID.MaxBytes, dependencies.Random)
	if err != nil {
		return nil, errors.Join(ErrComposition, err)
	}
	runtime, err := publicapi.NewRuntime(policy, dependencies.Clock, requestIDs)
	if err != nil {
		return nil, errors.Join(ErrComposition, err)
	}
	router := chi.NewRouter()
	router.NotFound(func(response http.ResponseWriter, _ *http.Request) {
		writeExact(response, http.StatusNotFound, notFound)
	})
	router.MethodNotAllowed(func(response http.ResponseWriter, _ *http.Request) {
		writeExact(response, http.StatusNotFound, notFound)
	})
	bindings := []publicapi.Binding{
		{OperationID: ListBoardOperation, Handler: BoardHandler{Registry: registry, Cursors: cursors, Runtime: runtime, Boards: dependencies.Boards}},
		{OperationID: ListEpochsOperation, Handler: EpochsHandler{Registry: registry, Cursors: cursors, Runtime: runtime, Epochs: dependencies.Epochs}},
	}
	if err := registry.Mount(router, bindings, nil); err != nil {
		return nil, errors.Join(ErrComposition, err)
	}
	return runtime.WithRequestID(router), nil
}
