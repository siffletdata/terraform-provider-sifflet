package providertests

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

var staticPrefix = "tf-acc-test-"
var testSessionPrefix = staticPrefix + acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

// AcceptanceTestPrefix returns a prefix used for naming resources created during Terraform acceptance tests. This prefix is common to all test
// sessions (not unique to a single test session).
func AcceptanceTestPrefix() string {
	return staticPrefix
}

// SessionPrefix returns a prefix used for naming resources created during a single test session. This prefix is unique to the test session and is generated
// using known prefix and a random string. It is used to avoid name collisions between different test sessions.
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
	return RandomName() + "s"
}

func RandomEmail() string {
	return "alerts-sandbox+" + RandomName() + "@siffletdata.com"
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
