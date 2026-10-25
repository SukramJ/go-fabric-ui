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
	// window reports whether a commissioning window is open and the
	// duration it was opened for.
	window() (open bool, durationSeconds uint16)
	openBasicWindow(ctx context.Context, creds node.WindowCredentials) error
	openOwnWindow(ctx context.Context, timeout time.Duration) error
	closeWindow(ctx context.Context) error
	pairingCodes() (qr, manual string, err error)
	pairingCodesFor(creds node.WindowCredentials) (qr, manual string, err error)
	topology() node.BridgeTopology
	sessions() ([]sessionInfo, occupancyResponse)
	mdns() mdnsResponse
	diagnosticEvents() []node.DiagnosticEvent
}

// sessionInfo is one open secure session, without key material. Mirrors
// loom's MatterSessionInfo (internal/north/rest/handlers/matter.go:305-318).
type sessionInfo struct {
	SessionID        uint16
	FabricIndex      uint8
	PeerNodeID       uint64
	LocalNodeID      uint64
	IsPASE           bool
	Subscriptions    int
	LastActivity     time.Time
	LastPeerActivity time.Time
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

// window reads IsOpen rather than the AdministratorCommissioning snapshot:
// the node's own window reads WindowNotOpen there (matter.js keeps those
// attributes for administrator windows only), and the status would call a
// live window closed — loom cmd/openccu-loom/matter_status_adapter.go:63-74.
func (a nodeAdapter) window() (open bool, durationSeconds uint16) {
	w := a.n.CommissioningWindow()
	if w == nil {
		return false, 0
	}
	return w.IsOpen(), w.RequestedDurationSeconds()
}

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

func (a nodeAdapter) pairingCodes() (qr, manual string, err error) { return a.n.PairingCodes() }

func (a nodeAdapter) pairingCodesFor(creds node.WindowCredentials) (qr, manual string, err error) {
	return a.n.PairingCodesFor(creds)
}

func (a nodeAdapter) topology() node.BridgeTopology { return a.n.Topology() }

// sessions joins the session table with the subscription counts per
// session. Mirrors loom cmd/openccu-loom/daemon_matter.go:1198-1240
// (matterSessionLister), reading the facade's Sessions and Subscriptions.
func (a nodeAdapter) sessions() ([]sessionInfo, occupancyResponse) {
	op := a.n.Sessions()
	if op == nil {
		return nil, occupancyResponse{}
	}
	var counts map[uint16]int
	if sub := a.n.Subscriptions(); sub != nil {
		counts = sub.CountBySession()
	}
	list := op.Sessions()
	out := make([]sessionInfo, 0, len(list))
	for _, s := range list {
		out = append(out, sessionInfo{
			SessionID:        s.SessionID,
			FabricIndex:      s.FabricIndex,
			PeerNodeID:       s.PeerNodeID,
			LocalNodeID:      s.LocalNodeID,
			IsPASE:           s.IsPASE,
			Subscriptions:    counts[s.SessionID],
			LastActivity:     s.LastActivity,
			LastPeerActivity: s.LastPeerActivity,
		})
	}
	occ := op.Occupancy()
	return out, occupancyResponse{Live: occ.Live, Reserved: occ.Reserved, Capacity: occ.Capacity, Free: occ.Free()}
}

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
