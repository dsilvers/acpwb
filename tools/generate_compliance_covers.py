#!/usr/bin/env python3
"""
Generate course cover images for the /compliance/ ComplianceCenter with
FLUX.1-schnell, seeded from the course catalog itself.

Each course type in acpwb_go/compliance/data/course_types.txt gets
--per-topic images (default 6, so ~1,000 across ~170 topics). Prompts are
built from the topic's question categories (its "mix") — a CPR course draws
AED-cabinet and first-aid-station scenes, a hybrid-work course draws
hot-desk neighborhoods, a union course draws empty break rooms and bulletin
boards — combined with the course's own locations (data/vars/PLACE.txt), a
time of day, and a camera angle, so no two covers repeat.

No people and no writing: scenes describe only objects and empty spaces,
and avoid screens, documents, signs, and anything else that invites
garbled text. FLUX schnell ignores negative prompts (guidance 0), so the
scene wording itself has to keep people and lettering out.

Outputs:
    acpwb/static/img/compliance-covers/<topic-slug>-<n>.webp   (1024×576)
    acpwb/static/img/compliance-covers/manifest.json           (prompt per file)
    acpwb_go/compliance/data/compliance_covers.txt             (file list acpwb_go embeds)

Usage:
    python tools/generate_compliance_covers.py --dry-run | head -40   # preview prompts
    python tools/generate_compliance_covers.py                        # everything missing
    python tools/generate_compliance_covers.py --per-topic 3 --limit 50
    python tools/generate_compliance_covers.py --topics cpr-and-aed-awareness,union-election-neutrality-briefing
    python tools/generate_compliance_covers.py --list-only            # just rewrite the Go file list
    python tools/generate_compliance_covers.py --report-pool          # ~1,400 covers for generated reports
    python tools/generate_compliance_covers.py --report-covers        # replace the 100 report covers in place
                                                                      # (originals are in git: git checkout acpwb/static/img/report-covers)

Then rebuild acpwb_go (the list is embedded) and run collectstatic.
Install: pip install torch diffusers transformers accelerate pillow sentencepiece protobuf
"""
import argparse
import hashlib
import json
import os
import random
import re
import sys
import tempfile

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from cover_pool_subjects import POOL_SETTINGS, POOL_SUBJECTS  # noqa: E402
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
REPORT_DIR = ROOT / 'acpwb' / 'static' / 'img' / 'report-covers'
REPORT_SIZE = (1280, 704)  # keep the existing report covers' size
DATA = ROOT / 'acpwb_go' / 'compliance' / 'data'
OUT_DIR = ROOT / 'acpwb' / 'static' / 'img' / 'compliance-covers'
GO_LIST = DATA / 'compliance_covers.txt'

WIDTH, HEIGHT = 1024, 576            # saved size (16:9)
GEN_WIDTH, GEN_HEIGHT = 1360, 768    # FLUX renders larger, then downscales

# ── Scenes per question category ──────────────────────────────────────────
# Objects and empty spaces only. No people, no screens with content, no
# paper with writing, no signs.

SCENES = {
    'safety': [
        'a red fire extinguisher mounted on a clean corridor wall beside a closed fire door',
        'a yellow wet-floor cone on a freshly mopped polished concrete floor in an empty hallway',
        'an emergency eyewash station mounted above a stainless steel sink in a spotless utility room',
        'a neat row of white hard hats on hooks above a bench in an empty equipment room',
        'an illuminated green exit light above a steel stairwell door at the end of a quiet corridor',
    ],
    'firstaid': [
        'a wall-mounted white first aid cabinet with a green cross shape, slightly open, rolls of gauze and bandages neatly stacked inside',
        'an open first aid kit on a clean break room counter, bandages, gauze pads, and a pair of trauma shears arranged neatly',
        'a folded grey wool blanket and a first aid bag on a chair in an empty office wellness room',
        'a tourniquet and rolled gauze resting on a stainless steel tray under bright clinical light',
    ],
    'cpr': [
        'a bright yellow defibrillator cabinet mounted on a white lobby wall beside the elevators',
        'an automated external defibrillator in an open protective case resting on a polished conference table',
        'a wall-mounted defibrillator box beside a drinking fountain in a quiet office corridor',
    ],
    'threat': [
        'a long empty office corridor with closed solid wooden doors, calm even lighting',
        'a sturdy conference room door closed and locked, heavy table pushed nearby, quiet empty room',
        'an empty stairwell with a closed steel fire door and concrete steps, calm and orderly',
        'a dim interior storage room with metal shelving and a closed door, a single overhead light',
    ],
    'violence': [
        'a calm empty reception desk with a closed glass partition and a single potted plant',
        'a well-lit parking structure at dusk with evenly spaced lights and empty bays',
        'a quiet security desk in an empty marble lobby with a turnstile row',
    ],
    'bombthreat': [
        'a plain cardboard parcel sitting alone on a mailroom table, rows of empty sorting cubbies behind it',
        'an empty mailroom with wooden sorting cubbies and a canvas mail cart',
        'a desk phone on a tidy reception desk in an empty lobby, soft morning light',
    ],
    'weather': [
        'dark storm clouds over a glass office tower, rain streaking the windows',
        'an interior windowless conference room with stacked chairs and a flashlight on the table',
        'a corporate campus at dusk under a heavy green-grey storm sky, trees bending in the wind',
        'snow piled along an empty office parking lot under orange streetlights',
    ],
    'elevator': [
        'brushed steel elevator doors in a marble lobby, polished floor reflections',
        'the interior of an empty modern elevator car with brushed metal walls and a handrail',
        'an elevator machine room with large motors and cables, clean industrial lighting',
    ],
    'hazcom': [
        'a row of unlabeled spray bottles and cleaning supplies on a janitorial closet shelf',
        'a yellow chemical storage cabinet with closed doors in a clean utility room',
        'a laser printer with its toner drawer pulled open in a quiet copy room',
    ],
    'bloodborne': [
        'a wall-mounted red sharps container beside a stainless sink in a clean restroom',
        'a box of blue nitrile gloves and a spill kit on a clinic-style counter',
    ],
    'badge': [
        'a blank white access card on a navy lanyard resting on a glass desk',
        'a row of glass security turnstiles in a modern marble lobby',
        'a small black card reader beside a frosted glass office door, soft blue indicator light',
        'a security office counter with a card printer and a tray of blank badge cards',
    ],
    'visitor': [
        'a modern reception desk with an empty visitor chair, a vase of flowers, and a marble wall',
        'a quiet visitor waiting area with leather armchairs and a low coffee table',
        'a lobby with glass doors, a revolving door, and morning sunlight across the floor',
    ],
    'security': [
        'a locked server rack cage behind chain-link fencing, blue indicator lights',
        'a laptop closed and secured with a steel cable lock to a desk',
        'a security camera dome mounted on the ceiling of an empty corridor',
        'a row of locked grey steel filing cabinets in a records room',
    ],
    'phishing': [
        'a closed laptop with a fishing lure resting on its lid, dark wood desk, moody light',
        'a desk with a closed laptop, a cup of coffee, and a steel fishing hook on a coaster, editorial still life',
        'an empty desk with a laptop at night, the room lit only by a desk lamp',
    ],
    'email': [
        'a tidy desk with a closed laptop, a mug, and an empty in-tray',
        'a row of brass mail slots in a wood-paneled corporate mailroom',
    ],
    'cleandesk': [
        'a perfectly clear white desk with only a closed laptop and a single plant',
        'a locked under-desk pedestal drawer with a small key, clean office',
        'a row of empty identical desks at the end of the day, chairs pushed in, soft evening light',
        'a confidential shredding console bin with a slot in a quiet copy room',
    ],
    'records': [
        'tall rolling archive shelves filled with uniform grey storage boxes',
        'rows of identical bankers boxes on metal shelving in a records room',
        'a document shredder beside a stack of sealed boxes in a storage room',
        'a long corridor of compact mobile shelving units in a quiet archive',
    ],
    'privacy': [
        'a frosted glass privacy wall in an empty modern office, soft diffused light',
        'a padlock resting on a closed leather folio on a dark desk, editorial still life',
        'a locked steel cabinet in a quiet human resources office',
    ],
    'ai': [
        'a sleek server room aisle with rows of glowing blue lights, symmetrical perspective',
        'abstract glowing network of light filaments in a dark glass-walled room',
        'a futuristic empty office with curved white desks and soft ambient lighting',
        'a brass compass resting on a closed laptop, symbolic still life about guidance',
    ],
    'ethics': [
        'a brass balance scale on a mahogany desk in a wood-paneled office',
        'a single chess piece on a polished conference table, dramatic window light',
        'a fountain pen resting on a blank sheet of heavy paper, warm desk lamp',
    ],
    'conduct': [
        'an empty corporate atrium with tall windows and a polished stone floor',
        'a long conference table with neatly aligned empty chairs and water glasses',
        'a quiet library-style corporate lounge with leather chairs and bookshelves',
    ],
    'harassment': [
        'a quiet private meeting room with two empty chairs and a box of tissues on a small table',
        'a calm human resources office with a round table and soft lamp light',
        'an empty hallway with an open door into a bright private office',
    ],
    'gifts': [
        'a wrapped gift box with a satin ribbon on an office desk',
        'a fruit basket on a reception counter in a corporate lobby',
        'a small potted orchid tied with a satin ribbon on an office desk',
    ],
    'insider': [
        'an empty trading floor with rows of dark monitors at night',
        'a glass boardroom high above a city at night, chairs pushed in',
        'a closed leather portfolio and a brass desk lamp in a dark executive office',
    ],
    'export': [
        'stacked shipping containers at a port at dusk, gantry cranes overhead',
        'wooden shipping crates on pallets in a clean warehouse loading bay',
        'a cargo plane on an empty airport apron at dawn',
    ],
    'procurement': [
        'neatly stacked cardboard boxes on pallets in a corporate supply room',
        'a loading dock with a roll-up door half open and a hand truck',
    ],
    'travel': [
        'a carry-on suitcase beside an airport gate window with a parked jet outside',
        'an empty hotel business center with a desk lamp and leather chair',
        'a rental car key and a coffee cup on a hotel desk at sunrise',
        'an empty airport terminal walkway with floor-to-ceiling windows',
    ],
    'kitchen': [
        'a shared office refrigerator with its door slightly open, neatly stacked lunch containers inside',
        'a corporate break room with a coffee machine, mugs on hooks, and a clean counter',
        'a microwave and a kettle on a spotless office kitchenette counter',
        'a dish rack with clean mugs beside a stainless sink in an office kitchen',
    ],
    'meetings': [
        'an empty conference room with chairs around a long table, glass walls, afternoon light',
        'a small huddle room with four chairs and a round table, a pitcher of water',
        'a boardroom with a speakerphone in the center of a polished table',
    ],
    'calendar': [
        'a wall clock above an empty conference room door, soft corridor lighting',
        'a desk with a closed laptop, a brass desk clock, and a cup of tea',
        'an hourglass on a polished conference table, chairs blurred in the background',
    ],
    'hotdesk': [
        'rows of identical unassigned desks with monitors turned off and chairs pushed in, open-plan office',
        'a wall of small grey personal lockers in a modern office',
        'an empty hot-desking neighborhood with plants dividing clusters of white desks',
    ],
    'hybrid': [
        'a half-empty open-plan office on a sunny weekday, every other desk unoccupied',
        'a video conference room with a large camera bar and empty chairs facing a blank wall display',
        'a laptop bag resting on an empty desk next to a window overlooking the city',
    ],
    'remote': [
        'a tidy home office desk by a window with a laptop, a plant, and a mug',
        'a kitchen table set up as a workspace with a laptop and headphones, morning light',
        'a cozy spare-room home office with a bookshelf and a desk lamp at dusk',
    ],
    'rto': [
        'an office lobby on a Monday morning with sunlight and empty turnstiles',
        'freshly cleaned empty desks waiting under bright office lights',
        'an office coffee station with a row of clean mugs, morning light',
    ],
    'pto': [
        'a closed laptop on a desk beside a pair of sunglasses and a beach hat',
        'an empty office chair facing a window with a view of a lake on a sunny day',
        'a suitcase and a straw hat by an office door at the end of the day',
    ],
    'leave': [
        'a quiet empty desk with the chair pushed in and a plant in the sun',
        'a calm wellness room with a reclining chair and soft lighting',
    ],
    'parental': [
        'a bright private lactation room with a comfortable armchair, a small refrigerator, and a sink',
        'a small potted plant and a baby rattle on an empty office desk',
        'a calm nursing room in an office with soft lamp light and an armchair',
    ],
    'union': [
        'an empty break room with long tables, folding chairs, and a cork bulletin board with blank pinned cards',
        'a factory floor at dawn with idle machinery and long shadows',
        'a row of empty folding chairs set up in a community hall',
        'a time clock and card rack on a plain wall near a loading dock',
    ],
    'pay': [
        'a neat stack of coins and a brass scale on a dark desk, editorial still life',
        'a calculator and a fountain pen on a polished desk, warm lamp light',
        'a leather wallet and a stack of folded bills beside a closed ledger on a walnut desk',
        'a glass jar of coins on a windowsill in a quiet office, soft light',
        'rows of empty desks in a finance department at dusk, chairs pushed in',
        'a brass desk lamp lighting a closed ledger and a mechanical pencil, dark wood desk',
    ],
    'governance': [
        'an empty mahogany boardroom with leather chairs and a city view',
        'a gavel resting on a polished wooden table in a formal meeting room',
    ],
    'investigations': [
        'a small quiet interview room with a table, two chairs, and a window',
        'a locked evidence-style storage cabinet in a plain office',
    ],
    'accessibility': [
        'a wide automatic door with a push-button opener at a modern office entrance',
        'a gently sloped ramp beside stone steps at a corporate entrance',
        'an adjustable-height desk raised in an empty modern office',
    ],
    'ergonomics': [
        'an ergonomic office chair at a height-adjustable desk with a monitor arm',
        'a keyboard tray, a vertical mouse, and a footrest under a tidy desk',
    ],
    'wellness': [
        'an empty office fitness center with treadmills facing a window',
        'a quiet meditation room with floor cushions and plants',
        'a bowl of fresh fruit on a break room counter in morning light',
    ],
    'parking': [
        'an empty multi-level parking structure with painted bays and concrete pillars',
        'an electric vehicle charging station in an office parking garage',
        'a parking gate arm at the entrance of an office garage at dusk',
    ],
    'continuity': [
        'a backup generator in a clean mechanical room',
        'emergency flashlights and supplies on a shelf in an office storage room',
        'a dark office floor lit only by emergency lighting during a power outage',
    ],
    'printing': [
        'a large office multifunction printer in a quiet copy room',
        'reams of blank paper stacked beside a printer in a supply alcove',
    ],
}

