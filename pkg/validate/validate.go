package validate

import (
	"encoding/json"
	"github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"vod-ms-kubernetes-admission-webhook/pkg/config"
)

var (
	ignoredNamespaces = []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}
)

const (
	admissionWebhookAnnotationValidateKey = "za.co.vodacom.admission/validate"
	admissionWebhookAnnotationStatusKey   = "za.co.vodacom.admission/status"
	NA                                    = "not_available"
)

func init() {

}

func admissionRequired(ignoredList []string, admissionAnnotationKey string, metadata *metav1.ObjectMeta) bool {
	// skip special kubernetes system namespaces
	for _, namespace := range ignoredList {
		if metadata.Namespace == namespace {
			glog.Infof("Skip validation for %v for it's in special namespace:%v", metadata.Name, metadata.Namespace)
			return false
		}
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	var required bool
	switch strings.ToLower(annotations[admissionAnnotationKey]) {
	default:
		required = true
	case "n", "no", "false", "off":
		required = false
	}
	return required
}

func validationRequired(ignoredList []string, metadata *metav1.ObjectMeta) bool {
	if metadata == nil {
		glog.Infof("Skip Validation for  for it's in special as object metadata is nill")
		return false
	}

	required := admissionRequired(ignoredList, admissionWebhookAnnotationValidateKey, metadata)
	glog.Infof("Validation policy for %v/%v: required:%v", metadata.Namespace, metadata.Name, required)
	return required
}

func Validate(req *v1beta1.AdmissionRequest, configuration *config.HookConfig) *v1beta1.AdmissionResponse {
	var (
		availableLabels                 map[string]string
		availableAnnotations            map[string]string
		objectMeta                      *metav1.ObjectMeta
		resourceNamespace, resourceName string
		availablePodSpec                corev1.PodSpec
	)

	glog.Infof("AdmissionReview validate for Kind=%v, Namespace=%v Name=%v (%v) UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, resourceName, req.UID, req.Operation, req.UserInfo)

	switch req.Kind.Kind {
	case "Deployment":
		var deployment appsv1.Deployment
		if err := json.Unmarshal(req.Object.Raw, &deployment); err != nil {
			glog.Errorf("Could not unmarshal raw object: %v", err)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		resourceName, resourceNamespace, objectMeta = deployment.Name, deployment.Namespace, &deployment.ObjectMeta
		availableLabels = deployment.Labels
		availableAnnotations = deployment.Annotations
		availablePodSpec = deployment.Spec.Template.Spec
	case "Service":
		var service corev1.Service
		if err := json.Unmarshal(req.Object.Raw, &service); err != nil {
			glog.Errorf("Could not unmarshal raw object: %v", err)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		resourceName, resourceNamespace, objectMeta = service.Name, service.Namespace, &service.ObjectMeta
		availableLabels = service.Labels
	}

	if !validationRequired(ignoredNamespaces, objectMeta) {
		glog.Infof("Skipping validation for %s/%s due to policy check", resourceNamespace, resourceName)
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}
	validationCmds := MacroCommandValidation{}
	validationCmds.Append(&ValidateAnnotations{availableAnnotations, configuration.Validation.RequiredAnnotations})
	validationCmds.Append(&ValidateLabels{availableLabels, configuration.Validation.RequiredLabels})
	validationCmds.Append(&ValidateImages{availablePodSpec, configuration.Validation.RequiredImageTags})
	validationCmds.Append(&ValidateContainsResourcesDefintions{availablePodSpec})

	allowed := true
	var result *metav1.Status
	validationResult := validationCmds.Execute()
	if validationResult != nil {
		allowed = false
		result = &metav1.Status{
			Message: validationResult.Error(),
		}
	}

	return &v1beta1.AdmissionResponse{
		Allowed: allowed,
		Result:  result,
	}
}
