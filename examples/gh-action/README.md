# Go ArcTest GitHub Action Examples

This directory contains example configurations for the Go ArcTest GitHub Action.

## Basic Configuration

The [basic-config.yml](basic-config.yml) file contains a simple configuration with layer definitions and dependency rules.

```yaml
# Basic configuration example
layers:
  - name: Domain
    pattern: "^domain/.*$"
  - name: Application
    pattern: "^application/.*$"
  - name: Infrastructure
    pattern: "^infrastructure/.*$"
  - name: Presentation
    pattern: "^presentation/.*$"

rules:
  - from: Application
    to: Domain
  - from: Infrastructure
    to: Domain
  - from: Infrastructure
    to: Application
  - from: Presentation
    to: Domain
  - from: Presentation
    to: Application
  - from: Presentation
    to: Infrastructure
```

## Advanced Configuration

The [advanced-config.yml](advanced-config.yml) file contains a more advanced configuration with interface implementation rules, parameter type rules, and direct layer dependency rules.

## Layer-Specific Configuration

The [layer-specific-config.yml](layer-specific-config.yml) file contains a configuration with layer-specific rules.

## Hexagonal Architecture Configuration

The [hexagonal-config.yml](hexagonal-config.yml) file contains a configuration for a hexagonal architecture.

## Clean Architecture Configuration

The [clean-architecture-config.yml](clean-architecture-config.yml) file contains a configuration for a clean architecture.

## Usage

To use these configurations with the GitHub Action, create a workflow file in your repository at `.github/workflows/go-arctest.yml`:

```yaml
name: Architecture Test

on: [push, pull_request]

jobs:
  check-architecture:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Run go-arctest
        uses: mstrYoda/go-arctest@v1
        with:
          config: '.github/arctest-config.yml'
```

Then, copy one of the example configurations to `.github/arctest-config.yml` in your repository and customize it to match your project's architecture.
