#!/usr/bin/env python3
"""
Standalone OG image generator — no Django required.

Install:
    pip install torch torchvision --index-url https://download.pytorch.org/whl/cpu
    pip install diffusers transformers accelerate pillow

Usage:
    python generate_og_images.py                        # all slugs
    python generate_og_images.py compensation-review    # specific slugs
    python generate_og_images.py --force                # re-generate existing
    python generate_og_images.py --steps 30             # inference steps (default 30 for SDXL)
    python generate_og_images.py --dry-run              # preview prompts only
"""
import argparse
import hashlib
import os
import re
import sys
import tempfile
from pathlib import Path

HERE = Path(__file__).resolve().parent

# Add acpwb package to path so we can import archive_data slugs
sys.path.insert(0, str(HERE / 'acpwb'))


# ── Word-level visual lookup ───────────────────────────────────────────────────
# Each meaningful slug word maps to a short visual description.
# _slug_to_scene picks the top 2 matching words and combines their visuals,
# giving each slug its own distinct scene rather than collapsing into broad buckets.

_WORD_VISUALS = {
    # Pay & cash comp
    'compensation':  'executive boardroom, long polished table, evening city lights',
    'salary':        'formal meeting room, structured corporate interior',
    'pay':           'executive office, polished desk, afternoon light',
    'wage':          'functional industrial office, organized workspace',
    'merit':         'bright performance review room, minimal and focused',
    'bonus':         'elegant executive suite, warm celebratory light',
    'incentive':     'modern motivational workspace, dynamic atmosphere',
    'deferred':      'quiet corner office at dusk, contemplative mood',
    'pension':       'serene professional office, warm afternoon sun',
    'retirement':    'calm corner office, late golden hour light',
    'allowance':     'organized administrative office, clean and structured',
    # Equity
    'equity':        'panoramic corner office, golden hour cityscape',
    'stock':         'trading floor interior, glowing terminal screens',
    'option':        'financial office with sweeping city view',
    'grant':         'formal signing room, dark polished wood table',
    'vesting':       'corner office at sunrise, long shadows',
    'rsu':           'modern equity office, glass and steel',
    'dilution':      'financial district glass tower, reflective facade',
    'ownership':     'commanding executive suite, city panorama',
    # Benefits
    'benefit':       'bright open office, welcoming and modern',
    'benefits':      'bright airy workspace, plants along window wall',
    'healthcare':    'clean clinical space, white walls, chrome accents',
    'health':        'wellness-oriented office, greenery, natural light',
    'medical':       'sterile clinical interior, white and chrome, precise',
    'dental':        'bright clinical room, clean white surfaces',
    'wellness':      'serene light-filled space, indoor plants, calm',
    'mental':        'quiet soft-lit office, diffused warm light',
    'insurance':     'traditional corporate office, dark polished wood',
    'hsa':           'clean contemporary open office',
    'hra':           'modern organized HR suite',
    'ichra':         'modern benefits office, open floor plan',
    'cobra':         'formal HR office, organized and institutional',
    # Retirement & long-term
    '401k':          'calm professional office, warm focused lamp',
    'pension':       'serene executive office, late afternoon light',
    'actuarial':     'analytical office, rows of files, focused lamp light',
    # Technology
    'technology':    'dark server room, blue LED glow, cable rows',
    'tech':          'modern tech workspace, ambient blue accent light',
    'software':      'developer workspace, multiple monitors, dark room',
    'digital':       'sleek workspace, ambient screen glow',
    'cloud':         'data center hall, dramatic server rows, cool light',
    'microservices': 'network operations center, wall of glowing screens',
    'migration':     'server room, dense cable management, blue light',
    'cybersecurity': 'dark security operations center, red and blue displays',
    'security':      'secure operations room, controlled dim lighting',
    'phishing':      'dark security operations center, analyst at glowing monitors',
    'simulation':    'empty corporate training room, projector light, rows of chairs',
    'training':      'empty corporate training room, projector glow, clean rows of seats',
    'awareness':     'bright open office common area, natural light, collaborative space',
    'testing':       'dark lab environment, multiple screens, focused technical workspace',
    'exercise':      'conference room setup, formal atmosphere, city view beyond',
    'infrastructure':'industrial network facility, dramatic overhead light',
    'platform':      'modern tech office, open collaborative floor plan',
    'data':          'data center corridor, racks of servers, blue glow',
    'saas':          'modern SaaS office, glass and open space',
    'hris':          'modern HR technology office, organized screens',
    'hcm':           'human capital management hub, modern and open',
    'api':           'developer workspace, terminal on dark monitor, ambient blue glow, keyboard foreground',
    'integration':   'server room, dense cable bundles, systematic rack rows, cool blue light',
    'webhook':       'network operations center, blinking status lights, dark console room',
    'endpoint':      'data center corridor, server rows, cool blue LED strips',
    'sdk':           'solo developer at multiple monitors, dark room, code on screens',
    'documentation': 'organized technical workspace, focused desk lamp, reference binders, clean desk',
    'docs':          'clean organized technical office, reference materials, focused lamp',
    'workflow':      'open modern workspace, systematic layout, organized flow',
    'automation':    'industrial control room, blinking panels, systematic monitoring',
    'pipeline':      'data center hall, organized rack infrastructure, dramatic overhead light',
    'batch':         'server room at night, minimal lighting, humming equipment',
    'sync':          'network operations center, synchronized displays, dark environment',
    'async':         'quiet dark server room, single status light, isolated terminal',
    # Leadership & C-suite
    'executive':     'penthouse corner office, panoramic city view',
    'ceo':           'top-floor office, floor-to-ceiling windows, twilight',
    'cfo':           'financial executive suite, commanding city skyline',
    'coo':           'operational command center, monitoring displays',
    'cto':           'technology executive office, modern and minimal',
    'chro':          'people leadership suite, warm and open interior',
    'president':     'formal presidential office, wood paneling, gravitas',
    'board':         'empty boardroom, long mahogany table, evening',
    'director':      'formal conference room, serious corporate atmosphere',
    'committee':     'formal committee room, dark oak table, recessed light',
    'leadership':    'lone silhouette at panoramic boardroom window',
    'management':    'formal meeting room, long conference table',
    'governance':    'boardroom with institutional gravitas, dark wood',
    'officer':       'formal executive corridor, polished and serious',
    # Compliance & legal
    'compliance':    'regulated institutional office, dark formal interior',
    'audit':         'dark-paneled audit room, serious and formal',
    'regulatory':    'governmental architecture, stone columns, formal',
    'legal':         'law library, tall dark wood shelves, reading lamp',
    'contract':      'formal signing room, polished table',
    'counsel':       'law library corridor, single warm lamp ahead',
    'litigation':    'formal law office, serious controlled atmosphere',
    'policy':        'institutional office, formal and deliberate',
    'sox':           'compliance operations room, formal interior',
    'proxy':         'formal shareholder meeting room, institutional',
    'disclosure':    'formal regulatory office, controlled and institutional',
    'transparency':  'bright glass-walled open office, visible and clear',
    # Industry verticals
    'pharmaceutical':'clean room laboratory, white walls, chrome equipment',
    'pharma':        'sterile laboratory interior, clean and precise',
    'energy':        'industrial control room, monitoring displays, dramatic',
    'mining':        'rugged industrial facility, heavy equipment silhouette',
    'aerospace':     'vast aircraft hangar, dramatic overhead lighting',
    'defense':       'formal dark institutional corridor, serious mood',
    'manufacturing': 'industrial facility, machinery silhouettes at dusk',
    'retail':        'sleek modern retail space, minimalist curated display',
    'hospitality':   'luxury hotel lobby, warm golden accent lighting',
    'banking':       'marble bank lobby, brass fixtures, formal gravitas',
    'finance':       'trading floor interior, glowing terminal screens',
    'investment':    'glass-walled investment office, city panorama',
    'nonprofit':     'bright community office, plants, warm natural light',
    'education':     'bright modern learning space, open and airy',
    'media':         'creative studio, exposed brick, warm amber light',
    'music':         'recording studio, mixing console, moody shadows',
    'royalty':       'vintage recording studio, rich warm dramatic light',
    'entertainment': 'creative production studio, dynamic warm atmosphere',
    'publishing':    'editorial office, focused desk lamp, organized',
    'construction':  'steel building frame at sunset, dramatic sky',
    'estate':        'sleek property lobby, marble and polished glass',
    'property':      'modern architectural exterior, glass and steel',
    'insurance':     'traditional corporate office, formal dark wood',
    'veterinary':    'bright clean clinical space, white and chrome',
    'sports':        'athletic facility interior, dramatic overhead light',
    'arts':          'gallery space, dramatic spotlights, white walls',
    'museum':        'museum gallery, exhibition spot lighting, dramatic',
    'hospitality':   'luxury hotel lobby, golden accent lighting',
    'staffing':      'modern professional recruitment office, bright',
    'restaurant':    'elegant dining establishment, warm intimate light',
    # Talent & HR
    'talent':        'dynamic modern workspace, collaborative open design',
    'hiring':        'bright welcoming HR suite, open atmosphere',
    'recruitment':   'modern open recruitment office, bright and inviting',
    'workforce':     'open-plan office, workstations at dusk, ambient glow',
    'employee':      'warm collaborative workspace, natural light',
    'retention':     'comfortable inviting modern office, warm atmosphere',
    'engagement':    'vibrant energetic office, bright open space',
    'culture':       'creative open workspace, warm and welcoming',
    'diversity':     'vibrant bright office with open natural light',
    'inclusion':     'open collaborative workspace, bright and warm',
    # Analysis & type
    'performance':   'minimal focused review room, sleek and deliberate',
    'benchmark':     'analyst workstation, multiple screens, dark office',
    'survey':        'research workspace, organized and methodical',
    'analysis':      'dark analytical workspace, ambient screen glow',
    'assessment':    'formal evaluation room, structured atmosphere',
    'strategy':      'war room, panoramic windows, commanding view',
    'planning':      'conference room with sweeping city panorama',
    'forecast':      'dark workspace, ambient monitor glow, focused',
    'research':      'organized research office, focused and methodical',
    'quarterly':     'end-of-quarter boardroom, documents, evening lights',
    'annual':        'year-end executive meeting room, evening city lights',
    'international': 'cosmopolitan executive office, global city panorama',
    'global':        'world-facing executive suite, expansive city view',
    'expat':         'international executive lounge, global city skyline',
    'mobility':      'modern international office, cosmopolitan atmosphere',
    'remote':        'clean home office setup, bright minimal interior',
    'hybrid':        'flexible modern workspace, open and adaptable',
    # Pay equity & structure
    'equity':        'panoramic corner office, golden hour cityscape',
    'compa':         'structured compensation office, organized and formal',
    'broadband':     'modern structured office, systematic interior',
    'grade':         'organized HR office, methodical atmosphere',
    'range':         'formal compensation review room',
    'structure':     'architectural interior, geometric and deliberate',
    # PERCH Conference
    'perch':        'grand convention center ballroom, rows of conference seating, dramatic event lighting',
    'conference':   'large professional convention hall, sweeping interior, organized event setup',
    'venue':        'grand convention center atrium, polished marble lobby, architectural interior',
    'dinner':       'elegant supper club dining room, white tablecloths, warm amber candlelight',
    'speakers':     'conference stage with podium, dramatic spotlight, auditorium seating in background',
    'schedule':     'large conference hall, professional event atmosphere, organized session layout',
    'sponsors':     'corporate event space, branded exhibition hall, professional partnership atmosphere',
    'register':     'modern event registration lobby, organized welcome desk, bright professional interior',
    'about':        'established convention hall, institutional gravitas, warm even lighting',
    'milwaukee':    'lakefront convention center, dramatic modern interior, evening event lighting',
    'wisconsin':    'midwestern professional convention facility, organized and polished interior',
    # M&A / finance events
    'acquisition':   'formal deal-signing room, dark polished table',
    'merger':        'executive negotiation suite, serious atmosphere',
    'ipo':           'financial district trading floor, anticipatory energy',
    'spac':          'modern financial office, glass and forward-looking',
    'private':       'private equity suite, exclusive and refined',
    'capital':       'financial district office, commanding city view',
    'fund':          'investment management suite, panoramic view',
    'portfolio':     'investment office, city panorama, polished',
}

