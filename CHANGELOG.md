# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2025-01-15

### Added
- Initial implementation of genkit-migrate CLI
- Support for migrating GenKit projects from GCP to AWS using [genkit-aws](https://github.com/scttfrdmn/genkit-aws) plugin
- Project analyzer to detect GenKit flows, models, and dependencies
- Code transformer for updating imports and model references
- Project generator for creating migrated project structure
- Template system for AWS deployment configurations
- Interactive CLI with progress indicators and confirmations
- Configuration management system with YAML support
- Comprehensive test coverage across all packages
- Build and release automation with cross-platform support
- GitHub Actions for CI/CD pipeline
- Docker and Terraform template generation for AWS deployment
- Model mapping system for GCP → AWS migrations

### Features
- `migrate` command for full project migration with dry-run support
- `analyze` command for project inspection and reporting
- `version` command with detailed build information
- Interactive and non-interactive modes
- Verbose output for debugging and troubleshooting
- Configuration file support (~/.genkit-migrate.yaml)
- Multi-platform binary builds (Linux, macOS, Windows)

### Supported Migrations
- **GCP to AWS**: Complete migration using genkit-aws plugin
  - Google AI models → Anthropic Claude models (claude-3-5-sonnet, claude-3-sonnet, claude-3-haiku)
  - Google AI models → Amazon Nova models (nova-pro, nova-lite, nova-micro)
  - Plugin initialization transformation
  - Dependency updates for genkit-aws integration

### Infrastructure
- Cross-platform build support with semantic versioning
- Automated testing with Go 1.21, 1.22, and 1.23
- Code coverage reporting with Codecov integration
- Static analysis with golangci-lint and CodeQL security scanning
- Automated releases with GitHub Actions
- Pre-commit hooks for code quality assurance

### Templates
- AWS Terraform configurations for Lambda deployment
- Docker multi-stage builds optimized for AWS
- GitHub Actions workflows for AWS deployment
- CloudWatch monitoring and logging setup
- AWS IAM roles and policies for Bedrock access

### Documentation
- Comprehensive README with usage examples
- Migration guide with before/after code samples
- API documentation for all packages
- Troubleshooting guide for common issues