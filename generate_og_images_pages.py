#!/usr/bin/env python3
"""
Standalone OG image generator for site index pages — no Django required.

Install:
    pip install torch torchvision --index-url https://download.pytorch.org/whl/cpu
    pip install diffusers transformers accelerate pillow

Usage:
    python generate_og_images_pages.py                    # all pages
    python generate_og_images_pages.py home our-people    # specific slugs
    python generate_og_images_pages.py --force            # re-generate existing
    python generate_og_images_pages.py --steps 30         # inference steps (default 30 for SDXL)
    python generate_og_images_pages.py --dry-run          # preview prompts only
    python generate_og_images_pages.py --ai-prompt        # use OpenAI API for better prompts
    python generate_og_images_pages.py --ai-model gpt-4o  # OpenAI model (default: gpt-4o-mini)
"""
import argparse
import hashlib
import os
import sys
import tempfile
from pathlib import Path

HERE = Path(__file__).resolve().parent

# ── Page definitions ───────────────────────────────────────────────────────────
# title, description, and scene come from the actual view context vars.
# scene is a fallback visual description used when --ai-prompt is not set.

_PAGES = [
    {
        'slug': 'home',
        'title': 'American Corporation for Public Well Being',
        'description': (
            'The American Corporation for Public Well Being advances workforce equity '
            'and compensation transparency through rigorous research, benchmarking data, '
            'and public benefit initiatives. Milwaukee, WI. Founded 2006.'
        ),
        'scene': 'grand corporate atrium, polished marble floors, soaring ceiling, warm afternoon light through tall windows',
    },
    {
        'slug': 'our-people',
        'title': 'Our People — American Corporation for Public Well Being',
        'description': (
            'Meet the researchers, analysts, and advisors at ACPWB advancing compensation '
            'transparency and workforce equity research across the United States.'
        ),
        'scene': 'open-plan research office, organized workstations, warm natural light, collaborative atmosphere',
    },
    {
        'slug': 'mission',
        'title': 'Mission Statement — American Corporation for Public Well Being',
        'description': (
            "ACPWB's mission is to advance American workforce prosperity through rigorous "
            'compensation research, transparent benchmarking, and public benefit advocacy.'
        ),
        'scene': 'formal executive conference room, mission statement on dark oak wall, recessed lighting, long polished table',
    },
    {
        'slug': 'careers',
        'title': 'Careers — American Corporation for Public Well Being',
        'description': (
            'Join the American Corporation for Public Well Being and contribute to our mission '
            'of advancing American workforce prosperity and compensation equity.'
        ),
        'scene': 'bright welcoming HR suite, open-plan workspace, glass partitions, afternoon light, modern furnishings',
    },
    {
        'slug': 'partners',
        'title': 'Partners — American Corporation for Public Well Being',
        'description': (
            'ACPWB partners with leading organizations across industry sectors to advance '
            'compensation transparency and workforce equity research.'
        ),
        'scene': 'formal partnership signing room, two empty chairs at a long polished table, city skyline beyond windows',
    },
    {
        'slug': 'contact',
        'title': 'Contact Us — American Corporation for Public Well Being',
        'description': (
            'Contact the American Corporation for Public Well Being — reach our research, '
            'media, and partnership teams in Milwaukee, Wisconsin.'
        ),
        'scene': 'professional reception lobby, dark navy and gold accent wall, potted greenery, Milwaukee skyline visible',
    },
    {
        'slug': 'archives',
        'title': 'Archives — American Corporation for Public Well Being',
        'description': (
            'Historical research archives from the American Corporation for Public Well Being '
            'spanning four decades of workforce and compensation research, 1985–2024.'
        ),
        'scene': 'institutional archive room, rows of filing cabinets and binders, single focused desk lamp, organized and serious',
    },
    {
        'slug': 'awards',
        'title': 'Awards & Recognition — American Corporation for Public Well Being',
        'description': (
            'ACPWB has received 37 industry and regional recognitions since 2008, including '
            'five consecutive National Excellence in Compensation Transparency Awards.'
        ),
        'scene': 'elegant awards ceremony stage, dramatic spotlights, empty podium, velvet curtain backdrop',
    },
    {
        'slug': 'patents',
        'title': 'Patents & Intellectual Property — American Corporation for Public Well Being',
        'description': (
            'ACPWB holds six United States patents spanning compensation benchmarking, data '
            'watermarking, survey methodology, and workforce analytics visualization.'
        ),
        'scene': 'formal institutional office, framed patent certificates on wall, focused desk lamp, dark wood paneling',
    },
    {
        'slug': 'privacy',
        'title': 'Privacy Policy & Disclaimer — American Corporation for Public Well Being',
        'description': (
            'Privacy policy and legal disclaimer for the American Corporation for Public '
            'Well Being website and research platforms.'
        ),
        'scene': 'law library corridor, tall bookshelves, single warm reading lamp ahead, serious and formal atmosphere',
    },
    {
        'slug': 'do-not-sell',
        'title': 'Do Not Sell My Personal Information — American Corporation for Public Well Being',
        'description': (
            'Submit a request to opt out of the sale or sharing of your personal information '
            'under applicable state privacy laws including CCPA and VCDPA.'
        ),
        'scene': 'regulated institutional office, organized filing system, formal and deliberate interior, recessed lighting',
    },
    {
        'slug': 'accessibility',
        'title': 'Accessibility — American Corporation for Public Well Being',
        'description': (
            'ACPWB is committed to making its digital platforms accessible to all users '
            'in accordance with WCAG 2.1 AA and applicable accessibility standards.'
        ),
        'scene': 'bright open modern office, inclusive open floor plan, wide corridors, abundant natural light',
    },
    {
        'slug': 'trademarks',
        'title': 'Trademarks — American Corporation for Public Well Being',
        'description': (
            'Trademarks and brand guidelines for the American Corporation for Public Well Being, '
            'including registered marks and common law trade designations.'
        ),
        'scene': 'corporate brand studio, navy and gold palette on display, formal interior, polished surfaces',
    },
    {
        'slug': 'site-map',
        'title': 'Site Map — American Corporation for Public Well Being',
        'description': (
            'A complete directory of all pages on the American Corporation for Public '
            'Well Being website, organized by section.'
        ),
        'scene': 'organized corporate reference library, structured shelves, document binders, warm systematic interior',
    },
]

