#!/usr/bin/env python3
"""
Standalone speaker headshot generator for PERCH conference pages — no Django required.

Generates a professional portrait for each unique speaker across all conference years.
Images are saved to acpwb/static/img/speakers/<slug>.webp and picked up automatically
by the speaker_avatar template tag on conference pages.

Install:
    pip install torch torchvision --index-url https://download.pytorch.org/whl/cpu
    pip install diffusers transformers accelerate pillow

Usage:
    python generate_speaker_headshots.py               # all speakers
    python generate_speaker_headshots.py --year 2026   # one year only
    python generate_speaker_headshots.py --force        # re-generate existing
    python generate_speaker_headshots.py --dry-run      # preview prompts only
    python generate_speaker_headshots.py --steps 25
    python generate_speaker_headshots.py --model SG161222/RealVisXL_V4.0

Model notes (all work well under 64 GB RAM):
  - stabilityai/stable-diffusion-xl-base-1.0  — default, safe, matches press release script
  - SG161222/RealVisXL_V4.0                   — SDXL-based, photorealistic portraits (recommended)
  - Lykon/dreamshaper-xl-1-0                  — SDXL-based, fast, good diversity
"""
import argparse
import hashlib
import os
import re
import sys
import tempfile
from pathlib import Path

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE / 'acpwb'))

# ── Output ─────────────────────────────────────────────────────────────────────

SPEAKERS_DIR = HERE / 'acpwb' / 'static' / 'img' / 'speakers'
OUTPUT_SIZE = 512   # final webp square side; circles on the site are ≤192px @2×
GEN_SIZE = 1024     # SDXL native square; portrait-focused prompts keep faces centered

# ── Gender inference ────────────────────────────────────────────────────────────
# Strip honorifics, take first name, match against curated sets.
# Falls back to seed-based random for genuinely ambiguous names.

_HONORIFICS = frozenset([
    'dr', 'dr.', 'prof', 'prof.', 'judge', 'congressman', 'congresswoman',
    'senator', 'rev', 'rev.', 'mr', 'mr.', 'ms', 'ms.', 'mrs', 'mrs.',
    'hon', 'hon.', 'the',
])

_FEMALE_NAMES = frozenset([
    # Names appearing in PERCH conference data
    'margaret', 'sandra', 'alicia', 'teresa', 'yolanda', 'patricia', 'nadia',
    'rachel', 'jennifer', 'anita', 'tiffany', 'claudia', 'elena', 'miriam',
    # Broad common names
    'mary', 'barbara', 'elizabeth', 'linda', 'susan', 'karen', 'nancy',
    'lisa', 'betty', 'helen', 'dorothy', 'donna', 'carol', 'ruth', 'sharon',
    'michelle', 'laura', 'sarah', 'kimberly', 'deborah', 'jessica', 'shirley',
    'angela', 'emily', 'brenda', 'pamela', 'emma', 'nicole', 'kathleen',
    'samantha', 'christine', 'diana', 'virginia', 'evelyn', 'kelly', 'judith',
    'marilyn', 'gloria', 'beverly', 'cheryl', 'mildred', 'amber', 'amy',
    'anna', 'annie', 'ashley', 'brittany', 'carolyn', 'charlotte', 'chloe',
    'christina', 'claire', 'cynthia', 'danielle', 'debra', 'diane', 'eleanor',
    'elaine', 'felicia', 'frances', 'gina', 'grace', 'hannah', 'heather',
    'irene', 'jacqueline', 'janice', 'jean', 'joyce', 'june', 'katelyn',
    'kayla', 'kathy', 'kim', 'kristen', 'kristin', 'latoya', 'lauren',
    'leah', 'lillian', 'lorraine', 'madison', 'maria', 'marina', 'megan',
    'melanie', 'melissa', 'miranda', 'molly', 'monique', 'natalie', 'olivia',
    'phyllis', 'priscilla', 'raven', 'rhonda', 'rose', 'sheila', 'stacy',
    'stephanie', 'tammy', 'taylor', 'tina', 'tracey', 'tracy', 'valerie',
    'vera', 'vicki', 'wendy', 'yvonne', 'abigail', 'aisha', 'ava', 'bella',
    'bethany', 'bianca', 'camille', 'denise', 'dora', 'edith', 'ella',
    'fiona', 'gwendolyn', 'ida', 'ingrid', 'iris', 'ivy', 'jade', 'jenna',
    'josephine', 'julia', 'katarina', 'kylie', 'lena', 'lily', 'loretta',
    'lori', 'luna', 'lydia', 'mae', 'marcia', 'maxine', 'mia', 'mina',
    'myra', 'naomi', 'nina', 'nora', 'paige', 'pearl', 'peggy', 'penelope',
    'phoebe', 'raquel', 'rebekah', 'renee', 'rhea', 'rina', 'rosa',
    'roxanne', 'ruby', 'sabrina', 'sally', 'selena', 'serena', 'shannon',
    'sierra', 'sofia', 'sonja', 'sophia', 'stella', 'tamara', 'tatiana',
    'thea', 'toni', 'tonya', 'ursula', 'vanessa', 'violet', 'vivian',
    'whitney', 'wilma', 'zoe', 'fatima', 'amara', 'keisha', 'layla',
    'destiny', 'brianna', 'jasmine', 'aaliyah', 'imani', 'latasha',
    'lakisha', 'tamika', 'shanice', 'yuki', 'mei', 'sakura', 'lin',
    'linh', 'priya', 'deepa', 'sunita', 'ananya', 'pooja', 'rekha',
    'svetlana', 'natasha', 'olga', 'irina', 'katerina', 'oksana',
    'agnieszka', 'katarzyna', 'magdalena', 'zofia',
    # missing from initial list
    'adriana', 'renata', 'helena', 'simone', 'lydia', 'eleanor',
    'miriam', 'carolyn', 'maria', 'nadia', 'priya', 'elena',
    'tiffany', 'anita', 'yolanda', 'jennifer', 'alicia', 'patricia',
])

