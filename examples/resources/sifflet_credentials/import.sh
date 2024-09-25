# Due to API limitations, Terraform can't detect changes on imported credentials.
# Importing credentials will always generate a diff during the first apply, even
# if the configured value is the same as the imported one.
terraform import sifflet_credentials.example 'credentialname'
