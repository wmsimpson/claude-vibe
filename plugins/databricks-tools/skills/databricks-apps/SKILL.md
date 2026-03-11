---
name: databricks-apps
description: Build and deploy full-stack Databricks Apps with Lakebase, Foundation Model API, and React/FastAPI
---

# Databricks Apps Development Skill

Build production-ready Databricks Apps with database integration (Lakebase), AI features (Foundation Model API), and beautiful React frontends.

## Prerequisites

1. **FE-VM Workspace** - Required for Lakebase and Foundation Models
   - Use `/databricks-fe-vm-workspace-deployment` skill to get a workspace
   - Need "serverless" workspace type for Lakebase support

2. **Databricks CLI** - Version 0.229.0+
   - Authenticate: `databricks auth login --host <workspace-url> --profile <profile-name>`

3. **uv** - Python package manager (required for Python backends)
   - Install: `curl -LsSf https://astral.sh/uv/install.sh | sh`

## Quick Start

### Step 1: Choose Architecture

**Node.js Frontend + FastAPI Backend** (Recommended for complex apps):
```
my-app/
├── app.yaml              # Databricks app config
├── app.py                # FastAPI entry point
├── pyproject.toml        # Python dependencies (uv)
├── requirements.txt      # Generated for deployment
├── server/               # Backend code
│   ├── config.py         # Dual-mode auth
│   ├── db.py             # Lakebase connection
│   ├── llm.py            # Foundation Model client
│   └── routes/           # API endpoints
├── frontend/             # React app
│   ├── package.json
│   ├── vite.config.ts
│   └── src/
└── .gitignore            # CRITICAL - exclude node_modules, .venv
```

**Pure Node.js** (Simpler apps):
```
my-app/
├── app.yaml
├── package.json
├── server.js             # Express server
└── client/               # React app (optional)
```

### Step 2: Set Up Project

```bash
# Create directory
mkdir my-app && cd my-app

# Initialize Python backend with uv
uv init
uv add fastapi uvicorn asyncpg aiohttp openai databricks-sdk pydantic python-multipart

# Export clean requirements.txt for deployment
cat > requirements.txt << 'EOF'
fastapi>=0.115.0
uvicorn[standard]>=0.30.0
asyncpg>=0.29.0
aiohttp>=3.9.0
openai>=1.52.0
databricks-sdk>=0.30.0
pydantic>=2.0.0
python-multipart>=0.0.9
EOF

# Initialize React frontend
cd frontend && npm create vite@latest . -- --template react-ts
npm install zustand react-router-dom lucide-react
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
cd ..
```

### Step 3: Configure .gitignore

**CRITICAL** - Create this BEFORE deployment to avoid uploading thousands of files:

```gitignore
# Python
__pycache__/
*.py[cod]
.venv/
venv/
.env

# Node
node_modules/
npm-debug.log*

# Build outputs (keep frontend/dist for deployment!)
# frontend/dist/  # DO NOT exclude this!

# IDE
.idea/
.vscode/
*.swp

# Databricks
.databricks/

# OS
.DS_Store
```

### Step 4: Create app.yaml

```yaml
command:
  - "python"
  - "-m"
  - "uvicorn"
  - "app:app"
  - "--host"
  - "0.0.0.0"
  - "--port"
  - "8000"

env:
  # Lakebase connection (auto-populated when resource added)
  - name: PGHOST
    valueFrom: database
  - name: PGPORT
    valueFrom: database
  - name: PGDATABASE
    valueFrom: database
  - name: PGUSER
    valueFrom: database

  # Foundation Model endpoint
  - name: SERVING_ENDPOINT
    value: databricks-claude-sonnet-4-5
```

### Step 5: Implement Dual-Mode Authentication

The key pattern: detect if running in Databricks Apps vs locally.

**server/config.py**:
```python
import os
from databricks.sdk import WorkspaceClient

# Detect environment
IS_DATABRICKS_APP = bool(os.environ.get("DATABRICKS_APP_NAME"))

def get_workspace_client() -> WorkspaceClient:
    """Get authenticated WorkspaceClient."""
    if IS_DATABRICKS_APP:
        # Remote: Uses auto-injected service principal credentials
        return WorkspaceClient()
    else:
        # Local: Uses Databricks CLI profile
        profile = os.environ.get("DATABRICKS_PROFILE", "DEFAULT")
        return WorkspaceClient(profile=profile)

def get_oauth_token() -> str:
    """Get OAuth token for Lakebase authentication."""
    client = get_workspace_client()
    return client.config.authenticate().token

def get_workspace_host() -> str:
    """Get workspace host URL with https:// prefix."""
    if IS_DATABRICKS_APP:
        # IMPORTANT: DATABRICKS_HOST in Databricks Apps is just hostname, no scheme
        host = os.environ.get("DATABRICKS_HOST", "")
        if host and not host.startswith("http"):
            host = f"https://{host}"
        return host
    client = get_workspace_client()
    return client.config.host  # SDK includes https://
```

### Step 6: Set Up Lakebase Connection

