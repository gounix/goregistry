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
	"github.com/gounix/gosecret"
	"log/slog"
	"net/http"
)

func (token TokenT) deleteByDigest(scheme string, tlsVerify bool, host string, repo string, digest string) error {

	url := fmt.Sprintf(deleteUrlPattern, scheme, host, repo, digest)
	slog.Info("goregistry.deleteByDigest", "url", url)

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
        if ! tlsVerify {
                customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
        }

        client := &http.Client{Transport: customTransport}
        req, err := http.NewRequest("DELETE", url, nil)

	req.Header.Add("Authorization", "Bearer "+string(token))

	resp, err := client.Do(req)
        if err != nil {
                slog.Error("goregistry.deleteByDigest", "client.do error", err)
                return err
        }

	defer resp.Body.Close()
        if resp.StatusCode != 202 {
                slog.Error("goregistry.deleteByDigest", "status", resp.Status)
                str := fmt.Sprintf("status code %d", resp.StatusCode)
                return errors.New(str)
        }

	slog.Info("goregistry.deleteByDigest", "status", resp.Status)
	return nil
}

func (token TokenT) DeleteImage(scheme string, tlsVerify bool, host string, repo string, tag string) error {

	digest, err := getDigestFromManifest(scheme, tlsVerify, host, token, repo, tag)
	if err != nil {
		slog.Error("goregistry.DeleteImage", "err", err)
		return err
	}
	return token.deleteByDigest(scheme, tlsVerify, host, repo, digest)
}

func AcquireDeleteToken(scheme string, tlsVerify bool, host string, repo string, regcred gosecret.RegCredT) (TokenT, error) {
	var token TokenT

	realm, service, err := checkAuth(scheme, tlsVerify, host, repo)
	if err != nil {
		slog.Error("goregistry.AcquireToken", "checkAuth", err)
		return TokenT(""), err
	}
	if realm != "" && service != "" {
		token = getToken(realm, tlsVerify, service, repo, regcred.User, regcred.Passwd, "pull,push,delete")
	}
	return token, nil
}
