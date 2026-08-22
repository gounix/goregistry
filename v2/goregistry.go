/*
MIT License

Copyright (c) 2026 gounix

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package goregistry

import (
	"encoding/json"
	"errors"
	"log/slog"
	"time"
)

// assuming that all architectures were built at the same time, so pick the first architecture that is not "unknown"
func getArchDigest(manifest JsonManifestListT) (string, string, error) {
	for _, entry := range manifest.Manifest {
		if entry.Platform.Architecture != "unknown" && entry.Platform.Os != "unknown" {
			slog.Info("goregistry.getArchDigest", "arch", entry.Platform.Architecture, "os", entry.Platform.Os)
			return entry.MediaType, entry.Digest, nil
		}
	}

	slog.Error("goregistry.getArchDigest not found")
	return "", "", errors.New("architecture not found")
}

func (registry RegistryT) GetLastUpdate(tag string) (time.Time, error) {
	var digest string
	var mediaType string
	var dat BlobT

	registry.RenewToken()
	retML, err := registry.GetManifestList(tag)
	if err != nil {
		// there is no image index manifest, try a normal manifest
		slog.Info("goregistry.GetLastUpdate GetManifestList", "err", err)
		digest = tag
		mediaType = acceptImageManifest
	} else {
		mediaType, digest, err = getArchDigest(retML.Json)
		if err != nil {
			slog.Error("goregistry.GetLastUpdate", "err", err)
			return time.Time{}, err
		}
	}

	slog.Info("goregistry.GetLastUpdate", "mediaType", mediaType, "digest", digest)
	retM, err := registry.GetManifest(mediaType, digest)
	if err != nil {
		slog.Error("goregistry.GetLastUpdate GetManifest", "err", err)
		return time.Time{}, err
	}

	blob, err := registry.GetBlob(retM.Json.Config.MediaType, retM.Json.Config.Digest)
	if err != nil {
		slog.Error("goregistry.GetLastUpdate", "err", err)
		return time.Time{}, err
	}

	json.Unmarshal(blob, &dat)
	slog.Info("goregistry.GetLastUpdate", "date", dat.Created)
	return dat.Created, nil
}
