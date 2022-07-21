package k8s

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/shundezhang/oidc-config/pkg/logger"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"

	certmanager "github.com/cert-manager/cert-manager/pkg/api"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultApiUriPath  = "/api"
	defaultApisUriPath = "/apis"
)

var (
	defaultYamlDelimiter = []byte("---")
)

func getKubernetesConfigInCluster() (*rest.Config, error) {
	var config *rest.Config
	config, err := rest.InClusterConfig()
	if err != nil {
		return getKubernetesLocalConfig()
	}
	return config, nil
}

func getKubernetesLocalConfig() (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	clientCfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	return clientCfg.ClientConfig()
}

func GetKubernetesConfig(kubePath string) (*rest.Config, error) {
	var (
		config *rest.Config
		err    error
	)

	// if kubeconfig path is not provided, try to auto detect
	if kubePath == "" {
		config, err = getKubernetesConfigInCluster()
		if err != nil {
			return nil, err
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubePath)
		if err != nil {
			return nil, err
		}
	}

	return config, err
}

func GetKubernetesClient(kubePath string) (kubernetes.Interface, error) {
	config, err := GetKubernetesConfig(kubePath)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)

	if err != nil {
		return nil, err
	}
	return client, nil
}

func Apply(kubePath string, yaml []byte) error {
	log := logger.NewLogger()
	k, err := GetKubernetesClient(kubePath)
	if err != nil {
		return err
	}
	objs, err := getObjects(yaml)
	if err != nil {
		return err
	}
	log.Info("Got %d objects.", len(objs))
	// Create a REST mapper that tracks information about the available resources in the cluster.
	groupResources, err := restmapper.GetAPIGroupResources(k.Discovery())
	if err != nil {
		return err
	}
	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	for i := range objs {
		// Get some metadata needed to make the REST request.
		gvk := objs[i].GetObjectKind().GroupVersionKind()
		gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
		mapping, err := mapper.RESTMapping(gk, gvk.Version)
		if err != nil {
			return err
		}
		namespace, name, err := retrievesMetaFromObject(objs[i])
		log.Info("Applying %s in %s...", name, namespace)
		if err != nil {
			return err
		}
		cli, err := getResourceClient(kubePath, mapping.GroupVersionKind.GroupVersion())
		if err != nil {
			return err
		}
		helper := resource.NewHelper(cli, mapping)
		err = applyObject(helper, namespace, name, objs[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func applyObject(helper *resource.Helper, namespace, name string, obj runtime.Object) error {
	if _, err := helper.Get(namespace, name); err != nil {
		_, err = helper.Create(namespace, false, obj)
		if err != nil {
			return err
		}
	} else {
		_, err = helper.Replace(namespace, name, true, obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func getResourceClient(kubePath string, gv schema.GroupVersion) (rest.Interface, error) {
	cfg, err := GetKubernetesConfig(kubePath)
	if err != nil {
		return nil, err
	}
	cfg.ContentConfig = resource.UnstructuredPlusDefaultContentConfig()
	cfg.GroupVersion = &gv
	if len(gv.Group) == 0 {
		cfg.APIPath = defaultApiUriPath
	} else {
		cfg.APIPath = defaultApisUriPath
	}
	return rest.RESTClientFor(cfg)
}

func GetYAML(yamlURL string) ([]byte, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", yamlURL, nil)
	if err != nil {
		return nil, fmt.Errorf("Got error %s", err.Error())
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Got error %s", err.Error())
	}
	data, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	return data, nil
}

func getObjects(content []byte) ([]runtime.Object, error) {
	objs := make([]runtime.Object, 0)

	delimited := bytes.Split(content, defaultYamlDelimiter)
	for _, del := range delimited {
		if len(del) == 0 {
			continue
		}
		apiextensionsv1.AddToScheme(scheme.Scheme)
		apiextensionsv1beta1.AddToScheme(scheme.Scheme)
		certmanager.AddToScheme(scheme.Scheme)
		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode(del, nil, nil)
		if err != nil {
			return nil, err
		}
		objs = append(objs, obj)
	}
	return objs, nil
}

func retrievesMetaFromObject(obj runtime.Object) (namespace, name string, err error) {
	name, err = meta.NewAccessor().Name(obj)
	if err != nil {
		return
	}
	namespace, err = meta.NewAccessor().Namespace(obj)
	if err != nil {
		return
	}
	return
}
