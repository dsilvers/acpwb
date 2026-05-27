PROBLEM_TEMPLATES = [
    "The current {area} process at ACPWB suffers from significant inefficiencies that have accumulated over {years} years of incremental change without systematic review. Manual handoffs between departments create delays of up to {days} business days, and the absence of standardized documentation results in inconsistent outcomes that are difficult to audit or improve.",
    "ACPWB's {area} function has experienced a {pct}% increase in volume over the past {years} years without a corresponding increase in staffing or process sophistication. As a result, error rates have climbed to {error_pct}%, approval cycle times have extended to an average of {days} business days, and stakeholder satisfaction scores have declined steadily.",
    "A review of the {area} process conducted by the Business Process Management office identified {count} discrete failure points that collectively account for {pct}% of all process exceptions. These failures impose measurable costs on the organization, estimated at ${cost}K annually in rework, delays, and compliance risk.",
    "The existing {area} process was designed for an organization of {old_size} employees and has not been fundamentally redesigned since {old_year}. Growth to the current staffing level and the introduction of {count} new systems over the intervening period have created a fragmented, redundant workflow that consumes {hours} FTE-hours per week in non-value-added activities.",
    "Stakeholder interviews conducted in support of this initiative revealed widespread dissatisfaction with the current {area} process. {pct}% of respondents rated the process as 'inefficient' or 'very inefficient,' and {count} distinct pain points were identified across {dept_count} participating departments.",
]

SOLUTION_TEMPLATES = [
    "The proposed solution redesigns the {area} process end-to-end using {methodology} principles, eliminating {count} non-value-added steps and consolidating {old_systems} legacy systems into a single integrated workflow. The redesigned process is projected to reduce cycle time by {pct}% and error rates by {error_pct}% within {months} months of full deployment.",
    "ACPWB will implement a {methodology}-driven {area} framework that introduces standardized templates, automated routing, and real-time exception alerting. This approach addresses the root causes identified in the diagnostic phase and provides a scalable foundation that can accommodate {pct}% volume growth without additional headcount.",
    "The initiative will deploy a phased {methodology} approach, beginning with a {dept_count}-department pilot and expanding to the full organization over {months} months. Technology enhancements will be limited to configuration changes within existing licensed platforms, minimizing implementation risk and capital expenditure.",
    "Drawing on benchmarking data from {count} peer organizations, including {company}, the proposed {area} process redesign adopts industry-leading practices in automation, exception management, and continuous improvement. Expected benefits include a {pct}% reduction in processing costs and a {months}-month payback period on implementation investment.",
    "The redesigned {area} process will be governed by a newly established Center of Excellence, staffed with {count} dedicated process analysts and supported by executive sponsorship from the Chief Operating Officer. This governance model ensures sustained improvement beyond the initial implementation period.",
]

ROOT_CAUSE_TEMPLATES = [
    "Lack of standardized procedures has led to inconsistent execution across departments and regions",
    "Manual data entry at multiple touchpoints introduces errors that propagate through downstream processes",
    "Insufficient training has resulted in {pct}% of process participants being unable to correctly handle exception scenarios",
    "Legacy system limitations prevent end-to-end visibility and require redundant data entry across {count} separate platforms",
    "Unclear ownership at process handoff points results in delays averaging {days} business days per transaction",
    "Approval authority matrices have not been updated since {old_year}, creating unnecessary escalations for routine transactions",
    "Absence of defined SLAs allows transactions to remain in queue indefinitely without triggering escalation",
    "Siloed departmental metrics create local optimization at the expense of end-to-end process performance",
    "Workarounds developed to address system limitations have been institutionalized as informal procedures, bypassing controls",
    "Inadequate change management during prior system implementations resulted in low user adoption and persistent manual workarounds",
    "Reporting deficiencies prevent timely identification of process bottlenecks and performance degradation",
    "Vendor and contractor interfaces are not standardized, requiring custom handling for each supplier relationship",
]

RISK_TEMPLATES = [
    ("Change resistance from process participants accustomed to current workflows", "Medium"),
    ("Integration delays with legacy systems extending implementation timeline", "Medium"),
    ("Temporary productivity dip during transition to new process", "Low"),
    ("Data migration errors affecting historical records and reporting", "Medium"),
    ("Key personnel departures during implementation reducing institutional knowledge", "High"),
    ("Budget overruns due to scope creep or unforeseen technical complexity", "Medium"),
    ("Regulatory compliance gaps during transition period", "High"),
    ("Vendor support delays affecting technology component delivery", "Low"),
    ("Insufficient pilot scope resulting in undiscovered edge cases at full deployment", "Medium"),
    ("Executive sponsorship changes affecting initiative priority", "High"),
]

SUMMARY_TEMPLATES = [
    "Applying {methodology} principles to redesign {area} workflows and eliminate {count} non-value-added process steps.",
    "Systematic redesign of {area} processes to reduce cycle time by {pct}% and improve stakeholder satisfaction scores.",
    "End-to-end process transformation for {area} leveraging {methodology} methodology and automation to cut costs by ${cost}K annually.",
    "Structured {methodology} initiative to standardize {area} across all ACPWB business units and reduce error rates by {pct}%.",
    "Process redesign initiative targeting {area} inefficiencies identified through stakeholder interviews and value stream mapping.",
]
