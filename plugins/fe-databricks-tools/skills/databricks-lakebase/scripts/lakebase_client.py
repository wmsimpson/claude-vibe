"""
Lakebase Client - Python helper for connecting to Databricks Lakebase

Supports both Autoscaling Tier (preferred) and Provisioned Tier.

Autoscaling Tier Usage:
    from lakebase_client import AutoscalingLakebaseClient

    client = AutoscalingLakebaseClient("my-app", "production", "primary", "my-profile")
    conn = client.connect("mydb")

    cur = conn.cursor()
    cur.execute("SELECT * FROM users")
    print(cur.fetchall())
    conn.close()

Provisioned Tier Usage:
    from lakebase_client import LakebaseClient

    client = LakebaseClient("my-instance", "my-profile")
    conn = client.connect("mydb")

    cur = conn.cursor()
    cur.execute("SELECT * FROM users")
    print(cur.fetchall())
    conn.close()
"""

import subprocess
import json
from typing import Optional
from urllib.parse import quote_plus


class LakebaseClient:
    """Client for connecting to Databricks Lakebase databases."""

    def __init__(self, instance_name: str, profile: str, user_email: Optional[str] = None):
        """
        Initialize Lakebase client.

        Args:
            instance_name: Name of the Lakebase instance
            profile: Databricks CLI profile name
            user_email: Your Databricks email (optional, auto-detected if not provided)
        """
        self.instance_name = instance_name
        self.profile = profile
        self._user_email = user_email
        self._host: Optional[str] = None
        self._token: Optional[str] = None

    @property
    def user_email(self) -> str:
        """Get user email for authentication."""
        if self._user_email:
            return self._user_email

        result = subprocess.run([
            'databricks', 'current-user', 'me',
            '--profile', self.profile,
            '--output', 'json'
        ], capture_output=True, text=True)

        if result.returncode != 0:
            raise RuntimeError(f"Failed to get user info: {result.stderr}")

        user_info = json.loads(result.stdout)
        self._user_email = user_info['userName']
        return self._user_email

    @property
    def host(self) -> str:
        """Get the instance hostname."""
        if self._host:
            return self._host

        result = subprocess.run([
            'databricks', 'database', 'get-database-instance', self.instance_name,
            '--profile', self.profile,
            '--output', 'json'
        ], capture_output=True, text=True)

        if result.returncode != 0:
            raise RuntimeError(f"Failed to get instance info: {result.stderr}")

        instance = json.loads(result.stdout)

        if instance.get('state') != 'AVAILABLE':
            raise RuntimeError(f"Instance is not available. State: {instance.get('state')}")

        self._host = instance['read_write_dns']
        return self._host

    def generate_token(self) -> str:
        """Generate a new OAuth token."""
        result = subprocess.run([
            'databricks', 'database', 'generate-database-credential',
            '--json', json.dumps({
                "request_id": "python-client",
                "instance_names": [self.instance_name]
            }),
            '--profile', self.profile,
            '--output', 'json'
        ], capture_output=True, text=True)

        if result.returncode != 0:
            raise RuntimeError(f"Failed to generate credentials: {result.stderr}")

        creds = json.loads(result.stdout)
        self._token = creds['token']
        return self._token

    def connect(self, database: str = "postgres"):
        """
        Connect to a Lakebase database.

        Args:
            database: Database name (default: postgres)

        Returns:
            psycopg2 connection object
        """
        try:
            import psycopg2
        except ImportError:
            raise ImportError("psycopg2 is required. Install with: pip install psycopg2-binary")

        token = self.generate_token()

        return psycopg2.connect(
            host=self.host,
            port=5432,
            database=database,
            user=self.user_email,
            password=token,
            sslmode='require'
        )

    def get_sqlalchemy_url(self, database: str = "postgres") -> str:
        """
        Get SQLAlchemy connection URL.

        Args:
            database: Database name (default: postgres)

        Returns:
            SQLAlchemy connection URL string
        """
        token = self.generate_token()
        encoded_token = quote_plus(token)
        return f"postgresql://{self.user_email}:{encoded_token}@{self.host}:5432/{database}?sslmode=require"

    def get_sqlalchemy_engine(self, database: str = "postgres"):
        """
        Get SQLAlchemy engine.

        Args:
            database: Database name (default: postgres)

        Returns:
            SQLAlchemy engine
        """
        try:
            from sqlalchemy import create_engine
        except ImportError:
            raise ImportError("SQLAlchemy is required. Install with: pip install sqlalchemy")

        return create_engine(self.get_sqlalchemy_url(database))

    def get_connection_info(self) -> dict:
        """Get connection information for manual connections."""
        token = self.generate_token()
        return {
            'host': self.host,
            'port': 5432,
            'user': self.user_email,
            'password': token,
            'sslmode': 'require'
        }


