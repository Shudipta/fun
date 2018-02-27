package use_clair

import (
	"fmt"
	"os"
	"net/http"
	"crypto/tls"
	"time"
	"k8s.io/kubernetes/pkg/util/parsers"
)

func keepFootStep(f string, a ...interface{}) {
	s := fmt.Sprintf("%s\n", f)
	fmt.Fprintf(os.Stderr, s, a...)
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
				InsecureSkipVerify: false,
			},
		},
		Timeout: time.Minute,
	}
	repo, tag, _, err := parsers.ParseImageName(imageName)
	repo = repo[10:]
	registry := "https://registry-1.docker.io/v2"
}
