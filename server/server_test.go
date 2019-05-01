package server

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"k8s.io/api/admission/v1beta1"

)


func createAdmissionRequest() *v1beta1.AdmissionReview {

	deployment := `
	{
		"apiVersion": "apps/v1",
		"kind": "Deployment",
		"metadata": {
			"labels": {
			  "k8s-app": "nginx"
			},
			"namespace": "nginx",
			"name": "nginx"
		},		
		"spec": {
		  "replicas": 1,
		  "template": {
			"spec": {
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
			},
			"metadata": {
			  "labels": {
				"k8s-app": "nginx"
			  }
			}
		  },
		  "selector": {
			"matchLabels": {
			  "k8s-app": "nginx"
			}
		  }
		}
		
		
	  }
`

	
	admissionRequestNS := v1beta1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind: "AdmissionReview",
		},
		Request: &v1beta1.AdmissionRequest{
			UID: "e911857d-c318-11e8-bbad-025000000001",
			Kind: metav1.GroupVersionKind{
				Kind: "Deployment",
			},
			Operation: "CREATE",
			Object: runtime.RawExtension{
				Raw: []byte(deployment),
			},
		},
	}

	return &admissionRequestNS
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.TextFormatter{})

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

func TestServerCaptureDeployment(t *testing.T) {
	nsc := &RunAsUserAdmission{}
	server := httptest.NewServer(GetAdmissionServerNoSSL(nsc, ":8080").Handler)
	requestString := string(encodeRequest(createAdmissionRequest()))
	myr := strings.NewReader(requestString)
	r, _ := http.Post(server.URL, "application/json", myr)
	review := decodeResponse(r.Body)

	logrus.Debugf("************************* %t", review.Response.Allowed)
	 if review.Response.Allowed != true {
	 	t.Error("Request and response UID don't match")
	 }
}