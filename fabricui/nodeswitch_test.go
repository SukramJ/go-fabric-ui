// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui_test

import (
	"sync"
	"testing"

	"github.com/SukramJ/go-fabric-ui/fabricui"
)

// stubSource is a partial NodeSource: the embedded nil interface stands
// for every method but LocalAddr.
type stubSource struct {
	fabricui.NodeSource
	addr string
}

func (s *stubSource) LocalAddr() string { return s.addr }

// Readers on several goroutines see one target or the other while the
// test swaps them, never a torn or empty one; -race (make race) is the
// assertion that matters.
func TestNodeSwitch_SwapUnderConcurrentReader(t *testing.T) {
	a, b := &stubSource{addr: "[::]:5540"}, &stubSource{addr: "[::]:5541"}
	sw := fabricui.NewNodeSwitch(a)
	stop := make(chan struct{})
	var wg sync.WaitGroup
	errs := make(chan string, 8)
	for range 8 {
		wg.Go(func() {
			for {
				select {
				case <-stop:
					return
				default:
				}
				if got := sw.LocalAddr(); got != a.addr && got != b.addr {
					errs <- "LocalAddr " + got
					return
				}
				if c := sw.Current(); c != a && c != b {
					errs <- "Current is neither target"
					return
				}
			}
		})
	}
	for i := range 1000 {
		if i%2 == 0 {
			sw.Set(b)
		} else {
			sw.Set(a)
		}
	}
	close(stop)
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
	if sw.Current() != a || sw.LocalAddr() != a.addr {
		t.Fatalf("after the last Set: %v", sw.LocalAddr())
	}
}

// A switch never holds nil: NewNodeSwitch and Set refuse it at the call
// that is wrong.
func TestNodeSwitch_NilRefused(t *testing.T) {
	for _, c := range []struct {
		want string
		call func()
	}{
		{"fabricui: NewNodeSwitch: nil source", func() { fabricui.NewNodeSwitch(nil) }},
		{"fabricui: NodeSwitch.Set: nil source", func() { fabricui.NewNodeSwitch(&stubSource{}).Set(nil) }},
	} {
		func() {
			defer func() {
				if r := recover(); r != c.want {
					t.Errorf("panic %v, want %q", r, c.want)
				}
			}()
			c.call()
		}()
	}
}
