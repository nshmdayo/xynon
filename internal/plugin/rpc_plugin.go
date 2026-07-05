package plugin

import (
	"context"
	"fmt"
	"net/http"
	"net/rpc"
	"os/exec"

	hplugin "github.com/hashicorp/go-plugin"
	"github.com/nshmdayo/xynon/internal/plugin/abi"
)

// Types for RPC requests and responses
type OnRequestArgs struct {
	Header http.Header
}

type OnRequestReply struct {
	Action       abi.Action
	ShortCircuit bool
	StatusCode   int
	Header       http.Header
}

type OnResponseArgs struct {
	Header     http.Header
	StatusCode int
}

type OnResponseReply struct {
	Action       abi.Action
	ShortCircuit bool
	StatusCode   int
	Header       http.Header
}

// HandlerRPCServer is the RPC server that exposes the plugin.Handler implementation over RPC.
type HandlerRPCServer struct {
	Impl Handler
}

func (s *HandlerRPCServer) OnRequest(args *OnRequestArgs, reply *OnRequestReply) error {
	action, sc, scStatus, err := s.Impl.OnRequest(context.Background(), args.Header)
	reply.Action = action
	reply.ShortCircuit = sc
	reply.StatusCode = scStatus
	reply.Header = args.Header
	return err
}

func (s *HandlerRPCServer) OnResponse(args *OnResponseArgs, reply *OnResponseReply) error {
	action, sc, scStatus, err := s.Impl.OnResponse(context.Background(), args.Header, args.StatusCode)
	reply.Action = action
	reply.ShortCircuit = sc
	reply.StatusCode = scStatus
	reply.Header = args.Header
	return err
}

// HandlerRPCClient is the RPC client that implements the plugin.Handler interface by forwarding to RPC Server.
type HandlerRPCClient struct {
	client *rpc.Client
	name   string
	hc     *hplugin.Client
}

func (c *HandlerRPCClient) Name() string {
	return c.name
}

func (c *HandlerRPCClient) OnRequest(ctx context.Context, header http.Header) (abi.Action, bool, int, error) {
	args := &OnRequestArgs{Header: header}
	var reply OnRequestReply
	err := c.client.Call("Plugin.OnRequest", args, &reply)
	if err != nil {
		return abi.ActionError, false, 0, err
	}
	// Apply header changes
	for k := range header {
		delete(header, k)
	}
	for k, v := range reply.Header {
		header[k] = v
	}
	return reply.Action, reply.ShortCircuit, reply.StatusCode, nil
}

func (c *HandlerRPCClient) OnResponse(ctx context.Context, header http.Header, statusCode int) (abi.Action, bool, int, error) {
	args := &OnResponseArgs{Header: header, StatusCode: statusCode}
	var reply OnResponseReply
	err := c.client.Call("Plugin.OnResponse", args, &reply)
	if err != nil {
		return abi.ActionError, false, 0, err
	}
	for k := range header {
		delete(header, k)
	}
	for k, v := range reply.Header {
		header[k] = v
	}
	return reply.Action, reply.ShortCircuit, reply.StatusCode, nil
}

func (c *HandlerRPCClient) Close(ctx context.Context) error {
	c.hc.Kill()
	return nil
}

// XynonPlugin is the implementation of hashicorp plugin.Plugin interface
type XynonPlugin struct {
	Impl Handler
}

func (p *XynonPlugin) Server(*hplugin.MuxBroker) (interface{}, error) {
	return &HandlerRPCServer{Impl: p.Impl}, nil
}

func (p *XynonPlugin) Client(b *hplugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &HandlerRPCClient{client: c}, nil
}

// HandshakeConfig is a common handshake configuration used by client and server
var HandshakeConfig = hplugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "XYNON_PLUGIN",
	MagicCookieValue: "hello",
}

// PluginMap is the map of plugins we can dispense
var PluginMap = map[string]hplugin.Plugin{
	"handler": &XynonPlugin{},
}

// LoadRpcPlugin launches the RPC subprocess and returns a Handler wrapper.
func LoadRpcPlugin(ctx context.Context, name, path string) (Handler, error) {
	client := hplugin.NewClient(&hplugin.ClientConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins:         PluginMap,
		Cmd:             exec.Command(path),
		Managed:         true,
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("rpc client start: %w", err)
	}

	raw, err := rpcClient.Dispense("handler")
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("rpc dispense: %w", err)
	}

	handlerClient := raw.(*HandlerRPCClient)
	handlerClient.name = name
	handlerClient.hc = client
	return handlerClient, nil
}
