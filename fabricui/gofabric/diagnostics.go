// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import (
	"github.com/SukramJ/go-fabric/diag"

	"github.com/SukramJ/go-fabric-ui/fabricui"
)

// The controller's diagnostics (contract 0.9.0) map one to one from
// go-fabric's Controller.SessionDiagnostics, DiagnoseMDNS,
// DiagnosticEvents and OnDiagnosticEvent (controller/diagnostics.go at
// v0.21.0) onto the port's own types.

// Diagnostics implements [fabricui.ControllerSource]: go-fabric's
// snapshot of every peer's sessions and connection state.
func (s *source) Diagnostics() (fabricui.ControllerDiagnostics, bool) {
	return diagnosticsOf(s.ctrl.SessionDiagnostics()), true
}

// diagnosticsOf is go-fabric's snapshot as the port's.
func diagnosticsOf(d diag.SessionDiagnostics) fabricui.ControllerDiagnostics {
	out := fabricui.ControllerDiagnostics{
		AsOf: d.AsOf,
		Occupancy: fabricui.SessionOccupancy{
			Live: d.Occupancy.Live, Reserved: d.Occupancy.Reserved, Capacity: d.Occupancy.Capacity, Free: d.Occupancy.Free,
		},
		Sessions: make([]fabricui.ControllerSession, 0, len(d.Sessions)),
		Peers:    make([]fabricui.ControllerPeer, 0, len(d.Peers)),
	}
	for i := range d.Sessions {
		x := &d.Sessions[i]
		out.Sessions = append(out.Sessions, fabricui.ControllerSession{
			ID: x.ID, FabricIndex: x.FabricIndex, PeerNodeID: x.PeerNodeID, LocalNodeID: x.LocalNodeID,
			PASE: x.PASE, Subscriptions: x.Subscriptions,
			LastActivity: x.LastActivity, LastPeerActivity: x.LastPeerActivity, Idle: x.Idle, PeerIdle: x.PeerIdle,
			CreatedAt: x.CreatedAt, Resumed: x.Resumed, Transport: x.Transport, PeerLost: x.PeerLost,
		})
	}
	for i := range d.Peers {
		p := &d.Peers[i]
		out.Peers = append(out.Peers, fabricui.ControllerPeer{
			FabricIndex: p.FabricIndex, NodeID: p.NodeID, Registered: p.Registered,
			State: peerStateOf(p.State), LoadError: p.LoadError, TimeOffline: p.TimeOffline,
			OperationalAddress: p.OperationalAddress, OperationalHost: p.OperationalHost,
			TCPUnsupported: p.TCPUnsupported, Reachable: p.Reachable, DiscoveredNames: p.DiscoveredNames,
			LastCheckInAt: p.LastCheckInAt, NextExpectedCheckIn: p.NextExpectedCheckIn,
			SubscriptionActive: p.SubscriptionActive, SessionIDs: p.SessionIDs,
		})
	}
	return out
}

// peerStateOf is diag.PeerInfo.State — client.ConnectionState's String —
// as the port's connection state; "" without a registered node, and for a
// name go-fabric adds later.
func peerStateOf(s string) fabricui.ConnectionState {
	switch s {
	case "Connected":
		return fabricui.ConnectionConnected
	case "Disconnected":
		return fabricui.ConnectionDisconnected
	case "Reconnecting":
		return fabricui.ConnectionReconnecting
	case "WaitingForDeviceDiscovery":
		return fabricui.ConnectionWaitingForDiscovery
	}
	return ""
}

// DiagnoseMDNS implements [fabricui.ControllerSource]. go-fabric's call
// waits for the DNS-SD cache, so it must not run from a callback the cache
// runs: the handler calls it only from GET /controller/mdns.
func (s *source) DiagnoseMDNS() []fabricui.MDNSFinding {
	findings := s.ctrl.DiagnoseMDNS()
	out := make([]fabricui.MDNSFinding, 0, len(findings))
	for _, f := range findings {
		out = append(out, fabricui.MDNSFinding{Severity: string(f.Severity), Code: f.Code, Message: f.Message, Service: f.Service})
	}
	return out
}

// DiagnosticEvents implements [fabricui.ControllerSource]: the trace,
// oldest first.
func (s *source) DiagnosticEvents() []fabricui.DiagnosticEvent {
	events := s.ctrl.DiagnosticEvents()
	out := make([]fabricui.DiagnosticEvent, 0, len(events))
	for i := range events {
		out = append(out, diagnosticEventOf(&events[i]))
	}
	return out
}

// OnDiagnosticEvent implements [fabricui.ControllerSource]. go-fabric runs
// fn inline on the goroutine that recorded the event; the handler's fn
// only queues it.
func (s *source) OnDiagnosticEvent(fn func(fabricui.DiagnosticEvent)) (unsubscribe func()) {
	return s.ctrl.OnDiagnosticEvent(func(e diag.Event) { fn(diagnosticEventOf(&e)) })
}

func diagnosticEventOf(e *diag.Event) fabricui.DiagnosticEvent {
	return fabricui.DiagnosticEvent{
		At: e.At, Kind: string(e.Kind), Severity: string(e.Severity), Message: e.Message, Detail: e.Detail,
	}
}