class AutoscalingLakebaseClient:
    """Client for connecting to Databricks Lakebase Autoscaling Tier databases.

    Uses the `databricks postgres` CLI commands (requires CLI v0.285.0+).
    """

    def __init__(self, project_id: str, branch_id: str, endpoint_id: str,
                 profile: str, user_email: Optional[str] = None):
        """
        Initialize Autoscaling Lakebase client.

        Args:
            project_id: Project ID (e.g., "my-app")
            branch_id: Branch ID (e.g., "production")
            endpoint_id: Endpoint ID (e.g., "primary")
            profile: Databricks CLI profile name
            user_email: Your Databricks email (optional, auto-detected if not provided)
        """
        self.project_id = project_id
        self.branch_id = branch_id
        self.endpoint_id = endpoint_id
        self.profile = profile
        self._user_email = user_email
        self._host: Optional[str] = None
        self._token: Optional[str] = None

    @property
    def endpoint_path(self) -> str:
        """Full endpoint resource path."""
        return f"projects/{self.project_id}/branches/{self.branch_id}/endpoints/{self.endpoint_id}"

    @property
    def branch_path(self) -> str:
        """Full branch resource path."""
        return f"projects/{self.project_id}/branches/{self.branch_id}"

    @property
    def user_email(self) -> str:
        """Get user email for authentication."""
        if self._user_email:
            return self._user_email

        result = subprocess.run([
            'databricks', 'current-user', 'me',
            '--profile', self.profile,
            '--output', 'json'
        ], capture_output=True, text=True)

        if result.returncode != 0:
            raise RuntimeError(f"Failed to get user info: {result.stderr}")

        user_info = json.loads(result.stdout)
        self._user_email = user_info['userName']
        return self._user_email

    @property
    def host(self) -> str:
        """Get the endpoint hostname."""
        if self._host:
            return self._host

        result = subprocess.run([
            'databricks', 'postgres', 'get-endpoint', self.endpoint_path,
            '--profile', self.profile,
            '--output', 'json'
        ], capture_output=True, text=True)

        if result.returncode != 0:
            raise RuntimeError(f"Failed to get endpoint info: {result.stderr}")

        endpoint = json.loads(result.stdout)

        state = endpoint.get('status', {}).get('current_state')
        if state != 'ACTIVE':
            raise RuntimeError(f"Endpoint is not active. State: {state}")

        self._host = endpoint['status']['hosts']['host']
        return self._host

    def generate_token(self) -> str:
        """Generate a new OAuth token."""
        result = subprocess.run([
            'databricks', 'postgres', 'generate-database-credential',
            self.endpoint_path,
            '--profile', self.profile,
            '--output', 'json'
        ], capture_output=True, text=True)

        if result.returncode != 0:
            raise RuntimeError(f"Failed to generate credentials: {result.stderr}")

        creds = json.loads(result.stdout)
        self._token = creds['token']
        return self._token

    def connect(self, database: str = "postgres"):
        """
        Connect to a Lakebase database.

        Args:
            database: Database name (default: postgres)

        Returns:
            psycopg2 connection object
        """
        try:
            import psycopg2
        except ImportError:
            raise ImportError("psycopg2 is required. Install with: pip install psycopg2-binary")

        token = self.generate_token()

        return psycopg2.connect(
            host=self.host,
            port=5432,
            database=database,
            user=self.user_email,
            password=token,
            sslmode='require'
        )

    def get_sqlalchemy_url(self, database: str = "postgres") -> str:
        """
        Get SQLAlchemy connection URL.

        Args:
            database: Database name (default: postgres)

        Returns:
            SQLAlchemy connection URL string
        """
        token = self.generate_token()
        encoded_token = quote_plus(token)
        return f"postgresql://{self.user_email}:{encoded_token}@{self.host}:5432/{database}?sslmode=require"

    def get_sqlalchemy_engine(self, database: str = "postgres"):
        """
        Get SQLAlchemy engine.

        Args:
            database: Database name (default: postgres)

        Returns:
            SQLAlchemy engine
        """
        try:
            from sqlalchemy import create_engine
        except ImportError:
            raise ImportError("SQLAlchemy is required. Install with: pip install sqlalchemy")

        return create_engine(self.get_sqlalchemy_url(database))

    def get_connection_info(self) -> dict:
        """Get connection information for manual connections."""
        token = self.generate_token()
        return {
            'host': self.host,
            'port': 5432,
            'user': self.user_email,
            'password': token,
            'sslmode': 'require'
        }


