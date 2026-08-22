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
)

func (registry *RegistryT) getToken(realm string, service string) error {
        var dat TokenRespT

        url := fmt.Sprintf(getTokenUrlPattern, realm, service, registry.Image, registry.TokenScope)
        slog.Info("goregistry.getToken", "url", url)

        customTransport := http.DefaultTransport.(*http.Transport).Clone()
        customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: ! registry.TlsVerify}

        client := &http.Client{Transport: customTransport}

        req, err := http.NewRequest("GET", url, nil)
        if registry.Regcred.User != "" {
                slog.Info("goregistry.getToken", "user", registry.Regcred.User)
                req.SetBasicAuth(registry.Regcred.User, registry.Regcred.Passwd)
        }

        resp, err := client.Do(req)
        if err != nil {
                slog.Error("goregistry.getToken", "http.Get error", err)
                return err 
        }
        defer resp.Body.Close()
        if resp.StatusCode != 200 {
                slog.Info("goregistry.getToken", "status", resp.Status)
                return errors.New(resp.Status)
        }

        body, err := io.ReadAll(resp.Body)
        if err != nil {
                slog.Error("goregistry.getToken", "io.ReadAll error", err)
                return err
        }

        err = json.Unmarshal(body, &dat)
        if err != nil {
                slog.Error("goregistry.getToken", "json.Unmarshal error", err)
                return err
        }

	start := dat.Token[:5]
	end := dat.Token[len(dat.Token)-5:]
        slog.Info("goregistry.getToken", "token(truncated)", fmt.Sprintf("%s...%s", start, end), "expires_in", dat.ExpiresIn, "issued_at", dat.IssuedAt)
	registry.Token = dat.Token
	registry.FullToken = dat
        return  nil
}

func getValueFromString(str string, substr string) string {

        startPos := strings.Index(str, substr)

        // including starting quote
        subStartPos := startPos + len(substr) + 1
        endPos := strings.Index(str[subStartPos:], "\"")
        endPos += subStartPos

        found := str[subStartPos:endPos]
        return found
}

func getRealmService(header string) (string, string) {
        slog.Info("goregistry.getRealmService", "header", header)

        realm := getValueFromString(header, "realm=")
        service := getValueFromString(header, "service=")
        slog.Info("goregistry.getRealmService", "realm", realm, "service", service)

        return realm, service
}

func (registry *RegistryT) checkAuth() (string, string, error) {

        // first check the v2 endpoint tot see if authentication is needed
        url := fmt.Sprintf(checkAuthUrlPattern, registry.Scheme, registry.Host)
        slog.Info("goregistry.checkAuth", "url", url)

        customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: ! registry.TlsVerify}

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

func (registry *RegistryT) acquireTokenCommon() error {

        realm, service, err := registry.checkAuth()
        if err != nil {
                slog.Error("goregistry.acquireTokenCommon", "checkAuth", err)
                return err
        }
        if realm != "" && service != "" {
		err := registry.getToken(realm, service)
		if err != nil {
			slog.Error("goregistry.acquireTokenCommon", "getToken", err)
			return err
		}
        }
        return nil
}

func (registry *RegistryT) AcquireToken() error {
	registry.TokenScope = "pull"
	return registry.acquireTokenCommon()
}

func (registry *RegistryT) AcquirePushToken() error {
	registry.TokenScope = "pull,push"
	return registry.acquireTokenCommon()
}

func (registry *RegistryT) AcquireDeleteToken() error {
	registry.TokenScope = "pull,push,delete"
	return registry.acquireTokenCommon()
}

func (registry *RegistryT) RenewToken() error {
	// check if token still valid
	expire := registry.FullToken.IssuedAt.Add(time.Duration(registry.FullToken.ExpiresIn) * time.Second)
	if expire.After(time.Now()) {
		slog.Info("goregistry.RenewToken token still valid")
		return nil
	}
	return registry.acquireTokenCommon()
}
