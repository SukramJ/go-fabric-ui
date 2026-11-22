// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"testing"
)

// onboardingHandler is a handler over a fresh in-memory controller in host
// mode, whose admin check admits every request with the request header:
// the onboarding inventory is served only with Options.Admin (owner
// decision Q2 of brief U12).
func onboardingHandler(t *testing.T) (*http.ServeMux, *MemoryController) {
	t.Helper()
	m := NewMemoryController()
	_, mux := newTestHandler(t, testOptions{noNode: true, controller: m, admin: func(*http.Request) bool { return true }})
	return mux, m
}

// withInventory pairs the commissionable device by its QR code (node 0x13,
// its code kept as paired) and attaches a manual code to the sensor (node
// 0x12); the light (0x11) keeps none.
func withInventory(t *testing.T, m *MemoryController) {
	t.Helper()
	if _, err := m.Pair(context.Background(), memoryQRCode, PairOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.SetOnboardingCode(context.Background(), 0x12, "3497-011-2332"); err != nil {
		t.Fatal(err)
	}
}

// The onboarding fixtures (contract 0.13.0), each against a fresh
// in-memory controller.
func TestOnboardingFixtures(t *testing.T) {
	for _, c := range []struct {
		name  string
		setup func(*testing.T, *MemoryController)
		token bool
	}{
		{name: "controller-inventory.200", setup: withInventory},
		{name: "controller-inventory.403"},
		{name: "controller-inventory.501", token: true},
		{name: "controller-node-onboarding-code-set.200"},
		{name: "controller-node-onboarding-code-set.400-invalid"},
		{name: "controller-node-onboarding-code-set.409-mismatch", setup: func(_ *testing.T, m *MemoryController) {
			m.OnboardingMismatch = &OnboardingMismatch{Field: OnboardingFieldProduct, Code: 0x8000, Node: 0x8001}
		}},
		{name: "controller-node-onboarding-code-set.404"},
		{name: "controller-node-onboarding-code-clear.204", setup: withInventory},
		{name: "controller-nodes.200-onboarding", setup: withInventory},
		{name: "controller-status.200-onboarding"},
	} {
		t.Run(c.name, func(t *testing.T) {
			mux, m := onboardingHandler(t)
			if c.token {
				_, mux, m = controllerHandler(t, nil)
			}
			if c.setup != nil {
				c.setup(t, m)
			}
			runFixture(t, mux, c.name)
		})
	}
}

// The inventory is by node id; a paired QR code has both forms, an
// attached manual code no QR code; a decommissioned node is gone with its
// code, and DELETE removes a code and answers 204 again without one.
func TestOnboarding_InventoryLifecycle(t *testing.T) {
	mux, m := onboardingHandler(t)
	withInventory(t, m)
	var inv struct {
		Entries []struct {
			NodeIDHex  string  `json:"node_id_hex"`
			HasCode    bool    `json:"has_code"`
			QRCode     *string `json:"qr_code"`
			ManualCode *string `json:"manual_code"`
			Source     string  `json:"source"`
		} `json:"entries"`
	}
	get := func() {
		t.Helper()
		rec := do(t, mux, http.MethodGet, "/controller/inventory", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("inventory %d %s", rec.Code, rec.Body)
		}
		inv.Entries = nil
		if err := json.Unmarshal(rec.Body.Bytes(), &inv); err != nil {
			t.Fatal(err)
		}
	}
	get()
	ids := make([]string, 0, len(inv.Entries))
	for _, e := range inv.Entries {
		ids = append(ids, e.NodeIDHex)
	}
	if fmt.Sprint(ids) != "[0000000000000011 0000000000000012 0000000000000013]" {
		t.Fatalf("order %v", ids)
	}
	if e := inv.Entries[2]; !e.HasCode || e.QRCode == nil || e.ManualCode == nil || e.Source != "paired" {
		t.Fatalf("paired entry %+v", e)
	}
	if e := inv.Entries[1]; !e.HasCode || e.QRCode != nil || e.ManualCode == nil || *e.ManualCode != "34970112332" || e.Source != "attached" {
		t.Fatalf("attached entry %+v", e)
	}
	if e := inv.Entries[0]; e.HasCode || e.ManualCode != nil {
		t.Fatalf("entry without a code %+v", e)
	}
	for range 2 {
		if rec := do(t, mux, http.MethodDelete, "/controller/nodes/0000000000000012/onboarding-code", ""); rec.Code != http.StatusNoContent {
			t.Fatalf("clear %d %s", rec.Code, rec.Body)
		}
	}
	if rec := do(t, mux, http.MethodDelete, "/controller/nodes/0000000000000013", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("decommission %d %s", rec.Code, rec.Body)
	}
	get()
	if len(inv.Entries) != 2 || inv.Entries[1].HasCode {
		t.Fatalf("after clear and decommission %+v", inv.Entries)
	}
	// The decommission dropped the node's code in the port (U12-R-05),
	// and the node takes none any more.
	m.mu.Lock()
	_, kept := m.codes[0x13]
	m.mu.Unlock()
	if kept {
		t.Fatal("the memory port kept a decommissioned node's code")
	}
	if e, err := m.SetOnboardingCode(context.Background(), 0x13, memoryManualCode); !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("a decommissioned node took a code: %v, %v", e, err)
	}
}

