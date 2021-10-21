# Opensearch Cluster Settings

Opensearch is a community-driven, open source search and analytics suite derived from 
 Apache 2.0 licensed Elasticsearch 7.10.2 & Kibana 7.10.2 which is notably offered as manged service
 from AWS.

Resources offered by this provider allow for configuration of cluster settings,
 which are described in more detail in
 [OpenSearch documentation](https://opensearch.org/docs/latest/opensearch/rest-api/cluster-settings/).

## Basic example

Following example can be used to only allow automatic creation of indexes prefixed with given strings.

```hcl
terraform {
  required_providers {
    opensearch = {
      version = "1.0.0"
      source = "marcin-maciej-seweryn/opensearch"
    }
  }
}

provider "opensearch" {
  endpoint = "http://localhost:9201"
}

resource "opensearch_cluster_settings" "this" {
  persistent {
    auto_create_index = "+my-index-00*,+my-other-index-*"
  }
}
```

## Argument reference

### Required
- `endpoint`: (String) Server's URL

### Optional
- `aws_request_signing`: (Block List, Max: 1)
 Sign requests according to AWS requirements.
 It requires AWS credentials to be accessible through default provider chain
  - `region`: (String) AWS Region
  - `role`: (String, Optional) ARN of the role to assume to perform the operation
