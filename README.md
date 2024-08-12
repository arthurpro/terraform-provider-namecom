# Terraform Provider for [Name.com](https://name.com)

[Name.com API Docs](https://www.name.com/api-docs)

Supported features:

- DNS records
- NS records
- DNSSEC

## Usage

Username and token must be generated from `https://www.name.com/account/settings/api`

```HCL
# Set up the provider
terraform {
  required_providers {
    namecom = {
      source  = "arthurpro/namecom"
      version = "~>1.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~>5.0"
    }
  }
}

provider "aws" {
  region = "us-west-2"
}

resource "aws_route53_zone" "example_com" {
  name = "example.com"
}

# Create the provider with your account details
provider "namecom" {
  username = var.namecom_username
  token    = var.namecom_token
  # test   = true # uncomment to use test API
}
# username and token default to NAMECOM_USER and NAMECOM_TOKEN environment variables

# Example usage for creating DNS records
resource "namecom_record" "bar" {
  zone   = "example.com"
  host   = ""
  type   = "cname"
  answer = "foo.com"
}

resource "namecom_record" "foo" {
  zone   = "example.com"
  host   = "foo"
  type   = "A"
  answer = "1.2.3.4"
}

# Example usage for creating many records per domain
resource "namecom_record" "test" {
  zone = "test.com"
  type = "A"

  for_each = {
    ""   = "1.2.3.4"
    www  = "2.3.4.5"
    www1 = "3.4.5.6"
    www2 = "4.5.6.7"
  }

  host   = each.key
  answer = each.value
}

# Example usage for setting nameservers from a generated hosted_zone
resource "namecom_nameservers" "example_com" {
  zone = "example.com"
  nameservers = [
    "${aws_route53_zone.example_com.name_servers.0}",
    "${aws_route53_zone.example_com.name_servers.1}",
    "${aws_route53_zone.example_com.name_servers.2}",
    "${aws_route53_zone.example_com.name_servers.3}",
  ]
}

# Example usage for using DNSSEC
resource "aws_route53_key_signing_key" "dnssec" {
  name           = data.aws_route53_zone.example_com.name
  hosted_zone_id = data.aws_route53_zone.example_com.id

  key_management_service_arn = aws_kms_key.dnssec.arn

  lifecycle {
    create_before_destroy = true
  }
}

resource "namecom_dnssec" "dnssec" {
  zone        = aws_route53_zone.example_com.name
  key_tag     = aws_route53_key_signing_key.dnssec.key_tag
  algorithm   = aws_route53_key_signing_key.dnssec.signing_algorithm_type
  digest_type = aws_route53_key_signing_key.dnssec.digest_algorithm_type
  digest      = aws_route53_key_signing_key.dnssec.digest_value
}
```

### Importing DNS records

You need to use format `domain/recordID` as last parameter for import command

```bash
# Import single record
terraform import namecom_record.example_record example.com/23456

# Import one of the mentioned records in for_each
terraform import 'namecom_record.example_record["www"]' example.com/23456
```

To get recordId, you need to use Name.com API for domain ListRecords and use ID for appropriate host

```bash
curl -u 'username:token' 'https://api.name.com/v4/domains/example.org/records'
```

### Importing DNSSEC

You need to use format "domain" as last parameter for import command

```bash
# Import single record
terraform import namecom_dnssec.dnssec example.com/digest

# Import one of the mentioned records in for_each
terraform import 'namecom_dnssec.dnssec["example.com"]' example.com
```