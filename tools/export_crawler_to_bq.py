#!/usr/bin/env python3
"""
Export honeypot_crawlervisit to BigQuery via GCS.

Streams data from PostgreSQL using COPY TO STDOUT (no full-table buffering),
writes gzipped CSV chunks, uploads each to GCS, then submits a BigQuery load job.

Requirements:
    pip install "psycopg[binary]" google-cloud-storage google-cloud-bigquery

Auth:
    export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
    # or: gcloud auth application-default login

DB config from env vars (same as .env):
    DB_HOST, DB_NAME, DB_USER, DB_PASSWORD, DB_PORT
    (or DATABASE_URL as fallback)

Usage:
    python tools/export_crawler_to_bq.py \\
        --bucket my-gcs-bucket \\
        --bq-dataset acpwb \\
        --bq-table crawler_visits

    # Test locally without uploading:
    python tools/export_crawler_to_bq.py --skip-upload --rows-per-file 100000 \\
        --bucket unused --bq-dataset unused --bq-table unused
"""
import argparse
import gzip
import os
import sys
import tempfile
import time
from pathlib import Path


# ── BigQuery schema ────────────────────────────────────────────────────────────

BQ_SCHEMA = [
    ("id",           "INTEGER"),
    ("timestamp",    "TIMESTAMP"),
    ("ip_address",   "STRING"),
    ("user_agent",   "STRING"),
    ("host",         "STRING"),
    ("path",         "STRING"),
    ("referrer",     "STRING"),
    ("trap_type",    "STRING"),
    ("query_string", "STRING"),
    ("bot_type",     "STRING"),
    ("bot_group",    "STRING"),
]


# ── Helpers ────────────────────────────────────────────────────────────────────

def _db_dsn():
    if url := os.environ.get("DATABASE_URL"):
        return url
    host     = os.environ.get("DB_HOST",     "localhost")
    port     = os.environ.get("DB_PORT",     "5432")
    name     = os.environ.get("DB_NAME",     "acpwb")
    user     = os.environ.get("DB_USER",     "acpwb")
    password = os.environ.get("DB_PASSWORD", "")
    return f"host={host} port={port} dbname={name} user={user} password={password}"


def _fmt_bytes(n):
    for unit in ("B", "KB", "MB", "GB"):
        if n < 1024:
            return f"{n:.1f} {unit}"
        n /= 1024
    return f"{n:.1f} TB"


def _fmt_elapsed(seconds):
    m, s = divmod(int(seconds), 60)
    h, m = divmod(m, 60)
    return f"{h}h {m:02d}m {s:02d}s" if h else f"{m}m {s:02d}s" if m else f"{s}s"


# ── GCS helpers ────────────────────────────────────────────────────────────────

def _ensure_bucket(storage_client, bucket_name, location="US"):
    from google.cloud.exceptions import NotFound
    bucket = storage_client.bucket(bucket_name)
    try:
        storage_client.get_bucket(bucket_name)
        print(f"  Bucket gs://{bucket_name} already exists.")
    except NotFound:
        storage_client.create_bucket(bucket_name, location=location)
        print(f"  Created bucket gs://{bucket_name} in {location}.")
    return bucket


def _upload(storage_client, bucket_name, gcs_path, local_path):
    bucket = storage_client.bucket(bucket_name)
    blob = bucket.blob(gcs_path)
    blob.upload_from_filename(local_path)


# ── Export ─────────────────────────────────────────────────────────────────────

def export(args):
    import psycopg

    cols = ", ".join(name for name, _ in BQ_SCHEMA)
    copy_sql = f"COPY (SELECT {cols} FROM honeypot_crawlervisit) TO STDOUT WITH (FORMAT CSV, HEADER)"

    tmp_dir = Path(args.tmp_dir)
    tmp_dir.mkdir(parents=True, exist_ok=True)

    if not args.skip_upload:
        from google.cloud import storage as gcs
        storage_client = gcs.Client()
        print("Ensuring GCS bucket exists ...")
        _ensure_bucket(storage_client, args.bucket)

    gcs_paths = []
    part = 0
    t_total = time.monotonic()

    print(f"\nConnecting to PostgreSQL ...")
    with psycopg.connect(_db_dsn()) as conn:
        conn.autocommit = True  # COPY doesn't need a transaction
        with conn.cursor() as cur:
            print(f"Starting COPY TO STDOUT ({cols}) ...\n")
            t_part = time.monotonic()
            rows_in_part = 0
            total_rows = 0
            gz_path = tmp_dir / f"part-{part:03d}.csv.gz"
            gz_file = gzip.open(gz_path, "wb", compresslevel=6)

            with cur.copy(copy_sql) as copy:
                for data in copy:
                    # data is bytes; each chunk ends with a newline per row
                    rows_in_chunk = data.count(b"\n")
                    gz_file.write(data)
                    rows_in_part += rows_in_chunk
                    total_rows   += rows_in_chunk

                    if rows_in_part >= args.rows_per_file:
                        gz_file.close()
                        _finish_part(
                            args, storage_client if not args.skip_upload else None,
                            gz_path, part, rows_in_part, t_part, gcs_paths,
                        )
                        part += 1
                        rows_in_part = 0
                        t_part = time.monotonic()
                        gz_path = tmp_dir / f"part-{part:03d}.csv.gz"
                        gz_file = gzip.open(gz_path, "wb", compresslevel=6)

            gz_file.close()
            if rows_in_part > 0:
                _finish_part(
                    args, storage_client if not args.skip_upload else None,
                    gz_path, part, rows_in_part, t_part, gcs_paths,
                )

    elapsed = time.monotonic() - t_total
    print(f"\nExport complete: {total_rows:,} rows, {part + 1} files, {_fmt_elapsed(elapsed)}")

    if args.skip_upload:
        print(f"Files written to {tmp_dir}/ (--skip-upload; no GCS upload performed)")
        return

    if args.skip_bq:
        print("GCS upload done (--skip-bq; BigQuery load skipped)")
        return

    _load_bigquery(args, gcs_paths)