// U12-R-06: the very code a node was paired with stays paired when it is
// attached again; another code replaces it as attached.
func TestOnboarding_PairedCodeStaysPaired(t *testing.T) {
	mux, m := onboardingHandler(t)
	withInventory(t, m)
	var e struct {
		Source string `json:"source"`
	}
	rec := do(t, mux, http.MethodPut, "/controller/nodes/0000000000000013/onboarding-code", `{"code":"`+memoryQRCode+`"}`)
	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil || rec.Code != http.StatusOK || e.Source != "paired" {
		t.Fatalf("the paired code again: %d %s", rec.Code, rec.Body)
	}
	rec = do(t, mux, http.MethodPut, "/controller/nodes/0000000000000013/onboarding-code", `{"code":"1111-222-3333"}`)
	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil || rec.Code != http.StatusOK || e.Source != "attached" {
		t.Fatalf("another code: %d %s", rec.Code, rec.Body)
	}
}

// U12-R-04: the contract's maxLength 512 of OnboardingCodeRequest.code is
// enforced: a longer code is 400 bad_request and reaches no source.
func TestOnboarding_CodeMaxLength(t *testing.T) {
	src := &failingOnboarding{MemoryController: NewMemoryController(), fail: true}
	_, mux := newTestHandler(t, testOptions{noNode: true, controller: src, admin: func(*http.Request) bool { return true }})
	const path = "/controller/nodes/0000000000000011/onboarding-code"
	long := strings.Repeat("ä", 513)
	rec := do(t, mux, http.MethodPut, path, `{"code":"`+long+`"}`)
	if rec.Code != http.StatusBadRequest || problemOf(t, rec).Code != codeBadRequest || strings.Contains(rec.Body.String(), "ä") {
		t.Fatalf("513 characters: %d %s", rec.Code, rec.Body)
	}
	// 512 characters reach the source (which fails: 500, not 400).
	if rec := do(t, mux, http.MethodPut, path, `{"code":"`+strings.Repeat("ä", 512)+`"}`); rec.Code != http.StatusInternalServerError {
		t.Fatalf("512 characters: %d %s", rec.Code, rec.Body)
	}
}

// has_onboarding_code is on GET /controller/nodes only while the
// inventory is served (decision 8).
func TestOnboarding_NodesHaveTheFlagOnlyWhenServed(t *testing.T) {
	mux, m := onboardingHandler(t)
	withInventory(t, m)
	if body := do(t, mux, http.MethodGet, "/controller/nodes", "").Body.String(); strings.Count(body, `"has_onboarding_code":true`) != 2 || strings.Count(body, `"has_onboarding_code":false`) != 1 {
		t.Fatalf("nodes %s", body)
	}
	m.NoOnboardingCodes = true
	if body := do(t, mux, http.MethodGet, "/controller/nodes", "").Body.String(); strings.Contains(body, "has_onboarding_code") {
		t.Fatalf("nodes without the inventory %s", body)
	}
}

