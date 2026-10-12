"""
Python mirror of acpwb_go/legacyjob/identity.go plus the article-page link
markup for the legacy "source code" pages (served by acpwb_go at
<article>/source-code/).

Exists ONLY so Django's pyrender output stays byte-identical to the Go
renderer for the archive fixture tests (acpwb_go/archive/testdata). The
source-code page itself has no Python implementation. Any change here must
be made in identity.go / archive/sourcecode.go too, and the archive fixtures
re-dumped.
"""
import hashlib

from apps.core.htmlgen import escape as e

_PROGRAM_NAMES = [
    'BNCHMK', 'PCTLAD', 'ENGTAB', 'STATPK', 'CMPBNK', 'RWDTAB', 'MKTPOS', 'PAYLAD',
    'SURVTB', 'BMKRPT', 'QTLCLC', 'ENGMET', 'DLTCLC', 'TABGEN', 'PCTRPT', 'WGSTAT',
    'COMPRA', 'BENCHX', 'KPITAB', 'MEDLAD', 'PERCNT', 'SALSRV', 'TRWTAB', 'EQUTAB',
    'MRKBNK', 'LADDER', 'RPTPCT', 'ENGDLT', 'CMPTAB', 'STPCTL', 'BNKSUM', 'QRTILE',
    'PAYSTA', 'WAGRPT', 'MKTLAD', 'SVYTAB', 'DELTAS', 'PCTGEN', 'BMKTAB', 'RSCHTB',
]


_LANGUAGE_TABLE = [
    ('f4', 'FORTRAN IV', 1985, 1993, 12),
    ('f77', 'FORTRAN 77', 1985, 2001, 18),
    ('f77', 'FORTRAN 77', 2002, 2012, 10),
    ('gf', 'Fortran (gfortran)', 2013, 2025, 14),
    ('cobol', 'COBOL', 1985, 2001, 20),
    ('cobol', 'COBOL', 2002, 2025, 6),
    ('basic', 'BASIC', 1985, 1993, 10),
    ('atari', 'Atari BASIC', 1985, 1991, 5),
    ('pick', 'Pick BASIC', 1985, 2008, 6),
    ('mumps', 'MUMPS', 1985, 2025, 5),
    ('ada', 'Ada 83', 1985, 1995, 6),
    ('ada', 'Ada 95', 1996, 2012, 4),
    ('ada', 'Ada 2012', 2013, 2025, 2),
    ('pascal', 'Turbo Pascal', 1985, 1998, 7),
    ('pascalbtr', 'Turbo Pascal + Btrieve', 1986, 2001, 4),
    ('pli', 'PL/I', 1985, 2005, 8),
    ('pli', 'Enterprise PL/I', 2006, 2025, 2),
    ('pal', 'Paradox PAL', 1987, 1999, 4),
    ('dbase', 'dBASE IV', 1989, 1994, 5),
    ('clipper', 'Clipper', 1988, 1999, 5),
    ('foxpro', 'FoxPro', 1991, 2007, 4),
    ('natural', 'Natural (Adabas)', 1985, 2025, 5),
    ('informix', 'Informix 4GL', 1987, 2008, 5),
    ('c', 'C', 1985, 2001, 16),
    ('c', 'C', 2002, 2025, 8),
    ('cbtr', 'C + Btrieve', 1988, 2004, 4),
    ('vb', 'Visual Basic 3.0', 1993, 1997, 5),
    ('vb', 'Visual Basic 6.0', 1998, 2012, 6),
    ('vba', 'Excel VBA', 1994, 2025, 7),
    ('vbnet', 'VB.NET', 2002, 2025, 6),
    ('csharp', 'C#', 2002, 2025, 7),
    ('perl4', 'Perl 4', 1991, 1996, 5),
    ('perl', 'Perl 5', 1995, 2025, 14),
    ('php', 'PHP 4', 2000, 2004, 6),
    ('php', 'PHP 5', 2005, 2015, 8),
    ('php', 'PHP 7', 2016, 2025, 6),
    ('py2', 'Python 2', 2001, 2012, 10),
    ('py2', 'Python 2.7', 2013, 2025, 10),
    ('sqlds', 'SQL/DS', 1985, 1995, 3),
    ('db2sql', 'DB2 SQL', 1985, 2025, 6),
    ('sqlplus', 'Oracle SQL*Plus', 1988, 2012, 5),
    ('plsql', 'PL/SQL', 1992, 2025, 6),
    ('tsql', 'Transact-SQL', 1993, 2025, 6),
    ('sqlpl', 'DB2 SQL PL', 1999, 2025, 4),
    ('sql', 'SQL', 1990, 2025, 5),
    ('rpg', 'RPG II', 1985, 1990, 4),
    ('rpg', 'RPG/400', 1988, 1996, 5),
    ('rpg', 'ILE RPG IV', 1995, 2012, 4),
    ('rpg', 'RPG IV (free-form)', 2013, 2025, 3),
    ('rexx', 'REXX', 1985, 2025, 5),
    ('clist', 'TSO CLIST', 1985, 2001, 3),
    ('cl400', 'AS/400 CL', 1988, 2007, 3),
    ('cl400', 'IBM i CL', 2008, 2025, 2),
    ('ezt', 'Easytrieve Plus', 1985, 2025, 4),
    ('abap', 'ABAP/4', 1993, 2012, 4),
    ('abap', 'ABAP', 2013, 2025, 4),
    ('apl', 'VS APL', 1985, 1988, 2),
    ('apl', 'APL2', 1989, 2004, 2),
    ('apl', 'Dyalog APL', 2005, 2025, 2),
    ('focus', 'FOCUS', 1985, 2010, 3),
    ('nomad', 'NOMAD', 1985, 1998, 2),
    ('ramis', 'RAMIS II', 1985, 1994, 2),
    ('sas', 'SAS', 1985, 2025, 6),
    ('spss', 'SPSS-X', 1985, 1993, 3),
    ('spss', 'SPSS', 1994, 2025, 3),
    ('splus', 'S-PLUS', 1988, 2007, 2),
    ('r', 'R', 2000, 2025, 4),
    ('matlab', 'MATLAB', 1988, 2025, 3),
    ('octave', 'GNU Octave', 1997, 2025, 2),
    ('stata', 'Stata', 1986, 2025, 3),
    ('awk', 'nawk', 1988, 1995, 4),
    ('awk', 'gawk', 1996, 2025, 4),
    ('ksh', 'Korn shell', 1988, 2025, 4),
    ('tcl', 'Tcl', 1991, 2025, 3),
    ('dc', 'dc', 1985, 2025, 1),
    ('bc', 'bc', 1985, 2025, 2),
    ('cfml', 'ColdFusion', 1996, 2025, 3),
    ('asp', 'Classic ASP', 1997, 2025, 3),
    ('modula2', 'Modula-2', 1985, 2000, 2),
    ('lisp', 'Franz Lisp', 1985, 1989, 2),
    ('lisp', 'Common Lisp', 1990, 2025, 2),
    ('prolog', 'Turbo Prolog', 1986, 1991, 2),
    ('prolog', 'Quintus Prolog', 1992, 2000, 1),
    ('prolog', 'SWI-Prolog', 2001, 2025, 1),
    ('forth', 'Forth-83', 1985, 1995, 2),
    ('forth', 'ANS Forth', 1996, 2008, 1),
    ('smalltalk', 'Smalltalk-80', 1985, 1995, 1),
    ('smalltalk', 'VisualWorks Smalltalk', 1996, 2012, 2),
]


