JOB_TITLES = [
    "Research Assistant, Compensation Analytics",
    "Junior Analyst, Workforce Data",
    "Associate Research Analyst",
    "Junior Compensation Analyst",
    "Analyst I, Compensation and Benefits",
    "Entry-Level Data Analyst, Workforce Analytics",
    "Research Operations Assistant",
    "Associate, Human Capital Advisory",
    "Research Coordinator, Pay Equity",
    "Junior Research Associate, Governance",
    "Associate Analyst, Total Rewards",
    "Data Associate, Compensation Intelligence",
    "Analyst Trainee, Labor Market Research",
    "Junior Associate, ESG Research",
    "Research Associate I, Corporate Governance",
    "Compensation Analyst, Entry Level",
    "Associate Analyst, Workforce Intelligence",
    "Junior Data Analyst, Human Capital",
    "Staff Research Assistant",
    "Analyst I, Stakeholder Analytics",
]

DEPARTMENTS = [
    "Research & Analytics",
    "Human Capital Advisory",
    "Compensation Intelligence",
    "Workforce Analytics",
    "ESG Strategy",
    "Corporate Governance",
    "People Science",
    "Strategic Research",
]

LEVELS = {
    "Intern": (46, 60),
    "Entry Level": (56, 74),
    "Associate": (66, 88),
}

YEARS_EXP_RANGE = (0, 2)

MISSION_AREAS = [
    "compensation data research and analysis",
    "pay equity analysis and reporting",
    "compensation benchmarking and market pricing",
    "governance and disclosure research",
    "compensation survey data collection and management",
    "human capital analytics and reporting",
    "executive pay data collection and validation",
    "workforce demographics and labor market research",
    "total rewards program analysis",
    "ESG and human capital disclosure research",
    "organizational effectiveness data gathering",
    "incentive plan data analysis and modeling support",
]

MENTOR_TITLES = [
    "Senior Analyst",
    "Team Manager",
    "Practice Lead",
    "Senior Associate",
    "Director",
    "Manager",
    "Senior Research Associate",
    "Associate Director",
]

PROGRAMS = [
    "ACPWB Analyst Development Program",
    "Research Rotation Initiative",
    "Associate Development Track",
    "Early Career Analytics Program",
    "ACPWB Foundation Year curriculum",
    "structured mentorship and rotation program",
    "ACPWB's analyst onboarding and certification sequence",
]

