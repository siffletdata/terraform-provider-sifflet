package tfhttp

import "testing"

func TestUserAgent(t *testing.T) {
	actual := userAgent("1.9.2", "0.1.1")
	expected := "Terraform/1.9.2 (+https://www.terraform.io) terraform-provider-sifflet/0.1.1"
	if actual != expected {
		t.Fatalf("Expected User-Agent string: %s, got: %s", expected, actual)
	}
}
