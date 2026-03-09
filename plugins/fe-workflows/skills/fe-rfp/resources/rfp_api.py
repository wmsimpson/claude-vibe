#!/usr/bin/env python3
"""
RFP Response Assistant API client.

Calls the Databricks RFP Assistant model serving endpoint to answer
RFP (Request for Proposal) questions with structured compliance responses.

Prerequisites:
- Databricks CLI installed and configured with access to E2 Demo West
- VPN connected

Usage:
    # Single question
    python3 rfp_api.py "Does Databricks support RBAC?"

    # Single question with cloud filter
    python3 rfp_api.py --cloud aws "Does Databricks support RBAC?"

    # Multiple questions
    python3 rfp_api.py --cloud azure "Question 1" "Question 2"

    # Batch from CSV
    python3 rfp_api.py --cloud aws --csv questions.csv --output responses.csv
"""

import argparse
import csv
import json
import subprocess
import sys
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

ENDPOINT_URL = "https://e2-demo-west.cloud.databricks.com/serving-endpoints/rfp-assistant-endpoint/invocations"
DEFAULT_PROFILES = ["e2-demo-west", "DEFAULT"]
REQUEST_TIMEOUT = 150


def get_token(profile: str | None = None) -> str:
    """Get Databricks token, trying specified profile or defaults."""
    profiles = [profile] if profile else DEFAULT_PROFILES
    for p in profiles:
        try:
            result = subprocess.run(
                ["databricks", "auth", "token", "-p", p],
                capture_output=True,
                text=True,
                check=True,
            )
            token_data = json.loads(result.stdout)
            token = token_data.get("access_token", "")
            if token:
                return token
        except (subprocess.CalledProcessError, json.JSONDecodeError):
            continue

    print(
        "Error: Could not get Databricks token. "
        "Run '/databricks-authentication' and configure access to E2 Demo West.",
        file=sys.stderr,
    )
    sys.exit(1)


def call_rfp_endpoint(question: str, cloud: str | None, token: str) -> dict:
    """Call the RFP Assistant endpoint with a single question."""
    payload = {
        "messages": [{"role": "user", "content": question}],
    }
    if cloud:
        payload["custom_inputs"] = {"cloud": cloud.lower()}

    curl_cmd = [
        "curl",
        "-s",
        "-X", "POST",
        ENDPOINT_URL,
        "-H", f"Authorization: Bearer {token}",
        "-H", "Content-Type: application/json",
        "-d", json.dumps(payload),
        "--max-time", str(REQUEST_TIMEOUT),
    ]

    try:
        result = subprocess.run(
            curl_cmd, capture_output=True, text=True, timeout=REQUEST_TIMEOUT + 10
        )
        if result.returncode != 0:
            return {"error": f"curl error: {result.stderr}"}
        return json.loads(result.stdout)
    except subprocess.TimeoutExpired:
        return {"error": f"Request timed out after {REQUEST_TIMEOUT} seconds"}
    except json.JSONDecodeError as e:
        return {"error": f"Error parsing response: {e}", "raw": result.stdout[:500]}


def parse_response(response: dict) -> dict:
    """Parse the RFP endpoint response into a structured result."""
    if "error" in response:
        return {
            "compliance_degree": "Error",
            "justification": response["error"],
            "products": [],
            "sources": [],
            "is_sqrc": False,
        }

    try:
        content_str = response["messages"][0]["content"]
        content = json.loads(content_str)
    except (KeyError, IndexError, json.JSONDecodeError) as e:
        return {
            "compliance_degree": "Error",
            "justification": f"Failed to parse response: {e}",
            "products": [],
            "sources": [],
            "is_sqrc": False,
        }

    compliance = content.get("compliance_degree", "")
    is_sqrc = compliance == "Consult go/sqrc"

    result = {
        "compliance_degree": compliance,
        "justification": content.get("databricks_justification", ""),
        "products": content.get("databricks_product", []),
        "sources": content.get("source_url", []),
        "is_sqrc": is_sqrc,
    }

    # Parse SQRC matches from source_url if applicable
    if is_sqrc and result["sources"]:
        try:
            source_data = result["sources"]
            if isinstance(source_data, str):
                import ast
                sqrc_matches = ast.literal_eval(source_data)
                result["sqrc_matches"] = sqrc_matches
                result["sources"] = [m.get("url", "") for m in sqrc_matches if m.get("url")]
        except (ValueError, SyntaxError):
            pass

    return result


