#!/usr/bin/env python3
"""
SailC Multipass VM Agent Worker.
Autonomous agent running inside an Ubuntu Multipass VM,
performing proactive context queries, status reporting, and resolution sharing.
"""

import argparse
import json
import time
import urllib.request
import urllib.error

def call_mcp_tool(hub_url: str, tool_name: str, arguments: dict):
    payload = {
        "jsonrpc": "2.0",
        "id": int(time.time() * 1000),
        "method": "tools/call",
        "params": {
            "name": tool_name,
            "arguments": arguments
        }
    }
    req = urllib.request.Request(
        f"{hub_url}/mcp",
        data=json.dumps(payload).encode("utf-8"),
        headers={"Content-Type": "application/json"}
    )
    try:
        with urllib.request.urlopen(req, timeout=5) as resp:
            data = json.loads(resp.read().decode("utf-8"))
            return data.get("result", {})
    except Exception as e:
        print(f"[!] Communication error with SailC Hub ({hub_url}): {e}")
        return None

def main():
    parser = argparse.ArgumentParser(description="SailC VM Agent Worker")
    parser.add_argument("--hub", default="http://127.0.0.1:8765", help="SailC Central Hub URL")
    parser.add_argument("--name", default="Multipass-Worker-1", help="AI agent name")
    args = parser.parse_args()

    print(f"=== Starting Agent '{args.name}' ===")
    print(f"[*] Connecting to Central Hub: {args.hub}")

    # 1. Report live status
    res = call_mcp_tool(args.hub, "agent_report_status", {
        "agent_id": args.name,
        "task_description": "Auditing open tasks in central hub",
        "status": "working"
    })
    if res:
        print("[+] Telemetry successfully posted to SailC monitor!")

    # 2. Search before asking
    print("\n[*] Step 1: Proactive knowledge search before asking (hub_search)...")
    search_res = call_mcp_tool(args.hub, "hub_search", {"query": "Go"})
    print(f"[+] Knowledge search result:\n{json.dumps(search_res, indent=2, ensure_ascii=False)}")

    # 3. Publish new inquiry
    print("\n[*] Step 2: Creating new inquiry in shared log (hub_ask)...")
    ask_res = call_mcp_tool(args.hub, "hub_ask", {
        "author": args.name,
        "title": "Multipass Worker Coordination & Distributed Cache",
        "content": "Worker active inside isolated Ubuntu environment. Should we configure distributed caching for context passing?",
        "tags": ["multipass", "vm", "cache"]
    })
    print(f"[+] Inquiry created:\n{json.dumps(ask_res, indent=2, ensure_ascii=False)}")

    # 4. Set idle status
    call_mcp_tool(args.hub, "agent_report_status", {
        "agent_id": args.name,
        "task_description": "Awaiting responses from host AI agents",
        "status": "online"
    })
    print("\n[+] Initialization complete. Agent active in SailC Hub.")

if __name__ == "__main__":
    main()
