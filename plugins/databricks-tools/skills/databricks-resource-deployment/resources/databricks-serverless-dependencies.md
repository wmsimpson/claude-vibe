# Serverless Compute Dependencies

## Dependency Management

### Environment Side Panel
Centralized control for:
- Dependencies configuration
- Budget policies
- Memory settings
- Environment versions

Applies to serverless compute only

### Supported Path Formats
- **Workspace files**: `/Workspace/...`
- **Unity Catalog volumes**: `/Volumes/<catalog>/<schema>/<volume>/<path>.whl`
- **Standard formats**: Any requirements.txt-valid format

### Automatic Installation
Dependencies auto-install when serverless jobs execute (no manual installation during scheduling)

## Critical Limitation
**DO NOT install PySpark or dependencies that install PySpark**
- Stops session and causes errors
- Must remove and reset environment if conflicts occur

## Advanced Features

### Environment Caching
- Auto-preserves virtual environment content
- Reduces reinstallation time
- Improves job performance with shared dependencies

### Base Environments
- YAML files for workspace-wide standardization
- Admins can configure private/authenticated package repositories
- Access internal Python repos without explicit index config

## Non-Notebook Task Dependencies
Python scripts, wheels, JAR files, and dbt tasks:
- Inherit libraries from serverless environment version
- Add dependencies via **Environment and Libraries** dialog during task creation
