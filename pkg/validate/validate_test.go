package validate

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"testing"
	"vod-ms-kubernetes-admission-webhook/pkg/fake"
)

var (
	factory *fake.Factory
)

func init() {
	// create a webhook which uses its fake client to seed the sidecar configmap
	factory = fake.NewFactory()
	log.SetOutput(ioutil.Discard)
}

func TestValidationDisabled(t *testing.T) {

	assertSet := assert.New(t)

	fakeAdmistionReview, err := factory.AdmissionReview("admissitionreview_deployment_no_validate.yaml")
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	config, err := factory.Config()
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	response := Validate(fakeAdmistionReview.Request, config)
	assertSet.NotEmpty(response)
	assertSet.True(response.Allowed)
	assertSet.Empty(response.Patch)

}

func TestValidationFail(t *testing.T) {

	assertSet := assert.New(t)

	fakeAdmistionReview, err := factory.AdmissionReview("admissitionreview_deployment_validate_fail.yaml")
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	config, err := factory.Config()
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	response := Validate(fakeAdmistionReview.Request, config)
	assertSet.NotEmpty(response)
	assertSet.False(response.Allowed)
	fmt.Printf("%+v\n", response.Result.Message)

}

func TestValidationOK(t *testing.T) {

	assertSet := assert.New(t)

	fakeAdmistionReview, err := factory.AdmissionReview("admissitionreview_deployment_validate_successs.yaml")
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}
	config, err := factory.Config()
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	response := Validate(fakeAdmistionReview.Request, config)
	assertSet.NotEmpty(response)
	assertSet.True(response.Allowed)

}