_MALE_NAMES = frozenset([
    # Names appearing in PERCH conference data
    'james', 'clifton', 'robert', 'marcus', 'david', 'patrick', 'victor',
    'christopher', 'luis', 'thomas', 'gerald', 'raymond', 'henry',
    # Broad common names
    'john', 'william', 'richard', 'charles', 'joseph', 'daniel', 'paul',
    'mark', 'donald', 'george', 'kenneth', 'steven', 'edward', 'brian',
    'ronald', 'anthony', 'kevin', 'jason', 'matthew', 'gary', 'timothy',
    'jose', 'larry', 'jeffrey', 'frank', 'scott', 'eric', 'stephen',
    'andrew', 'gregory', 'joshua', 'jerry', 'dennis', 'walter', 'peter',
    'harold', 'douglas', 'carl', 'arthur', 'ryan', 'roger', 'joe', 'juan',
    'jack', 'albert', 'jonathan', 'justin', 'terry', 'keith', 'samuel',
    'willie', 'ralph', 'lawrence', 'nicholas', 'roy', 'benjamin', 'bruce',
    'brandon', 'adam', 'harry', 'fred', 'wayne', 'billy', 'steve', 'louis',
    'jeremy', 'aaron', 'randy', 'howard', 'eugene', 'carlos', 'russell',
    'bobby', 'martin', 'ernest', 'phillip', 'todd', 'jesse', 'craig',
    'alan', 'shawn', 'sean', 'philip', 'chris', 'johnny', 'earl', 'jimmy',
    'antonio', 'danny', 'bryan', 'tony', 'mike', 'stanley', 'leonard',
    'nathan', 'dale', 'manuel', 'rodney', 'curtis', 'norman', 'allen',
    'marvin', 'vincent', 'glen', 'jeffery', 'travis', 'jeff', 'chad',
    'jacob', 'melvin', 'alfred', 'kyle', 'neil', 'floyd', 'alvin', 'tim',
    'darryl', 'reginald', 'oscar', 'clifford', 'willard', 'darrell',
    'ross', 'brent', 'andre', 'felix', 'ted', 'herman', 'derek', 'lester',
    'alexander', 'elijah', 'oliver', 'evan', 'gabriel', 'dylan', 'caleb',
    'mason', 'ethan', 'lucas', 'liam', 'noah', 'aidan', 'miles', 'nolan',
    'dominic', 'jordan', 'omar', 'malik', 'darius', 'jerome', 'terrence',
    'warren', 'trevor', 'spencer', 'grant', 'blake', 'brett', 'lance',
    'cory', 'dustin', 'colin', 'rory', 'ian', 'finn', 'axel', 'benedict',
    'cedric', 'damien', 'dawson', 'dean', 'devin', 'donovan', 'dorian',
    'drake', 'drew', 'edgar', 'elliot', 'emilio', 'emmett', 'enrique',
    'ezra', 'fabian', 'ford', 'franklin', 'gavin', 'gene', 'geoffrey',
    'gordon', 'graham', 'gunnar', 'harris', 'harrison', 'hayes', 'hector',
    'holden', 'homer', 'houston', 'hudson', 'hugo', 'hunter', 'ivan',
    'jace', 'jaime', 'jake', 'jasper', 'javier', 'jefferson', 'jett',
    'joaquin', 'joel', 'jonah', 'jonas', 'jorge', 'julian', 'julius',
    'kai', 'keegan', 'kelvin', 'kent', 'kian', 'kieran', 'kingsley',
    'knox', 'kurt', 'lamar', 'langston', 'lars', 'lawson', 'leandro',
    'lee', 'leo', 'leon', 'levi', 'lewis', 'lincoln', 'lionel', 'lloyd',
    'logan', 'luca', 'luke', 'luther', 'mack', 'marco', 'mario',
    'maxwell', 'max', 'micah', 'michael', 'mitchell', 'morgan', 'myron',
    'neal', 'ned', 'nelson', 'nigel', 'noel', 'norbert', 'orion', 'otto',
    'pablo', 'parker', 'pascal', 'pedro', 'pierce', 'quinn', 'randall',
    'raphael', 'reed', 'rhett', 'ricardo', 'riley', 'roberto', 'rod',
    'rodrigo', 'rohan', 'roland', 'romeo', 'roscoe', 'ruben', 'rupert',
    'salvador', 'sam', 'samir', 'santiago', 'sebastian', 'sergio', 'silas',
    'simon', 'solomon', 'sterling', 'stewart', 'stuart', 'sylvester',
    'terence', 'thaddeus', 'theo', 'theodore', 'titus', 'tobias', 'toby',
    'tomas', 'trent', 'tristan', 'tucker', 'tyler', 'ulrich', 'uriah',
    'vance', 'wade', 'walker', 'warner', 'watson', 'wendell', 'weston',
    'wilbur', 'winston', 'woodrow', 'wyatt', 'xander', 'yusuf', 'zachary',
    'zane', 'kwame', 'kofi', 'kweku', 'ade', 'emeka', 'chidi', 'obinna',
    'tariq', 'jamal', 'rashid', 'kareem', 'devonte', 'deshawn', 'xavier',
    'kenji', 'hiroshi', 'akira', 'takeshi', 'raj', 'arjun', 'vikram',
    'rahul', 'sanjay', 'arun', 'mikhail', 'dmitri', 'alexei', 'igor',
    'piotr', 'andrzej', 'krzysztof', 'wojciech',
    # missing from initial list
    'nathaniel', 'kevin', 'william', 'thomas', 'marcus', 'clifton',
])


