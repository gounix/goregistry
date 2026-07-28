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
        "fmt"
        "github.com/gounix/gojsonreq"
	"log/slog"
	"regexp"
)
const (
        versionUrlPattern  = "%s://%s/v2/%s/tags/list"
)

type (
	TagsT struct {
                Name string   `json:"name"`
                Tags []string `json:"tags"`
        }
)

func (token TokenT) GetVersions(scheme string, tlsVerify bool, host string, repo string, filter string) ([]string, error) {
        var dat TagsT
	var filtered []string

        url := fmt.Sprintf(versionUrlPattern, scheme, host, repo)
        slog.Info("goregistry.getVersions", "url", url)

        if err := gojsonreq.GetJsonResp(tlsVerify, url, string(token), "", &dat); err != nil {
                slog.Error("goregistry.getVersions", "err", err)
                return []string{}, err
        }

	for _, entry := range dat.Tags {
		matched, err := regexp.Match(filter, []byte(entry))
		if err == nil && matched {
			filtered = append(filtered, entry)
		}
	}

        slog.Info("goregistry.getVersions", "repo", repo, "versions", filtered)
        return filtered, nil
}

