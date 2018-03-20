package docker

import (
	appsv1 "k8s.io/api/apps/v1"
	"encoding/json"
	//"log"

	//_ "github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/heroku/docker-registry-client/registry"
	"github.com/tamalsaha/go-oneliners"
	"k8s.io/kubernetes/pkg/util/parsers"
	"fmt"
	//"log"
)

func Annotate(depl *appsv1.Deployment) (*appsv1.Deployment, error) {
	var annotations map[string]string
	conts := depl.Spec.Template.Spec.Containers
	//sec := depl.Spec.Template.Spec.ImagePullSecrets
	for _,cont := range conts {
		img := cont.Image

		lbls, err := GetLabels(img)
		if err != nil {
			return nil, err
		}
		annotations = Merge(annotations, lbls)
	}

	depl.ObjectMeta.SetAnnotations(annotations)

	return depl, nil
}

func Merge(mp1, mp2 map[string]string) map[string]string {
	for key, val := range mp2 {
		mp1[key] = val
	}

	return mp1
}

func GetLabels(img string) (map[string]string, error) {
	var labels map[string]string
	var err error

	url := "https://registry-1.docker.io/"
	username := "" // anonymous
	password := "" // anonymous

	hub, err := registry.New(url, username, password)
	if err != nil {
		return labels, fmt.Errorf("couldn't create registry %v: ", err)
	}
	oneliners.FILE(err)

	// https://github.com/kubernetes/kubernetes/blob/a7a3dcfc527123b6cca15913fbb309172ef2c6e4/pkg/util/parsers/parsers.go#L33
	// https://github.com/kubernetes/kubernetes/blob/master/pkg/util/parsers/parsers_test.go

	repo, tag, digestHash, err := parsers.ParseImageName("shudipta/labels")
	if err != nil {
		return labels, err
	}
	oneliners.FILE(err)
	oneliners.FILE("repo = ", repo)
	oneliners.FILE("tag = ", tag)
	oneliners.FILE("digest = ", digestHash)

	//repo = "tigerworks/labels"
	//tags, err := hub.Tags("shudipta/labels")
	//if err != nil {
	//	return labels, fmt.Errorf("couldn't get the tags: %v", err)
	//}
	//oneliners.FILE(tags, err)

	//m2, err := hub.Manifest("tigerworks/labels", "latest")
	//oneliners.FILE(m2.Name, m2.Tag)
	//d2, err := m2.MarshalJSON()
	//oneliners.FILE(string(d2))

	//tag = "latest"
	repoName := repo[10:]
	fmt.Println("-___", repoName, "_______")
	manifest, err := hub.ManifestV2(repoName, tag)
	//manifest, err := hub.ManifestV2("shudipta/labels", "latest")
	//manifest, err := hub.ManifestV2(repo, tag)
	if err != nil {
		return labels, fmt.Errorf("couldn't get the manifest: %v", err)
	}
	//fmt.Println("\n\n\n_____________manifest is\n", manifest)
	canonical, err := manifest.MarshalJSON()
	fmt.Println("\n\n\n_____________manifest is\n", string(canonical))
	oneliners.FILE("manifest.Config.Digest________", manifest.Config.Digest)
	oneliners.FILE("manifest.Config.Digest________", manifest.Config.Digest.Encoded())
	//oneliners.FILE("manifest.Layers[0].Digest__________", manifest.Layers[0].Digest.Encoded())

	reader, err := hub.DownloadLayer(repoName, manifest.Config.Digest)
	if err != nil {
		return labels, fmt.Errorf("couldn't get encoded imageInspect: %v", err)
		//log.Fatalln(">>>>>>>>>", err)
	}

	var cfg types.ImageInspect
	//fmt.Println(">>>>> reader is:", reader)
	//buf := new(bytes.Buffer)
	//buf.ReadFrom(reader)
	//fmt.Println("\n\n\n\n\n>>>>> reader is:", buf.String(), "\n\n\n")
	//json.Unmarshal([]byte(reader), &cfg)
	err = json.NewDecoder(reader).Decode(&cfg)
	if err != nil {
		return labels, fmt.Errorf("couldn't get decode imageInspect: %v", err)
	}
	//oneliners.FILE(err)
	defer reader.Close()

	oneliners.FILE("labels are:", cfg.Config.Labels)
	return cfg.Config.Labels, nil
}
