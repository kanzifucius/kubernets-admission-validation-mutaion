package mutatation

import (
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

func TestMutationDisableds(t *testing.T) {

	assertSet := assert.New(t)

	fakeAdmistionReview, err := factory.AdmissionReview("Admissitionreview.yaml")
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	config, err := factory.ConfigFileName("config_muatationDisabled")
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	response := Mutate(fakeAdmistionReview.Request, config)
	assertSet.NotEmpty(response)
	assertSet.True(response.Allowed)
	assertSet.Empty(response.Patch)

}

func TestMutationEnabled(t *testing.T) {

	assertSet := assert.New(t)

	fakeAdmistionReview, err := factory.AdmissionReview("Admissitionreview.yaml")
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	config, err := factory.ConfigFileName("config_muatationEnabled")
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	response := Mutate(fakeAdmistionReview.Request, config)
	assertSet.NotEmpty(response)
	assertSet.NotEmpty(response.Patch)
	assertSet.True(response.Allowed)

}
