"""
Databricks Fiscal Calendar utility.

Databricks fiscal year starts February 1.
  FY27 = Feb 1, 2026 → Jan 31, 2027
  FY28 = Feb 1, 2027 → Jan 31, 2028

Month → Fiscal Quarter mapping:
  Feb, Mar, Apr  → Q1  (fiscal_year = calendar_year + 1)
  May, Jun, Jul  → Q2  (fiscal_year = calendar_year + 1)
  Aug, Sep, Oct  → Q3  (fiscal_year = calendar_year + 1)
  Nov, Dec       → Q4  (fiscal_year = calendar_year + 1)
  Jan            → Q4  (fiscal_year = calendar_year)

Quick reference for FY27:
  FY27 Q1: Feb 1 – Apr 30, 2026
  FY27 Q2: May 1 – Jul 31, 2026
  FY27 Q3: Aug 1 – Oct 31, 2026
  FY27 Q4: Nov 1, 2026 – Jan 31, 2027

Usage from a skill:
    # Run directly to get JSON for shell interpolation:
    #   python3 ~/.claude/plugins/cache/fe-vibe/fe-shared-resources/*/resources/fiscal_calendar.py
    # Or import the function directly in Python code:
    #   from resources.fiscal_calendar import databricks_fiscal_context
"""

import json
from datetime import date


def databricks_fiscal_context(run_date: date = None) -> dict:
    """
    Returns fiscal context for any given date (defaults to today).

    Returns a dict with:
        fiscal_year              (int)   e.g. 27 for FY27
        fiscal_quarter           (int)   1-4
        fiscal_year_quarter_str  (str)   e.g. "FY'27 Q1"  — use in account_active_users_quarterly SQL
        quarter_start            (date)  calendar start of current fiscal quarter
        quarter_end              (date)  calendar end   of current fiscal quarter
        prior_fiscal_year        (int)
        prior_fiscal_quarter     (int)
        next_fiscal_year         (int)
        next_fiscal_quarter      (int)

    SQL format notes:
        account_active_users_quarterly  → string filter: fiscal_year_quarter = 'FY''27 Q1'
            Build: fq_str.replace("'", "''")   (SQL-escaped apostrophe)
        rpt_individual_obt_gtm          → integer filter: fiscal_quarter = 1 AND fiscal_year = 27
    """
    d = run_date or date.today()
    m, y = d.month, d.year

    if m == 1:
        fy, fq = y, 4
    elif m in (2, 3, 4):
        fy, fq = y + 1, 1
    elif m in (5, 6, 7):
        fy, fq = y + 1, 2
    elif m in (8, 9, 10):
        fy, fq = y + 1, 3
    else:  # Nov, Dec
        fy, fq = y + 1, 4

    q_starts = {
        1: date(fy - 1, 2, 1),
        2: date(fy - 1, 5, 1),
        3: date(fy - 1, 8, 1),
        4: date(fy - 1, 11, 1),
    }
    q_ends = {
        1: date(fy - 1, 4, 30),
        2: date(fy - 1, 7, 31),
        3: date(fy - 1, 10, 31),
        4: date(fy, 1, 31),
    }

    fy_short = fy % 100
    fq_str = f"FY'{fy_short:02d} Q{fq}"

    prior_fy, prior_fq = (fy - 1, 4) if fq == 1 else (fy, fq - 1)
    next_fy, next_fq = (fy + 1, 1) if fq == 4 else (fy, fq + 1)

    return {
        "fiscal_year": fy,
        "fiscal_quarter": fq,
        "fiscal_year_quarter_str": fq_str,
        "quarter_start": str(q_starts[fq]),
        "quarter_end": str(q_ends[fq]),
        "prior_fiscal_year": prior_fy,
        "prior_fiscal_quarter": prior_fq,
        "next_fiscal_year": next_fy,
        "next_fiscal_quarter": next_fq,
    }


if __name__ == "__main__":
    # Print JSON for shell use:
    #   CTX=$(python3 ~/.claude/plugins/cache/fe-vibe/fe-shared-resources/*/resources/fiscal_calendar.py)
    #   FQ_STR=$(echo "$CTX" | python3 -c "import json,sys; print(json.load(sys.stdin)['fiscal_year_quarter_str'])")
    print(json.dumps(databricks_fiscal_context(), indent=2))
