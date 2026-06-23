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
	"testing"

	"github.com/NVIDIA/go-nvfm/pkg/dl"
)

func requireLibNVFM(t *testing.T) {
	t.Helper()

	lib := dl.New(defaultNvfmLibraryName, defaultNvfmLibraryLoadFlags)
	if err := lib.Open(); err != nil {
		t.Skipf("This test requires %s: %v", defaultNvfmLibraryName, err)
	}
	_ = lib.Close()
}

func TestInitIntegration(t *testing.T) {
	requireLibNVFM(t)

	if ret := Init(); ret != SUCCESS {
		t.Fatalf("Init() = %v, want SUCCESS", ret)
	}
	if ret := Shutdown(); ret != SUCCESS {
		t.Fatalf("Shutdown() = %v, want SUCCESS", ret)
	}
}
