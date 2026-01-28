"""RRULE unroller for expanding recurrence rules to concrete dates.

This module uses python-dateutil to expand RFC 5545 RRULE strings into
lists of concrete dates within a given time horizon.

Key functions:
- unroll_occurrences: Expand RRULE to list of ISO date strings
- task_instance_id: Generate stable deterministic ID for deduplication
"""

import hashlib
from datetime import datetime, timedelta
from typing import Optional


def unroll_occurrences(
    rrule_str: str,
    start: Optional[datetime] = None,
    horizon_days: int = 7
) -> list[str]:
    """Expand an RRULE string to a list of ISO 8601 date strings.

    Uses dateutil.rrule to parse and expand the RRULE. Returns dates
    between start and start+horizon_days.

    Args:
        rrule_str: RFC 5545 RRULE string (e.g., "RRULE:FREQ=DAILY").
        start: Start datetime for expansion. Defaults to now.
        horizon_days: Number of days to look ahead. Default 7.

    Returns:
        List of ISO 8601 date strings (YYYY-MM-DD) for occurrences
        within the horizon. Returns empty list if RRULE is invalid
        or has no occurrences in the horizon.

    Examples:
        >>> unroll_occurrences("RRULE:FREQ=DAILY", horizon_days=3)
        ['2026-01-28', '2026-01-29', '2026-01-30']

        >>> unroll_occurrences("INVALID", horizon_days=3)
        []
    """
    if not rrule_str:
        return []

    try:
        from dateutil.rrule import rrulestr

        # Use current time if start not provided
        if start is None:
            start = datetime.now()

        # Calculate horizon end
        end = start + timedelta(days=horizon_days)

        # Parse RRULE - dtstart is required for expansion
        rule = rrulestr(rrule_str, dtstart=start)

        # Get occurrences in range (inc=True to include start boundary)
        occurrences = rule.between(start, end, inc=True)

        # Convert to ISO date strings
        return [occ.strftime("%Y-%m-%d") for occ in occurrences]

    except ImportError:
        # dateutil not installed
        return []

    except Exception:
        # Invalid RRULE or other parsing error
        return []


def task_instance_id(task_text: str, rrule: str, occurrence_date: str) -> str:
    """Generate a stable deterministic ID for a task instance.

    Creates a unique identifier for a specific occurrence of a recurring
    task. The same inputs always produce the same ID, enabling deduplication
    when re-parsing tasks.

    Args:
        task_text: The task text (without recurrence words).
        rrule: The RRULE string for this task.
        occurrence_date: The specific date (YYYY-MM-DD) of this occurrence.

    Returns:
        A 16-character hex string (first 16 chars of SHA256 hash).

    Examples:
        >>> task_instance_id("Review PRs", "RRULE:FREQ=WEEKLY;BYDAY=FR", "2026-01-31")
        'a1b2c3d4e5f67890'
    """
    # Combine all inputs into a single string
    combined = f"{task_text}|{rrule}|{occurrence_date}"

    # Hash with SHA256 and take first 16 chars
    hash_obj = hashlib.sha256(combined.encode("utf-8"))
    return hash_obj.hexdigest()[:16]