// G4 (owner decision Q2): with Options.Admin nil — token mode, a valid
// token — the three routes answer 501 and the status reports
// onboarding_codes false although the source serves the inventory; in
// host mode without admin rights 403; a source without the capability
// 501 in host mode too.
func TestOnboarding_TokenModeNeverServesACode(t *testing.T) {
	_, mux, m := controllerHandler(t, nil)
	if !m.Capabilities().OnboardingCodes {
		t.Fatal("the memory source serves the inventory")
	}
	for _, c := range []struct{ method, path, body string }{
		{http.MethodGet, "/controller/inventory", ""},
		{http.MethodPut, "/controller/nodes/0000000000000011/onboarding-code", `{"code":"34970112332"}`},
		{http.MethodDelete, "/controller/nodes/0000000000000011/onboarding-code", ""},
	} {
		rec := do(t, mux, c.method, c.path, c.body)
		if rec.Code != http.StatusNotImplemented || problemOf(t, rec).Code != codeNotImplemented {
			t.Fatalf("token mode %s %s: %d %s", c.method, c.path, rec.Code, rec.Body)
		}
	}
	if body := do(t, mux, http.MethodGet, "/controller/status", "").Body.String(); !strings.Contains(body, `"onboarding_codes":false`) {
		t.Fatalf("status %s", body)
	}
	if body := do(t, mux, http.MethodGet, "/controller/nodes", "").Body.String(); strings.Contains(body, "has_onboarding_code") {
		t.Fatalf("nodes %s", body)
	}

	_, denied := newTestHandler(t, testOptions{noNode: true, controller: NewMemoryController(), admin: func(*http.Request) bool { return false }})
	if rec := do(t, denied, http.MethodGet, "/controller/inventory", ""); rec.Code != http.StatusForbidden {
		t.Fatalf("no admin rights: %d %s", rec.Code, rec.Body)
	}
	if body := do(t, denied, http.MethodGet, "/controller/status", "").Body.String(); !strings.Contains(body, `"onboarding_codes":true`) {
		t.Fatalf("the capability is readable without admin rights: %s", body)
	}

	host, hm := onboardingHandler(t)
	hm.NoOnboardingCodes = true
	if rec := do(t, host, http.MethodGet, "/controller/inventory", ""); rec.Code != http.StatusNotImplemented {
		t.Fatalf("no capability: %d %s", rec.Code, rec.Body)
	}
}

// G3: Cache-Control no-store on every answer of the three routes that
// carries or refuses a secret — the 200s of GET and PUT, and the 403, 400,
// 409 and 501 (decision 4: apiHeaders sets it before any route writes).
func TestOnboarding_AnswersAreNoStore(t *testing.T) {
	mux, m := onboardingHandler(t)
	withInventory(t, m)
	check := func(what string, rec interface {
		Result() *http.Response
	}, want int,
	) {
		t.Helper()
		res := rec.Result()
		defer res.Body.Close()
		if res.StatusCode != want || res.Header.Get("Cache-Control") != "no-store" {
			t.Fatalf("%s: %d, Cache-Control %q", what, res.StatusCode, res.Header.Get("Cache-Control"))
		}
	}
	check("GET 200", do(t, mux, http.MethodGet, "/controller/inventory", ""), http.StatusOK)
	check("PUT 200", do(t, mux, http.MethodPut, "/controller/nodes/0000000000000011/onboarding-code", `{"code":"34970112332"}`), http.StatusOK)
	check("PUT 400", do(t, mux, http.MethodPut, "/controller/nodes/0000000000000011/onboarding-code", `{"code":"nope"}`), http.StatusBadRequest)
	m.OnboardingMismatch = &OnboardingMismatch{Field: OnboardingFieldDiscriminator, Code: 3, Node: 15}
	check("PUT 409", do(t, mux, http.MethodPut, "/controller/nodes/0000000000000011/onboarding-code", `{"code":"34970112332"}`), http.StatusConflict)
	check("DELETE 204", do(t, mux, http.MethodDelete, "/controller/nodes/0000000000000011/onboarding-code", ""), http.StatusNoContent)
	_, denied := newTestHandler(t, testOptions{noNode: true, controller: NewMemoryController(), admin: func(*http.Request) bool { return false }})
	check("GET 403", do(t, denied, http.MethodGet, "/controller/inventory", ""), http.StatusForbidden)
	_, token, _ := controllerHandler(t, nil)
	check("GET 501", do(t, token, http.MethodGet, "/controller/inventory", ""), http.StatusNotImplemented)
	check("PUT 501", do(t, token, http.MethodPut, "/controller/nodes/0000000000000011/onboarding-code", `{"code":"34970112332"}`), http.StatusNotImplemented)
	check("DELETE 501", do(t, token, http.MethodDelete, "/controller/nodes/0000000000000011/onboarding-code", ""), http.StatusNotImplemented)
}

