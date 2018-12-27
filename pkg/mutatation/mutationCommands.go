package mutatation

import (
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/json"
	"strconv"
	"strings"
	"vod-ms-kubernetes-admission-webhook/pkg/config"
)

type command interface {
	Execute() patchOperation
}

type MacroCommand struct {
	commands []command
}

func (self *MacroCommand) Execute() ([]byte, error) {
	var result []patchOperation
	for _, command := range self.commands {
		result = append(result, command.Execute())
	}
	return json.Marshal(result)
}

func (self *MacroCommand) Append(command command) {
	self.commands = append(self.commands, command)
}

func (self *MacroCommand) Undo() {
	if len(self.commands) != 0 {
		self.commands = self.commands[:len(self.commands)-1]
	}
}

func (self *MacroCommand) Clear() {
	self.commands = []command{}
}

type IngressMutateCmd struct {
	Ingress  *extensions.Ingress
	Mutation config.Mutation
}

func (self *IngressMutateCmd) Execute() patchOperation {

	var newRule extensions.IngressRule
	var patch patchOperation

	if self.Mutation.Ingress.Enabled {

		for index, rule := range self.Ingress.Spec.Rules {

			if self.Mutation.Ingress.MutationType == "append" {
				newRule = rule
				newRule.Host = strings.Replace(rule.Host, self.Mutation.Ingress.OldSuffix, self.Mutation.Ingress.NewSuffix, -1)
				path := "/spec/rules/-"
				patch = patchOperation{
					Op:    "add",
					Path:  path,
					Value: newRule,
				}
			} else if self.Mutation.Ingress.MutationType == "replace" {

				values := strings.Replace(rule.Host, "grafana.ms.vodacom.corp", "grafana.teststing.cloud.vodacom.corp", -1)
				path := "/spec/rules/" + strconv.Itoa(index) + "/host"
				patch = patchOperation{
					Op:    "replace",
					Path:  path,
					Value: values,
				}

			}
		}
	}

	return patch
}

type AnnotationsMutateCmd struct {
	target map[string]string
	added  map[string]string
}

func (self *AnnotationsMutateCmd) Execute() patchOperation {
	var patch patchOperation
	for key, value := range self.added {
		if self.target == nil || self.target[key] == "" {
			self.target = map[string]string{}
			patch = patchOperation{
				Op:   "add",
				Path: "/metadata/annotations",
				Value: map[string]string{
					key: value,
				},
			}
		} else {
			patch = patchOperation{
				Op:    "replace",
				Path:  "/metadata/annotations/" + key,
				Value: value,
			}
		}
	}

	return patch
}

type LabelMutateCmd struct {
	target map[string]string
	added  map[string]string
}

func (self *LabelMutateCmd) Execute() patchOperation {
	var patch patchOperation
	values := make(map[string]string)
	for key, value := range self.added {
		if self.target == nil || self.target[key] == "" {
			values[key] = value
		}
	}
	patch = patchOperation{
		Op:    "add",
		Path:  "/metadata/labels",
		Value: values,
	}

	return patch
}
