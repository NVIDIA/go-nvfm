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

import "testing"

func TestNewConnectParams(t *testing.T) {
	params := newConnectParams()
	if got := int8ArrayString(params.AddressInfo[:]); got != defaultAddress {
		t.Fatalf("default address = %q, want %q", got, defaultAddress)
	}
	if params.TimeoutMs != defaultTimeoutMs {
		t.Fatalf("default timeout = %d, want %d", params.TimeoutMs, defaultTimeoutMs)
	}
	if params.AddressType != ADDR_TYPE_INET {
		t.Fatalf("default address type = %d, want ADDR_TYPE_INET", params.AddressType)
	}
	if params.AddressIsUnixSocket != 0 {
		t.Fatalf("default unix socket flag = %d, want 0", params.AddressIsUnixSocket)
	}
	if params.Version != ConnectParamsVersion {
		t.Fatalf("version = %d, want %d", params.Version, ConnectParamsVersion)
	}

	params = newConnectParams(WithAddress("10.0.0.1:7777"), WithTimeoutMs(2500))
	if got := int8ArrayString(params.AddressInfo[:]); got != "10.0.0.1:7777" {
		t.Fatalf("address = %q, want 10.0.0.1:7777", got)
	}
	if params.TimeoutMs != 2500 {
		t.Fatalf("timeout = %d, want 2500", params.TimeoutMs)
	}
	if params.AddressType != ADDR_TYPE_INET {
		t.Fatalf("address type = %d, want ADDR_TYPE_INET", params.AddressType)
	}

	params = newConnectParams(WithUnixSocket("/run/nvidia-fm.sock"))
	if got := int8ArrayString(params.AddressInfo[:]); got != "/run/nvidia-fm.sock" {
		t.Fatalf("unix address = %q, want /run/nvidia-fm.sock", got)
	}
	if params.AddressType != ADDR_TYPE_UNIX {
		t.Fatalf("unix address type = %d, want ADDR_TYPE_UNIX", params.AddressType)
	}
	if params.AddressIsUnixSocket != 1 {
		t.Fatalf("unix socket flag = %d, want 1", params.AddressIsUnixSocket)
	}
}

func TestConnectWithParams(t *testing.T) {
	original := fmConnectFunc
	defer func() { fmConnectFunc = original }()

	var raw nvfmHandle
	fmConnectFunc = func(params *ConnectParams, out **nvfmHandle) Return {
		if params.Version != ConnectParamsVersion {
			t.Fatalf("version = %d, want %d", params.Version, ConnectParamsVersion)
		}
		*out = &raw
		return SUCCESS
	}

	handle, ret := newLibrary().Connect()
	if ret != SUCCESS {
		t.Fatalf("Connect() ret = %v, want SUCCESS", ret)
	}
	if handle == nil {
		t.Fatal("Connect() handle is nil")
	}
}
