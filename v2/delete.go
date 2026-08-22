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
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

func (registry RegistryT) deleteByDigest(digest string) error {

	url := fmt.Sprintf(deleteUrlPattern, registry.Scheme, registry.Host, registry.Image, digest)
	slog.Info("goregistry.deleteByDigest", "url", url)

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: ! registry.TlsVerify}

        client := &http.Client{Transport: customTransport}
        req, err := http.NewRequest("DELETE", url, nil)

	req.Header.Add("Authorization", "Bearer "+string(registry.Token))

	resp, err := client.Do(req)
        if err != nil {
                slog.Error("goregistry.deleteByDigest", "client.do error", err)
                return err
        }

	defer resp.Body.Close()
        if resp.StatusCode != 202 {
                slog.Error("goregistry.deleteByDigest", "status", resp.Status)
                //str := fmt.Sprintf("status code %d", resp.StatusCode)
                return errors.New(resp.Status)
        }

	slog.Info("goregistry.deleteByDigest", "status", resp.Status)
	return nil
}

func (registry RegistryT) DeleteImage(tag string) error {
	var digest string

	registry.RenewToken()
        retML, err := registry.GetManifestList(tag)
        if err != nil {
                // there is no image index manifest, try a normal manifest
		retM, err := registry.GetManifest(acceptImageManifest, tag)
		if err != nil {
			slog.Error("goregistry.DeleteImage", "err", err)
			return err
		}
		digest = retM.Digest
	} else {
		digest = retML.Digest
	}
	return registry.deleteByDigest(digest)
}

