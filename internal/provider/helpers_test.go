package provider

import (
	"fmt"

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
