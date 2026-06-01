"""
Traffic graph generation for the ACPWB dashboard.

Generates 5 PNG stacked-area charts saved to staticfiles/graphs/:
  traffic_1h.png   — last hour,    per-minute buckets     (live query)
  traffic_8h.png   — last 8 hours, per-10-minute buckets  (live query)
  traffic_24h.png  — last 24 hrs,  per-hour buckets       (live query)
  traffic_7d.png   — last 7 days,  per-day buckets        (stored DashboardStat)
  traffic_all.png  — all time,     per-day buckets        (stored DashboardStat)

Live queries (1h/8h/24h) are timestamp-range-bounded — they touch only the
recent slice of the table and use the (bot_type, timestamp) composite index.
At any reasonable traffic rate the scanned row count is small regardless of
the total table size.

Long-window graphs (7d/all) read from pre-aggregated DashboardStat rows
(crawlers.daily_by_bot_type) written by precalc_dashboard — no table scan.

Bots with < 1% share of total traffic are grouped into "Others".

Requires matplotlib (pip install matplotlib).
"""
import os
import tempfile
import time as _time
from collections import defaultdict
from datetime import datetime, timedelta, timezone as dt_timezone
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
# Assigned sequentially by bot volume (index 0 = highest-traffic bot). "Others"
# is always neutral gray. Palette alternates warm/cool across the hue wheel so
# adjacent ranks are always visually distinct.

_BOT_COLOR_PALETTE = [
    '#4878CF',  # steel blue
    '#E84D4D',  # coral red
    '#6ACC65',  # sage green
    '#F39B2D',  # amber
    '#8172B3',  # medium purple
    '#4CB4B7',  # teal
    '#DD8452',  # terracotta
    '#55A868',  # forest green
    '#DA8BC3',  # mauve pink
    '#CCB974',  # gold tan
    '#C44E52',  # brick red
    '#64B5CD',  # sky blue
    '#937860',  # warm brown
    '#7B4D9E',  # deep purple
    '#2A9D8F',  # dark teal
    '#F4A261',  # peach orange
    '#457B9D',  # dark slate blue
    '#E76F51',  # burnt orange
    '#A8C256',  # olive green
    '#B07AA1',  # dusty purple
]
_OTHERS_COLOR = '#95A5A6'  # neutral gray


def _assign_colors(groups):
    """
    Return a color list aligned to groups, assigned sequentially by rank.

    Colors are drawn in palette order so adjacent legend entries are visually
    distinct. 'Others' always gets the fixed neutral gray.
    """
    palette_idx = 0
    colors = []
    for g in groups:
        if g == 'Others':
            colors.append(_OTHERS_COLOR)
        else:
            colors.append(_BOT_COLOR_PALETTE[palette_idx % len(_BOT_COLOR_PALETTE)])
            palette_idx += 1
    return colors

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


# ── Threshold helper ───────────────────────────────────────────────────────────

def _apply_threshold(series, threshold_pct=1.0):
    """
    Group bot types with < threshold_pct% total share into 'Others'.

    Returns a new ordered dict: significant bots sorted by total descending,
    followed by 'Others' (if any sub-threshold bots exist).
    """
    if not series:
        return {}

    totals = {bot: sum(vals) for bot, vals in series.items()}
    grand_total = sum(totals.values())
    if grand_total == 0:
        return {}

    n = len(next(iter(series.values())))
    significant = {}
    others = [0] * n

    for bot, vals in series.items():
        if totals[bot] / grand_total >= threshold_pct / 100:
            significant[bot] = vals
        else:
            for i, v in enumerate(vals):
                others[i] += v

    result = dict(sorted(significant.items(), key=lambda kv: totals[kv[0]], reverse=True))
    if any(v > 0 for v in others):
        result['Others'] = others
    return result


# ── DB queries (live — recent windows only) ────────────────────────────────────

def _query_windowed(cutoff, interval_minutes):
    """
    Query CrawlerVisit since cutoff, grouped by bot_type, using timestamp index.

    Safe for large tables: the WHERE timestamp >= cutoff clause is satisfied
    via an index range scan on the recent fraction of rows only.

    Returns (buckets, series) where:
      buckets — sorted list of datetime objects covering [cutoff, now]
      series  — dict[bot_name → list[int]] aligned to buckets; bots < 1% share
                are merged into 'Others'
    """
    from django.db.models import Count
    from django.db.models.functions import TruncHour, TruncMinute
    from django.utils import timezone

    from apps.honeypot.models import CrawlerVisit

    now = timezone.now()
    trunc_fn = TruncMinute if interval_minutes < 60 else TruncHour
    buckets = _make_buckets(cutoff, now, interval_minutes)

    cv_rows = (
        CrawlerVisit.objects
        .filter(timestamp__gte=cutoff)
        .annotate(bucket=trunc_fn('timestamp'))
        .values('bucket', 'bot_type')
        .annotate(c=Count('id'))
    )

    data = defaultdict(lambda: defaultdict(int))
    for row in cv_rows:
        b = _floor_to_interval(row['bucket'], interval_minutes)
        g = row['bot_type'] or '(empty user agent)'
        data[b][g] += row['c']

    present = set()
    for b in data:
        present.update(data[b].keys())

    raw_series = {g: [data[b].get(g, 0) for b in buckets] for g in present}
    return buckets, _apply_threshold(raw_series)


