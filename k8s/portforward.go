package k8s

import (
	"context"
	"fmt"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func PortForward(ctx context.Context, restConfig *rest.Config, namespace, pod, address string, ports []string) error {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", namespace, pod)
	hostIP := strings.TrimLeft(restConfig.Host, "htps:/")
	url := &url.URL{
		Scheme: "https", // ou de acordo com o cluster
		Path:   path,
		Host:   hostIP,
	}

	transport, upgrader, err := spdy.RoundTripperFor(restConfig)
	if err != nil {
		return err
	}

	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{})
	go func() {
		<-ctx.Done()
		close(stopChan)
	}()

	pf, err := portforward.New(
		spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url),
		ports,
		stopChan,
		readyChan,
		os.Stdout,
		os.Stderr,
	)
	if err != nil {
		return err
	}
	return pf.ForwardPorts()
}
