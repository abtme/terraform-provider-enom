# Terraform Provider for eNom

A [Terraform](https://www.terraform.io) provider for managing domain
nameservers via the [eNom reseller API](https://www.enom.com/APICommandCatalog/).

## Using the provider

```hcl
terraform {
  required_providers {
    enom = {
      source = "abtme/enom"
    }
  }
}

provider "enom" {
  # uid/pw can also be set via the ENOM_UID/ENOM_PW environment variables
  uid = "your-reseller-id"
  pw  = "your-reseller-password"
}

resource "enom_domain_nameservers" "example" {
  domain_name = "example.com"
  nameservers = ["ns1.yourdns.com", "ns2.yourdns.com"]
}
```

Full documentation, including the provider schema and resource import
syntax, is published on the
[Terraform Registry](https://registry.terraform.io/providers/abtme/enom/latest/docs).

## Resources

| Resource | Description |
|---|---|
| [`enom_domain_nameservers`](docs/resources/domain_nameservers.md) | Manages the nameservers for an existing domain registration. |

## Developing the provider

Requires [Go](https://go.dev/) (see `go.mod` for the version) and
[Terraform](https://www.terraform.io/downloads) locally.

```shell
go build ./...
```

### Generating docs

Documentation under `docs/` is generated from the provider's schema plus
the example `.tf` files in `examples/`, via
[tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs):

```shell
go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest generate
```

### Releasing

Pushing a `v*` tag triggers `.github/workflows/release.yml`, which builds
and signs release artifacts with [GoReleaser](https://goreleaser.com/) and
publishes a GitHub Release. The Terraform Registry picks up new versions
automatically once connected.

## License

See [LICENSE](LICENSE).