# ── Stored-stat queries (no table scan) ────────────────────────────────────────

def _query_stored_daily(cutoff=None):
    """
    Read per-day per-bot-type totals from DashboardStat written by precalc_dashboard.

    Reads 'crawlers.daily_by_bot_type': {date_str: {bot_type: count}}.
    cutoff — if provided, only return dates >= cutoff.

    Returns (dates, series) where series bots < 1% share are merged into 'Others'.
    Falls back to the legacy 'crawlers.daily' total if the new stat is absent.
    """
    from apps.core.models import DashboardStat

    try:
        daily_by_bot = DashboardStat.objects.get(key='crawlers.daily_by_bot_type').value
    except DashboardStat.DoesNotExist:
        daily_by_bot = {}

    # Legacy fallback: show a single 'Crawler Hits' series
    if not daily_by_bot:
        crawlers_daily = {}
        try:
            crawlers_daily = DashboardStat.objects.get(key='crawlers.daily').value
        except DashboardStat.DoesNotExist:
            pass
        all_dates = sorted(crawlers_daily)
        if cutoff:
            cutoff_str = cutoff.date().isoformat()
            all_dates = [d for d in all_dates if d >= cutoff_str]
        if not all_dates:
            return [], {}
        date_objs = [datetime.fromisoformat(d).replace(tzinfo=dt_timezone.utc) for d in all_dates]
        return date_objs, {'Crawler Hits': [crawlers_daily.get(d, 0) for d in all_dates]}

    all_dates = sorted(daily_by_bot)
    if cutoff:
        cutoff_str = cutoff.date().isoformat()
        all_dates = [d for d in all_dates if d >= cutoff_str]
    if not all_dates:
        return [], {}

    date_objs = [datetime.fromisoformat(d).replace(tzinfo=dt_timezone.utc) for d in all_dates]

    all_bots = set()
    for d in all_dates:
        all_bots.update(daily_by_bot.get(d, {}).keys())

    raw_series = {
        bot: [daily_by_bot.get(d, {}).get(bot, 0) for d in all_dates]
        for bot in all_bots
    }
    return date_objs, _apply_threshold(raw_series)


# ── Rendering helpers ──────────────────────────────────────────────────────────

def _apply_style(ax, title):
    """Apply consistent visual style to an axes object."""
    ax.set_facecolor('#FAFBFC')
    ax.set_title(title, fontsize=12, color='#2C3E50', pad=7, fontweight='bold')
    ax.tick_params(axis='both', labelsize=9, colors='#7F8C8D')
    ax.grid(axis='y', color='#E8ECEF', linewidth=0.5, zorder=0)
    for spine in ax.spines.values():
        spine.set_edgecolor('#E8ECEF')


def _render_stacked(ax, xs, series, title, x_locator, x_fmt):
    """
    Draw a stacked area chart.

    series must already be ordered (significant bots desc, Others last) as
    returned by _apply_threshold(). Colors are assigned from _BOT_COLOR_PALETTE
    in order; 'Others' always gets _OTHERS_COLOR.

    x_fmt may be a strftime format string or a matplotlib Formatter object.
    """
    _apply_style(ax, title)
    if not xs or not series:
        ax.text(0.5, 0.5, 'No data', ha='center', va='center',
                transform=ax.transAxes, color='#95A5A6', fontsize=9)
        return

    groups = list(series.keys())
    ys = [series[g] for g in groups]
    colors = _assign_colors(groups)

    ax.stackplot(xs, ys, labels=groups, colors=colors, alpha=0.88, zorder=2)
    # Set ylim AFTER stackplot so autoscaling has already computed the data range.
    # Setting it before (on a fresh axes with default top=1.0) would lock the scale to [0,1].
    ax.set_ylim(bottom=0)
    ax.yaxis.set_major_formatter(plt.FuncFormatter(lambda x, _: f'{int(x):,}'))
    ax.set_xlim(xs[0], xs[-1])
    ax.xaxis.set_major_locator(x_locator)
    if isinstance(x_fmt, str):
        ax.xaxis.set_major_formatter(mdates.DateFormatter(x_fmt))
    else:
        ax.xaxis.set_major_formatter(x_fmt)
    ax.legend(loc='upper center', bbox_to_anchor=(0.5, -0.18), fontsize=8,
              framealpha=0.85, ncol=4, handlelength=1.2, handletextpad=0.5,
              columnspacing=1.0)


