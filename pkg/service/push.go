package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/luigizuccarelli/golang-container-tools/pkg/schema"
)

// OCIPushToRegistry - pushes a local OCI image to remote registry
func OCIPushToRegistry(ss schema.ServiceSchema) error {

	var ba *schema.BasicAuth
	var ociManifest = &schema.OCIImageManifest{SchemaVersion: 2}
	var layers []schema.Layer

	repo, err := name.NewRepository(ss.Image)
	if err != nil {
		return err
	}

	// Fetch credentials based on your docker config file, which is $HOME/.docker/config.json or $DOCKER_CONFIG.
	auth, err := authn.DefaultKeychain.Resolve(repo.Registry)
	if err != nil {
		return err
	}

	// Construct an http.Client that is authorized to push
	scopes := []string{repo.Scope(transport.PushScope)}
	t, err := transport.New(repo.Registry, auth, http.DefaultTransport, scopes)
	if err != nil {
		return err
	}
	client := &http.Client{Transport: t}

	// read the index.json
	var index = schema.Index{}
	data, err := ioutil.ReadFile(ss.Path + indexJSON)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &index)

	// read the blobs
	files, err := ioutil.ReadDir(ss.Path + blobsPath)
	if err != nil {
		return err
	}

	var errs []error
	ch := make(chan byte, 1)
	for _, file := range files {

		go func(file string) {

			// set up the request
			req, err := http.NewRequest(http.MethodPost, ss.URL+blobs+uploads, nil)
			if ss.Auth {
				ba, _ = GetBasicAuthCredentials()
				req.SetBasicAuth(ba.User, ba.Password)
			}
			if err != nil {
				errs = append(errs, err)
			}
			resp, err := client.Do(req)
			if err != nil {
				errs = append(errs, err)
			}
			defer resp.Body.Close()
			uuid := resp.Header.Get("Docker-Upload-UUID")
			location, err := url.QueryUnescape(resp.Header.Get("Location"))
			if err != nil {
				errs = append(errs, err)
			}
			fmt.Println("INFO: POST response: " + resp.Status)
			fmt.Println("INFO: POST location: ", location)
			fmt.Println("INFO: POST uuid: ", uuid)

			data, err := ioutil.ReadFile(ss.Path + blobsPath + file)
			if err != nil {
				errs = append(errs, err)
			}

			// get ths sha256 digest for each layer
			h := sha256.New()
			if _, err := io.Copy(h, bytes.NewBuffer(data)); err != nil {
				errs = append(errs, err)
			}

			// we now have the digest
			digestID := h.Sum(nil)
			fmt.Println("INFO: POST digestID: ", hex.EncodeToString(digestID))
			manifest := index.Manifests[0].Digest[7:]

			// check to see if the blob exists
			req, err = http.NewRequest(http.MethodHead, ss.URL+blobs+hex.EncodeToString(digestID), nil)
			if ss.Auth {
				req.SetBasicAuth(ba.User, ba.Password)
			}
			if err != nil {
				errs = append(errs, err)
			}
			resp, err = client.Do(req)
			if err != nil {
				errs = append(errs, err)
			}
			defer resp.Body.Close()
			fmt.Println("INFO: HEAD response: " + resp.Status)

			if resp.StatusCode != http.StatusOK {
				url := location + "&digest=" + SHA256 + hex.EncodeToString(digestID)
				fmt.Println("INFO: Uploading blob ", hex.EncodeToString(digestID))
				reqUpload, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(data))
				if err != nil {
					errs = append(errs, err)
				}
				if ss.Auth {
					reqUpload.SetBasicAuth(ba.User, ba.Password)
				}
				reqUpload.Header.Set("Content-Length", fmt.Sprintf("%d", len(data)))
				reqUpload.Header.Set("Content-Type", "application/octet-stream")
				respUpload, err := client.Do(reqUpload)
				if err != nil {
					errs = append(errs, err)
				}
				defer respUpload.Body.Close()
				fmt.Println("INFO: PUT upload response: " + respUpload.Status)
				if respUpload.StatusCode <= 200 || respUpload.StatusCode >= 300 {
					errs = append(errs, fmt.Errorf("response from data upload %d", respUpload.StatusCode))
				}
				// all good we append to layer struct and update config
				if manifest == file {
					ociManifest.Config.MediaType = "application/vnd.oci.image.config.v1+json"
					ociManifest.Config.Digest = SHA256 + hex.EncodeToString(digestID)
					ociManifest.Config.Size = len(data)

				} else {
					layer := schema.Layer{Digest: SHA256 + hex.EncodeToString(digestID), MediaType: "application/vnd.oci.image.layer.v1.tar+gzip", Size: len(data)}
					layers = append(layers, layer)
				}
			}
			ch <- 1
		}(file.Name())
		<-ch
	}

	// iterate throigh errors
	for _, e := range errs {
		if e != nil {
			return e
		}
	}

	// update the index.json file
	// put the manifest for the given image
	// set the HTTP method, url, and request body
	ociManifest.Layers = layers
	jsonOCIManifest, err := json.Marshal(ociManifest)
	if err != nil {
		return err
	}

	// set up the request
	req, err := http.NewRequest(http.MethodPut, ss.URL+manifests+ss.Version, bytes.NewBuffer(jsonOCIManifest))
	fmt.Println("DEBUG LMZ : ", string(jsonOCIManifest))
	req.Header.Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
	if ss.Auth {
		req.SetBasicAuth(ba.User, ba.Password)
	}
	if err != nil {
		return err
	}

	// set the request header Content-Type for json
	resp, err := client.Do(req)
	d, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("INFO: " + resp.Status)
	if err != nil {
		return err
	}
	if resp.StatusCode <= 200 || resp.StatusCode >= 300 {
		return fmt.Errorf(string(d))
	}

	return nil
}
