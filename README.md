# GenKit Migrate

**Migrate Google GenKit applications between cloud providers**

A CLI tool for migrating existing GenKit applications (built with Google's GenKit framework) from one cloud provider to another.

> **Important**: This migrates applications that use Google's GenKit framework, not the framework itself. Your app continues using Google's GenKit APIs - we just change which cloud provider plugins it uses.

> **GenKit Go 1.0**: This tool is designed to work with GenKit Go 1.0+, the first stable, production-ready release of GenKit for Go. It supports type-safe AI flows, unified model interfaces, tool calling, RAG, multimodal content, and deployment as HTTP endpoints.

[![CI](https://github.com/genkit-migrate/genkit-migrate/actions/workflows/ci.yml/badge.svg)](https://github.com/genkit-migrate/genkit-migrate/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/genkit-migrate/genkit-migrate/branch/main/graph/badge.svg)](https://codecov.io/gh/genkit-migrate/genkit-migrate)
[![Go Report Card](https://goreportcard.com/badge/github.com/genkit-migrate/genkit-migrate)](https://goreportcard.com/report/github.com/genkit-migrate/genkit-migrate)
[![Release](https://img.shields.io/github/v/release/genkit-migrate/genkit-migrate)](https://github.com/genkit-migrate/genkit-migrate/releases)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

## What It Does

Takes a GenKit app like this:
```go
// Original app using Google AI
import "github.com/firebase/genkit/go/plugins/googleai"

googleai.Init(ctx, &googleai.Config{})
resp, _ := genkit.Generate(ctx, &ai.GenerateRequest{
    Model: genkit.Model("googleai/gemini-1.5-pro"),
    // ...
})
```

And transforms it to:
```go  
// Migrated app using AWS Bedrock
import (
    "github.com/firebase/genkit/go/genkit"
    genkitaws "github.com/scttfrdmn/genkit-aws/pkg/genkit-aws"
    "github.com/scttfrdmn/genkit-aws/pkg/bedrock"
)

genkit.Init(ctx, &genkit.Config{
    Plugins: []genkit.Plugin{
        genkitaws.New(&genkitaws.Config{
            Region: "us-east-1",
            Bedrock: &bedrock.Config{
                Models: []string{"anthropic.claude-3-sonnet-20240229-v1:0"},
            },
        }),
    },
})
resp, _ := genkit.Generate(ctx, &ai.GenerateRequest{
    Model: genkit.Model("anthropic.claude-3-sonnet-20240229-v1:0"),
    // ...
})
```

## Migration Process

1. **Analyzes** your GenKit app to find flows, models, and dependencies
2. **Transforms** code to use different cloud provider plugins
3. **Maps models** between providers (e.g., gemini → claude)
4. **Updates** go.mod dependencies  
5. **Generates** new deployment configs (Terraform, Docker, CI/CD)

## Quick Start

### Install

```bash
# Using go install
go install github.com/genkit-migrate/genkit-migrate/cmd/genkit-migrate@latest

# Or download from releases
curl -sSL https://raw.githubusercontent.com/genkit-migrate/genkit-migrate/main/scripts/install.sh | bash
```

### Migrate GCP → AWS

```bash
genkit-migrate migrate --from=gcp --to=aws --source=./my-genkit-app
```

### Preview Changes (Dry Run)

```bash
genkit-migrate migrate --from=gcp --to=aws --source=./my-genkit-app --dry-run
```

## Example Migration

**Before (GCP):**
```
my-genkit-app/
├── go.mod                 # Depends on googleai plugin
├── main.go               # Uses googleai.Init(), googleai/gemini models
└── config.yaml          # GCP-specific configuration
```

**After (AWS):**
```
my-genkit-app_aws/
├── go.mod                # Depends on genkit-aws plugin  
├── main.go              # Uses genkit.Init() with AWS plugin, anthropic/claude models  
├── config.yaml          # AWS-specific configuration
├── terraform/           # AWS deployment configs
│   ├── main.tf
│   └── variables.tf
├── Dockerfile           # Container for AWS Lambda/ECS
└── .github/workflows/   # CI/CD for AWS
    └── deploy.yml
```

## Supported Migrations

| From | To | Status | Models Mapped |
|------|-------|--------|---------------|
| GCP | AWS | ✅ Ready | googleai/gemini → anthropic/claude |
| GCP | Azure | 🚧 Planned | TBD |
| AWS | GCP | 🚧 Planned | TBD |

## Command Reference

### `migrate`
```bash
genkit-migrate migrate [flags]
```

**Flags:**
- `--from`: Source provider (gcp, aws, azure) 
- `--to`: Target provider (aws, gcp, azure)
- `--source, -s`: Source GenKit project path
- `--target, -t`: Target path (default: source_target)
- `--dry-run`: Preview without changes
- `--interactive, -i`: Interactive prompts (default: true)

### `analyze` 
```bash
genkit-migrate analyze --source=./my-genkit-app
```
Analyze a GenKit project without migrating.

## What Gets Transformed

### Code Changes
- **Import statements**: `googleai` → `genkit-aws` packages
- **Plugin initialization**: `googleai.Init()` → `genkit.Init()` with AWS plugin
- **Model references**: `googleai/gemini-1.5-pro` → `anthropic.claude-3-sonnet-20240229-v1:0`
- **Configuration**: AWS region, Bedrock models, CloudWatch monitoring

### Dependencies  
- **go.mod**: Replace provider-specific packages
- **Provider plugins**: Remove old, add new cloud provider plugins
- **Maintain GenKit**: Keep Google's GenKit framework unchanged

### Model Mappings (GCP → AWS)
| GCP Model | AWS Model |
|-----------|-----------|
| googleai/gemini-1.5-flash | anthropic.claude-3-haiku-20240307-v1:0 |
| googleai/gemini-1.5-pro | anthropic.claude-3-sonnet-20240229-v1:0 |
| googleai/gemini-2.0-flash | anthropic.claude-3-5-sonnet-20241022-v2:0 |
| vertexai/gemini-pro | anthropic.claude-3-sonnet-20240229-v1:0 |
| googleai/gemini-1.5-flash-8b | amazon.nova-lite-v1:0 |
| googleai/text-bison | amazon.nova-micro-v1:0 |

### Generated Files
- **Terraform**: AWS infrastructure as code
- **Docker**: Container configuration for AWS services
- **CI/CD**: GitHub Actions for AWS deployment
- **Documentation**: Migration notes and next steps

## Configuration

Create `.genkit-migrate.yaml`:
```yaml
# Default settings
default_source_provider: gcp
default_target_provider: aws
interactive: true

# Provider settings  
providers:
  aws:
    region: us-east-1
    profile: default
  gcp:
    project_id: my-project
    region: us-central1
```

## Contributing

Help expand cloud provider support and model mappings!

### Development Setup
```bash
git clone https://github.com/genkit-migrate/genkit-migrate
cd genkit-migrate
go mod download
./scripts/test.sh
```

### Building
```bash
./scripts/build.sh
```

### Testing
```bash
./scripts/test.sh
```

### Code Quality
This project maintains a **Go Report Card Grade A** through automated quality checks:

```bash
# Run comprehensive quality checks
./scripts/check-reportcard.sh

# Install pre-commit hooks (optional but recommended)
pip install pre-commit
pre-commit install
```

Quality standards:
- ✅ **gofmt**: Code formatting
- ✅ **goimports**: Import organization  
- ✅ **go vet**: Static analysis
- ✅ **golangci-lint**: Comprehensive linting
- ✅ **Test coverage**: ≥80% coverage required
- ✅ **ineffassign**: No ineffective assignments
- ✅ **misspell**: No spelling errors
- ✅ **go mod tidy**: Clean dependencies

## License

Apache License 2.0 - Copyright 2025 Scott Friedman

## AWS Plugin

This migration tool uses the [genkit-aws](https://github.com/scttfrdmn/genkit-aws) plugin for AWS Bedrock integration. The plugin provides:

- **Anthropic Claude Models**: claude-3-5-sonnet, claude-3-sonnet, claude-3-haiku
- **Amazon Nova Models**: nova-pro, nova-lite, nova-micro  
- **CloudWatch Integration**: Metrics and monitoring
- **Bedrock Runtime**: Direct integration with AWS Bedrock service

## Support

- 📖 [Documentation](https://docs.genkit-migrate.com)
- 🐛 [Report Issues](https://github.com/genkit-migrate/genkit-migrate/issues)
- 💬 [Discussions](https://github.com/genkit-migrate/genkit-migrate/discussions)
- 🔧 [GenKit AWS Plugin](https://github.com/scttfrdmn/genkit-aws)