# Scenes for the newer compliance banks (fire, heat, ladders, employment law,
# collaboration tools, celebrations). Topics only; the report pool doesn't
# use these categories.
SCENES.update({
    'fire': [
        'a red fire extinguisher mounted on a clean office wall beside a fire alarm pull station',
        'a row of fire extinguishers on a shelf in a building maintenance room',
        'a glass-fronted fire hose cabinet in an empty office corridor',
        'an extinguisher beside a stainless steel commercial kitchen range',
        'a stairwell with a heavy fire door held closed and a red handrail',
    ],
    'heat': [
        'an empty loading dock in harsh midday summer sun with heat shimmer',
        'a shaded break area with a water cooler and paper cups beside a warehouse',
        'a pop-up canopy over a cooler of water bottles on a hot paved lot',
        'a rooftop of HVAC units under a blazing summer sky',
        'a box fan and a pitcher of ice water on an office desk in summer light',
    ],
    'ladder': [
        'an aluminum extension ladder leaning against a brick wall at the correct angle',
        'a fiberglass stepladder fully opened in an empty office hallway',
        'a rolling library ladder against tall shelves of blank binders',
        'a two-step safety stool beside a high storage shelf',
        'ladders hanging neatly on a wall in a facilities workshop',
    ],
    'lockout': [
        'a red padlock and blank tag on an industrial electrical disconnect',
        'a lockout station board with rows of padlocks and keys',
        'a clean factory machine with a safety guard and an emergency stop button',
        'a group lockout box with several colored padlocks',
        'an electrical panel room with gray switchgear',
    ],
    'mentalhealth': [
        'a calm quiet room with a soft armchair, a plant, and warm lamp light',
        'two comfortable chairs by a window in a small private office',
        'a cup of tea on a side table beside a window with rain outside',
        'a peaceful office courtyard garden with a bench',
        'a softly lit wellness room with floor cushions',
    ],
    'wage': [
        'a wall-mounted time clock with a rack of blank punch cards',
        'a desk with a calculator, a blank timesheet, and a cup of coffee',
        'a payroll office with gray filing cabinets and a wall clock',
        'a wristwatch and a stack of blank pay envelopes on a desk',
        'a wall clock above an empty office at 6 p.m. with long shadows',
    ],
    'hiring': [
        'two chairs facing each other across a small interview table with water glasses',
        'a bright reception area with a sofa and fresh flowers',
        'a glass-walled interview room with a round table',
        'a neat stack of blank folders and a pen on a meeting table',
        'an empty lobby chair beside a briefcase',
    ],
    'whistleblower': [
        'a single desk lamp lit in a dark office at night',
        'a closed manila envelope on a polished table',
        'an old corded telephone on a desk beside a window',
        'a quiet empty corridor with one office door slightly open',
        'a locked drop box mounted on an office wall',
    ],
    'layoff': [
        'an empty desk with a single cardboard box packed on top',
        'a row of empty chairs in a large conference room after hours',
        'a quiet open-plan office with many unoccupied desks',
        'a coat rack with empty hangers by an office door',
        'a parking lot with only a few cars at dusk',
    ],
    'chat': [
        'a laptop with a dark screen and a coffee mug on a tidy desk',
        'a phone lying face down beside a keyboard',
        'a pair of headphones resting on a laptop on a shared desk',
        'a desk with two blank monitors glowing softly at night',
    ],
    'video': [
        'a home office desk with a webcam, ring light, and a plant',
        'a video conference room with a large blank display and empty chairs',
        'a laptop on a stack of books at eye level beside a window',
        'a whiteboard wiped clean behind an empty desk chair',
    ],
    'openoffice': [
        'a long open-plan office of bench desks with acoustic panels',
        'an empty office phone booth pod with its door open',
        'a quiet zone of desks with desk lamps and plants',
        'a wall-mounted thermostat in an open office',
    ],
    'holiday': [
        'an office lobby decorated with string lights and a wrapped gift table',
        'a buffet table set for an office party with empty glasses',
        'a stack of wrapped gifts on an office desk',
        'a conference room set for a party with balloons and a cake',
    ],
    'pets': [
        'a dog bed and a water bowl beside an office desk',
        'a leash hanging on a coat hook by a desk',
        'a basket of dog toys in an office lounge',
        'a pet-friendly office patio with water bowls',
    ],
    'performance': [
        'a desk with a closed laptop, a notebook, and two coffee cups',
        "a manager's office with two chairs facing a desk",
        'a small meeting room with a whiteboard covered in blank sticky notes',
        'a desk calendar beside a closed folder and a pen',
    ],
    'social': [
        'a smartphone on a desk beside a cup of coffee and a notebook',
        'a laptop and a camera on a tidy desk by a window',
        'a corporate communications office with blank screens',
        'a sunny desk set up for a photo with a plant and a mug',
    ],
})

