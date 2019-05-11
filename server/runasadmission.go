package server

import (
	"encoding/json"

	"github.com/balchua/uid-validating-webhook/config"
	"github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RunAsUserAdmission Use this struct to verify the Admission review
type RunAsUserAdmission struct {
	AppConfig config.Configuration
}

func (r *RunAsUserAdmission) isInExcludedNamespace(podNamespace string) (isExcluded bool) {

	for _, excluded := range r.AppConfig.Spec.Excluded {
		logrus.Infof("Excluded namespace: %s", excluded.Name)
		if podNamespace == excluded.Name {
			return true
		}
	}

	return false
}

func (r *RunAsUserAdmission) isSecurityContextValid(pod corev1.Pod) (isExcluded bool) {
	logrus.Info("isSecurityContextValid")
	if pod.Spec.SecurityContext != nil && pod.Spec.SecurityContext.RunAsUser != nil {
		runAsUser := *pod.Spec.SecurityContext.RunAsUser
		for _, includeNamespace := range r.AppConfig.Spec.Included {
			logrus.Infof("Namespace : %s", includeNamespace.Name)
			if pod.ObjectMeta.Namespace == includeNamespace.Name {
				for _, uid := range includeNamespace.Uids {
					logrus.Infof("Uid : %d", uid.Uid)
					if runAsUser == uid.Uid {
						return true
					}
				}
			}

		}
	}

	return false
}

func createAdmissionReviewResponse(allowed bool, message string) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Allowed: allowed,
		Result: &metav1.Status{
			Message: message,
		},
	}
}

func (r *RunAsUserAdmission) validate(pod corev1.Pod) *v1beta1.AdmissionResponse {
	podNamespace := pod.ObjectMeta.Namespace
	if r.isInExcludedNamespace(podNamespace) {
		return createAdmissionReviewResponse(true, "The pod is deployed to an excluded namespace.")
	} else if r.isSecurityContextValid(pod) {
		return createAdmissionReviewResponse(true, "The pod has the right security context.")
	} else if r.AppConfig.Spec.IgnoreOnFailure {
		logrus.Warn("Pod is denied but ignore on failure is set to true, bypassing the validation.")
		return createAdmissionReviewResponse(true, "Overriding the validation due to IgnoreOnFailure is true.")
	}
	return createAdmissionReviewResponse(false, "Pod denied.")
}

// HandleAdmission - The main code that captures the AdmissionRequest to be used to verify
func (r *RunAsUserAdmission) HandleAdmission(review *v1beta1.AdmissionReview) error {

	req := review.Request
	logrus.Info(req.Kind.Kind)
	switch req.Kind.Kind {
	case "Pod":
		var pod corev1.Pod
		if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
			logrus.Errorf("Could not unmarshal raw object: %v", err)
			review.Response = &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		} else {
			review.Response = r.validate(pod)
		}

	}
	return nil
}
