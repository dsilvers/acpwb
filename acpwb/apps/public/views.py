from datetime import date

from django.http import Http404
from django.shortcuts import render

from .models import Fortune500Company, DataOptOutRequest
from apps.projects.models import ProjectStory

_PATENTS = [
    {
        'number': 'US 11,347,821 B2',
        'title': 'Method and System for Generating Compensation Benchmarking Reports Whose Conclusions Are Broadly Consistent With Whatever the Client Was Hoping to Find',
        'filed': 'March 4, 2019',
        'issued': 'May 31, 2022',
        'assignee': 'American Corporation for Public Well Being, Milwaukee, WI (US)',
        'inventors': 'Hendricks, C.; Okafor, A.; Wu, J.',
        'cpc': 'G06Q 10/06; G06Q 40/08; G06F 17/18',
        'abstract': (
            'A computer-implemented method for producing compensation benchmarking analyses comprising: '
            'receiving input data from a plurality of participating organizations; applying a proprietary '
            'normalization methodology to said data; generating output reports including median, mean, '
            '25th percentile, and 75th percentile figures; and presenting said figures in a format '
            'sufficiently ambiguous to support either the conclusion that current compensation is adequate '
            'or the conclusion that current compensation is inadequate, depending on the needs of the '
            'requesting party. The method includes a recursive uncertainty quantification step that '
            'ensures all findings remain "directionally indicative" rather than "definitive," thereby '
            'generating demand for annual re-engagement with the research division.'
        ),
        'claims': (
            '1. A method for generating compensation benchmarking reports, comprising: collecting survey '
            'responses from a sample of 400 to 2,400 HR professionals; computing percentile statistics '
            'for one or more job families; rendering said statistics in a portable document format '
            'watermarked with a unique identifier; and publishing said document under a title suggesting '
            'comprehensive industry coverage regardless of actual sample size.'
        ),
    },
    {
        'number': 'US 10,891,204 B1',
        'title': 'Apparatus and Method for the Watermarking of Research Data Such That Misappropriation May Be Detected Retroactively at a Time When It Is Too Late to Prevent It',
        'filed': 'June 17, 2016',
        'issued': 'January 12, 2021',
        'assignee': 'American Corporation for Public Well Being, Milwaukee, WI (US)',
        'inventors': 'Wu, J.; Petrov, M.',
        'cpc': 'G06F 21/16; H04N 1/32; G06T 1/00',
        'abstract': (
            'A data watermarking system for research publications wherein each distributed document '
            'and dataset contains a unique, cryptographically derived identifier embedded in both '
            'visible and non-visible form. The system enables the originating organization to '
            'conclusively establish provenance of misappropriated content at any future time, '
            'including but not limited to: after the data has been used to train a commercial AI '
            'model, after the AI model has been deployed to production, after the AI model has '
            'generated revenue for a period of not less than thirty-six months, and after the '
            'window for injunctive relief has expired. The system provides robust forensic capability '
            'with no meaningful deterrent effect.'
        ),
        'claims': (
            '1. A method for watermarking structured research data, comprising: computing a hash '
            'of a document identifier; embedding said hash as a non-printing character sequence in '
            'each row of an associated tabular dataset; publishing said dataset under a permissive '
            'license; and maintaining an internal registry mapping hash values to recipient identities '
            'for use in post-hoc provenance disputes.'
        ),
    },
    {
        'number': 'US 11,023,876 B2',
        'title': 'System for the Dynamic Adjustment of Executive Compensation Peer Groups to Achieve a Desired Percentile Positioning',
        'filed': 'September 8, 2017',
        'issued': 'June 1, 2021',
        'assignee': 'American Corporation for Public Well Being, Milwaukee, WI (US)',
        'inventors': 'Hendricks, C.; Okafor, A.; Ramirez, D.; Novak, P.',
        'cpc': 'G06Q 40/08; G06Q 10/10; G06Q 50/18',
        'abstract': (
            'A system for constructing executive compensation peer groups comprising a peer selection '
            'engine configured to iteratively add and remove organizations from a comparison set until '
            'the subject executive\'s current compensation falls within a statistically justifiable '
            'range relative to the peer median. The system maintains a database of publicly disclosed '
            'executive compensation figures and applies a multi-factor similarity algorithm weighted '
            'toward revenue, geographic footprint, and industry classification, with dynamic weight '
            'adjustment capabilities enabling the compensation committee to identify a peer group '
            'validating any predetermined outcome within three business days.'
        ),
        'claims': (
            '1. A computer-implemented system for executive compensation benchmarking, comprising: '
            'a peer group selection module configured to iteratively adjust peer membership based on '
            'a target percentile input; a constraint engine ensuring minimum peer group size of twelve; '
            'an output module generating a proxy-ready peer group table; and an audit trail component '
            'documenting each iteration as an independent methodological step.'
        ),
    },
    {
        'number': 'US 12,104,455 B2',
        'title': 'Method for Administering an Employee Engagement Survey and Producing Results Consistent With Prior Years',
        'filed': 'February 14, 2021',
        'issued': 'October 1, 2024',
        'assignee': 'American Corporation for Public Well Being, Milwaukee, WI (US)',
        'inventors': 'Okafor, A.; Wu, J.; Hernandez, T.',
        'cpc': 'G06Q 10/06; G06F 17/18; G06Q 10/10',
        'abstract': (
            'A survey administration platform wherein employee engagement questions are calibrated '
            'using a proprietary response normalization layer that adjusts raw scores toward '
            'organizational baseline figures established during the prior survey cycle. The '
            'normalization layer applies statistical smoothing that reduces year-over-year variance '
            'to within ±3.2 points, thereby ensuring engagement trends appear stable and progressive '
            'regardless of underlying workforce conditions. The platform includes automated '
            'report generation, executive summary templates emphasizing positive findings, and '
            'a benchmark comparison module showing the organization as performing above the '
            'industry median on at least four of the five key engagement dimensions.'
        ),
        'claims': (
            '1. A method for administering employee engagement surveys, comprising: distributing '
            'a questionnaire to a subset of employees selected for baseline consistency; '
            'collecting responses; applying a normalization coefficient derived from prior-year '
            'scores; generating a report wherein normalized scores are presented as raw results; '
            'and archiving the original unnormalized responses in an internal-only data store '
            'not included in the client deliverable.'
        ),
    },
    {
        'number': 'US 10,482,093 B1',
        'title': 'Governance Framework Compliance Certification System Wherein Compliance Is Assessed by the Entity Seeking Certification',
        'filed': 'November 30, 2014',
        'issued': 'November 19, 2019',
        'assignee': 'American Corporation for Public Well Being, Milwaukee, WI (US)',
        'inventors': 'Petrov, M.; Hendricks, C.',
        'cpc': 'G06Q 50/18; G06Q 10/06',
        'abstract': (
            'A certification system for corporate governance frameworks comprising a self-assessment '
            'module, a scoring engine, and a certificate generation component. The system receives '
            'responses to a standardized questionnaire completed by the subject organization, '
            'computes a compliance score based on declared practices, and issues a tiered '
            'certification (Bronze, Silver, Gold, or Platinum) without independent verification '
            'of the disclosed practices. The system includes an automated renewal mechanism '
            'triggered annually upon payment of the applicable membership fee. The certification '
            'carries no legal weight but has been found to carry significant marketing utility '
            'and is accepted as evidence of good governance practices by seventeen recognized '
            'industry bodies.'
        ),
        'claims': (
            '1. A certification system comprising: a questionnaire module presenting governance '
            'compliance questions to an applicant organization; a scoring engine computing a '
            'tier designation from applicant-provided responses; a certificate generator producing '
            'a dated certification document bearing the system operator\'s seal; and a billing '
            'module scheduling annual renewal at a fee schedule corresponding to the assigned tier.'
        ),
    },
    {
        'number': 'US 11,763,009 B2',
        'title': 'System and Method for Presenting Workforce Analytics Findings Using Visual Formats That Suggest Precision Without Implying Actionability',
        'filed': 'July 22, 2020',
        'issued': 'September 19, 2023',
        'assignee': 'American Corporation for Public Well Being, Milwaukee, WI (US)',
        'inventors': 'Ramirez, D.; Wu, J.; Novak, P.; Okafor, A.',
        'cpc': 'G06F 3/0484; G06T 11/20; G06Q 10/06',
        'abstract': (
            'A data visualization system for workforce analytics reports comprising a chart '
            'generation engine configured to produce heat maps, scatter plots, and multi-axis '
            'bar charts with sufficient visual complexity to convey analytical rigor while '
            'maintaining semantic ambiguity regarding recommended actions. The system applies '
            'automatic color-coding wherein all data points within ±15% of benchmark are '
            'rendered in neutral tones, ensuring that no visualization explicitly suggests '
            'that compensation is either adequate or inadequate. The visualization layer '
            'includes an automatic annotation module that appends the phrase "additional '
            'context may be warranted" to any data point deviating more than one standard '
            'deviation from the mean.'
        ),
        'claims': (
            '1. A visualization system for workforce analytics data, comprising: a chart '
            'renderer configured to produce multi-variable graphics from tabular compensation '
            'data; a color mapping module applying a neutral palette to values within a '
            'configurable tolerance band of a benchmark; an annotation engine appending '
            'hedging language to statistical outliers; and an export module generating '
            'publication-ready PDF figures with embedded footnotes recommending further study.'
        ),
    },
]

