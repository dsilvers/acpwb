"""
Traffic graph generation for the ACPWB dashboard.

Generates 5 PNG stacked-area charts saved to staticfiles/graphs/:
  traffic_1h.png   — last hour,    per-minute buckets     (live query)
  traffic_8h.png   — last 8 hours, per-10-minute buckets  (live query)
  traffic_24h.png  — last 24 hrs,  per-hour buckets       (live query)
  traffic_7d.png   — last 7 days,  per-day buckets        (stored DashboardStat)
  traffic_all.png  — all time,     per-day buckets        (stored DashboardStat)

Live queries (1h/8h/24h) are timestamp-range-bounded — they touch only the
recent slice of the table and use the (bot_group, timestamp) composite index.
At any reasonable traffic rate the scanned row count is small regardless of
the total table size.

Long-window graphs (7d/all) read from pre-aggregated DashboardStat rows
(crawlers.daily, archive.daily) written by precalc_dashboard — no table scan.

Requires matplotlib (pip install matplotlib).
"""
import os
import tempfile
import time as _time
from collections import defaultdict
from datetime import datetime, timedelta
from pathlib import Path

try:
    import matplotlib
    matplotlib.use('Agg')
    import matplotlib.dates as mdates
    import matplotlib.pyplot as plt
    _MATPLOTLIB_AVAILABLE = True
except ImportError:
    _MATPLOTLIB_AVAILABLE = False

# ── Colour palette ─────────────────────────────────────────────────────────────

GROUP_COLORS = {
    'AI Crawlers':        '#E74C3C',
    'Search Engines':     '#2980B9',
    'SEO / Other Bots':   '#8E44AD',
    'Generic Scrapers':   '#27AE60',
    'Other / Browser':    '#7F8C8D',
    '(empty user agent)': '#95A5A6',
    'Archive Visits':     '#16A085',
    # stored-stat series names (7d / all-time)
    'Crawler Hits':       '#C9A84C',
}

# Stacking order — archive/empty at bottom, AI at top
GROUP_STACK_ORDER = [
    'Archive Visits',
    '(empty user agent)',
    'Other / Browser',
    'Generic Scrapers',
    'SEO / Other Bots',
    'Search Engines',
    'AI Crawlers',
    # stored-stat fallback
    'Crawler Hits',
]

# ── Bucket helpers ─────────────────────────────────────────────────────────────

