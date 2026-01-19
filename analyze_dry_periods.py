#!/usr/bin/env python3
"""
Analyze rainfall data to find the top 25 longest periods with no rain.
A period with no rain is defined as consecutive days between recorded rainfall events.
"""

import argparse
import json
from datetime import datetime, timedelta
from typing import List, Tuple


def load_rainfall_data(filename: str) -> dict:
    """Load rainfall data from JSON file."""
    with open(filename, 'r') as f:
        return json.load(f)


def find_dry_periods(data: dict, include_to_today: bool = True, min_rainfall: float = 0.0) -> List[Tuple[str, str, int]]:
    """
    Find all periods with no rain (gaps between rainfall events).
    
    Args:
        data: Rainfall data dictionary
        include_to_today: If True, include period from last rain to today
        min_rainfall: Minimum rainfall in mm to count as a rain day (default: 0.0)
    
    Returns:
        List of tuples: (start_date, end_date, days_count)
    """
    # Collect all dates with rainfall and sort them
    all_dates = []
    
    for year, entries in data.items():
        if not entries:  # Skip empty years
            continue
        for entry in entries:
            # Only include dates with rainfall >= min_rainfall
            if entry['rainfall_mm'] >= min_rainfall:
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
    # Parse command line arguments
    parser = argparse.ArgumentParser(
        description='Analyze rainfall data to find longest dry periods',
        formatter_class=argparse.RawDescriptionHelpFormatter
    )
    parser.add_argument(
        '-m', '--mode',
        choices=['a', 'l'],
        default='a',
        help='Mode: "a" for all rain days (default), "l" for limited (only days with >= 2mm)'
    )
    args = parser.parse_args()
    
    # Set minimum rainfall threshold based on mode
    min_rainfall = 2.0 if args.mode == 'l' else 0.0
    mode_description = "all rain days" if args.mode == 'a' else "days with >= 2mm rainfall"
    
    # Load the data
    data = load_rainfall_data('rainfall_data.json')
    
    # Find all dry periods
    dry_periods = find_dry_periods(data, min_rainfall=min_rainfall)
    
    # Sort by duration (descending)
    dry_periods.sort(key=lambda x: x[2], reverse=True)
    
    # Display top 25
    print(f"Top 25 Longest Periods with No Rain (Mode: {mode_description})")
    print("=" * 70)
    print(f"{'Rank':<6} {'Start Date':<12} {'End Date':<12} {'Days':<6}")
    print("-" * 70)
    
    for rank, (start_date, end_date, days) in enumerate(dry_periods[:25], 1):
        print(f"{rank:<6} {start_date:<12} {end_date:<12} {days:<6}")
    
    print("-" * 70)
    
    # Filter dry periods for the past year (from today's date)
    today = datetime.now().replace(hour=0, minute=0, second=0, microsecond=0)
    past_year_start = today - timedelta(days=365)
    past_year_end = today
    
    past_year_dry_periods = []
    for start_date, end_date, days in dry_periods:
        start = datetime.strptime(start_date, '%Y-%m-%d')
        end = datetime.strptime(end_date, '%Y-%m-%d')
        # Include if the dry period overlaps with the past year
        if start <= past_year_end and end >= past_year_start:
            past_year_dry_periods.append((start_date, end_date, days))
    
    # Display top 10 for past year
    print(f"\n\nTop 10 Longest Dry Periods for the Past Year ({past_year_start.strftime('%Y-%m-%d')} to {past_year_end.strftime('%Y-%m-%d')})")
    print("=" * 70)
    print(f"{'Rank':<6} {'Start Date':<12} {'End Date':<12} {'Days':<6}")
    print("-" * 70)
    
    for rank, (start_date, end_date, days) in enumerate(past_year_dry_periods[:10], 1):
        print(f"{rank:<6} {start_date:<12} {end_date:<12} {days:<6}")
    
    print("-" * 70)
    
    # Calculate additional statistics
    all_dates = []
    for year, entries in data.items():
        if entries:
            for entry in entries:
                # Count rain days based on the same minimum rainfall threshold
                if entry['rainfall_mm'] >= min_rainfall:
                    all_dates.append(datetime.strptime(entry['date'], '%Y-%m-%d'))
    
    if all_dates:
        all_dates.sort()
        first_date = all_dates[0]
        today = datetime.now().replace(hour=0, minute=0, second=0, microsecond=0)
        total_days = (today - first_date).days + 1
        total_rain_days = len(all_dates)
        
        print(f"\nTotal days: {total_days}")
        print(f"Total days with rain: {total_rain_days}")
        if len(dry_periods) > 0:
            total_dry_days = sum(period[2] for period in dry_periods)
            print(f"Total days without rain: {total_dry_days}")
        print(f"Total dry periods found: {len(dry_periods)}")
        


if __name__ == "__main__":
    main()
