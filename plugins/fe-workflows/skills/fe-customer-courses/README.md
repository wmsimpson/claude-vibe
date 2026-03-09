# Customer Training Courses

Generate customer-specific Databricks training course recommendations based on Salesforce use cases, Google Drive research, and the official course catalog. Creates a customer-facing Google Doc with personalized course suggestions organized by skill level.

## How to Invoke

### Slash Command

```
/fe-customer-courses
```

### Example Prompts

```
"Create training recommendations for Meta"
"Generate a course plan for Acme Corp focusing on data engineering"
"Make training recommendations for our ML use cases at BigCo"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Salesforce Auth | Run `/salesforce-authentication` for use case lookups |
| Google Auth | Run `/google-auth` for reading the course catalog and creating output doc |

## What This Skill Does

1. Queries Salesforce for the customer's active use cases to identify technology focus areas
2. Identifies target personas (Data Engineers, Data Scientists, Analysts, etc.) and confirms with the user
3. Searches Google Drive for recent account documents to understand engagement context
4. References the official Databricks course catalog, filtering to customer-available courses only
5. Matches courses to personas and use case maturity levels (Foundation, Intermediate, Advanced)
6. Creates a professional, customer-facing Google Doc with training path tables and course links

## Related Skills

- `/salesforce-authentication` - Required for Salesforce use case queries
- `/google-auth` - Required for catalog access and doc creation
- `/google-docs` - Used for reading the course catalog spreadsheet
