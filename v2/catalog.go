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
	"log/slog"
	"regexp"
)

// the catalog api is 
//  - available in docker v2 registries
//  - disabled or not yet implemented by registry-1.docker.io, quay.io, ghcr.io and gcr.io
func (registry RegistryT) GetCatalog(filter string, negateFilter bool) ([]string, error) {
	var filtered []string
	var err error

	registry.RenewToken()
        baseUrl := fmt.Sprintf(catalogBaseUrlPattern, registry.Scheme, registry.Host)
	//linkUrl := fmt.Sprintf(versionLinkUrlPattern, registry.Image)
	linkUrl := "/v2/_catalog"
        slog.Info("goregistry.GetCatalog", "baseUrl", baseUrl)//, "linkUrl", linkUrl)

	for {
		var dat CatalogT

		url := baseUrl + linkUrl
		linkUrl, err = fetchPage(registry.TlsVerify, url, string(registry.Token), "", &dat)
		if err != nil {
			slog.Error("goregistry.GetCatalog", "err", err)
			break
		}
		slog.Info("goregistry.GetCatalog", "baseUrl", baseUrl, "linkUrl", linkUrl)

		for _, entry := range dat.Repositories {
			slog.Info("goregistry.GetCatalog", "entry", entry)
			matched, err := regexp.Match(filter, []byte(entry))
			if err == nil && ((matched && ! negateFilter) || (! matched && negateFilter)) {
				filtered = append(filtered, entry)
			}
		}
		if linkUrl == "" {
			slog.Info("goregistry.GetCatalog last page")
			break
		}
	}
	return filtered, nil
}
