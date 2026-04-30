JOB_TITLES = [
    "IT Systems Administrator",
    "Senior Software Engineer, Internal Tools",
    "Data Engineer, Compensation Platforms",
    "Full Stack Developer, Client Portals",
    "Database Administrator",
    "Business Intelligence Developer",
    "Infrastructure Engineer",
    "DevOps Engineer",
    "Senior Data Engineer",
    "Software Engineer, Analytics Infrastructure",
    "Information Security Analyst",
    "IT Help Desk Specialist",
    "Systems Integration Engineer",
    "Cloud Infrastructure Engineer",
    "Senior Backend Engineer",
    "Data Analyst, Business Intelligence",
    "Frontend Developer, Research Tools",
    "Network Administrator",
    "Platform Engineer",
    "IT Project Manager",
    "Analytics Engineer",
    "Security Operations Analyst",
    "Application Support Engineer",
    "Senior BI Developer",
    "Data Platform Architect",
]

DEPARTMENTS = [
    "Data Infrastructure",
    "Technology & Engineering",
    "Business Intelligence",
    "Information Technology",
    "Platform Engineering",
    "Analytics Engineering",
    "IT Operations",
    "Security & Compliance",
]

LEVELS = {
    "Junior": (78, 108),
    "Mid-Level": (100, 142),
    "Senior": (130, 178),
    "Staff": (158, 218),
    "Lead": (148, 205),
}

YEARS_EXP_RANGE = (1, 7)

MISSION_AREAS = [
    "data infrastructure reliability and scalability",
    "compensation analytics platform development",
    "internal systems performance and security",
    "business intelligence and self-service reporting",
    "application development and maintenance",
    "cloud platform operations and optimization",
    "data engineering and pipeline development",
    "IT operations and service delivery",
    "cybersecurity and access management",
    "data governance and quality assurance",
    "internal tooling and developer productivity",
    "database administration and performance tuning",
]

TECH_TOOLS = [
    "Python and SQL",
    "Tableau and Snowflake",
    "dbt and BigQuery",
    "React and TypeScript",
    "PowerShell and Azure Active Directory",
    "Python, Django, and PostgreSQL",
    "Terraform and AWS",
    "Spark and Databricks",
    "Power BI and Azure Synapse",
    "Node.js and PostgreSQL",
    "Airflow and Redshift",
    "FastAPI and Docker",
]

SYSTEMS = [
    "ACPWB compensation benchmarking platform",
    "internal data pipeline and ETL infrastructure",
    "client-facing analytics and reporting portal",
    "employee data warehouse",
    "internal reporting and dashboard infrastructure",
    "compensation survey processing system",
    "ACPWB's core data platform",
    "internal authentication and access management systems",
    "the firm's business intelligence environment",
    "ACPWB's client data exchange platform",
]

PLATFORMS = [
    "Amazon Web Services (AWS)",
    "Microsoft Azure",
    "Google Cloud Platform",
    "on-premise infrastructure and hybrid cloud",
    "Azure and on-premise hybrid environment",
    "AWS and Snowflake",
]

INDUSTRY_SECTORS = [
    "financial services",
    "healthcare and life sciences",
    "technology",
    "professional services",
    "compensation analytics",
    "HR technology",
]

