"""
Generate presentation-quality graphs from the TrafficMinuteStat table.

Outputs two 16×6 @ 200 DPI PNGs:
  presentation_total.png  — total requests per day across all time
  presentation_by_bot.png — stacked area by bot_type across all time

Run build_minute_stats first to populate the source table.

Usage:
    manage.py generate_presentation_graphs
    manage.py generate_presentation_graphs --output-dir /tmp/graphs
"""
import time
from datetime import timezone as dt_tz

from django.core.management.base import BaseCommand
from django.db import connection
from pathlib import Path


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

        # ── Fetch aggregated data ──────────────────────────────────────────────
        self.stdout.write('Querying TrafficMinuteStat (grouped by day + bot_type) ...')
        t0 = time.monotonic()

        with connection.cursor() as cursor:
            cursor.execute("""
                SELECT date_trunc('day', minute) AS day,
                       bot_type,
                       SUM(count)::bigint AS total
                FROM core_trafficminutestat
                GROUP BY day, bot_type
                ORDER BY day, bot_type
            """)
            rows = cursor.fetchall()

        if not rows:
            self.stdout.write(
                'No data in TrafficMinuteStat. '
                'Run: manage.py build_minute_stats --full'
            )
            return

        self.stdout.write(f'  {len(rows):,} day/bot_type rows in {time.monotonic()-t0:.1f}s')

        # ── Build series ───────────────────────────────────────────────────────
        all_days = sorted({row[0] for row in rows})
        day_idx = {d: i for i, d in enumerate(all_days)}
        n = len(all_days)

        all_bots = sorted({row[1] for row in rows})
        series = {bot: [0] * n for bot in all_bots}
        for day, bot_type, total in rows:
            series[bot_type][day_idx[day]] = int(total)

        date_objs = [d.replace(tzinfo=dt_tz.utc) if d.tzinfo is None else d for d in all_days]
        series_thresh = _apply_threshold(series, threshold_pct=1.0)
        grand_total = sum(sum(v) for v in series.values())

        # ── Graph 1: Total traffic ─────────────────────────────────────────────
        self.stdout.write('Generating presentation_total.png ...')
        totals_by_day = [sum(series[bot][i] for bot in all_bots) for i in range(n)]

        fig, ax = plt.subplots(figsize=(16, 6))
        fig.patch.set_facecolor('white')
        ax.set_facecolor('#FAFBFC')
        ax.fill_between(date_objs, totals_by_day, alpha=0.72, color='#4878CF', zorder=2)
        ax.plot(date_objs, totals_by_day, color='#2C5F9E', linewidth=0.8, zorder=3)
        ax.set_ylim(bottom=0)
        ax.yaxis.set_major_formatter(plt.FuncFormatter(lambda v, _: f'{int(v):,}'))
        x_loc = mdates.AutoDateLocator()
        ax.xaxis.set_major_locator(x_loc)
        ax.xaxis.set_major_formatter(mdates.AutoDateFormatter(x_loc))
        ax.set_title(
            'Total Requests Per Day — All Time',
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

        # ── Graph 2: By bot type ───────────────────────────────────────────────
        self.stdout.write('Generating presentation_by_bot.png ...')
        fig, ax = plt.subplots(figsize=(16, 6))
        fig.patch.set_facecolor('white')
        _apply_style(ax, 'Requests By Bot Type Per Day — All Time')
        ax.title.set_fontsize(14)

        groups = list(series_thresh.keys())
        ys = [series_thresh[g] for g in groups]
        colors = _assign_colors(groups)

        ax.stackplot(date_objs, ys, labels=groups, colors=colors, alpha=0.88, zorder=2)
        ax.set_ylim(bottom=0)
        ax.yaxis.set_major_formatter(plt.FuncFormatter(lambda v, _: f'{int(v):,}'))
        ax.set_xlim(date_objs[0], date_objs[-1])
        x_loc2 = mdates.AutoDateLocator()
        ax.xaxis.set_major_locator(x_loc2)
        ax.xaxis.set_major_formatter(mdates.AutoDateFormatter(x_loc2))
        ax.tick_params(axis='x', labelsize=10, colors='#7F8C8D')
        ax.legend(
            loc='upper center', bbox_to_anchor=(0.5, -0.14),
            fontsize=9, framealpha=0.85, ncol=5,
            handlelength=1.2, handletextpad=0.5, columnspacing=1.0,
        )
        fig.tight_layout(pad=1.0)
        _save_atomic(fig, output_dir / 'presentation_by_bot.png', dpi=200)
        plt.close(fig)
        self.stdout.write(
            f'  Saved presentation_by_bot.png  ({len(groups)} series: '
            f'{", ".join(groups[:4])}{"..." if len(groups) > 4 else ""})'
        )

        self.stdout.write(f'\nGraphs written to {output_dir}/')