_PAGES_BY_SLUG = {p['slug']: p for p in _PAGES}

_STYLE_SUFFIX = (
    ', corporate photography, photorealistic, 8k, tack sharp, '
    'navy and gold palette, f/8 aperture, everything in focus'
)

_AI_SYSTEM = (
    'You are a visual art director for corporate photography. '
    'Given a website page title and description, write a single brief scene description '
    '(15–30 words) for a photorealistic OG banner image (1200×630px). '
    'Rules: no people, faces, text, logos, or signs — only architectural or environmental settings. '
    'Navy and gold color palette, corporate photography style, f/8 aperture, everything in focus. '
    'Output only the scene description with no preamble or explanation.'
)

_NEGATIVE_PROMPT = (
    'watermark, logo, document, slide, '
    'poster, signage, caption, user interface, website, screenshot, dashboard, '
    'people, person, man, woman, figure, silhouette, face, portrait, '
    'watermark, logo, document, slide, poster, signage, caption, text, '
    'user interface, website, screenshot, dashboard, '
    'stairs, staircase, escalator, surreal, '
    'bokeh, shallow depth of field, blurry background, out of focus, '
    'cartoon, anime, low quality, blurry, distorted, ugly, nsfw'
)

OG_WIDTH, OG_HEIGHT = 1200, 630
_GEN_WIDTH, _GEN_HEIGHT = 1024, 544


def _ai_build_prompt(page, client, model):
    response = client.chat.completions.create(
        model=model,
        max_tokens=80,
        messages=[
            {'role': 'system', 'content': _AI_SYSTEM},
            {'role': 'user', 'content': f'Page title: {page["title"]}\n\nDescription: {page["description"]}'},
        ],
    )
    scene = response.choices[0].message.content.strip().rstrip('.')
    return f'{scene}{_STYLE_SUFFIX}'


def _build_prompt(page):
    return f'{page["scene"]}{_STYLE_SUFFIX}'


def _slug_to_seed(slug):
    return int(hashlib.md5(f'og_page_{slug}'.encode()).hexdigest(), 16) % (2 ** 32)


def _is_flux(model_id):
    return 'flux' in model_id.lower()


