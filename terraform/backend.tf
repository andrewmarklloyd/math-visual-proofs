terraform {
  backend "s3" {
    endpoint                    = "sfo3.digitaloceanspaces.com"
    key                         = "math-visual-proofs/terraform.tfstate"
    bucket                      = "math-visual-proofs"
    region                      = "us-west-1"
    skip_credentials_validation = true
    skip_metadata_api_check     = true
  }
}
