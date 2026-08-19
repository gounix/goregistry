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
	"regexp"
	"strings"
)

func parseLinkHeader(linkHeader string) string {
	// Link: </v2/prometheus/prometheus/tags/list?n=100&last=v2.20.0-rc.1>; rel="next"
	left := strings.Index(linkHeader, "<")
	if left < 0 {
		slog.Error("goregistry.parseLinkHeader left anchor not found", "linkHeader", linkHeader)
		return ""
	}

	right := strings.Index(linkHeader, ">")
	if right < 0 {
		slog.Error("goregistry.parseLinkHeader right anchor not found", "linkHeader", linkHeader)
		return ""
	}
	return linkHeader[left+1:right]
}

func fetchPage(tlsVerify bool, url string, token string, accept string, dat any) (string, error) {

        customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: ! tlsVerify}

        client := &http.Client{Transport: customTransport}
        req, err := http.NewRequest("GET", url, nil)
        if accept != "" {
                req.Header.Add("accept", accept)
        }

        if token != "" {
                req.Header.Add("Authorization", "Bearer "+token)
        }

	slog.Info("goregistry.fetchPage", "url", url)
        resp, err := client.Do(req)
        if err != nil {
                slog.Error("goregistry.fetchPage", "client.do error", err)
                return "", err
        }

        defer resp.Body.Close()
        if resp.StatusCode != 200 {
                slog.Error("goregistry.fetchPage", "status", resp.Status)
                //str := fmt.Sprintf("status code %d", resp.StatusCode)
                return "", errors.New(resp.Status)
        }

	linkHeader := resp.Header.Get("Link")
	link := ""
	if linkHeader != "" {
		link = parseLinkHeader(linkHeader)
	}

        slog.Info("goregistry.fetchPage", "ContentLength", resp.ContentLength)
        body, err := io.ReadAll(resp.Body)
        if err != nil {
                slog.Error("goregistry.fetchPage", "io.ReadAll error", err)
                return "", err
        }

        return link, json.Unmarshal(body, dat)
}

func (registry RegistryT) GetVersions(filter string, negateFilter bool) ([]string, error) {
	var filtered []string
	var err error

        baseUrl := fmt.Sprintf(versionBaseUrlPattern, registry.Scheme, registry.Host)
        linkUrl := fmt.Sprintf(versionLinkUrlPattern, registry.Image)
        slog.Info("goregistry.getVersions", "baseUrl", baseUrl, "linkUrl", linkUrl)

	for {
		var dat TagsT

		url := baseUrl + linkUrl
		linkUrl, err = fetchPage(registry.TlsVerify, url, string(registry.Token), "", &dat)
		if err != nil {
			slog.Error("goregistry.getVersions", "err", err)
			break
		}
		slog.Info("goregistry.getVersions", "baseUrl", baseUrl, "linkUrl", linkUrl)

		for _, entry := range dat.Tags {
			matched, err := regexp.Match(filter, []byte(entry))
			if err == nil && ((matched && ! negateFilter) || (! matched && negateFilter)) {
				filtered = append(filtered, entry)
			}
		}
		if linkUrl == "" {
			slog.Info("goregistry.getVersions last page")
			break
		}
	}
	return filtered, nil
}
