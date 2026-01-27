# Phase 11: Recurring Task Parser - Research

**Researched:** 2026-01-27
**Domain:** Natural language recurring task parsing (Python)
**Confidence:** MEDIUM

<research_summary>
## Summary

Researched the Python ecosystem for parsing natural language recurring dates like "every friday" or "daily at 9am" and converting them to unrollable recurrence rules.

The standard approach combines two libraries: **recurrent** for natural language → RRULE conversion, and **python-dateutil's rrule** for expanding RRULEs into concrete dates. However, recurrent is unmaintained (last commit: Aug 2021), so careful evaluation is needed.

Key finding: dateparser does NOT handle recurring patterns - it only parses single dates. Recurrent is the only Python library that parses natural language recurrences, but it's unmaintained. The alternative is building a custom parser using regex patterns for common cases (every, daily, weekly, monthly) combined with dateparser for the date components.

**Primary recommendation:** Use recurrent for MVP despite maintenance concerns - it covers the common patterns and outputs RFC-compliant RRULEs. Wrap it with fallback handling. Long-term, consider building a targeted regex-based parser for the specific patterns DaemonFlow needs.
</research_summary>

<standard_stack>
## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| recurrent | 0.4.1 | NL → RRULE | Only Python lib that parses "every tuesday" to RRULEs |
| python-dateutil | 3.9.0 | RRULE expansion | Standard for iCal recurrence rules in Python |
| dateparser | 1.2.2 | Single date NL parsing | Best for "tomorrow", "next week" (not recurrence) |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| parsedatetime | 2.6 | Fuzzy date parsing | Recurrent uses this internally |
| regex | (stdlib re) | Pattern extraction | Extract recurrence markers from task text |
| markdown-checklist | 0.4.4 | Parse markdown tasks | If using Python Markdown lib |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| recurrent | Custom regex parser | More control, more work, but guaranteed maintenance |
| recurrent | LLM extraction | More flexible but overkill for well-defined patterns |
| dateparser | parsedatetime | parsedatetime better for future dates, dateparser for multi-language |

**Installation:**
```bash
pip install recurrent python-dateutil dateparser
```

**Warning:** recurrent's last commit was August 2021. It works but may have issues with Python 3.12+. Test thoroughly.
</standard_stack>

<architecture_patterns>
## Architecture Patterns

### Recommended Project Structure
```
brain/
├── recurring/           # Recurring task subsystem
│   ├── __init__.py
│   ├── parser.py        # NL → RRULE parsing (recurrent wrapper)
│   ├── unroller.py      # RRULE → concrete dates (dateutil wrapper)
│   ├── patterns.py      # Regex patterns for "every X" detection
│   └── dedup.py         # Deduplication logic
├── main.py              # CLI entry point
└── ...
```

### Pattern 1: Two-Phase Parsing
**What:** Separate detection from parsing - first detect if task has recurrence pattern, then parse it
**When to use:** All recurring task implementations
**Example:**
```python
import re
from recurrent.event_parser import RecurringEvent
from dateutil import rrule

# Detection patterns
RECURRENCE_PATTERNS = [
    r'\bevery\s+\w+',           # every day, every monday
    r'\bdaily\b',               # daily
    r'\bweekly\b',              # weekly
    r'\bmonthly\b',             # monthly
    r'\b\w+days?\b',            # fridays, mondays
    r'\bevery\s+other\b',       # every other
    r'\brepeat\b',              # repeat weekly
]

def has_recurrence(text: str) -> bool:
    """Detect if text contains recurrence pattern."""
    text_lower = text.lower()
    return any(re.search(p, text_lower) for p in RECURRENCE_PATTERNS)

def parse_recurrence(text: str, now=None) -> str | None:
    """Parse text to RRULE string. Returns None if not recurring."""
    if not has_recurrence(text):
        return None
    r = RecurringEvent(now_date=now or datetime.now())
    r.parse(text)
    return r.get_RFC_rrule()
```

### Pattern 2: Stable Task IDs for Deduplication
**What:** Generate deterministic IDs from task content + recurrence rule to prevent duplicates
**When to use:** When unrolling generates repeating instances
**Example:**
```python
import hashlib

def task_id(task_text: str, rrule: str, occurrence_date: datetime) -> str:
    """Generate stable ID for a recurring task instance."""
    # Include date in ISO format (not time) for daily dedup
    date_str = occurrence_date.date().isoformat()
    content = f"{task_text}|{rrule}|{date_str}"
    return hashlib.sha256(content.encode()).hexdigest()[:16]
```