# Extra scenes used only for the report covers' categories.
SCENES.update({
    'esg': [
        'a modern office building with a green living wall and rooftop solar panels under a clear blue sky',
        'wind turbines on a ridge seen through the floor-to-ceiling window of an empty boardroom',
        'a rooftop garden with raised planters on a corporate tower, city skyline behind',
        'a row of bicycles in a bike rack outside a glass office entrance, morning light',
        'a rain garden with native grasses at the entrance of a corporate campus',
    ],
    'economic': [
        'a container port at dusk with gantry cranes and stacked shipping containers',
        'a financial district skyline at blue hour reflected in a still water feature',
        'long-exposure light trails on a downtown avenue between glass towers at night',
        'an aerial view of a dense downtown commercial district at golden hour',
    ],
    'leadership': [
        'an empty auditorium with a single lectern lit by a spotlight on stage',
        'a corner executive office with a clear desk and a leather chair facing a city view',
        'a mountain-lodge conference room with floor-to-ceiling windows and empty chairs',
    ],
    'talent': [
        'two empty chairs facing each other across a small glass table in a bright meeting room',
        'a new-hire welcome kit on a desk: a blank notebook, a lanyard, and a coffee mug',
        'the glass entrance of a modern office building on a sunny morning',
    ],
})

# Report cover categories (static/img/report-covers/manifest.json) → scene
# categories.
REPORT_CATEGORIES = {
    'Compensation': ['pay', 'governance', 'insider'],
    'Workforce': ['hotdesk', 'hybrid', 'rto', 'conduct'],
    'Benefits': ['wellness', 'parental', 'pto', 'accessibility'],
    'Governance': ['governance', 'conduct', 'records'],
    'ESG': ['esg', 'weather'],
    'Retention': ['wellness', 'pto', 'meetings', 'leadership'],
    'Technology': ['ai', 'security'],
    'Economic': ['economic', 'export', 'insider'],
    'Diversity & Inclusion': ['accessibility', 'conduct', 'meetings'],
    'Compliance': ['records', 'security', 'privacy', 'cleandesk'],
    'Leadership': ['leadership', 'governance', 'meetings'],
    'Talent Acquisition': ['talent', 'visitor', 'meetings'],
}

# Round three: categories that look like their own place.
SCENES.update({
    'winter': [
        'a corporate parking lot after fresh snowfall, salted walkway leading to a glass entrance',
        'an office entrance with a snow-dusted canopy and a rubber mat inside the doors, early morning',
        'a row of icicles hanging from an office building roofline against a pale winter sky',
        'a snow shovel and a bucket of de-icing salt beside a brick office entrance',
        'a frosted office window looking out over a snowy corporate campus',
    ],
    'electrical': [
        'a tidy electrical panel room with closed grey breaker cabinets and a yellow floor line',
        'a single heavy-duty power strip mounted neatly under an empty desk, cables bundled',
        'coiled orange extension cords hanging on hooks in a facilities workshop',
        'a clean office floor box with covered outlets in polished concrete',
        'a utility closet with conduit running neatly along a painted wall',
    ],
    'driving': [
        'a clean white fleet sedan parked alone in a corporate garage bay',
        'a quiet highway at dawn seen from a car interior, dashboard dark and unmarked',
        'a row of identical company vans parked at a depot under overcast sky',
        'a car key fob and a closed logbook on the passenger seat of a parked car',
        'a rain-slicked two-lane road through farmland seen through a windshield',
    ],
    'warehouse': [
        'a tall warehouse aisle of pallet racking with blank cardboard boxes, yellow floor lines',
        'a loading dock with a roll-up door open to a parked trailer, wheel chocks in place',
        'a parked forklift beside stacked plain pallets in a clean distribution center',
        'a pallet jack and stretch-wrapped pallets beside a dock leveler',
        'a mezzanine walkway with yellow guardrails above a quiet warehouse floor',
    ],
    'spaces': [
        'an open steel access hatch in a concrete floor, guardrail and tripod beside it',
        'a utility vault ladder descending into a dimly lit concrete chamber',
        'a portable gas monitor clipped to a railing beside an open tank manhole',
        'a narrow mechanical crawlspace lit by a single work lamp',
        'a ventilation blower and yellow duct hose beside a sewer access cover',
    ],
    'ppe': [
        'a wall rack of white hard hats, safety glasses and hi-visibility vests by a doorway',
        'safety goggles, earmuffs and work gloves laid out neatly on a steel bench',
        'a PPE vending cabinet with clear doors in a clean plant corridor',
        'steel-toe boots on a mat beside a locker room bench',
        'a row of respirators hanging on hooks above a stainless sink',
    ],
    'medical': [
        'a wall-mounted AED cabinet in a bright office corridor, door closed',
        'an on-site health room with an exam table and a closed supply cabinet',
        'a stretcher-ready first-response bag on a shelf beside an elevator lobby',
        'a corporate wellness clinic waiting area with empty chairs and soft light',
        'an oxygen cylinder cart and blanket on a clean clinic counter',
    ],
    'earthquake': [
        'a sturdy oak desk in an empty office with shelves strapped to the wall',
        'a bookcase anchored to a wall with steel brackets in a quiet office',
        'an emergency supply cabinet with water jugs and flashlights in a stairwell landing',
        'a seismically braced server rack with cross bracing in a cool data room',
        'a calm open courtyard away from buildings, used as an assembly area',
    ],
    'foodsafety': [
        'a catering buffet line with covered chafing dishes and a probe thermometer on a tray',
        'a stainless commercial kitchen with a walk-in cooler door slightly open',
        'a tray of iced salads on a buffet table in a bright conference foyer',
        'a hand-washing sink with soap and paper towels in a clean office kitchen',
        'color-coded cutting boards stacked on a stainless prep table',
    ],
    'potluck': [
        'a long break-room table covered with casserole dishes and serving spoons, lids on',
        'a slow cooker and a stack of paper plates on an office kitchen counter',
        'a potluck spread on a folding table in a decorated conference room',
        'a pie, a salad bowl and a tray of cookies under plastic covers on a counter',
        'mismatched serving dishes cooling on a windowsill in an office kitchen',
    ],
    'mail': [
        'a corporate mailroom with wooden pigeonhole sorting shelves and a rolling cart',
        'a postal scale and stacks of plain parcels on a mailroom counter',
        'a row of blank metal mailboxes in an office building lobby',
        'canvas mail bins on wheels in a back-of-house corridor',
        'a mailroom screening table with a lamp and gloves beside a plain package',
    ],
    'deliveries': [
        'a reception desk buried in plain unlabeled cardboard boxes',
        'a wall of smart parcel lockers in a corporate lobby, doors closed',
        'a hand truck stacked with plain boxes beside an elevator',
        'a package room with shelves of plain parcels sorted by size',
        'a delivery entrance with a buzzer panel and a roll-up door',
    ],
    'serverroom': [
        'a server room hot aisle with tidy cable management and blue status lights',
        'a locked glass door leading into a cool data hall, raised floor tiles visible',
        'a single rack with its door open, patch cables perfectly combed',
        'a fire suppression nozzle and smoke detector on the ceiling above server racks',
        'a badge reader beside a steel data center door in a white corridor',
    ],
    'lobby': [
        'a corporate lobby with a curved reception desk and glass turnstiles',
        'a visitor badge printer and a closed sign-in tablet on a reception counter',
        'a quiet lobby seating area with leather chairs and a coffee table',
        'security turnstiles in a marble lobby seen from the revolving door',
        'a reception desk at night lit by a single pendant lamp',
    ],
    'physicalsec': [
        'a propped-open fire door with a doorstop in a quiet stairwell',
        'a badge reader on a glass security door in a bright corridor',
        'a ceiling security camera dome above an office hallway',
        'a perimeter fence and gate outside a corporate campus at dusk',
        'a key cabinet with numbered hooks, door closed, in a facilities office',
    ],
    'retro': [
        'a beige desktop computer with a CRT monitor and a floppy drive on a 1990s office desk',
        'a fax machine and a desk phone with a coiled cord on a laminate counter',
        'a box of floppy disks beside a dot-matrix printer with blank tractor-feed paper',
        'a 1990s cubicle with fabric walls, a Rolodex and an overhead projector',
        'a pager and a PalmPilot-style device on a mouse pad, warm desk lamp',
    ],
    'drones': [
        'a small quadcopter drone resting on a landing pad on a corporate rooftop',
        'a drone in an open hard case on a table in a facilities office',
        'an empty corporate campus seen from high above, drone-eye view',
        'a drone hovering above an empty parking structure at golden hour',
        'a drone controller and spare propellers on a workbench',
    ],
    'construction': [
        'a construction site beside an office tower, cranes against a clear sky, site empty',
        'an office floor under renovation with exposed ductwork and stacked drywall',
        'hard hats hanging on a site trailer wall beside rolled plans',
        'orange barriers and caution tape across a renovated office corridor',
        'scaffolding on a building facade at dawn, no workers present',
    ],
    'fieldwork': [
        'a company pickup truck parked on a gravel road beside open farmland',
        'a clipboard and hard hat on the hood of a truck at a quiet utility site',
        'a field equipment case open on a tailgate at sunrise',
        'a remote substation fence line under a wide prairie sky',
        'a rural road with a parked inspection vehicle under storm clouds',
    ],
    'nightshift': [
        'an office floor at 3 a.m. lit by a few desk lamps, city lights outside',
        'a brightly lit break room at night with vending machines and empty tables',
        'a well-lit parking structure at night with a single parked car',
        'a security desk at night facing a dark glass lobby',
        'a quiet production floor under night lighting, machines idle',
    ],
    'coffee': [
        'an office coffee station with a drip machine, mugs on hooks and a sugar jar',
        'an espresso machine and stacked cups on a sleek office kitchen counter',
        'a nearly empty glass coffee carafe on a warmer, morning light',
        'a café-style office coffee bar with stools and pendant lights',
        'a row of mismatched office mugs drying on a rack',
    ],
    'celebrations': [
        'a conference room decorated with balloons and a sheet cake on the table',
        'paper streamers and an unlit candle on a cupcake in an office kitchen',
        'a break room set for a retirement party with punch bowl and plates',
        'a banner-free festive office lobby with gold balloons and string lights',
        'a gift bag and a card envelope on an empty decorated desk',
    ],
    'costumes': [
        'carved pumpkins and paper bats decorating an office reception desk',
        'a costume rack with a cape and a witch hat in an office closet',
        'a conference room decorated with fake cobwebs and orange string lights',
        'a candy bowl shaped like a cauldron on a cubicle shelf',
        'a mask and a wizard hat on an empty office chair',
    ],
    'sports': [
        'an empty softball field at dusk with a dugout and corporate campus behind',
        'a bin of kickballs and orange cones beside an office green',
        'a ping-pong table in an office game room, paddles resting on it',
        'a trophy shelf with plain unengraved cups in an office break room',
        'a corporate fitness court with basketball hoop at sunrise',
    ],
    'kids': [
        'a conference room set up with crayons, paper and small chairs for kids day',
        'a small desk with a toy laptop beside a grown-up desk in an office',
        'a box of building blocks and picture books in a bright office lounge',
        'child-size safety vests hanging by an office door',
        'a juice box and a snack plate on a reception counter',
    ],
    'vr': [
        'a VR headset and two controllers on a white table in an empty training room',
        'a cleared padded floor area with a boundary mat in a training studio',
        'a rack of VR headsets charging in a learning lab',
        'a swivel chair in the center of an empty VR room with soft lighting',
        'a VR headset hanging on a hook beside a sanitizing cabinet',
    ],
    'photos': [
        'a camera on a tripod facing an empty office backdrop and softbox lights',
        'a team-photo spot in an office atrium with a stool and ring light',
        'a camera and lenses on a conference table beside a closed laptop',
        'a photo studio corner in an office with a seamless grey backdrop',
        'framed blank photos on an office corridor wall',
    ],
    'agile': [
        'a team room with a wall of blank sticky notes in columns and a standing table',
        'a whiteboard wall with empty colored columns and magnets in a bright project room',
        'a standing desk cluster with a small timer and a bell in an agile space',
        'a retro room with beanbags and a blank flip chart',
        'colorful blank index cards spread on a project table',
    ],
    'spreadsheets': [
        'a dual-monitor desk with both screens dark and a calculator beside the keyboard',
        'a finance analyst desk with stacked binders and a mechanical keyboard, monitors off',
        'a laptop closed on a tidy desk beside a mug and a printed grid of blank cells',
        'an empty open-plan finance floor at closing time with monitors off',
        'a single desk lamp over a keyboard and calculator late at night',
    ],
    'aviation': [
        'an empty airport apron at dawn with a parked white business jet and wheel chocks in place',
        'a quiet general aviation hangar with a twin-engine airplane and a polished concrete floor',
        'an empty airline cockpit with the instrument panel dark, seen from the jump seat',
        'a long runway at sunset seen from the threshold, edge lights glowing',
        'a corporate flight department lounge with leather chairs overlooking the ramp',
    ],
    'ramp': [
        'an airport ramp with orange safety cones and a baggage tug parked beside a regional jet',
        'a FOD collection bin and a row of chocks on the edge of an airport apron',
        'a jet bridge docked to an empty airliner at dusk, ramp lights on',
        'a fuel truck parked at a distance from a business jet on a wet apron',
        'hearing protection and high-visibility vests hanging in a ramp break room',
    ],
    'cabin': [
        'an empty airliner cabin with rows of seats and overhead bins open',
        'an over-wing exit row in an empty airliner cabin, window shade up',
        'an empty business jet cabin with cream leather seats and a polished table',
        'yellow oxygen masks hanging from the ceiling of an empty airliner cabin',
        'a galley of an airliner with carts stowed and latched',
    ],
    'innovation': [
        'an innovation lab with 3D printers and blank prototypes on shelves',
        'a brainstorming room with a wall of blank sticky notes and beanbags',
        'a workshop bench with cardboard prototypes and cutting mats',
        'a glass-walled idea lab with whiteboard walls wiped clean',
        'a lightbulb-shaped pendant lamp over a long collaboration table',
    ],
})