**server/db.py**:
```python
import os
import asyncpg
from typing import Optional
from .config import get_oauth_token, IS_DATABRICKS_APP

class DatabasePool:
    def __init__(self):
        self._pool: Optional[asyncpg.Pool] = None
        self._demo_mode = False

    async def get_pool(self) -> Optional[asyncpg.Pool]:
        # Check if Lakebase is configured
        if not os.environ.get("PGHOST"):
            self._demo_mode = True
            return None

        # Create or refresh pool
        if self._pool is None:
            try:
                token = get_oauth_token()
                self._pool = await asyncpg.create_pool(
                    host=os.environ["PGHOST"],
                    port=int(os.environ.get("PGPORT", "5432")),
                    database=os.environ["PGDATABASE"],
                    user=os.environ["PGUSER"],
                    password=token,
                    ssl="require",
                    min_size=2,
                    max_size=10,
                )
            except Exception as e:
                print(f"Lakebase connection failed: {e}")
                self._demo_mode = True
                return None
        return self._pool

    async def refresh_token(self):
        """Refresh OAuth token (call every ~45 minutes)."""
        if self._pool:
            await self._pool.close()
            self._pool = None
        await self.get_pool()

    @property
    def is_demo_mode(self) -> bool:
        return self._demo_mode

db = DatabasePool()
```

### Step 7: Set Up Foundation Model API

**server/llm.py**:
```python
import os
from openai import AsyncOpenAI
from .config import get_oauth_token, get_workspace_host, IS_DATABRICKS_APP

def get_llm_client() -> AsyncOpenAI:
    """Get OpenAI-compatible client for Databricks Foundation Models."""
    host = get_workspace_host()

    if IS_DATABRICKS_APP:
        # Remote: Use service principal token
        token = os.environ.get("DATABRICKS_TOKEN") or get_oauth_token()
    else:
        # Local: Use profile token
        token = get_oauth_token()

    return AsyncOpenAI(
        api_key=token,
        base_url=f"{host}/serving-endpoints"
    )

async def chat_completion(messages: list, model: str = "databricks-claude-sonnet-4-5"):
    """Get chat completion from Foundation Model."""
    client = get_llm_client()
    response = await client.chat.completions.create(
        model=model,
        messages=messages,
        max_tokens=4096,
        temperature=0.7,
    )
    return response.choices[0].message.content
```

### Step 8: Create FastAPI App

**app.py**:
```python
from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
from fastapi.responses import FileResponse
import os

app = FastAPI(title="My Databricks App")

# Import routes
from server.routes import restaurants, cart, orders, chat

app.include_router(restaurants.router, prefix="/api")
app.include_router(cart.router, prefix="/api")
app.include_router(orders.router, prefix="/api")
app.include_router(chat.router, prefix="/api")

# Serve React frontend
frontend_dir = os.path.join(os.path.dirname(__file__), "frontend", "dist")
if os.path.exists(frontend_dir):
    app.mount("/assets", StaticFiles(directory=os.path.join(frontend_dir, "assets")), name="assets")

    @app.get("/{full_path:path}")
    async def serve_spa(full_path: str):
        return FileResponse(os.path.join(frontend_dir, "index.html"))
```

## Local Testing Workflow

### Step 1: Start Backend
```bash
cd my-app
export DATABRICKS_PROFILE=my-fevm-profile
uv run uvicorn app:app --reload --port 8000
```

### Step 2: Start Frontend (Dev Mode)
```bash
cd frontend
npm run dev  # Runs on port 5173 with proxy to 8000
```

### Step 3: Test with Chrome DevTools MCP

**IMPORTANT**: Always validate UI with Chrome DevTools before deployment.

```bash
# Navigate to app
mcp-cli call chrome-devtools/navigate_page '{"type": "url", "url": "http://localhost:5173"}'

# Take screenshot
mcp-cli call chrome-devtools/take_screenshot '{"filePath": "/tmp/app-screenshot.png"}'

# Check for console errors
mcp-cli call chrome-devtools/list_console_messages '{}'

# Get page snapshot for interactions
mcp-cli call chrome-devtools/take_snapshot '{}'

# Click elements
mcp-cli call chrome-devtools/click '{"uid": "element-uid"}'
```

### Step 4: Build Frontend for Production
```bash
cd frontend
npm run build  # Outputs to frontend/dist/
```

## Deployment

### Step 1: Create App
```bash
databricks apps create my-app --description "My Databricks App" -p my-profile
```

### Step 2: Upload Files (Excluding node_modules/.venv)
```bash
# Use databricks sync with excludes
databricks sync . /Users/user@example.com/my-app \
  --exclude node_modules \
  --exclude .venv \
  --exclude __pycache__ \
  --exclude .git \
  --exclude "frontend/src" \
  --exclude "frontend/public" \
  -p my-profile

# Upload built frontend separately if needed
databricks workspace import-dir frontend/dist /Users/user@example.com/my-app/frontend/dist -p my-profile
```

### Step 3: Deploy
```bash
databricks apps deploy my-app \
  --source-code-path /Workspace/Users/user@example.com/my-app \
  -p my-profile
```