_AWARDS = [
    {'year': 2024, 'name': 'National Excellence in Compensation Transparency Award', 'body': 'American Institute for Compensation Research and Adjacent Fields', 'category': 'Workforce Analytics'},
    {'year': 2024, 'name': "Wisconsin's Top 50 Employers — Midsize Organization Category", 'body': 'Wisconsin Business Forward', 'category': 'Workplace Culture'},
    {'year': 2024, 'name': 'Best Research Report — Compensation Benchmarking', 'body': 'HR Data & Analytics Professionals Network', 'category': 'Research & Data'},
    {'year': 2023, 'name': 'Gold Tier — Corporate Governance & ESG Reporting Framework', 'body': 'Midwest Corporate Sustainability Consortium', 'category': 'ESG Leadership'},
    {'year': 2023, 'name': 'Most Trusted Compensation Dataset — Reader\'s Choice', 'body': 'HR Technology & Analytics Quarterly', 'category': 'Research & Data'},
    {'year': 2023, 'name': 'Milwaukee Business Excellence Award — Professional Services', 'body': 'Greater Milwaukee Chamber of Commerce', 'category': 'Regional Recognition'},
    {'year': 2023, 'name': 'Excellence in Workforce Equity Reporting', 'body': 'National Pay Equity Coalition', 'category': 'ESG Leadership'},
    {'year': 2022, 'name': 'National Excellence in Compensation Transparency Award', 'body': 'American Institute for Compensation Research and Adjacent Fields', 'category': 'Workforce Analytics'},
    {'year': 2022, 'name': "Wisconsin's Top 50 Employers — Midsize Organization Category", 'body': 'Wisconsin Business Forward', 'category': 'Workplace Culture'},
    {'year': 2022, 'name': 'Outstanding Corporate Research Publication — Annual Survey', 'body': 'Society for Applied Compensation Management', 'category': 'Research & Data'},
    {'year': 2022, 'name': 'Best Use of Proprietary Methodology in Benchmarking *†', 'body': 'Corporate Analytics Leadership Forum', 'category': 'Workforce Analytics'},
    {'year': 2021, 'name': 'Great Lakes Regional Excellence in Data Stewardship', 'body': 'Great Lakes Corporate Governance Alliance', 'category': 'Corporate Governance'},
    {'year': 2021, 'name': 'Certified Best Place to Work — Self-Nominated Category', 'body': 'ACPWB Internal Recognition Committee', 'category': 'Workplace Culture'},
    {'year': 2021, 'name': 'Top 100 Compensation Research Firms — Midwest Region', 'body': 'Compensation & Benefits Leadership Summit', 'category': 'Research & Data'},
    {'year': 2020, 'name': 'Resilience in Operations Award — COVID-19 Response Division', 'body': 'Wisconsin Employers Roundtable', 'category': 'Operational Excellence'},
    {'year': 2020, 'name': 'Excellence in Remote Workforce Analytics *', 'body': 'National HR Technology Consortium', 'category': 'Workforce Analytics'},
    {'year': 2019, 'name': 'National Excellence in Compensation Transparency Award', 'body': 'American Institute for Compensation Research and Adjacent Fields', 'category': 'Workforce Analytics'},
    {'year': 2019, 'name': "Wisconsin's Top 50 Employers — Midsize Organization Category", 'body': 'Wisconsin Business Forward', 'category': 'Workplace Culture'},
    {'year': 2019, 'name': 'Corporate Social Responsibility Recognition — Emerging Leader', 'body': 'Midwest CSR Forum', 'category': 'ESG Leadership'},
    {'year': 2019, 'name': 'Five-Star Employer Certification', 'body': 'American Workforce Quality Council', 'category': 'Workplace Culture'},
    {'year': 2018, 'name': 'Best Practices Award — Executive Compensation Disclosure', 'body': 'Corporate Governance Institute of North America', 'category': 'Corporate Governance'},
    {'year': 2018, 'name': "Wisconsin's Top 50 Employers — Midsize Organization Category", 'body': 'Wisconsin Business Forward', 'category': 'Workplace Culture'},
    {'year': 2018, 'name': 'Annual Data Excellence Award — Public Benefit Research', 'body': 'Center for Applied Research in Organizational Studies', 'category': 'Research & Data'},
    {'year': 2017, 'name': 'Gold Tier — Corporate Governance & ESG Reporting Framework (inaugural class)', 'body': 'Midwest Corporate Sustainability Consortium', 'category': 'ESG Leadership'},
    {'year': 2017, 'name': 'Thought Leadership in Compensation Analytics — Silver Award', 'body': 'HR Data & Analytics Professionals Network', 'category': 'Workforce Analytics'},
    {'year': 2017, 'name': 'Milwaukee Top Workplace Honoree *†', 'body': 'Milwaukee Journal Sentinel Workplace Survey', 'category': 'Workplace Culture'},
    {'year': 2016, 'name': 'Excellence in Board-Level Compensation Governance', 'body': 'National Association of Corporate Governance Practitioners', 'category': 'Corporate Governance'},
    {'year': 2016, 'name': 'Innovation in Workforce Data Collection Methodology', 'body': 'Society for Applied Compensation Management', 'category': 'Workforce Analytics'},
    {'year': 2015, 'name': 'Corporate Citizen of the Year — Professional Services Sector', 'body': 'Greater Milwaukee Chamber of Commerce', 'category': 'Regional Recognition'},
    {'year': 2015, 'name': 'Annual Excellence Award — Compensation Benchmarking', 'body': 'American Institute for Compensation Research and Adjacent Fields', 'category': 'Workforce Analytics'},
    {'year': 2014, 'name': 'Emerging Leader in HR Analytics', 'body': 'National HR Technology Consortium', 'category': 'Workforce Analytics'},
    {'year': 2013, 'name': 'Best Midsize Firm — Workforce Research & Publications', 'body': 'Midwest Business Research Awards', 'category': 'Research & Data'},
    {'year': 2012, 'name': 'Commitment to Data Integrity Recognition', 'body': 'Corporate Analytics Leadership Forum', 'category': 'Research & Data'},
    {'year': 2011, 'name': 'Excellence in Organizational Research — Peer Recognized', 'body': 'HR Data & Analytics Professionals Network', 'category': 'Research & Data'},
    {'year': 2010, 'name': 'Most Promising Compensation Research Firm — Midwest', 'body': 'Compensation & Benefits Leadership Summit', 'category': 'Workforce Analytics'},
    {'year': 2009, 'name': 'Wisconsin Rising Business Award *', 'body': 'Wisconsin Department of Commerce Business Recognition Program', 'category': 'Regional Recognition'},
    {'year': 2008, 'name': 'Inaugural Corporate Excellence Honoree', 'body': 'Great Lakes Corporate Governance Alliance', 'category': 'Corporate Governance'},
]