def _floor_to_interval(dt, interval_minutes):
    """Floor a timezone-aware datetime to the nearest interval_minutes boundary."""
    interval_seconds = interval_minutes * 60
    ts = dt.timestamp()
    floored_ts = (ts // interval_seconds) * interval_seconds
    return datetime.fromtimestamp(floored_ts, tz=dt.tzinfo)


def _make_buckets(start, end, interval_minutes):
    """All bucket boundaries from start to end at interval_minutes spacing."""
    current = _floor_to_interval(start, interval_minutes)
    buckets = []
    while current <= end:
        buckets.append(current)
        current += timedelta(minutes=interval_minutes)
    return buckets


# ── DB queries (live — recent windows only) ────────────────────────────────────

def _query_windowed(cutoff, interval_minutes):
    """
    Query CrawlerVisit and ArchiveVisit since cutoff using timestamp index.

    Safe for large tables: the WHERE timestamp >= cutoff clause is satisfied
    via an index range scan on the recent fraction of rows only.

    Returns (buckets, series) where:
      buckets — sorted list of datetime objects covering [cutoff, now]
      series  — dict[group_name → list[int]] aligned to buckets
    """
    from django.db.models import Count
    from django.db.models.functions import TruncHour, TruncMinute
    from django.utils import timezone

    from apps.honeypot.models import ArchiveVisit, CrawlerVisit

    now = timezone.now()
    trunc_fn = TruncMinute if interval_minutes < 60 else TruncHour
    buckets = _make_buckets(cutoff, now, interval_minutes)

    # CrawlerVisit — stacked by bot_group (field is pre-denormalized, fast GROUP BY)
    cv_rows = (
        CrawlerVisit.objects
        .filter(timestamp__gte=cutoff)
        .annotate(bucket=trunc_fn('timestamp'))
        .values('bucket', 'bot_group')
        .annotate(c=Count('id'))
    )

    data = defaultdict(lambda: defaultdict(int))
    for row in cv_rows:
        b = _floor_to_interval(row['bucket'], interval_minutes)
        g = row['bot_group'] or '(empty user agent)'
        data[b][g] += row['c']

    # ArchiveVisit — no bot_group field; shown as its own series
    av_rows = (
        ArchiveVisit.objects
        .filter(timestamp__gte=cutoff)
        .annotate(bucket=trunc_fn('timestamp'))
        .values('bucket')
        .annotate(c=Count('id'))
    )
    for row in av_rows:
        b = _floor_to_interval(row['bucket'], interval_minutes)
        data[b]['Archive Visits'] += row['c']

    present = set()
    for b in data:
        present.update(data[b].keys())

    groups = [g for g in GROUP_STACK_ORDER if g in present]
    series = {g: [data[b].get(g, 0) for b in buckets] for g in groups}

    return buckets, series


# ── Stored-stat queries (no table scan) ────────────────────────────────────────

def _query_stored_daily(cutoff=None):
    """
    Read per-day totals from DashboardStat rows written by precalc_dashboard.

    Returns (dates, series) with 'Crawler Hits' and 'Archive Visits' keys.
    cutoff — if provided, only return dates >= cutoff.
    """
    from django.utils import timezone

    from apps.core.models import DashboardStat

    crawlers_daily = {}
    archive_daily = {}
    try:
        crawlers_daily = DashboardStat.objects.get(key='crawlers.daily').value
    except DashboardStat.DoesNotExist:
        pass
    try:
        archive_daily = DashboardStat.objects.get(key='archive.daily').value
    except DashboardStat.DoesNotExist:
        pass

    all_dates = sorted(set(crawlers_daily) | set(archive_daily))
    if cutoff:
        cutoff_str = cutoff.date().isoformat()
        all_dates = [d for d in all_dates if d >= cutoff_str]

    if not all_dates:
        return [], {}

    date_objs = [
        datetime.fromisoformat(d).replace(tzinfo=timezone.utc)
        for d in all_dates
    ]
    series = {
        'Crawler Hits':   [crawlers_daily.get(d, 0) for d in all_dates],
        'Archive Visits': [archive_daily.get(d, 0)  for d in all_dates],
    }
    return date_objs, series


# ── Rendering helpers ──────────────────────────────────────────────────────────

def _apply_style(ax, title):
    """Apply consistent visual style to an axes object."""
    ax.set_facecolor('#FAFBFC')
    ax.set_title(title, fontsize=8.5, color='#2C3E50', pad=5, fontweight='bold')
    ax.tick_params(axis='both', labelsize=6.5, colors='#7F8C8D')
    ax.grid(axis='y', color='#E8ECEF', linewidth=0.5, zorder=0)
    for spine in ax.spines.values():
        spine.set_edgecolor('#E8ECEF')
    ax.set_ylim(bottom=0)


def _render_stacked(ax, xs, series, title, x_locator, x_fmt):
    """
    Draw a stacked area chart.

    x_fmt may be a strftime format string or a matplotlib Formatter object.
    """
    _apply_style(ax, title)
    if not xs or not series:
        ax.text(0.5, 0.5, 'No data', ha='center', va='center',
                transform=ax.transAxes, color='#95A5A6', fontsize=9)
        return

    groups = [g for g in GROUP_STACK_ORDER if g in series]
    ys     = [series[g] for g in groups]
    colors = [GROUP_COLORS.get(g, '#95A5A6') for g in groups]

    ax.stackplot(xs, ys, labels=groups, colors=colors, alpha=0.88, zorder=2)
    ax.set_xlim(xs[0], xs[-1])
    ax.xaxis.set_major_locator(x_locator)
    if isinstance(x_fmt, str):
        ax.xaxis.set_major_formatter(mdates.DateFormatter(x_fmt))
    else:
        ax.xaxis.set_major_formatter(x_fmt)
    ax.legend(loc='upper left', fontsize=6, framealpha=0.75,
              ncol=4, handlelength=1, handletextpad=0.4, columnspacing=0.8)


# ── File I/O ───────────────────────────────────────────────────────────────────

def _save_atomic(fig, path):
    """Write to a temp file then atomically rename — nginx never sees a partial PNG."""
    path = Path(path)
    fd, tmp = tempfile.mkstemp(suffix='.png', dir=path.parent)
    try:
        os.close(fd)
        fig.savefig(tmp, dpi=100, bbox_inches='tight',
                    facecolor=fig.get_facecolor())
        os.replace(tmp, path)
    except Exception:
        try:
            os.unlink(tmp)
        except OSError:
            pass
        raise


# ── Public entry point ─────────────────────────────────────────────────────────

def generate_traffic_graphs(output_dir, stdout=None):
    """
    Generate 5 PNG traffic charts and save to output_dir.

    output_dir — pathlib.Path or str; created if absent.
    stdout     — optional file-like for progress messages.
    """
    if not _MATPLOTLIB_AVAILABLE:
        if stdout:
            stdout.write('  graphs: matplotlib not installed — skipping\n')
        return

    from django.utils import timezone

    output_dir = Path(output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    now = timezone.now()

    # Live-query windows (timestamp-range-bounded, safe on large tables)
    live_windows = [
        (
            '1h',
            now - timedelta(hours=1),
            1,
            '%H:%M',
            mdates.MinuteLocator(byminute=[0, 15, 30, 45]),
            'Last Hour (per minute)',
        ),
        (
            '8h',
            now - timedelta(hours=8),
            10,
            '%H:%M',
            mdates.HourLocator(),
            'Last 8 Hours (per 10 min)',
        ),
        (
            '24h',
            now - timedelta(hours=24),
            60,
            '%H:%M',
            mdates.HourLocator(byhour=[0, 4, 8, 12, 16, 20]),
            'Last 24 Hours (per hour)',
        ),
    ]

    for name, cutoff, interval_min, xfmt, x_locator, title in live_windows:
        t0 = _time.monotonic()
        try:
            buckets, series = _query_windowed(cutoff, interval_min)
        except Exception as exc:
            if stdout:
                stdout.write(f'    traffic_{name}.png — query error: {exc}\n')
            continue

        fig, ax = plt.subplots(figsize=(11, 2.8))
        fig.patch.set_facecolor('white')
        _render_stacked(ax, buckets, series, title, x_locator, xfmt)
        fig.tight_layout(pad=0.6)
        _save_atomic(fig, output_dir / f'traffic_{name}.png')
        plt.close(fig)

        elapsed = _time.monotonic() - t0
        total = sum(sum(v) for v in series.values()) if series else 0
        if stdout:
            stdout.write(
                f'    traffic_{name}.png — {total:,} requests ({elapsed:.1f}s)\n'
            )

    # Stored-stat windows — read pre-aggregated DashboardStat, no table scan
    stored_windows = [
        (
            '7d',
            now - timedelta(days=7),
            'Last 7 Days (per day)',
            mdates.DayLocator(),
            '%-d %b',
        ),
        (
            'all',
            None,
            'All Time (per day)',
            mdates.AutoDateLocator(),
            None,   # formatter set per-window below
        ),
    ]

    for name, cutoff, title, x_locator, xfmt in stored_windows:
        t0 = _time.monotonic()
        try:
            dates, series = _query_stored_daily(cutoff=cutoff)
        except Exception as exc:
            if stdout:
                stdout.write(f'    traffic_{name}.png — stat read error: {exc}\n')
            continue

        fig, ax = plt.subplots(figsize=(11, 2.8))
        fig.patch.set_facecolor('white')
        # xfmt is a format string for windowed, or None → AutoDateFormatter for all-time
        fmt = xfmt if xfmt else mdates.AutoDateFormatter(x_locator)
        _render_stacked(ax, dates, series, title, x_locator, fmt)
        fig.tight_layout(pad=0.6)
        _save_atomic(fig, output_dir / f'traffic_{name}.png')
        plt.close(fig)

        elapsed = _time.monotonic() - t0
        total = sum(sum(v) for v in series.values()) if series else 0
        if stdout:
            stdout.write(
                f'    traffic_{name}.png — {total:,} total ({elapsed:.1f}s)\n'
            )
