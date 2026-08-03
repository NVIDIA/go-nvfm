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
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/NVIDIA/go-nvfm/pkg/nvfm"
)

const defaultUnixSocketPath = "/run/nvidia-fabricmanager/socket"

// This example is only usable when Fabric Manager is configured for
// multi-tenancy fabric mode (FABRIC_MODE=1) and its Unix socket interface
// (FM_CMD_UNIX_SOCKET_PATH) is configured.
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	partitionIDValue := os.Getenv("NVFM_PARTITION_ID")
	if partitionIDValue == "" {
		return fmt.Errorf("NVFM_PARTITION_ID must be set")
	}
	parsedPartitionID, err := strconv.ParseUint(partitionIDValue, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid NVFM_PARTITION_ID %q: %w", partitionIDValue, err)
	}
	partitionID := nvfm.FabricPartitionId(parsedPartitionID)

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

	ret = handle.DeactivateFabricPartition(partitionID)
	if ret != nvfm.SUCCESS {
		return fmt.Errorf("unable to deactivate fabric partition %d: %s (%d)", partitionID, nvfm.ErrorString(ret), ret)
	}

	fmt.Printf("Deactivated fabric partition %d\n", partitionID)
	return nil
}