def _infer_gender(name, rng):
    """Infer gender from first name; fall back to even 50/50 for ambiguous cases."""
    parts = name.lower().split()
    first = next((p for p in parts if p not in _HONORIFICS), '')
    if first in _FEMALE_NAMES:
        return 'woman'
    if first in _MALE_NAMES:
        return 'man'
    return rng.choice(['man', 'woman'])


# ── Appearance pools ────────────────────────────────────────────────────────────

_AGES = [
    'in their mid-30s',
    'in their early 40s',
    'in their late 40s',
    'in their early 50s',
    'in their late 50s',
    'in their early 60s',
]

# Wide diversity of complexions, hair textures, and features
_LOOKS = [
    # Light / fair
    'with light complexion and straight brown hair',
    'with fair complexion and blonde hair, blue eyes',
    'with pale complexion and dark auburn hair',
    'with fair complexion and silver-gray hair, distinguished',
    'with light complexion and short red hair',
    # Medium / olive
    'with olive complexion and dark wavy hair',
    'with tan complexion and thick dark hair',
    'with light brown complexion and black hair',
    'with warm medium complexion and curly dark hair',
    'with golden-brown complexion and dark straight hair',
    # South Asian
    'with warm brown complexion and black hair, South Asian features',
    'with medium brown complexion and dark eyes, South Asian heritage',
    # East / Southeast Asian
    'with light complexion and straight black hair, East Asian features',
    'with warm complexion and black hair, Southeast Asian features',
    # Middle Eastern / North African
    'with olive complexion and dark hair, Middle Eastern features',
    'with warm tan complexion and dark eyes, North African heritage',
    # Latino / Hispanic
    'with warm medium complexion and dark curly hair, Latin heritage',
    'with light olive complexion and dark wavy hair, Hispanic features',
    # Black / African / African-American
    'with dark brown complexion and natural coily hair',
    'with deep brown complexion and closely cropped hair',
    'with rich dark complexion and natural hair, West African features',
    'with medium-dark complexion and short natural hair',
    'with dark complexion and locs, distinguished and composed',
    # Mixed / ambiguous
    'with warm medium complexion and curly brown hair',
    'with light brown complexion and wavy dark-brown hair',
    'with medium complexion and salt-and-pepper hair',
    # Older / distinguished
    'with fair complexion and white hair, distinguished and sharp-eyed',
    'with dark complexion and close-cropped gray hair, distinguished',
]