# Words too generic to anchor a scene on their own
_SKIP_WORDS = {  # noqa: E241
    'a', 'an', 'the', 'and', 'or', 'for', 'of', 'in', 'at', 'to', 'by', 'on',
    'with', 'from', 'into', 'as', 'is', 'are', 'was', 'be', 'has', 'have',
    'this', 'that', 'its', 'their', 'our', 'your', 'new', 'old', 'key', 'top',
    'best', 'full', 'total', 'base', 'core', 'main', 'high', 'low', 'long',
    'short', 'joint', 'lead', 'role', 'type', 'rate', 'cost', 'level', 'year',
    'review', 'report', 'study', 'guide', 'overview', 'summary', 'update',
    'analysis', 'assessment', 'findings', 'output', 'results', 'approach', 'vs',
    'process', 'model', 'system', 'program', 'plan', 'design', 'framework',
    'draft', 'final', 'revised', 'current', 'second',
    'standard', 'criteria', 'elements', 'principles', 'practices', 'factors',
}

_STYLE_SUFFIX = ', corporate photography, photorealistic, 8k, tack sharp, navy and gold palette, f/8 aperture, everything in focus'

# FLUX ignores negative prompts (flow matching, no CFG); used only for SD models.
_NEGATIVE_PROMPT = (
    'watermark, logo, document, slide, '
    'poster, signage, caption, user interface, website, screenshot, dashboard, '
    'stairs, staircase, escalator, surreal, '
    'bokeh, shallow depth of field, blurry background, out of focus, '
    'cartoon, anime, low quality, blurry, distorted, ugly, nsfw'
)

