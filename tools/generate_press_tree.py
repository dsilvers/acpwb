#!/usr/bin/env python3
"""Pre-render the press release answer tree as a static SVG (the no-JS / feed fallback for the
interactive figure in static/js/press-tree.js).

    python tools/generate_press_tree.py                       # default release, attempt 3
    python tools/generate_press_tree.py SLUG --attempt 3 --winner dbaaaa

Reads static/img/press/<slug>/answer-tree.json (leaves, first-request minutes, passing pages) and
writes answer-tree-attempt-<n>.svg beside it. Pure standard library.
"""
import argparse
import json
import math
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
DEPTH = 6
NAVY, GOLD, EMPTY, RED, PASS = '#0A1628', '#C9A84C', '#e9e4d3', '#8b1e1e', '#1e3560'
SIZE, STEPS = 720, 28


STOPS = [(242, 223, 160), (201, 168, 76), (93, 111, 148), (10, 22, 40)]  # pale gold, gold, slate, navy


def ramp(f):
    """Early = pale gold, late = navy; kept in step with ramp() in static/js/press-tree.js."""
    x = max(0.0, min(1.0, f)) * (len(STOPS) - 1)
    i = min(int(x), len(STOPS) - 2)
    a, b, k = STOPS[i], STOPS[i + 1], x - i
    return '#%02x%02x%02x' % tuple(round(u + (v - u) * k) for u, v in zip(a, b))


def build(att, span):
    leaves, seen = att['leaves'], att['seen']
    n = len(leaves)
    levels = []
    for d in range(1, DEPTH + 1):
        arr, cur = [], None
        for i, leaf in enumerate(leaves):
            s = leaf[:d]
            if cur is None or cur['s'] != s:
                cur = {'s': s, 'i0': i, 'i1': i + 1}
                arr.append(cur)
            else:
                cur['i1'] = i + 1
        for node in arr:
            t = seen.get(node['s'])
            node['fill'] = EMPTY if t is None else ramp(round(t / span * STEPS) / STEPS)
        levels.append(arr)
    return levels, n


def pt(cx, r, a):
    return cx + r * math.cos(a), cx + r * math.sin(a)


def arc(cx, ri, ro, a0, a1):
    large = 1 if a1 - a0 > math.pi else 0
    x0, y0 = pt(cx, ro, a0)
    x1, y1 = pt(cx, ro, a1)
    x2, y2 = pt(cx, ri, a1)
    x3, y3 = pt(cx, ri, a0)
    return (f'M{x0:.2f} {y0:.2f}A{ro:.2f} {ro:.2f} 0 {large} 1 {x1:.2f} {y1:.2f}'
            f'L{x2:.2f} {y2:.2f}A{ri:.2f} {ri:.2f} 0 {large} 0 {x3:.2f} {y3:.2f}Z')


