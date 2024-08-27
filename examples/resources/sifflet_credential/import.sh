# Due to API limitations, Terraform can't detect changes on imported credentials.
# Importing a credential will always generate a diff during the first apply, even
# if the configrued value is the same as the imported one.
terraform import sifflet_credential.example 'credentialname'
