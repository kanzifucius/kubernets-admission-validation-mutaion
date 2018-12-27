package mutatation

import (
	"encoding/json"
	"github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"
	extentions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"strings"
	"vod-ms-kubernetes-admission-webhook/pkg/config"
)

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

var (
	ignoredNamespaces = []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}
)

const (
	AdmissionWebhookAnnotationMutateKey = "za.co.vodacom.admission/mutate"
	AdmissionWebhookAnnotationStatusKey = "za.co.vodacom.admission/status"
	NA                                  = "not_available"
)

func admissionRequired(ignoredList []string, admissionAnnotationKey string, metadata *metav1.ObjectMeta) bool {
	// skip special kubernetes system namespaces
	for _, namespace := range ignoredList {
		if metadata.Namespace == namespace {
			glog.Infof("Skip Mutation for %v for it's in special namespace:%v", metadata.Name, metadata.Namespace)
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

func mutationRequired(ignoredList []string, metadata *metav1.ObjectMeta) bool {

	if metadata == nil {
		glog.Infof("Skip Mutation for  for it's in special as object metadata is nill")
		return false
	}

	required := admissionRequired(ignoredList, AdmissionWebhookAnnotationMutateKey, metadata)
	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	status := annotations[AdmissionWebhookAnnotationStatusKey]

	if strings.ToLower(status) == "mutated" {
		required = false
	}

	glog.Infof("Mutation policy for %v/%v: required:%v", metadata.Namespace, metadata.Name, required)
	return required
}

func Mutate(req *v1beta1.AdmissionRequest, configuration *config.HookConfig) *v1beta1.AdmissionResponse {

	var (
		//availableLabels,
		availableAnnotations            map[string]string
		objectMeta                      *metav1.ObjectMeta
		resourceNamespace, resourceName string
		currentIngress                  *extentions.Ingress
	)

	macro := MacroCommand{}
	glog.Infof("AdmissionReview  mutate for Kind=%v, Namespace=%v Name=%v (%v) UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, resourceName, req.UID, req.Operation, req.UserInfo)

	switch req.Kind.Kind {
	case "Ingress":
		var ingress extentions.Ingress
		if err := json.Unmarshal(req.Object.Raw, &ingress); err != nil {
			glog.Errorf("Could not unmarshal raw object: %v", err)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		resourceName, resourceNamespace, objectMeta = ingress.Name, ingress.Namespace, &ingress.ObjectMeta
		//availableLabels = ingress.Labels
		availableAnnotations = ingress.Annotations
		currentIngress = &ingress

		if !configuration.Mutation.Ingress.Enabled {
			glog.Infof("Skipping Mutation Ingress as disabled in config  for %s/%s due to policy check", resourceNamespace, resourceName)
			return &v1beta1.AdmissionResponse{
				Allowed: true,
			}
		}

		if !mutationRequired(ignoredNamespaces, objectMeta) {
			glog.Infof("Skipping Mutatation  for %s/%s due to policy check", resourceNamespace, resourceName)
			return &v1beta1.AdmissionResponse{
				Allowed: true,
			}
		}

		macro.Append(&IngressMutateCmd{
			Ingress:  currentIngress,
			Mutation: configuration.Mutation,
		})
		annotations := map[string]string{AdmissionWebhookAnnotationStatusKey: "mutated"}
		macro.Append(&AnnotationsMutateCmd{availableAnnotations, annotations})

	}

	patchBytes, err := macro.Execute()
	if err != nil {
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	glog.Infof("AdmissionResponse: patch=%v\n", string(patchBytes))
	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}