GENERIC = [
    'a sunlit corporate atrium with a polished stone floor and tall windows',
    'an empty modern office floor at golden hour, long shadows across the carpet',
    'a glass office tower against a clear blue sky, viewed from below',
    'a quiet corporate corridor with framed abstract art and soft lighting',
    'a training room with rows of empty desks facing a blank wall',
]

# Generated question kinds and exhibit kinds fall back to related scenes.
ALIASES = {
    'agency': 'governance', 'statute': 'governance', 'wrongfact': 'governance', 'report': 'records',
    'chart': 'records', 'trivia': 'conduct', 'perch': 'meetings', 'scenario': 'conduct', 'definition': 'records',
    'exhibit': 'safety', 'ex-evac': 'safety', 'ex-tipschat': 'union', 'ex-invite': 'calendar', 'ex-fmla': 'parental',
    'ex-badgelog': 'rto', 'ex-ptostatement': 'pto', 'ex-hybridgrid': 'hybrid', 'ex-email': 'phishing',
    'ex-document': 'records', 'ex-visitorlog': 'visitor', 'ex-screening': 'export', 'ex-retention': 'records',
    'ex-printqueue': 'printing', 'ex-payrange': 'pay', 'ex-gifts': 'gifts', 'ex-fridge': 'kitchen',
    'ex-expense': 'travel', 'ex-deskmap': 'hotdesk', 'ex-desk': 'cleandesk', 'ex-calendar': 'meetings',
    'ex-ladder': 'ladder', 'ex-extinguisher': 'fire', 'ex-heatindex': 'heat', 'ex-chatthread': 'chat',
    'ex-timesheet': 'wage', 'ex-warn': 'layoff', 'ex-interview': 'hiring', 'ex-videogrid': 'video',
}