def _render_stacked_bar(ax, dates, series, title, x_locator, x_fmt):
    """
    Draw a stacked bar chart for daily data.

    Each bar represents one day; bars are stacked by bot type in the same colour
    scheme as _render_stacked().  dates must be a list of datetime objects;
    series must be ordered (significant bots desc, Others last) as returned by
    _apply_threshold().
    """
    _apply_style(ax, title)
    if not dates or not series:
        ax.text(0.5, 0.5, 'No data', ha='center', va='center',
                transform=ax.transAxes, color='#95A5A6', fontsize=9)
        return

    groups = list(series.keys())
    colors = _assign_colors(groups)

    x = mdates.date2num(dates)
    width = 0.8  # fraction of 1-day spacing; leaves a gap between bars
    bottom = [0.0] * len(dates)

    for g, color in zip(groups, colors):
        vals = series[g]
        ax.bar(x, vals, bottom=bottom, width=width, label=g, color=color, alpha=0.88, zorder=2)
        bottom = [b + v for b, v in zip(bottom, vals)]

    ax.set_ylim(bottom=0)
    ax.yaxis.set_major_formatter(plt.FuncFormatter(lambda v, _: f'{int(v):,}'))
    ax.xaxis_date()
    ax.xaxis.set_major_locator(x_locator)
    if isinstance(x_fmt, str):
        ax.xaxis.set_major_formatter(mdates.DateFormatter(x_fmt))
    else:
        ax.xaxis.set_major_formatter(x_fmt)
    ax.legend(loc='upper center', bbox_to_anchor=(0.5, -0.18), fontsize=8,
              framealpha=0.85, ncol=4, handlelength=1.2, handletextpad=0.5,
              columnspacing=1.0)


# ── File I/O ───────────────────────────────────────────────────────────────────

def _save_atomic(fig, path):
    """Write to a temp file then atomically rename — nginx never sees a partial PNG."""
    path = Path(path)
    fd, tmp = tempfile.mkstemp(suffix='.png', dir=path.parent)
    try:
        os.close(fd)
        os.chmod(tmp, 0o644)  # ensure nginx (non-root) can read the file
        fig.savefig(tmp, dpi=150, bbox_inches='tight',
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
            2,
            '%H:%M',
            mdates.HourLocator(),
            'Last 8 Hours (per 2 min)',
        ),
        (
            '24h',
            now - timedelta(hours=24),
            5,
            '%H:%M',
            mdates.HourLocator(byhour=[0, 4, 8, 12, 16, 20]),
            'Last 24 Hours (per 5 min)',
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

        fig, ax = plt.subplots(figsize=(13, 3.5))
        fig.patch.set_facecolor('white')
        _render_stacked(ax, buckets, series, title, x_locator, xfmt)
        fig.tight_layout(pad=0.8)
        _save_atomic(fig, output_dir / f'traffic_{name}.png')
        plt.close(fig)

        elapsed = _time.monotonic() - t0
        total = sum(sum(v) for v in series.values()) if series else 0
        if stdout:
            stdout.write(
                f'    traffic_{name}.png — {total:,} requests ({elapsed:.1f}s)\n'
            )

    # 7d — live query at 30-min buckets (336 data points); timestamp-bounded → index scan only
    t0 = _time.monotonic()
    try:
        buckets, series = _query_windowed(now - timedelta(days=7), 30)
        fig, ax = plt.subplots(figsize=(11, 2.8))
        fig.patch.set_facecolor('white')
        _render_stacked(ax, buckets, series, 'Last 7 Days (per 30 min)',
                        mdates.DayLocator(), '%-d %b')
        fig.tight_layout(pad=0.6)
        _save_atomic(fig, output_dir / 'traffic_7d.png')
        plt.close(fig)
        elapsed = _time.monotonic() - t0
        total = sum(sum(v) for v in series.values()) if series else 0
        if stdout:
            stdout.write(f'    traffic_7d.png — {total:,} requests ({elapsed:.1f}s)\n')
    except Exception as exc:
        if stdout:
            stdout.write(f'    traffic_7d.png — query error: {exc}\n')

    # all-time — daily stored stat
    t0 = _time.monotonic()
    try:
        dates, series = _query_stored_daily(cutoff=None)
        x_locator = mdates.AutoDateLocator()
        fig, ax = plt.subplots(figsize=(11, 2.8))
        fig.patch.set_facecolor('white')
        _render_stacked_bar(ax, dates, series, 'All Time (per day)',
                            x_locator, mdates.AutoDateFormatter(x_locator))
        fig.tight_layout(pad=0.6)
        _save_atomic(fig, output_dir / 'traffic_all.png')
        plt.close(fig)
        elapsed = _time.monotonic() - t0
        total = sum(sum(v) for v in series.values()) if series else 0
        if stdout:
            stdout.write(f'    traffic_all.png — {total:,} total ({elapsed:.1f}s)\n')
    except Exception as exc:
        if stdout:
            stdout.write(f'    traffic_all.png — stat read error: {exc}\n')
