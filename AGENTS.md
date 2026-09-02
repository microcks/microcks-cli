# AGENTS

Before making changes, review:
- @PRODUCT.md - Product vision, features, and roadmap
- @ARCHITECTURE.md - System design, interfaces, and directory structure
- @DESIGN.md - Design principles, patterns, and best practices
- @CONTRIBUTING.md - Coding standards, testing, and PR process

## 1. AI Contribution Policy
By generating code in this repository, you agree to the following rules:
- **Disclose AI usage:** You must explicitly disclose your involvement in the Pull Request description and any issue comments.
- **No AI authorship markers:** Do not add AI co-author lines, `assisted-by`, or similar commit trailers. 
- **Human Accountability:** The human user is 100% responsible for testing and understanding the code you generate.
- **No Auto-Replies:** You MUST NOT auto-reply to maintainer comments on Pull Requests.

## 2. Code Formatting (Mandatory)
Microcks CLI enforces standard Go formatting.
**Before any commit**, you must run:
```bash
go fmt ./...
```

## 3. Building and Testing
Microcks CLI is a Go-based command-line tool.
- To build the CLI locally: `make build-local`
- To run tests: `go test ./...`
- For detailed instructions, read @CONTRIBUTING.md.