# Round three: categories that borrow a related scene set.
ALIASES.update({
    'acceptableuse': 'security', 'accommodation': 'accessibility', 'aihiring': 'ai', 'airquality': 'wellness',
    'aml': 'export', 'antitrust': 'governance', 'attendance': 'rto', 'audit': 'records', 'authority': 'governance',
    'bedbugs': 'travel', 'board': 'governance', 'brand': 'social', 'bribery': 'gifts', 'careers': 'talent',
    'chatbots': 'ai', 'civicleave': 'leave', 'civility': 'conduct', 'classification': 'wage', 'clawback': 'pay',
    'climate': 'esg', 'closures': 'weather', 'commuting': 'parking', 'conflicts': 'ethics', 'contracts': 'records',
    'crypto': 'economic', 'culture': 'conduct', 'customerdata': 'privacy', 'customers': 'meetings',
    'deferredcomp': 'pay', 'devices': 'security', 'dilemmas': 'ethics', 'disclosure': 'insider',
    'doccontrol': 'records', 'dresscode': 'conduct', 'eeo': 'hiring', 'emoji': 'chat', 'enrollment': 'wellness',
    'equity': 'pay', 'evacassist': 'accessibility', 'execcomp': 'pay', 'eyestrain': 'ergonomics',
    'finance': 'economic', 'finwell': 'pay', 'focus': 'openoffice', 'futurework': 'ai', 'gambling': 'social',
    'globalteams': 'calendar', 'goals': 'performance', 'health': 'privacy', 'hospitality': 'gifts', 'hsa': 'wellness',
    'incentives': 'pay', 'injuries': 'firstaid', 'insurance': 'records', 'intltravel': 'travel', 'ip': 'records',
    'languages': 'meetings', 'learning': 'talent', 'legalhold': 'records', 'managers': 'leadership', 'mfa': 'security',
    'network': 'security', 'networking': 'social', 'offboarding': 'layoff', 'officemove': 'warehouse',
    'offsites': 'meetings', 'onboarding': 'talent', 'pandemic': 'wellness', 'passwords': 'security',
    'payequity': 'pay', 'payratio': 'pay', 'payroll': 'wage', 'paytransparency': 'pay', 'pcard': 'procurement',
    'pettycash': 'pay', 'political': 'governance', 'presentations': 'meetings', 'privacy2': 'privacy',
    'ransomware': 'security', 'recognition': 'performance', 'reimburse': 'travel', 'relationships': 'conduct',
    'retirement': 'pay', 'sanctions': 'export', 'sayonpay': 'governance', 'signage': 'safety', 'signatures': 'email',
    'socialeng': 'phishing', 'sox': 'records', 'substances': 'wellness', 'supplychain': 'procurement',
    'sustainability': 'esg', 'thirdparty': 'procurement', 'timeoff': 'pto', 'travelbooking': 'travel',
    'volunteering': 'social', 'workation': 'remote',
    'ex-aedpads': 'cpr', 'ex-agenda': 'meetings', 'ex-alertphone': 'threat', 'ex-allergens': 'foodsafety',
    'ex-approvals': 'procurement', 'ex-boardingpass': 'travel', 'ex-bonuscalc': 'pay', 'ex-breachclock': 'privacy',
    'ex-cashlog': 'pay', 'ex-chemlabel': 'hazcom', 'ex-classlabels': 'records', 'ex-clawback': 'pay',
    'ex-commission': 'pay', 'ex-dogboard': 'pets', 'ex-dresscal': 'conduct', 'ex-dsarclock': 'privacy',
    'ex-emissions': 'esg', 'ex-evcharge': 'parking', 'ex-faxcover': 'retro', 'ex-firedrill': 'fire',
    'ex-floppy': 'retro', 'ex-foodtemps': 'foodsafety', 'ex-gantt': 'agile', 'ex-gasmonitor': 'spaces',
    'ex-hiddencols': 'spreadsheets', 'ex-holdscope': 'records', 'ex-hotelfolio': 'travel', 'ex-hsalimit': 'wellness',
    'ex-i9': 'talent', 'ex-invoice': 'procurement', 'ex-jobposting': 'hiring', 'ex-kanban': 'agile',
    'ex-lifeevent': 'parental', 'ex-liftweights': 'warehouse', 'ex-match401k': 'pay', 'ex-mealpunch': 'wage',
    'ex-meetingcost': 'meetings', 'ex-mfaprompt': 'security', 'ex-mileage': 'driving', 'ex-monitorsetup': 'ergonomics',
    'ex-napbook': 'wellness', 'ex-ndaterm': 'records', 'ex-noisemeter': 'ppe', 'ex-occupancy': 'fire',
    'ex-okr': 'performance', 'ex-options': 'pay', 'ex-orgchart': 'leadership', 'ex-parking': 'parking',
    'ex-passwords': 'security', 'ex-payratio': 'pay', 'ex-payscatter': 'pay', 'ex-paystub': 'wage',
    'ex-pcardstmt': 'procurement', 'ex-perdiem': 'travel', 'ex-planpick': 'wellness', 'ex-powerstrips': 'electrical',
    'ex-ppecheck': 'ppe', 'ex-qrposter': 'phishing', 'ex-quakeroom': 'earthquake', 'ex-quorum': 'governance',
    'ex-ransomnote': 'security', 'ex-receipt': 'travel', 'ex-recycling': 'esg', 'ex-relocation': 'travel',
    'ex-restgap': 'wage', 'ex-retirevest': 'pay', 'ex-rsvp': 'celebrations', 'ex-santatags': 'holiday',
    'ex-sayonpay': 'governance', 'ex-sds': 'hazcom', 'ex-sharing': 'privacy', 'ex-signins': 'security',
    'ex-smishing': 'phishing', 'ex-summarycomp': 'governance', 'ex-surveygroups': 'privacy',
    'ex-suspiciouspkg': 'mail', 'ex-tailgate': 'physicalsec', 'ex-tradeagenda': 'meetings', 'ex-urlbar': 'phishing',
    'ex-usrra': 'leave', 'ex-vesting': 'pay', 'ex-walkways': 'winter', 'ex-wifilist': 'remote',
    'ex-windchill': 'winter', 'ex-wire': 'procurement', 'ex-y2k': 'retro',
})

# Aviation and the diagram exhibits.
ALIASES.update({
    'faa': 'aviation', 'flightdept': 'aviation', 'sida': 'badge', 'dgr': 'hazcom',
    'ex-runway': 'aviation', 'ex-rampzones': 'ramp', 'ex-cgenvelope': 'aviation', 'ex-lithium': 'cabin',
    'ex-fuelres': 'aviation', 'ex-sterile': 'aviation', 'ex-droneagl': 'drones', 'ex-cabinexit': 'cabin',
    'ex-metar': 'aviation', 'ex-badgezones': 'physicalsec', 'ex-lightning': 'weather', 'ex-extaim': 'fire',
    'ex-sprinkler': 'fire', 'ex-panelclear': 'electrical', 'ex-extheight': 'fire', 'ex-windleak': 'hazcom',
    'ex-tornado': 'weather', 'ex-shelfload': 'warehouse', 'ex-following': 'driving', 'ex-screenview': 'privacy',
    'ex-netmap': 'security', 'ex-venn': 'meetings', 'ex-pie': 'records', 'ex-worldclock': 'calendar',
    'ex-stepladder': 'ladder', 'ex-eyewash': 'hazcom', 'ex-elevload': 'elevator', 'ex-clearance': 'parking',
    'ex-cords': 'safety', 'ex-stairs': 'fire', 'ex-ramp': 'accessibility', 'ex-lockpoints': 'lockout',
})

