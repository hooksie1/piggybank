// Copyright 2024 Cover Whale Insurance Solutions Inc.
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

package gupdate

import (
	"bufio"
	"fmt"
	"io"
	"runtime"
	"strings"
)

func GoReleaserChecksum(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, strings.ToLower(runtime.GOOS)) && strings.Contains(line, strings.ToLower(runtime.GOARCH)) {
			return strings.Split(line, " ")[0], nil
		}
	}

	return "", fmt.Errorf("valid checksum not found")
}
