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
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
	//"rebuilder/jsonreq"
	//"rebuilder/secret"
)

const (
	checkAuthUrlPattern = "%s://%s/v2/"
	getTokenUrlPattern  = "%s?service=%s&scope=repository:%s:pull"
	manifestUrlPattern  = "%s://%s/v2/%s/manifests/%s"
	blobUrlPattern      = "%s://%s/v2/%s/blobs/%s"
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
)

func getValueFromString(str string, substr string) string {

	startPos := strings.Index(str, substr)

	// including starting quote
	subStartPos := startPos + len(substr) + 1
	endPos := strings.Index(str[subStartPos:], "\"")
	endPos += subStartPos

	found := str[subStartPos:endPos]
	return found
}

// parse header and split in realm + service
// www-authenticate: Bearer realm="https://auth.docker.io/token",service="registry.docker.io"
// www-authenticate: Bearer realm="https://quay.io/v2/auth",service="quay.io"
// www-authenticate: Bearer realm="https://ghcr.io/token",service="ghcr.io",scope="repository:user/image:pull"
func getRealmService(header string) (string, string) {
	slog.Info("goregistry.getRealmService", "header", header)

	realm := getValueFromString(header, "realm=")
	service := getValueFromString(header, "service=")
	slog.Info("goregistry.getRealmService", "realm", realm, "service", service)

	return realm, service
}

func checkAuth(scheme string, tlsVerify bool, host string, repo string) (string, string, error) {

	// first check the v2 endpoint tot see if authentication is needed
	url := fmt.Sprintf(checkAuthUrlPattern, scheme, host)
	slog.Info("goregistry.checkAuth", "url", url)

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
        if ! tlsVerify {
                customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
        }

        client := &http.Client{Transport: customTransport}

	req, err := http.NewRequest("GET", url, nil)

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("goregistry.checkAuth", "http.Get error", err)
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		slog.Info("goregistry.checkAuth no authentication needed", "status", resp.Status)
		return "", "", nil // no authentication needed
	}
	if resp.StatusCode != 401 {
		// something else
		slog.Error("goregistry.checkAuth", "status", resp.Status)
		return "", "", errors.New(resp.Status)
	}
	// error 401, authentication needed
	// https://datatracker.ietf.org/doc/html/rfc6750#section-3
	authHeader := resp.Header.Get("www-authenticate")
	realm, service := getRealmService(authHeader)
	return realm, service, nil
}

// https://docs.docker.com/reference/api/registry/auth/
func getToken(realm string, tlsVerify bool, service string, repo string, user string, passwd string) TokenT {
	var dat TokenRespT

	url := fmt.Sprintf(getTokenUrlPattern, realm, service, repo)
	slog.Info("goregistry.getToken", "url", url)

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
        if ! tlsVerify {
                customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
        }

        client := &http.Client{Transport: customTransport}

	req, err := http.NewRequest("GET", url, nil)
	if user != "" {
		slog.Info("goregistry.getToken", "user", user)
		req.SetBasicAuth(user, passwd)
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("goregistry.getToken", "http.Get error", err)
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		slog.Info("goregistry.getToken", "status", resp.Status)
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("goregistry.getToken", "io.ReadAll error", err)
		return ""
	}

	err = json.Unmarshal(body, &dat)
	if err != nil {
		slog.Error("goregistry.getToken", "json.Unmarshal error", err)
		return ""
	}

	slog.Info("goregistry.getToken", "token(truncated)", dat.Token[:10], "expires_in", dat.ExpiresIn, "issued_at", dat.IssuedAt)
	return dat.Token
}

