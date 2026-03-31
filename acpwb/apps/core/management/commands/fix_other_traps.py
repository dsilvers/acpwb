"""
One-shot cleanup for the double-logging bugs that inflated the 'other' trap type.

Steps performed (in order, under the precalc lock):
  1. Acquire /tmp/precalc_dashboard.lock — blocks the cron from running concurrently.
  2. Run dedupe_crawler_visits to delete 'other' duplicates of archive/view-logged rows.
  3. Delete all DashboardStat rows (stats and high-water marks).
  4. Run precalc_dashboard to recompute everything from the clean CrawlerVisit table.

Run once after deploying the middleware fix.

Usage:
    python manage.py fix_other_traps
    python manage.py fix_other_traps --dry-run   # dedupes with --dry-run, skips stat reset
"""
import fcntl
import time

from django.core.management import call_command
from django.core.management.base import BaseCommand

from apps.core.models import DashboardStat

LOCK_PATH = '/tmp/precalc_dashboard.lock'


class Command(BaseCommand):
    help = 'Dedupe CrawlerVisit rows and recompute all dashboard stats from scratch.'

    def add_arguments(self, parser):
        parser.add_argument(
            '--dry-run',
            action='store_true',
            help='Pass --dry-run to dedupe_crawler_visits; skip stat reset and precalc.',
        )

    def handle(self, *args, **options):
        dry_run = options['dry_run']

        self.stdout.write(f'Acquiring precalc lock at {LOCK_PATH} …')
        lockfile = open(LOCK_PATH, 'w')
        try:
            fcntl.flock(lockfile, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError:
            self.stdout.write(self.style.ERROR(
                'precalc_dashboard is currently running — wait for it to finish and retry.'
            ))
            lockfile.close()
            return

        self.stdout.write(self.style.SUCCESS('Lock acquired.'))

        try:
            self._step1_dedupe(dry_run)
            if not dry_run:
                self._step2_reset_stats()
                self._step3_precalc()
            else:
                self.stdout.write('\n--dry-run: skipping stat reset and precalc.')
        finally:
            fcntl.flock(lockfile, fcntl.LOCK_UN)
            lockfile.close()
            self.stdout.write('Lock released.')

    # ── Steps ─────────────────────────────────────────────────────────────────

    def _step1_dedupe(self, dry_run):
        self.stdout.write('\n── Step 1: dedupe_crawler_visits ─────────────────────────────')
        kwargs = {}
        if dry_run:
            kwargs['dry_run'] = True
        call_command('dedupe_crawler_visits', stdout=self.stdout, stderr=self.stderr, **kwargs)

    def _step2_reset_stats(self):
        self.stdout.write('\n── Step 2: reset DashboardStat rows ──────────────────────────')
        t0 = time.monotonic()
        count, _ = DashboardStat.objects.all().delete()
        elapsed = time.monotonic() - t0
        self.stdout.write(f'  Deleted {count} DashboardStat rows in {elapsed:.1f}s')

    def _step3_precalc(self):
        self.stdout.write('\n── Step 3: precalc_dashboard ─────────────────────────────────')
        # precalc_dashboard also tries to acquire the lock — call its internals via
        # call_command but note it will fail to get the lock since we hold it.
        # Instead, invoke it directly via call_command with the lock already held;
        # precalc uses LOCK_NB so it would bail.  We release ours first, then call it.
        #
        # To avoid that, we just import and call the internal update methods directly.
        from apps.core.management.commands.precalc_dashboard import Command as PrecalcCommand
        cmd = PrecalcCommand(stdout=self.stdout, stderr=self.stderr)
        cmd.stdout = self.stdout
        cmd.stderr = self.stderr
        cmd.style = self.style

        t0 = time.monotonic()
        cmd._update_crawlers()
        cmd._update_archive()
        cmd._update_emails()
        cmd._update_people()
        cmd._update_projects()
        cmd._update_login_attempts()
        cmd._update_opt_outs()
        cmd._update_canary()
        elapsed = time.monotonic() - t0
        self.stdout.write(f'\n  precalc_dashboard completed in {elapsed:.1f}s')
