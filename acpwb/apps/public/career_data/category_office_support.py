JOB_TITLES = [
    "Office Manager",
    "Executive Assistant to the CEO",
    "Administrative Assistant, Research Operations",
    "Receptionist and Office Coordinator",
    "Facilities and Office Services Coordinator",
    "Executive Assistant, Advisory Services",
    "Administrative Manager",
    "Office Services Specialist",
    "Executive Coordinator, Leadership Team",
    "Finance and Operations Associate",
    "Accounting Coordinator",
    "Human Resources Generalist",
    "Talent Acquisition Coordinator",
    "Marketing and Communications Coordinator",
    "Events and Conferences Coordinator",
    "Payroll Specialist",
    "Benefits Administrator",
    "Staff Accountant",
    "HR Operations Coordinator",
    "Office Administrator",
]

DEPARTMENTS = [
    "Operations",
    "Human Resources",
    "Finance & Accounting",
    "Executive Office",
    "Facilities",
    "Marketing & Communications",
    "Talent Acquisition",
    "People Operations",
]

LEVELS = {
    "Administrative": (50, 68),
    "Senior Administrative": (60, 82),
    "Specialist": (62, 88),
    "Manager": (76, 108),
}

YEARS_EXP_RANGE = (0, 5)

MISSION_AREAS = [
    "executive and leadership support",
    "office operations and facilities management",
    "HR administration and employee services",
    "financial operations and accounting support",
    "talent acquisition and onboarding",
    "corporate communications and marketing",
    "payroll and benefits administration",
    "event planning and firm programming",
    "vendor management and procurement",
    "administrative operations and team support",
]

ABOUT_ROLE_TEMPLATES = [
    "The {title} is the person who keeps ACPWB's office running. That is a genuine compliment. "
    "Without someone doing this role well, {n} people have a noticeably worse day at work. "
    "You will own {mission_area} and be the first person others turn to when something needs to be handled. "
    "We are looking for someone who finds that responsibility satisfying rather than stressful.",

    "ACPWB is hiring a {title} to manage {mission_area} for the {department} team. "
    "This is a hands-on, operational role for someone who is organized, reliable, and "
    "comfortable being the person who makes the office work. If you do your job well, "
    "most people will not notice all the things you prevented from going wrong. We will notice.",

    "The {title} will join ACPWB's {department} team and take ownership of {mission_area}. "
    "This is not a role where you wait to be told what needs to be done — it is a role where "
    "you identify what needs to be done and do it. The {n} people who share this office "
    "will depend on you for that. They will also depend on you to keep the coffee situation "
    "under control, which is a responsibility we do not take lightly.",

    "ACPWB's {department} team needs a {title} who is genuinely good at {mission_area} "
    "and takes professional pride in operational excellence. You will manage day-to-day "
    "administrative and support functions for the firm and be the operational anchor "
    "for a team that does complex, client-facing advisory work. "
    "The better you are at your job, the better they can be at theirs.",

    "This {title} role is for someone who brings organization, professionalism, and a service "
    "orientation to {mission_area}. You will work within ACPWB's {department} group, "
    "supporting leadership and staff with the operational and administrative functions that "
    "keep a {n}-person firm operating smoothly. We are direct about expectations "
    "and generous in recognition when those expectations are met.",

    "The {title} will be an important part of ACPWB's {department} team, responsible for "
    "{mission_area} in an organization that values its support staff as much as its advisors. "
    "You will handle a range of tasks that require both attention to detail and good judgment. "
    "Some will be routine. Some will require you to figure things out as you go. "
    "We are looking for someone who handles both with equal professionalism.",

    "ACPWB is looking for a {title} to own {mission_area} for a firm that has grown to a size "
    "where these functions require dedicated expertise. You will be embedded in the {department} "
    "team and work closely with a range of staff at all levels. "
    "Clear communication, reliability, and discretion are all required. "
    "Familiarity with Milwaukee's vendor landscape is a bonus.",

    "The {title} manages {mission_area} for ACPWB's {department} team of {n}. "
    "This is a role where your work product is the operational environment everyone else works in, "
    "which means the standard is high but the impact is immediate and visible. "
    "We are a professional services firm that practices what it publishes about treating employees well, "
    "and that starts with ensuring the people who keep the firm running are recognized and supported.",

    "This {title} role is an opportunity to take ownership of {mission_area} as the firm continues to grow. "
    "The role is operational at its core — managing the systems, processes, and relationships "
    "that keep {department} functioning — but it requires real judgment and professional maturity. "
    "We are not looking for someone to follow a checklist. We are looking for someone who writes the checklist.",

    "The {title} position in ACPWB's {department} team is an opportunity to be the operational "
    "foundation of a growing professional services firm. You will handle {mission_area} "
    "and everything adjacent to it, working with leadership and staff who are genuinely good at their jobs "
    "and genuinely appreciative of people who make it easier to do those jobs. "
    "We have a very clean office. We intend to keep it that way.",
]

