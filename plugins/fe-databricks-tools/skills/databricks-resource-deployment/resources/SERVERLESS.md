# Databricks Serverless Compute Documentation Index

## Overview
This index covers serverless compute functionality for Databricks bundles, jobs, notebooks, and Databricks Connect.

## Getting Started with Serverless

**When to use:**
- Setting up serverless compute for the first time
- Configuring Databricks Connect for local development with serverless
- Need authentication and environment setup

**Documents:**
- `databricks-serverless-connect-tutorial.md` - Complete setup guide for Databricks Connect with serverless (Python 3.12, virtual env, authentication)
- `databricks-serverless-config.md` - Configuration methods (environment variables, profiles, programmatic setup)

## Deploying Serverless Jobs with Bundles

**When to use:**
- Creating jobs that run on serverless compute
- Writing bundle YAML configurations for serverless
- Need serverless job examples

**Documents:**
- `databricks-bundles-serverless-example.md` - **CRITICAL** - Two methods for serverless jobs (omit cluster for notebooks, use environment_key for Python/wheel/dbt)
- `databricks-bundles-job-task-types.md` - Task types compatible with serverless
- `databricks-bundles-library-dependencies.md` - Managing libraries for serverless jobs

## Managing Dependencies

**When to use:**
- Installing Python packages for serverless notebooks or jobs
- Setting up base environments for workspace-wide dependencies
- Understanding dependency caching
- Troubleshooting dependency errors

**Documents:**
- `databricks-serverless-dependencies.md` - Dependency management (Environment panel, workspace files, UC volumes, base environments, caching)
- `databricks-serverless-limitations.md` - Critical PySpark restriction and other dependency limitations

## Understanding Limitations

**When to use:**
- Planning which features can/cannot work with serverless
- Troubleshooting serverless errors
- Understanding API and language restrictions
- Choosing between serverless and cluster-based compute

**Documents:**
- `databricks-serverless-limitations.md` - Comprehensive list of restrictions (language support, APIs, storage, query limits, features, data sources)

## Configuration & Connection

**When to use:**
- Configuring Databricks Connect to use serverless
- Setting up authentication for serverless access
- Managing session timeouts
- Using serverless from local Python environments

**Documents:**
- `databricks-serverless-config.md` - Databricks Connect serverless configuration (Public Preview, Python only, session management)
- `databricks-serverless-connect-tutorial.md` - Full tutorial with OAuth, virtual env, and best practices

## Quick Reference by Task

### "I need to create a serverless job in a bundle"
Start with `databricks-bundles-serverless-example.md` (CRITICAL), then `databricks-bundles-job-task-types.md`

### "I need to install Python packages for serverless"
`databricks-serverless-dependencies.md` - **WARNING**: Do not install PySpark!

### "I need to connect to serverless from my local Python IDE"
`databricks-serverless-connect-tutorial.md`, then `databricks-serverless-config.md`

### "Why is my serverless job/notebook failing?"
`databricks-serverless-limitations.md` - Check for unsupported features, APIs, or configurations

### "How do I configure serverless for different environments?"
`databricks-serverless-config.md` - Environment variables, profiles, programmatic options

### "Can I use [specific feature] with serverless?"
`databricks-serverless-limitations.md` - Comprehensive restrictions list

## Key Serverless Concepts

### Two Job Configuration Approaches
1. **Notebook tasks**: Omit cluster definition entirely
2. **Python/wheel/dbt tasks**: Use `environment_key` with environments section

### Critical Restrictions to Remember
- No PySpark installation
- Python only (no R, no Scala in notebooks)
- Spark Connect APIs only (no RDD)
- No Spark UI or logs
- No DataFrame caching
- Trigger.AvailableNow only for streaming
- 10-minute session timeout for Databricks Connect

### Dependency Best Practices
- Use workspace files or UC volumes for wheels
- Leverage base environments for workspace standardization
- Environment caching improves performance
- Auto-installation for jobs (no manual scheduling needed)
