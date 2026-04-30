#!/usr/bin/env python3
"""
Standalone professional headshot generator for ACPWB fake employees — no Django required.

Install:
    pip install torch torchvision --index-url https://download.pytorch.org/whl/cpu
    pip install diffusers transformers accelerate pillow

Usage:
    python generate_headshots2.py                                  # all 400 headshots
    python generate_headshots2.py 5 42 100                         # specific indices
    python generate_headshots2.py --count 600                      # generate 0–599
    python generate_headshots2.py --force                          # re-generate existing
    python generate_headshots2.py --dry-run                        # preview prompts only
    python generate_headshots2.py --steps 4                        # inference steps (default 4)

Recommended model:
    python generate_headshots2.py --model black-forest-labs/FLUX.1-schnell --steps 4
"""
import argparse
import hashlib
import os
import random
import sys
import tempfile
from pathlib import Path

HERE = Path(__file__).resolve().parent

DEFAULT_COUNT = 400
DEFAULT_OUTPUT = HERE / "acpwb" / "static" / "img" / "headshots2"

# ── Person attribute pools ─────────────────────────────────────────────────────

# 98% male/female, 2% other
_GENDER_CHOICES = ["man", "woman", "non-binary professional", "androgynous professional"]
_GENDER_WEIGHTS = [49, 49, 1, 1]

AGES = [
    24, 26, 27, 29, 31, 33, 35, 37, 39, 41, 43, 45, 47, 49,
    51, 53, 55, 57, 59, 61, 64, 67,
]

# (value, weight) — Milwaukee/Midwest professional workforce distribution
# European heritages broken out to reflect Milwaukee's German, Irish, Polish roots
_ETHNICITY_CHOICES = [
    "German American",
    "Irish American",
    "Polish American",
    "Italian American",
    "Scandinavian American",
    "English American",
    "Eastern European American",
    "Black American",
    "Hispanic American",
    "Mexican American",
    "Puerto Rican American",
    "East Asian American",
    "South Asian American",
    "Middle Eastern American",
    "Southeast Asian American",
    "mixed-race American",
    "Native American",
]
_ETHNICITY_WEIGHTS = [10, 8, 7, 5, 4, 4, 3, 14, 8, 5, 3, 8, 8, 5, 4, 3, 1]

HAIR_MAN = [
    "short dark hair",
    "short brown hair",
    "short blonde hair",
    "short gray hair",
    "short black hair",
    "close-cropped curly black hair",
    "neatly combed dark hair",
    "neatly combed brown hair parted on the side",
    "salt-and-pepper hair",
    "salt-and-pepper hair parted on the side",
    "bald",
    "shaved head",
    "closely cropped silver hair",
    "short wavy brown hair",
    "short wavy black hair",
    "short wavy dark hair",
    "receding dark hair",
    "receding blonde hair",
    "neat dark auburn hair",
    "short natural hair",
    "short tight curls",
    "locs",
    "short locs",
    "medium-length dark hair slicked back",
    "short red hair",
    "short strawberry blonde hair",
    "closely cropped dark brown hair",
    "short sandy blonde hair",
    "thick dark hair combed neatly",
    "graying brown hair",
    "closely cropped gray hair",
    "buzz cut",
    "short dark hair with a widow's peak",
    "neatly trimmed dark beard and short hair",
    "short hair with slight wave",
]

HAIR_WOMAN = [
    "shoulder-length brown hair",
    "shoulder-length black hair",
    "shoulder-length blonde hair",
    "shoulder-length red hair",
    "shoulder-length auburn hair",
    "shoulder-length gray hair",
    "short blonde hair",
    "short dark hair",
    "short black hair",
    "short red hair",
    "short gray pixie cut",
    "short silver pixie cut",
    "dark hair in a neat updo",
    "brown hair in a professional bun",
    "black hair in a tight bun",
    "blonde hair pulled back",
    "natural curly dark hair",
    "natural curly brown hair",
    "tight natural curls",
    "straight black bob",
    "straight brown bob",
    "neat silver hair",
    "gray hair in a professional style",
    "neat white hair",
    "auburn shoulder-length hair",
    "wavy dark brown hair",
    "wavy auburn hair",
    "wavy blonde hair",
    "neat brunette bob",
    "long straight dark hair pulled back",
    "long straight brown hair",
    "long black hair",
    "short natural curly hair",
    "highlighted brown shoulder-length hair",
    "highlighted blonde shoulder-length hair",
    "balayage brown-to-blonde hair",
    "long dark locs",
    "medium-length locs",
    "medium-length blonde hair",
    "medium-length chestnut hair",
    "glossy straight black hair",
    "layered brown shoulder-length hair",
    "swept-back dark hair",
    "chignon updo",
    "French twist updo",
]

