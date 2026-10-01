#!/usr/bin/env python3
"""Compute monthly-report stats from CrawlerVisit CSV chunk dumps.

The 2026-09-03..09 crawlervisit chunk was dropped from production (it was too
large to compress) after being dumped to crawlervisit_<date>.csv files. This
script produces the same stats the end-of-month report pulls from the live
hypertable, so the two can be combined into full-month numbers.

Needs only DuckDB (pip install duckdb). Nothing touches production.

Usage:
    python tools/dump_stats.py ingest /path/to/dump-dir   # safe to re-run
    python tools/dump_stats.py report [--json out.json] [--export-ips DIR]

`ingest` skips files that are still extracting (a sibling .csv.gz exists) and
files already ingested, so run it again as more files finish unzipping. Each
file is scanned twice: once for a per-second aggregate of every row, once to
keep the rows of the small traps the report drills into.

Merging with the production numbers:
  - counts (totals, per-trap, per-bot, pptx/pdf, per-file) -> add
  - peaks (busiest second/minute)                           -> take the max
  - distinct counts (IPs, unique files)                    -> NOT additive;
    use --export-ips to write the IP sets and union them with the prod sets.
"""
import argparse
import json
import sys
from pathlib import Path

import duckdb

DEFAULT_DB = Path(__file__).resolve().parent / 'dump_stats.duckdb'

COLUMNS = {
    'id': 'VARCHAR', 'timestamp': 'VARCHAR', 'ip_address': 'VARCHAR',
    'user_agent': 'VARCHAR', 'host': 'VARCHAR', 'path': 'VARCHAR',
    'referrer': 'VARCHAR', 'trap_type': 'VARCHAR', 'query_string': 'VARCHAR',
    'bot_type': 'VARCHAR', 'bot_group': 'VARCHAR', 'idempotency_key': 'VARCHAR',
}

# Rows kept verbatim for drill-down; everything else only survives as the
# per-second aggregate.
DETAIL_TRAPS = (
    'presentation', 'report_download', 'ghost_link', 'canary_trigger',
    'env_probe', 'wp_probe', 'webshell_probe',
)

EXEC_COMP_CSV = '/reports/executive-compensation-study-2024/download.csv'
DECK_RE = r'^(/presentations/[^/]+/\d+/\d+/\d+/[^/]+)/'


def connect(db_path, memory_limit='8GB', temp_dir=None):
    con = duckdb.connect(str(db_path))
    # Hard caps so a 60 GB CSV can't exhaust RAM/swap: DuckDB's default is
    # 80% of physical memory, spilling next to the database file.
    con.execute(f"SET memory_limit = '{memory_limit}'")
    con.execute('SET preserve_insertion_order = false')
    con.execute('SET threads = 6')
    if temp_dir:
        con.execute(f"SET temp_directory = '{temp_dir}'")
        con.execute("SET max_temp_directory_size = '200GB'")
    con.execute("""
        CREATE TABLE IF NOT EXISTS files (
            name VARCHAR PRIMARY KEY, day VARCHAR, rows BIGINT, expected BIGINT);
        -- `sec` is the timestamp text truncated to the second. Timestamps in
        -- the dump are all UTC ('+00'), so string truncation is exact and
        -- avoids parsing 650M timestamps.
        CREATE TABLE IF NOT EXISTS per_second (
            file VARCHAR, sec VARCHAR, trap_type VARCHAR, bot_type VARCHAR,
            kind VARCHAR, n BIGINT);
        CREATE TABLE IF NOT EXISTS detail (
            file VARCHAR, ts VARCHAR, ip VARCHAR, host VARCHAR, path VARCHAR,
            trap_type VARCHAR, bot_type VARCHAR, query_string VARCHAR);
    """)
    return con


def csv_source(path):
    cols = ', '.join(f"'{k}': '{v}'" for k, v in COLUMNS.items())
    # Explicit dialect: the Postgres COPY output escapes quotes by doubling
    # them ("/""/static/x.js"""), which DuckDB's sniffer can miss.
    return (f"read_csv('{path}', header=true, delim=',', quote='\"', escape='\"', "
            f"new_line='\\n', columns={{{cols}}})")


