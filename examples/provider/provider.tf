provider "sifflet" {
  host = "https://tenant.siffletdata.com/api"
  // Sifflet API token. If not set, the token will be read from the SIFFLET_TOKEN environment variable.
  // (it's recommended to use environment variables to avoid storing sensitive information in your
  // configuration)
  token = "123azert"
}
