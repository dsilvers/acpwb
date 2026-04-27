"""
Generate Open Graph banner images (1200×630 JPEG) for archive pages.

Each image is derived deterministically from the slug — same slug always
produces the same image. Images are saved to staticfiles/og/.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
INSTALL (one-time):

    pip install diffusers transformers accelerate pillow

    # PyTorch — pick ONE:
    pip install torch torchvision               # Mac (MPS) or CPU
    pip install torch torchvision --index-url https://download.pytorch.org/whl/cu121
                                               # NVIDIA GPU (CUDA 12.1)

Model (~6 GB) downloads automatically on first run to ~/.cache/huggingface/
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

USAGE:

    # Generate for every base slug in _ARCHIVE_SLUGS
    python manage.py generate_og_images

    # One or more specific slugs
    python manage.py generate_og_images compensation-committee-review market-analysis-2024

    # Overwrite existing images
    python manage.py generate_og_images --force

    # Fewer steps = faster but rougher (default 4; minimum 1)
    python manage.py generate_og_images --steps 2

    # Custom output dir (default: staticfiles/og/)
    python manage.py generate_og_images --output-dir /path/to/dir
"""
import hashlib
import os
import re as _re
import sys
import tempfile
from pathlib import Path

from django.core.management.base import BaseCommand, CommandError


# ── Keyword → scene mapping ────────────────────────────────────────────────────
# Keywords are matched against words in the de-hyphenated slug.  First match wins.

_KEYWORD_SCENES = [
    (
        {'compensation', 'salary', 'pay', 'wage', 'cash', 'merit', 'bonus',
         'incentive', 'deferred', 'ltip', 'stip', 'tdc', 'allowance'},
        'lone executive silhouette standing at floor-to-ceiling boardroom windows, '
        'back to camera, dramatic city skyline at dusk, navy and gold reflections on glass',
    ),
    (
        {'equity', 'stock', 'option', 'grant', 'vesting', 'share', 'rsu',
         'sar', 'ownership', 'dilution'},
        'single business figure seen from behind at panoramic corner office window, '
        'golden hour city skyline, long shadows, cinematic atmosphere',
    ),
    (
        {'benefit', 'healthcare', 'health', 'wellness', 'medical', 'insurance',
         'mental', 'wellbeing', 'clinical', 'pension', 'retirement'},
        'small group of professional silhouettes in a bright modern open workspace, '
        'back-lit by floor-to-ceiling windows, lush greenery visible, warm natural light',
    ),
    (
        {'survey', 'benchmark', 'analysis', 'data', 'metric', 'assessment',
         'study', 'findings', 'research', 'regression', 'statistics', 'output'},
        'analyst figure from behind at a wide curved workstation, '
        'surrounded by glowing monitors in a dark office, moody blue ambient light',
    ),
    (
        {'leadership', 'executive', 'ceo', 'cfo', 'coo', 'president',
         'officer', 'management', 'director', 'committee', 'board'},
        'solitary executive silhouette at the head of a long boardroom table, '
        'seen from far end of room, city panorama behind, dramatic overhead lighting',
    ),
    (
        {'governance', 'compliance', 'regulatory', 'audit', 'policy',
         'framework', 'documentation', 'procedure', 'sox', 'dodd', 'sec'},
        'two figures in formal attire seen from behind in a corridor of dark oak paneling, '
        'walking away toward a lit conference room, serious atmosphere',
    ),
    (
        {'talent', 'hiring', 'workforce', 'employee', 'recruitment',
         'staffing', 'retention', 'engagement', 'culture', 'people', 'diversity'},
        'diverse group of professional silhouettes gathered around a bright window wall, '
        'seen from a distance, collaborative energy, warm modern office interior',
    ),
    (
        {'technology', 'tech', 'software', 'digital', 'cloud', 'microservices',
         'migration', 'infrastructure', 'platform', 'cybersecurity', 'security'},
        'solitary figure from behind at a standing desk in a dark tech office, '
        'blue LED accent lighting, glass partitions, polished concrete, moody atmosphere',
    ),
    (
        {'financial', 'finance', 'investment', 'banking', 'fund', 'capital',
         'private', 'portfolio', 'market', 'asset', 'revenue'},
        'financial district glass tower at twilight, two silhouetted figures on an outdoor '
        'terrace seen from below, city lights, navy sky, gold window reflections',
    ),
    (
        {'real', 'estate', 'property', 'construction', 'building', 'facility'},
        'architect silhouette against a dramatic sunset behind a modern glass tower, '
        'seen from ground level, bold geometric facade, long shadows',
    ),
    (
        {'media', 'creative', 'marketing', 'brand', 'entertainment',
         'music', 'royalty', 'publishing', 'content'},
        'creative professional silhouette at a large studio window, '
        'warm industrial interior behind, city view, golden afternoon light',
    ),
    (
        {'legal', 'contract', 'counsel', 'attorney', 'litigation', 'dispute'},
        'lone figure from behind in a dark law library corridor, '
        'tall bookshelves on both sides, single reading lamp ahead, serious mood',
    ),
    (
        {'nonprofit', 'foundation', 'charitable', 'mission', 'social', 'impact'},
        'small group of silhouettes at a bright floor-to-ceiling window in a modern office, '
        'plants visible, warm natural light, optimistic atmosphere',
    ),
]

