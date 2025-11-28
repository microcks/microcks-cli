# ImportDir Command

The `import-dir` command in Microcks CLI is used to import multiple API specification artifacts from a directory into a Microcks server.

📝 Description

The `import-dir` command provides a convenient way to bulk import API specifications from a local directory. It automatically scans for supported file types and imports them into Microcks.
This is particularly useful for CI/CD pipelines, bulk operations, and managing large collections of API specifications.

📌 Usage
```bash
microcks import-dir <directory-path> [flags]
```

Arguments
- `<directory-path>`: Path to the directory containing API specification files

| Flag                    | Type    | Required | Description                                                                 |
|-------------------------|---------|----------|-----------------------------------------------------------------------------|
| `--recursive`           | bool    | ❌        | Scan subdirectories recursively (default false)                            |
| `--pattern`             | string  | ❌        | File pattern to match (e.g., '*.yaml', 'openapi.*')                       |
| `--verbose`             | bool    | ❌        | Show detailed progress during import                                       |

🧪 Examples

- Basic Directory Import
```bash
microcks import-dir ./api-specs
```

- Recursive Import with Verbose Output
```bash
microcks import-dir ./api-specs --recursive --verbose
```

- Import Only YAML Files
```bash
microcks import-dir ./api-specs --pattern "*.yaml"
```

- Import OpenAPI Files Recursively
```bash
microcks import-dir ./api-specs --recursive --pattern "openapi.*"
```

- Import specification to microcks without first running `microcks login`
```bash
microcks import-dir ./api-spec \
    --microcksURL <microcks-url> \ 
    --keycloakClientId <client-id> \
    --keycloakClientSecret <client-secret> 
```

- Import specification to microcks running without authentication (ie. local uber instance typically)
```bash
microcks import-dir ./api-spec --microcksURL <microcks-url>
```

📋 Supported File Types

The command automatically detects and imports the following file types:
- `.yaml` / `.yml` - OpenAPI, AsyncAPI, and other YAML-based specifications
- `.json` - OpenAPI, AsyncAPI, Postman collections, and other JSON-based specifications  
- `.xml` - SOAP WSDL and other XML-based specifications

🔍 File Type Detection

The command automatically determines which files should be marked as primary artifacts:
- Files containing "openapi" or "swagger" in the filename are marked as primary
- Files containing "postman", "collection", "metadata" or "examples" in the filename are marked as secondary
- All other files default to primary

📊 Output

The command provides:
- Progress reporting showing which files are being imported
- Success/failure status for each file
- Summary of total files found and successfully imported
- Detailed error messages for failed imports

💡 Use Cases

- **CI/CD Pipelines**: Automatically import API specs during build processes
- **Bulk Operations**: Import large collections of API specifications
- **Development Workflows**: Import specs from development directories
- **Testing**: Import test specifications for validation

🔧 Integration

This command integrates with the existing Microcks CLI context system and authentication, making it compatible with all existing workflows and configurations. 