OG_WIDTH, OG_HEIGHT = 1200, 630
# Use dimensions with a closer aspect ratio to 1200x630, divisible by 16.
_GEN_WIDTH, _GEN_HEIGHT = 1024, 544

# Hardcoded overrides for slugs where word-level matching produces bad results.
# Tuple of (subject, scene) — replaces both the subject and scene for that slug.
_SLUG_OVERRIDES = {
    'perch-conference':          ('PERCH annual conference', 'grand convention center ballroom, rows of conference seating, dramatic event lighting'),
    'perch-conference-speakers': ('conference keynote stage', 'large auditorium stage, dramatic spotlight on podium, rows of professional seating'),
    'perch-conference-schedule': ('professional conference hall', 'large convention hall, multiple concurrent session rooms, organized event atmosphere'),
    'perch-conference-venue':    ('convention center venue', 'grand convention center atrium, polished marble lobby, dramatic architectural interior'),
    'perch-conference-about':    ('professional conference history', 'established convention hall, institutional gravitas, warm even lighting, long corridor'),
    'perch-conference-dinner':   ('annual conference dinner', 'elegant Wisconsin supper club, white tablecloths, warm amber candlelight, intimate dining room'),
    'perch-conference-sponsors': ('corporate conference sponsors', 'professional exhibition hall, branded sponsor displays, corporate partnership atmosphere'),
    'perch-conference-register': ('conference registration', 'modern event registration lobby, welcoming desk, organized bright professional interior'),
}