class AsyncAutoscalingLakebaseClient:
    """Async client for Autoscaling Lakebase databases using asyncpg."""

    def __init__(self, project_id: str, branch_id: str, endpoint_id: str,
                 profile: str, user_email: Optional[str] = None):
        self._sync_client = AutoscalingLakebaseClient(
            project_id, branch_id, endpoint_id, profile, user_email
        )

    async def connect(self, database: str = "postgres"):
        """Connect asynchronously."""
        try:
            import asyncpg
        except ImportError:
            raise ImportError("asyncpg is required. Install with: pip install asyncpg")

        token = self._sync_client.generate_token()

        return await asyncpg.connect(
            host=self._sync_client.host,
            port=5432,
            database=database,
            user=self._sync_client.user_email,
            password=token,
            ssl='require'
        )

    async def create_pool(self, database: str = "postgres", min_size: int = 2, max_size: int = 10):
        """Create an asyncpg connection pool."""
        try:
            import asyncpg
        except ImportError:
            raise ImportError("asyncpg is required. Install with: pip install asyncpg")

        token = self._sync_client.generate_token()

        return await asyncpg.create_pool(
            host=self._sync_client.host,
            port=5432,
            database=database,
            user=self._sync_client.user_email,
            password=token,
            ssl='require',
            min_size=min_size,
            max_size=max_size
        )


class AsyncLakebaseClient:
    """Async client for connecting to Databricks Lakebase databases using asyncpg."""

    def __init__(self, instance_name: str, profile: str, user_email: Optional[str] = None):
        """
        Initialize async Lakebase client.

        Args:
            instance_name: Name of the Lakebase instance
            profile: Databricks CLI profile name
            user_email: Your Databricks email (optional, auto-detected if not provided)
        """
        self._sync_client = LakebaseClient(instance_name, profile, user_email)

    async def connect(self, database: str = "postgres"):
        """
        Connect to a Lakebase database asynchronously.

        Args:
            database: Database name (default: postgres)

        Returns:
            asyncpg connection object
        """
        try:
            import asyncpg
        except ImportError:
            raise ImportError("asyncpg is required. Install with: pip install asyncpg")

        token = self._sync_client.generate_token()

        return await asyncpg.connect(
            host=self._sync_client.host,
            port=5432,
            database=database,
            user=self._sync_client.user_email,
            password=token,
            ssl='require'
        )

    async def create_pool(self, database: str = "postgres", min_size: int = 2, max_size: int = 10):
        """
        Create an asyncpg connection pool.

        Args:
            database: Database name (default: postgres)
            min_size: Minimum pool size
            max_size: Maximum pool size

        Returns:
            asyncpg pool object
        """
        try:
            import asyncpg
        except ImportError:
            raise ImportError("asyncpg is required. Install with: pip install asyncpg")

        token = self._sync_client.generate_token()

        return await asyncpg.create_pool(
            host=self._sync_client.host,
            port=5432,
            database=database,
            user=self._sync_client.user_email,
            password=token,
            ssl='require',
            min_size=min_size,
            max_size=max_size
        )


if __name__ == "__main__":
    import sys

    if len(sys.argv) < 3:
        print("Usage (Autoscaling): python lakebase_client.py --autoscaling <project> <branch> <endpoint> <profile> [database]")
        print("Usage (Provisioned): python lakebase_client.py <instance_name> <profile> [database]")
        print()
        print("Examples:")
        print("  python lakebase_client.py --autoscaling my-app production primary my-profile mydb")
        print("  python lakebase_client.py my-lakebase my-profile mydb")
        sys.exit(1)

    if sys.argv[1] == "--autoscaling":
        if len(sys.argv) < 6:
            print("Usage: python lakebase_client.py --autoscaling <project> <branch> <endpoint> <profile> [database]")
            sys.exit(1)
        project_id = sys.argv[2]
        branch_id = sys.argv[3]
        endpoint_id = sys.argv[4]
        profile = sys.argv[5]
        database = sys.argv[6] if len(sys.argv) > 6 else "postgres"

        print(f"Connecting to projects/{project_id}/branches/{branch_id}/endpoints/{endpoint_id} (database: {database})...")
        client = AutoscalingLakebaseClient(project_id, branch_id, endpoint_id, profile)
    else:
        instance_name = sys.argv[1]
        profile = sys.argv[2]
        database = sys.argv[3] if len(sys.argv) > 3 else "postgres"

        print(f"Connecting to {instance_name} (database: {database})...")
        client = LakebaseClient(instance_name, profile)

    conn = client.connect(database)

    cur = conn.cursor()
    cur.execute("SELECT version()")
    print(f"Connected! PostgreSQL version: {cur.fetchone()[0]}")

    conn.close()
    print("Connection closed.")