# ── Attire pools ────────────────────────────────────────────────────────────────

_ATTIRE_FORMAL = [
    'wearing a dark charcoal business suit and tie',
    'wearing a navy pinstripe suit with a white pocket square',
    'wearing a classic black blazer and crisp white dress shirt',
    'wearing a tailored charcoal blazer and black turtleneck',
    'wearing a deep burgundy blazer over a white blouse',
    'wearing a navy double-breasted suit',
    'wearing a formal dark suit with a subtle plaid pattern',
]

_ATTIRE_BUSINESS_PROFESSIONAL = [
    'wearing a professional navy blazer over a patterned blouse',
    'wearing a well-fitted dark blazer and silk blouse',
    'wearing a smart business blazer over a button-down shirt',
    'wearing a classic grey blazer and dark turtleneck',
    'wearing a structured forest-green blazer over a black top',
    'wearing a rich plum blazer with a simple neckline',
    'wearing a tailored cobalt blue blazer and slacks',
    'wearing a camel-colored blazer and ivory blouse',
]

_ATTIRE_ACADEMIC = [
    'wearing a herringbone tweed jacket over a collared shirt',
    'wearing a brown corduroy blazer over a dark turtleneck',
    'wearing a relaxed linen blazer with an open-collar shirt',
    'wearing a patterned academic jacket over a simple blouse',
    'wearing a dark sweater over a buttoned plaid shirt',
    'wearing a crew-neck cardigan over a checked dress shirt',
    'wearing a classic academia blazer with elbow patches',
]

_ATTIRE_SMART_CASUAL = [
    'wearing a fitted merino turtleneck sweater in charcoal',
    'wearing a v-neck cashmere sweater over a collared shirt',
    'wearing a dark mock-neck sweater, relaxed professional',
    'wearing a smart casual blazer over a jersey crewneck',
    'wearing a zip-up cardigan and open-collar shirt',
    'wearing a structured knit top in a deep jewel tone',
]

_ATTIRE_CREATIVE_PROFESSIONAL = [
    'wearing a bold-color structured blazer with modern lapels',
    'wearing a fashionable blazer in a deep terracotta tone',
    'wearing a contemporary blazer with a geometric pattern',
    'wearing a modern unstructured jacket in warm olive',
    'wearing a statement blazer in deep teal over a dark top',
    'wearing an oversized artistic jacket and simple shirt',
]

_ATTIRE_LEGAL_GOVT = [
    'wearing a conservative dark suit with a subtle tie',
    'wearing formal business attire in deep navy or charcoal',
    'wearing a well-pressed professional blazer and dress shirt',
    'wearing a classic legal-professional dark suit',
    'wearing a formal blazer with an American flag lapel pin',
]