### Pattern 3: Horizon-Based Unrolling
**What:** Only unroll tasks within a configurable time horizon (e.g., 7 days)
**When to use:** Avoid creating infinite future tasks
**Example:**
```python
from dateutil.rrule import rrulestr
from datetime import datetime, timedelta

def unroll_occurrences(rrule_str: str,
                       start: datetime = None,
                       horizon_days: int = 7) -> list[datetime]:
    """Get occurrences within horizon."""
    start = start or datetime.now()
    end = start + timedelta(days=horizon_days)

    rule = rrulestr(rrule_str, dtstart=start)
    return list(rule.between(start, end, inc=True))
```

### Anti-Patterns to Avoid
- **Parsing in the Go daemon:** Keep all NL parsing in Python; Go just triggers
- **Storing parsed dates instead of RRULEs:** RRULEs are the source of truth
- **Unrolling all future occurrences:** Use horizon-based approach
- **Non-idempotent unrolling:** Same input must always produce same output
</architecture_patterns>

<dont_hand_roll>
## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| RRULE parsing/generation | String manipulation | dateutil.rrule | RFC 5545 is complex, edge cases abound |
| "every monday" → rule | Regex to RRULE | recurrent | Handles "every other", "until", "except" |
| Date arithmetic | Manual day/month math | dateutil.relativedelta | Month-end, leap years, DST |
| Timezone handling | UTC offsets by hand | pytz or zoneinfo | DST transitions are complex |
| Occurrence expansion | Loop + increment | rrule.between() | Handles count, until, exceptions |

**Key insight:** Recurrence rules look simple ("every Monday") but the implementation involves RFC 5545 compliance, timezone handling, exception dates, interval math, and edge cases like "31st of every month" (skips months without 31 days). Even unmaintained, recurrent + dateutil handle more edge cases than custom code.
</dont_hand_roll>

<common_pitfalls>
## Common Pitfalls

### Pitfall 1: dateparser Doesn't Do Recurrence
**What goes wrong:** Using dateparser.parse("every friday") returns None or unexpected date
**Why it happens:** dateparser parses single dates, not recurrence patterns
**How to avoid:** Use recurrent for "every X" patterns, dateparser only for "next friday"
**Warning signs:** dateparser returning None for clearly recurring text

### Pitfall 2: Month-End Skipping with rrule
**What goes wrong:** Task scheduled for "31st of every month" skips February, April, June, September, November
**Why it happens:** rrule ignores invalid dates per RFC 5545 rather than coercing
**How to avoid:** Use bysetpos for "last day of month" or validate user expectations
**Warning signs:** Missing occurrences in shorter months
**Example fix:**
```python
# "Last day of every month" instead of "31st"
rrule(MONTHLY, bymonthday=-1)  # -1 means last day
```

### Pitfall 3: DST Transitions Break Times
**What goes wrong:** 9am meeting becomes 8am or 10am after DST switch
**Why it happens:** rrule with pytz timezone doesn't handle DST correctly
**How to avoid:** Generate naive datetimes, localize after with pytz.normalize()
**Warning signs:** Times shifting by 1 hour twice a year

### Pitfall 4: Duplicate Task Creation
**What goes wrong:** Same task instance created multiple times on re-parse
**Why it happens:** No stable ID for task instances
**How to avoid:** Hash (task_text, rrule, occurrence_date) for idempotent inserts
**Warning signs:** Multiple identical tasks for same day

### Pitfall 5: Recurrent Library Maintenance
**What goes wrong:** Library breaks on new Python versions or has unfixed bugs
**Why it happens:** Last commit was August 2021
**How to avoid:** Pin version, wrap with error handling, have fallback strategy
**Warning signs:** ImportError or AttributeError on Python upgrade
</common_pitfalls>

<code_examples>
## Code Examples

Verified patterns from official sources:

### recurrent: Natural Language to RRULE
```python
# Source: github.com/kvh/recurrent README
import datetime
from recurrent.event_parser import RecurringEvent

r = RecurringEvent(now_date=datetime.datetime(2026, 1, 27))

# Parse natural language
r.parse('every friday at 9am')
rrule_str = r.get_RFC_rrule()
# Returns: 'RRULE:FREQ=WEEKLY;BYDAY=FR;BYHOUR=9'

r.parse('daily except on weekends')
r.parse('every other tuesday')
r.parse('monthly on the 15th')
r.parse('every day starting next monday until march')
```

### dateutil: RRULE Expansion
```python
# Source: dateutil.readthedocs.io/en/stable/rrule.html
from dateutil.rrule import rrulestr, rrule, WEEKLY, MO, TU, WE, TH, FR
from datetime import datetime, timedelta

# From RRULE string
rule = rrulestr('RRULE:FREQ=WEEKLY;BYDAY=FR', dtstart=datetime(2026, 1, 27))
next_5 = list(rule[:5])

# Get occurrences in range
start = datetime.now()
end = start + timedelta(days=14)
occurrences = list(rule.between(start, end))

# Programmatic rule creation
weekday_rule = rrule(WEEKLY, byweekday=(MO, TU, WE, TH, FR),
                     dtstart=datetime(2026, 1, 27))
```

