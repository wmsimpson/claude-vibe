---
name: github-workflow
description: Manage GitHub workflows for personal app projects. Create branches, PRs, issues, releases, and GitHub Actions CI/CD. Use for "create a PR", "set up CI", "create a release", "manage issues", "branch strategy", "GitHub Actions". Works with any personal or team GitHub repo.
user-invocable: true
---

# GitHub Workflow

Personal project GitHub workflow management — branches, PRs, issues, releases, and CI/CD.

## Quick Start

```
"Create a PR for my current branch"
"Set up GitHub Actions CI for my React Native app"
"Create a new release v1.0.0"
"Open a GitHub issue for this bug"
"Show my open PRs"
```

---

## Branch Strategy

### Recommended for personal projects

```
main          ← production (protected, auto-deploy to prod)
develop       ← integration branch (optional for solo projects)
feature/xxx   ← new features
fix/xxx       ← bug fixes
release/x.x.x ← release prep
```

For solo projects, a simpler approach works well:
```
main          ← stable production code
dev           ← active development
```

### Create a feature branch
```bash
git checkout -b feature/my-new-feature
git push -u origin feature/my-new-feature
```

---

## Pull Requests

### Create PR with gh CLI
```bash
gh pr create \
  --title "feat: add user authentication" \
  --body "## Summary
- Adds Firebase Auth integration
- Login/logout flow
- Protected routes

## Testing
- [ ] Tested on iOS simulator
- [ ] Tested on Android emulator
- [ ] Auth persists across app restarts" \
  --assignee @me
```

### List open PRs
```bash
gh pr list
gh pr view --web  # open in browser
```

### Merge PR
```bash
gh pr merge --squash --delete-branch
```

---

## Issues

### Create an issue
```bash
gh issue create \
  --title "Bug: login screen freezes on Android" \
  --body "Steps to reproduce:
1. Open app
2. Tap Login
3. Screen freezes after 2 seconds

Expected: Login dialog appears
Actual: Freeze" \
  --label "bug"
```

### Useful labels to set up
```bash
gh label create "bug" --color "d73a4a"
gh label create "enhancement" --color "a2eeef"
gh label create "ios" --color "e4e669"
gh label create "android" --color "0075ca"
gh label create "web" --color "7057ff"
gh label create "blocked" --color "e11d48"
```

### List and close issues
```bash
gh issue list
gh issue close 42 --comment "Fixed in PR #43"
```

---

## Releases

### Create a release
```bash
# Tag the release
git tag -a v1.0.0 -m "First public release"
git push origin v1.0.0

# Create GitHub release with notes
gh release create v1.0.0 \
  --title "v1.0.0 — Initial Release" \
  --notes "## What's New
- Feature A
- Feature B
- Bug fix for issue #12" \
  --latest
```

### Semantic versioning guide
| Change | Version bump | Example |
|--------|-------------|---------|
| Breaking change | Major | 1.0.0 → 2.0.0 |
| New feature | Minor | 1.0.0 → 1.1.0 |
| Bug fix | Patch | 1.0.0 → 1.0.1 |
| Pre-release | Pre | 1.0.0-beta.1 |

---

## GitHub Actions CI/CD

### React Native / Expo — Basic CI
Create `.github/workflows/ci.yml`:
```yaml
name: CI
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
      - run: npm ci
      - run: npm test -- --passWithNoTests
      - run: npm run lint

  type-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
      - run: npm ci
      - run: npx tsc --noEmit
```

### Next.js — CI + Deploy to Vercel
```yaml
name: CI + Deploy
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
      - run: npm ci
      - run: npm run build
      - run: npm test -- --passWithNoTests

  deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - uses: amondnet/vercel-action@v25
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }}
          vercel-org-id: ${{ secrets.VERCEL_ORG_ID }}
          vercel-project-id: ${{ secrets.VERCEL_PROJECT_ID }}
          vercel-args: '--prod'
```

### Expo EAS Build — Mobile CI
```yaml
name: EAS Build
on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
      - run: npm ci
      - uses: expo/expo-github-action@v8
        with:
          expo-version: latest
          eas-version: latest
          token: ${{ secrets.EXPO_TOKEN }}
      - run: eas build --platform all --non-interactive
```

### Set up GitHub secrets
```bash
# For Vercel
gh secret set VERCEL_TOKEN
gh secret set VERCEL_ORG_ID
gh secret set VERCEL_PROJECT_ID

# For Expo
gh secret set EXPO_TOKEN

# For Firebase
gh secret set FIREBASE_TOKEN
```

---

## Repo Configuration Best Practices

### Branch protection (main branch)
```bash
gh api repos/:owner/:repo/branches/main/protection \
  --method PUT \
  --field required_status_checks='{"strict":true,"contexts":["test"]}' \
  --field enforce_admins=false \
  --field required_pull_request_reviews=null \
  --field restrictions=null
```

### Useful repo settings
```bash
# Enable auto-delete of merged branches
gh repo edit --delete-branch-on-merge

# Set default branch
gh repo edit --default-branch main
```

---

## Common Workflows

### Start a new feature
```bash
git checkout main && git pull
git checkout -b feature/my-feature
# ... make changes ...
git add -p  # stage interactively
git commit -m "feat: describe the change"
git push -u origin feature/my-feature
gh pr create --fill
```

### Fix a bug
```bash
git checkout -b fix/bug-description
# ... fix the bug ...
git commit -m "fix: describe what was broken and how it's fixed"
gh pr create --title "fix: bug description"
```

### Review your work before PR
```bash
git diff main...HEAD         # all changes from main
git log main...HEAD --oneline  # all commits
gh pr diff                   # if PR already exists
```