# Round four: new banks and the gauge, chart, plan, timeline, flowchart,
# dimension, sign, settings, file and scene exhibits borrow related scenes.
ALIASES.update({
    'accounting': 'economic', 'agediscrim': 'hiring', 'allyship': 'conduct', 'apis': 'security',
    'assetdisposal': 'serverroom', 'async': 'remote', 'awards': 'celebrations', 'background': 'hiring',
    'backups': 'serverroom', 'beneficiary': 'pay', 'bereavement': 'leave', 'biometrics': 'physicalsec',
    'bonuses': 'pay', 'boomerang': 'talent', 'botpolicy': 'serverroom', 'breach': 'privacy',
    'breaks': 'wage', 'bullying': 'harassment', 'bystander': 'harassment', 'calibration': 'performance',
    'canaries': 'security', 'carbon': 'esg', 'caregiving': 'parental', 'catering': 'potluck',
    'ccpa': 'privacy', 'certificates': 'security', 'changemgmt': 'serverroom', 'chemical': 'hazcom',
    'cloud': 'serverroom', 'cloudcost': 'serverroom', 'cobra': 'wellness', 'codereview': 'security',
    'collections': 'economic', 'complaints': 'meetings', 'conferences': 'meetings', 'contractors': 'construction',
    'cookies': 'privacy', 'copyright': 'records', 'crisiscomms': 'meetings', 'customs': 'export',
    'dataresidency': 'serverroom', 'datasets': 'records', 'decisions': 'ethics', 'devsecops': 'security',
    'disability': 'accessibility', 'discipline': 'performance', 'discrimination': 'hiring', 'dlp': 'privacy',
    'dns': 'serverroom', 'domesticviolence': 'violence', 'dotcom': 'retro', 'eap': 'mentalhealth',
    'encryption': 'security', 'energy': 'esg', 'entertainment': 'gifts', 'espp': 'pay',
    'events': 'celebrations', 'exitinterviews': 'layoff', 'facilities': 'lobby', 'fallprotection': 'construction',
    'falseclaims': 'governance', 'fantasy': 'sports', 'fatigue': 'nightshift', 'favoritism': 'ethics',
    'fax': 'retro', 'feedback': 'performance', 'fleet': 'driving', 'fmla': 'parental',
    'forklift': 'warehouse', 'fragrance': 'wellness', 'fraud': 'ethics', 'fsa': 'wellness',
    'gdpr': 'privacy', 'genai': 'ai', 'genetic': 'privacy', 'geopay': 'pay',
    'gig': 'talent', 'gossip': 'conduct', 'grants': 'economic', 'handbook': 'records',
    'hearing': 'ppe', 'helpdesk': 'security', 'hipaa': 'medical', 'holidaysched': 'holiday',
    'hostilework': 'harassment', 'hotwork': 'construction', 'humor': 'conduct', 'i9': 'hiring',
    'identity': 'security', 'idtheft': 'privacy', 'incident': 'security', 'inclusion': 'conduct',
    'integrity': 'ethics', 'internalportal': 'security', 'internships': 'talent', 'inventory': 'warehouse',
    'investor': 'insider', 'iot': 'serverroom', 'knowledge': 'records', 'lab': 'hazcom',
    'leansixsigma': 'agile', 'legacycode': 'retro', 'licensing': 'procurement', 'lifting': 'warehouse',
    'lithium': 'electrical', 'lobbying': 'governance', 'location': 'privacy', 'logging': 'serverroom',
    'lostfound': 'lobby', 'machineguarding': 'lockout', 'maintenance': 'lockout', 'media': 'meetings',
    'mentoring': 'talent', 'merit': 'pay', 'microfiche': 'retro', 'military': 'leave',
    'mna': 'insider', 'mobile': 'security', 'mobility': 'travel', 'modernslavery': 'export',
    'monitoring': 'privacy', 'moonlighting': 'ethics', 'nda': 'records', 'nearmiss': 'safety',
    'newsletters': 'email', 'noise': 'openoffice', 'noncompete': 'records', 'okrs': 'performance',
    'oncall': 'nightshift', 'opensource': 'security', 'overtime': 'wage', 'pagers': 'retro',
    'partners': 'meetings', 'patching': 'security', 'patents': 'innovation', 'pci': 'privacy',
    'perchconf': 'meetings', 'phones': 'openoffice', 'pip': 'performance', 'pregnancy': 'parental',
    'pressreleases': 'meetings', 'privilege': 'records', 'privilegedaccess': 'security', 'projects': 'agile',
    'pronouns': 'conduct', 'recall': 'warehouse', 'recruiting': 'hiring', 'references': 'hiring',
    'referrals': 'hiring', 'releases': 'agile', 'religion': 'conduct', 'relocation': 'travel',
    'research': 'records', 'respect': 'conduct', 'respirator': 'ppe', 'retaliation': 'whistleblower',
    'retention': 'talent', 'returntowork': 'firstaid', 'revenue': 'economic', 'sabbatical': 'pto',
    'salarydata': 'pay', 'salescomp': 'pay', 'sampling': 'records', 'scraping': 'serverroom',
    'secrets': 'security', 'severance': 'layoff', 'sexharassment': 'harassment', 'sickleave': 'leave',
    'skiplevel': 'leadership', 'sla': 'serverroom', 'slips': 'safety', 'spam': 'email',
    'speakup': 'whistleblower', 'sponsorships': 'celebrations', 'standups': 'agile', 'studentloans': 'pay',
    'subpoena': 'records', 'succession': 'leadership', 'surveys': 'spreadsheets', 'tax': 'economic',
    'tickets': 'security', 'timekeeping': 'wage', 'tips': 'kitchen', 'townhalls': 'meetings',
    'trademarks': 'records', 'translation': 'meetings', 'tuition': 'talent', 'unclaimed': 'economic',
    'unconsciousbias': 'hiring', 'values': 'conduct', 'vehicles': 'driving', 'vendorrisk': 'procurement',
    'vpn': 'remote', 'waste': 'esg', 'water': 'esg', 'wearables': 'wellness',
    'wikiedits': 'records', 'wires': 'economic', 'workerscomp': 'firstaid', 'y2k': 'retro',
    'ex-abcflow': 'wage', 'ex-accessaisle': 'accessibility', 'ex-accesslog': 'serverroom', 'ex-adminlog': 'security',
    'ex-aedwalk': 'cpr', 'ex-aiflow': 'ai', 'ex-airnozzle': 'lockout', 'ex-aisettings': 'ai',
    'ex-apikeys': 'security', 'ex-apikeysettings': 'security', 'ex-arcflash': 'electrical', 'ex-assetlist': 'security',
    'ex-attacklog': 'security', 'ex-attendant': 'construction', 'ex-auditfindings': 'records', 'ex-backupsettings': 'serverroom',
    'ex-batterywh': 'cabin', 'ex-breachedpw': 'security', 'ex-breachflow': 'privacy', 'ex-breakerdir': 'lockout',
    'ex-bucketpolicy': 'privacy', 'ex-bucketsettings': 'privacy', 'ex-bystanderflow': 'harassment', 'ex-camcover': 'badge',
    'ex-canaryconfig': 'security', 'ex-changeflow': 'serverroom', 'ex-changelog': 'serverroom', 'ex-changewindow': 'serverroom',
    'ex-chatexport': 'privacy', 'ex-chatharass': 'harassment', 'ex-ciyaml': 'security', 'ex-cloudspend': 'serverroom',
    'ex-co2': 'wellness', 'ex-cobolsrc': 'security', 'ex-coiflow': 'ethics', 'ex-coldhold': 'foodsafety',
    'ex-concernflow': 'whistleblower', 'ex-confbadges': 'meetings', 'ex-contractorpct': 'wage', 'ex-cookiebanner': 'privacy',
    'ex-cookprobe': 'celebrations', 'ex-coolingcurve': 'potluck', 'ex-corehours': 'calendar', 'ex-counterht': 'accessibility',
    'ex-coverage': 'pto', 'ex-cranesignals': 'construction', 'ex-crawlerhits': 'serverroom', 'ex-crontab': 'security',
    'ex-cylinderlabel': 'hazcom', 'ex-cylinderstore': 'construction', 'ex-datasetcard': 'ai', 'ex-dbcpu': 'serverroom',
    'ex-dinnertables': 'celebrations', 'ex-dlpsettings': 'privacy', 'ex-dmthread': 'harassment', 'ex-dockerfile': 'security',
    'ex-doorwidth': 'accessibility', 'ex-dosimeter': 'ppe', 'ex-dotenv': 'security', 'ex-doublebooked': 'meetings',
    'ex-drenchwalk': 'hazcom', 'ex-dsotrend': 'economic', 'ex-dumpsterspot': 'esg', 'ex-energyuse': 'esg',
    'ex-engagement': 'conduct', 'ex-escortlog': 'badge', 'ex-exitwalk': 'accessibility', 'ex-exportflow': 'export',
    'ex-extwalk': 'lobby', 'ex-eyewashht': 'hazcom', 'ex-fallclearance': 'construction', 'ex-fallht': 'construction',
    'ex-fatigue': 'driving', 'ex-filterdp': 'wellness', 'ex-fireriser': 'construction', 'ex-firewatch': 'construction',
    'ex-firstaidwalk': 'firstaid', 'ex-flsaflow': 'wage', 'ex-fmlaflow': 'parental', 'ex-focusblock': 'calendar',
    'ex-forkliftplate': 'warehouse', 'ex-fortranyr': 'retro', 'ex-freezer': 'potluck', 'ex-genspot': 'weather',
    'ex-gfsbackup': 'serverroom', 'ex-ghs': 'hazcom', 'ex-giftflow': 'gifts', 'ex-gitconfig': 'security',
    'ex-gitdiff': 'security', 'ex-grabbar': 'accessibility', 'ex-grillspot': 'celebrations', 'ex-guardrailht': 'construction',
    'ex-handrailht': 'accessibility', 'ex-harassflow': 'harassment', 'ex-headroom': 'safety', 'ex-hoslog': 'driving',
    'ex-hothold': 'potluck', 'ex-hotworkflow': 'construction', 'ex-hydrantpark': 'driving', 'ex-i9flow': 'hiring',
    'ex-jcl': 'security', 'ex-jndilog': 'security', 'ex-ladderduty': 'construction', 'ex-ladderext': 'construction',
    'ex-laptopsettings': 'security', 'ex-licensecount': 'procurement', 'ex-liftposture': 'warehouse', 'ex-lithiumflow': 'cabin',
    'ex-lockhandoff': 'lockout', 'ex-loginlog': 'security', 'ex-loneworker': 'fieldwork', 'ex-lotoflow': 'lockout',
    'ex-mailboxsettings': 'email', 'ex-marshalling': 'aviation', 'ex-mdmsettings': 'security', 'ex-mediaflow': 'meetings',
    'ex-meetingsettings': 'meetings', 'ex-meetingslot': 'calendar', 'ex-mfaenroll': 'security', 'ex-mfasettings': 'security',
    'ex-microaggress': 'conduct', 'ex-minorhours': 'hiring', 'ex-minutes': 'governance', 'ex-nfpa704': 'hazcom',
    'ex-noisedose': 'ppe', 'ex-npstrend': 'meetings', 'ex-offeraccept': 'hiring', 'ex-officetemp': 'wellness',
    'ex-oncall': 'nightshift', 'ex-ossflow': 'records', 'ex-overtimetrend': 'wage', 'ex-parcelscale': 'deliveries',
    'ex-paspeaker': 'weather', 'ex-patentfig': 'innovation', 'ex-paygaptrend': 'pay', 'ex-payratiotrend': 'pay',
    'ex-payrollsettings': 'ethics', 'ex-perchregs': 'meetings', 'ex-permitflow': 'construction', 'ex-phishclick': 'phishing',
    'ex-phoneperms': 'security', 'ex-pipemarkers': 'lobby', 'ex-placards': 'hazcom', 'ex-plankoverhang': 'construction',
    'ex-posttest': 'driving', 'ex-pressdraft': 'insider', 'ex-propanefill': 'warehouse', 'ex-protrude': 'safety',
    'ex-pullwalk': 'accessibility', 'ex-punchcard': 'retro', 'ex-rackload': 'warehouse', 'ex-radarspeed': 'parking',
    'ex-radiocover': 'deliveries', 'ex-railwall': 'accessibility', 'ex-rampwidth': 'accessibility', 'ex-reachrange': 'accessibility',
    'ex-recordflow': 'firstaid', 'ex-recordingsettings': 'privacy', 'ex-reposettings': 'security', 'ex-requirements': 'security',
    'ex-resincodes': 'esg', 'ex-retainflow': 'records', 'ex-retaliationmsg': 'whistleblower', 'ex-reviewcomments': 'hiring',
    'ex-robots': 'serverroom', 'ex-safetycolors': 'safety', 'ex-sanitizer': 'hazcom', 'ex-scaffoldht': 'construction',
    'ex-serverinlet': 'serverroom', 'ex-sevflow': 'security', 'ex-shellhistory': 'security', 'ex-shelterwalk': 'weather',
    'ex-shrink': 'warehouse', 'ex-signalwords': 'safety', 'ex-sitetraffic': 'serverroom', 'ex-slabreach': 'records',
    'ex-smartspeaker': 'security', 'ex-smokedet': 'lobby', 'ex-smokezone': 'wellness', 'ex-socialflow': 'social',
    'ex-spamvolume': 'email', 'ex-spillflow': 'hazcom', 'ex-spillwalk': 'hazcom', 'ex-spoilpile': 'construction',
    'ex-storagecompat': 'hazcom', 'ex-sudoers': 'security', 'ex-tepid': 'hazcom', 'ex-thermostat': 'esg',
    'ex-ticketbacklog': 'security', 'ex-ticketnotes': 'security', 'ex-timeclock': 'wage', 'ex-timetofill': 'hiring',
    'ex-tiredate': 'driving', 'ex-tirepressure': 'driving', 'ex-tiretread': 'driving', 'ex-toiletseat': 'accessibility',
    'ex-tphc': 'potluck', 'ex-trademarkcopy': 'records', 'ex-trainingcomp': 'talent', 'ex-treaddepth': 'construction',
    'ex-trenchprot': 'construction', 'ex-trenchslope': 'construction', 'ex-trir': 'firstaid', 'ex-turnover': 'talent',
    'ex-upsruntime': 'serverroom', 'ex-urlparams': 'security', 'ex-useragents': 'serverroom', 'ex-valvepos': 'lockout',
    'ex-vendorlist': 'procurement', 'ex-voicemail': 'security', 'ex-wateruse': 'esg', 'ex-webconfig': 'security',
    'ex-wifiap': 'security', 'ex-wireflow': 'economic', 'ex-wpconfig': 'security',
})