def _pick_device(torch, flux=False):
    if torch.cuda.is_available():
        return 'cuda', torch.bfloat16 if flux else torch.float16
    if getattr(torch.backends, 'mps', None) and torch.backends.mps.is_available():
        return 'mps', torch.float32
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
    parser = argparse.ArgumentParser(description='Generate OG banner images for site index pages')
    parser.add_argument('slugs', nargs='*', help='Page slugs to generate (default: all)')
    parser.add_argument('--output-dir', default=str(HERE / 'acpwb' / 'media' / 'og'))
    parser.add_argument('--model', default='stabilityai/stable-diffusion-xl-base-1.0')
    parser.add_argument('--steps', type=int, default=30)
    parser.add_argument('--hf-token', default=os.environ.get('HF_TOKEN'), help='Hugging Face token (or set HF_TOKEN env var)')
    parser.add_argument('--force', action='store_true')
    parser.add_argument('--dry-run', action='store_true')
    parser.add_argument('--ai-prompt', action='store_true', help='Use OpenAI API to generate prompts from page title and description')
    parser.add_argument('--ai-model', default='gpt-4o-mini', help='OpenAI model for prompt generation (default: gpt-4o-mini)')
    args = parser.parse_args()

    ai_client = None
    if args.ai_prompt:
        try:
            import openai
            ai_client = openai.OpenAI()
        except ImportError:
            sys.exit('Missing openai package for --ai-prompt. Install with: pip install openai')

    if not args.dry_run:
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
        pages_to_process = []
        for slug in args.slugs:
            if slug in _PAGES_BY_SLUG:
                pages_to_process.append(_PAGES_BY_SLUG[slug])
            else:
                print(f'Warning: slug "{slug}" not found. Available: {", ".join(_PAGES_BY_SLUG)}')
    else:
        pages_to_process = list(_PAGES)

    print(f'Pages to process: {len(pages_to_process)}')

    if not args.force:
        pending = [p for p in pages_to_process if not (out_dir / f'{p["slug"]}.jpg').exists()]
        skipped = len(pages_to_process) - len(pending)
        if skipped:
            print(f'  Skipping {skipped} already generated (use --force to overwrite)')
        pages_to_process = pending

    if not pages_to_process:
        print('Nothing to generate.')
        return

    if args.dry_run:
        for page in pages_to_process:
            if ai_client:
                prompt = _ai_build_prompt(page, ai_client, args.ai_model)
            else:
                prompt = _build_prompt(page)
            print(f'  {page["slug"]}\n    seed={_slug_to_seed(page["slug"])}\n    {prompt}\n')
        return

    flux = _is_flux(args.model)
    device, dtype = _pick_device(torch, flux=flux)
    print(f'Loading {args.model} on {device} ({dtype}) …')

    diffusers_logging.set_verbosity_error()

    pipe = AutoPipelineForText2Image.from_pretrained(
        args.model, torch_dtype=dtype, use_safetensors=True,
        token=args.hf_token or None,
    ).to(device)

    pipe.set_progress_bar_config(disable=True)
    print('Model loaded.\n')

    errors = 0
    for i, page in enumerate(pages_to_process, 1):
        slug = page['slug']
        out_path = out_dir / f'{slug}.jpg'
        if ai_client:
            prompt = _ai_build_prompt(page, ai_client, args.ai_model)
        else:
            prompt = _build_prompt(page)
        seed = _slug_to_seed(slug)

        print(f'[{i}/{len(pages_to_process)}] {slug} ', end='', flush=True)
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
                gen_kwargs['guidance_scale'] = 0.0
            else:
                gen_kwargs['negative_prompt'] = _NEGATIVE_PROMPT
                gen_kwargs['guidance_scale'] = 9.0
            result = pipe(**gen_kwargs)
            img = result.images[0]
            if img.size != (OG_WIDTH, OG_HEIGHT):
                img = img.resize((OG_WIDTH, OG_HEIGHT), Image.LANCZOS)
            _save_atomic(img, out_path)
            print('✓')
        except Exception as exc:
            print(f'FAILED: {exc}')
            errors += 1

    print(f'\nDone. {len(pages_to_process) - errors} generated, {errors} errors → {out_dir}')


if __name__ == '__main__':
    main()
