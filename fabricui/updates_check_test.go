// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"thingfabric.dev/go-fabric-ui/internal/api"
)

// checkWith is the memory controller's update source with Check
// answering found and err.
type checkWith struct {
	*memoryUpdates
	found []AvailableUpdate
	err   error
}

func (c checkWith) Check(context.Context) ([]AvailableUpdate, error) { return c.found, c.err }

type checkController struct {
	*MemoryController
	found []AvailableUpdate
	err   error
}

func (c checkController) SoftwareUpdates() SoftwareUpdateSource {
	return checkWith{c.updates, c.found, c.err}
}

// A node whose update check failed does not fail the others': the
// updates found are answered 200 with the failed node in errors
// (contract 0.11.0, decision 70, UR-P4-adpt-32).
func TestCheckUpdates_APartialFailureListsTheNode(t *testing.T) {
	ctrl := checkController{
		MemoryController: NewMemoryController(),
		found:            []AvailableUpdate{{NodeID: 0x11, VendorID: 0xFFF1, ProductID: 0x8001, SoftwareVersion: 3_000_000_000, SoftwareVersionString: "3.0"}},
		err:              UpdateCheckFailures{{NodeID: 0x12, Err: errors.New("image download failed")}},
	}
	_, mux := newTestHandler(t, testOptions{noNode: true, controller: ctrl})
	rec := do(t, mux, http.MethodPost, "/controller/updates/check", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("partial check: %d %s", rec.Code, rec.Body)
	}
	var got api.AvailableUpdateList
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Updates) != 1 || got.Updates[0].NodeIDHex != "0000000000000011" || got.Updates[0].SoftwareVersion != 3_000_000_000 {
		t.Errorf("updates: %+v", got.Updates)
	}
	if got.Errors == nil || len(*got.Errors) != 1 || (*got.Errors)[0].NodeIDHex != "0000000000000012" || (*got.Errors)[0].Error != "image download failed" {
		t.Errorf("errors: %+v", got.Errors)
	}
}

// Without a failure the answer has no errors member.
func TestCheckUpdates_NoFailureNoErrors(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{noNode: true, controller: NewMemoryController()})
	rec := do(t, mux, http.MethodPost, "/controller/updates/check", "")
	var got map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || rec.Code != http.StatusOK {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	if _, ok := got["errors"]; ok {
		t.Errorf("errors present: %s", rec.Body)
	}
}

// A check that fails as a whole is 500, a declared status, also when the
// source names a node's failure (the status guard holds the rest).
func TestCheckUpdates_AWholeFailureIs500(t *testing.T) {
	ctrl := checkController{MemoryController: NewMemoryController(), err: fmt.Errorf("%w: ledger", ErrNodeFailed)}
	_, mux := newTestHandler(t, testOptions{noNode: true, controller: ctrl})
	if rec := do(t, mux, http.MethodPost, "/controller/updates/check", ""); rec.Code != http.StatusInternalServerError {
		t.Errorf("whole failure: %d %s", rec.Code, rec.Body)
	}
}
