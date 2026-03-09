---
name: web-deployment
description: Deploy web and mobile apps to production. Handles Vercel, Netlify, Firebase Hosting, GitHub Pages, Expo EAS, and App Store/Google Play submission. Use for "deploy my app", "publish to Vercel", "submit to App Store", "set up CI/CD deployment", "TestFlight build".
user-invocable: true
---

# Web & Mobile App Deployment

Deploy your app to production — web hosting, mobile app stores, and CI/CD pipelines.

## Quick Start

```
"Deploy my Next.js app to Vercel"
"Set up Firebase Hosting for my web app"
"Build and submit my Expo app to TestFlight"
"Publish my React Native app to the App Store"
"Deploy to GitHub Pages"
```

---

## Web Deployment

### Vercel (recommended for Next.js / React)

**First deploy:**
```bash
npm install -g vercel
vercel login
vercel  # runs from your project directory
```

Vercel will:
1. Auto-detect framework (Next.js, React, Vue, etc.)
2. Set up build command and output directory
3. Provide a preview URL

**Production deploy:**
```bash
vercel --prod
```

**Link existing project:**
```bash
vercel link  # connect to existing Vercel project
```

**Environment variables:**
```bash
vercel env add VARIABLE_NAME production
vercel env pull .env.local  # pull envs to local
```

**Auto-deploy from GitHub:** Connect repo at vercel.com → New Project → Import Git Repository. Vercel auto-deploys on every push to main.

---

### Netlify

**First deploy:**
```bash
npm install -g netlify-cli
netlify login
netlify deploy --dir=out  # or dist/, build/, .next/ etc.
netlify deploy --dir=out --prod  # production
```

**Create `netlify.toml` for automatic config:**
```toml
[build]
  command = "npm run build"
  publish = "out"  # Next.js static: "out", Vite: "dist", CRA: "build"

[[redirects]]
  from = "/*"
  to = "/index.html"
  status = 200  # Required for SPA routing
```

**Deploy from GitHub:** netlify.com → New site from Git → Connect GitHub.

---

### Firebase Hosting (good for React SPAs + Firebase backend)

**Setup:**
```bash
npm install -g firebase-tools
firebase login
firebase init hosting
```

Firebase will ask:
- Public directory: `build` (CRA) or `dist` (Vite) or `out` (Next.js static)
- Single-page app: Yes (for React/Vue SPAs)
- GitHub Actions: Yes (for auto-deploy)

**Deploy:**
```bash
npm run build
firebase deploy --only hosting
```

**Preview channels:**
```bash
firebase hosting:channel:deploy preview-name --expires 7d
```

---

### GitHub Pages (free, good for static sites/portfolios)

**For Next.js static export:**
```js
// next.config.js
const nextConfig = {
  output: 'export',
  basePath: '/your-repo-name',  // if not using custom domain
}
module.exports = nextConfig
```

**GitHub Actions deploy:**
Create `.github/workflows/deploy.yml`:
```yaml
name: Deploy to GitHub Pages
on:
  push:
    branches: [main]
permissions:
  contents: read
  pages: write
  id-token: write
jobs:
  deploy:
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
      - run: npm ci && npm run build
      - uses: actions/upload-pages-artifact@v3
        with:
          path: out
      - uses: actions/deploy-pages@v4
        id: deployment
```

Enable in repo Settings → Pages → GitHub Actions source.

---

## Mobile Deployment

### Expo EAS Build + Submit

**Setup:**
```bash
npm install -g eas-cli
eas login
eas build:configure
```

This creates `eas.json`:
```json
{
  "build": {
    "development": {
      "developmentClient": true,
      "distribution": "internal"
    },
    "preview": {
      "distribution": "internal"
    },
    "production": {}
  },
  "submit": {
    "production": {}
  }
}
```

**Build for TestFlight (iOS):**
```bash
eas build --platform ios --profile preview
```

**Build for Google Play (Android):**
```bash
eas build --platform android --profile preview
```

**Submit to App Store:**
```bash
eas submit --platform ios
# Requires: Apple ID + app-specific password or API key
```

**Submit to Google Play:**
```bash
eas submit --platform android
# Requires: Google Play service account JSON key
```

**OTA Updates (no app store review needed for JS changes):**
```bash
eas update --branch production --message "Fix login bug"
```

---

### React Native (bare) — Manual Build

**iOS (Xcode):**
```bash
cd ios && pod install && cd ..
# Open Xcode: open ios/MyApp.xcworkspace
# Product → Archive → Distribute App → App Store Connect
```

**Android:**
```bash
cd android
./gradlew bundleRelease
# AAB at: android/app/build/outputs/bundle/release/app-release.aab
# Upload to Google Play Console
```

---

### TestFlight Distribution (iOS beta)

1. Archive in Xcode: Product → Archive
2. Distribute: Distribute App → App Store Connect → Upload
3. In App Store Connect: TestFlight → Add external testers
4. Testers receive invite email

For Expo: `eas build --platform ios --profile preview` then `eas submit --platform ios` with TestFlight distribution.

---

### App Store / Google Play Submission Checklist

**App Store (iOS):**
- [ ] Bundle ID set and registered in Apple Developer portal
- [ ] App icons (1024x1024 + all sizes)
- [ ] Screenshots for all required device sizes
- [ ] App description and keywords
- [ ] Privacy policy URL
- [ ] Age rating questionnaire completed
- [ ] In-app purchases configured (if applicable)
- [ ] Build uploaded via Xcode Archive or EAS Submit

**Google Play (Android):**
- [ ] Package name set
- [ ] App icons (512x512 + feature graphic 1024x500)
- [ ] Screenshots for phone + tablet
- [ ] Short/full description
- [ ] Content rating questionnaire
- [ ] Target API level meets requirements
- [ ] Signed AAB/APK uploaded

---

## Environment Management

### Staging vs Production

**Vercel — Multiple environments:**
```bash
vercel env add API_URL development  # dev
vercel env add API_URL preview       # PRs
vercel env add API_URL production    # main branch
```

**Expo — Multiple app variants:**
```json
// app.json / app.config.js
{
  "expo": {
    "name": "MyApp",
    "slug": "my-app",
    "extra": {
      "apiUrl": process.env.API_URL || "https://api.example.com"
    }
  }
}
```

---

## Custom Domain Setup

**Vercel:**
```bash
vercel domains add yourdomain.com
# Then update DNS at your registrar: CNAME → cname.vercel-dns.com
```

**Netlify:**
Site settings → Domain management → Add custom domain

**Firebase:**
```bash
firebase hosting:sites:create yourdomain-com
firebase target:apply hosting yourdomain yourdomain-com
```

---

## Monitoring

### Basic analytics (free)
- **Vercel Analytics:** Built-in for Vercel-hosted apps (enable in dashboard)
- **Firebase Analytics:** `expo install expo-firebase-analytics`
- **Sentry (error tracking):** `npx expo install sentry-expo`

### Performance
- **Vercel Speed Insights:** Enable in Vercel dashboard
- **Lighthouse:** `npx lighthouse https://your-app.com --view`
