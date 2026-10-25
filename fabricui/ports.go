// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"time"

	"github.com/SukramJ/go-fabric/node"
)

// nodePort is what the handlers read from the node, already in the shape
// they serve. nodeAdapter implements it over a [NodeSource]; the handler
// tests implement it with a fake, as loom's handler tests fake each narrow
// port (internal/north/rest/handlers/matter_test.go at b40b289a).
type nodePort interface {
	health() (healthy bool, note string, ok bool)
	localAddr() string
	fabrics(ctx context.Context) ([]node.FabricInfo, error)
	removeFabric(ctx context.Context, fabricIndex uint8) error
	// window is the commissioning window as the node serves it now.
	window() node.WindowSnapshot
	openBasicWindow(ctx context.Context, creds node.WindowCredentials) error
	openOwnWindow(ctx context.Context, timeout time.Duration) error
	closeWindow(ctx context.Context) error
	setupPayload() (node.SetupPayload, error)
	pairingCodesFor(creds node.WindowCredentials) (qr, manual string, err error)
	topology() node.BridgeTopology
	sessions() node.SessionDiagnostics
	// now is the node's clock, the one the window's deadline runs on.
	now() time.Time
	// observe registers o on the node until the returned function runs.
	observe(o node.Observers) (unsubscribe func())
	mdns() mdnsResponse
	diagnosticEvents() []node.DiagnosticEvent
}

// nodeAdapter reads a [NodeSource] through the facade only.
type nodeAdapter struct{ n NodeSource }

func (a nodeAdapter) health() (healthy bool, note string, ok bool) {
	s, ok := a.n.Health()
	return s.Healthy, s.Note, ok
}

func (a nodeAdapter) localAddr() string { return a.n.LocalAddr() }

func (a nodeAdapter) fabrics(ctx context.Context) ([]node.FabricInfo, error) {
	return a.n.FabricInfos(ctx)
}

func (a nodeAdapter) removeFabric(ctx context.Context, fabricIndex uint8) error {
	return a.n.RemoveFabric(ctx, fabricIndex)
}

// window is node.Window: Open is true for whoever opened the window, the
// node's own included, which the AdministratorCommissioning snapshot
// reports as WindowNotOpen (matter.js keeps those attributes for
// administrator windows only).
func (a nodeAdapter) window() node.WindowSnapshot { return a.n.Window() }

func (a nodeAdapter) openBasicWindow(ctx context.Context, creds node.WindowCredentials) error {
	return a.n.OpenBasicWindow(ctx, creds)
}

func (a nodeAdapter) openOwnWindow(ctx context.Context, timeout time.Duration) error {
	w := a.n.CommissioningWindow()
	if w == nil {
		return node.ErrCommissioningWindowNotConfigured
	}
	return w.OpenOwnWindow(ctx, timeout)
}

func (a nodeAdapter) closeWindow(ctx context.Context) error {
	w := a.n.CommissioningWindow()
	if w == nil {
		return nil
	}
	return w.RevokeWindow(ctx)
}

func (a nodeAdapter) setupPayload() (node.SetupPayload, error) { return a.n.SetupPayload() }

func (a nodeAdapter) pairingCodesFor(creds node.WindowCredentials) (qr, manual string, err error) {
	return a.n.PairingCodesFor(creds)
}

func (a nodeAdapter) topology() node.BridgeTopology { return a.n.Topology() }

func (a nodeAdapter) sessions() node.SessionDiagnostics { return a.n.SessionDiagnostics() }

func (a nodeAdapter) now() time.Time { return a.n.Clock().Now() }

func (a nodeAdapter) observe(o node.Observers) (unsubscribe func()) { return a.n.Observe(o) }

// mdns reads the live advertisement and its diagnosis. Mirrors loom
// cmd/openccu-loom/daemon_matter.go:1255-1305 (matterMdnsReporter), on the
// facade's MDNSEnabled, MDNSServices and DiagnoseMDNS.
func (a nodeAdapter) mdns() mdnsResponse {
	out := mdnsResponse{Services: []mdnsService{}, Findings: []mdnsFinding{}}
	out.Advertising = a.n.MDNSEnabled()
	services := a.n.MDNSServices()
	for i := range services {
		svc := &services[i]
		addrs := make([]string, 0, len(svc.Addresses))
		for _, ip := range svc.Addresses {
			addrs = append(addrs, ip.String())
		}
		txt := make(map[string]string, len(svc.TXT))
		for _, rec := range svc.TXT {
			txt[rec.Key] = rec.Value
		}
		subtypes := append([]string{}, svc.Subtypes...)
		out.Services = append(out.Services, mdnsService{
			ServiceType:  svc.ServiceType,
			InstanceName: svc.InstanceName,
			HostName:     svc.HostName,
			Port:         svc.Port,
			Addresses:    addrs,
			Subtypes:     subtypes,
			TXT:          txt,
		})
	}
	if !out.Advertising {
		return out
	}
	for _, f := range a.n.DiagnoseMDNS() {
		out.Findings = append(out.Findings, mdnsFinding{
			Severity: string(f.Severity),
			Code:     f.Code,
			Message:  f.Message,
			Service:  f.Service,
		})
	}
	return out
}

func (a nodeAdapter) diagnosticEvents() []node.DiagnosticEvent { return a.n.DiagnosticEvents() }
