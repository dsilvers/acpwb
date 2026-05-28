#!/usr/bin/env python3
"""
Standalone OG image generator for press releases — no Django required.

Install:
    pip install torch torchvision --index-url https://download.pytorch.org/whl/cpu
    pip install diffusers transformers accelerate pillow

Usage:
    python generate_og_images_press.py                        # all press releases
    python generate_og_images_press.py 200-million-pages-served # specific slugs
    python generate_og_images_press.py --force                # re-generate existing
    python generate_og_images_press.py --steps 30             # inference steps (default 30 for SDXL)
    python generate_og_images_press.py --dry-run              # preview prompts only
"""
import argparse
import hashlib
import os
import re
import sys
import tempfile
from pathlib import Path

HERE = Path(__file__).resolve().parent

# Add acpwb package to path so we can import from apps
sys.path.insert(0, str(HERE / 'acpwb'))


# ── Word-level visual lookup for press release headlines ───────────────────────
# Each meaningful word maps to a short visual description.

_WORD_VISUALS = {
    # Pay & cash comp
    'compensation':  'executive boardroom, long polished table, evening city lights',
    'salary':        'formal meeting room, structured corporate interior',
    'pay':           'executive office, polished desk, afternoon light',
    'ceo':           'top-floor office, floor-to-ceiling windows, twilight',
    'executive':     'penthouse corner office, panoramic city view',
    'bonus':         'elegant executive suite, warm celebratory light',
    'incentive':     'modern motivational workspace, dynamic atmosphere',
    # Equity
    'equity':        'panoramic corner office, golden hour cityscape',
    'stock':         'trading floor interior, glowing terminal screens',
    'option':        'financial office with sweeping city view',
    # Technology & AI
    'ai':            'dark server room, glowing blue geometric patterns, abstract data streams',
    'artificial':    'abstract network of glowing nodes, dark background',
    'intelligence':  'glowing neural network visualization, complex and interconnected',
    'model':         'abstract 3d wireframe of a complex algorithm, blue and gold',
    'models':        'multiple glowing wireframe structures, data center background',
    'data':          'data center corridor, racks of servers, cool blue glow',
    'platform':      'modern tech office, open collaborative floor plan, glass walls',
    'digital':       'sleek workspace, ambient screen glow, minimalist',
    'technology':    'dark server room, blue LED glow, rows of cables',
    'cybersecurity': 'dark security operations center, red and blue displays',
    'software':      'developer workspace, multiple monitors, dark room, code on screens',
    # Corporate & Governance
    'board':         'empty boardroom, long mahogany table, evening city lights',
    'governance':    'boardroom with institutional gravitas, dark wood, recessed lighting',
    'ceo':           'top-floor office, floor-to-ceiling windows, twilight city view',
    'officer':       'formal executive corridor, polished marble floors, serious atmosphere',
    'leadership':    'empty panoramic boardroom window with commanding city view, no people',
    'management':    'formal meeting room, long conference table, city skyline',
    'patent':        'formal institutional office, framed documents on wall, focused desk lamp',
    'award':         'elegant awards ceremony stage, spotlights, empty podium',
    'recognition':   'formal event space, stage with spotlights, celebratory mood',
    # Workforce & People
    'workforce':     'open-plan office, workstations at dusk, ambient screen glow',
    'employee':      'warm collaborative workspace, natural light, modern furniture',
    'talent':        'dynamic modern workspace, collaborative open design',
    'hiring':        'bright welcoming HR suite, open atmosphere, glass partitions',
    'people':        'vibrant energetic office, bright open space, collaborative groups',
    'engagement':    'energetic all-hands meeting space, large screen, modern seating',
    'culture':       'creative open workspace, warm and welcoming, exposed brick',
    'team':          'collaborative project room, whiteboards, city view',
    # General Business & Milestones
    'launch':        'modern product launch stage, dramatic lighting, empty podium',
    'launches':      'sleek event stage, spotlights, anticipation',
    'milestone':     'abstract geometric structure marking a point in a timeline, gold and navy',
    'growth':        'upward-trending abstract light graph, dark background',
    'expansion':     'architectural model of a growing city skyline, dramatic lighting',
    'report':        'organized research office, focused desk lamp, data visualizations on screen',
    'survey':        'research workspace, organized data charts, methodical atmosphere',
    'study':         'academic research library, single reading lamp, quiet mood',
    'analysis':      'dark analytical workspace, ambient screen glow, complex charts',
    'framework':     'architectural blueprint, clean lines, precise and structured',
    'initiative':    'project war room, large screens, strategic maps, focused energy',
    'partnership':   'formal signing room, two empty chairs at a polished table',
    'strategy':      'command center with panoramic city view, strategic displays',
}

