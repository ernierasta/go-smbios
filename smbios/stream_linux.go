// Copyright 2017 DigitalOcean.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//+build linux

package smbios

import (
	"errors"
	"io"
	"os"
)

// stream opens a SMBIOS structure stream.
func stream() (io.ReadCloser, error) {
	// First, check for the sysfs location present in modern kernels.
	f, err := os.Open("/sys/firmware/dmi/tables/DMI")
	switch {
	case err == nil:
		return f, nil
	case os.IsNotExist(err):
		// TODO(mdlayher): try reading /dev/mem and fail if no data present.
		return nil, errors.New("reading from /dev/mem not yet supported on Linux")
	default:
		return nil, err
	}
}
