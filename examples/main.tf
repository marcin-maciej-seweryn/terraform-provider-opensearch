terraform {
  required_providers {
    opensearch = {
      version = "1.1.0"
      source  = "localhost/marcin-maciej-seweryn/opensearch"
    }
  }
}

provider "opensearch" {
  endpoint = "http://localhost:9201"
}

resource "opensearch_cluster_settings" "this" {
  persistent {
    auto_create_index = "+index,-temporal*"
  }
}