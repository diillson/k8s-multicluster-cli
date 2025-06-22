package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"net/http"
	"os"
)

func ExecPod(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	restConfig *rest.Config,
	namespace string,
	pod string,
	container string,
	command []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	tty bool,
) error {

	req := clientset.CoreV1().RESTClient().
		Post().
		Namespace(namespace).
		Resource("pods").
		Name(pod).
		SubResource("exec")

	req.Param("container", container)
	req.Param("stdout", "true")
	req.Param("stderr", "true")
	for _, c := range command {
		req.Param("command", c)
	}
	if stdin != nil {
		req.Param("stdin", "true")
	}
	if tty {
		req.Param("tty", "true")
	}

	exec, err := remotecommand.NewSPDYExecutor(restConfig, http.MethodPost, req.URL())
	if err != nil {
		return fmt.Errorf("failed to create SPDY executor: %w", err)
	}

	options := remotecommand.StreamOptions{
		Stdin:             stdin,
		Stdout:            stdout,
		Stderr:            stderr,
		Tty:               tty,
		TerminalSizeQueue: nil, // pode adicionar suporte a resize se quiser
	}

	return exec.StreamWithContext(ctx, options)
}

// Dá suporte a gravação de stdout/stderr em buffer para output clusterizado.
func ExecPodToBuffer(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	restConfig *rest.Config,
	namespace, pod, container string,
	command []string,
	stdinEnabled, tty bool,
) (string, error) {
	var (
		stdout, stderr bytes.Buffer
		in             io.Reader
	)
	if stdinEnabled {
		in = os.Stdin
	}
	err := ExecPod(ctx, clientset, restConfig, namespace, pod, container, command, in, &stdout, &stderr, tty)
	out := stdout.String() + stderr.String()
	return out, err
}