HAIR_NEUTRAL = [
    "neat short dark hair",
    "natural close-cropped hair",
    "professional short hair",
    "neat medium-length dark hair",
    "short natural hair",
    "close-cropped black hair",
    "short wavy hair",
    "neatly styled short hair",
]

ATTIRE_MAN = [
    "dark navy suit and white dress shirt",
    "charcoal suit and light blue dress shirt",
    "dark suit and burgundy tie",
    "classic black suit and white shirt",
    "navy blazer and open-collar white dress shirt",
    "dark gray suit and striped tie",
    "professional charcoal blazer",
    "dark suit and silk tie",
    "two-button dark wool suit",
    "business casual dark blazer",
    "dark suit and solid tie",
    "navy suit and patterned tie",
]

ATTIRE_WOMAN = [
    "professional navy blazer and white blouse",
    "tailored dark business suit jacket",
    "charcoal blazer over a silk blouse",
    "professional navy jacket",
    "formal dark blazer over a light blouse",
    "elegant gray blazer",
    "professional black blazer over a patterned blouse",
    "tailored dark business jacket",
    "classic dark suit jacket",
    "sophisticated navy blazer and pearl necklace",
    "burgundy blazer over a white blouse",
    "dark turtleneck under a blazer",
]

ATTIRE_NEUTRAL = [
    "professional dark blazer",
    "tailored dark business jacket",
    "formal dark jacket",
    "neat dark blazer and collared shirt",
]

_BASE_STYLE = (
    ", direct eye contact, neutral expression, "
    "gray studio background, studio lighting, "
    "photorealistic, sharp focus, 8k"
)

_NEGATIVE = (
    "cartoon, anime, illustration, painting, art, digital art, "
    "sunglasses, hat, cap, hood, headwear, "
    "outdoor background, office background, busy background, colored background, "
    "multiple people, group photo, crowd, "
    "watermark, logo, text, signature, "
    "blurry, out of focus, overexposed, underexposed, grainy, noise, "
    "distorted, deformed, bad anatomy, disfigured, ugly, nsfw, "
    "extreme makeup, costume, sportswear, casual clothing, t-shirt, hoodie"
)

_GEN_SIZE = 768
OUT_SIZE = 300


def _rng(idx: int) -> random.Random:
    seed = int(hashlib.md5(f"headshot_{idx}".encode()).hexdigest(), 16) % (2 ** 32)
    return random.Random(seed)


def _build_prompt(idx: int) -> str:
    rng = _rng(idx)
    gender = rng.choices(_GENDER_CHOICES, weights=_GENDER_WEIGHTS, k=1)[0]
    age = rng.choice(AGES)
    ethnicity = rng.choices(_ETHNICITY_CHOICES, weights=_ETHNICITY_WEIGHTS, k=1)[0]

    if gender == "man":
        hair = rng.choice(HAIR_MAN)
        attire = rng.choice(ATTIRE_MAN)
    elif gender == "woman":
        hair = rng.choice(HAIR_WOMAN)
        attire = rng.choice(ATTIRE_WOMAN)
    else:
        hair = rng.choice(HAIR_NEUTRAL)
        attire = rng.choice(ATTIRE_NEUTRAL)

    return (
        f"Professional corporate headshot photograph of a {age}-year-old "
        f"{ethnicity} {gender}, {hair}, wearing {attire}{_BASE_STYLE}"
    )


def _idx_to_seed(idx: int) -> int:
    return int(hashlib.md5(f"headshot_gen_{idx}".encode()).hexdigest(), 16) % (2 ** 32)


def _is_flux(model_id: str) -> bool:
    return "flux" in model_id.lower()


def _pick_device(torch, flux: bool = False):
    if torch.cuda.is_available():
        return "cuda", torch.bfloat16 if flux else torch.float16
    if getattr(torch.backends, "mps", None) and torch.backends.mps.is_available():
        return "mps", torch.float32
    return "cpu", torch.float32


