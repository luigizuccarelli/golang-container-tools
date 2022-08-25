package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/luigizuccarelli/golang-container-tools/pkg/schema"
)

var (
	ba *schema.BasicAuth
)

// OCICopyToDisk - pulls an image from a given registry and saves it to disk in OCI format
func OCICopyToDisk(ss schema.ServiceSchema) error {

	repo, err := name.NewRepository(ss.Image)
	if err != nil {
		return err
	}

	// Fetch credentials based on your docker config file, which is $HOME/.docker/config.json or $DOCKER_CONFIG.
	auth, err := authn.DefaultKeychain.Resolve(repo.Registry)
	if err != nil {
		return err
	}

	// Construct an http.Client that is authorized to pull from gcr.io/google-containers/pause.
	scopes := []string{repo.Scope(transport.PullScope)}
	t, err := transport.New(repo.Registry, auth, http.DefaultTransport, scopes)
	if err != nil {
		return err
	}
	client := &http.Client{Transport: t}

	// create the directory for all the blobs/layers
	err = os.MkdirAll(ss.Path+blobsPath, 0777)
	if err != nil {
		return err
	}

	// setup the GET request
	req, err := http.NewRequest(http.MethodGet, ss.URL+manifests+ss.Version, nil)
	req.Header.Set("Accept", "application/vnd.oci.image.manifest.v1+json")
	if ss.Auth {
		ba, _ = GetBasicAuthCredentials()
		req.SetBasicAuth(ba.User, ba.Password)
	}
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// Assert that we get a 200, otherwise attempt to parse body as a structured error.
	if err := transport.CheckError(resp, http.StatusOK); err != nil {
		return err
	}

	var ms schema.ManifestSchema
	// read and convert to schema
	data, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(data, &ms)
	if err != nil {
		return err
	}

	// for reference we write the original manifest version to disk
	err = ioutil.WriteFile(ss.Path+manifestJSON, data, 0777)
	if err != nil {
		return err
	}

	switch ms.SchemaVersion {
	case 1:
		err = convertAndSaveToOCI(client, ss, ms)
	case 2:
		// for schemaVersion 2 we use OCIImageManifest
		var ocim schema.OCIImageManifest
		err = json.Unmarshal(data, &ocim)
		if err == nil {
			err = saveToOCI(client, ss, ocim)
		}
	default:
		err = fmt.Errorf("version unknown")
	}
	return err
}

func saveToOCI(client *http.Client, ss schema.ServiceSchema, ocim schema.OCIImageManifest) error {
	// download all the blobs/layersa
	err := os.MkdirAll(ss.Path+blobsPath, 0777)
	if err != nil {
		return err
	}
	var errs []error

	ch := make(chan byte, 1)

	for _, x := range ocim.Layers {
		go func(x schema.Layer) {
			req, err := http.NewRequest(http.MethodGet, ss.URL+blobs+x.Digest, nil)
			if ss.Auth {
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

			// read blob fully
			contents, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				errs = append(errs, err)
			}

			// wrtite to file
			err = ioutil.WriteFile(ss.Path+blobsPath+x.Digest[7:], contents, 0777)
			if err != nil {
				errs = append(errs, err)
			}
			fmt.Println("INFO: writing blob ", x.Digest)
			ch <- 1
		}(x)
		<-ch
	}

	// check errors
	for _, e := range errs {
		if e != nil {
			return e
		}
	}

	ij, err := json.Marshal(ocim)
	if err != nil {
		return err
	}

	// write the new index.json file
	err = ioutil.WriteFile(ss.Path+indexJSON, ij, 0777)
	if err != nil {
		return err
	}

	return nil
}

func convertAndSaveToOCI(client *http.Client, ss schema.ServiceSchema, ms schema.ManifestSchema) error {
	var cs = schema.Compatibility{}
	// manipulate the compatibility data
	compatibility := ms.History[0].V1Compatibility
	// clean the compatibility data (will become 'Lables' in the manifest.json)
	clean := strings.Replace(compatibility, "\\", "", -1)

	// unmarshal json to Compatibilty struct
	err := json.Unmarshal([]byte(clean), &cs)
	if err != nil {
		return err
	}

	cs.Rootfs.Type = layers
	var ids []string
	var errs []error

	ch := make(chan byte, 1)

	for _, x := range ms.FsLayers {
		// launch each layer fetch concurrently
		go func(x schema.FsLayer) {
			req, err := http.NewRequest(http.MethodGet, ss.URL+blobs+x.BlobSum, nil)
			if ss.Auth {
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

			// read blob fully
			contents, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				errs = append(errs, err)
			}

			// write to file
			err = ioutil.WriteFile(ss.Path+blobsPath+x.BlobSum[7:], contents, 0777)
			if err != nil {
				errs = append(errs, err)
			}
			fmt.Println("INFO: writing blob ", x.BlobSum)
			ids = append(ids, x.BlobSum)
			ch <- 1
		}(x)
		<-ch
	}

	// check errors
	for _, e := range errs {
		if e != nil {
			return e
		}
	}

	cs.Rootfs.DiffIds = ids

	// add history to manifest
	var history = schema.HistorySchema{}
	var hArray []schema.HistorySchema
	var comp = &schema.Compatibility{}
	var cc = &schema.ContainerConfigSchema{}
	for _, h := range ms.History {
		cleaned := strings.Replace(h.V1Compatibility, "\\", "", -1)
		err := json.Unmarshal([]byte(cleaned), comp)
		if err != nil {
			fmt.Println("ERROR ", err)
		}
		err = json.Unmarshal([]byte(cleaned), cc)
		if err != nil {
			fmt.Println("ERROR ", err)
		}
		if !strings.Contains(cc.ContainerConfig.Cmd[0], "null") {
			history.CreatedBy = cc.ContainerConfig.Cmd[0]
		}
		history.Created = comp.Created
		history.Author = comp.Author
		history.Comment = comp.Comment
		hArray = append(hArray, history)
		history = schema.HistorySchema{}
		comp = &schema.Compatibility{}
		cc = &schema.ContainerConfigSchema{}
	}
	cs.History = hArray

	manifest, err := json.Marshal(cs)
	if err != nil {
		return err
	}

	// write the new manifest
	err = ioutil.WriteFile(ss.Path+blobsPath+cs.ID, manifest, 0777)
	if err != nil {
		return err
	}

	// finally create the index.json file
	var index = schema.Index{SchemaVersion: 2}
	var m = make([]schema.Manifest, 1)
	m[0].MediaType = mediatypeV1
	m[0].Digest = SHA256 + cs.ID
	m[0].Size = len(manifest)
	m[0].Annotations.OrgOpencontainersImageRefName = ss.Image + ":" + ss.Version
	index.Manifests = m

	// marshal index struct to json
	ij, err := json.Marshal(index)
	if err != nil {
		return err
	}

	// write the new index.json file
	err = ioutil.WriteFile(ss.Path+indexJSON, ij, 0777)
	if err != nil {
		return err
	}

	return nil
}
