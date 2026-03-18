#!/bin/bash
#
# Check SKILL.md files for progressive disclosure compliance
# Skills should be 600 lines or less to encourage modular resource files
#

set -e

MAX_LINES=600
CONCISE_THRESHOLD=100
NEAR_LIMIT_THRESHOLD=500

# ANSI color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Check marks and X marks
CHECK="${GREEN}✅${NC}"
WARN="${YELLOW}⚠${NC}"
FAIL="${RED}❌${NC}"

# Track failures
failures=0
declare -a results=()

# Find all SKILL.md files
while IFS= read -r skill_file; do
    # Extract plugin and skill names from path
    # Path format: plugins/<plugin>/skills/<skill>/SKILL.md
    plugin=$(echo "$skill_file" | sed -n 's|plugins/\([^/]*\)/skills/.*|\1|p')
    skill=$(echo "$skill_file" | sed -n 's|plugins/[^/]*/skills/\([^/]*\)/SKILL.md|\1|p')

    if [[ -z "$plugin" || -z "$skill" ]]; then
        continue
    fi

    # Count lines in SKILL.md
    lines=$(wc -l < "$skill_file" | tr -d ' ')

    # Check for resources folder (progressive disclosure)
    skill_dir=$(dirname "$skill_file")
    resource_count=0
    has_resources="no"
    for subdir in resources references scripts assets diagrams; do
        if [[ -d "$skill_dir/$subdir" ]]; then
            has_resources="yes"
            sub_count=$(find "$skill_dir/$subdir" -type f 2>/dev/null | wc -l | tr -d ' ')
            resource_count=$((resource_count + sub_count))
        fi
    done

    # Determine status
    if [[ $lines -le $CONCISE_THRESHOLD ]]; then
        status="Concise"
        status_icon="${CHECK}"
        over_ratio=""
    elif [[ $lines -le $NEAR_LIMIT_THRESHOLD ]]; then
        status="Good"
        status_icon="${CHECK}"
        over_ratio=""
    elif [[ $lines -le $MAX_LINES ]]; then
        if [[ $has_resources == "yes" ]]; then
            status="Best practice"
            status_icon="${CHECK}"
        else
            status="Near limit"
            status_icon="${WARN}"
        fi
        over_ratio=""
    else
        # Over the limit
        ratio=$(echo "scale=1; $lines / $MAX_LINES" | bc)
        over_ratio="${ratio}x over"
        if [[ $has_resources == "yes" ]]; then
            status="Over limit"
            status_icon="${WARN}"
        else
            status="Over limit"
            status_icon="${FAIL}"
            failures=$((failures + 1))
        fi
    fi

    # Format progressive disclosure column
    if [[ $has_resources == "yes" ]]; then
        if [[ $resource_count -gt 1 ]]; then
            prog_disclosure="${CHECK} ${resource_count} refs"
        else
            prog_disclosure="${WARN} ${resource_count} ref"
        fi
    else
        prog_disclosure="${FAIL}"
    fi

    # Store result
    results+=("$plugin|$skill|$lines|$prog_disclosure|$status_icon $status|$over_ratio")

done < <(find plugins -name "SKILL.md" -type f 2>/dev/null | sort)

# Print header
printf "\n"
printf "%-25s %-40s %8s   %-22s   %-20s\n" "Plugin" "Skill" "Lines" "Progressive Disclosure" "Status"
printf "%-25s %-40s %8s   %-22s   %-20s\n" "$(printf '%0.s-' {1..25})" "$(printf '%0.s-' {1..40})" "$(printf '%0.s-' {1..8})" "$(printf '%0.s-' {1..22})" "$(printf '%0.s-' {1..20})"

# Sort results by line count (descending) and print
IFS=$'\n' sorted=($(for r in "${results[@]}"; do
    lines=$(echo "$r" | cut -d'|' -f3)
    echo "$lines|$r"
done | sort -t'|' -k1 -nr | cut -d'|' -f2-))

for result in "${sorted[@]}"; do
    plugin=$(echo "$result" | cut -d'|' -f1)
    skill=$(echo "$result" | cut -d'|' -f2)
    lines=$(echo "$result" | cut -d'|' -f3)
    prog=$(echo "$result" | cut -d'|' -f4)
    status=$(echo "$result" | cut -d'|' -f5)
    over=$(echo "$result" | cut -d'|' -f6)

    # Add over ratio to status if present
    if [[ -n "$over" ]]; then
        status="$status ($over)"
    fi

    printf "%-25s %-40s %8s   %-22b   %-20b\n" "$plugin" "$skill" "$lines" "$prog" "$status"
done

printf "\n"

# Print summary
total=${#results[@]}
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Summary: $total skills checked, $failures over ${MAX_LINES}-line limit without progressive disclosure"
echo ""
echo "Legend:"
echo "  Lines: Number of lines in SKILL.md (limit: ${MAX_LINES})"
printf "  Progressive Disclosure: %b = has resource files, %b = no resource files\n" "$CHECK" "$FAIL"
echo "  Status:"
printf "    %b Concise       - Under %d lines\n" "$CHECK" "$CONCISE_THRESHOLD"
printf "    %b Good          - Under %d lines\n" "$CHECK" "$NEAR_LIMIT_THRESHOLD"
printf "    %b Best practice - Near limit but uses progressive disclosure\n" "$CHECK"
printf "    %b Near limit    - %d-%d lines, consider adding resource files\n" "$WARN" "$NEAR_LIMIT_THRESHOLD" "$MAX_LINES"
printf "    %b Over limit    - Exceeds %d lines without progressive disclosure\n" "$FAIL" "$MAX_LINES"
echo ""

if [[ $failures -gt 0 ]]; then
    echo -e "${YELLOW}WARNING: ${failures} skill(s) exceed ${MAX_LINES} lines without using progressive disclosure.${NC}"
    echo ""
    echo "Consider moving detailed content (examples, schemas, templates) to resources/ subfolder"
    echo "and referencing them from SKILL.md."
else
    echo -e "${GREEN}All skills comply with progressive disclosure guidelines.${NC}"
fi

exit 0