### Step 4: Add Resources (Via UI)
1. Go to Compute > Apps > my-app > Edit
2. Add "Database" resource → Select Lakebase instance → Permission: "Can connect"
3. Add "Model serving endpoint" → Select Foundation Model → Permission: "Can query"
4. Redeploy to pick up new environment variables

### Step 5: Verify Remote Deployment
```bash
# Get app URL
databricks apps get my-app -p my-profile

# Test with Chrome DevTools
mcp-cli call chrome-devtools/navigate_page '{"type": "url", "url": "https://my-app-xxxx.aws.databricksapps.com"}'
```

## Monitoring & Logs

### Viewing Application Logs

Access logs directly by appending `/logz` to your app URL:

```
https://my-app-1234567890.my-instance.databricksapps.com/logz
```

This provides real-time access to application logs for debugging without needing to navigate through the Databricks UI.

**Via CLI**: Get your app URL then construct the logs URL:
```bash
# Get app details including URL
databricks apps get my-app -p my-profile

# The logs URL is your app URL + /logz
# Example: https://my-app-xxxx.aws.databricksapps.com/logz
```

**Via Chrome DevTools MCP**: Navigate directly to logs:
```bash
mcp-cli call chrome-devtools/navigate_page '{"type": "url", "url": "https://my-app-xxxx.aws.databricksapps.com/logz"}'
```

For more details, see: https://docs.databricks.com/aws/en/dev-tools/databricks-apps/monitor#application-logs

## Troubleshooting

### "Error installing packages"
- **Cause**: requirements.txt has uv-specific format or invalid packages
- **Fix**: Create clean requirements.txt with simple `package>=version` format

### "App Not Available"
- **Cause**: App not listening on port 8000
- **Fix**: Ensure uvicorn binds to `--port 8000` in app.yaml command

### OAuth Token Expires
- **Cause**: Lakebase tokens expire after 1 hour
- **Fix**: Implement token refresh (see db.py pattern above)

### node_modules Uploaded
- **Cause**: Missing .gitignore or wrong sync command
- **Fix**: Add .gitignore, use `databricks sync --exclude` patterns

### Frontend Not Loading
- **Cause**: frontend/dist not uploaded or wrong path
- **Fix**: Verify `npm run build` succeeded, upload dist directory separately

### 401 Unauthorized from Foundation Model API
- **Cause**: `DATABRICKS_HOST` env var is just hostname without `https://` scheme
- **Fix**: Always add `https://` prefix when using DATABRICKS_HOST in remote:
  ```python
  host = os.environ.get("DATABRICKS_HOST", "")
  if host and not host.startswith("http"):
      host = f"https://{host}"
  ```

### OAuth Token Returns None Locally
- **Cause**: Using `w.config.token` which is `None` for OAuth/U2M auth
- **Fix**: Use `w.config.authenticate()` which returns `{'Authorization': 'Bearer <token>'}`:
  ```python
  auth_headers = w.config.authenticate()
  if auth_headers and "Authorization" in auth_headers:
      token = auth_headers["Authorization"].replace("Bearer ", "")
  ```

### Function Calling Returns Dictionaries
- **Cause**: Databricks serving endpoint returns `tool_calls` as raw JSON dictionaries, not objects
- **Fix**: Wrap tool_calls in wrapper classes for attribute access:
  ```python
  class FunctionCall:
      def __init__(self, func_dict):
          self.name = func_dict.get("name", "")
          self.arguments = func_dict.get("arguments", "{}")

  class ToolCall:
      def __init__(self, tc_dict):
          self.id = tc_dict.get("id", "")
          self.function = FunctionCall(tc_dict.get("function", {}))

  # In response parsing:
  raw_tool_calls = message.get("tool_calls")
  if raw_tool_calls:
      tool_calls = [ToolCall(tc) for tc in raw_tool_calls]
  ```

## Reference: Foundation Models

| Model | Endpoint | Best For |
|-------|----------|----------|
| Claude Sonnet 4.5 | `databricks-claude-sonnet-4-5` | General purpose, function calling |
| Claude Opus 4.5 | `databricks-claude-opus-4-5` | Complex reasoning |
| Gemini 2.5 Pro | `databricks-gemini-2.5-pro` | Long context (1M tokens) |
| GPT-5 | `databricks-gpt-5` | Multimodal |
| Llama 3.3 70B | `databricks-meta-llama-3-3-70b-instruct` | Cost-effective |

## Agents to Use

- **databricks-apps-developer**: Detailed patterns for Node.js/React and Python/FastAPI
- **web-devloop-tester**: UI testing with Chrome DevTools MCP
- **databricks-resource-deployment**: Create Lakebase instances, serving endpoints

## Example Projects

See `learnings/doordash-clone/` for a complete example with:
- FastAPI backend with dual-mode auth
- Lakebase integration with OAuth token refresh
- Foundation Model API chatbot with function calling
- React frontend with TailwindCSS and Zustand
- Full deployment workflow
