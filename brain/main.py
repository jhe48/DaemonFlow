#!/usr/bin/env python3
"""DaemonFlow Brain CLI entry point.

Usage:
    python -m brain --action <action> [--input <json>]

    If --input is not provided, reads JSON from stdin.

Actions:
    ping    Returns health check status with version
    echo    Returns the input unchanged (placeholder for Phase 11)

Exit codes:
    0   Success
    1   Error (JSON with {"error": "message"} written to stdout)
"""

import argparse
import json
import sys

from brain import __version__


def ping_action(data: dict) -> dict:
    """Ping action - returns health check status with version.

    Used by the daemon to verify Python brain is reachable and working.
    """
    return {
        "status": "ok",
        "version": __version__,
    }


def echo_action(data: dict) -> dict:
    """Echo action - returns input unchanged.

    Placeholder for future intelligent actions in Phase 11.
    """
    return data


# Action registry
ACTIONS = {
    "ping": ping_action,
    "echo": echo_action,
}


def main() -> int:
    """Main entry point for the brain CLI."""
    parser = argparse.ArgumentParser(
        description="DaemonFlow Python Brain CLI"
    )
    parser.add_argument(
        "--action",
        required=True,
        choices=list(ACTIONS.keys()),
        help="Action to perform"
    )
    parser.add_argument(
        "--input",
        help="JSON input string (reads from stdin if not provided)"
    )

    args = parser.parse_args()

    # Get input JSON
    try:
        if args.input:
            input_data = json.loads(args.input)
        else:
            input_data = json.load(sys.stdin)
    except json.JSONDecodeError as e:
        print(json.dumps({"error": f"Invalid JSON input: {e}"}))
        return 1
    except Exception as e:
        print(json.dumps({"error": f"Failed to read input: {e}"}))
        return 1

    # Execute action
    action_fn = ACTIONS[args.action]
    try:
        result = action_fn(input_data)
        print(json.dumps(result))
        return 0
    except Exception as e:
        print(json.dumps({"error": f"Action '{args.action}' failed: {e}"}))
        return 1


if __name__ == "__main__":
    sys.exit(main())
