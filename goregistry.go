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
	"github.com/gounix/gojsonreq"
	"github.com/gounix/gosecret"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
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
func getToken(realm string, tlsVerify bool, service string, repo string, user string, passwd string, usage string) TokenT {
	var dat TokenRespT

	url := fmt.Sprintf(getTokenUrlPattern, realm, service, repo, usage)
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

func getDigestFromManifest(scheme string, tlsVerify bool, host string, token TokenT, repo string, tag string) (string, error) {
	var digest string

	url := fmt.Sprintf(manifestUrlPattern, scheme, host, repo, tag)
	slog.Info("goregistry.getDigestFromManifest", "url", url)

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
        if ! tlsVerify {
                customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
        }

        client := &http.Client{Transport: customTransport}
        req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("accept", acceptAll)

        if token != "" {
                req.Header.Add("Authorization", "Bearer " + string(token))
        }

        resp, err := client.Do(req)
        if err != nil {
                slog.Error("goregistry.getDigestFromManifest", "client.do error", err)
                return "", err
        }

        defer resp.Body.Close()
        if resp.StatusCode != 200 {
                slog.Error("goregistry.getDigestFromManifest", "status", resp.Status)
                str := fmt.Sprintf("status code %d", resp.StatusCode)
                return "", errors.New(str)
        }

        if digest = resp.Header.Get("Docker-Content-Digest"); digest != "" {
                slog.Info("goregistry.getDigestFromManifest", "Docker-Content-Digest", digest)
        }
        return digest, nil

}
func getConfigFromImageIndex(scheme string, tlsVerify bool, host string, token TokenT, repo string, tag string) (ConfigT, error) {
	var dat ManifestsT
	var config ConfigT

	url := fmt.Sprintf(manifestUrlPattern, scheme, host, repo, tag)
	slog.Info("goregistry.getConfigFromImageIndex", "url", url)

	if err := gojsonreq.GetJsonResp(tlsVerify, url, string(token), acceptImageIndex, &dat); err != nil {
		// not always present, so no error
		slog.Info("goregistry.getConfigFromImageIndex", "err", err)
		return ConfigT{}, err
	}

	for _, entry := range dat.Manifest {
		if entry.Platform.Architecture == "amd64" {
			slog.Info("goregistry.getConfigFromImageIndex returning", "digest", entry.Digest, 
			            "arch", entry.Platform.Architecture, "created", entry.Annotations.Created, 
				    "url", entry.Annotations.Url, "version", entry.Annotations.Version)
				    config.Digest = entry.Digest
				    config.MediaType = entry.MediaType
			return config, nil
		}
	}
	slog.Error("goregistry.getConfigFromImageIndex return architecture not found")
	return ConfigT{}, errors.New("not found")
}

func getConfigFromManifest(scheme string, tlsVerify bool, host string, token TokenT, repo string, digest string) (ConfigT, error) {
	var dat SingleT

	url := fmt.Sprintf(manifestUrlPattern, scheme, host, repo, digest)
	slog.Info("goregistry.getConfigFromManifest", "url", url)

	if err := gojsonreq.GetJsonResp(tlsVerify, url, string(token), acceptImageManifest, &dat); err != nil {
		slog.Error("goregistry.getConfigFromManifest", "err", err)
		return ConfigT{}, err
	}

	slog.Info("goregistry.getConfigFromManifest returning", "digest", dat.Config.Digest, "mediaType", dat.Config.MediaType)
	return dat.Config, nil
}

func getBlob(scheme string, tlsVerify bool, host string, config ConfigT, token TokenT, repo string, tag string) (time.Time, error) {
	var dat BlobT

	url := fmt.Sprintf(blobUrlPattern, scheme, host, repo, config.Digest)
	slog.Info("goregistry.getBlob", "url", url)

	//if err := gojsonreq.GetJsonResp(tlsVerify, url, string(token), "application/vnd.oci.image.config.v1+json", &dat); err != nil {
	if err := gojsonreq.GetJsonResp(tlsVerify, url, string(token), config.MediaType, &dat); err != nil {
		slog.Error("goregistry.getBlob", "err", err)
		return time.Time{}, err
	}

	slog.Info("goregistry.getBlob", "repo", repo, "tag", tag, "digest", config.Digest, "mediaType", config.MediaType, "created", dat.Created)

	return dat.Created, nil
}

func (token TokenT) GetLastUpdate(scheme string, tlsVerify bool, host string, repo string, tag string) (time.Time, error) {
	var digest string

	config, err := getConfigFromImageIndex(scheme, tlsVerify, host, token, repo, tag)
	if err != nil {
		// there is no image index manifest, try a normal manifest
		digest = tag
	} else {
		digest = config.Digest
	}
	// get manifest for specific arch
	config, err = getConfigFromManifest(scheme, tlsVerify, host, token, repo, digest)
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

func AcquireToken(scheme string, tlsVerify bool, host string, repo string, regcred gosecret.RegCredT) (TokenT, error) {
	var token TokenT

	realm, service, err := checkAuth(scheme, tlsVerify, host, repo)
	if err != nil {
		slog.Error("goregistry.AcquireToken", "checkAuth", err)
		return TokenT(""), err
	}
	// no authentication needed
	if realm != "" && service != "" {
		token = getToken(realm, tlsVerify, service, repo, regcred.User, regcred.Passwd, "pull")
	}
	return token, nil
}