def render(data, attempt, winner):
    att = data['att'][str(attempt)]
    levels, n = build(att, data['span'])
    cx, R = SIZE / 2, SIZE / 2 - 40
    r0 = SIZE * 0.06
    th = (R - r0) / DEPTH
    out = [f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {SIZE} {SIZE + 90}" font-family="Arial,Helvetica,sans-serif">',
           f'<rect width="{SIZE}" height="{SIZE + 90}" fill="#fff"/>']
    for di, lv in enumerate(levels):
        ri, ro = r0 + di * th, r0 + di * th + th - 1
        runs = []  # merge neighbouring wedges with one fill into one path
        for node in lv:
            a0 = node['i0'] / n * 2 * math.pi - math.pi / 2
            a1 = node['i1'] / n * 2 * math.pi - math.pi / 2
            if runs and runs[-1][2] == node['fill'] and abs(runs[-1][1] - a0) < 1e-9:
                runs[-1][1] = a1
            else:
                runs.append([a0, a1, node['fill']])
        by = {}
        for a0, a1, fill in runs:
            by.setdefault(fill, []).append(arc(cx, ri, ro, a0, min(a1, a0 + 2 * math.pi - 1e-4)))
        for fill, paths in by.items():
            out.append(f'<path fill="{fill}" d="{"".join(paths)}"/>')
    # passing pages: ticks outside the ring
    for node in levels[-1]:
        sc = att['pass'].get(node['s'])
        t = att['seen'].get(node['s'])
        if sc is None or t is None:
            continue
        win = node['s'] == winner
        a = ((node['i0'] + node['i1']) / 2) / n * 2 * math.pi - math.pi / 2
        x0, y0 = pt(cx, R + 2, a)
        x1, y1 = pt(cx, R + (14 if win else 8), a)
        out.append(f'<line x1="{x0:.1f}" y1="{y0:.1f}" x2="{x1:.1f}" y2="{y1:.1f}" stroke="{RED if win else PASS}" stroke-width="{3 if win else 1.6}"/>')
    # winning path outlined ring by ring
    for d in range(1, DEPTH + 1):
        node = next((x for x in levels[d - 1] if x['s'] == winner[:d]), None)
        if not node or node['s'] not in att['seen']:
            continue
        a0 = node['i0'] / n * 2 * math.pi - math.pi / 2
        a1 = node['i1'] / n * 2 * math.pi - math.pi / 2
        ri, ro = r0 + (d - 1) * th, r0 + (d - 1) * th + th - 1
        out.append(f'<path fill="none" stroke="{RED}" stroke-width="2" d="{arc(cx, ri, ro, a0, a1)}"/>')
    # ring labels down the right-hand spoke
    for d in range(1, DEPTH + 1):
        r = r0 + (d - 0.5) * th
        out.append(f'<text x="{cx + 6:.1f}" y="{cx - r + 4:.1f}" font-size="11" font-weight="700" fill="{NAVY}" stroke="#fff" stroke-width="3" paint-order="stroke">{d}</text>')
    # legend
    gid = 'g'
    stops = ''.join(f'<stop offset="{i * 10}%" stop-color="{ramp(i / 10)}"/>' for i in range(11))
    out.append(f'<defs><linearGradient id="{gid}">{stops}</linearGradient></defs>')
    out.append(f'<text x="40" y="{SIZE + 14}" font-size="12" font-weight="700" fill="#444">First requested</text>')
    out.append(f'<rect x="40" y="{SIZE + 22}" width="220" height="10" fill="url(#{gid})"/>')
    out.append(f'<text x="40" y="{SIZE + 48}" font-size="12" fill="#444">Oct. 25, 09:48 UTC</text>')
    out.append(f'<text x="260" y="{SIZE + 48}" text-anchor="end" font-size="12" fill="#444">Oct. 26, 01:24</text>')
    out.append(f'<rect x="300" y="{SIZE + 22}" width="14" height="10" fill="{EMPTY}" stroke="#ccc"/>')
    out.append(f'<text x="320" y="{SIZE + 31}" font-size="12" fill="#444">not requested</text>')
    out.append(f'<line x1="430" y1="{SIZE + 21}" x2="430" y2="{SIZE + 33}" stroke="{PASS}" stroke-width="2"/>')
    out.append(f'<text x="438" y="{SIZE + 31}" font-size="12" fill="#444">passing page</text>')
    out.append(f'<line x1="540" y1="{SIZE + 21}" x2="540" y2="{SIZE + 33}" stroke="{RED}" stroke-width="3"/>')
    out.append(f'<text x="548" y="{SIZE + 31}" font-size="12" fill="#444">certificate, {winner}</text>')
    out.append(f'<text x="{cx:.0f}" y="{SIZE + 78}" text-anchor="middle" font-size="12" fill="#444">Attempt {attempt}: rings are answers 1 to 6, wedges are choices</text>')
    out.append('</svg>')
    return '\n'.join(out)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('slug', nargs='?', default='compliance-center-first-results')
    ap.add_argument('--attempt', type=int, default=3)
    ap.add_argument('--winner', default='dbaaaa')
    args = ap.parse_args()
    d = ROOT / 'acpwb' / 'static' / 'img' / 'press' / args.slug
    data = json.loads((d / 'answer-tree.json').read_text())
    svg = render(data, args.attempt, args.winner)
    out = d / f'answer-tree-attempt-{args.attempt}.svg'
    out.write_text(svg)
    print(f'wrote {out} ({len(svg) / 1024:.0f} KB)')


if __name__ == '__main__':
    main()