RESPONSIBILITY_TEMPLATES = [
    "Manage {mission_area} for the {department} team with accuracy, professionalism, and appropriate urgency",
    "Serve as the primary administrative point of contact for staff, visitors, and vendor inquiries",
    "Maintain organized systems for records, files, contracts, and correspondence in accordance with firm policies",
    "Coordinate executive calendars, travel arrangements, and meeting logistics for senior leadership",
    "Process invoices, purchase orders, and expense reports in accordance with ACPWB's financial procedures",
    "Support HR functions including new hire onboarding, offboarding, benefits administration, and records management",
    "Manage office supply inventory, vendor relationships, and facilities service providers",
    "Coordinate firm events including all-hands meetings, team offsites, and external client visits",
    "Prepare and distribute internal communications, announcements, and documentation on behalf of leadership",
    "Answer and direct incoming phone calls, emails, and visitors in a professional and welcoming manner",
    "Maintain confidentiality of sensitive personnel, financial, and client information at all times",
    "Assist with payroll preparation, timesheet review, and coordination with the firm's payroll provider",
    "Coordinate with IT on equipment procurement, setup, and support for staff as needed",
    "Support the talent acquisition process including interview scheduling, candidate communications, and offer logistics",
    "Manage the firm's physical facilities including space planning, maintenance requests, and safety compliance",
    "Prepare and file documents, reports, and correspondence in a timely and organized manner",
    "Assist with marketing and communications projects including collateral assembly and event promotion",
    "Track and report on {department} budget items including invoices, approvals, and recurring expenses",
    "Support the annual performance review cycle with administrative coordination and documentation",
    "Coordinate volunteer and community engagement activities as part of ACPWB's firm culture programs",
    "Serve as point of contact for building management, cleaning services, and facilities vendors",
    "Maintain the firm's employee directory, organizational charts, and internal resource materials",
    "Support audit, compliance, and regulatory filing processes by organizing and producing required documentation",
    "Assist with the planning and logistics of ACPWB's annual compensation benchmarking conference",
    "Handle sensitive correspondence and communications with appropriate discretion and professionalism",
]

REQUIREMENT_TEMPLATES = [
    "{n}+ years of experience in an administrative, HR, accounting, or office management role",
    "Strong organizational skills with a demonstrated ability to manage multiple tasks and shifting priorities",
    "Proficiency in Microsoft Office Suite including Outlook, Word, Excel, and PowerPoint",
    "Professional written and verbal communication skills with a polished, service-oriented style",
    "Demonstrated ability to maintain confidentiality and exercise judgment with sensitive information",
    "Experience in an office or professional services environment",
    "Comfortable managing vendor relationships, service contracts, and routine procurement",
    "Attention to detail and commitment to accuracy in administrative and financial tasks",
    "Friendly, professional presence and the ability to represent ACPWB well to visitors and clients",
    "Prior experience with scheduling, calendar management, or executive support",
    "Ability to work independently and take initiative without requiring constant direction",
    "Experience with accounts payable, expense management, or basic bookkeeping",
    "Familiarity with HR processes including onboarding, benefits enrollment, and employee record management",
    "Demonstrated ability to create and maintain organized filing and documentation systems",
    "Strong interpersonal skills and a genuine interest in supporting a professional team",
    "Reliable, punctual, and consistently professional in appearance and conduct",
    "Comfort working in a fast-paced environment with shifting priorities and occasional time pressure",
    "Ability to lift and carry up to 30 pounds for office supply management and event setup",
]

PREFERRED_TEMPLATES = [
    "Prior experience at a professional services, consulting, or advisory firm",
    "Familiarity with HR systems (Rippling, BambooHR, ADP, or equivalent)",
    "Experience with accounting software (QuickBooks, Sage, or equivalent)",
    "Exposure to payroll processing and tax withholding requirements",
    "Experience supporting C-suite executives in a high-volume environment",
    "Event planning or conference coordination background",
    "Familiarity with Milwaukee's vendor community and local service providers",
    "Experience with facilities management or workplace operations",
    "Background in benefits administration including health, dental, and retirement plan management",
    "Demonstrated interest in professional development and expanded scope over time",
    "Prior experience with legal document management or compliance filings",
    "Familiarity with digital collaboration tools including Slack, Zoom, and Teams",
    "Experience in talent acquisition coordination including applicant tracking systems",
    "Background in marketing, communications, or event production",
    "Prior experience supporting a distributed team in a hybrid work environment",
    "Familiarity with expense reporting platforms (Concur, Expensify, or equivalent)",
    "Track record of working effectively across multiple departments and levels",
]

FIRST_YEAR_MILESTONES = [
    "Learn ACPWB's office systems, processes, and vendors thoroughly enough to handle routine requests independently",
    "Own at least one recurring {mission_area} process end-to-end without requiring regular oversight",
    "Build positive, professional working relationships with staff at all levels of the firm",
    "Establish yourself as a reliable, responsive point of contact for the {department} team",
    "Complete any required compliance, HR, or role-specific training before their respective deadlines",
    "Identify and implement at least one small but meaningful improvement to an existing operational process",
    "Successfully coordinate at least one firm-wide event or initiative from planning through execution",
    "Receive formal acknowledgment from at least two stakeholders that your work has made their job easier",
    "Develop familiarity with ACPWB's services and clients sufficient to make informed routing and scheduling decisions",
    "Build and maintain accurate, well-organized records for the core administrative systems you own",
    "Demonstrate consistent accuracy in financial and HR administrative tasks with no material errors",
    "Propose at least one vendor, tool, or service improvement based on your experience managing {mission_area}",
    "Complete your first annual performance review with documented goals and an honest self-assessment",
    "Establish a reputation for reliability — that deadlines are met, requests are followed up on, and nothing is dropped",
    "Take ownership of the firm's physical environment by ensuring it is clean, functional, and professionally presented",
    "Support the successful onboarding of any new hires by managing the logistics and providing a welcoming first day",
]

REPORTS_TO_TEMPLATES = [
    "Director of Operations",
    "Chief of Staff",
    "VP of {department}",
    "Director of {department}",
    "Office Manager or Director of Operations",
    "Chief Administrative Officer",
    "Senior Manager, Operations and Administration",
    "Managing Director or COO",
    "Director of People Operations",
    "VP of Finance and Operations",
]