// The PUT's body is strict and its code required: an unknown member, an
// empty code or a missing body is 400 bad_request, not
// onboarding_code_invalid.
func TestOnboarding_BadBodies(t *testing.T) {
	mux, _ := onboardingHandler(t)
	for _, body := range []string{`{"code":"34970112332","x":1}`, `{"code":""}`, `{"code":"   "}`, `{}`, `{"code":"34970112332"} {}`} {
		rec := do(t, mux, http.MethodPut, "/controller/nodes/0000000000000011/onboarding-code", body)
		if rec.Code != http.StatusBadRequest || problemOf(t, rec).Code != codeBadRequest {
			t.Fatalf("%s: %d %s", body, rec.Code, rec.Body)
		}
	}
	rec := do(t, mux, http.MethodPut, "/controller/nodes/xyz/onboarding-code", `{"code":"34970112332"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad id: %d %s", rec.Code, rec.Body)
	}
}

// failingOnboarding is a host's own source whose writes fail with errors
// of other kinds: fail makes the next SetOnboardingCode a store failure;
// nodeFailed answers ErrNodeFailed on the PUT and conflict ErrConflict
// on the DELETE.
type failingOnboarding struct {
	*MemoryController
	fail, nodeFailed, conflict bool
	badField                   bool
}

func (f *failingOnboarding) SetOnboardingCode(ctx context.Context, id uint64, code string) (OnboardingEntry, error) {
	switch {
	case f.fail:
		return OnboardingEntry{}, errors.New("store write failed: disk full")
	case f.nodeFailed:
		return OnboardingEntry{}, fmt.Errorf("%w: timeout", ErrNodeFailed)
	case f.badField:
		return OnboardingEntry{}, &OnboardingMismatch{Field: "serial", Code: 1, Node: 2}
	}
	return f.MemoryController.SetOnboardingCode(ctx, id, code)
}

func (f *failingOnboarding) ClearOnboardingCode(ctx context.Context, id uint64) error {
	if f.conflict {
		return fmt.Errorf("%w: busy", ErrConflict)
	}
	return f.MemoryController.ClearOnboardingCode(ctx, id)
}

// G7: the three handlers map errors themselves, closed — a host source's
// ErrNodeFailed on the PUT and ErrConflict on the DELETE answer 500, not
// an undeclared 502 or 409; ErrConflict on the PUT is the declared 409
// conflict with a fixed detail; a mismatch field the contract does not
// name is 500.
func TestOnboarding_ClosedErrorMapping(t *testing.T) {
	src := &failingOnboarding{MemoryController: NewMemoryController()}
	var log bytes.Buffer
	_, mux := newTestHandler(t, testOptions{noNode: true, controller: src, log: &log, admin: func(*http.Request) bool { return true }})
	const path = "/controller/nodes/0000000000000011/onboarding-code"
	src.nodeFailed = true
	if rec := do(t, mux, http.MethodPut, path, `{"code":"34970112332"}`); rec.Code != http.StatusInternalServerError || problemOf(t, rec).Detail != "" {
		t.Fatalf("PUT ErrNodeFailed: %d %s", rec.Code, rec.Body)
	}
	src.nodeFailed, src.badField = false, true
	if rec := do(t, mux, http.MethodPut, path, `{"code":"34970112332"}`); rec.Code != http.StatusInternalServerError {
		t.Fatalf("PUT unknown field: %d %s", rec.Code, rec.Body)
	}
	src.badField, src.conflict = false, true
	if rec := do(t, mux, http.MethodDelete, path, ""); rec.Code != http.StatusInternalServerError {
		t.Fatalf("DELETE ErrConflict: %d %s", rec.Code, rec.Body)
	}
	src.conflict = false
	if rec := do(t, mux, http.MethodDelete, "/controller/nodes/00000000000000AA/onboarding-code", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("DELETE unknown: %d %s", rec.Code, rec.Body)
	}
	if !strings.Contains(log.String(), "busy") {
		t.Fatalf("the 500 is not logged:\n%s", log.String())
	}

	conflict := &conflictOnPut{MemoryController: NewMemoryController()}
	_, mux = newTestHandler(t, testOptions{noNode: true, controller: conflict, admin: func(*http.Request) bool { return true }})
	rec := do(t, mux, http.MethodPut, path, `{"code":"34970112332"}`)
	if p := problemOf(t, rec); rec.Code != http.StatusConflict || p.Code != codeConflict || p.Detail != "the node is no longer commissioned" {
		t.Fatalf("PUT ErrConflict: %d %s", rec.Code, rec.Body)
	}
}

type conflictOnPut struct{ *MemoryController }

func (*conflictOnPut) SetOnboardingCode(context.Context, uint64, string) (OnboardingEntry, error) {
	return OnboardingEntry{}, ErrConflict
}

// A mismatch's problem carries field, code_value and node_value, and a
// detail the handler words from them.
func TestOnboarding_MismatchMembers(t *testing.T) {
	mux, m := onboardingHandler(t)
	for _, c := range []struct {
		mm     OnboardingMismatch
		detail string
	}{
		{OnboardingMismatch{Field: OnboardingFieldVendor, Code: 0x1234, Node: 0xFFF1}, "the code's vendor id is 0x1234, the node's is 0xFFF1"},
		{OnboardingMismatch{Field: OnboardingFieldProduct, Code: 0x8000, Node: 0x8001}, "the code's product id is 0x8000, the node's is 0x8001"},
		{OnboardingMismatch{Field: OnboardingFieldDiscriminator, Code: 3, Node: 15}, "the code's discriminator is 3, the node advertised 15"},
	} {
		m.OnboardingMismatch = &c.mm
		rec := do(t, mux, http.MethodPut, "/controller/nodes/0000000000000011/onboarding-code", `{"code":"34970112332"}`)
		var body struct {
			Code      string `json:"code"`
			Detail    string `json:"detail"`
			Field     string `json:"field"`
			CodeValue uint16 `json:"code_value"`
			NodeValue uint16 `json:"node_value"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusConflict || body.Code != codeOnboardingMismatch || body.Field != c.mm.Field ||
			body.CodeValue != c.mm.Code || body.NodeValue != c.mm.Node || body.Detail != c.detail {
			t.Fatalf("%+v: %d %s", c.mm, rec.Code, rec.Body)
		}
		if !errors.Is(&c.mm, ErrOnboardingMismatch) || !strings.Contains(c.mm.Error(), c.detail) {
			t.Fatalf("error %v", c.mm.Error())
		}
	}
	if inv := (&OnboardingInvalid{Reason: "bad check digit"}); !errors.Is(inv, ErrOnboardingCode) || inv.Error() != "fabricui: not an onboarding code: bad check digit" {
		t.Fatalf("invalid %v", inv)
	}
}