def _save_atomic(img, path: Path):
    fd, tmp = tempfile.mkstemp(suffix=".webp", dir=path.parent)
    try:
        os.close(fd)
        img.save(tmp, "WEBP", quality=88, method=6)
        os.replace(tmp, path)
    except Exception:
        try:
            os.unlink(tmp)
        except OSError:
            pass
        raise


def main():
    parser = argparse.ArgumentParser(description="Generate professional headshot images for ACPWB fake employees")
    parser.add_argument("indices", nargs="*", type=int, help="Specific indices to generate (default: 0 to count-1)")
    parser.add_argument("--count", type=int, default=DEFAULT_COUNT, help=f"Total headshot pool size (default: {DEFAULT_COUNT})")
    parser.add_argument("--output-dir", default=str(DEFAULT_OUTPUT))
    parser.add_argument("--model", default="black-forest-labs/FLUX.1-schnell")
    parser.add_argument("--steps", type=int, default=4)
    parser.add_argument("--hf-token", default=os.environ.get("HF_TOKEN"))
    parser.add_argument("--force", action="store_true", help="Re-generate images that already exist")
    parser.add_argument("--dry-run", action="store_true", help="Print prompts without generating images")
    args = parser.parse_args()

    out_dir = Path(args.output_dir)
    out_dir.mkdir(parents=True, exist_ok=True)

    if args.indices:
        bad = [i for i in args.indices if not (0 <= i < args.count)]
        if bad:
            parser.error(f"Indices out of range (0–{args.count - 1}): {bad}")
        indices = args.indices
    else:
        indices = list(range(args.count))

    if not args.force:
        pending = [i for i in indices if not (out_dir / f"{i:03d}.webp").exists()]
        skipped = len(indices) - len(pending)
        if skipped:
            print(f"Skipping {skipped} already generated (use --force to overwrite)")
        indices = pending

    if not indices:
        print("Nothing to generate.")
        return

    print(f"Headshots to generate: {len(indices)}")

    if args.dry_run:
        for idx in indices:
            prompt = _build_prompt(idx)
            print(f"  [{idx:03d}] seed={_idx_to_seed(idx)}\n        {prompt}\n")
        return

    try:
        import torch
        from diffusers import AutoPipelineForText2Image
        from diffusers.utils import logging as diffusers_logging
        from PIL import Image
    except ImportError as e:
        sys.exit(
            f"Missing dependency: {e}\n\n"
            "Install with:\n"
            "  pip install torch torchvision --index-url https://download.pytorch.org/whl/cpu\n"
            "  pip install diffusers transformers accelerate pillow\n"
        )

    flux = _is_flux(args.model)
    device, dtype = _pick_device(torch, flux=flux)
    print(f"Loading {args.model} on {device} ({dtype}) …")

    diffusers_logging.set_verbosity_error()

    pipe = AutoPipelineForText2Image.from_pretrained(
        args.model, torch_dtype=dtype, use_safetensors=True,
        token=args.hf_token or None,
    ).to(device)
    pipe.set_progress_bar_config(disable=True)
    print("Model loaded.\n")

    errors = 0
    for n, idx in enumerate(indices, 1):
        out_path = out_dir / f"{idx:03d}.webp"
        prompt = _build_prompt(idx)
        seed = _idx_to_seed(idx)

        print(f"[{n}/{len(indices)}] {idx:03d} ", end="", flush=True)
        try:
            generator = torch.Generator(device=device).manual_seed(seed)
            gen_kwargs = dict(
                prompt=prompt,
                num_inference_steps=args.steps,
                width=_GEN_SIZE,
                height=_GEN_SIZE,
                generator=generator,
            )
            if flux:
                gen_kwargs["guidance_scale"] = 0.0
            else:
                gen_kwargs["negative_prompt"] = _NEGATIVE
                gen_kwargs["guidance_scale"] = 9.0

            result = pipe(**gen_kwargs)
            img = result.images[0]
            if img.size != (OUT_SIZE, OUT_SIZE):
                img = img.resize((OUT_SIZE, OUT_SIZE), Image.LANCZOS)
            _save_atomic(img, out_path)
            print("✓")
        except Exception as exc:
            print(f"FAILED: {exc}")
            errors += 1

    print(f"\nDone. {len(indices) - errors} generated, {errors} errors → {out_dir}")


if __name__ == "__main__":
    main()
