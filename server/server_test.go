package server

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"gotest.tools/assert"
	"k8s.io/api/admission/v1beta1"

	"github.com/balchua/uid-validating-webhook/config"

)

type Config struct {
	ListenOn  string `default:"0.0.0.0:9443"`
	TlsCert   string `default:"/etc/webhook/certs/cert.pem"`
	TlsKey    string `default:"/etc/webhook/certs/key.pem"`
	Debug     bool   `default:"true"`
	PodPolicy string `default:"/etc/webhook/"`
}

var (

	podInExcludedNamespace string = `
		{
			"apiVersion": "core/v1",
			"kind": "Pod",
			"metadata": {
				"labels": {
					"k8s-app": "nginx"
				},
				"namespace": "kube-system",
				"name": "nginx"
			},		
			"spec": 
			{
				"securityContext" : {
					"runAsUser":1000,
					"fsGroups": 1000
				},
				"containers": [
					{
						"image": "nginx",
						"name": "nginx",
						"resources": {
							"requests": {
								"cpu": "80m",
								"memory": "140Mi"
							},
							"limits": {
								"cpu": "80m",
								"memory": "140Mi"
							}
						}
					}
				]
			}
		}`

	podWithSecurityContext string = `
		{
			"apiVersion": "core/v1",
			"kind": "Pod",
			"metadata": {
				"labels": {
					"k8s-app": "nginx"
				},
				"namespace": "dev",
				"name": "nginx"
			},		
			"spec": 
			{
				"securityContext" : {
					"runAsUser": 1000,
					"fsGroup" : 1000
				},
				"containers": [
					{
						"image": "nginx",
						"name": "nginx",
						"resources": {
							"requests": {
								"cpu": "80m",
								"memory": "140Mi"
							},
							"limits": {
								"cpu": "80m",
								"memory": "140Mi"
							}
						}
					}
				]
			}
		}`
		
		podWithoutSecurityContextOnExcludedNamespace string = `
		{
			"apiVersion": "core/v1",
			"kind": "Pod",
			"metadata": {
				"labels": {
					"k8s-app": "nginx"
				},
				"namespace": "kube-system",
				"name": "nginx"
			},		
			"spec": 
			{
				"containers": [
					{
						"image": "nginx",
						"name": "nginx",
						"resources": {
							"requests": {
								"cpu": "80m",
								"memory": "140Mi"
							},
							"limits": {
								"cpu": "80m",
								"memory": "140Mi"
							}
						}
					}
				]
			}
		}`		

		podWithoutSecurityContextOnIncludedNamespace string = `
		{
			"apiVersion": "core/v1",
			"kind": "Pod",
			"metadata": {
				"labels": {
					"k8s-app": "nginx"
				},
				"namespace": "dev",
				"name": "nginx"
			},		
			"spec": 
			{
				"securityContext" : {},
				"containers": [
					{
						"image": "nginx",
						"name": "nginx",
						"resources": {
							"requests": {
								"cpu": "80m",
								"memory": "140Mi"
							},
							"limits": {
								"cpu": "80m",
								"memory": "140Mi"
							}
						}
					}
				]
			}
		}`				
)


func createAdmissionRequest(pod string) *v1beta1.AdmissionReview {
	
	admissionRequestNS := v1beta1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind: "AdmissionReview",
		},
		Request: &v1beta1.AdmissionRequest{
			UID: "e911857d-c318-11e8-bbad-025000000001",
			Kind: metav1.GroupVersionKind{
				Kind: "Pod",
			},
			Operation: "CREATE",
			Object: runtime.RawExtension{
				Raw: []byte(pod),
			},
		},
	}

	return &admissionRequestNS
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.TextFormatter{})

	// Output to stdout instead of the default stderr
    // Can be any io.Writer, see below for File example
    logrus.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	logrus.SetLevel(logrus.DebugLevel)
}

func decodeResponse(body io.ReadCloser) *v1beta1.AdmissionReview {
	response, _ := ioutil.ReadAll(body)
	review := &v1beta1.AdmissionReview{}
	codecs.UniversalDeserializer().Decode(response, nil, review)
	return review
}

func encodeRequest(review *v1beta1.AdmissionReview) []byte {
	ret, err := json.Marshal(review)
	if err != nil {
		logrus.Errorln(err)
	}
	return ret
}

func TestPodWithSecurityContext(t *testing.T) {
	appConfig := config.GetAppConfig("/home/thor/workspace/go-work/src/github.com/balchua/uid-validating-webhook/hack")
	appConfig.Spec.IgnoreOnFailure = false
	nsc := &RunAsUserAdmission{
		AppConfig: appConfig,
	}
	server := httptest.NewServer(GetAdmissionServerNoSSL(nsc, ":8080").Handler)
	requestString := string(encodeRequest(createAdmissionRequest(podWithSecurityContext)))
	myr := strings.NewReader(requestString)
	r, _ := http.Post(server.URL, "application/json", myr)
	review := decodeResponse(r.Body)

	logrus.Debugf("************************* %t", review.Response.Allowed)
	assert.Equal(t, review.Response.Allowed, true)
	
}


func TestPodInExcludedNamespace(t *testing.T) {
	appConfig := config.GetAppConfig("/home/thor/workspace/go-work/src/github.com/balchua/uid-validating-webhook/hack")
	nsc := &RunAsUserAdmission{
		AppConfig: appConfig,
	}
	server := httptest.NewServer(GetAdmissionServerNoSSL(nsc, ":8080").Handler)
	requestString := string(encodeRequest(createAdmissionRequest(podInExcludedNamespace)))
	myr := strings.NewReader(requestString)
	r, _ := http.Post(server.URL, "application/json", myr)
	review := decodeResponse(r.Body)

	logrus.Debugf("************************* %t", review.Response.Allowed)
	assert.Equal(t, review.Response.Allowed, true)
	
}

func TestPodWithoutSecurityContextButIgnoredOnFailure(t *testing.T) {
	appConfig := config.GetAppConfig("/home/thor/workspace/go-work/src/github.com/balchua/uid-validating-webhook/hack")
	nsc := &RunAsUserAdmission{
		AppConfig: appConfig,
	}
	server := httptest.NewServer(GetAdmissionServerNoSSL(nsc, ":8080").Handler)
	requestString := string(encodeRequest(createAdmissionRequest(podWithoutSecurityContextOnIncludedNamespace)))
	myr := strings.NewReader(requestString)
	r, _ := http.Post(server.URL, "application/json", myr)
	review := decodeResponse(r.Body)

	logrus.Debugf("************************* %t", review.Response.Allowed)
	assert.Equal(t, review.Response.Allowed, true)
	
}

func TestPodWithoutSecurityContext(t *testing.T) {
	appConfig := config.GetAppConfig("/home/thor/workspace/go-work/src/github.com/balchua/uid-validating-webhook/hack")
	appConfig.Spec.IgnoreOnFailure = false
	nsc := &RunAsUserAdmission{
		AppConfig: appConfig,
	}
	server := httptest.NewServer(GetAdmissionServerNoSSL(nsc, ":8080").Handler)
	requestString := string(encodeRequest(createAdmissionRequest(podWithoutSecurityContextOnIncludedNamespace)))
	myr := strings.NewReader(requestString)
	r, _ := http.Post(server.URL, "application/json", myr)
	review := decodeResponse(r.Body)

	logrus.Debugf("************************* %t", review.Response.Allowed)
	assert.Equal(t, review.Response.Allowed, false)
	
}