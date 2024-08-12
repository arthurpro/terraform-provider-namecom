# Terraform Provider for [name.com](https://name.com)

[API Docs](https://www.name.com/api-docs)

Currently only supports DNS records and setting nameservers for a domain zone.

## Usage

Username and token must be generated from
`https://www.name.com/account/settings/api`

```HCL
provider "namecom" {
  token = "0123456789"
  username = "user"
}

// example.com CNAME -> bar.com
resource "namecom_record" "bar" {
  domain_name = "example.com"
  host = ""
  record_type = "cname"
  answer = "bar.com"
  ttl = 300
}

// foo.example.com -> 10.1.2.3
resource "namecom_record" "foo" {
  domain_name = "example.com"
  host = "foo"
  record_type = "A"
  answer = "10.1.2.3"
  ttl = 300
}
```

Many records per domain example

```HCL
resource "namecom_record" "domain-me" {
  domain_name = "domain.me"
  record_type = "A"
  
  for_each = {
    "" = local.t6
    www = local.t8
    www1 = local.t8
    www2 = local.t9
  }

  host = each.key
  answer = each.value
  ttl = 300
}
```

Setting nameservers from a generated hosted_zone

```HCL
provider "aws" {
  region = "us-west-2"
}

provider "namecom" {
  token = "0123456789"
  username = "user"
}

resource "aws_route53_zone" "example_com" {
  name = "example.com"
}

resource "namecom_domain_nameservers" "example_com" {
  domain_name = "example.com"
  nameservers = [
    "${aws_route53_zone.example_com.name_servers.0}",
    "${aws_route53_zone.example_com.name_servers.1}",
    "${aws_route53_zone.example_com.name_servers.2}",
    "${aws_route53_zone.example_com.name_servers.3}",
  ]
}
```

