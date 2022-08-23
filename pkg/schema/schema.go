package schema

// OCIImageManifest - oci image manifest
type OCIImageManifest struct {
	SchemaVersion int `json:"schemaVersion"`
	Config        struct {
		MediaType string `json:"mediaType"`
		Digest    string `json:"digest"`
		Size      int    `json:"size"`
	} `json:"config"`
	Layers      []Layer `json:"layers"`
	Annotations struct {
		OrgOpencontainersImageBaseDigest string `json:"org.opencontainers.image.base.digest"`
		OrgOpencontainersImageBaseName   string `json:"org.opencontainers.image.base.name"`
	} `json:"annotations"`
}

// ManifestSchema - manifest from registry
type ManifestSchema struct {
	Tag           string `json:"tag"`
	Name          string `json:"name"`
	Architecture  string `json:"architecture"`
	SchemaVersion int    `json:"schemaVersion"`
	History       []struct {
		V1Compatibility string `json:"v1Compatibility"`
	} `json:"history"`
	FsLayers []FsLayer `json:"fsLayers"`
}

// Layer schemaVersion 2
type Layer struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int    `json:"size"`
}

// FsLayer - schemaVersion 1 - blobsum for each layer
type FsLayer struct {
	BlobSum string `json:"blobSum"`
}

// Index - used to crete index.json in oci layout
type Index struct {
	SchemaVersion int        `json:"schemaVersion"`
	Manifests     []Manifest `json:"manifests"`
}

// Manifest - used in newly create oci manifest
type Manifest struct {
	MediaType   string `json:"mediaType"`
	Digest      string `json:"digest"`
	Size        int    `json:"size"`
	Annotations struct {
		OrgOpencontainersImageRefName string `json:"org.opencontainers.image.ref.name"`
	} `json:"annotations"`
}

// Compatibility - taken from History[0].V1Compatibility in ManifestSchema
type Compatibility struct {
	Created      string `json:"created"`
	Architecture string `json:"architecture"`
	//Container     string    `json:"container"`
	//DockerVersion string    `json:"docker_version"`
	Os     string `json:"os"`
	Config struct {
		User       string   `json:"User"`
		Env        []string `json:"Env"`
		Entrypoint []string `json:"Entrypoint"`
		WorkingDir string   `json:"WorkingDir"`
		Labels     struct {
			Architecture              string `json:"architecture"`
			BuildDate                 string `json:"build-date"`
			ComRedhatBuildHost        string `json:"com.redhat.build-host"`
			ComRedhatComponent        string `json:"com.redhat.component"`
			ComRedhatLicenseTerms     string `json:"com.redhat.license_terms"`
			Description               string `json:"description"`
			DistributionScope         string `json:"distribution-scope"`
			IoK8SDescription          string `json:"io.k8s.description"`
			IoK8SDisplayName          string `json:"io.k8s.display-name"`
			IoOpenshiftExposeServices string `json:"io.openshift.expose-services"`
			Maintainer                string `json:"maintainer"`
			Name                      string `json:"name"`
			Release                   string `json:"release"`
			Summary                   string `json:"summary"`
			URL                       string `json:"url"`
			VcsRef                    string `json:"vcs-ref"`
			VcsType                   string `json:"vcs-type"`
			Vendor                    string `json:"vendor"`
			Version                   string `json:"version"`
		} `json:"Labels"`
	} `json:"config"`
	Rootfs struct {
		Type    string   `json:"type"`
		DiffIds []string `json:"diff_ids"`
	} `json:"rootfs"`
	History []HistorySchema `json:"history"`
	ID      string          `json:"id"`
	Comment string          `json:"comment,omitempty"`
	Author  string          `json:"author,omitempty"`
}

// ContainerConfigSchema used to extract ContainerConfig.Cmd
type ContainerConfigSchema struct {
	ContainerConfig struct {
		Cmd []string `json:"Cmd"`
	} `json:"container_config"`
}

// HistorySchema used in Manifest
type HistorySchema struct {
	Created   string `json:"created"`
	CreatedBy string `json:"created_by,omitempty"`
	Author    string `json:"author,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

// ServiceSchema - holds all relevent data
type ServiceSchema struct {
	Name      string
	User      string
	Component string
	Version   string
	Path      string
	Image     string
	URL       string
	TLS       bool
	Auth      bool
}

// BasicAuth struct
type BasicAuth struct {
	User     string
	Password string
}