_DEFAULT_SCENE = (
    'executive silhouette from behind at floor-to-ceiling office windows, '
    'dramatic city skyline view, navy and gold tones, cinematic lighting'
)

# Fixed style wrapper applied to every prompt
_STYLE_PREFIX = (
    'professional architectural interior photography, photorealistic, 8k, '
    'sharp focus, no people, empty space, cinematic lighting, '
)
_STYLE_SUFFIX = (
    ', navy blue and gold color palette, clean architectural lines, '
    'high-end interior design, polished surfaces, depth of field, color graded'
)

_NEGATIVE_PROMPT = (
    'face, faces, portrait, close-up face, deformed face, blurry face, '
    'stairs, staircase, stairwell, escalator, mezzanine, upper floor, balcony, multi-level, '
    'floating objects, surreal, distorted perspective, impossible architecture, '
    'text, words, letters, numbers, watermark, logo, sign, label, caption, '
    'typography, font, writing, signage, billboard, poster, screen text, '
    'cartoon, anime, illustration, painting, drawing, sketch, render, 3d cgi, '
    'low quality, blurry, distorted, ugly, deformed, artifacts, '
    'oversaturated, overexposed, nsfw'
)

OG_WIDTH  = 1200
OG_HEIGHT = 630
# Generate slightly smaller (faster) then scale up cleanly
_GEN_WIDTH  = 960
_GEN_HEIGHT = 504


def _slug_to_scene(slug: str) -> str:
    """Map a slug to a visual scene description by keyword matching."""
    words = set(slug.lower().replace('-', ' ').split())
    for keywords, scene in _KEYWORD_SCENES:
        if words & keywords:
            return scene
    return _DEFAULT_SCENE


def _slug_to_seed(slug: str) -> int:
    """Deterministic seed so the same slug always produces the same image."""
    return int(hashlib.md5(f'og_{slug}'.encode()).hexdigest(), 16) % (2 ** 32)


def _slug_to_subject(slug: str) -> str:
    """Convert a slug into a natural-language subject description."""
    # Strip trailing numeric ID (e.g. -7381)
    clean = _re.sub(r'-\d{3,}$', '', slug)
    # Convert hyphens to spaces and title-case
    return clean.replace('-', ' ')


def _build_prompt(slug: str) -> str:
    subject = _slug_to_subject(slug)
    scene = _slug_to_scene(slug)
    # Subject drives the image; scene provides compositional/atmospheric context
    return f'{_STYLE_PREFIX}{subject}, {scene}{_STYLE_SUFFIX}'


