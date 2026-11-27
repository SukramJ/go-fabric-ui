// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"thingfabric.dev/go-fabric/controller"
	"thingfabric.dev/go-fabric/controller/softwareupdate"
	"thingfabric.dev/go-fabric/device"
	"thingfabric.dev/go-fabric/fabrictest"
	"thingfabric.dev/go-fabric/node"

	"thingfabric.dev/go-fabric-ui/fabricui"
)

// go-fabric's real update manager fails one product and checks the other
// (UR-P4-test-403). A product's check fails only where
// Images.DownloadUpdate fails, and the manager goes on with the next
// product (controller/softwareupdate/manager.go:1980-1981 and
// :1907-1911 at v0.28.0); offline — NewImages without a ledger — it
// fails when no valid image is stored (images.go:1105-1106). Which stored
// image does that for one product alone was not read; this test
// establishes it: two commissioned nodes of two products, a newer local
// image of each stored, and one product's image corrupted in the store
// afterwards (its last payload byte flipped, the header intact). The
// route lists the good product's update, names the other node in errors,
// and the stream carries one check_finished.
func TestEndToEnd_UpdateCheckFailsOneProduct(t *testing.T) {
	blobs := &softwareupdate.MemoryBlobs{}
	images := softwareupdate.NewImages(softwareupdate.ImagesConfig{Blobs: blobs})
	product := func(pid uint16) func(*node.Config) {
		dir := t.TempDir()
		return func(cfg *node.Config) {
			cfg.ProductDescription.ProductID = pid
			// The OTA requestor makes it a node the manager updates.
			cfg.OTA = &node.OTAConfig{Dir: dir}
		}
	}
	s := newCtrlSiteWith(t, ctrlSiteOptions{
		// The images are local ones: the manager asks the store for
		// production images only without AllowTestImages
		// (manager.go:1971-1975 at v0.28.0).
		ota:       &controller.OTAConfig{Images: images, AllowTestImages: true},
		configure: product(0x8101),
	})
	plug, err := device.NewOnOffPlugInUnit(device.Options{StableKey: device.StringKey("plug:2"), Name: "Plug 2"})
	if err != nil {
		t.Fatal(err)
	}
	second := s.site.AddDevice(fabrictest.Topology{Bridged: []node.Endpoint{plug}, Configure: product(0x8102)})
	good := commissionSiteDevice(t, s)
	sp, err := second.Node().SetupPayload()
	if err != nil {
		t.Fatal(err)
	}
	var added nodeDetail
	s.json(http.MethodPost, "/controller/commission", fmt.Sprintf(`{"code":%q}`, sp.Manual), http.StatusOK, &added)
	bad := added.Node.NodeIDHex
	s.waitEvent(fabricui.EventSubscription)

	ctx := context.Background()
	for _, n := range s.src.Nodes() {
		b := n.Basic()
		img, err := node.CreateOTAImage(node.OTAImageOptions{
			VendorID: b.VendorID, ProductID: b.ProductID, SoftwareVersion: 0x10000, SoftwareVersionString: "99.0",
			Payload: []byte("the payload of the image"),
		})
		if err != nil {
			t.Fatal(err)
		}
		st, err := images.Store(ctx, img, softwareupdate.ModeLocal, nil)
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%016X", n.ID()) == bad {
			// The header still reads; the payload no longer matches its digest.
			img[len(img)-1] ^= 0xFF
			if err := blobs.Put(ctx, st.Filename, img); err != nil {
				t.Fatal(err)
			}
		}
	}

	updates := s.controllerStream(fabricui.EventUpdate)
	var list struct {
		Updates []struct {
			NodeIDHex string `json:"node_id_hex"`
		} `json:"updates"`
		Errors []struct {
			NodeIDHex string `json:"node_id_hex"`
			Error     string `json:"error"`
		} `json:"errors"`
	}
	s.json(http.MethodPost, "/controller/updates/check", "", http.StatusOK, &list)
	if len(list.Errors) != 1 || list.Errors[0].NodeIDHex != bad || list.Errors[0].Error == "" {
		t.Fatalf("errors %+v, want the node %s", list.Errors, bad)
	}
	if len(list.Updates) != 1 || list.Updates[0].NodeIDHex != good {
		t.Fatalf("updates %+v, want the node %s", list.Updates, good)
	}
	finished := 0
	quiet := time.After(2 * time.Second)
	for done := false; !done; {
		select {
		case d := <-updates:
			if strings.Contains(d, `"update_kind":"check_finished"`) {
				finished++
			}
		case <-quiet:
			done = true
		}
	}
	if finished != 1 {
		t.Fatalf("%d check_finished, want 1", finished)
	}
}