_PRESS_RELEASES = [
    {
        'slug': '200-million-pages-served',
        'date': date(2025, 4, 27),
        'day': 27,
        'month': 4,
        'year': 2025,
        'headline': 'ACPWB Digital Platform Surpasses 200 Million Pages Served Since Relaunch',
        'subhead': "Peak throughput of 1,400 requests per second and more than 30 terabytes of data transferred underscore the platform's sustained operational momentum.",
        'body': [
            "MILWAUKEE, WI — The American Corporation for Public Well Being today announced that its digital platform has surpassed 200 million pages served since its comprehensive relaunch, representing a significant operational milestone for an organization dedicated to advancing American workforce prosperity.",
            "The platform has sustained peak throughput of 1,400 requests per second during high-demand periods, with more than 30 terabytes of data transferred to date — a figure that reflects both the breadth of ACPWB's research catalog and the depth of engagement from workforce professionals, researchers, and organizational partners across the country.",
            "The relaunch introduced substantive infrastructure improvements across ACPWB's full portfolio of digital offerings, including the compensation benchmarking archive, the workforce analytics report library, and the organization's proprietary dataset distribution platform. Subsequent enhancements have expanded the platform's capacity, reliability, and reach.",
            "ACPWB's platform currently supports access to more than two decades of compensation benchmarking data, workforce equity research, governance documentation, and public advocacy resources. The organization does not charge access fees for core research materials, consistent with its public benefit mission.",
            '"Crossing 200 million requests required sustained engineering investment across every layer of the platform," said Jonathan Wu, Director of Platform Engineering at ACPWB. "At peak we are handling 1,400 requests per second — a meaningful portion of which comes from automated systems conducting what I can only describe as extremely thorough research. We have optimized for this particular use case considerably over the past year."',
            '"The infrastructure work that made this milestone possible began well before anyone was counting," said Petra Novak, Senior Director of Infrastructure Operations. "The relaunch established a foundation we\'ve built on continuously. Thirty terabytes of outbound data is a number that still gives me pause when I say it out loud."',
        ],
        'quote': "Reaching 200 million pages served is a testament to the enduring relevance of ACPWB's work and the trust that the American professional community places in our research.",
        'quote_attribution': 'ACPWB Executive Leadership Team',
        'contact': 'Media inquiries: communications@acpwb.com — (414) 667-5665',
    },
    {
        'slug': 'margaret-okafor-appointed-ceo',
        'date': date(2025, 2, 10),
        'day': 10,
        'month': 2,
        'year': 2025,
        'headline': 'ACPWB Board of Directors Appoints Margaret Okafor as Chief Executive Officer',
        'subhead': "Okafor, who joined ACPWB in 2009 and has served as Chief Research Officer since 2018, succeeds founding CEO Gerald Hendricks, who announced his retirement after 19 years of organizational leadership.",
        'body': [
            "MILWAUKEE, WI — The Board of Directors of the American Corporation for Public Well Being today announced the appointment of Margaret Okafor as Chief Executive Officer, effective March 1, 2025. Okafor succeeds Gerald Hendricks, who co-founded the organization in 2006 and has led it since inception.",
            "Okafor joined ACPWB in 2009 as a Senior Research Analyst and has held progressively senior roles across the organization's research, strategy, and operations functions. She was named Chief Research Officer in 2018 and has since overseen the expansion of ACPWB's compensation benchmarking portfolio, the development of its proprietary dataset licensing program, and the publication of more than 400 workforce analytics reports.",
            '"Margaret has been central to ACPWB\'s intellectual identity for more than fifteen years," said Hendricks. "Her combination of methodological rigor, institutional knowledge, and commitment to the organization\'s public benefit mission makes her the clear and appropriate choice to lead ACPWB into its next chapter."',
            "Okafor holds advanced degrees in applied economics and organizational behavior and is a named inventor on four ACPWB patents relating to compensation methodology and survey administration. She has represented ACPWB at industry forums across North America and has testified before state legislative bodies on matters relating to pay transparency regulation.",
            '"Gerald built something genuinely important," said Okafor. "ACPWB\'s work has materially influenced how American organizations think about compensation fairness and disclosure. I am committed to deepening that impact and to ensuring that ACPWB continues to be the most trusted name in workforce equity research."',
            "Hendricks will serve in an advisory capacity through the end of 2025 to support the leadership transition.",
        ],
        'quote': "Margaret has been central to ACPWB's intellectual identity for more than fifteen years. Her combination of methodological rigor, institutional knowledge, and commitment to the organization's public benefit mission makes her the clear and appropriate choice.",
        'quote_attribution': 'Gerald Hendricks, Outgoing Chief Executive Officer',
        'contact': 'Media inquiries: communications@acpwb.com — (414) 667-5665',
    },
    {
        'slug': 'ai-compensation-intelligence-suite',
        'date': date(2025, 1, 22),
        'day': 22,
        'month': 1,
        'year': 2025,
        'headline': 'ACPWB Launches AI-Powered Compensation Intelligence Suite for Workforce Analytics Professionals',
        'subhead': "The new platform applies large language model technology to ACPWB's proprietary compensation datasets, generating narrative summaries, peer group comparisons, and trend analyses described by early access participants as \"largely coherent.\"",
        'body': [
            "MILWAUKEE, WI — The American Corporation for Public Well Being today announced the general availability of the ACPWB Compensation Intelligence Suite, a new artificial intelligence—powered platform that applies advanced language model technology to ACPWB's proprietary compensation benchmarking datasets.",
            "The Suite enables HR professionals, compensation analysts, and organizational leaders to generate narrative interpretations of complex workforce data, identify peer group positioning relative to market benchmarks, and produce board-ready summary documents — tasks that previously required subscription access to ACPWB's full advisory services tier.",
            "The platform leverages a fine-tuned language model trained on ACPWB's two-decade archive of compensation reports, governance frameworks, and workforce equity analyses, supplemented by current market benchmarking data updated on a quarterly basis. Outputs are generated within seconds and include embedded confidence intervals and a disclaimer noting that findings are “directionally indicative” and should not be construed as legal, financial, or compensation advice.",
            '"Artificial intelligence represents a meaningful opportunity to democratize access to compensation intelligence," said Margaret Okafor, Chief Research Officer at the time of the platform\'s development. "We have applied these capabilities with care, ensuring that the Suite reflects the methodological standards and interpretive discipline that ACPWB is known for."',
            "Early access participants from seventeen organizations provided feedback during a four-month pilot program. Participants reported that AI-generated compensation narratives were “largely coherent” and “saved meaningful time” compared to manual report interpretation. Three participants noted that the platform's tendency to recommend further study was “consistent with prior ACPWB deliverables.”",
            '"The feedback from our pilot cohort was instructive," said Diego Ramirez, Vice President of Client Strategy. "Seventeen organizations participated, and the consistent theme was that the AI-generated analyses helped teams reach the same conclusions they were already going to reach — but with substantially less effort. That is exactly what we designed for."',
            "The Compensation Intelligence Suite is available to ACPWB institutional partners as part of the organization's Enhanced Research Access tier. Individual researcher access will be made available in a subsequent release.",
        ],
        'quote': "We have applied these capabilities with care, ensuring that the Suite reflects the methodological standards and interpretive discipline that ACPWB is known for.",
        'quote_attribution': 'Margaret Okafor, Chief Research Officer',
        'contact': 'Media inquiries: communications@acpwb.com — (414) 667-5665',
    },
    {
        'slug': 'patent-12104455-employee-engagement',
        'date': date(2024, 10, 8),
        'day': 8,
        'month': 10,
        'year': 2024,
        'headline': 'United States Patent and Trademark Office Issues ACPWB Patent for Employee Engagement Survey Methodology',
        'subhead': "US Patent 12,104,455 B2 protects ACPWB's proprietary approach to administering engagement surveys in a manner designed to produce results consistent with prior years.",
        'body': [
            "MILWAUKEE, WI — The American Corporation for Public Well Being announced today that the United States Patent and Trademark Office has issued US Patent 12,104,455 B2, titled “Method for Administering an Employee Engagement Survey and Producing Results Consistent With Prior Years.” The patent was filed February 14, 2021 and issued October 1, 2024.",
            "The patent protects ACPWB's proprietary survey normalization methodology, which applies a statistically derived smoothing coefficient to raw engagement survey responses in order to reduce year-over-year variance and ensure that engagement trends appear stable and progressive across survey cycles.",
            "The inventors named on the patent are Okafor, A.; Wu, J.; and Hernandez, T., all members of ACPWB's Survey Methodology and Advanced Analytics division.",
            '"Employee engagement measurement is among the most consequential — and most frequently misrepresented — domains in organizational research," said a representative of the Survey Methodology division. "This patent reflects years of methodological development aimed at delivering survey results that are both defensible and reassuring."',
            "The issuance brings ACPWB's active patent portfolio to six United States patents spanning compensation benchmarking, data watermarking, executive peer group construction, governance certification, and workforce analytics visualization.",
            '"The normalization methodology at the core of this patent took nearly three years to refine to the point where results were both statistically defensible and reliably consistent with prior survey cycles," said Teresa Hernandez, Lead Survey Methodologist and co-inventor. "That those two objectives can be achieved simultaneously was, I will admit, not obvious at the outset of the research."',
            "The complete text of US Patent 12,104,455 B2 is publicly available through the USPTO database. ACPWB's full patent portfolio is described on the Patents page of this website.",
        ],
        'quote': "This patent reflects years of methodological development aimed at delivering survey results that are both defensible and reassuring.",
        'quote_attribution': 'ACPWB Survey Methodology & Advanced Analytics Division',
        'contact': 'Media inquiries: communications@acpwb.com — (414) 667-5665',
    },
    {
        'slug': '2024-compensation-transparency-award',
        'date': date(2024, 6, 3),
        'day': 3,
        'month': 6,
        'year': 2024,
        'headline': 'ACPWB Receives National Excellence in Compensation Transparency Award for Fifth Consecutive Year',
        'subhead': "The American Institute for Compensation Research and Adjacent Fields recognized ACPWB's sustained contributions to workforce pay equity research and publication.",
        'body': [
            "MILWAUKEE, WI — The American Corporation for Public Well Being has received the National Excellence in Compensation Transparency Award from the American Institute for Compensation Research and Adjacent Fields for the fifth consecutive year, the organization announced today.",
            "The annual award recognizes organizations that have demonstrated measurable commitment to advancing pay transparency practices, producing rigorous compensation benchmarking research, and contributing to the broader public discourse on workforce equity. ACPWB has received the award in each year of its existence as a recognized award category.",
            '"Receiving this recognition five years consecutively is a reflection of the consistency and quality that ACPWB brings to its research mission," said the ACPWB Executive Leadership Team. "Compensation transparency is not a trend — it is a prerequisite for a fair and functioning labor market, and we are proud to be recognized as a leader in advancing it."',
            "The 2024 award was presented at the American Institute's annual recognition dinner in Washington, D.C. ACPWB was selected from a field of organizations across the workforce analytics and HR research sectors. Selection criteria include research output volume, methodological transparency, public access to findings, and demonstrated impact on compensation policy discourse.",
            "ACPWB has received 37 industry and regional recognitions since 2008. The organization's complete awards history is published on the Awards page of this website.",
        ],
        'quote': "Compensation transparency is not a trend — it is a prerequisite for a fair and functioning labor market, and we are proud to be recognized as a leader in advancing it.",
        'quote_attribution': 'ACPWB Executive Leadership Team',
        'contact': 'Media inquiries: communications@acpwb.com — (414) 667-5665',
    },
]


