#!/usr/bin/env python3
"""DaemonFlow Brain CLI entry point.

Usage:
    python -m brain --action <action> [--input <json>]

    If --input is not provided, reads JSON from stdin.

Actions:
    ping              Returns health check status with version
    echo              Returns the input unchanged
    parse_recurring   Parse natural language text for recurring patterns
                      Input: {"text": "task text"}
                      Output: {"recurring": true/false, "rrule": "...", ...}
    unroll_recurring  Expand RRULE to list of concrete dates with instance IDs
                      Input: {"rrule": "RRULE:...", "task_text": "...", "horizon_days": 7}
                      Output: {"occurrences": [{"date": "...", "instance_id": "..."}]}

Exit codes:
    0   Success
    1   Error (JSON with {"error": "message"} written to stdout)
"""

import argparse
import json
import sys

from brain import __version__
from brain.recurring.parser import parse_recurrence
from brain.recurring.unroller import unroll_occurrences, task_instance_id


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

    Placeholder for future intelligent actions.
    """
    return data


def parse_recurring_action(data: dict) -> dict:
    """Parse recurring action - detect recurring patterns in task text.

    Parses natural language text to detect recurring patterns like
    "every friday", "daily at 9am", etc. and returns an RRULE string.

    Args:
        data: Dict with "text" key containing the task text to parse.

    Returns:
        Dict with:
            - recurring: True if text contains recurring pattern, False otherwise
            - rrule: RFC 5545 RRULE string (only if recurring is True)
            - clean_text: Task text with recurrence words removed (only if recurring)
            - original_text: The original input text (only if recurring)
    """
    text = data.get("text", "")
    result = parse_recurrence(text)

    if result:
        return {
            "recurring": True,
            "rrule": result["rrule"],
            "clean_text": result["clean_text"],
            "original_text": result["original_text"],
        }
    else:
        return {"recurring": False}


def unroll_recurring_action(data: dict) -> dict:
    """Unroll recurring action - expand RRULE to concrete dates.

    Takes an RRULE string and returns a list of occurrences within
    a time horizon. Each occurrence includes a stable instance ID
    for deduplication purposes.

    Args:
        data: Dict with:
            - rrule: RFC 5545 RRULE string to expand
            - task_text: Original task text (for ID generation)
            - horizon_days: Optional, days to look ahead (default 7)

    Returns:
        Dict with:
            - occurrences: List of dicts, each containing:
                - date: ISO 8601 date string (YYYY-MM-DD)
                - instance_id: Stable 16-char hex ID for deduplication
    """
    rrule = data.get("rrule", "")
    task_text = data.get("task_text", "")
    horizon_days = data.get("horizon_days", 7)

    # Get date strings from unroller
    dates = unroll_occurrences(rrule, horizon_days=horizon_days)

    # Build occurrences with instance IDs
    occurrences = []
    for date in dates:
        instance_id = task_instance_id(task_text, rrule, date)
        occurrences.append({
            "date": date,
            "instance_id": instance_id,
        })

    return {"occurrences": occurrences}


# Action registry
ACTIONS = {
    "ping": ping_action,
    "echo": echo_action,
    "parse_recurring": parse_recurring_action,
    "unroll_recurring": unroll_recurring_action,
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