class Command(BaseCommand):
    help = 'Generate OG banner images for archive slugs using SDXL-Turbo'

    def add_arguments(self, parser):
        parser.add_argument(
            'slugs',
            nargs='*',
            help='Specific slugs to generate (default: all _ARCHIVE_SLUGS)',
        )
        parser.add_argument(
            '--output-dir',
            default=None,
            help='Output directory (default: staticfiles/og/ relative to manage.py)',
        )
        parser.add_argument(
            '--steps',
            type=int,
            default=20,
            help='Inference steps (default: 20)',
        )
        parser.add_argument(
            '--force',
            action='store_true',
            help='Regenerate images that already exist',
        )
        parser.add_argument(
            '--dry-run',
            action='store_true',
            help='Print what would be generated without running the model',
        )
        parser.add_argument(
            '--model',
            default='stabilityai/stable-diffusion-xl-base-1.0',
            help='Hugging Face model ID (default: stabilityai/stable-diffusion-xl-base-1.0)',
        )

    def handle(self, **options):
        try:
            import torch
            from diffusers import AutoPipelineForText2Image
            from PIL import Image
        except ImportError as e:
            raise CommandError(
                f'Missing dependency: {e}\n\n'
                'Install with:\n'
                '  pip install diffusers transformers accelerate pillow torch torchvision\n'
            )

        # ── Resolve output dir ─────────────────────────────────────────────────
        if options['output_dir']:
            out_dir = Path(options['output_dir'])
        else:
            # media/og/ next to manage.py
            manage_dir = Path(sys.argv[0]).resolve().parent
            out_dir = manage_dir / 'media' / 'og'
        out_dir.mkdir(parents=True, exist_ok=True)

        # ── Resolve slugs ──────────────────────────────────────────────────────
        if options['slugs']:
            slugs = options['slugs']
        else:
            from apps.honeypot.archive_data import _ARCHIVE_SLUGS
            slugs = list(_ARCHIVE_SLUGS)
        self.stdout.write(f'Slugs to process: {len(slugs)}')

        # Filter already-existing unless --force
        if not options['force']:
            pending = [s for s in slugs if not (out_dir / f'{s}.jpg').exists()]
            skipped = len(slugs) - len(pending)
            if skipped:
                self.stdout.write(f'  Skipping {skipped} already generated (use --force to overwrite)')
            slugs = pending

        if not slugs:
            self.stdout.write(self.style.SUCCESS('Nothing to generate.'))
            return

        if options['dry_run']:
            for slug in slugs:
                prompt = _build_prompt(slug)
                seed = _slug_to_seed(slug)
                self.stdout.write(f'  {slug}\n    seed={seed}\n    prompt={prompt[:120]}...\n')
            return

        # ── Load model ─────────────────────────────────────────────────────────
        device, dtype = _pick_device(torch)
        self.stdout.write(f'Loading {options["model"]} on {device} ({dtype}) …')

        pipe = AutoPipelineForText2Image.from_pretrained(
            options['model'],
            torch_dtype=dtype,
            variant='fp16' if dtype == torch.float16 else None,
        )
        pipe = pipe.to(device)
        pipe.set_progress_bar_config(disable=True)

        self.stdout.write(self.style.SUCCESS('Model loaded.'))

        # ── Generate ───────────────────────────────────────────────────────────
        steps  = max(1, options['steps'])
        errors = 0

        for i, slug in enumerate(slugs, 1):
            out_path = out_dir / f'{slug}.jpg'
            prompt   = _build_prompt(slug)
            seed     = _slug_to_seed(slug)

            self.stdout.write(f'[{i}/{len(slugs)}] {slug}', ending=' ')
            self.stdout.flush()

            try:
                generator = torch.Generator(device=device).manual_seed(seed)
                result = pipe(
                    prompt=prompt,
                    negative_prompt=_NEGATIVE_PROMPT,
                    num_inference_steps=steps,
                    guidance_scale=7.5,
                    width=_GEN_WIDTH,
                    height=_GEN_HEIGHT,
                    generator=generator,
                )
                img = result.images[0]

                # Scale to OG dimensions with high-quality Lanczos
                if img.size != (OG_WIDTH, OG_HEIGHT):
                    img = img.resize((OG_WIDTH, OG_HEIGHT), Image.LANCZOS)

                _save_atomic(img, out_path)
                self.stdout.write(self.style.SUCCESS('✓'))

            except Exception as exc:
                self.stdout.write(self.style.ERROR(f'FAILED: {exc}'))
                errors += 1

        self.stdout.write(
            f'\nDone. {len(slugs) - errors} generated, {errors} errors → {out_dir}'
        )


def _pick_device(torch):
    """Return (device_str, dtype) for the best available accelerator."""
    if torch.cuda.is_available():
        return 'cuda', torch.float16
    if getattr(torch.backends, 'mps', None) and torch.backends.mps.is_available():
        return 'mps', torch.float16
    return 'cpu', torch.float32


def _save_atomic(img, path: Path):
    """Write JPEG to a temp file then rename — avoids partial files on disk."""
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