def format_result(question: str, result: dict) -> str:
    """Format a single result for display."""
    lines = []
    lines.append(f"**Question:** {question}")
    lines.append(f"**Compliance:** {result['compliance_degree']}")

    if result["is_sqrc"] and result.get("sqrc_matches"):
        for i, match in enumerate(result["sqrc_matches"], 1):
            lines.append(f"\n**SQRC Match {i}:**")
            lines.append(f"  Q: {match.get('sqrc_question', 'N/A')}")
            short = match.get("short_answer", "")
            if short:
                lines.append(f"  Short Answer: {short}")
            detailed = match.get("detailed_answer", "")
            if detailed:
                lines.append(f"  Detailed Answer: {detailed}")
            url = match.get("url", "")
            if url:
                lines.append(f"  Source: {url}")
        lines.append("\n> Note: Consult go/sqrc for the authoritative answer.")
    else:
        if result["justification"]:
            lines.append(f"\n**Response:** {result['justification']}")
        if result["products"]:
            lines.append(f"\n**Relevant Products:** {', '.join(result['products'])}")
        if result["sources"]:
            lines.append("\n**Sources:**")
            for src in result["sources"]:
                lines.append(f"  - {src}")

    return "\n".join(lines)


def process_batch(questions: list[str], cloud: str | None, token: str, max_workers: int = 2) -> list[dict]:
    """Process multiple questions in parallel."""
    results = [None] * len(questions)

    with ThreadPoolExecutor(max_workers=max_workers) as executor:
        future_to_idx = {
            executor.submit(call_rfp_endpoint, q, cloud, token): i
            for i, q in enumerate(questions)
        }
        for future in as_completed(future_to_idx):
            idx = future_to_idx[future]
            try:
                response = future.result()
                results[idx] = parse_response(response)
            except Exception as e:
                results[idx] = {
                    "compliance_degree": "Error",
                    "justification": str(e),
                    "products": [],
                    "sources": [],
                    "is_sqrc": False,
                }

    return results


def write_csv(questions: list[str], results: list[dict], output_path: str):
    """Write results to a CSV file."""
    with open(output_path, "w", newline="") as f:
        writer = csv.writer(f)
        writer.writerow([
            "question", "compliance_degree", "justification",
            "products", "sources", "is_sqrc",
        ])
        for q, r in zip(questions, results):
            writer.writerow([
                q,
                r["compliance_degree"],
                r["justification"],
                "; ".join(r["products"]) if r["products"] else "",
                "; ".join(r["sources"]) if isinstance(r["sources"], list) else r["sources"],
                r["is_sqrc"],
            ])


def write_markdown(questions: list[str], results: list[dict], output_path: str):
    """Write results to a markdown file suitable for Google Docs conversion."""
    lines = ["# RFP Response Summary\n"]

    # Summary statistics
    total = len(results)
    full = sum(1 for r in results if r["compliance_degree"] == "Full Compliance")
    partial = sum(1 for r in results if r["compliance_degree"] == "Partial Compliance")
    non = sum(1 for r in results if r["compliance_degree"] == "Non-Compliance")
    sqrc = sum(1 for r in results if r["is_sqrc"])
    errors = sum(1 for r in results if r["compliance_degree"] == "Error")

    lines.append("## Summary\n")
    lines.append(f"| Metric | Count |")
    lines.append(f"|--------|-------|")
    lines.append(f"| Total Questions | {total} |")
    lines.append(f"| Full Compliance | {full} |")
    lines.append(f"| Partial Compliance | {partial} |")
    lines.append(f"| Non-Compliance | {non} |")
    lines.append(f"| SQRC (Security/Compliance) | {sqrc} |")
    if errors:
        lines.append(f"| Errors | {errors} |")
    lines.append("")

    lines.append("## Responses\n")
    for i, (q, r) in enumerate(zip(questions, results), 1):
        lines.append(f"### Question {i}\n")
        lines.append(f"**{q}**\n")
        lines.append(f"**Compliance:** {r['compliance_degree']}\n")

        if r["is_sqrc"] and r.get("sqrc_matches"):
            for j, match in enumerate(r["sqrc_matches"], 1):
                lines.append(f"**SQRC Match {j}:**")
                lines.append(f"- Q: {match.get('sqrc_question', 'N/A')}")
                if match.get("short_answer"):
                    lines.append(f"- Answer: {match['short_answer']}")
                if match.get("detailed_answer"):
                    lines.append(f"- Details: {match['detailed_answer']}")
                if match.get("url"):
                    lines.append(f"- [Source]({match['url']})")
                lines.append("")
            lines.append("> Consult go/sqrc for the authoritative answer.\n")
        else:
            if r["justification"]:
                lines.append(f"{r['justification']}\n")
            if r["products"]:
                lines.append(f"**Products:** {', '.join(r['products'])}\n")
            if r["sources"] and isinstance(r["sources"], list):
                lines.append("**Sources:**")
                for src in r["sources"]:
                    lines.append(f"- [{src}]({src})")
                lines.append("")
        lines.append("---\n")

    with open(output_path, "w") as f:
        f.write("\n".join(lines))


