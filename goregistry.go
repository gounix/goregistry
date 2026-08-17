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
	"encoding/json"
	"log/slog"
	"time"
)

//func getDigestFromManifest(scheme string, tlsVerify bool, host string, token TokenT, repo string, tag string) (string, error) {
//	var digest string
//
//	url := fmt.Sprintf(manifestUrlPattern, scheme, host, repo, tag)
//	slog.Info("goregistry.getDigestFromManifest", "url", url)
//
//	customTransport := http.DefaultTransport.(*http.Transport).Clone()
//        if ! tlsVerify {
//                customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
//        }
//
//        client := &http.Client{Transport: customTransport}
//        req, err := http.NewRequest("GET", url, nil)
//	req.Header.Add("accept", acceptAll)
//
//        if token != "" {
//                req.Header.Add("Authorization", "Bearer " + string(token))
//        }
//
//        resp, err := client.Do(req)
//        if err != nil {
//                slog.Error("goregistry.getDigestFromManifest", "client.do error", err)
//                return "", err
//        }
//
//        defer resp.Body.Close()
//        if resp.StatusCode != 200 {
//                slog.Error("goregistry.getDigestFromManifest", "status", resp.Status)
//                str := fmt.Sprintf("status code %d", resp.StatusCode)
//                return "", errors.New(str)
//        }
//
//        if digest = resp.Header.Get("Docker-Content-Digest"); digest != "" {
//                slog.Info("goregistry.getDigestFromManifest", "Docker-Content-Digest", digest)
//        }
//        return digest, nil
//
//}
//func getConfigFromImageIndex(scheme string, tlsVerify bool, host string, token TokenT, repo string, tag string) (ConfigT, error) {
//	var dat JsonManifestListT
//	var config ConfigT
//
//	url := fmt.Sprintf(manifestUrlPattern, scheme, host, repo, tag)
//	slog.Info("goregistry.getConfigFromImageIndex", "url", url)
//
//	if err := gojsonreq.GetJsonResp(tlsVerify, url, string(token), acceptImageIndex, &dat); err != nil {
//		// not always present, so no error
//		slog.Info("goregistry.getConfigFromImageIndex", "err", err)
//		return ConfigT{}, err
//	}
//
//	for _, entry := range dat.Manifest {
//		if entry.Platform.Architecture == "amd64" {
//			slog.Info("goregistry.getConfigFromImageIndex returning", "digest", entry.Digest, 
//			            "arch", entry.Platform.Architecture, "created", entry.Annotations.Created, 
//				    "url", entry.Annotations.Url, "version", entry.Annotations.Version)
//				    config.Digest = entry.Digest
//				    config.MediaType = entry.MediaType
//			return config, nil
//		}
//	}
//	slog.Error("goregistry.getConfigFromImageIndex return architecture not found")
//	return ConfigT{}, errors.New("not found")
//}
//
//func getConfigFromManifest(scheme string, tlsVerify bool, host string, token TokenT, repo string, digest string) (ConfigT, error) {
//	var dat JsonManifestT
//
//	url := fmt.Sprintf(manifestUrlPattern, scheme, host, repo, digest)
//	slog.Info("goregistry.getConfigFromManifest", "url", url)
//
//	if err := gojsonreq.GetJsonResp(tlsVerify, url, string(token), acceptImageManifest, &dat); err != nil {
//		slog.Error("goregistry.getConfigFromManifest", "err", err)
//		return ConfigT{}, err
//	}
//
//	slog.Info("goregistry.getConfigFromManifest returning", "digest", dat.Config.Digest, "mediaType", dat.Config.MediaType)
//	slog.Info("goregistry.getConfigFromManifest", "dat", dat)
//	return dat.Config, nil
//}
//
//func getBlob(scheme string, tlsVerify bool, host string, config ConfigT, token TokenT, repo string, tag string) (time.Time, error) {
//	var dat BlobT
//
//	url := fmt.Sprintf(blobUrlPattern, scheme, host, repo, config.Digest)
//	slog.Info("goregistry.getBlob", "url", url)
//
//	//if err := gojsonreq.GetJsonResp(tlsVerify, url, string(token), "application/vnd.oci.image.config.v1+json", &dat); err != nil {
//	if err := gojsonreq.GetJsonResp(tlsVerify, url, string(token), config.MediaType, &dat); err != nil {
//		slog.Error("goregistry.getBlob", "err", err)
//		return time.Time{}, err
//	}
//
//	slog.Info("goregistry.getBlob", "repo", repo, "tag", tag, "digest", config.Digest, "mediaType", config.MediaType, "created", dat.Created)
//
//	return dat.Created, nil
//}

//func (token TokenT) GetLastUpdate(scheme string, tlsVerify bool, host string, repo string, tag string) (time.Time, error) {
//	var digest string
//
//	config, err := getConfigFromImageIndex(scheme, tlsVerify, host, token, repo, tag)
//	if err != nil {
//		// there is no image index manifest, try a normal manifest
//		digest = tag
//	} else {
//		digest = config.Digest
//	}
//	// get manifest for specific arch
//	config, err = getConfigFromManifest(scheme, tlsVerify, host, token, repo, digest)
//	if err != nil {
//		slog.Error("goregistry.GetLastUpdate", "err", err)
//		return time.Time{}, err
//	}
//	datum, err := getBlob(scheme, tlsVerify, host, config, token, repo, tag)
//	if err != nil {
//		slog.Error("goregistry.GetLastUpdate", "err", err)
//		return time.Time{}, err
//	}
//
//	return datum, nil
//}

func getArchDigest(manifest JsonManifestListT) (string, string) {
	for _, entry := range manifest.Manifest {
		if entry.Platform.Architecture == "amd64" && entry.Platform.Os == "linux" {
			return entry.MediaType, entry.Digest
		}
	}
	return "", ""
}

func (registry RegistryT) GetLastUpdate(tag string) (time.Time, error) {
	var digest string
	var mediaType string
	var dat BlobT

	retML, err := registry.GetManifestList(tag)
	if err != nil {
		// there is no image index manifest, try a normal manifest
		digest = tag
		mediaType = "application/vnd.docker.distribution.manifest.list.v2+json,application/vnd.oci.image.index.v1+json"
	} else {
		mediaType, digest = getArchDigest(retML.Json)
	}
	// get manifest for specific arch
	retM, err := registry.GetManifest(mediaType, digest)
	blob, err := registry.GetBlob(retM.Json.Config.MediaType, retM.Json.Config.Digest)
	if err != nil {
		slog.Error("goregistry.GetLastUpdate", "err", err)
		return time.Time{}, err
	}

	json.Unmarshal(blob, &dat)
	slog.Info("goregistry.GetLastUpdate", "date", dat.Created)
	return dat.Created, nil
}