LIGHT = ['soft morning light', 'warm golden-hour light', 'cool overcast daylight', 'bright even midday light',
         'late evening light with long shadows', 'soft diffused window light']
DECOR = ['mid-century modern decor', 'minimalist Scandinavian decor', 'classic wood-paneled decor',
         'industrial loft decor with exposed brick', 'sleek glass-and-steel decor', '1990s corporate decor with beige carpet',
         'warm Midwestern office decor', 'contemporary decor with navy accents']
CAMERA = ['wide-angle architectural view', 'eye-level view', 'three-quarter view', 'symmetrical straight-on view',
          'slightly elevated view', 'close still-life composition']

STYLE = ('. Photorealistic corporate photography, empty and unoccupied, unbranded blank surfaces, '
         'muted navy and warm gold palette, sharp focus, natural colors')

# A scene that comes round again (a pool category has ~15 scenes and 118
# images) gets one of these framings instead of a plain camera angle, never
# the same one twice for that scene. They change framing, light and weather
# only: every cover stays a photograph in the house palette (other media —
# watercolor, isometric renders, card models, studio backdrops — came out
# looking like a different site, and strip the room away from the subject).
TREATMENTS = [
    'close-up detail of the main object, shallow depth of field, soft background blur',
    'seen through a glass partition with soft reflections on the glass',
    'seen through an open doorway from a dim corridor',
    'at night, lit only by warm desk lamps, city lights glowing through the windows',
    'on a rainy afternoon, raindrops streaking the windows, soft grey light',
    'on a snowy winter morning, snow visible through the windows',
    'with shafts of sunlight and long geometric shadows across the floor',
    'from a low angle near the floor, strong leading lines',
    'wide establishing shot with generous empty space',
    'at blue hour, interior lights glowing warmly',
    'in warm late-afternoon light, dust motes in the sunbeams',
    'on a bright clear morning, crisp shadows',
]

def load_topics():
    topics = []
    for line in (DATA / 'course_types.txt').read_text().splitlines():
        if not line.strip() or line.startswith('#'):
            continue
        parts = [p.strip() for p in line.split('|')]
        slug, title, mix, blurb = parts[0], parts[1], parts[5], parts[6]
        cats = []
        for m in mix.split(','):
            cat, _, w = m.partition(':')
            cat = ALIASES.get(cat.strip(), cat.strip())
            cats += [cat] * int(w or 1)
        topics.append({'slug': slug, 'title': title, 'cats': cats, 'blurb': blurb})
    return topics


def places():
    out = []
    for line in (DATA / 'vars' / 'PLACE.txt').read_text().splitlines():
        line = line.strip()
        if line and not line.startswith('#') and '|' not in line:
            out.append(line)
    return out