# Words too generic to anchor a scene on their own
_SKIP_WORDS = {
    'a', 'an', 'the', 'and', 'or', 'for', 'of', 'in', 'at', 'to', 'by', 'on',
    'with', 'from', 'into', 'as', 'is', 'are', 'was', 'be', 'has', 'have',
    'its', 'new', 'key', 'top', 'best', 'full', 'total', 'vs',
    'acpwb', 'american', 'corporation', 'public', 'well', 'being',
    'announces', 'receives', 'surpasses', 'becomes', 'named', 'appointed',
    'issues', 'since', 'after', 'for', 'over', 'about',
}

_STYLE_SUFFIX = ', corporate photography, photorealistic, 8k, tack sharp, navy and gold palette, f/8 aperture, everything in focus'
_FLUX_STYLE_SUFFIX = ', no people, no faces, empty interior, architecture only, corporate photography, photorealistic, navy gold palette, 8k'

_AI_SYSTEM = (
    'You are a visual art director for corporate photography. '
    'Given a press release, write a single brief scene description (15–30 words) '
    'for a photorealistic OG banner image (1200×630px). '
    'Rules: no people, faces, text, logos, or signs — only architectural or environmental settings. '
    'Navy and gold color palette, corporate photography style, f/8 aperture, everything in focus. '
    'Output only the scene description with no preamble or explanation.'
)


def _ai_build_prompt(pr, client, model, flux=False):
    body_text = '\n\n'.join(pr.get('body', []))[:3000]
    response = client.chat.completions.create(
        model=model,
        max_tokens=80,
        messages=[
            {'role': 'system', 'content': _AI_SYSTEM},
            {'role': 'user', 'content': f'Headline: {pr["headline"]}\n\n{body_text}'},
        ],
    )
    scene = response.choices[0].message.content.strip().rstrip('.')
    suffix = _FLUX_STYLE_SUFFIX if flux else _STYLE_SUFFIX
    return f'{scene}{suffix}'

# FLUX ignores negative prompts (flow matching, no CFG); used only for SD models.
_NEGATIVE_PROMPT = (
    'people, person, man, woman, figure, silhouette, face, portrait, '
    'watermark, logo, document, slide, poster, signage, caption, text, '
    'user interface, website, screenshot, dashboard, '
    'stairs, staircase, escalator, surreal, '
    'bokeh, shallow depth of field, blurry background, out of focus, '
    'cartoon, anime, low quality, blurry, distorted, ugly, nsfw'
)

OG_WIDTH, OG_HEIGHT = 1200, 630
# Use dimensions with a closer aspect ratio to 1200x630, divisible by 16.
_GEN_WIDTH, _GEN_HEIGHT = 1024, 544


def _headline_words(headline):
    """Extract meaningful words from a headline, most specific first."""
    clean = re.sub(r'[^\w\s-]', '', headline.lower())
    return [w for w in clean.split() if w not in _SKIP_WORDS and len(w) > 2]


def _headline_to_scene(headline):
    """Build a scene by combining visuals for the headline's most distinctive words."""
    for word in _headline_words(headline):
        if word in _WORD_VISUALS:
            return _WORD_VISUALS[word]
    return None


def _build_prompt(headline, flux=False):
    suffix = _FLUX_STYLE_SUFFIX if flux else _STYLE_SUFFIX
    scene = _headline_to_scene(headline)
    if scene:
        return f'{headline}, {scene}{suffix}'
    return f'{headline}{suffix}'


