package main

import (
	"context"
	"errors"
	"net/http"

	hplugin "github.com/hashicorp/go-plugin"
	"github.com/nshmdayo/xynon/internal/plugin"
	"github.com/nshmdayo/xynon/internal/plugin/abi"
)

type DummyHandler struct{}

func (d *DummyHandler) Name() string {
	return "dummy_rpc"
}

func (d *DummyHandler) OnRequest(ctx context.Context, header http.Header) (abi.Action, bool, int, error) {
	header.Set("X-Dummy-Request", "handled")
	if header.Get("X-Fail") == "true" {
		return abi.ActionError, false, 0, errors.New("simulated request error")
	}
	if header.Get("X-Short") == "true" {
		return abi.ActionContinue, true, 403, nil
	}
	return abi.ActionContinue, false, 0, nil
}

func (d *DummyHandler) OnResponse(ctx context.Context, header http.Header, statusCode int) (abi.Action, bool, int, error) {
	header.Set("X-Dummy-Response", "handled")
	if header.Get("X-Fail") == "true" {
		return abi.ActionError, false, 0, errors.New("simulated response error")
	}
	return abi.ActionContinue, false, 0, nil
}

func (d *DummyHandler) Close(ctx context.Context) error {
	return nil
}

func main() {
	hplugin.Serve(&hplugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig,
		Plugins: map[string]hplugin.Plugin{
			"handler": &plugin.XynonPlugin{Impl: &DummyHandler{}},
		},
	})
}