func getDigestFromImageIndex(scheme string, tlsVerify bool, host string, token TokenT, repo string, tag string) (string, error) {
	var dat ManifestsT

	url := fmt.Sprintf(manifestUrlPattern, scheme, host, repo, tag)
	slog.Info("goregistry.getDigestFromImageIndex", "url", url)

	// application/vnd.docker.distribution.manifest.list.v2+json added for gcr.io
	// application/vnd.oci.image.manifest.v1+json for dockerhub
	if err := jsonreq.GetJsonResp(tlsVerify, url, string(token), "application/vnd.docker.distribution.manifest.list.v2+json,application/vnd.oci.image.index.v1+json", &dat); err != nil {
		// not always present, so no error
		slog.Info("goregistry.getDigestFromImageIndex", "err", err)
		return "", err
	}
	// multi architecture manifest list
	if dat.MediaType != "application/vnd.oci.image.index.v1+json" &&
	   dat.MediaType != "application/vnd.docker.distribution.manifest.list.v2+json" {
		slog.Warn("goregistry.getDigestFromImageIndex", "MediaType", dat.MediaType)
	}

	b,_:=json.MarshalIndent(dat, "", "  ")
	fmt.Print(string(b))

	for _, entry := range dat.Manifest {
		if entry.Platform.Architecture == "amd64" {
			slog.Info("goregistry.getDigestFromImageIndex returning", "digest", entry.Digest, 
			            "arch", entry.Platform.Architecture, "created", entry.Annotations.Created, 
				    "url", entry.Annotations.Url, "version", entry.Annotations.Version)
			return entry.Digest, nil
		}
	}
	slog.Error("goregistry.getDigestFromImageIndex return architecture not found")
	return "", errors.New("not found")
}

func getDigestFromManifest(scheme string, tlsVerify bool, host string, token TokenT, repo string, digest string) (ConfigT, error) {
	var dat SingleT

	url := fmt.Sprintf(manifestUrlPattern, scheme, host, repo, digest)
	slog.Info("goregistry.getDigestFromManifest", "url", url)

	if err := jsonreq.GetJsonResp(tlsVerify, url, string(token), "application/vnd.docker.distribution.manifest.v2+json,application/vnd.oci.image.manifest.v1+json", &dat); err != nil {
		slog.Error("goregistry.getDigestFromManifest", "err", err)
		return ConfigT{}, err
	}
	// docker manifest
	if dat.MediaType != "application/vnd.oci.image.manifest.v1+json" && 
	   dat.MediaType != "application/vnd.docker.distribution.manifest.v2+json" {
		slog.Warn("goregistry.getDigestFromManifest", "MediaType", dat.MediaType)
	}

	b,_:=json.MarshalIndent(dat, "", "  ")
	fmt.Print(string(b))

	slog.Info("goregistry.getDigestFromManifest returning", "digest", dat.Config.Digest, "mediaType", dat.Config.MediaType)
	return dat.Config, nil
}

func getBlob(scheme string, tlsVerify bool, host string, config ConfigT, token TokenT, repo string, tag string) (time.Time, error) {
	var dat BlobT

	url := fmt.Sprintf(blobUrlPattern, scheme, host, repo, config.Digest)
	slog.Info("goregistry.getBlob", "url", url)

	//if err := jsonreq.GetJsonResp(tlsVerify, url, string(token), "application/vnd.oci.image.config.v1+json", &dat); err != nil {
	if err := jsonreq.GetJsonResp(tlsVerify, url, string(token), config.MediaType, &dat); err != nil {
		slog.Error("goregistry.getBlob", "err", err)
		return time.Time{}, err
	}

	b,_:=json.MarshalIndent(dat, "", "  ")
	fmt.Print(string(b))

	slog.Info("goregistry.getBlob", "repo", repo, "tag", tag, "digest", config.Digest, "mediaType", config.MediaType, "created", dat.Created)

	return dat.Created, nil
}

func (token TokenT) GetLastUpdate(scheme string, tlsVerify bool, host string, repo string, tag string) (time.Time, error) {
	digest1, err := getDigestFromImageIndex(scheme, tlsVerify, host, token, repo, tag)
	if err != nil {
		// there is no image index manifest, try a normal manifest
		digest1 = tag
	}
	// get manifest for specific arch
	config, err := getDigestFromManifest(scheme, tlsVerify, host, token, repo, digest1)
	if err != nil {
		slog.Error("goregistry.GetLastUpdate", "err", err)
		return time.Time{}, err
	}
	datum, err := getBlob(scheme, tlsVerify, host, config, token, repo, tag)
	if err != nil {
		slog.Error("goregistry.GetLastUpdate", "err", err)
		return time.Time{}, err
	}

	return datum, nil
}

func AcquireToken(scheme string, tlsVerify bool, host string, repo string, regcred secret.RegCredT) (TokenT, error) {
	var token TokenT

	realm, service, err := checkAuth(scheme, tlsVerify, host, repo)
	if err != nil {
		slog.Error("goregistry.AcquireToken", "checkAuth", err)
		return TokenT(""), err
	}
	// no authentication needed
	if realm != "" && service != "" {
		token = getToken(realm, tlsVerify, service, repo, regcred.User, regcred.Passwd)
	}
	return token, nil
}