def _slug_to_seed(slug):
    return int(hashlib.md5(f'og_press_{slug}'.encode()).hexdigest(), 16) % (2 ** 32)


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
    parser = argparse.ArgumentParser(description='Generate OG banner images for press releases')
    parser.add_argument('slugs', nargs='*', help='Press release slugs to generate (default: all)')
    parser.add_argument('--output-dir', default=str(HERE / 'acpwb' / 'media' / 'og'))
    parser.add_argument('--model', default='stabilityai/stable-diffusion-xl-base-1.0')
    parser.add_argument('--steps', type=int, default=30)
    parser.add_argument('--hf-token', default=os.environ.get('HF_TOKEN'), help='Hugging Face token (or set HF_TOKEN env var)')
    parser.add_argument('--force', action='store_true')
    parser.add_argument('--dry-run', action='store_true')
    parser.add_argument('--ai-prompt', action='store_true', help='Use OpenAI API to generate prompts from press release body text')
    parser.add_argument('--ai-model', default='gpt-4o-mini', help='OpenAI model for prompt generation (default: gpt-4o-mini)')
    parser.add_argument('--quantize', action='store_true', help='Quantize transformer+T5 to qint8 via optimum-quanto (reduces RAM ~50%%)')
    args = parser.parse_args()

    ai_client = None
    if args.ai_prompt:
        try:
            import openai
            ai_client = openai.OpenAI()
        except ImportError:
            sys.exit('Missing openai package for --ai-prompt. Install with: pip install openai')

    try:
        import torch
        from diffusers import AutoPipelineForText2Image
        from diffusers.utils import logging as diffusers_logging
        from PIL import Image
        from apps.public.press_data import _PRESS_RELEASES
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
        all_press_releases = {pr['slug']: pr for pr in _PRESS_RELEASES}
        press_releases_to_process = []
        for slug in args.slugs:
            if slug in all_press_releases:
                press_releases_to_process.append(all_press_releases[slug])
            else:
                print(f'Warning: slug "{slug}" not found in _PRESS_RELEASES.')
    else:
        press_releases_to_process = _PRESS_RELEASES

    print(f'Press releases to process: {len(press_releases_to_process)}')

    import random
    random.shuffle(press_releases_to_process)

    if not args.force:
        pending = [pr for pr in press_releases_to_process if not (out_dir / f'{pr["slug"]}.jpg').exists()]
        skipped = len(press_releases_to_process) - len(pending)
        if skipped:
            print(f'  Skipping {skipped} already generated (use --force to overwrite)')
        press_releases_to_process = pending

    if not press_releases_to_process:
        print('Nothing to generate.')
        return

    flux = _is_flux(args.model)

    if args.dry_run:
        for pr in press_releases_to_process:
            if ai_client:
                prompt = _ai_build_prompt(pr, ai_client, args.ai_model, flux=flux)
            else:
                prompt = _build_prompt(pr['headline'], flux=flux)
            print(f'  {pr["slug"]}\n    seed={_slug_to_seed(pr["slug"])}\n    {prompt}\n')
        return

    device, dtype = _pick_device(torch, flux=flux)
    print(f'Loading {args.model} on {device} ({dtype}) …')

    # Suppress verbose model loading warnings
    diffusers_logging.set_verbosity_error()

    pipe = AutoPipelineForText2Image.from_pretrained(
        args.model, torch_dtype=dtype, use_safetensors=True,
        token=args.hf_token or None,
    )

    if args.quantize:
        try:
            from optimum.quanto import freeze, qint8, quantize as quanto_quantize
        except ImportError:
            sys.exit('Missing optimum-quanto. Install with: pip install optimum-quanto')
        print('Quantizing transformer to qint8 …')
        quanto_quantize(pipe.transformer, weights=qint8)
        freeze(pipe.transformer)
        if hasattr(pipe, 'text_encoder_2') and pipe.text_encoder_2 is not None:
            print('Quantizing T5 text encoder to qint8 …')
            quanto_quantize(pipe.text_encoder_2, weights=qint8)
            freeze(pipe.text_encoder_2)

    pipe.enable_model_cpu_offload()

    pipe.set_progress_bar_config(disable=True)
    print('Model loaded.\n')

    errors = 0
    for i, pr in enumerate(press_releases_to_process, 1):
        slug = pr['slug']
        headline = pr['headline']
        out_path = out_dir / f'{slug}.jpg'
        if ai_client:
            prompt = _ai_build_prompt(pr, ai_client, args.ai_model, flux=flux)
        else:
            prompt = _build_prompt(headline, flux=flux)
        seed = _slug_to_seed(slug)

        print(f'[{i}/{len(press_releases_to_process)}] {slug} ', end='', flush=True)
        try:
            generator = torch.Generator(device='cpu').manual_seed(seed)
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

    print(f'\nDone. {len(press_releases_to_process) - errors} generated, {errors} errors → {out_dir}')


if __name__ == '__main__':
    main()