ABOUT_ROLE_TEMPLATES = [
    "The {title} is responsible for the technical infrastructure that makes ACPWB's analytical work possible. "
    "The {department} team of {n} builds and maintains the {system} that powers our advisory practice, "
    "and this role is central to ensuring that infrastructure is reliable, secure, and scalable. "
    "Our analytical staff depends on this work. So do our clients, indirectly. You will know both are counting on you.",

    "ACPWB's {department} team is looking for a {title} who takes genuine ownership of the {system}. "
    "This is not a break-fix role or a ticket queue. You will work in {tech_stack}, contribute to the "
    "{mission_area} agenda, and collaborate closely with analysts and advisors who need technology "
    "that simply works. We have a small, excellent team. We are looking for someone who fits.",

    "The {title} will join ACPWB's {department} group and own a defined portion of the {system}. "
    "The work spans {tech_stack} on {platform}, balancing feature development with the maintenance "
    "and reliability that a professional services firm depends on. We treat engineering as a professional discipline "
    "here — not a cost center. That distinction affects how this team operates day to day.",

    "ACPWB is hiring a {title} to support {mission_area} within the {department} team. "
    "The work involves {tech_stack} development on {platform}, with an emphasis on building "
    "tools and systems that allow ACPWB's analysts to do better work faster. "
    "You will be embedded in a {n}-person team that is technical without being insular "
    "and collaborative without being slow.",

    "The {title} role in ACPWB's {department} group is for someone who finds real satisfaction "
    "in making complex systems work well. You will take responsibility for {mission_area} "
    "using {tech_stack} on {platform}, and your work will have immediate, visible impact "
    "on the quality and efficiency of ACPWB's analytical operations. We have {n} people on this team. "
    "Every contribution matters.",

    "ACPWB's {department} team powers the {system} that supports our advisory practice. "
    "The {title} will contribute to {mission_area}, working in {tech_stack} and collaborating "
    "with a small team of technically rigorous engineers. We do not have layers of process. "
    "We do have high standards. These are different things.",

    "This {title} role requires someone who can own {mission_area} within {department} "
    "without waiting to be told what needs to be done. The {system} is critical infrastructure, "
    "and the people who depend on it — our analysts, our clients, our firm — need it to work. "
    "You will use {tech_stack} on {platform} to build and maintain systems that meet that standard.",

    "ACPWB is seeking a {title} to help the {department} team scale the {system} "
    "to meet growing demand from the advisory practice. You will work in {tech_stack}, "
    "contribute to {mission_area} initiatives, and collaborate with a team of {n} "
    "who take engineering craft seriously and treat technical debt as a real liability.",

    "The {title} will be a key contributor to ACPWB's {department} practice, "
    "working on {mission_area} that directly supports our compensation advisory work. "
    "If you have experience with {tech_stack} and are looking for a role where your engineering output "
    "has clear business impact rather than disappearing into a backlog, ACPWB is worth your attention.",

    "ACPWB's {department} team of {n} runs the technical infrastructure that makes our analytical work credible. "
    "The {title} will take ownership of specific components of the {system}, "
    "working in {tech_stack} on {platform}. We are a small team that covers significant technical scope, "
    "which means you will have breadth, ownership, and visibility that larger organizations rarely offer.",

    "The {title} will join ACPWB's {department} team to drive {mission_area} improvements "
    "to the {system}. Working in {tech_stack}, you will partner with data engineers, analysts, "
    "and business stakeholders to build solutions that have measurable impact on firm performance. "
    "We deploy frequently, review code carefully, and invest in the quality of our technical foundation.",

    "ACPWB's analytical practice runs on data, and the {title} in our {department} team "
    "is responsible for ensuring that data infrastructure is sound. The work spans {mission_area} "
    "for a team of {n} engineers who treat reliability, security, and developer experience as first-class concerns. "
    "If systems going down at inopportune moments bothers you personally, you will fit in here.",
]

RESPONSIBILITY_TEMPLATES = [
    "Design, build, and maintain components of the {system} using {tech_stack}",
    "Ensure reliability, performance, and security of {mission_area} infrastructure on {platform}",
    "Monitor system health, troubleshoot issues, and respond to incidents with appropriate urgency and documentation",
    "Partner with analytical and advisory teams to understand data requirements and translate them into technical solutions",
    "Contribute to code reviews, technical documentation, and engineering standards for the {department} team",
    "Build and maintain automated data pipelines and ETL processes supporting the {system}",
    "Design and enforce access controls, authentication policies, and data security protocols",
    "Develop and maintain BI dashboards and analytical reports for internal stakeholders using {tech_stack}",
    "Manage infrastructure-as-code configurations and deployment automation on {platform}",
    "Support end users of the {system} with technical guidance, training, and issue resolution",
    "Evaluate and recommend tools, platforms, and architectural approaches for {mission_area} initiatives",
    "Document system architecture, data models, and operational runbooks for the {department} team",
    "Participate in sprint planning, standups, and retrospectives as part of the {department} engineering cadence",
    "Manage vendor relationships for technical services and software subscriptions relevant to {mission_area}",
    "Implement and maintain backup, recovery, and disaster preparedness procedures for the {system}",
    "Build internal tooling to improve analyst productivity and data access across the firm",
    "Lead or contribute to technical evaluation and procurement for new systems and infrastructure",
    "Conduct performance analysis and capacity planning for the {system} to support firm growth",
    "Enforce data quality standards and develop automated validation and monitoring for data pipelines",
    "Collaborate with Information Security on vulnerability assessments, patching, and compliance audits",
    "Maintain network and systems infrastructure for ACPWB's Milwaukee office and remote work environments",
    "Develop APIs and integrations connecting the {system} with third-party data sources and internal tools",
    "Support the migration of legacy systems to modern platforms including {platform}",
    "Provide technical guidance and code review to junior engineers and associate developers",
    "Contribute to ACPWB's technical roadmap by identifying investment opportunities in {mission_area}",
]

