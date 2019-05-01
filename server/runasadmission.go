package server

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RunAsUserAdmission Use this struct to verify the Admission review
type RunAsUserAdmission struct {
}

// HandleAdmission - The main code that captures the AdmissionRequest to be used to verify
func (*RunAsUserAdmission) HandleAdmission(review *v1beta1.AdmissionReview) error {

	// TODO perform check here.

	req := review.Request
	logrus.Debug(req.Kind.Kind)
	switch req.Kind.Kind {
	case "Deployment":
		var deployment appsv1.Deployment
		if err := json.Unmarshal(req.Object.Raw, &deployment); err != nil {
			logrus.Errorf("Could not unmarshal raw object: %v", err)
			review.Response = &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		} else {
			logrus.Debugf("Deployment Name: %s", deployment.ObjectMeta.Name)
			review.Response = &v1beta1.AdmissionResponse{
				Allowed: true,
				Result: &metav1.Status{
					Message: "Deployment captured!",
				},
			}
		}

	default:
		review.Response = &v1beta1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Welcome aboard!",
			},
		}

	}
	return nil
}
