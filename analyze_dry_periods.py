#!/usr/bin/env python3
"""
Analyze rainfall data to find the top 25 longest periods with no rain.
A period with no rain is defined as consecutive days between recorded rainfall events.
"""

import json
from datetime import datetime, timedelta
from typing import List, Tuple


def load_rainfall_data(filename: str) -> dict:
    """Load rainfall data from JSON file."""
    with open(filename, 'r') as f:
        return json.load(f)


def find_dry_periods(data: dict, include_to_today: bool = True) -> List[Tuple[str, str, int]]:
    """
    Find all periods with no rain (gaps between rainfall events).
    
    Args:
        data: Rainfall data dictionary
        include_to_today: If True, include period from last rain to today
    
    Returns:
        List of tuples: (start_date, end_date, days_count)
    """
    # Collect all dates with rainfall and sort them
    all_dates = []
    
    for year, entries in data.items():
        if not entries:  # Skip empty years
            continue
        for entry in entries:
            date_str = entry['date']
            all_dates.append(datetime.strptime(date_str, '%Y-%m-%d'))
    
    # Sort dates chronologically
    all_dates.sort()
    
    # Find gaps between consecutive rainfall events
    dry_periods = []
    
    for i in range(len(all_dates) - 1):
        current_date = all_dates[i]
        next_date = all_dates[i + 1]
        
        # Calculate the gap (days between rain events, excluding the rain days themselves)
        gap_days = (next_date - current_date).days - 1
        
        if gap_days > 0:
            # Start date is the day after current rain
            start_date = current_date + timedelta(days=1)
            # End date is the day before next rain
            end_date = next_date - timedelta(days=1)
            
            dry_periods.append((
                start_date.strftime('%Y-%m-%d'),
                end_date.strftime('%Y-%m-%d'),
                gap_days
            ))
    
    # Add period from last rainfall to today
    if include_to_today and all_dates:
        last_rain = all_dates[-1]
        today = datetime.now().replace(hour=0, minute=0, second=0, microsecond=0)
        
        # Only add if today is after the last rain
        if today > last_rain:
            gap_days = (today - last_rain).days - 1
            if gap_days > 0:
                start_date = last_rain + timedelta(days=1)
                dry_periods.append((
                    start_date.strftime('%Y-%m-%d'),
                    today.strftime('%Y-%m-%d'),
                    gap_days
                ))
    
    return dry_periods


def main():
    # Load the data
    data = load_rainfall_data('rainfall_data.json')
    
    # Find all dry periods
    dry_periods = find_dry_periods(data)
    
    # Sort by duration (descending)
    dry_periods.sort(key=lambda x: x[2], reverse=True)
    
    # Display top 25
    print("Top 25 Longest Periods with No Rain")
    print("=" * 70)
    print(f"{'Rank':<6} {'Start Date':<12} {'End Date':<12} {'Days':<6}")
    print("-" * 70)
    
    for rank, (start_date, end_date, days) in enumerate(dry_periods[:40], 1):
        print(f"{rank:<6} {start_date:<12} {end_date:<12} {days:<6}")
    
    print("-" * 70)
    print(f"\nTotal dry periods found: {len(dry_periods)}")
    
    if len(dry_periods) > 0:
        total_dry_days = sum(period[2] for period in dry_periods)
        print(f"Total days without rain: {total_dry_days}")


if __name__ == "__main__":
    main()