REQUIREMENT_TEMPLATES = [
    "{n}+ years of professional experience in software engineering, data engineering, or a closely related technical discipline",
    "Strong proficiency in {tech_stack} or an equivalent technical stack with demonstrated production experience",
    "Experience building and operating systems on {platform} in a professional environment",
    "Solid understanding of relational database design, SQL, and data modeling principles",
    "Demonstrated ability to write clean, tested, maintainable code with attention to documentation",
    "Experience with version control (Git), code review practices, and CI/CD pipelines",
    "Strong troubleshooting and debugging skills with the ability to diagnose issues under time pressure",
    "Experience with monitoring, alerting, and observability tooling for production systems",
    "Ability to communicate technical concepts clearly to non-technical stakeholders",
    "Demonstrated ability to take ownership of technical systems and follow through on commitments independently",
    "Familiarity with information security fundamentals including authentication, encryption, and access control",
    "Experience with ETL or data pipeline development and maintenance",
    "Understanding of networking fundamentals including DNS, TCP/IP, firewalls, and VPN",
    "{n}+ years of experience managing or contributing to production systems with reliability requirements",
    "Experience with cloud infrastructure management including compute, storage, and networking on {platform}",
    "Demonstrated experience working collaboratively in a small, cross-functional technical team",
    "Strong attention to documentation, runbooks, and knowledge transfer in technical environments",
    "Experience working in an Agile or iterative development environment",
    "Ability to prioritize technical work based on business impact and operational risk",
    "Demonstrated curiosity about new tools and technologies with evidence of self-directed technical learning",
]

PREFERRED_TEMPLATES = [
    "Relevant cloud certification (AWS Solutions Architect, Azure Administrator, Google Professional Data Engineer, or equivalent)",
    "Experience with compensation analytics, HR technology, or financial services data platforms",
    "Familiarity with data warehousing concepts and columnar storage platforms",
    "Experience with data orchestration tools such as Apache Airflow, Prefect, or equivalent",
    "Background in BI development using Tableau, Power BI, Looker, or equivalent",
    "Experience with containerization and orchestration (Docker, Kubernetes)",
    "Familiarity with infrastructure-as-code tools (Terraform, Pulumi, CloudFormation)",
    "Prior experience in a professional services or advisory firm technology environment",
    "Experience with Python for data processing including pandas, NumPy, or PySpark",
    "Familiarity with API design and RESTful service development",
    "Experience with PostgreSQL, Snowflake, or other enterprise relational or analytical databases",
    "Background in application security including OWASP principles and secure development practices",
    "Prior experience mentoring or reviewing work of junior engineers",
    "Familiarity with compliance requirements such as SOC 2, GDPR, or CCPA as they apply to data systems",
    "Experience with feature flagging, A/B testing infrastructure, or progressive deployment patterns",
    "Demonstrated interest in data quality, data contracts, or observability as engineering practices",
    "Experience building internal developer tooling or productivity infrastructure",
    "Familiarity with dbt, Great Expectations, or other data transformation and validation frameworks",
    "Prior involvement in technology selection or architectural decision-making in a professional context",
]

FIRST_YEAR_MILESTONES = [
    "Develop a thorough working knowledge of the {system} architecture, data flows, and dependencies",
    "Ship at least two meaningful features or improvements to the {system} with full documentation",
    "Pass any required security, compliance, or systems access certifications for the {department} team",
    "Establish a reliable on-call rotation and demonstrate competent incident response",
    "Reduce at least one identified pain point for the analytical team through tooling or infrastructure improvement",
    "Document the components of the {system} you own at a level sufficient for another engineer to maintain them",
    "Complete a successful deployment to production with no unplanned rollbacks or escalations",
    "Build a productive working relationship with at least two non-engineering stakeholders who depend on the {system}",
    "Identify and resolve at least one performance, security, or reliability gap in the existing infrastructure",
    "Contribute meaningfully to the team's technical roadmap discussions with grounded, evidence-based input",
    "Establish your own development environment, testing practices, and review standards consistent with team norms",
    "Complete any required compliance training modules and system access certifications",
    "Support at least one cross-functional project where your technical work enables an advisory deliverable",
    "Lead at least one sprint or project phase end-to-end from planning through deployment",
    "Receive positive feedback from a non-technical stakeholder on the reliability or usability of a system you own",
    "Demonstrate the ability to scope technical work accurately and deliver against that scope",
    "Build familiarity with ACPWB's business context to make better technical prioritization decisions",
]

REPORTS_TO_TEMPLATES = [
    "Director of {department}",
    "VP of {department}",
    "Head of Technology, with dotted-line to {department} practice leads",
    "Engineering Manager, {department}",
    "Director of Data Infrastructure",
    "VP of Technology and Data",
    "Principal Engineer or Technical Lead, {department}",
    "CTO or Head of Engineering (for senior or staff-level roles)",
    "Senior Engineer, {department} — with technical mentorship from a Staff or Principal",
    "Director of {department}, with cross-functional responsibilities to analytical practice leads",
]
