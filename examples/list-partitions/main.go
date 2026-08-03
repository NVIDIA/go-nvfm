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

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/NVIDIA/go-nvfm/pkg/nvfm"
)

const defaultUnixSocketPath = "/run/nvidia-fabricmanager/socket"

type partitionList struct {
	Version          uint32          `json:"version"`
	NumPartitions    uint32          `json:"numPartitions"`
	MaxNumPartitions uint32          `json:"maxNumPartitions"`
	PartitionInfo    []partitionInfo `json:"partitionInfo"`
}

type partitionInfo struct {
	PartitionID uint32    `json:"partitionId"`
	IsActive    uint32    `json:"isActive"`
	NumGPUs     uint32    `json:"numGpus"`
	GPUInfo     []gpuInfo `json:"gpuInfo"`
}

type gpuInfo struct {
	PhysicalID          uint32 `json:"physicalId"`
	UUID                string `json:"uuid"`
	PCIBusID            string `json:"pciBusId"`
	NumNVLinksAvailable uint32 `json:"numNvLinksAvailable"`
	MaxNumNVLinks       uint32 `json:"maxNumNvLinks"`
	NVLinkLineRateMBps  uint32 `json:"nvlinkLineRateMBps"`
}

// This example is only usable when Fabric Manager is configured for
// multi-tenancy fabric mode (FABRIC_MODE=1) and its Unix socket interface
// (FM_CMD_UNIX_SOCKET_PATH) is configured.
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ret := nvfm.Init()
	if ret != nvfm.SUCCESS {
		return fmt.Errorf("unable to initialize NVFM: %s (%d)", nvfm.ErrorString(ret), ret)
	}
	defer nvfm.Shutdown()

	unixSocketPath := os.Getenv("NVFM_UNIX_SOCKET_PATH")
	if unixSocketPath == "" {
		unixSocketPath = defaultUnixSocketPath
	}

	handle, ret := nvfm.Connect(nvfm.WithUnixSocket(unixSocketPath))
	if ret != nvfm.SUCCESS {
		return fmt.Errorf("unable to connect to Fabric Manager: %s (%d)", nvfm.ErrorString(ret), ret)
	}
	defer handle.Disconnect()

	partitions, ret := handle.GetSupportedFabricPartitions()
	if ret != nvfm.SUCCESS {
		return fmt.Errorf("unable to query supported fabric partitions: %s (%d)", nvfm.ErrorString(ret), ret)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(newPartitionList(partitions)); err != nil {
		return fmt.Errorf("unable to encode supported fabric partitions: %w", err)
	}

	return nil
}

func newPartitionList(partitions nvfm.FabricPartitionList) partitionList {
	output := partitionList{
		Version:          partitions.Version,
		NumPartitions:    partitions.NumPartitions,
		MaxNumPartitions: partitions.MaxNumPartitions,
		PartitionInfo:    make([]partitionInfo, 0, partitions.NumPartitions),
	}
	for partitionIndex := uint32(0); partitionIndex < partitions.NumPartitions; partitionIndex++ {
		partition := partitions.PartitionInfo[partitionIndex]
		partitionOutput := partitionInfo{
			PartitionID: partition.PartitionId,
			IsActive:    partition.IsActive,
			NumGPUs:     partition.NumGpus,
			GPUInfo:     make([]gpuInfo, 0, partition.NumGpus),
		}
		for gpuIndex := uint32(0); gpuIndex < partition.NumGpus; gpuIndex++ {
			gpu := partition.GpuInfo[gpuIndex]
			partitionOutput.GPUInfo = append(partitionOutput.GPUInfo, gpuInfo{
				PhysicalID:          gpu.PhysicalId,
				UUID:                int8ArrayString(gpu.Uuid[:]),
				PCIBusID:            int8ArrayString(gpu.PciBusId[:]),
				NumNVLinksAvailable: gpu.NumNvLinksAvailable,
				MaxNumNVLinks:       gpu.MaxNumNvLinks,
				NVLinkLineRateMBps:  gpu.NvlinkLineRateMBps,
			})
		}
		output.PartitionInfo = append(output.PartitionInfo, partitionOutput)
	}
	return output
}

func int8ArrayString(value []int8) string {
	length := len(value)
	for index, character := range value {
		if character == 0 {
			length = index
			break
		}
	}
	bytes := make([]byte, length)
	for index := range bytes {
		bytes[index] = byte(value[index])
	}
	return string(bytes)
}
