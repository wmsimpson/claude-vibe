---
name: app-debug
description: Debug mobile and web app issues — build errors, runtime crashes, performance problems, network issues, and platform-specific bugs. Use for "debug my app", "fix build error", "app crashes", "slow performance", "network request failing", "white screen", "React Native error", "iOS build failing".
user-invocable: true
---

# App Debug

Systematic debugging for mobile (React Native, Expo, iOS/Swift, Flutter) and web (React, Next.js) apps.

## Quick Start

```
"My React Native app crashes on launch"
"Expo build fails with metro error"
"My Next.js page shows a white screen"
"Network requests are failing in my app"
"iOS build fails with CocoaPods error"
"Android build fails with Gradle error"
```

---

## Step 1 — Understand the Problem

Ask the user:
1. **What is the error?** (paste the full error message / stack trace)
2. **When does it occur?** (build time / runtime / specific action)
3. **Platform?** (iOS / Android / Web / all)
4. **Recent changes?** (what changed before this broke?)
5. **Environment?** (development / staging / production)

Read the project structure to understand the tech stack:
```bash
ls ~/code/$APP_NAME/
cat ~/code/$APP_NAME/package.json | python3 -c "import sys,json; d=json.load(sys.stdin); print('Framework:', list(d.get('dependencies',{}).keys())[:10])"
```

---

## React Native / Expo Debugging

### Metro bundler errors

**Clear all caches:**
```bash
npx expo start --clear
# Or for bare React Native:
npx react-native start --reset-cache
```

**Delete node_modules and reinstall:**
```bash
rm -rf node_modules
npm install  # or: yarn install / bun install
```

**Common Metro errors:**

| Error | Fix |
|-------|-----|
| `Unable to resolve module` | Check import path, run `npm install` |
| `Haste module naming collision` | `npx react-native start --reset-cache` |
| `EMFILE: too many open files` | `ulimit -n 65536` in terminal |
| `Cannot find module 'xxx'` | `npm install xxx` or check if it's in package.json |

### iOS-specific

**CocoaPods sync:**
```bash
cd ios
pod install --repo-update
cd ..
```

**Clean Xcode build:**
```bash
xcodebuild clean -workspace ios/MyApp.xcworkspace -scheme MyApp
# Or in Xcode: Product → Clean Build Folder (Cmd+Shift+K)
```

**Common iOS errors:**

| Error | Fix |
|-------|-----|
| `No bundle URL present` | Metro bundler not running; start with `npx expo start` |
| CocoaPods version mismatch | `cd ios && pod deintegrate && pod install` |
| `Signing certificate not found` | Xcode → Project Settings → Signing → fix team/cert |
| `Symbol not found` | `cd ios && pod install` (native module not linked) |
| Simulator can't be found | Xcode → Preferences → Locations → Command Line Tools |

**View iOS logs:**
```bash
xcrun simctl spawn booted log stream --predicate 'subsystem == "com.your.app"'
# Or in Xcode: Window → Devices and Simulators → device logs
```

### Android-specific

**Clean Gradle build:**
```bash
cd android
./gradlew clean
cd ..
```

**Common Android errors:**

| Error | Fix |
|-------|-----|
| `SDK location not found` | Create `android/local.properties`: `sdk.dir=/Users/$USER/Library/Android/sdk` |
| `Duplicate class` | Check for dependency conflicts in `build.gradle` |
| `Could not resolve xxx` | Add maven repo or update dependency version |
| `Manifest merger failed` | Identify conflicting attributes in AndroidManifest.xml |
| JDK version mismatch | Set JDK in Android Studio → Project Structure |

**View Android logs:**
```bash
adb logcat | grep -i "ReactNativeJS\|FATAL\|Error"
# Or: npx react-native log-android
```

### Expo-specific

**EAS Build failures:**
```bash
# Check build logs
eas build:list
eas build:view [BUILD_ID]

# Common fixes:
# - Update eas-cli: npm install -g eas-cli
# - Clear EAS cache: eas build --clear-cache
# - Check eas.json for config errors
```

