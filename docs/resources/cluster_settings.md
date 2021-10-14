# opensearch_cluster_settings

This resource can be used to set cluster settings.

## Example

The following example disable automatic index creation completely.
```hcl
resource "opensearch_cluster_settings" "this" {
  persistent {
    auto_create_index = "false"
  }
}
```

This setting can also be used to allow creation of indexes that match a pattern.
```hcl
resource "opensearch_cluster_settings" "this" {
  persistent {
    auto_create_index = "index-01,fancy-documents-*"
  }
}
```
## Argument reference

### Required
- `persistent`: (Block List, Min: 1, Max: 1) Persistent cluster setting.
  - `auto_create_index`: (String) Automatically create indexes when a request is received.
  The operation automatically creates the index and applies any matching index templates.
  If no mapping exists, the index operation creates a dynamic mapping.
  Accepted values are: `true`, `false` or comma-separated list of patterns you want to allow,
  or each pattern prefixed with + or - to indicate whether it should be allowed or blocked