def _classify_attire(title, org, rng):
    t = (title + ' ' + org).lower()
    is_academic = any(k in t for k in (
        'professor', 'lecturer', 'academic', 'university', 'college', 'school of', 'faculty', 'research'))
    is_legal_govt = any(k in t for k in (
        'llp', 'law', 'attorney', 'counsel', 'judge', 'legal', 'partner',
        'department of', 'dol', 'eeoc', 'federal', 'government', 'secretary',
        'bureau', 'senator', 'congressman', 'congresswoman', 'ministry'))

    if is_legal_govt:
        pool = _ATTIRE_LEGAL_GOVT + _ATTIRE_FORMAL
    elif is_academic:
        pool = _ATTIRE_ACADEMIC + _ATTIRE_SMART_CASUAL + _ATTIRE_BUSINESS_PROFESSIONAL
    else:
        # Corporate/consulting: full variety
        pool = (
            _ATTIRE_FORMAL + _ATTIRE_BUSINESS_PROFESSIONAL +
            _ATTIRE_SMART_CASUAL + _ATTIRE_CREATIVE_PROFESSIONAL
        )
    return rng.choice(pool)


# ── Background pools ────────────────────────────────────────────────────────────
# Each entry is the background + lighting description (replaces _STYLE_SUFFIX's BG portion).

_BACKGROUNDS = [
    # Studio / clean
    'neutral light gray studio background, soft professional studio lighting, subtle shadow on one side',
    'soft warm cream studio background, gentle diffused studio lighting',
    'clean white background with a soft gradient, bright even lighting',
    'dark charcoal-to-black gradient background, dramatic Rembrandt lighting',
    'deep navy gradient background, crisp studio lighting',
    # Indoor / environmental
    'blurred warm wooden bookshelf in the background, soft academic library light',
    'blurred office bookshelves in background, warm reading-room glow',
    'blurred modern glass office interior, cool natural window light from the side',
    'blurred floor-to-ceiling window with soft city view, golden afternoon light',
    'blurred conference room interior, neutral professional lighting',
    'blurred white brick interior wall, warm ambient lighting',
    'blurred indoor plants and greenery, soft natural light from above',
    # Outdoor / architectural (soft focus)
    'blurred outdoor courtyard in background, bright overcast natural light',
    'blurred modern architectural facade, cool daylight from the left',
    'blurred autumn trees in background, warm golden-hour natural light',
    'blurred urban exterior in background, clean bright daylight',
]

_STYLE_FIXED = (
    ', sharp focus on face and eyes, head and shoulders portrait, '
    'photorealistic portrait photography, 8k, high resolution, no text, no watermark'
)

_NEGATIVE_PROMPT = (
    'cartoon, anime, painting, illustration, sketch, drawing, render, '
    'sunglasses, hat, bokeh, overexposed, underexposed, '
    'low quality, blurry, distorted, deformed, disfigured, ugly, '
    'watermark, logo, text, nsfw, nudity, multiple people, group'
)


def _build_prompt(speaker, rng):
    gender = _infer_gender(speaker['name'], rng)
    age = rng.choice(_AGES)
    looks = rng.choice(_LOOKS)
    attire = _classify_attire(speaker['title'], speaker['org'], rng)
    background = rng.choice(_BACKGROUNDS)
    return (
        f'close-up professional headshot portrait of a {gender} {age} {looks}, '
        f'{attire}, {background}{_STYLE_FIXED}'
    )


# ── Utilities ───────────────────────────────────────────────────────────────────

def _speaker_slug(name):
    slug = name.lower()
    slug = re.sub(r'[^a-z0-9]+', '-', slug)
    return slug.strip('-')


def _name_seed(name):
    return int(hashlib.md5(f'perch_speaker_{name}'.encode()).hexdigest(), 16) % (2 ** 32)


def _is_flux(model_id):
    return 'flux' in model_id.lower()


def _pick_device(torch, flux=False):
    if torch.cuda.is_available():
        return 'cuda', torch.bfloat16 if flux else torch.float16
    if getattr(torch.backends, 'mps', None) and torch.backends.mps.is_available():
        return 'mps', torch.float32
    return 'cpu', torch.float32


def _save_atomic(img, path):
    fd, tmp = tempfile.mkstemp(suffix='.webp', dir=path.parent)
    try:
        os.close(fd)
        img.save(tmp, 'WEBP', quality=90, method=6)
        os.replace(tmp, path)
    except Exception:
        try:
            os.unlink(tmp)
        except OSError:
            pass
        raise