def ingest(con, dump_dir, only=None):
    files = sorted(Path(dump_dir).glob('crawlervisit_*.csv'))
    if only:
        files = [f for f in files if only in f.name]
    if not files:
        sys.exit(f'no crawlervisit_*.csv files in {dump_dir}')
    done = {r[0] for r in con.execute('SELECT name FROM files').fetchall()}

    for f in files:
        if f.name in done:
            print(f'{f.name}: already ingested')
            continue
        if f.with_name(f.name + '.gz').exists():
            print(f'{f.name}: .gz still present, probably extracting — skipping')
            continue

        day = f.stem.removeprefix('crawlervisit_')
        rc = f.with_suffix('.rowcount')
        expected = int(rc.read_text().strip()) if rc.exists() else None
        src = csv_source(f)
        print(f'{f.name}: scanning ...', flush=True)

        con.execute('BEGIN')
        con.execute('DELETE FROM per_second WHERE file = ?', [f.name])
        con.execute('DELETE FROM detail WHERE file = ?', [f.name])
        con.execute(f"""
            INSERT INTO per_second
            SELECT ?, left(timestamp, 19), trap_type, bot_type,
                   CASE WHEN trap_type = 'presentation' AND path LIKE '%.pptx' THEN 'pptx'
                        WHEN trap_type = 'presentation' AND path LIKE '%.pdf' THEN 'pdf'
                        WHEN trap_type = 'report_download' AND path LIKE '%.csv' THEN 'report_csv'
                        WHEN trap_type = 'report_download' AND path LIKE '%.pdf' THEN 'report_pdf'
                        ELSE '' END,
                   count(*)
            FROM {src} GROUP BY ALL
        """, [f.name])
        con.execute(f"""
            INSERT INTO detail
            SELECT ?, timestamp, ip_address, host, path, trap_type, bot_type, query_string
            FROM {src}
            WHERE trap_type IN {DETAIL_TRAPS} OR path LIKE '/employees/export%'
        """, [f.name])
        rows = con.execute('SELECT sum(n) FROM per_second WHERE file = ?', [f.name]).fetchone()[0]
        con.execute('INSERT INTO files VALUES (?, ?, ?, ?)', [f.name, day, rows, expected])
        con.execute('COMMIT')

        status = 'OK' if expected in (None, rows) else f'MISMATCH (expected {expected:,})'
        print(f'{f.name}: {rows:,} rows — {status}', flush=True)


def q(con, sql, *params):
    cur = con.execute(sql, list(params))
    cols = [d[0] for d in cur.description]
    return [dict(zip(cols, r)) for r in cur.fetchall()]