def _slug_words(slug):
    """Extract meaningful words from a slug, most specific first."""
    clean = re.sub(r'-\d{3,}$', '', slug)
    return [w for w in clean.split('-') if w not in _SKIP_WORDS and len(w) > 2]


def _slug_to_scene(slug):
    """Build a scene by combining visuals for the slug's most distinctive words."""
    for word in _slug_words(slug):
        if word in _WORD_VISUALS:
            return _WORD_VISUALS[word]
    return None


def _slug_to_subject(slug):
    clean = re.sub(r'-\d{3,}$', '', slug)
    return clean.replace('-', ' ')


def _build_prompt(slug):
    if slug in _SLUG_OVERRIDES:
        subject, scene = _SLUG_OVERRIDES[slug]
        return f'{subject}, {scene}{_STYLE_SUFFIX}'
    subject = _slug_to_subject(slug)
    scene = _slug_to_scene(slug)
    if scene:
        return f'{subject}, {scene}{_STYLE_SUFFIX}'
    return f'{subject}{_STYLE_SUFFIX}'


def _slug_to_seed(slug):
    return int(hashlib.md5(f'og_{slug}'.encode()).hexdigest(), 16) % (2 ** 32)


def _is_flux(model_id):
    return 'flux' in model_id.lower()


def _pick_device(torch, flux=False):
    if torch.cuda.is_available():
        # FLUX is trained in bfloat16; SD models use float16
        return 'cuda', torch.bfloat16 if flux else torch.float16
    if getattr(torch.backends, 'mps', None) and torch.backends.mps.is_available():
        return 'mps', torch.float32  # float16/bfloat16 both produce artifacts on Apple Silicon
    return 'cpu', torch.float32


