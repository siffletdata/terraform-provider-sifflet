package providertests

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

var testSessionPrefix = fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum))

// SessionPrefix returns a prefix that can be used to create unique names for test resources created during Terraform acceptance tests.
func SessionPrefix() string {
	return testSessionPrefix
}

// RandomName returns a random name that can be used for test resources created during Terraform acceptance tests.
// It starts with the session prefix, see [SessionPrefix].
func RandomName() string {
	return SessionPrefix() + "-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
}

func RandomCredentialsName() string {
	// Add a trailing "s" to the name because credential names can't end with a digit, as returned by RandomName
	return strings.ReplaceAll(RandomName(), "-", "") + "s"
}

func ProviderConfig() string {
	return providerConfig
}

const (
	// Use environment variables to configure the provider during tests (SIFFLET_HOST and SIFFLET_TOKEN, see README.md).
	providerConfig = `
provider "sifflet" { }
`
)
