// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES.  All rights reserved.
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nvfm

import (
	"errors"
	"testing"
)

type dynamicLibraryMock struct {
	openFunc   func() error
	closeFunc  func() error
	lookupFunc func(string) error
	openCalls  int
	closeCalls int
	lookups    []string
}

func (m *dynamicLibraryMock) Open() error {
	m.openCalls++
	if m.openFunc != nil {
		return m.openFunc()
	}
	return nil
}

func (m *dynamicLibraryMock) Close() error {
	m.closeCalls++
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func (m *dynamicLibraryMock) Lookup(name string) error {
	m.lookups = append(m.lookups, name)
	if m.lookupFunc != nil {
		return m.lookupFunc(name)
	}
	return nil
}

func newTestLibrary(dl dynamicLibrary) *library {
	return &library{path: "test-libnvfm.so", dl: dl}
}

func withLifecycleStubs(initRet, shutdownRet Return) func() (int, int) {
	originalInit := fmLibInitFunc
	originalShutdown := fmLibShutdownFunc
	initCalls := 0
	shutdownCalls := 0
	fmLibInitFunc = func() Return {
		initCalls++
		return initRet
	}
	fmLibShutdownFunc = func() Return {
		shutdownCalls++
		return shutdownRet
	}
	return func() (int, int) {
		fmLibInitFunc = originalInit
		fmLibShutdownFunc = originalShutdown
		return initCalls, shutdownCalls
	}
}

func TestInitShutdownRefcount(t *testing.T) {
	restore := withLifecycleStubs(SUCCESS, SUCCESS)
	defer restore()

	dl := &dynamicLibraryMock{}
	l := newTestLibrary(dl)

	if ret := l.Init(); ret != SUCCESS {
		t.Fatalf("Init() = %v, want SUCCESS", ret)
	}
	if ret := l.Init(); ret != SUCCESS {
		t.Fatalf("second Init() = %v, want SUCCESS", ret)
	}
	if dl.openCalls != 1 {
		t.Fatalf("Open calls = %d, want 1", dl.openCalls)
	}
	if l.refcount != 2 {
		t.Fatalf("refcount = %d, want 2", l.refcount)
	}

	if ret := l.Shutdown(); ret != SUCCESS {
		t.Fatalf("first Shutdown() = %v, want SUCCESS", ret)
	}
	if dl.closeCalls != 0 {
		t.Fatalf("Close calls after first shutdown = %d, want 0", dl.closeCalls)
	}
	if ret := l.Shutdown(); ret != SUCCESS {
		t.Fatalf("second Shutdown() = %v, want SUCCESS", ret)
	}
	if dl.closeCalls != 1 {
		t.Fatalf("Close calls after final shutdown = %d, want 1", dl.closeCalls)
	}
	if l.refcount != 0 {
		t.Fatalf("refcount = %d, want 0", l.refcount)
	}

	initCalls, shutdownCalls := restore()
	if initCalls != 1 {
		t.Fatalf("fmLibInit calls = %d, want 1", initCalls)
	}
	if shutdownCalls != 1 {
		t.Fatalf("fmLibShutdown calls = %d, want 1", shutdownCalls)
	}
}

func TestInitAndShutdownErrors(t *testing.T) {
	openErr := errors.New("open error")
	closeErr := errors.New("close error")

	t.Run("open error", func(t *testing.T) {
		restore := withLifecycleStubs(SUCCESS, SUCCESS)
		defer restore()

		l := newTestLibrary(&dynamicLibraryMock{openFunc: func() error { return openErr }})
		if ret := l.Init(); ret != GENERIC_ERROR {
			t.Fatalf("Init() = %v, want GENERIC_ERROR", ret)
		}
		if l.refcount != 0 {
			t.Fatalf("refcount = %d, want 0", l.refcount)
		}
	})

	t.Run("fmLibInit error closes library", func(t *testing.T) {
		restore := withLifecycleStubs(BADPARAM, SUCCESS)
		defer restore()

		dl := &dynamicLibraryMock{}
		l := newTestLibrary(dl)
		if ret := l.Init(); ret != BADPARAM {
			t.Fatalf("Init() = %v, want BADPARAM", ret)
		}
		if dl.closeCalls != 1 {
			t.Fatalf("Close calls = %d, want 1", dl.closeCalls)
		}
		if l.refcount != 0 {
			t.Fatalf("refcount = %d, want 0", l.refcount)
		}
	})

	t.Run("close error keeps refcount", func(t *testing.T) {
		restore := withLifecycleStubs(SUCCESS, SUCCESS)
		defer restore()

		l := newTestLibrary(&dynamicLibraryMock{closeFunc: func() error { return closeErr }})
		if ret := l.Init(); ret != SUCCESS {
			t.Fatalf("Init() = %v, want SUCCESS", ret)
		}
		if ret := l.Shutdown(); ret != GENERIC_ERROR {
			t.Fatalf("Shutdown() = %v, want GENERIC_ERROR", ret)
		}
		if l.refcount != 1 {
			t.Fatalf("refcount = %d, want 1", l.refcount)
		}
	})
}

func TestLookupSymbol(t *testing.T) {
	lookupErr := errors.New("lookup error")
	restore := withLifecycleStubs(SUCCESS, SUCCESS)
	defer restore()

	dl := &dynamicLibraryMock{lookupFunc: func(string) error { return lookupErr }}
	l := newTestLibrary(dl)

	if err := l.LookupSymbol("fmConnect"); !errors.Is(err, errLibraryNotLoaded) {
		t.Fatalf("LookupSymbol before Init error = %v, want errLibraryNotLoaded", err)
	}
	if ret := l.Init(); ret != SUCCESS {
		t.Fatalf("Init() = %v, want SUCCESS", ret)
	}
	if err := l.LookupSymbol("fmConnect"); !errors.Is(err, lookupErr) {
		t.Fatalf("LookupSymbol error = %v, want lookupErr", err)
	}
	if len(dl.lookups) != 1 || dl.lookups[0] != "fmConnect" {
		t.Fatalf("lookups = %#v, want fmConnect", dl.lookups)
	}
}