def report(con, json_out=None, export_ips=None):
    r = {}
    r['files'] = q(con, 'SELECT day, rows, expected FROM files ORDER BY day')
    r['total_rows'] = sum(f['rows'] for f in r['files'])

    r['daily'] = q(con, """
        SELECT left(sec, 10) AS day, sum(n) n FROM per_second GROUP BY 1 ORDER BY 1""")
    r['by_trap_type'] = q(con, """
        SELECT trap_type, sum(n) n FROM per_second GROUP BY 1 ORDER BY 2 DESC""")
    r['by_bot_type'] = q(con, """
        SELECT bot_type, sum(n) n FROM per_second GROUP BY 1 ORDER BY 2 DESC LIMIT 25""")

    # Site-wide peaks.
    r['busiest_seconds'] = q(con, """
        SELECT sec, sum(n) n FROM per_second GROUP BY 1 ORDER BY 2 DESC LIMIT 5""")
    r['busiest_minutes'] = q(con, """
        SELECT left(sec, 16) AS minute, sum(n) n FROM per_second GROUP BY 1 ORDER BY 2 DESC LIMIT 5""")

    # PowerPoint / PDF / report-file requests: totals and peak rates.
    r['file_requests'] = q(con, """
        SELECT kind, sum(n) n FROM per_second WHERE kind <> '' GROUP BY 1 ORDER BY 1""")
    r['file_peak_second'] = q(con, """
        WITH s AS (SELECT sec, kind, sum(n) n FROM per_second WHERE kind <> '' GROUP BY 1, 2)
        SELECT kind, max(n) peak FROM s GROUP BY 1
        UNION ALL
        SELECT 'pptx+pdf', max(n) FROM (
            SELECT sec, sum(n) n FROM s WHERE kind IN ('pptx', 'pdf') GROUP BY 1)""")
    r['file_peak_minute'] = q(con, """
        WITH m AS (SELECT left(sec, 16) AS minute, kind, sum(n) n
                   FROM per_second WHERE kind <> '' GROUP BY 1, 2)
        SELECT kind, arg_max(minute, n) AS minute, max(n) peak FROM m GROUP BY 1""")
    r['pptx_peak_seconds'] = q(con, """
        SELECT sec, sum(n) n FROM per_second WHERE kind = 'pptx'
        GROUP BY 1 ORDER BY 2 DESC LIMIT 5""")

    # Presentation drill-down.
    r['presentation_downloads'] = q(con, """
        SELECT CASE WHEN path LIKE '%.pptx' THEN 'pptx' ELSE 'pdf' END kind,
               count(*) n, count(DISTINCT path) unique_files, count(DISTINCT ip) ips
        FROM detail WHERE trap_type = 'presentation'
          AND (path LIKE '%download.pptx' OR path LIKE '%download.pdf')
        GROUP BY 1""")
    r['top_decks'] = q(con, f"""
        SELECT regexp_extract(path, '{DECK_RE}', 1) deck, count(*) n,
               count(*) FILTER (WHERE path LIKE '%/download.%') downloads,
               count(DISTINCT ip) ips
        FROM detail WHERE trap_type = 'presentation'
        GROUP BY 1 HAVING deck <> '' ORDER BY 2 DESC LIMIT 20""")

    # Reports.
    r['top_report_files'] = q(con, """
        SELECT path, count(*) n FROM detail WHERE trap_type = 'report_download'
        GROUP BY 1 ORDER BY 2 DESC LIMIT 10""")
    r['exec_comp_csv'] = q(con, """
        SELECT count(*) n, count(DISTINCT ip) ips, min(ts) AS first, max(ts) AS last
        FROM detail WHERE path = ?""", EXEC_COMP_CSV)

    # Ghost links, probes, canaries.
    r['ghost_links'] = q(con, """
        SELECT count(*) n, count(DISTINCT ip) ips FROM detail WHERE trap_type = 'ghost_link'""")
    r['ghost_links_by_bot'] = q(con, """
        SELECT bot_type, count(*) n, count(DISTINCT ip) ips FROM detail
        WHERE trap_type = 'ghost_link' GROUP BY 1 ORDER BY 2 DESC LIMIT 15""")
    r['employees_export_probes'] = q(con, """
        SELECT count(*) n, count(DISTINCT ip) ips FROM detail WHERE path LIKE '/employees/export%'""")
    r['canary_triggers'] = q(con, """
        SELECT ts, ip, path FROM detail WHERE trap_type = 'canary_trigger' ORDER BY ts""")

    if export_ips:
        out = Path(export_ips)
        out.mkdir(parents=True, exist_ok=True)
        sets = {
            'exec_comp_csv': f"path = '{EXEC_COMP_CSV}'",
            'pptx': "trap_type = 'presentation' AND path LIKE '%download.pptx'",
            'pdf': "trap_type = 'presentation' AND path LIKE '%download.pdf'",
            'ghost_link': "trap_type = 'ghost_link'",
        }
        for name, where in sets.items():
            con.execute(f"""COPY (SELECT DISTINCT ip FROM detail WHERE {where})
                            TO '{out / f"{name}_ips.parquet"}' (FORMAT parquet)""")
        print(f'IP sets written to {out}/', file=sys.stderr)

    if json_out:
        Path(json_out).write_text(json.dumps(r, indent=2, default=str))
        print(f'wrote {json_out}', file=sys.stderr)
    else:
        for key, val in r.items():
            print(f'\n== {key}')
            if isinstance(val, list):
                for row in val:
                    print('  ' + '  '.join(
                        f'{v:,}' if isinstance(v, int) else str(v) for v in row.values()))
            else:
                print(f'  {val:,}')


def main():
    p = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument('--db', default=DEFAULT_DB, help=f'results database (default {DEFAULT_DB})')
    p.add_argument('--memory-limit', default='8GB', help='DuckDB memory cap (default 8GB)')
    p.add_argument('--temp-dir', help='where DuckDB spills when over the memory cap')
    sub = p.add_subparsers(dest='cmd', required=True)
    i = sub.add_parser('ingest')
    i.add_argument('dump_dir')
    i.add_argument('--only', help='only ingest files whose name contains this (e.g. 2026-09-07)')
    rp = sub.add_parser('report')
    rp.add_argument('--json', help='write the stats as JSON instead of printing')
    rp.add_argument('--export-ips', help='write distinct-IP sets as parquet for merging')
    args = p.parse_args()

    con = connect(args.db, args.memory_limit, args.temp_dir)
    if args.cmd == 'ingest':
        ingest(con, args.dump_dir, args.only)
    else:
        report(con, args.json, args.export_ips)


if __name__ == '__main__':
    main()
