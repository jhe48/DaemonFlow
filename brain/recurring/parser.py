"""RRULE parser for converting natural language to RFC 5545 recurrence rules.

This module wraps the recurrent library to parse natural language recurrence
patterns like "every friday" or "daily at 9am" into RFC 5545 RRULE strings
that can be expanded by dateutil.

Note: recurrent is unmaintained (last commit: Aug 2021) but functional.
All calls are wrapped in try/except for graceful degradation.
"""

from datetime import datetime
from typing import Optional

from brain.recurring.patterns import has_recurrence, extract_task_text


def parse_recurrence(text: str, now: Optional[datetime] = None) -> Optional[dict]:
    """Parse natural language text to extract RRULE and clean task text.

    Uses the recurrent library to convert natural language recurrence patterns
    into RFC 5545 RRULE strings. The recurrent library is wrapped with error
    handling to gracefully degrade if parsing fails.

    Args:
        text: The task text to parse for recurrence patterns.
        now: Optional datetime to use as reference for relative dates.
             Defaults to current datetime if not provided.

    Returns:
        A dict with keys:
            - rrule: The RFC 5545 RRULE string (e.g., "RRULE:FREQ=WEEKLY;BYDAY=FR")
            - clean_text: Task text with recurrence words removed
            - original_text: The original input text
        Returns None if text is not recurring or parsing fails.

    Examples:
        >>> result = parse_recurrence("Review PRs every friday")
        >>> result["rrule"]
        'RRULE:FREQ=WEEKLY;BYDAY=FR'
        >>> result["clean_text"]
        'Review PRs'

        >>> parse_recurrence("Buy groceries")
        None
    """
    # Early detection: don't try to parse if no recurrence pattern found
    if not has_recurrence(text):
        return None

    # Try to parse with recurrent library
    try:
        from recurrent.event_parser import RecurringEvent

        reference_date = now or datetime.now()
        r = RecurringEvent(now_date=reference_date)
        r.parse(text)
        rrule_str = r.get_RFC_rrule()

        # recurrent returns empty string or None if parsing fails
        if not rrule_str:
            return None

        return {
            "rrule": rrule_str,
            "clean_text": extract_task_text(text),
            "original_text": text,
        }

    except ImportError:
        # recurrent library not installed
        # This is a critical error that should be logged
        return None

    except Exception:
        # recurrent library failed to parse
        # Graceful degradation - treat as non-recurring
        return None