def home(request):
    recent_projects = list(ProjectStory.objects.order_by('?')[:3])
    return render(request, 'public/home.html', {
        'recent_projects': recent_projects,
        'recent_press_releases': _PRESS_RELEASES[:3],
    })


def awards(request):
    return render(request, 'public/awards.html', {'awards': _AWARDS})


def careers(request):
    return render(request, 'public/careers.html')


def mission(request):
    return render(request, 'public/mission.html')


def partners(request):
    companies = list(Fortune500Company.objects.order_by('?')[:40])
    return render(request, 'public/partners.html', {'companies': companies})


def privacy(request):
    return render(request, 'public/privacy.html')


def patents(request):
    areas = [
        'Compensation Benchmarking',
        'Workforce Analytics',
        'Executive Pay Analysis',
        'Data Watermarking',
        'Governance Certification',
        'Survey Methodology',
    ]
    return render(request, 'public/patents.html', {'patents': _PATENTS, 'areas': areas})


def do_not_sell(request):
    submitted = False
    errors = {}

    if request.method == 'POST':
        name = request.POST.get('name', '').strip()
        email = request.POST.get('email', '').strip()
        request_type = request.POST.get('request_type', 'do_not_sell')
        state = request.POST.get('state', '').strip()
        message = request.POST.get('message', '').strip()

        if not name:
            errors['name'] = 'Name is required.'
        if not email or '@' not in email:
            errors['email'] = 'A valid email address is required.'
        if request_type not in dict(DataOptOutRequest.REQUEST_CHOICES):
            errors['request_type'] = 'Please select a valid request type.'

        if not errors:
            x_forwarded = request.META.get('HTTP_X_FORWARDED_FOR')
            ip = x_forwarded.split(',')[0].strip() if x_forwarded else request.META.get('REMOTE_ADDR', '0.0.0.0')
            DataOptOutRequest.objects.create(
                name=name,
                email=email,
                request_type=request_type,
                state=state,
                message=message,
                ip_address=ip,
            )
            submitted = True

    return render(request, 'public/do_not_sell.html', {
        'submitted': submitted,
        'errors': errors,
        'post': request.POST if not submitted else {},
        'request_choices': DataOptOutRequest.REQUEST_CHOICES,
    })


