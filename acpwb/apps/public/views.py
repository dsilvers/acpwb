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


def home(request):
    recent_projects = list(ProjectStory.objects.order_by('?')[:3])
    return render(request, 'public/home.html', {'recent_projects': recent_projects})


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
