"""Recurrence pattern detection for natural language task parsing.

This module provides regex-based detection of recurring patterns in task text,
such as "every friday", "daily", "weekly", etc.

The detection is intentionally conservative to avoid false positives on words
like "everyday" used as an adjective rather than a recurrence pattern.
"""

import re
from typing import List

# Patterns that indicate a recurring task
# These are designed to match explicit recurrence markers, not ambiguous uses
RECURRENCE_PATTERNS: List[str] = [
    r'\bevery\s+\w+',           # every day, every monday, every week
    r'\bdaily\b',               # daily
    r'\bweekly\b',              # weekly
    r'\bmonthly\b',             # monthly
    r'\byearly\b',              # yearly
    r'\bannually\b',            # annually
    r'\b(mon|tues|wednes|thurs|fri|satur|sun)days?\b',  # mondays, fridays, etc.
    r'\bevery\s+other\b',       # every other
]

# Compile patterns for efficiency
_COMPILED_PATTERNS = [re.compile(p, re.IGNORECASE) for p in RECURRENCE_PATTERNS]

# Words that indicate recurrence, used for extracting clean task text
_RECURRENCE_WORDS = [
    r'\bevery\s+(\w+\s+)*\w+',  # "every friday", "every other week"
    r'\bdaily\b',
    r'\bweekly\b',
    r'\bmonthly\b',
    r'\byearly\b',
    r'\bannually\b',
    r'\bat\s+\d{1,2}(:\d{2})?\s*(am|pm)?\b',  # "at 9am", "at 9:30 pm"
]


def has_recurrence(text: str) -> bool:
    """Check if text contains a recurrence pattern.

    Args:
        text: The task text to check for recurrence patterns.

    Returns:
        True if the text contains any recognized recurrence pattern,
        False otherwise.

    Examples:
        >>> has_recurrence("Review PRs every friday")
        True
        >>> has_recurrence("daily standup at 9am")
        True
        >>> has_recurrence("Buy groceries")
        False
        >>> has_recurrence("This is an everyday task")
        False
    """
    return any(pattern.search(text) for pattern in _COMPILED_PATTERNS)


def extract_task_text(text: str) -> str:
    """Extract clean task text by removing recurrence words.

    Removes recurrence patterns like "every friday", "daily", "at 9am"
    from the task text to get the core task description.

    Args:
        text: The full task text including recurrence patterns.

    Returns:
        The task text with recurrence patterns removed and cleaned up.

    Examples:
        >>> extract_task_text("Review PRs every friday")
        'Review PRs'
        >>> extract_task_text("daily standup at 9am")
        'standup'
        >>> extract_task_text("Buy groceries")
        'Buy groceries'
    """
    result = text

    # Remove recurrence patterns
    for pattern in _RECURRENCE_WORDS:
        result = re.sub(pattern, '', result, flags=re.IGNORECASE)

    # Clean up extra whitespace
    result = re.sub(r'\s+', ' ', result).strip()

    return result