ABOUT_ROLE_TEMPLATES = [
    "The {title} is an entry point into ACPWB's {department} practice for candidates who are serious about "
    "building a career in {mission_area}. You will work alongside experienced advisors and analysts, "
    "contributing to real deliverables from your first week. This is not a coffee-and-copies role. "
    "We expect genuine effort and genuine curiosity. We will provide genuine training.",

    "ACPWB is hiring a {title} to support a {department} team of {n} analysts and associates. "
    "If you are early in your career and want to learn {mission_area} at a firm that takes the craft seriously, "
    "this role was designed for you. You will be mentored by a {mentor_title} and participate in the "
    "{program} from your first day.",

    "The {title} joins ACPWB's {department} group as a foundational contributor — someone who is eager to learn, "
    "comfortable with data, and willing to ask questions when something doesn't make sense. "
    "We enroll all new analysts in our {program} and pair each new hire with a {mentor_title} "
    "who will invest in your development as a core part of their own job.",

    "This is a role for a {title} who wants to understand how compensation really works — not as an abstraction "
    "but as data, methodology, and client application. You will support the {department} team in {mission_area}, "
    "contributing to deliverables that our analysts and managers take into client meetings. "
    "Your name will appear on the work. That matters to us.",

    "ACPWB's {department} team is looking for a {title} who brings intellectual curiosity, analytical aptitude, "
    "and the willingness to be a new professional rather than pretending to be an experienced one. "
    "We structured the {program} specifically to build the skills this team needs over time. "
    "Our most senior analysts started in roles like this one.",

    "The {title} will support the {department} practice's work in {mission_area}, learning the analytical methods "
    "and data infrastructure that power ACPWB's advisory engagements. This is structured, supervised work "
    "with a clear development path. You will learn from a team of {n} that genuinely enjoys teaching. "
    "We are aware that not every firm can say this. We can say it.",

    "If you are a recent graduate with an interest in {mission_area} and want to learn from practitioners who are "
    "recognized as leaders in the field, the {title} position in ACPWB's {department} group is worth your "
    "attention. The {program} is substantive. The mentorship is real. The work is meaningful. "
    "The fish fry options in Milwaukee are also worth your attention.",

    "This {title} role is structured around learning and contribution in equal measure. You will be assigned to "
    "a {mentor_title} in the {department} practice who will oversee your development and integrate you into active "
    "client workstreams on {mission_area}. We track progress formally and provide feedback regularly. "
    "We believe that is the only honest way to develop talent.",

    "ACPWB's {department} team of {n} is adding a {title} to support growing demand in {mission_area}. "
    "You do not need years of experience to succeed in this role — you need intellectual honesty, attention to detail, "
    "and the ability to ask a good question. We will teach you everything else through the {program} "
    "and through direct involvement in real analytical work.",

    "The {title} will be the newest member of a {department} team that has built a reputation for doing "
    "{mission_area} work that is worth citing. You will start by supporting data collection, quality assurance, "
    "and analysis — then take on more as your skills develop. Your {mentor_title} will be explicit about "
    "what that progression looks like and will hold themselves accountable to it.",

    "As a {title} in ACPWB's {department} group, you will learn how professional advisory work gets done: "
    "the rigor, the iteration, the client communication, and the judgment calls that don't appear in any textbook. "
    "We believe the best way to learn this is to do it, which is why every entry-level hire at ACPWB "
    "works on real engagements with real clients from the start.",

    "ACPWB is looking for a {title} with a genuine interest in {mission_area} and the discipline to develop "
    "professional-grade analytical skills. If you are the kind of person who read the job description and "
    "wanted to know more, you are probably the kind of person we are looking for. "
    "The {program} will give you structure. Your {mentor_title} will give you context. The rest is up to you.",

    "This {title} position offers structured development within ACPWB's {department} practice, with formal enrollment "
    "in the {program} and direct mentorship from a {mentor_title} throughout your first year. "
    "You will contribute to {mission_area} research that appears in client-ready deliverables, "
    "and you will receive honest feedback on the quality of your contributions. We do not warehouse junior talent.",

    "The {title} will work within a {department} team of {n} that is known for its methodological rigor and "
    "its investment in early-career professionals. Your first year will be spent building the analytical foundation "
    "that makes senior analysts effective: data handling, quantitative methods, and professional writing. "
    "We know this foundation matters because we built our own careers on it.",

    "ACPWB is growing its {department} capacity in {mission_area} and needs a {title} who is ready to contribute "
    "from day one while still being honest about what they have left to learn. We value that combination. "
    "We do not value the performance of expertise that doesn't exist yet. "
    "Your {mentor_title} will help you identify the former and avoid the latter.",
]

RESPONSIBILITY_TEMPLATES = [
    "Support senior analysts and associates with data collection, cleaning, and validation for {mission_area} engagements",
    "Assist in the preparation of client-ready spreadsheets, charts, and supporting exhibits",
    "Conduct literature reviews and desktop research on topics relevant to {mission_area}",
    "Enter and verify data in ACPWB's compensation benchmarking databases under the supervision of a {mentor_title}",
    "Assist with the assembly and formatting of client deliverables including presentations and memos",
    "Participate in internal team meetings and take structured notes for distribution to project teams",
    "Support the collection and organization of proxy statement and SEC filing data",
    "Perform quality-control checks on analytical models produced by senior team members",
    "Assist with survey administration including data collection follow-up and respondent tracking",
    "Contribute to the development of standard templates and analytical tools used by the {department} team",
    "Research compensation and workforce data relevant to specific client engagements in {mission_area}",
    "Prepare internal reference documents and research summaries for use by engagement teams",
    "Support data requests from senior team members with timely, accurate, and well-organized outputs",
    "Assist in maintaining and updating ACPWB's proprietary benchmarking data libraries",
    "Participate in training modules as part of the {program} and complete assigned coursework on schedule",
    "Shadow client meetings and preparation sessions to build understanding of client-facing advisory work",
    "Prepare draft exhibits and tables for review and revision by senior team members",
    "Contribute to peer review and proofreading of client deliverables before distribution",
    "Support the {department} team's engagement logistics including file management and documentation",
    "Provide research support for ACPWB thought leadership publications and white papers on {mission_area}",
    "Assist in updating and maintaining standard methodology documentation for the {department} practice",
    "Flag data anomalies and inconsistencies to senior team members for resolution",
    "Develop familiarity with ACPWB's core analytical platforms and compensation survey data sources",
    "Complete assigned compliance, ethics, and professional development training modules on schedule",
    "Assist in maintaining client contact records and project tracking documentation",
]