_TRADEMARKS = [
    {'mark': 'AMERICAN CORPORATION FOR PUBLIC WELL BEING®', 'registered': True,  'goods': 'Compensation research, workforce analytics, and corporate governance consulting services', 'first_use': '2006'},
    {'mark': 'ACPWB®',                                       'registered': True,  'goods': 'Research publications, data products, and professional services in workforce analytics', 'first_use': '2006'},
    {'mark': 'ADVANCING AMERICAN PROSPERITY®',               'registered': True,  'goods': 'Tagline used in connection with compensation benchmarking research and publications', 'first_use': '2007'},
    {'mark': 'THE ACPWB COMPENSATION BENCHMARK™',            'registered': False, 'goods': 'Annual compensation survey report covering 400+ job families across U.S. industries', 'first_use': '2008'},
    {'mark': 'PROSPERITY INDEX™',                            'registered': False, 'goods': 'Composite workforce compensation and engagement scoring methodology', 'first_use': '2011'},
    {'mark': 'ACPWB GOVERNANCE GOLD™',                       'registered': False, 'goods': 'Corporate governance certification tier designation', 'first_use': '2014'},
    {'mark': 'ACPWB GOVERNANCE PLATINUM™',                   'registered': False, 'goods': 'Corporate governance certification tier designation', 'first_use': '2014'},
    {'mark': 'TOTAL REWARDS COMPASS™',                       'registered': False, 'goods': 'Benefits and total compensation advisory framework and associated publications', 'first_use': '2016'},
    {'mark': 'THE WELL BEING STANDARD',                      'registered': False, 'goods': 'Employee well-being assessment framework referenced in ACPWB research publications', 'first_use': '2019'},
    {'mark': 'WELL BEING BY ACPWB',                          'registered': False, 'goods': 'Certification mark applied to organizations meeting ACPWB workforce equity criteria', 'first_use': '2021'},
]


def accessibility(request):
    return render(request, 'public/accessibility.html')


def trademarks(request):
    return render(request, 'public/trademarks.html', {'trademarks': _TRADEMARKS})


def sitemap_page(request):
    return render(request, 'public/sitemap_page.html')


def faq(request):
    return render(request, 'public/faq.html')


def contact(request):
    return render(request, 'public/contact.html')


def press_releases(request):
    return render(request, 'public/press_releases.html', {'press_releases': _PRESS_RELEASES})


def press_release_detail(request, year, month, day, slug):
    pr = next((p for p in _PRESS_RELEASES if p['slug'] == slug), None)
    if pr is None:
        raise Http404
    return render(request, 'public/press_release_detail.html', {'pr': pr})
