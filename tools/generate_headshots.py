"""
Generate fake corporate headshots using SDXL on Apple Silicon (MPS).

Usage:
    pip install torch diffusers transformers accelerate pillow
    python tools/generate_headshots.py

Outputs ~400 JPEGs to acpwb/static/img/headshots/
"""

import json
import random
import sys
from pathlib import Path

OUTPUT_DIR = Path(__file__).parent.parent / "acpwb" / "static" / "img" / "headshots"
MANIFEST_PATH = OUTPUT_DIR / "manifest.json"
NUM_IMAGES = 400
MODEL_ID = "stabilityai/stable-diffusion-xl-base-1.0"

AGES = [
    "25 year old", "28 year old", "32 year old", "36 year old",
    "40 year old", "44 year old", "48 year old", "52 year old",
    "56 year old", "60 year old", "63 year old",
]
GENDERS = ["man", "man", "woman", "woman"]  # rough 50/50
ETHNICITIES = [
    "white American", "Black American", "Hispanic American",
    "East Asian American", "South Asian American", "Middle Eastern American",
    "mixed race American",
]
HAIR = [
    "short dark hair", "short brown hair", "short blonde hair", "short gray hair",
    "short black hair", "short red hair", "close-cropped black hair",
    "medium length dark hair", "medium length brown hair", "medium length blonde hair",
    "long dark hair", "long brown hair", "long black hair", "long blonde hair",
    "natural hair", "locs", "bald", "salt and pepper hair", "silver hair",
]
ATTIRE = [
    "navy suit and tie", "dark blazer and white shirt", "gray suit",
    "business casual blazer", "white dress shirt", "charcoal suit",
    "professional blouse", "navy blazer", "black blazer",
    "light blue dress shirt", "burgundy blazer", "dark turtleneck",
]

BASE_PROMPT = (
    "professional corporate headshot portrait of a {age} {ethnicity} {gender}, "
    "{hair}, wearing {attire}, "
    "neutral light gray background, soft studio lighting, "
    "sharp focus on face, photorealistic, shot on Canon 5D, 85mm lens, "
    "LinkedIn profile photo, high resolution"
)
NEGATIVE_PROMPT = (
    "cartoon, illustration, anime, painting, watermark, text, logo, "
    "blurry, out of focus, low quality, deformed face, ugly, disfigured, "
    "extra people, crowd, multiple faces, collage, grid, contact sheet, "
    "harsh shadows, dramatic lighting, outdoor, casual clothing, sunglasses"
)


def build_prompt(rng):
    return BASE_PROMPT.format(
        age=rng.choice(AGES),
        ethnicity=rng.choice(ETHNICITIES),
        gender=rng.choice(GENDERS),
        hair=rng.choice(HAIR),
        attire=rng.choice(ATTIRE),
    )


def main():
    try:
        import torch
        from diffusers import StableDiffusionXLPipeline
    except ImportError:
        print("Missing dependencies. Run:")
        print("  pip install torch diffusers transformers accelerate pillow")
        sys.exit(1)

    if not torch.backends.mps.is_available():
        print("MPS not available — are you on Apple Silicon?")
        sys.exit(1)

    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    # Find which images already exist so we can resume interrupted runs
    existing = {int(p.stem) for p in OUTPUT_DIR.glob("*.jpg") if p.stem.isdigit()}
    to_generate = [i for i in range(NUM_IMAGES) if i not in existing]

    if not to_generate:
        print(f"All {NUM_IMAGES} images already exist in {OUTPUT_DIR}")
        return

    print(f"Loading {MODEL_ID} (~7GB download on first run) ...")
    pipe = StableDiffusionXLPipeline.from_pretrained(
        MODEL_ID,
        torch_dtype=torch.float32,
        use_safetensors=True,
    )
    pipe = pipe.to("mps")
    pipe.enable_attention_slicing()

    print(f"Generating {len(to_generate)} images (skipping {len(existing)} existing) ...")

    manifest = {}
    if MANIFEST_PATH.exists():
        manifest = json.loads(MANIFEST_PATH.read_text())

    for idx in to_generate:
        rng = random.Random(idx)  # deterministic per index
        prompt = build_prompt(rng)
        out_path = OUTPUT_DIR / f"{idx:03d}.jpg"

        print(f"  [{idx+1}/{NUM_IMAGES}] {out_path.name}: {prompt[:80]}...")

        result = pipe(
            prompt=prompt,
            negative_prompt=NEGATIVE_PROMPT,
            num_inference_steps=35,
            guidance_scale=8.0,
            height=1024,
            width=1024,
            generator=torch.Generator("mps").manual_seed(idx),
        )

        img = result.images[0]
        img.save(out_path, "JPEG", quality=88, optimize=True)

        manifest[str(idx)] = {"prompt": prompt, "path": f"img/headshots/{idx:03d}.jpg"}

        # Write manifest after each image so progress survives interruption
        MANIFEST_PATH.write_text(json.dumps(manifest, indent=2))

    print(f"\nDone. {NUM_IMAGES} headshots saved to {OUTPUT_DIR}")
    print(f"Manifest: {MANIFEST_PATH}")


if __name__ == "__main__":
    main()