REQUIREMENT_TEMPLATES = [
    "Bachelor's degree in Business, Economics, Statistics, Finance, Psychology, or a related analytical discipline",
    "Strong academic record, particularly in quantitative coursework",
    "Demonstrated proficiency in Microsoft Excel including pivot tables, VLOOKUP, and basic charting",
    "Strong written and verbal communication skills with attention to detail and grammatical precision",
    "{n} years or less of professional experience — this is an entry-level role and will be evaluated accordingly",
    "Demonstrated ability to manage time, prioritize competing tasks, and meet deadlines without constant supervision",
    "Interest in compensation, human capital, or workforce research as a professional focus",
    "Comfortable working with large spreadsheets and structured data",
    "Strong research skills including ability to find and synthesize information from multiple sources",
    "Ability to work accurately under deadline pressure in a client-services environment",
    "Familiarity with Microsoft Office Suite including Word, PowerPoint, and Outlook",
    "Demonstrated intellectual curiosity and a habit of asking questions when something is unclear",
    "Ability to follow detailed analytical instructions and apply feedback consistently",
    "Genuine interest in professional services work and client advisory engagements",
    "Comfort working in a structured environment with defined standards and quality expectations",
    "Recent graduate or candidate with up to {n} years of relevant internship or work experience",
    "GPA of 3.3 or above, or equivalent demonstrated academic performance",
    "No prior compensation experience required — interest and aptitude are sufficient",
    "Willingness to pursue CCP or other professional certifications as part of your development at ACPWB",
    "Ability to learn proprietary analytical tools and data platforms quickly and independently",
]

PREFERRED_TEMPLATES = [
    "Prior internship in HR, finance, consulting, or data analysis",
    "Coursework in statistics, econometrics, industrial-organizational psychology, or labor economics",
    "Experience with any data analysis tool beyond Excel (Python, R, Stata, or similar)",
    "Familiarity with compensation concepts such as base pay, total cash, or equity compensation",
    "Experience with survey research methods or primary data collection",
    "Prior research assistant experience in an academic or professional setting",
    "Demonstrated interest in workforce equity, pay transparency, or labor market policy",
    "Familiarity with proxy statements, annual reports, or SEC filings",
    "Experience with professional writing — reports, memos, or analytical summaries",
    "Participation in relevant student organizations, case competitions, or academic research projects",
    "Demonstrated ability to produce clean, professional-quality work products",
    "Strong attention to formatting and visual presentation of data",
    "Familiarity with any benchmarking or survey platform (Radford, Mercer, WTW, or similar)",
    "Experience supporting a professional team in a coordinator or assistant capacity",
    "Prior exposure to financial statements or corporate disclosures",
    "Knowledge of basic statistical concepts including mean, median, percentile, and regression",
    "Demonstrated commitment to professional development, including self-directed learning",
    "Multilingual capability, particularly Spanish, for support on multinational client research",
    "Prior work or volunteer experience in a mission-driven organization",
    "Interest in Milwaukee and the Midwest as a place to build a professional career",
]

FIRST_YEAR_MILESTONES = [
    "Complete all modules of the {program} and pass the associated skills assessment",
    "Develop working proficiency with ACPWB's primary data platforms and benchmarking tools",
    "Contribute independently to at least two full engagement cycles, from data collection through final review",
    "Receive a mid-year performance review that confirms you are progressing on the expected development path",
    "Build a strong working relationship with your assigned {mentor_title} and seek feedback proactively",
    "Complete at least one professional development activity outside ACPWB — webinar, workshop, or certification module",
    "Demonstrate consistent accuracy in data entry and analysis work with a declining error rate over time",
    "Take ownership of at least one recurring team responsibility — a data update, a tracking log, a template — and execute it reliably",
    "Produce at least one deliverable component that requires no significant revision from a senior reviewer",
    "Ask good questions consistently — about methodology, about client context, about why we do things the way we do",
    "Complete all mandatory compliance and training requirements on or ahead of schedule",
    "Build positive working relationships with at least three peers across different parts of the firm",
    "Understand ACPWB's core analytical methodology well enough to explain it to a new hire",
    "Establish a clear picture of your own strengths and development areas through honest self-assessment and manager feedback",
    "Shadow at least two client-facing interactions to build understanding of how ACPWB presents its work",
    "Identify one area where you can add unique value to the team — and begin doing so",
    "Complete your first annual performance review cycle with documented goals for year two",
    "Demonstrate the ability to independently manage a small, well-defined research task from start to finish",
]

REPORTS_TO_TEMPLATES = [
    "Senior Analyst, {department}",
    "Manager, {department}",
    "{department} Team Lead",
    "Senior Associate, {department}",
    "Associate Director, {department}",
    "Director, {department} — with day-to-day oversight from a designated Senior Analyst",
    "Manager or Senior Manager, {department} (depending on team structure at time of hire)",
    "{department} Practice Manager",
    "Senior Analyst or Manager, assigned at time of onboarding",
    "Team Manager, {department}, with formal mentorship from a Senior Analyst",
]