def _save_atomic(img, path):
    fd, tmp = tempfile.mkstemp(suffix='.jpg', dir=path.parent)
    try:
        os.close(fd)
        img.save(tmp, 'JPEG', quality=88, optimize=True)
        os.replace(tmp, path)
    except Exception:
        try:
            os.unlink(tmp)
        except OSError:
            pass
        raise


def main():
    parser = argparse.ArgumentParser(description='Generate OG banner images from archive slugs')
    parser.add_argument('slugs', nargs='*', help='Slugs to generate (default: all _ARCHIVE_SLUGS)')
    parser.add_argument('--output-dir', default=str(HERE / 'acpwb' / 'media' / 'og'))
    parser.add_argument('--model', default='stabilityai/stable-diffusion-xl-base-1.0')
    parser.add_argument('--steps', type=int, default=30)
    parser.add_argument('--hf-token', default=os.environ.get('HF_TOKEN'), help='Hugging Face token (or set HF_TOKEN env var)')
    parser.add_argument('--force', action='store_true')
    parser.add_argument('--dry-run', action='store_true')
    args = parser.parse_args()

    try:
        import torch
        from diffusers import AutoPipelineForText2Image
        from diffusers.utils import logging as diffusers_logging
        from PIL import Image
    except ImportError as e:
        sys.exit(
            f'Missing dependency: {e}\n\n'
            'Install with:\n'
            '  pip install torch torchvision --index-url https://download.pytorch.org/whl/cpu\n'
            '  pip install diffusers transformers accelerate pillow\n'
        )

    out_dir = Path(args.output_dir)
    out_dir.mkdir(parents=True, exist_ok=True)

    if args.slugs:
        slugs = args.slugs
    else:
        from apps.honeypot.archive_data import _ARCHIVE_SLUGS
        slugs = list(_ARCHIVE_SLUGS)

    print(f'Slugs to process: {len(slugs)}')

    import random
    random.shuffle(slugs)

    if not args.force:
        pending = [s for s in slugs if not (out_dir / f'{s}.jpg').exists()]
        skipped = len(slugs) - len(pending)
        if skipped:
            print(f'  Skipping {skipped} already generated (use --force to overwrite)')
        slugs = pending

    if not slugs:
        print('Nothing to generate.')
        return

    if args.dry_run:
        for slug in slugs:
            print(f'  {slug}\n    seed={_slug_to_seed(slug)}\n    {_build_prompt(slug)}\n')
        return

    flux = _is_flux(args.model)
    device, dtype = _pick_device(torch, flux=flux)
    print(f'Loading {args.model} on {device} ({dtype}) …')

    # Suppress verbose model loading warnings
    diffusers_logging.set_verbosity_error()

    pipe = AutoPipelineForText2Image.from_pretrained(
        args.model, torch_dtype=dtype, use_safetensors=True,
        token=args.hf_token or None,
    ).to(device)

    pipe.set_progress_bar_config(disable=True)
    print('Model loaded.\n')

    errors = 0
    for i, slug in enumerate(slugs, 1):
        out_path = out_dir / f'{slug}.jpg'
        prompt = _build_prompt(slug)
        seed = _slug_to_seed(slug)

        print(f'[{i}/{len(slugs)}] {slug} ', end='', flush=True)
        try:
            generator = torch.Generator(device=device).manual_seed(seed)
            gen_kwargs = dict(
                prompt=prompt,
                num_inference_steps=args.steps,
                width=_GEN_WIDTH,
                height=_GEN_HEIGHT,
                generator=generator,
            )
            if flux:
                # FLUX.1-schnell uses flow matching — no CFG, no negative prompt
                gen_kwargs['guidance_scale'] = 0.0
            else:
                gen_kwargs['negative_prompt'] = _NEGATIVE_PROMPT
                gen_kwargs['guidance_scale'] = 9.0  # higher CFG = more prompt-adherent for SDXL
            result = pipe(**gen_kwargs)
            img = result.images[0]
            if img.size != (OG_WIDTH, OG_HEIGHT):
                img = img.resize((OG_WIDTH, OG_HEIGHT), Image.LANCZOS)
            _save_atomic(img, out_path)
            print('✓')
        except Exception as exc:
            print(f'FAILED: {exc}')
            errors += 1

    print(f'\nDone. {len(slugs) - errors} generated, {errors} errors → {out_dir}')


if __name__ == '__main__':
    main()
