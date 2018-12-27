package validate

import (
	"errors"
	"github.com/golang/glog"
	"github.com/hashicorp/go-multierror"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

type validationCommand interface {
	Execute() (err error)
}

type MacroCommandValidation struct {
	commands []validationCommand
}

func (self *MacroCommandValidation) Execute() (validationErrors error) {
	validationErrors = nil
	for _, command := range self.commands {
		err := command.Execute()
		if err != nil {
			validationErrors = multierror.Append(validationErrors, err)
		}
	}

	return validationErrors
}

func (self *MacroCommandValidation) Append(command validationCommand) {
	self.commands = append(self.commands, command)
}

func (self *MacroCommandValidation) Undo() {
	if len(self.commands) != 0 {
		self.commands = self.commands[:len(self.commands)-1]
	}
}

func (self *MacroCommandValidation) Clear() {
	self.commands = []validationCommand{}
}

type ValidateLabels struct {
	availableLabels map[string]string
	requiredLabels  []string
}

func (self *ValidateLabels) Execute() error {
	glog.Infof("ValidateLabelsCmd - >  availableLabels : %+v requiredLabels %+v ", self.availableLabels, self.requiredLabels)
	for _, rl := range self.requiredLabels {
		if _, ok := self.availableLabels[rl]; !ok {
			return errors.New(rl + "-> label missing ")

		}
	}
	return nil
}

type ValidateAnnotations struct {
	availbleAnnotatations map[string]string
	requiredAnnotaions    []string
}

func (self *ValidateAnnotations) Execute() error {

	glog.Infof("ValidateAnnotationsCmd  - >  %s %s", self.availbleAnnotatations, self.requiredAnnotaions)

	for _, rl := range self.requiredAnnotaions {
		if _, ok := self.availbleAnnotatations[rl]; !ok {
			return errors.New(rl + "->Annotation missing ")

		}
	}
	return nil
}

type ValidateImages struct {
	podSec            corev1.PodSpec
	requiredImageTags []string
}

func (self *ValidateImages) Execute() error {

	for _, container := range self.podSec.Containers {
		for _, tag := range self.requiredImageTags {

			glog.Infof("Checking container %+v for tag %s ", container.Image, tag)
			if strings.Contains(container.Image, tag) {
				return errors.New("Image does not contain " + tag)
			}
		}

	}
	return nil
}

type ValidateContainsResourcesDefintions struct {
	podSec corev1.PodSpec
}

func (self *ValidateContainsResourcesDefintions) Execute() error {

	for _, container := range self.podSec.Containers {

		resourcesLimits := &container.Resources.Limits
		resourcesRequests := &container.Resources.Requests
		if resourcesLimits == nil && resourcesRequests == nil {
			return errors.New("Missing  resources defintions for container" + container.Name)
		}

	}
	return nil
}
