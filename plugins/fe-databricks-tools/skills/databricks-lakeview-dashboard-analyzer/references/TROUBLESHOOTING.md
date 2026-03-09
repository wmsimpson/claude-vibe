# Troubleshooting Guide

Common issues and solutions when analyzing Databricks dashboards.

## MCP Server Disconnected

**Symptom:** `Error: Server 'chrome-devtools' is not connected (failed)`

**Causes:**
- Large screenshot response (500KB-1MB+)
- Network timeout
- Chrome DevTools protocol error

**Solutions:**

1. Check server status:
   ```bash
   mcp-cli servers
   ```

2. Restart if needed:
   ```
   /mcp
   ```

3. Use smaller screenshots (viewport only, not fullPage):
   ```bash
   mcp__chrome-devtools__take_screenshot '{}'  # viewport only
   # Avoid: '{"fullPage": true}' for complex dashboards
   ```

4. For complex dashboards, scroll and take multiple viewport screenshots instead of one full page screenshot.

---

## Authentication Issues

**Symptom:** Stuck on login page, can't access dashboard

**Solutions:**

1. User must authenticate manually in the browser
2. Wait longer for SSO redirects (30-60 seconds)
3. Check if cookies/session expired
4. Verify user has dashboard access in the workspace

**Checking auth status:**
```bash
mcp__chrome-devtools__evaluate_script '{"function": "() => ({ url: location.href, title: document.title })"}'
```

If URL contains `login`, `auth`, or `signin`, authentication is not complete.

---

## Session Expiration

**Symptom:** Dialog appears saying "Your session has expired" or "Log in"

**Solutions:**

```bash
# 1. Take snapshot to see the dialog
mcp__chrome-devtools__take_snapshot
# Look for: uid=X_Y button "Log in" or dialog elements

# 2. Click the login button
mcp__chrome-devtools__click '{"uid": "LOGIN_BUTTON_UID"}'

# 3. Wait for re-authentication to complete
mcp__chrome-devtools__wait_for '{"text": "Dashboard Title", "timeout": 30000}'

# 4. Continue with analysis
```

**Prevention:**
- Sessions typically last 30-60 minutes
- For long analysis workflows, save progress frequently
- Consider refreshing the page before starting complex operations

---

## Can't Find Elements

**Symptom:** UID not found or element not in snapshot

**Causes:**
- Snapshot is stale (page changed since snapshot was taken)
- Element is not visible (needs scrolling)
- Dynamic content hasn't loaded yet
- Element is in an iframe

**Solutions:**

1. **Take fresh snapshot** (snapshots become stale after page changes):
   ```bash
   mcp__chrome-devtools__take_snapshot
   ```

2. **Scroll to make element visible:**
   ```bash
   mcp__chrome-devtools__evaluate_script '{"function": "() => { const c = document.querySelector(\"[class*=dbsql-legacy-mfe-page]\"); if(c) c.scrollBy(0, 600); return true; }"}'
   ```

3. **Wait for dynamic content:**
   ```bash
   mcp__chrome-devtools__wait_for '{"text": "Expected Text", "timeout": 10000}'
   ```

4. **Check for iframes** (elements inside iframes may not appear in snapshot)

---

## Dashboard Won't Load

**Symptom:** Blank page, timeout, or error message

**Solutions:**

1. **Check URL pattern** - use `/dashboardsv3/` not `/sql/dashboards/`:
   - Correct: `https://adb-XXX.azuredatabricks.net/dashboardsv3/UUID/published`
   - Wrong: `https://adb-XXX.azuredatabricks.net/sql/dashboards/UUID`

2. **Verify dashboard is published** (not draft mode)

3. **Check for JavaScript errors:**
   ```bash
   mcp__chrome-devtools__list_console_messages '{}'
   ```

4. **Increase timeout:**
   ```bash
   mcp__chrome-devtools__wait_for '{"text": "Dashboard", "timeout": 30000}'
   ```

---

## Data Extraction Incomplete

**Symptom:** Tables paginated, can't see all rows

**Solutions:**

1. Look for pagination controls in snapshot
2. Click "Next page" or "Show all" buttons
3. Increase rows per page if option available
4. Extract page by page in a loop

```bash
# Find pagination
mcp__chrome-devtools__take_snapshot
# Look for "Next", "1 of 5", "Show 100"

# Click to show more
mcp__chrome-devtools__click '{"uid": "SHOW_ALL_UID"}'
```

---

## Scrolling Doesn't Work

**Symptom:** `window.scrollTo()` has no effect, content stays in same position

**Cause:** Dashboard content is in a scrollable container, not the main window.

**Solution:** Use container scrolling:

```bash
mcp__chrome-devtools__evaluate_script '{"function": "() => { const container = document.querySelector(\"[class*=dbsql-legacy-mfe-page]\"); if (container) { container.scrollBy(0, 600); return { scrolled: true, scrollTop: container.scrollTop }; } return { scrolled: false }; }"}'
```

**Verification:** After scrolling, `scrollTop` should increase. If it stays at 0, try a different container selector.

---

## Menu Collapses When Clicking

**Symptom:** Ellipsis menu or dropdown closes unexpectedly

**Cause:** Hover state changes when moving mouse to click submenu items.

**Solution:** Use keyboard navigation instead:

```bash
# After opening menu, navigate with keyboard
mcp__chrome-devtools__press_key '{"key": "ArrowDown"}'   # Move down
mcp__chrome-devtools__press_key '{"key": "ArrowRight"}'  # Expand submenu
mcp__chrome-devtools__press_key '{"key": "Enter"}'       # Select
```

---

## Screenshots Are Cut Off

**Symptom:** Chart or table is only partially visible in screenshot

**Solutions:**

1. **Scroll to center the element:**
   ```bash
   mcp__chrome-devtools__evaluate_script '{"function": "() => { const c = document.querySelector(\"[class*=dbsql-legacy-mfe-page]\"); if(c) c.scrollBy(0, 300); return true; }"}'
   ```

2. **Take full page screenshot** (use with caution - may disconnect MCP):
   ```bash
   mcp__chrome-devtools__take_screenshot '{"fullPage": true}'
   ```

3. **Take multiple viewport screenshots** at different scroll positions

---

## Filter Changes Don't Apply

**Symptom:** Clicking filter options doesn't update the dashboard

**Solutions:**

1. **Wait for refresh after filter change:**
   ```bash
   mcp__chrome-devtools__wait_for '{"text": "Loading", "timeout": 5000}'
   mcp__chrome-devtools__wait_for '{"text": "Showing results", "timeout": 10000}'
   ```

2. **Use URL parameter manipulation instead** (bypasses UI issues):
   ```bash
   mcp__chrome-devtools__navigate_page '{"type": "url", "url": "DASHBOARD_URL&f_PAGEID%7EFILTERID=NewValue"}'
   ```

3. **For React controlled inputs**, use the native setter technique (see ELEMENT_INTERACTION.md)

---

## Network Requests Failing

**Symptom:** Dashboard shows errors, data not loading

**Check network requests:**
```bash
mcp__chrome-devtools__list_network_requests '{"resourceTypes": ["xhr", "fetch"]}'
```

**Get details on a specific failed request:**
```bash
mcp__chrome-devtools__get_network_request '{"reqid": 123}'
```

Look for:
- 401/403: Authentication/authorization issues
- 404: Wrong URL or deleted resource
- 500: Server-side error
- CORS errors: May need different access approach
