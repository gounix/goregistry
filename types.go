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
	"time"
)

const (
	checkAuthUrlPattern   = "%s://%s/v2/"
	getTokenUrlPattern    = "%s?service=%s&scope=repository:%s:%s"
	manifestUrlPattern    = "%s://%s/v2/%s/manifests/%s"
	blobUrlPattern        = "%s://%s/v2/%s/blobs/%s"
	versionBaseUrlPattern = "%s://%s"
	versionLinkUrlPattern = "/v2/%s/tags/list"
	deleteUrlPattern      = "%s://%s/v2/%s/manifests/%s"
	acceptImageIndex      = "application/vnd.docker.distribution.manifest.list.v2+json,application/vnd.oci.image.index.v1+json"
	acceptImageManifest   = "application/vnd.docker.distribution.manifest.v2+json,application/vnd.oci.image.manifest.v1+json"
	acceptAll             = "application/vnd.docker.distribution.manifest.list.v2+json,application/vnd.oci.image.index.v1+json,application/vnd.docker.distribution.manifest.v2+json,application/vnd.oci.image.manifest.v1+json"
)

type (
	TokenT string
	TokenRespT struct {
		Token       TokenT    `json:"token"`
		AccessToken string    `json:"access_token"`
		ExpiresIn   int64     `json:"expires_in"`
		IssuedAt    time.Time `json:"issued_at"`
	}
	AnnotationsT struct {
		Created time.Time `json:"org.opencontainers.image.created"`
		Url     string    `json:"org.opencontainers.image.url"`
		Version string    `json:"org.opencontainers.image.version"`
	}
	PlatformT struct {
		Architecture string `json:"architecture"`
		Os           string `json:"os"`
	}
	ManifestT struct {
		MediaType   string       `json:"mediaType"`
		Digest      string       `json:"digest"`
		Platform    PlatformT    `json:"platform"`
		Annotations AnnotationsT `json:"annotations"`
	}
	ManifestsT struct {
		MediaType string      `json:"mediaType"`
		Manifest  []ManifestT `json:"manifests"`
	}
	ConfigT struct {
		MediaType string `json:"mediaType"`
		Digest    string `json:"digest"`
	}
	SingleT struct {
		MediaType string  `json:"mediaType"`
		Config    ConfigT `json:"config"`
	}
	BlobT struct {
		Created time.Time `json:"created"`
	}
	TagsT struct {
                Name string   `json:"name"`
                Tags []string `json:"tags"`
        }
)

