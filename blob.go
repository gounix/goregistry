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
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)


//func (token TokenT) ReadBinaryBlob(scheme string, tlsVerify bool, host string, config ConfigT, repo string, ch chan []byte) error {
//	url := fmt.Sprintf(blobUrlPattern, scheme, host, repo, config.Digest)
//        slog.Info("goregistry.readBinaryBlob", "url", url)
//
//        customTransport := http.DefaultTransport.(*http.Transport).Clone()
//	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: ! tlsVerify}
//
//        client := &http.Client{Transport: customTransport}
//        req, err := http.NewRequest("GET", url, nil)
//        if config.MediaType != "" {
//                req.Header.Add("accept", config.MediaType)
//        }
//	req.Header.Add("docker-distribution-api-version", "registry/2.0")
//
//        if token != "" {
//                req.Header.Add("Authorization", "Bearer " + string(token))
//        }
//
//	resp, err := client.Do(req)
//        if err != nil {
//                slog.Error("goregistry.readBinaryBlob", "client.do error", err)
//                return err
//        }
//
//        defer resp.Body.Close()
//        if resp.StatusCode != 200 {
//                slog.Error("goregistry.readBinaryBlob", "status", resp.Status)
//                str := fmt.Sprintf("status code %d", resp.StatusCode)
//                return errors.New(str)
//        }
//        body, err := io.ReadAll(resp.Body)
//        if err != nil {
//                slog.Error("goregistry.readBinaryBlob", "io.ReadAll error", err)
//                return err
//        }
//	ch <- body
//	close(ch)
//
//        return nil
//}

func (registry RegistryT) GetBlob(mediaType string, digest string) ([]byte, error) {
	url := fmt.Sprintf(blobUrlPattern, registry.Scheme, registry.Host, registry.Image, digest)
        slog.Info("goregistry.GetBlob", "url", url)

        customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: ! registry.TlsVerify}

        client := &http.Client{Transport: customTransport}
        req, err := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", mediaType)
	req.Header.Add("docker-distribution-api-version", "registry/2.0")

        if registry.Token != "" {
                req.Header.Add("Authorization", "Bearer " + string(registry.Token))
        }

	resp, err := client.Do(req)
        if err != nil {
                slog.Error("goregistry.GetBlob", "client.do error", err)
                return []byte{}, err
        }

        defer resp.Body.Close()
        if resp.StatusCode != 200 {
                slog.Error("goregistry.GetBlob", "status", resp.Status)
                str := fmt.Sprintf("status code %d", resp.StatusCode)
                return []byte{}, errors.New(str)
        }
        body, err := io.ReadAll(resp.Body)
        if err != nil {
                slog.Error("goregistry.GetBlob", "io.ReadAll error", err)
                return []byte{}, err
        }

        return body, nil
}

func (registry RegistryT) CheckBlob(mediaType string, digest string) (int, error) {
	url := fmt.Sprintf(blobUrlPattern, registry.Scheme, registry.Host, registry.Image, digest)
        slog.Info("goregistry.GetBlob", "url", url)

        customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: ! registry.TlsVerify}

        client := &http.Client{Transport: customTransport}
        req, err := http.NewRequest("HEAD", url, nil)

	req.Header.Add("accept", mediaType)
	req.Header.Add("docker-distribution-api-version", "registry/2.0")

        if registry.Token != "" {
                req.Header.Add("Authorization", "Bearer " + string(registry.Token))
        }

	resp, err := client.Do(req)
        if err != nil {
                slog.Error("goregistry.CheckBlob", "client.do error", err)
                return 0, err
        }

        defer resp.Body.Close()

	slog.Info("goregistry.CheckBlob", "code", resp.StatusCode)
        return resp.StatusCode, nil
}

// start the blob upload, this will return the upload location
func (registry RegistryT) PostBlob(mediaType string, digest string) (string, error) {
	url := fmt.Sprintf(blobUrlPattern, registry.Scheme, registry.Host, registry.Image, "uploads/")
        slog.Info("goregistry.PostBlob", "url", url)

        customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: ! registry.TlsVerify}

        client := &http.Client{Transport: customTransport}
        req, err := http.NewRequest("POST", url, nil)

	//req.Header.Add("content-type", "application/octet-stream") // mediaType)
	req.Header.Add("docker-distribution-api-version", "registry/2.0")

        if registry.Token != "" {
                req.Header.Add("Authorization", "Bearer " + string(registry.Token))
        }

	resp, err := client.Do(req)
        if err != nil {
                slog.Error("goregistry.PostBlob", "client.do error", err)
                return "", err
        }
	location := resp.Header.Get("Location")
	slog.Info("goregistry.PostBlob", "location", location)

        defer resp.Body.Close()
        if resp.StatusCode != 202 { // accepted
                slog.Error("goregistry.PostBlob", "status", resp.Status)
                str := fmt.Sprintf("status code %d", resp.StatusCode)
                return "", errors.New(str)
        }

        slog.Info("goregistry.PostBlob success")
        return location, nil
}