// holder keeps a secret in an unexported field, as a host's struct might.
type holder struct {
	name   string
	secret SecretText
}

// G2: a SecretText, an OnboardingEntry and a struct holding one in an
// unexported field render no code under any fmt verb, slog's text and
// JSON handlers and encoding/json.
func TestSecretTextRedacts(t *testing.T) {
	const code = "MT:Y.K9042C00KA0648G00"
	s := NewSecretText(code)
	e := OnboardingEntry{NodeID: 0x12, HasCode: true, QRCode: s, ManualCode: NewSecretText("34970112332")}
	h := holder{name: "x", secret: s}
	out := make([]string, 0, 36)
	for _, v := range []any{s, &s, e, &e, h, &h} {
		for _, verb := range []string{"%v", "%+v", "%#v", "%s"} {
			out = append(out, fmt.Sprintf(verb, v))
		}
		j, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, string(j))
		var buf bytes.Buffer
		slog.New(slog.NewTextHandler(&buf, nil)).Info("m", "v", v)
		slog.New(slog.NewJSONHandler(&buf, nil)).Info("m", "v", v)
		out = append(out, buf.String())
	}
	for _, o := range out {
		if strings.Contains(o, code) || strings.Contains(o, "34970112332") {
			t.Fatalf("a code in %q", o)
		}
	}
	if s.Reveal() != code || s.IsZero() || !NewSecretText("").IsZero() || NewSecretText("").String() != "" || s.String() != "redacted" {
		t.Fatal("Reveal, IsZero or String")
	}
	if got := e.String(); got != "OnboardingEntry(node 0000000000000012, code redacted)" {
		t.Fatalf("entry %q", got)
	}
	if got := (OnboardingEntry{NodeID: 1}).String(); !strings.Contains(got, "no code") {
		t.Fatalf("entry %q", got)
	}
}