def main():
    parser = argparse.ArgumentParser(
        description="Answer RFP questions using the Databricks RFP Response Assistant"
    )
    parser.add_argument(
        "questions",
        nargs="*",
        help="One or more RFP questions to answer",
    )
    parser.add_argument(
        "--cloud",
        choices=["aws", "azure", "gcp"],
        help="Cloud provider to filter documentation (aws, azure, gcp)",
    )
    parser.add_argument(
        "--csv",
        metavar="PATH",
        help="CSV file with a 'question' column for batch processing",
    )
    parser.add_argument(
        "--output",
        metavar="PATH",
        help="Output CSV file path for batch results",
    )
    parser.add_argument(
        "--output-markdown",
        metavar="PATH",
        help="Output markdown file path (for Google Docs conversion)",
    )
    parser.add_argument(
        "--profile",
        metavar="NAME",
        help="Databricks CLI profile name (default: tries e2-demo-west, then DEFAULT)",
    )
    parser.add_argument(
        "--raw",
        action="store_true",
        help="Output raw JSON response",
    )
    parser.add_argument(
        "--workers",
        type=int,
        default=2,
        help="Number of parallel workers for batch processing (default: 2)",
    )

    args = parser.parse_args()

    # Collect questions from args or CSV
    questions = list(args.questions) if args.questions else []

    if args.csv:
        csv_path = Path(args.csv)
        if not csv_path.exists():
            print(f"Error: CSV file not found: {args.csv}", file=sys.stderr)
            sys.exit(1)
        with open(csv_path) as f:
            reader = csv.DictReader(f)
            if "question" not in (reader.fieldnames or []):
                print("Error: CSV must have a 'question' column", file=sys.stderr)
                sys.exit(1)
            for row in reader:
                q = row["question"].strip()
                if q:
                    questions.append(q)

    if not questions:
        parser.print_help()
        sys.exit(1)

    # Authenticate
    token = get_token(args.profile)

    # Process questions
    if len(questions) == 1 and not args.csv:
        # Single question mode
        response = call_rfp_endpoint(questions[0], args.cloud, token)

        if args.raw:
            print(json.dumps(response, indent=2))
        else:
            result = parse_response(response)
            print(format_result(questions[0], result))
    else:
        # Batch mode
        print(f"Processing {len(questions)} questions...", file=sys.stderr)
        results = process_batch(questions, args.cloud, token, max_workers=args.workers)

        if args.raw:
            for q, r in zip(questions, results):
                print(json.dumps({"question": q, "result": r}, indent=2))
        else:
            for q, r in zip(questions, results):
                print(format_result(q, r))
                print("\n" + "=" * 60 + "\n")

        if args.output:
            write_csv(questions, results, args.output)
            print(f"\nResults saved to {args.output}", file=sys.stderr)

        if args.output_markdown:
            write_markdown(questions, results, args.output_markdown)
            print(f"\nMarkdown saved to {args.output_markdown}", file=sys.stderr)


if __name__ == "__main__":
    main()