// do the real uploading to the location we got from POST
func (registry RegistryT) PatchBlob(location string, mediaType string, digest string, blob []byte) (string, error) {

	url := fmt.Sprintf("%s://%s%s", registry.Scheme, registry.Host, location)
	// gitea returns location starting with /v2, docker registry uses full url
	if strings.HasPrefix(location, "http") {
		url = fmt.Sprintf("%s", location)
	}
        slog.Info("goregistry.PatchBlob", "url", url)

        customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: ! registry.TlsVerify}

        client := &http.Client{Transport: customTransport}
        req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(blob))

	req.Header.Add("content-type", "application/octet-stream")
	req.Header.Add("docker-distribution-api-version", "registry/2.0")

        if registry.Token != "" {
                req.Header.Add("Authorization", "Bearer " + string(registry.Token))
        }

	resp, err := client.Do(req)
        if err != nil {
                slog.Error("goregistry.PatchBlob", "client.do error", err)
		return "", err
        }

	locationNew := resp.Header.Get("Location")
	slog.Info("goregistry.PatchBlob", "location", locationNew)

        defer resp.Body.Close()
        if resp.StatusCode != 202 { // chunk accepted and stored
                slog.Error("goregistry.PatchBlob", "status", resp.Status)
                str := fmt.Sprintf("status code %d", resp.StatusCode)
                return "", errors.New(str)
        }

        slog.Info("goregistry.PatchBlob success")
        return locationNew, nil
}

// if the PATCH went wrong, abort the upload
func (registry RegistryT) DelBlob(location string, mediaType string, digest string) error {

	url := registry.createLocationUrl(location, digest)
        slog.Info("goregistry.DelBlob", "url", url)

        customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: ! registry.TlsVerify}

        client := &http.Client{Transport: customTransport}
        req, err := http.NewRequest("DEL", url, nil) // bytes.NewBuffer(blob))

	req.Header.Add("content-type", "application/octet-stream") // mediaType)
	req.Header.Add("docker-distribution-api-version", "registry/2.0")

        if registry.Token != "" {
                req.Header.Add("Authorization", "Bearer " + string(registry.Token))
        }

	resp, err := client.Do(req)
        if err != nil {
                slog.Error("goregistry.DelBlob", "client.do error", err)
                return err
        }

        defer resp.Body.Close()
        if resp.StatusCode != 204 { // upload succesfully cancelled
                slog.Error("goregistry.DelBlob", "status", resp.Status)
                str := fmt.Sprintf("status code %d", resp.StatusCode)
                return errors.New(str)
        }

        slog.Info("goregistry.DelBlob success")
        return nil
}

func (registry RegistryT) createLocationUrl(location string, digest string) string {

	// docker registry uses a location that has a query(?) so we need to append to the query
	// gitea not, so we have to create the query
	separator := "?"
	if strings.Contains(location, "?") {
		separator = "&"
	}

	url := fmt.Sprintf("%s://%s%s%sdigest=%s", registry.Scheme, registry.Host, location, separator, digest)
	// gitea returns location starting with /v2, docker registry uses full url
	if strings.HasPrefix(location, "http") {
		url = fmt.Sprintf("%s%sdigest=%s", location, separator, digest)
	}

	return url
}

// finalize the upload after a succesful PATCH
func (registry RegistryT) PutBlob(location string, mediaType string, digest string, blob []byte) error {

	url := registry.createLocationUrl(location, digest)
        slog.Info("goregistry.PutBlob", "url", url)

        customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: ! registry.TlsVerify}

        client := &http.Client{Transport: customTransport}
        req, err := http.NewRequest("PUT", url, nil)

	req.Header.Add("content-type", "application/octet-stream")
	req.Header.Add("docker-distribution-api-version", "registry/2.0")

        if registry.Token != "" {
                req.Header.Add("Authorization", "Bearer " + string(registry.Token))
        }

	resp, err := client.Do(req)
        if err != nil {
                slog.Error("goregistry.PutBlob", "client.do error", err)
                return err
        }

        defer resp.Body.Close()
        if resp.StatusCode != 201 {
                slog.Error("goregistry.PutBlob", "status", resp.Status)
                str := fmt.Sprintf("status code %d", resp.StatusCode)
                return errors.New(str)
        }

        slog.Info("goregistry.PutBlob success")
        return nil
}
