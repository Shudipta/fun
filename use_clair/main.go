package use_clair

import (
	"fmt"
	"os"
	//"net/http"
	//"crypto/tls"
	//"time"
	"k8s.io/kubernetes/pkg/util/parsers"
	"log"
	reg "github.com/heroku/docker-registry-client/registry"
	"bytes"
	"encoding/json"
	"net/http"
	"time"
	"fun/use_clair/clair"
	"strings"
	"crypto/tls"
)

type config struct {
	MediaType string
	Size int
	Digest    string
}

type layer struct {
	MediaType string
	Size int
	Digest    string
}

type Canonical struct {
	SchemaVersion int
	MediaType string
	Config        config
	Layers        []layer
}

func keepFootStep(f string, a ...interface{}) {
	s := fmt.Sprintf("%s\n", f)
	fmt.Fprintf(os.Stderr, s, a...)
}

func RequestBearerToken(repo, user, pass string) *http.Request {
	url := "https://auth.docker.io/token?service=registry.docker.io&scope=repository:" + repo +":pull&account=" + user
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Can't create a request: %v\n", err)
	}
	if user != "" {
		req.SetBasicAuth(user, pass)
	}

	return req
}

func GetBearerToken(resp *http.Response, err error) string {
	if err != nil {
		log.Fatalf("Can't get response for Bearer Token request: %v\n", err)
	}

	defer resp.Body.Close()

	var token struct {
		Token string
	}

	if err = json.NewDecoder(resp.Body).Decode(&token); err != nil {
		log.Fatal("Token response decode error:", err)
	}
	return fmt.Sprintf("Bearer %s", token.Token)
}

func RequestSendingLayer(l *clair.LayerType) *http.Request {
	var layerApi struct{
		Layer *clair.LayerType
	}
	layerApi.Layer = l
	reqBody, err := json.Marshal(layerApi)
	if err != nil {
		log.Fatalf("[]byte converting err:", err)
	}
	url := "http://192.168.99.100:30060/v1/layers"

	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		log.Fatalln("request error:", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return req
}

func RequestVulnerabilities(hashNameOfImage string) *http.Request {
	url := "http://192.168.99.100:30060/v1/layers/" + hashNameOfImage + "?features&vulnerabilities"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln("error in creating request for getting vulnerabilities:", err)
	}

	return req
}

func GetVulnerabilities(resp *http.Response, err error) []*clair.Vulnerability {
	if err != nil {
		log.Fatalf("Can't get response for vulnerabilities request: %v\n", err)
	}

	defer resp.Body.Close()

	var layerApi struct{
		Layer *clair.LayerType
	}
	err := json.NewDecoder(resp.Body).Decode(&layerApi)
	if err != nil {
		log.Fatalln("error in converting into structure:", err)
	}
	var vuls []*clair.Vulnerability
	for _, feature := range layerApi.Layer.Features {
		for _, vul := range feature.Vulnerabilities {
			vuls = append(vuls, &vul)
		}
	}

	return vuls
}

func main() {
	clairAddr := "http://192.168.99.100:30060"
	clairOutput := "Low"
	imageName := "shudipta/labels"
	user := "shudipta"
	pass := "pi-shudipta"

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false, //true
			},
		},
		Timeout: time.Minute,
	}
	repo, tag, _, err := parsers.ParseImageName(imageName)
	if err != nil {
		log.Fatal(err)
	}
	repo = repo[10:]
	registry := "https://registry-1.docker.io/v2"

	hub, err := reg.New("https://registry-1.docker.io/", user, pass)
	if err != nil {
		log.Fatalf("couldn't create registry %v: ", err)
	}

	manifest, err := hub.ManifestV2(repo, tag)
	if err != nil {
		log.Fatalf("couldn't get the manifest: %v", err)
	}
	canonical, err := manifest.MarshalJSON()
	if err != nil {
		log.Fatalf("couldn't get the manifest.canonical: %v", err)
	}
	can := bytes.NewReader(canonical)

	var img Canonical
	if err := json.NewDecoder(can).Decode(&img); err != nil {
		log.Fatalf("Image decode error")
	}

	var ls []layer
	for _, l := range img.Layers {
		if l.Digest == "" {
			continue
		}
		ls = append(ls, l)
	}
	digest := img.Config.Digest
	schemaVersion := img.SchemaVersion

	if len(ls) == 0 {
		keepFootStep("Can't pull fsLayers")
	} else {
		fmt.Println("Analysing", len(ls), "layers")
	}

	clairClient := http.Client{
		Timeout: time.Minute,
	}
	lsLen := len(ls)
	for i := 0; i < lsLen; i++ {
		var parent string
		if i > 0 {
			parent = strings.Replace(digest, "sha256:", "", 1) +
				strings.Replace(ls[i - 1].Digest, "sha256:", "", 1)
		}
		l := &clair.LayerType{
			Name: strings.Replace(digest, "sha256:", "", 1) +
				strings.Replace(ls[i].Digest, "sha256:", "", 1),
			Path: fmt.Sprintf("%s/%s/%s/%s", registry, repo, "blobs", ls[i].Digest),
			ParentName: parent,
			Format: "Docker",
			Headers: clair.HeadersType{
				Authorization: GetBearerToken(client.Do(RequestBearerToken(repo, user, pass))),
			},
		}

		_, err := clairClient.Do(RequestSendingLayer(l))
		if err != nil {
			log.Fatalf("can't send layer: %v", err)
		}
	}

	vuls := GetVulnerabilities(clairClient.Do(RequestVulnerabilities(
		strings.Replace(digest, "sha256:", "", 1) +
			strings.Replace(ls[lsLen - 1].Digest, "sha256:", "", 1),
	)))
}
