terraform {
  required_providers {
    enom = {
      source = "abtme/enom"
    }
  }
}

# uid/pw can also be set via the ENOM_UID/ENOM_PW environment variables
# instead of hardcoding them here.
provider "enom" {
  uid = "your-reseller-id"
  pw  = "your-reseller-password"
}
