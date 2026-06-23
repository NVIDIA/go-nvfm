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

func TestReturnError(t *testing.T) {
	if got := SUCCESS.Error(); got != "SUCCESS" {
		t.Fatalf("SUCCESS.Error() = %q, want SUCCESS", got)
	}
	if got := BADPARAM.String(); got != "BADPARAM" {
		t.Fatalf("BADPARAM.String() = %q, want BADPARAM", got)
	}
	if got := Return(-999).Error(); got != "unknown return value: -999" {
		t.Fatalf("unknown Error() = %q", got)
	}
}
