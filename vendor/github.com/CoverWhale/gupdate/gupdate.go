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
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/minio/selfupdate"
)

type LatestReleaseGetter interface {
	getLatestRelease() (Release, error)
}

type AllReleasesGetter interface {
	getAllReleases() ([]Release, error)
}

type CheckSumGetter interface {
	GetChecksum(io.Reader) (string, error)
}

type ChecksumFunc func(io.Reader) (string, error)

type RequestFunc func(r *http.Request)

type Release struct {
	Checksum string `json:"checksum,omitempty"`
	URL      string `json:"url"`
	ReqFunc  RequestFunc
}

func GetAllReleases(r AllReleasesGetter) ([]Release, error) {
	return r.getAllReleases()
}

func GetLatestRelease(r LatestReleaseGetter) (Release, error) {
	return r.getLatestRelease()
}

func (r Release) Update() error {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest(http.MethodGet, r.URL, nil)
	if err != nil {
		return err
	}

	if r.ReqFunc != nil {
		r.ReqFunc(req)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	cs, err := hex.DecodeString(r.Checksum)
	if err != nil {
		return err
	}

	if err := selfupdate.Apply(resp.Body, selfupdate.Options{
		Checksum: cs,
	}); err != nil {
		if updateErr := selfupdate.RollbackError(err); updateErr != nil {
			return fmt.Errorf("failed to rollback from bad update: %v", err)
		}

		return err
	}

	return nil
}
