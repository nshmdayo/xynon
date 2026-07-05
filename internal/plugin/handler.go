package plugin

import (
	"context"
	"net/http"

	"github.com/nshmdayo/xynon/internal/plugin/abi"
)

// Handler abstracts both WASM and RPC plugins.
type Handler interface {
	Name() string
	OnRequest(ctx context.Context, header http.Header) (abi.Action, bool, int, error)
	OnResponse(ctx context.Context, header http.Header, statusCode int) (abi.Action, bool, int, error)
	Close(ctx context.Context) error
}