def build_prompt(topic, n, place_list):
    """Deterministic prompt for image n (1-based) of a topic. Scenes come
    from the topic's question categories, heaviest first, each category's
    scenes shuffled per topic, so a topic's images are distinct scenes; decor
    style, light, and camera angle vary on top of that."""
    pool = scene_pool(topic, place_list)
    return scene_prompt(topic, n, pool[(n - 1) % len(pool)], (n - 1) // len(pool))


def scene_pool(topic, place_list):
    """The scenes build_prompt cycles through for a topic, in order."""
    order = random.Random(f"compliance_cover_order_{topic['slug']}")
    weights = {}
    for c in topic['cats']:
        weights[c] = weights.get(c, 0) + 1
    pool, seen = [], set()
    for c in sorted(weights, key=lambda c: (-weights[c], c)):
        scenes = list(SCENES.get(c, []))
        order.shuffle(scenes)
        for sc in scenes:
            if sc not in seen:
                seen.add(sc)
                pool.append(sc)
    if len(pool) < 8:  # pad thin topics so 6+ images never repeat a scene
        extra = GENERIC + [f'an empty {p} in a modern corporate headquarters' for p in place_list]
        order.shuffle(extra)
        pool += extra
    return pool


def scene_prompt(topic, n, scene, again):
    """Prompt for image n showing a scene; again counts earlier rounds of
    the topic's scenes (0 = first time round)."""
    rng = random.Random(f"compliance_cover_{topic['slug']}_{n}")
    outdoor = any(w in scene for w in ('sky', 'skyline', 'outside', 'port ', 'aerial', 'entrance', 'campus',
                                         'avenue', 'ridge', 'clouds', 'apron', 'snow', 'district'))
    decor = '' if outdoor else f', {rng.choice(DECOR)}'
    if not again:
        return f'{scene}{decor}, {rng.choice(LIGHT)}, {rng.choice(CAMERA)}{STYLE}'
    # Each scene's repeats walk its own shuffled treatment and palette lists,
    # so its 2nd, 3rd… images never share a look. The scene's own light
    # ("…, morning light") is dropped; the treatment sets the light.
    per = random.Random(f"compliance_cover_treat_{topic['slug']}_{scene}")
    treats = TREATMENTS[:]
    per.shuffle(treats)
    framing = treats[(again - 1) % len(treats)]
    scene = re.sub(r',[^,]*\blight$', '', scene)
    return f'{scene}{decor}, {framing}{STYLE}'


def report_jobs(force):
    """(slug, prompt) for every existing report cover not yet redone with
    FLUX (manifest entries marked model=flux-schnell are skipped, so a killed
    run resumes). Each report's scene comes from its category, picked by its
    slug, so reports in the same category get different scenes."""
    manifest = json.loads((REPORT_DIR / 'manifest.json').read_text())
    by_cat = {}
    for slug in sorted(manifest):
        by_cat.setdefault(manifest[slug].get('category', ''), []).append(slug)
    jobs = []
    for cat, slugs in sorted(by_cat.items()):
        topic = {'slug': 'report_' + cat, 'cats': REPORT_CATEGORIES.get(cat, ['conduct'])}
        for i, slug in enumerate(slugs, 1):
            if not force and manifest[slug].get('model') == 'flux-schnell':
                continue
            jobs.append((slug, build_prompt(topic, i, places())))
    return jobs


def category_slug(category):
    # Must match apps/honeypot/views.py _cover_category_slug.
    return category.lower().replace(' & ', '-and-').replace(' ', '-')


# Wider scene sets for the report pool, so ~120 images per category don't
# lean on three or four scenes.
POOL_CATEGORIES = {
    'Compensation': ['pay', 'governance', 'insider', 'records', 'economic'],
    'Workforce': ['hotdesk', 'hybrid', 'rto', 'conduct', 'meetings', 'remote'],
    'Benefits': ['wellness', 'parental', 'pto', 'accessibility', 'ergonomics', 'leave'],
    'Governance': ['governance', 'conduct', 'records', 'leadership', 'ethics'],
    'ESG': ['esg', 'weather', 'economic'],
    'Retention': ['wellness', 'pto', 'meetings', 'leadership', 'remote', 'hybrid'],
    'Technology': ['ai', 'security', 'privacy', 'continuity'],
    'Economic': ['economic', 'export', 'insider', 'travel'],
    'Diversity & Inclusion': ['accessibility', 'conduct', 'meetings', 'harassment', 'parental'],
    'Compliance': ['records', 'security', 'privacy', 'cleandesk', 'ethics', 'investigations'],
    'Leadership': ['leadership', 'governance', 'meetings', 'calendar'],
    'Talent Acquisition': ['talent', 'visitor', 'meetings', 'rto', 'badge'],
}


def pool_topic(cat):
    sc = POOL_CATEGORIES[cat]
    return {'slug': 'pool_' + cat, 'cats': [c for c in sc for _ in range(len(sc))]}


def pool_scenes(cat):
    """A category's first-pass scenes: its scene pool minus scenes an
    earlier category already uses, so no scene is in two pools."""
    claimed = set()
    for other in POOL_CATEGORIES:
        if other == cat:
            break
        claimed.update(scene_pool(pool_topic(other), places()))
    return [sc for sc in scene_pool(pool_topic(cat), places()) if sc not in claimed]


def pool_prompt(cat, topic, n):
    """Prompt for pool image n of a report category. The first pass is the
    category's whole scenes (build_prompt, house look); after that each image
    is one of the category's own subjects in one of its settings. A subject
    comes back only every len(subjects) images, each time in a new setting
    and with a new look (a treatment, or the house photo style)."""
    scenes = pool_scenes(cat)
    subjects, settings = POOL_SUBJECTS.get(cat), POOL_SETTINGS.get(cat)
    if scenes and (n <= len(scenes) or not subjects or not settings):
        return scene_prompt(topic, n, scenes[(n - 1) % len(scenes)], (n - 1) // len(scenes))
    k = n - len(scenes) - 1
    order = random.Random(f'cover_pool_subjects_{cat}')
    subjects = subjects[:]
    order.shuffle(subjects)
    subject, again = subjects[k % len(subjects)], k // len(subjects)
    per = random.Random(f'cover_pool_subject_{cat}_{subject}')
    looks = [None, None] + TREATMENTS  # None = a plain camera angle
    per.shuffle(looks)
    setting = settings[(per.randrange(len(settings)) + again) % len(settings)]
    look = looks[again % len(looks)]
    rng = random.Random(f'cover_pool_{cat}_{n}')
    if look is None:
        return f'{subject}, {setting}, {rng.choice(LIGHT)}, {rng.choice(CAMERA)}{STYLE}'
    return f'{subject}, {setting}, {look}{STYLE}'


def pool_jobs(per_category, force):
    """(name, prompt) for report-covers/pool/<category slug>-<n>.webp, the
    covers shown on generated reports that have no cover of their own."""
    out = REPORT_DIR / 'pool'
    mpath = out / 'manifest.json'
    manifest = json.loads(mpath.read_text()) if mpath.exists() else {}
    jobs = []
    for cat, scene_cats in POOL_CATEGORIES.items():
        topic = pool_topic(cat)
        for n in range(1, per_category + 1):
            name = f'{category_slug(cat)}-{n}'
            prompt = pool_prompt(cat, topic, n)
            # Redo an image whose prompt has since changed (the repeats that
            # used other media), so a rerun picks up prompt edits.
            stale = manifest.get(name, {}).get('prompt', prompt) != prompt
            if force or stale or not (out / f'{name}.webp').exists():
                jobs.append((name, prompt))
    return jobs


def seed_for(name):
    return int(hashlib.md5(f'compliance_cover_{name}'.encode()).hexdigest(), 16) % (2 ** 32)


def write_go_list():
    names = sorted(p.stem for p in OUT_DIR.glob('*.webp') if '-' in p.stem)
    GO_LIST.write_text(
        '# Generated by tools/generate_compliance_covers.py: one cover per line,\n'
        '# <course-type slug>-<n>, served from /static/img/compliance-covers/.\n'
        + ''.join(f'{n}\n' for n in names))
    print(f'Wrote {len(names)} names to {GO_LIST.relative_to(ROOT)}')


def write_atomic_text(path, text):
    fd, tmp = tempfile.mkstemp(prefix='.', suffix='.part', dir=path.parent)
    with os.fdopen(fd, 'w') as f:
        f.write(text)
    os.replace(tmp, path)


def save_atomic(img, path):
    fd, tmp = tempfile.mkstemp(prefix='.', suffix='.part', dir=path.parent)  # never matches *.webp
    try:
        os.close(fd)
        img.save(tmp, 'WEBP', quality=80, method=6)
        os.replace(tmp, path)
    except Exception:
        try:
            os.unlink(tmp)
        except OSError:
            pass
        raise


def main():
    ap = argparse.ArgumentParser(description=__doc__.split('\n\n')[0])
    ap.add_argument('--per-topic', type=int, default=6)
    ap.add_argument('--topics', default='', help='comma-separated course-type slugs (default: all)')
    ap.add_argument('--limit', type=int, default=0, help='stop after generating this many images')
    ap.add_argument('--model', default='black-forest-labs/FLUX.1-schnell')
    ap.add_argument('--steps', type=int, default=4)
    ap.add_argument('--hf-token', default=os.environ.get('HF_TOKEN'))
    ap.add_argument('--force', action='store_true', help='regenerate existing images')
    ap.add_argument('--dry-run', action='store_true', help='print prompts only')
    ap.add_argument('--list-only', action='store_true', help='only rewrite the acpwb_go file list')
    ap.add_argument('--report-pool', action='store_true',
                    help='generate the report cover pool (static/img/report-covers/pool/) for reports without a cover')
    ap.add_argument('--pool-per-category', type=int, default=118, help='pool images per report category (12 categories)')
    ap.add_argument('--report-covers', action='store_true',
                    help='replace the existing report covers (static/img/report-covers/<slug>.webp) in place instead')
    args = ap.parse_args()

    OUT_DIR.mkdir(parents=True, exist_ok=True)
    if args.list_only:
        write_go_list()
        return

    out_dir, size = OUT_DIR, (WIDTH, HEIGHT)
    if args.report_pool:
        out_dir = REPORT_DIR / 'pool'
        out_dir.mkdir(parents=True, exist_ok=True)
        jobs = pool_jobs(args.pool_per_category, args.force)
        if args.limit:
            jobs = jobs[:args.limit]
        print(f'{len(jobs)} report pool image(s) to generate.')
        return run(args, jobs, out_dir, REPORT_SIZE, report=True)
    if args.report_covers:
        out_dir, size = REPORT_DIR, REPORT_SIZE
        jobs = report_jobs(args.force)
        if args.limit:
            jobs = jobs[:args.limit]
        print(f'{len(jobs)} report cover(s) to replace.')
        return run(args, jobs, out_dir, size, report=True)
    topics = load_topics()
    if args.topics:
        want = set(args.topics.split(','))
        topics = [t for t in topics if t['slug'] in want]
    place_list = places()
    jobs = []
    for t in topics:
        for n in range(1, args.per_topic + 1):
            name = f"{t['slug']}-{n}"
            if args.force or not (OUT_DIR / f'{name}.webp').exists():
                jobs.append((name, build_prompt(t, n, place_list)))
    if args.limit:
        jobs = jobs[:args.limit]
    print(f'{len(jobs)} image(s) to generate across {len(topics)} topic(s).')

    return run(args, jobs, out_dir, size)


def run(args, jobs, out_dir, size, report=False):
    if args.dry_run:
        for name, prompt in jobs:
            print(f'{name}\n    {prompt}\n')
        return
    if not jobs:
        if not report:
            write_go_list()
        return

    try:
        import torch
        from diffusers import FluxPipeline
        from diffusers.utils import logging as diffusers_logging
        from PIL import Image
    except ImportError as e:
        sys.exit(f'Missing dependency: {e}\n  pip install torch diffusers transformers accelerate pillow sentencepiece protobuf')

    if torch.cuda.is_available():
        device, dtype = 'cuda', torch.bfloat16
    elif torch.backends.mps.is_available():
        device, dtype = 'mps', torch.bfloat16   # float32 FLUX is ~68GB and swaps on a 64GB Mac
    else:
        device, dtype = 'cpu', torch.float32
    print(f'Loading {args.model} on {device} ({dtype}) …')
    diffusers_logging.set_verbosity_error()
    pipe = FluxPipeline.from_pretrained(args.model, torch_dtype=dtype, token=args.hf_token or None).to(device)
    pipe.set_progress_bar_config(disable=True)

    manifest_path = out_dir / 'manifest.json'
    manifest = json.loads(manifest_path.read_text()) if manifest_path.exists() else {}
    gen_w, gen_h = (1408, 768) if report else (GEN_WIDTH, GEN_HEIGHT)  # multiples of 16, same aspect
    errors = 0
    for i, (name, prompt) in enumerate(jobs, 1):
        print(f'[{i}/{len(jobs)}] {name} ', end='', flush=True)
        try:
            img = pipe(
                prompt=prompt, num_inference_steps=args.steps, guidance_scale=0.0,
                width=gen_w, height=gen_h, max_sequence_length=256,
                generator=torch.Generator(device=device).manual_seed(seed_for(('report_' if report else '') + name)),
            ).images[0]
            img = img.resize(size, Image.LANCZOS)
            save_atomic(img, out_dir / f'{name}.webp')
            # Keep existing fields (a report's category and path); the model
            # marker is what lets a killed --report-covers run resume.
            manifest[name] = {**manifest.get(name, {}), 'prompt': prompt, 'model': 'flux-schnell'}
            write_atomic_text(manifest_path, json.dumps(manifest, indent=1, sort_keys=True))
            print('✓')
        except KeyboardInterrupt:
            # Exit non-zero (130, as a shell would) so generate_all_covers.sh
            # stops instead of treating Ctrl-C as a finished step and moving on.
            print('interrupted')
            if not report:
                write_go_list()
            sys.exit(130)
        except Exception as exc:
            print(f'FAILED: {exc}')
            errors += 1
    if not report:
        write_go_list()
    print(f'Done: {len(jobs) - errors} generated, {errors} failed → {out_dir.relative_to(ROOT)}')


if __name__ == '__main__':
    main()
