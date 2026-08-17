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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func (registry *RegistryT) getManifest(tag string, accept string) (string, []byte, error) {

	url := fmt.Sprintf(manifestUrlPattern, registry.Scheme, registry.Host, registry.Image, tag)
	slog.Info("goregistry.getManifest", "url", url)

        customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: ! registry.TlsVerify}

        client := &http.Client{Transport: customTransport}
        req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("accept", accept)

        if registry.Token != "" {
                req.Header.Add("Authorization", "Bearer " + string(registry.Token))
        }

	resp, err := client.Do(req)
        if err != nil {
                slog.Error("goregistry.getManifest", "client.do error", err)
                return "", []byte{}, err
        }

        defer resp.Body.Close()
        if resp.StatusCode != 200 {
                slog.Error("goregistry.getManifest", "status", resp.Status)
                str := fmt.Sprintf("status code %d", resp.StatusCode)
                return "", []byte{}, errors.New(str)
        }

        body, err := io.ReadAll(resp.Body)
        if err != nil {
                slog.Error("goregistry.getManifest", "io.ReadAll error", err)
                return "", []byte{}, err
        }
	digest := resp.Header.Get("docker-content-digest")

	return digest, body, err
}

func (registry *RegistryT) GetManifestList(tag string) (ManifestListT, error) {
	var dat JsonManifestListT
	var ret ManifestListT

	digest, body, err := registry.getManifest(tag, acceptImageIndex)
	if err != nil {
                slog.Error("goregistry.GetManifestList", "err", err)
		return ManifestListT{}, err
	}

	err = json.Unmarshal(body, &dat)
	ret.Digest = digest
	ret.Raw = body
	ret.Json = dat
	return ret, err
}

func (registry *RegistryT) GetManifest(accept string, tag string) (ManifestT, error) {
	var dat JsonManifestT
	var ret ManifestT

	digest, body, err := registry.getManifest(tag, accept)
	if err != nil {
                slog.Error("goregistry.GetManifest", "err", err)
		return ManifestT{}, err
	}

	err = json.Unmarshal(body, &dat)
	ret.Digest = digest
	ret.Raw = body
	ret.Json = dat
	return ret, err
}

func (registry *RegistryT) PutManifest(contentType string, content []byte, tag string) error {

	url := fmt.Sprintf(manifestUrlPattern, registry.Scheme, registry.Host, registry.Image, tag)
	slog.Info("goregistry.PutManifest", "url", url)

        customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: ! registry.TlsVerify}

        client := &http.Client{Transport: customTransport}
        req, err := http.NewRequest("PUT", url, bytes.NewBuffer(content))

	//req.Header.Add("accept-encoding", "gzip") // wordt door buildah gestuurd
	req.Header.Add("content-type", contentType)
	slog.Info("goregistry.PutManifest", "content-type", contentType)

        if registry.Token != "" {
                req.Header.Add("Authorization", "Bearer " + string(registry.Token))
        }

	resp, err := client.Do(req)
        if err != nil {
                slog.Error("goregistry.PutManifest", "client.do error", err)
                return err
        }

        defer resp.Body.Close()
        if resp.StatusCode != 201 {
                slog.Error("goregistry.PutManifest", "status", resp.Status)
                str := fmt.Sprintf("status code %d", resp.StatusCode)
                return errors.New(str)
        }

	slog.Info("goregistry.PutManifest OK")
	return nil
}