def _collect_speakers(conferences, year_filter=None):
    seen = {}
    for year, conf in sorted(conferences.items()):
        if year_filter and year != year_filter:
            continue
        for s in conf.get('speakers', []):
            name = s['name']
            if name not in seen:
                seen[name] = dict(s, year_first=year)
    return list(seen.values())


def main():
    parser = argparse.ArgumentParser(description='Generate headshot images for PERCH conference speakers')
    parser.add_argument('--year', type=int, default=None, help='Limit to speakers from a specific year')
    parser.add_argument('--output-dir', default=str(SPEAKERS_DIR))
    parser.add_argument('--model', default='stabilityai/stable-diffusion-xl-base-1.0',
                        help='HF model ID (default: SDXL; try SG161222/RealVisXL_V4.0 for better portraits)')
    parser.add_argument('--steps', type=int, default=30)
    parser.add_argument('--force', action='store_true', help='Re-generate images that already exist')
    parser.add_argument('--dry-run', action='store_true', help='Print prompts without generating')
    parser.add_argument('--hf-token', default=os.environ.get('HF_TOKEN'))
    parser.add_argument('--quantize', action='store_true',
                        help='Quantize transformer to qint8 via optimum-quanto (saves ~50%% RAM)')
    args = parser.parse_args()

    try:
        import random
        import torch
        from diffusers import AutoPipelineForText2Image
        from diffusers.utils import logging as diffusers_logging
        from PIL import Image
        from apps.public.conference_data import CONFERENCES
    except ImportError as e:
        sys.exit(
            f'Missing dependency: {e}\n\n'
            'Install with:\n'
            '  pip install torch torchvision --index-url https://download.pytorch.org/whl/cpu\n'
            '  pip install diffusers transformers accelerate pillow\n'
        )

    out_dir = Path(args.output_dir)
    out_dir.mkdir(parents=True, exist_ok=True)

    speakers = _collect_speakers(CONFERENCES, year_filter=args.year)
    print(f'Unique speakers to process: {len(speakers)}')

    if not args.force:
        pending = [s for s in speakers if not (out_dir / f'{_speaker_slug(s["name"])}.webp').exists()]
        skipped = len(speakers) - len(pending)
        if skipped:
            print(f'  Skipping {skipped} already generated (use --force to overwrite)')
        speakers = pending

    if not speakers:
        print('Nothing to generate.')
        return

    if args.dry_run:
        import random as _random
        for s in speakers:
            rng = _random.Random(_name_seed(s['name']))
            prompt = _build_prompt(s, rng)
            slug = _speaker_slug(s['name'])
            seed = _name_seed(s['name'])
            gender = _infer_gender(s['name'], _random.Random(seed))
            print(f'  {s["name"]}  [{gender}]\n    slug={slug}  seed={seed}\n    {prompt}\n')
        return

    flux = _is_flux(args.model)
    device, dtype = _pick_device(torch, flux=flux)
    print(f'Loading {args.model} on {device} ({dtype}) …')

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

    import random
    errors = 0
    for i, s in enumerate(speakers, 1):
        name = s['name']
        slug = _speaker_slug(name)
        out_path = out_dir / f'{slug}.webp'
        seed = _name_seed(name)
        rng = random.Random(seed)
        prompt = _build_prompt(s, rng)

        print(f'[{i}/{len(speakers)}] {name} ', end='', flush=True)
        try:
            generator = torch.Generator(device='cpu').manual_seed(seed)
            gen_kwargs = dict(
                prompt=prompt,
                num_inference_steps=args.steps,
                width=GEN_SIZE,
                height=GEN_SIZE,
                generator=generator,
            )
            if flux:
                gen_kwargs['guidance_scale'] = 0.0
            else:
                gen_kwargs['negative_prompt'] = _NEGATIVE_PROMPT
                gen_kwargs['guidance_scale'] = 7.5
            result = pipe(**gen_kwargs)
            img = result.images[0]
            if img.size != (OUTPUT_SIZE, OUTPUT_SIZE):
                img = img.resize((OUTPUT_SIZE, OUTPUT_SIZE), Image.LANCZOS)
            _save_atomic(img, out_path)
            print(f'✓  →  {slug}.webp')
        except Exception as exc:
            print(f'FAILED: {exc}')
            errors += 1

    print(f'\nDone. {len(speakers) - errors} generated, {errors} errors → {out_dir}')


if __name__ == '__main__':
    main()
