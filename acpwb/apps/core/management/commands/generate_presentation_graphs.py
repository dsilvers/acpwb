"""
Generate presentation-quality graphs from the TrafficMinuteStat table.

Outputs two 16×6 @ 200 DPI PNGs:
  presentation_total.png  — total requests per minute (bar chart)
  presentation_by_bot.png — stacked area by bot_type, hourly aggregation

Run build_minute_stats first to populate the source table.

Usage:
    manage.py generate_presentation_graphs
    manage.py generate_presentation_graphs --output-dir /tmp/graphs
"""
import time
from collections import defaultdict
from datetime import timezone as dt_tz

import numpy as np

import matplotlib.dates as mdates_mod

from django.core.management.base import BaseCommand
from django.db import connection
from pathlib import Path

# 1 minute expressed in matplotlib date units (days)
_BAR_WIDTH = 1.0 / 1440


class Command(BaseCommand):
    help = 'Generate presentation-quality traffic graphs from TrafficMinuteStat'

    def add_arguments(self, parser):
        parser.add_argument(
            '--output-dir', default=None,
            help='Output directory (default: staticfiles/graphs/)',
        )

    def handle(self, *args, **options):
        try:
            import matplotlib
            matplotlib.use('Agg')
            import matplotlib.dates as mdates
            import matplotlib.pyplot as plt
        except ImportError:
            self.stdout.write('matplotlib is not installed.')
            return

        from django.conf import settings
        from apps.core.graph_gen import (
            _apply_style, _apply_threshold, _assign_colors, _save_atomic,
        )

        output_dir = Path(options['output_dir'] or settings.STATIC_ROOT / 'graphs')
        output_dir.mkdir(parents=True, exist_ok=True)

        # ── Fetch per-minute data (for total graph) ────────────────────────────
        self.stdout.write('Querying TrafficMinuteStat (per-minute totals) ...')
        t0 = time.monotonic()

        with connection.cursor() as cursor:
            cursor.execute("""
                SELECT minute, SUM(count)::bigint
                FROM core_trafficminutestat
                GROUP BY minute
                ORDER BY minute
            """)
            minute_rows = cursor.fetchall()

        if not minute_rows:
            self.stdout.write(
                'No data in TrafficMinuteStat. '
                'Run: manage.py build_minute_stats --full'
            )
            return

        self.stdout.write(f'  {len(minute_rows):,} minute rows in {time.monotonic()-t0:.1f}s')

        # ── Fetch per-minute data by bot_type (for bot breakdown graph) ────────
        self.stdout.write('Querying TrafficMinuteStat (per-minute by bot_type) ...')
        t1 = time.monotonic()

        with connection.cursor() as cursor:
            cursor.execute("""
                SELECT date_trunc('hour', minute) AS hour, bot_type, SUM(count)::bigint
                FROM core_trafficminutestat
                GROUP BY hour, bot_type
                ORDER BY hour, bot_type
            """)
            hourly_rows = cursor.fetchall()

        self.stdout.write(f'  {len(hourly_rows):,} hour/bot_type rows in {time.monotonic()-t1:.1f}s')

        # ── Build per-minute series for Graph 1 ────────────────────────────────
        all_minutes = [r[0] for r in minute_rows]
        totals = [int(r[1]) for r in minute_rows]
        grand_total = sum(totals)

        minute_objs = [
            m.replace(tzinfo=dt_tz.utc) if m.tzinfo is None else m
            for m in all_minutes
        ]
        minute_nums = mdates_mod.date2num(minute_objs)

        # ── Graph 1: Total traffic per minute (bar chart) ─────────────────────
        self.stdout.write('Generating presentation_total.png ...')

        fig, ax = plt.subplots(figsize=(16, 6))
        fig.patch.set_facecolor('white')
        ax.set_facecolor('#FAFBFC')
        ax.bar(minute_nums, totals, width=_BAR_WIDTH, color='#2E9E8F', alpha=0.85, zorder=2)
        ax.set_xlim(minute_nums[0] - _BAR_WIDTH, minute_nums[-1] + _BAR_WIDTH)
        ax.set_ylim(bottom=0, top=np.percentile(totals, 99.9) * 1.10)
        ax.yaxis.set_major_formatter(plt.FuncFormatter(lambda v, _: f'{int(v):,}'))
        x_loc = mdates.AutoDateLocator()
        ax.xaxis.set_major_locator(x_loc)
        ax.xaxis.set_major_formatter(mdates.AutoDateFormatter(x_loc))
        ax.set_title(
            'Total Requests Per Minute — All Time',
            fontsize=14, color='#2C3E50', pad=10, fontweight='bold',
        )
        ax.tick_params(axis='both', labelsize=10, colors='#7F8C8D')
        ax.grid(axis='y', color='#E8ECEF', linewidth=0.5, zorder=0)
        for spine in ax.spines.values():
            spine.set_edgecolor('#E8ECEF')
        fig.tight_layout(pad=1.0)
        _save_atomic(fig, output_dir / 'presentation_total.png', dpi=200)
        plt.close(fig)
        self.stdout.write(f'  Saved presentation_total.png  ({grand_total:,} total requests)')

        # ── Build per-minute series for Graph 2 ───────────────────────────────
        all_bot_minutes = sorted({r[0] for r in hourly_rows})
        bot_min_idx = {m: i for i, m in enumerate(all_bot_minutes)}
        n = len(all_bot_minutes)

        all_bots = sorted({r[1] for r in hourly_rows})
        series = {bot: [0] * n for bot in all_bots}
        for minute, bot_type, count in hourly_rows:
            series[bot_type][bot_min_idx[minute]] = int(count)

        bot_min_objs = [
            m.replace(tzinfo=dt_tz.utc) if m.tzinfo is None else m
            for m in all_bot_minutes
        ]
        hour_nums = mdates_mod.date2num(bot_min_objs)

        series_thresh = _apply_threshold(series, threshold_pct=1.0)

        # ── Graph 2: By bot type (stacked area, hourly) ───────────────────────
        self.stdout.write('Generating presentation_by_bot.png ...')
        fig, ax = plt.subplots(figsize=(16, 6))
        fig.patch.set_facecolor('white')
        _apply_style(ax, 'Requests By Bot Type Per Hour — All Time')
        ax.title.set_fontsize(14)

        groups = sorted(series_thresh.keys(), key=lambda g: sum(series_thresh[g]))
        ys = [series_thresh[g] for g in groups]
        colors = _assign_colors(groups)

        minute_stack_totals = [sum(s[i] for s in ys) for i in range(n)]
        ax.stackplot(hour_nums, ys, labels=groups, colors=colors, alpha=0.88, zorder=2)
        ax.set_xlim(hour_nums[0], hour_nums[-1])
        ax.set_ylim(bottom=0, top=np.percentile(minute_stack_totals, 99.9) * 1.10)
        ax.yaxis.set_major_formatter(plt.FuncFormatter(lambda v, _: f'{int(v):,}'))
        x_loc2 = mdates.AutoDateLocator()
        ax.xaxis.set_major_locator(x_loc2)
        ax.xaxis.set_major_formatter(mdates.AutoDateFormatter(x_loc2))
        ax.tick_params(axis='x', labelsize=10, colors='#7F8C8D')
        ax.legend(
            loc='upper center', bbox_to_anchor=(0.5, -0.14),
            fontsize=9, framealpha=0.85, ncol=5,
            handlelength=1.2, handletextpad=0.5, columnspacing=1.0,
        )
        fig.subplots_adjust(bottom=0.22, top=0.93, left=0.07, right=0.98)
        _save_atomic(fig, output_dir / 'presentation_by_bot.png', dpi=200)
        plt.close(fig)
        self.stdout.write(
            f'  Saved presentation_by_bot.png  ({len(groups)} series: '
            f'{", ".join(groups[:4])}{"..." if len(groups) > 4 else ""})'
        )

        self.stdout.write(f'\nGraphs written to {output_dir}/')
