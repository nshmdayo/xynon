package main

import (
	"context"
	"net/http"

	hplugin "github.com/hashicorp/go-plugin"
	"github.com/nshmdayo/xynon/internal/plugin"
	"github.com/nshmdayo/xynon/internal/plugin/abi"
)

type dummyHandler struct{}

func (d *dummyHandler) Name() string {
	return "dummy"
}

func (d *dummyHandler) OnRequest(ctx context.Context, header http.Header) (abi.Action, bool, int, error) {
	header.Set("X-Dummy-Req", "ok")
	return abi.ActionContinue, false, 0, nil
}

func (d *dummyHandler) OnResponse(ctx context.Context, header http.Header, statusCode int) (abi.Action, bool, int, error) {
	header.Set("X-Dummy-Res", "ok")
	return abi.ActionContinue, false, 0, nil
}

func (d *dummyHandler) Close(ctx context.Context) error {
	return nil
}

func main() {
	hplugin.Serve(&hplugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig,
		Plugins: map[string]hplugin.Plugin{
			"handler": &plugin.XynonPlugin{Impl: &dummyHandler{}},
		},
	})
}