def _finish_part(args, storage_client, gz_path, part, rows, t_part, gcs_paths):
    size = gz_path.stat().st_size
    elapsed = time.monotonic() - t_part
    gcs_path = f"{args.gcs_prefix}part-{part:03d}.csv.gz"

    if storage_client:
        print(f"  part-{part:03d}.csv.gz  {rows:>10,} rows  {_fmt_bytes(size)}  uploading ...", end="", flush=True)
        t_up = time.monotonic()
        _upload(storage_client, args.bucket, gcs_path, gz_path)
        gz_path.unlink()
        print(f"  done ({_fmt_elapsed(time.monotonic() - t_up)})")
    else:
        print(f"  part-{part:03d}.csv.gz  {rows:>10,} rows  {_fmt_bytes(size)}  {_fmt_elapsed(elapsed)}")

    gcs_paths.append(f"gs://{args.bucket}/{gcs_path}")


# ── BigQuery load ──────────────────────────────────────────────────────────────

def _load_bigquery(args, gcs_paths):
    from google.cloud import bigquery

    bq = bigquery.Client()
    dataset_ref = bq.dataset(args.bq_dataset)

    print(f"\nEnsuring BigQuery dataset {args.bq_dataset} exists ...")
    bq.create_dataset(dataset_ref, exists_ok=True)

    schema = [bigquery.SchemaField(name, bq_type) for name, bq_type in BQ_SCHEMA]

    job_config = bigquery.LoadJobConfig(
        schema=schema,
        source_format=bigquery.SourceFormat.CSV,
        skip_leading_rows=1,
        allow_quoted_newlines=True,
        write_disposition=bigquery.WriteDisposition.WRITE_TRUNCATE,
        create_disposition=bigquery.CreateDisposition.CREATE_IF_NEEDED,
    )

    table_ref = dataset_ref.table(args.bq_table)
    uri = f"gs://{args.bucket}/{args.gcs_prefix}part-*.csv.gz"

    print(f"Submitting BigQuery load job from {uri} ...")
    t0 = time.monotonic()
    job = bq.load_table_from_uri(uri, table_ref, job_config=job_config)

    print(f"  Job ID: {job.job_id}")
    print("  Waiting ...", end="", flush=True)
    while not job.done():
        time.sleep(5)
        print(".", end="", flush=True)
    print()

    job.result()  # raises on error
    elapsed = time.monotonic() - t0
    table = bq.get_table(table_ref)
    print(f"  Done in {_fmt_elapsed(elapsed)}. {table.num_rows:,} rows in {args.bq_dataset}.{args.bq_table}")


# ── CLI ────────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--bucket",       required=True, help="GCS bucket name")
    parser.add_argument("--bq-dataset",   required=True, help="BigQuery dataset name")
    parser.add_argument("--bq-table",     required=True, help="BigQuery table name")
    parser.add_argument("--gcs-prefix",   default="crawler_export/", help="GCS object prefix (default: crawler_export/)")
    parser.add_argument("--rows-per-file",type=int, default=10_000_000, help="Rows per gzip file (default: 10M)")
    parser.add_argument("--tmp-dir",      default="/tmp/crawler_export", help="Local temp directory for CSV chunks")
    parser.add_argument("--skip-upload",  action="store_true", help="Write files locally only, skip GCS upload")
    parser.add_argument("--skip-bq",      action="store_true", help="Upload to GCS but skip BigQuery load job")
    args = parser.parse_args()

    if not args.gcs_prefix.endswith("/"):
        args.gcs_prefix += "/"

    export(args)


if __name__ == "__main__":
    main()
