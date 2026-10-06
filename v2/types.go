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
	"github.com/gounix/gosecret"
	"time"
)

const (
	checkAuthUrlPattern       = "%s://%s/v2/"
	getTokenUrlPattern        = "%s?service=%s&scope=repository:%s:%s"
	getCatalogTokenUrlPattern = "%s?service=%s&scope=registry:catalog:*"
	manifestUrlPattern        = "%s://%s/v2/%s/manifests/%s"
	blobUrlPattern            = "%s://%s/v2/%s/blobs/%s"
	versionBaseUrlPattern     = "%s://%s"
	versionLinkUrlPattern     = "/v2/%s/tags/list"
	catalogBaseUrlPattern     = "%s://%s"
	deleteUrlPattern          = "%s://%s/v2/%s/manifests/%s"

	acceptDockerImageIndex    = "application/vnd.docker.distribution.manifest.list.v2+json"
	acceptOCIImageIndex       = "application/vnd.oci.image.index.v1+json"
	acceptImageIndex          = "application/vnd.docker.distribution.manifest.list.v2+json,application/vnd.oci.image.index.v1+json"

	acceptDockerImageManifest = "application/vnd.docker.distribution.manifest.v2+json"
	acceptOCIImageManifest    = "application/vnd.oci.image.manifest.v1+json"
	acceptImageManifest       = "application/vnd.docker.distribution.manifest.v2+json,application/vnd.oci.image.manifest.v1+json"

	acceptAll                 = "application/vnd.docker.distribution.manifest.list.v2+json,application/vnd.oci.image.index.v1+json,application/vnd.docker.distribution.manifest.v2+json,application/vnd.oci.image.manifest.v1+json"
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
		Created         time.Time `json:"org.opencontainers.image.created,omitzero"`
		Url             string    `json:"org.opencontainers.image.url,omitempty"`
		Version         string    `json:"org.opencontainers.image.version,omitempty"`
		Revision        string    `json:"org.opencontainers.image.revision,omitempty"`
		Source          string    `json:"org.opencontainers.image.source,omitempty"`
		Base            string    `json:"org.opencontainers.image.base.name,omitempty"`
		Arch            string    `json:"com.docker.official-images.bashbrew.arch,omitempty"`
		ReferenceDigest string    `json:"vnd.docker.reference.digest,omitempty"`
		ReferenceType   string    `json:"vnd.docker.reference.type,omitempty"`
	}
	PlatformT struct {
		Architecture string `json:"architecture"`
		Os           string `json:"os"`
	}
	ArchManifestT struct {
		MediaType   string       `json:"mediaType"`
		Digest      string       `json:"digest"`
		Platform    PlatformT    `json:"platform"`
		Annotations AnnotationsT `json:"annotations"`
		Size        int64        `json:"size"`
	}
	JsonManifestListT struct {
		SchemaVersion int             `json:"schemaVersion"`
		MediaType     string          `json:"mediaType"`
		Manifest      []ArchManifestT `json:"manifests"`
	}
	// the config section in a manifest
	ConfigT struct {
		MediaType string `json:"mediaType"`
		Digest    string `json:"digest"`
		Size      int64  `json:"size"`
	}
	// a single layer in a manifest
	LayerT struct {
		MediaType string `json:"mediaType"`
		Digest    string `json:"digest"`
		Size      int64  `json:"size"`
	}
	JsonManifestT struct {
		SchemaVersion int          `json:"schemaVersion"`
		MediaType     string       `json:"mediaType"`
		Config        ConfigT      `json:"config"`
		Layers        []LayerT     `json:"layers"`
		Annotations   AnnotationsT `json:"annotations"`
	}
	// the structure of the config blob that is referenced in a manifest
	ConfigBlobT struct {
		Created time.Time `json:"created"`
	}
	TagsT struct {
                Name string   `json:"name"`
                Tags []string `json:"tags"`
        }
	CatalogT struct {
                Repositories []string `json:"repositories"`
        }
	RegistryT struct {
                Token                TokenT
		TokenScope           string
		Authentication       bool
		FullToken            TokenRespT
                Scheme               string
                TlsVerify            bool
                Host                 string
		SupportChunkedUpload bool
                Image                string
		Regcred              gosecret.RegCredT
        }
	BlobT struct {
		Digest       string
		Raw          []byte 
		LastModified time.Time
	}
	ManifestT struct {
		Digest string
		Raw    []byte 
		Json   JsonManifestT
	}
	ManifestListT struct {
		Digest string
		Raw    []byte 
		Json   JsonManifestListT
	}
	// error response given with http 4XX errors
	HttpError struct {
		Code string `json:"code"`
		Message string `json:"message"`
		Detail string `json:"detail"`
	}
	HttpErrorList struct {
		Error []HttpError `json:"errors"`
	}
)