**Expo Doctor:**
```bash
npx expo doctor
# Shows all dependency version issues
npx expo install --fix  # auto-fix compatible versions
```

---

## React / Next.js Debugging

### White screen / blank page

1. Open browser DevTools (F12) → Console tab
2. Look for red errors — most common: module not found, syntax error, undefined variable
3. Check Network tab for failed resource loads

```bash
# Check for TypeScript errors
npx tsc --noEmit

# Check for ESLint issues
npm run lint

# Build to find production errors
npm run build 2>&1 | head -50
```

### Next.js specific errors

| Error | Fix |
|-------|-----|
| `Hydration mismatch` | Component renders differently server vs client; check for `window` usage |
| `Cannot read properties of undefined` | Guard with optional chaining: `obj?.prop` |
| `Module not found` | Check import path case sensitivity (Linux is case-sensitive) |
| 404 on static assets | Check `public/` directory and `basePath` in next.config.js |
| API route not working | Check file is in `pages/api/` or `app/api/`, correct export |

**Next.js debug mode:**
```bash
NODE_OPTIONS='--inspect' npm run dev
# Open chrome://inspect in Chrome
```

### React performance issues

```bash
# Install React DevTools
npm install -g react-devtools
react-devtools

# Profile component re-renders
# In browser: React DevTools → Profiler tab → Record
```

---

## Network / API Debugging

### Log all network requests (Expo/React Native)

Add to app entry point:
```js
// Debug network requests
const originalFetch = global.fetch;
global.fetch = async (...args) => {
  console.log('FETCH:', args[0]);
  const result = await originalFetch(...args);
  console.log('RESPONSE:', result.status);
  return result;
};
```

### Common network errors

| Error | Fix |
|-------|-----|
| CORS error (web) | Backend needs `Access-Control-Allow-Origin` header |
| Certificate error (iOS) | Add domain to `NSAppTransportSecurity` in Info.plist |
| `Network request failed` (RN) | Check if backend is reachable; use computer's IP not `localhost` on device |
| 401 Unauthorized | Check auth token is being sent correctly |
| 429 Too Many Requests | Add rate limiting/backoff in app |

**Test API directly:**
```bash
curl -v https://your-api.com/endpoint \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json"
```

---

## Performance Debugging

### React Native performance

```bash
# Enable Hermes (if not already)
# In app.json: "jsEngine": "hermes"

# Profile with Flipper
npx flipper-pkg  # Flipper is RN's native debugging tool
```

**Common performance fixes:**
- Wrap expensive components with `React.memo`
- Use `useCallback` for event handlers passed to children
- Use `FlatList` instead of `ScrollView` for long lists
- Avoid inline object/array creation in render: `style={{ flex: 1 }}` → extract to `StyleSheet.create`

### Web performance

```bash
# Run Lighthouse
npx lighthouse https://localhost:3000 --view

# Bundle size analysis (Next.js)
npx @next/bundle-analyzer

# Bundle size analysis (Vite/CRA)
npx source-map-explorer build/static/js/*.js
```

---

## Dependency Issues

### Check for outdated / vulnerable packages
```bash
npm outdated
npm audit

# Fix vulnerabilities
npm audit fix

# Update to latest compatible
npx npm-check-updates -u && npm install
```

### Version conflicts
```bash
npm ls PACKAGE_NAME  # see all installed versions
```

---

## Environment Variable Issues

**Check env vars are loaded:**
```bash
# Next.js — only vars prefixed NEXT_PUBLIC_ are available in browser
# Expo — only vars prefixed EXPO_PUBLIC_ are available in app

# Debug: add temporary log
console.log('API URL:', process.env.NEXT_PUBLIC_API_URL);
```

**Expo env var precedence:**
```
.env.local > .env.development / .env.production > .env
```

---

## Getting More Help

If the error is still unresolved:
1. **Search error in GitHub issues:** `gh issue list` on the relevant repo
2. **Search Stack Overflow / GitHub:** include the full error message
3. **Check framework changelog:** breaking changes between versions
4. **Minimal reproduction:** isolate the issue in a fresh app to confirm it's not a project-specific config issue
