package k8s

import (
	"bytes"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/util/retry"
)

func ApplyManifest(clientset *kubernetes.Clientset, dynamicClient dynamic.Interface, clusterName string, data []byte) error {
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)
	for {
		var rawObj map[string]interface{}
		if err := decoder.Decode(&rawObj); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("failed to decode YAML: %v", err)
		}

		unstructuredObj := &unstructured.Unstructured{Object: rawObj}
		gvk := unstructuredObj.GroupVersionKind()

		// Cria um mapeador dinâmico
		gr, err := restmapper.GetAPIGroupResources(clientset.Discovery())
		if err != nil {
			return fmt.Errorf("failed to get API group resources: %v", err)
		}
		mapper := restmapper.NewDiscoveryRESTMapper(gr)

		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return fmt.Errorf("failed to get REST mapping: %v", err)
		}

		resourceClient := dynamicClient.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace())

		_, err = resourceClient.Create(context.TODO(), unstructuredObj, metav1.CreateOptions{})
		if errors.IsAlreadyExists(err) {
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				_, updateErr := resourceClient.Update(context.TODO(), unstructuredObj, metav1.UpdateOptions{})
				return updateErr
			})
			if retryErr != nil {
				logrus.Errorf("Failed to update resource in cluster %s: %v", clusterName, retryErr)
			} else {
				logrus.Errorf("Updated resource in cluster %s", clusterName)
			}
		} else if err != nil {
			logrus.Errorf("Failed to create resource in cluster %s: %v", clusterName, err)
		} else {
			logrus.Errorf("Created resource in cluster %s", clusterName)
		}
	}

	return nil
}
