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

func TestHandleWrappersInitializeVersions(t *testing.T) {
	var raw nvfmHandle
	handle := &fabricManager{handle: &raw}

	originalSupported := fmGetSupportedFabricPartitionsFunc
	originalUnsupported := fmGetUnsupportedFabricPartitionsFunc
	originalFailed := fmGetNvlinkFailedDevicesFunc
	defer func() {
		fmGetSupportedFabricPartitionsFunc = originalSupported
		fmGetUnsupportedFabricPartitionsFunc = originalUnsupported
		fmGetNvlinkFailedDevicesFunc = originalFailed
	}()

	fmGetSupportedFabricPartitionsFunc = func(_ *nvfmHandle, list *FabricPartitionList) Return {
		if list.Version != FabricPartitionListVersion {
			t.Fatalf("supported version = %d, want %d", list.Version, FabricPartitionListVersion)
		}
		return SUCCESS
	}
	if _, ret := handle.GetSupportedFabricPartitions(); ret != SUCCESS {
		t.Fatalf("GetSupportedFabricPartitions() = %v, want SUCCESS", ret)
	}

	fmGetUnsupportedFabricPartitionsFunc = func(_ *nvfmHandle, list *UnsupportedFabricPartitionList) Return {
		if list.Version != UnsupportedFabricPartitionListVersion {
			t.Fatalf("unsupported version = %d, want %d", list.Version, UnsupportedFabricPartitionListVersion)
		}
		return SUCCESS
	}
	if _, ret := handle.GetUnsupportedFabricPartitions(); ret != SUCCESS {
		t.Fatalf("GetUnsupportedFabricPartitions() = %v, want SUCCESS", ret)
	}

	fmGetNvlinkFailedDevicesFunc = func(_ *nvfmHandle, devices *NvlinkFailedDevices) Return {
		if devices.Version != NvlinkFailedDevicesVersion {
			t.Fatalf("failed devices version = %d, want %d", devices.Version, NvlinkFailedDevicesVersion)
		}
		return SUCCESS
	}
	if _, ret := handle.GetNvlinkFailedDevices(); ret != SUCCESS {
		t.Fatalf("GetNvlinkFailedDevices() = %v, want SUCCESS", ret)
	}
}

func TestSetActivatedFabricPartitions(t *testing.T) {
	var raw nvfmHandle
	handle := &fabricManager{handle: &raw}

	original := fmSetActivatedFabricPartitionsFunc
	defer func() { fmSetActivatedFabricPartitionsFunc = original }()

	fmSetActivatedFabricPartitionsFunc = func(_ *nvfmHandle, list *ActivatedFabricPartitionList) Return {
		if list.Version != ActivatedFabricPartitionListVersion {
			t.Fatalf("version = %d, want %d", list.Version, ActivatedFabricPartitionListVersion)
		}
		if list.NumPartitions != 3 {
			t.Fatalf("num partitions = %d, want 3", list.NumPartitions)
		}
		if got := []uint32{list.PartitionIds[0], list.PartitionIds[1], list.PartitionIds[2]}; got[0] != 1 || got[1] != 3 || got[2] != 5 {
			t.Fatalf("partition ids = %#v, want [1 3 5]", got)
		}
		return SUCCESS
	}

	if ret := handle.SetActivatedFabricPartitions([]FabricPartitionId{1, 3, 5}); ret != SUCCESS {
		t.Fatalf("SetActivatedFabricPartitions() = %v, want SUCCESS", ret)
	}

	tooMany := make([]FabricPartitionId, MAX_FABRIC_PARTITIONS+1)
	if ret := handle.SetActivatedFabricPartitions(tooMany); ret != BADPARAM {
		t.Fatalf("SetActivatedFabricPartitions(too many) = %v, want BADPARAM", ret)
	}
}

func TestActivateFabricPartitionWithVFs(t *testing.T) {
	var raw nvfmHandle
	handle := &fabricManager{handle: &raw}

	original := fmActivateFabricPartitionWithVFsFunc
	defer func() { fmActivateFabricPartitionWithVFsFunc = original }()

	calls := 0
	fmActivateFabricPartitionWithVFsFunc = func(_ *nvfmHandle, id FabricPartitionId, vfs *PciDevice, numVfs uint32) Return {
		calls++
		if id != 7 {
			t.Fatalf("partition id = %d, want 7", id)
		}
		if vfs == nil {
			t.Fatal("vf pointer is nil")
		}
		if numVfs != 2 {
			t.Fatalf("num VFs = %d, want 2", numVfs)
		}
		return SUCCESS
	}

	vfs := []PciDevice{{Domain: 1}, {Domain: 2}}
	if ret := handle.ActivateFabricPartitionWithVFs(7, vfs); ret != SUCCESS {
		t.Fatalf("ActivateFabricPartitionWithVFs() = %v, want SUCCESS", ret)
	}

	tooMany := make([]PciDevice, MAX_NUM_GPUS+1)
	if ret := handle.ActivateFabricPartitionWithVFs(7, tooMany); ret != BADPARAM {
		t.Fatalf("ActivateFabricPartitionWithVFs(too many) = %v, want BADPARAM", ret)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}