def _era_languages(year):
    out = [(k, label, w) for k, label, a, b, w in _LANGUAGE_TABLE if a <= year <= b]
    lo = min(a for _, _, a, _, _ in _LANGUAGE_TABLE)
    hi = max(b for _, _, _, b, _ in _LANGUAGE_TABLE)
    # Mirrors legacyjob/identity.go:eraLanguages — out-of-span years borrow the nearest covered year.
    if not out and (year < lo or year > hi):
        return _era_languages(max(lo, min(year, hi)))
    return out


def legacy_identity(year, month, day, slug):
    h = hashlib.md5(f"legacyid_{year}_{month}_{day}_{slug}".encode()).hexdigest()
    n0, n1, n2, n3 = (int(h[i * 8:(i + 1) * 8], 16) for i in range(4))
    langs = _era_languages(year)
    p = n0 % sum(w for _, _, w in langs)
    lang = langs[-1]
    for opt in langs:
        if p < opt[2]:
            lang = opt
            break
        p -= opt[2]
    release = f"{1 + n2 % 4}.{(n2 // 4) % 10}"
    if (n2 // 40) % 3 == 0:
        release += chr(ord('A') + (n2 // 120) % 3)
    return {
        'program': _PROGRAM_NAMES[n1 % len(_PROGRAM_NAMES)],
        'release': release,
        'lang_key': lang[0],
        'language': lang[1],
        'job_number': f"JOB{10000 + n3 % 90000:05d}",
    }


def source_code_href(export_csv_url):
    if export_csv_url.endswith('export.csv'):
        export_csv_url = export_csv_url[:-len('export.csv')]
    return export_csv_url + 'source-code/'


def _ident(c):
    return legacy_identity(c['year'], c['month'], c['day'], c['slug']), source_code_href(c['export_csv_url'])


def source_caption_html(c, era):
    i, href = _ident(c)
    if era:
        return (f'<div class="acpwb-source-caption" style="font-size:.72rem;opacity:.7;margin-top:.4rem">'
                f'Tabulated by <span style="font-family:monospace">{e(i["program"])} rel. {e(i["release"])}</span> ({e(i["language"])}) &middot; '
                f'<a href="{e(href)}" class="era-nav-link">view source code &rarr;</a></div>')
    return (f'<div class="acpwb-source-caption" style="font-size:.72rem;color:var(--muted);margin-top:.4rem">'
            f'Tabulated by <span style="font-family:monospace">{e(i["program"])} rel. {e(i["release"])}</span> ({e(i["language"])}) &middot; '
            f'<a href="{e(href)}" style="color:var(--navy);font-weight:700">view source code &rarr;</a></div>')


def source_sidebar_card_html(c, rid, era):
    i, href = _ident(c)
    if era:
        return (f'<div id="acpwb-archive-{rid}-sidebar-source-card" style="background:rgba(128,128,128,.07);'
                f'border:1px solid rgba(128,128,128,.2);padding:1rem;margin-bottom:1rem">'
                f'<div class="era-section-head" style="margin-bottom:.6rem">Computation Source</div>'
                f'<dl class="mb-2" style="font-size:.82rem">'
                f'<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Program</dt>'
                f'<dd class="fw-700 mb-2">{e(i["program"])} rel. {e(i["release"])}</dd>'
                f'<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Language</dt>'
                f'<dd class="mb-2">{e(i["language"])}</dd>'
                f'<dt style="opacity:.55;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em">Job</dt>'
                f'<dd class="mb-0" style="font-family:monospace;font-size:.75rem;opacity:.7">{e(i["job_number"])}</dd></dl>'
                f'<a href="{e(href)}" class="era-nav-link" style="font-size:.8rem">View source code &rarr;</a></div>')
    return (f'<div id="acpwb-archive-{rid}-sidebar-source-card" class="acpwb-card mb-4">'
            f'<h6 class="card-title mb-2">Computation Source</h6><dl class="mb-2" style="font-size:.82rem">'
            f'<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">Program</dt>'
            f'<dd class="fw-700 mb-2">{e(i["program"])} rel. {e(i["release"])}</dd>'
            f'<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">Language</dt>'
            f'<dd class="mb-2">{e(i["language"])}</dd>'
            f'<dt style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.08em">Job</dt>'
            f'<dd class="mb-0" style="font-family:monospace;font-size:.75rem;color:var(--muted)">{e(i["job_number"])}</dd></dl>'
            f'<a href="{e(href)}" style="font-size:.8rem;color:var(--gold);font-weight:700;text-decoration:none">View source code &rarr;</a></div>')


def source_ref_entry_html(c):
    i, href = _ident(c)
    date = f"{c['year']}-{c['month']:02d}-{c['day']:02d}"
    return (f'<div style="background:white;border:1px solid var(--border);border-left:3px solid var(--gold);'
            f'padding:.75rem 1rem;margin-bottom:.75rem">'
            f'<div style="font-size:.65rem;font-weight:800;text-transform:uppercase;letter-spacing:.08em;'
            f'color:var(--muted);margin-bottom:.2rem">Source &bull; JOB LISTING &bull; {e(date)}</div>'
            f'<a href="{e(href)}" style="font-size:.9rem;font-weight:700;color:var(--navy);text-decoration:none">{e(i["program"])} Job Listing {e(i["job_number"])}</a>'
            f'<p class="small text-muted mb-0" style="font-size:.78rem;margin-top:.25rem">'
            f'{e(i["language"])} batch job that tabulated the engagement metrics and market benchmark tables in this record.</p></div>')


def source_ref_entry_era_html(c):
    i, href = _ident(c)
    date = f"{c['year']}-{c['month']:02d}-{c['day']:02d}"
    return (f'<a href="{e(href)}" class="era-entry-card" style="margin-bottom:.5rem">'
            f'<div style="font-size:.6rem;opacity:.5;margin-bottom:.15rem">{e(date)} &bull; source</div>'
            f'<div style="font-size:.75rem;font-weight:600;line-height:1.35">{e(i["program"])} Job Listing {e(i["job_number"])}</div></a>')
