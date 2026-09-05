## `microcks import-url` – Import API Artifacts from URL
Imports API specification files (OpenAPI, AsyncAPI, etc.) hosted at a remote URL into the Microcks server.

### Usage
```bash
microcks import-url <specURL1[:main][:secret]>,<specURL2[:main][:secret]> [flags]
```

### Example
```bash
# Import a single artifact (marked as main)
microcks import-url https://example.com/openapi.yaml

# Specify mainArtifact flag for each file
microcks import-url https://example.com/spec1.yaml:true,https://example.com/spec2.yaml:false

# Import specification to microcks without logining to microcks
microcks import-url https://example.com/openapi.yaml \
    --microcksURL <microcks-url> \ 
    --keycloakClientId <client-id> \
    --keycloakClientSecret <client-secret> 

# Download a protected artifact using a stored Microcks secret
microcks import-url https://example.com/openapi.yaml:true:my-secret

# A port in the URL is not mistaken for a suffix
microcks import-url http://localhost:8585/spec.yaml:false
```

### Argument syntax

Each comma-separated entry is a URL optionally followed by a `mainArtifact` flag and
a secret name:

```
<specURL>[:<mainArtifact>[:<secretName>]]
```

| Part | Default | Description |
| ---- | ------- | ----------- |
| `<specURL>` | _(required)_ | `http://` or `https://` URL Microcks downloads the artifact from |
| `<mainArtifact>` | `true` | `true` marks the artifact as primary, `false` as secondary |
| `<secretName>` | _(none)_ | Name of a secret stored in Microcks, used to authenticate the download |

Because a URL contains colons of its own, the CLI resolves the suffixes from the
right: it scans backwards for the first segment that parses as a boolean, treats that
segment as `mainArtifact`, and takes everything after it as the secret name. A secret
name can therefore contain colons. The CLI applies suffix parsing only to arguments
starting with `http://` or `https://`, and passes anything else through unchanged.

### Options

This command has no flags of its own — it takes only the inherited options below.
It also has no `--output` flag, so it produces human-readable output only. See
[JSON output contracts](../json-output.md) for the commands that do emit structured
output.

### Options Inherited from Parent Commands
| Flag                     | Description                                 |
| ------------------------ | ------------------------------------------- |
| `--config`               | Path to Microcks config file                |
| `--microcks-context`     | Name of the Microcks context to use         |
| `--verbose`              | Produce dumps of HTTP exchanges             |
| `--insecure-tls`         | Allow insecure HTTPS connections            |
| `--caCerts`              | Comma-separated paths of CA cert files      |
| `--keycloakClientId`     | Keycloak Realm Service Account ClientId     |
| `--keycloakClientSecret` | Keycloak Realm Service Account ClientSecret |
| `--microcksURL`          | Microcks API URL                            |