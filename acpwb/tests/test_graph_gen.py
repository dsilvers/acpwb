"""
Tests for apps/core/graph_gen.py

These tests run against an in-memory SQLite-equivalent (whatever pytest-django
provides) without any actual CrawlerVisit/ArchiveVisit rows, so they verify
that graph_gen handles empty data gracefully and produces valid output files.
"""
import os
import tempfile
from pathlib import Path

import pytest


# ── helpers ────────────────────────────────────────────────────────────────────

def _graph_dir():
    """Return a fresh temp directory for each test."""
    d = tempfile.mkdtemp()
    return Path(d)


# ── import guard ───────────────────────────────────────────────────────────────

def test_matplotlib_available():
    """matplotlib must be installed (listed in requirements.txt)."""
    import matplotlib  # noqa: F401


def test_graph_gen_importable():
    from apps.core import graph_gen  # noqa: F401
    assert graph_gen._MATPLOTLIB_AVAILABLE, (
        'matplotlib is installed but graph_gen._MATPLOTLIB_AVAILABLE is False'
    )


# ── bucket helpers ─────────────────────────────────────────────────────────────

def test_floor_to_interval_rounds_down():
    from datetime import datetime, timezone
    from apps.core.graph_gen import _floor_to_interval

    dt = datetime(2024, 6, 15, 14, 37, 45, tzinfo=timezone.utc)
    floored = _floor_to_interval(dt, 10)
    assert floored.minute == 30
    assert floored.second == 0


def test_floor_to_interval_already_on_boundary():
    from datetime import datetime, timezone
    from apps.core.graph_gen import _floor_to_interval

    dt = datetime(2024, 6, 15, 14, 30, 0, tzinfo=timezone.utc)
    assert _floor_to_interval(dt, 10) == dt


def test_make_buckets_count():
    from datetime import datetime, timedelta, timezone
    from apps.core.graph_gen import _make_buckets

    start = datetime(2024, 6, 15, 14, 0, tzinfo=timezone.utc)
    end   = datetime(2024, 6, 15, 14, 59, tzinfo=timezone.utc)
    buckets = _make_buckets(start, end, 10)
    # 14:00, 14:10, 14:20, 14:30, 14:40, 14:50 → 6 buckets
    assert len(buckets) == 6
    assert buckets[0].minute == 0
    assert buckets[-1].minute == 50


# ── generate_traffic_graphs — no data ─────────────────────────────────────────

@pytest.mark.django_db
def test_generate_creates_all_five_files():
    """All 5 PNGs are written even when there are no rows in the DB."""
    from apps.core.graph_gen import generate_traffic_graphs

    out = _graph_dir()
    generate_traffic_graphs(out)

    expected = ['traffic_1h.png', 'traffic_8h.png', 'traffic_24h.png',
                'traffic_7d.png', 'traffic_all.png']
    for fname in expected:
        path = out / fname
        assert path.exists(), f'{fname} was not created'
        assert path.stat().st_size > 0, f'{fname} is empty'


@pytest.mark.django_db
def test_generate_files_are_valid_png():
    """Each output file starts with the PNG magic bytes."""
    from apps.core.graph_gen import generate_traffic_graphs

    PNG_MAGIC = b'\x89PNG\r\n\x1a\n'
    out = _graph_dir()
    generate_traffic_graphs(out)

    for fname in ['traffic_1h.png', 'traffic_8h.png', 'traffic_24h.png',
                  'traffic_7d.png', 'traffic_all.png']:
        with open(out / fname, 'rb') as f:
            header = f.read(8)
        assert header == PNG_MAGIC, f'{fname} is not a valid PNG'


@pytest.mark.django_db
def test_generate_stdout_messages(capsys):
    """generate_traffic_graphs writes progress lines when stdout is given."""
    import io
    from apps.core.graph_gen import generate_traffic_graphs

    buf = io.StringIO()
    out = _graph_dir()
    generate_traffic_graphs(out, stdout=buf)

    output = buf.getvalue()
    for name in ('1h', '8h', '24h', '7d', 'all'):
        assert f'traffic_{name}.png' in output


@pytest.mark.django_db
def test_generate_creates_output_dir_if_missing():
    """Output directory is created automatically."""
    from apps.core.graph_gen import generate_traffic_graphs

    out = _graph_dir() / 'nested' / 'graphs'
    assert not out.exists()
    generate_traffic_graphs(out)
    assert out.exists()


# ── generate with data ─────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_generate_with_crawler_visits(bot_client):
    """Graphs are generated successfully when CrawlerVisit rows exist."""
    from apps.core.graph_gen import generate_traffic_graphs

    # Create a few trap hits via the bot_client fixture
    bot_client.get('/internal/portal/')
    bot_client.get('/wiki/accounting/')
    bot_client.get('/internal/portal/')

    out = _graph_dir()
    generate_traffic_graphs(out)

    # All files must still be created and valid
    for fname in ['traffic_1h.png', 'traffic_8h.png', 'traffic_24h.png']:
        assert (out / fname).stat().st_size > 0


@pytest.mark.django_db
def test_generate_with_stored_stats():
    """7d and all-time graphs work when DashboardStat daily rows exist."""
    from apps.core.graph_gen import generate_traffic_graphs
    from apps.core.models import DashboardStat

    DashboardStat.objects.create(
        key='crawlers.daily',
        value={'2024-01-01': 120, '2024-01-02': 95, '2024-01-03': 210},
    )
    DashboardStat.objects.create(
        key='archive.daily',
        value={'2024-01-01': 40, '2024-01-02': 35, '2024-01-03': 88},
    )

    out = _graph_dir()
    generate_traffic_graphs(out)

    for fname in ['traffic_7d.png', 'traffic_all.png']:
        assert (out / fname).stat().st_size > 0


# ── _query_stored_daily ────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_query_stored_daily_empty():
    from apps.core.graph_gen import _query_stored_daily

    dates, series = _query_stored_daily()
    assert dates == []
    assert series == {}


@pytest.mark.django_db
def test_query_stored_daily_cutoff():
    from datetime import datetime, timezone
    from apps.core.graph_gen import _query_stored_daily
    from apps.core.models import DashboardStat

    DashboardStat.objects.create(
        key='crawlers.daily',
        value={'2024-01-01': 10, '2024-06-01': 20, '2024-12-01': 30},
    )

    cutoff = datetime(2024, 6, 1, tzinfo=timezone.utc)
    dates, series = _query_stored_daily(cutoff=cutoff)
    date_strs = [d.date().isoformat() for d in dates]
    assert '2024-01-01' not in date_strs
    assert '2024-06-01' in date_strs
    assert '2024-12-01' in date_strs


# ── _update_graphs in precalc_dashboard ───────────────────────────────────────

@pytest.mark.django_db
def test_precalc_update_graphs_runs(tmp_path, settings):
    """_update_graphs() completes without error and writes files."""
    import io

    settings.STATIC_ROOT = str(tmp_path)

    from apps.core.management.commands.precalc_dashboard import Command

    buf = io.StringIO()
    cmd = Command(stdout=buf, stderr=buf)
    cmd.style = type('S', (), {
        'SUCCESS': lambda self, s: s,
        'ERROR':   lambda self, s: s,
    })()
    cmd._update_graphs()

    graph_dir = tmp_path / 'graphs'
    assert graph_dir.exists()
    assert len(list(graph_dir.glob('traffic_*.png'))) == 5
