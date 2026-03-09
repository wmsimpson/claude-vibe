---
name: app-setup
description: Scaffold and set up new mobile or web app projects. Handles React Native, Expo, Swift/iOS, Flutter, React/Next.js, and vanilla web apps. Sets up GitHub repo, project structure, dependencies, and deployment targets. Use when starting a new app project. Triggers on "new app", "scaffold", "set up a project", "create a React Native app", "start a web app", "initialize project".
user-invocable: true
---

# App Setup — New Project Scaffolding

Scaffold a new mobile or web application with best-practice structure, GitHub repo, and CI/CD ready.

## Quick Start

```
"Set up a new React Native app called MyApp"
"Create a Next.js web app for my portfolio"
"Scaffold a new Swift iOS app"
"Start a new Expo app with TypeScript"
"Initialize a new Flutter project"
```

---

## Phase 1 — Gather Requirements

Ask the user:

1. **App name:** What is the app called? (used for repo name, bundle ID, etc.)
2. **Platform:** iOS only / Android only / iOS + Android / Web only / Web + Mobile
3. **Tech stack:**
   - **Cross-platform mobile:** React Native (bare) | Expo (managed/bare) | Flutter
   - **Native iOS:** Swift / SwiftUI | Objective-C (rare)
   - **Native Android:** Kotlin | Java (rare)
   - **Web:** Next.js | React (Vite) | Astro | SvelteKit | Vanilla HTML/CSS/JS
   - **Full-stack:** Next.js + API routes | React + Node.js/Express | React + FastAPI
4. **Backend/Database:**
   - None (static/client-only)
   - Firebase (Google — easy auth + Firestore)
   - Supabase (open source, PostgreSQL)
   - PlanetScale / Neon / Turso (serverless SQL)
   - Custom backend (Node.js, FastAPI, etc.)
5. **Auth needed?** Yes → recommend Firebase Auth or Supabase Auth
6. **Deployment target:**
   - Mobile: App Store / Google Play / TestFlight (iOS beta) / Expo EAS
   - Web: Vercel | Netlify | Firebase Hosting | GitHub Pages
7. **GitHub repo:** Create new / use existing? Private or public?

Check `~/.vibe/profile` for user preferences (mobile_platform, web_deploy_target) and pre-fill.

---

## Phase 2 — Scaffold the App

### React Native (bare)
```bash
APP_NAME="MyApp"
cd ~/code
npx react-native@latest init $APP_NAME --template react-native-template-typescript
cd $APP_NAME
```

### Expo (recommended for most cross-platform projects)
```bash
APP_NAME="MyApp"
cd ~/code
npx create-expo-app@latest $APP_NAME --template
cd $APP_NAME
# Install common extras
npx expo install expo-router expo-status-bar @expo/vector-icons
```

### Next.js (recommended for web)
```bash
APP_NAME="my-app"
cd ~/code
npx create-next-app@latest $APP_NAME \
  --typescript --tailwind --eslint --app --src-dir
cd $APP_NAME
```

### React + Vite
```bash
APP_NAME="my-app"
cd ~/code
npm create vite@latest $APP_NAME -- --template react-ts
cd $APP_NAME && npm install
```

### SvelteKit
```bash
APP_NAME="my-app"
cd ~/code
npm create svelte@latest $APP_NAME
cd $APP_NAME && npm install
```

### Flutter
```bash
APP_NAME="my_app"
cd ~/code
flutter create $APP_NAME --org com.yourname
cd $APP_NAME
```

### Swift / SwiftUI (iOS native)
- Create new Xcode project: File → New → Project → App
- Set bundle ID: `com.yourname.AppName`
- Team: Personal Team or your Apple Developer account
- Save to `~/code/AppName`

---

## Phase 3 — GitHub Repository Setup

```bash
cd ~/code/$APP_NAME

# Initialize git (if not already done by scaffold)
git init
git add .
git commit -m "Initial scaffold: $APP_NAME"

# Create GitHub repo
gh repo create $APP_NAME \
  --private \          # or --public
  --description "Description of your app" \
  --source=. \
  --remote=origin \
  --push
```

Set up `.gitignore` — ensure these are excluded:
- `node_modules/`, `.expo/`, `ios/Pods/`, `.env*`, `*.xcworkspace/xcuserdata`
- Platform-specific build artifacts: `android/app/build/`, `ios/build/`, `.dart_tool/`

---

## Phase 4 — Environment Configuration

Create `.env.local` (never commit this):
```bash
cat > .env.local << 'EOF'
# App configuration
APP_NAME=MyApp
APP_ENV=development

# Firebase (if using)
# EXPO_PUBLIC_FIREBASE_API_KEY=
# EXPO_PUBLIC_FIREBASE_AUTH_DOMAIN=
# EXPO_PUBLIC_FIREBASE_PROJECT_ID=

# Supabase (if using)
# EXPO_PUBLIC_SUPABASE_URL=
# EXPO_PUBLIC_SUPABASE_ANON_KEY=

# API endpoints
# EXPO_PUBLIC_API_URL=http://localhost:3000
EOF
```

Add to `.gitignore`:
```
.env.local
.env*.local
```

---

## Phase 5 — Integrations (Prompted Based on User's Stack)

### Firebase
```bash
npm install firebase
# Then: npx firebase login && npx firebase init
```

### Supabase
```bash
npm install @supabase/supabase-js
```

### Expo EAS Build (cross-platform mobile CI/CD)
```bash
npm install -g eas-cli
eas login
eas build:configure
```

### React Navigation (React Native)
```bash
npm install @react-navigation/native @react-navigation/native-stack
npx expo install react-native-screens react-native-safe-area-context
```

### Tailwind CSS (web)
```bash
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
```

---

## Phase 6 — First Run Verification

```bash
cd ~/code/$APP_NAME

# Expo / React Native
npx expo start

# Next.js / React
npm run dev

# Flutter
flutter run

# Swift — open Xcode
open *.xcodeproj 2>/dev/null || open *.xcworkspace 2>/dev/null
```

Report: URL/port the app is running on, and any errors to fix.

---

## Phase 7 — Completion Summary

Deliver:
1. **Repo URL** — `https://github.com/USERNAME/APP_NAME`
2. **Local path** — `~/code/APP_NAME`
3. **Run command** — how to start the dev server
4. **Next steps:**
   - Set up deployment: *"Deploy my app to Vercel"* or `web-deployment` skill
   - Set up GitHub workflow: *"Set up a GitHub PR workflow"* or `github-workflow` skill
   - Debug an issue: *"Debug my React Native build"* or `app-debug` skill

---

## Stack Decision Guide

| Need | Recommendation | Why |
|------|---------------|-----|
| iOS + Android, fast iteration | Expo (managed) | Zero native config needed |
| iOS + Android, native modules | React Native (bare) | Full native access |
| iOS performance-critical | Swift/SwiftUI | Native speed, best iOS APIs |
| Web only, SEO matters | Next.js | SSR, great ecosystem |
| Web only, SPA | React + Vite | Fast, minimal config |
| iOS + Android + Web | Flutter | Single codebase, great perf |
| Prototype fast | Expo | Easiest to get running |