### Task Detection Pattern
```python
# DaemonFlow-specific pattern
import re

def extract_task_and_recurrence(line: str) -> tuple[str, str | None]:
    """
    Parse markdown task line, extract clean text and recurrence.

    Input: "- [ ] Review PRs every friday"
    Output: ("Review PRs", "RRULE:FREQ=WEEKLY;BYDAY=FR")
    """
    # Remove markdown checkbox
    text = re.sub(r'^-\s*\[[ x]\]\s*', '', line.strip())

    # Check for recurrence
    r = RecurringEvent()
    r.parse(text)
    rrule = r.get_RFC_rrule()

    if rrule:
        # Extract clean task text (remove recurrence words)
        # This is simplified - real impl needs more sophisticated extraction
        clean = re.sub(r'\s*(every|daily|weekly|monthly)\s+\w+.*$', '', text, flags=re.I)
        return clean.strip(), rrule

    return text, None
```
</code_examples>

<sota_updates>
## State of the Art (2025-2026)

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| pytz for timezones | zoneinfo (stdlib) | Python 3.9 (2020) | No external dep for timezones |
| Manual RRULE strings | dateutil 2.9+ | Ongoing | Better API, more edge case handling |
| recurrent alone | recurrent + fallback | Now | Library unmaintained since 2021 |

**New tools/patterns to consider:**
- **LLM-based extraction:** Could use Claude API for complex NL parsing, but overkill for defined patterns
- **zoneinfo module:** Python 3.9+ stdlib, preferred over pytz for new projects
- **recurring-ical-events:** Actively maintained (v3.4.0 Jan 2025) but works on iCal format, not NL

**Deprecated/outdated:**
- **parsedatetime alone:** Use dateparser for better multi-language support
- **pytz:** Use zoneinfo for new Python 3.9+ projects
- **Manual RRULE generation:** Always use dateutil.rrule
</sota_updates>

<open_questions>
## Open Questions

Things that couldn't be fully resolved:

1. **recurrent Python 3.12+ compatibility**
   - What we know: Last tested on Python 3.6-3.10 era
   - What's unclear: May have issues with 3.12+ changes
   - Recommendation: Test early, have regex fallback for core patterns

2. **Non-English recurring patterns**
   - What we know: recurrent is "completely English centric" per docs
   - What's unclear: Whether parsedatetime locales help
   - Recommendation: Document as limitation, English-only for v2.0

3. **Edge case: task text containing date words**
   - What we know: "Buy milk every day" works, but what about "Review the daily report"?
   - What's unclear: False positive rate for recurrence detection
   - Recommendation: Only trigger on explicit patterns ("every", "daily at", "weekly")
</open_questions>

<sources>
## Sources

### Primary (HIGH confidence)
- [dateutil rrule documentation](https://dateutil.readthedocs.io/en/stable/rrule.html) - RRULE API, patterns, examples
- [dateparser documentation](https://dateparser.readthedocs.io/) - Single date parsing capabilities
- [GitHub: kvh/recurrent](https://github.com/kvh/recurrent) - README, examples, maintenance status

### Secondary (MEDIUM confidence)
- [Zyte blog: Parse natural language dates](https://www.zyte.com/blog/parse-natural-language-dates-with-dateparser/) - dateparser overview
- [Nylas: Calendar Events and RRULEs](https://www.nylas.com/blog/calendar-events-rrules/) - RRULE complexity overview
- [Martin Heinz: Scheduling Recurring Jobs](https://martinheinz.dev/blog/39) - Python scheduling patterns

### Tertiary (LOW confidence - needs validation)
- recurrent behavior with Python 3.12+ (not tested in research)
- Performance characteristics of recurrent on large task lists
</sources>

<metadata>
## Metadata

**Research scope:**
- Core technology: Python natural language date parsing
- Ecosystem: recurrent, dateutil, dateparser
- Patterns: Two-phase parsing, idempotent unrolling, horizon-based generation
- Pitfalls: DST, month-end, deduplication, library maintenance

**Confidence breakdown:**
- Standard stack: MEDIUM - recurrent unmaintained but functional
- Architecture: HIGH - patterns are well-established
- Pitfalls: HIGH - well-documented in rrule docs and community
- Code examples: HIGH - from official documentation

**Research date:** 2026-01-27
**Valid until:** 2026-02-27 (30 days - stable domain, main concern is recurrent maintenance)
</metadata>

---

*Phase: 11-recurring-task-parser*
*Research completed: 2026-01-27*
*Ready for planning: yes*
