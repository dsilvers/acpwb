"""
Deterministic report generator for ACPWB Reports & Publications.
All output is seeded from the slug — same slug always returns same content.
"""
import csv
import hashlib
import io
import random
import re
from datetime import date

from django.utils.text import slugify


def _rng_from_seed(seed_str):
    """Deterministic random.Random from a string seed (same as projects/generators.py)."""
    hex_digest = hashlib.md5(seed_str.encode()).hexdigest()
    return random.Random(int(hex_digest, 16))


# ── Year distribution (weighted toward recent, goes back to 1993) ─────────────

YEAR_POOL = (
    list(range(1993, 2000)) * 1 +   # rare early archive
    list(range(2000, 2008)) * 2 +   # growing cadence
    list(range(2008, 2016)) * 3 +   # established org
    list(range(2016, 2023)) * 5 +   # prolific period
    list(range(2023, 2026)) * 8     # current era
)


# ── Canonical catalog ─────────────────────────────────────────────────────────

REPORT_CATALOG = [
    # 2024 flagship reports
    {'slug': 'salary-compensation-benchmarking-survey-2024',      'title': 'Salary & Compensation Benchmarking Survey 2024',        'category': 'Compensation',          'file_type': 'csv'},
    {'slug': 'annual-workforce-analytics-report-2024',            'title': 'Annual Workforce Analytics Report 2024',                'category': 'Workforce',             'file_type': 'pdf'},
    {'slug': 'executive-compensation-study-2024',                 'title': 'Executive Compensation Study 2024',                     'category': 'Compensation',          'file_type': 'csv'},
    {'slug': 'corporate-governance-best-practices-white-paper',   'title': 'Corporate Governance Best Practices White Paper',       'category': 'Governance',            'file_type': 'pdf'},
    {'slug': 'esg-disclosure-framework-2024',                     'title': 'ESG Disclosure Framework 2024',                         'category': 'ESG',                   'file_type': 'pdf'},
    {'slug': 'q3-economic-indicators-2024',                       'title': 'Q3 Economic Indicators Report 2024',                    'category': 'Economic',              'file_type': 'csv'},
    {'slug': 'benefits-cost-analysis-2024',                       'title': 'Benefits Cost Analysis 2024',                           'category': 'Benefits',              'file_type': 'csv'},
    {'slug': 'industry-wage-gap-analysis-2024',                   'title': 'Industry Wage Gap Analysis 2024',                       'category': 'Compensation',          'file_type': 'pdf'},
    {'slug': 'fortune-500-ceo-pay-ratio-analysis',                'title': 'Fortune 500 CEO Pay Ratio Analysis',                    'category': 'Compensation',          'file_type': 'csv'},
    {'slug': 'state-of-workplace-report-2024',                    'title': 'State of the Workplace Report 2024',                    'category': 'Workforce',             'file_type': 'pdf'},
    {'slug': 'regulatory-compliance-costs-2024',                  'title': 'Regulatory Compliance Costs Report 2024',               'category': 'Governance',            'file_type': 'pdf'},
    {'slug': 'employee-satisfaction-survey-results-2024',         'title': 'Employee Satisfaction Survey Results 2024',             'category': 'Workforce',             'file_type': 'csv'},
    {'slug': 'regional-salary-differentials-2024',                'title': 'Regional Salary Differentials Study 2024',              'category': 'Compensation',          'file_type': 'csv'},
    {'slug': 'board-composition-study-2024',                      'title': 'Board Composition Study 2024',                          'category': 'Governance',            'file_type': 'csv'},
    {'slug': 'retirement-readiness-index-2024',                   'title': 'Retirement Readiness Index 2024',                       'category': 'Benefits',              'file_type': 'csv'},
    {'slug': 'public-sector-compensation-guide',                  'title': 'Public Sector Compensation Guide',                      'category': 'Compensation',          'file_type': 'pdf'},
    {'slug': 'healthcare-benefits-landscape-2024',                'title': 'Healthcare Benefits Landscape 2024',                    'category': 'Benefits',              'file_type': 'pdf'},
    {'slug': 'remote-work-impact-study-2024',                     'title': 'Remote Work Impact Study 2024',                         'category': 'Workforce',             'file_type': 'csv'},
    {'slug': 'pay-transparency-compliance-report-2024',           'title': 'Pay Transparency Compliance Report 2024',               'category': 'Compliance',            'file_type': 'pdf'},
    {'slug': 'dei-workforce-metrics-2024',                        'title': 'DEI Workforce Metrics Survey 2024',                     'category': 'Diversity & Inclusion', 'file_type': 'csv'},
    {'slug': 'technology-sector-compensation-analysis-2024',      'title': 'Technology Sector Compensation Analysis 2024',          'category': 'Technology',           'file_type': 'csv'},
    {'slug': 'leadership-pipeline-assessment-2024',               'title': 'Leadership Pipeline Assessment 2024',                   'category': 'Leadership',            'file_type': 'pdf'},
    {'slug': 'talent-acquisition-cost-index-2024',                'title': 'Talent Acquisition Cost Index 2024',                    'category': 'Talent Acquisition',    'file_type': 'csv'},
    {'slug': 'employee-retention-drivers-study-2024',             'title': 'Employee Retention Drivers Study 2024',                 'category': 'Retention',             'file_type': 'pdf'},
    {'slug': 'total-rewards-effectiveness-survey-2024',           'title': 'Total Rewards Effectiveness Survey 2024',               'category': 'Compensation',          'file_type': 'csv'},
    {'slug': 'q1-compensation-pulse-survey-2024',                 'title': 'Q1 Compensation Pulse Survey 2024',                     'category': 'Compensation',          'file_type': 'csv'},
    # 2023
    {'slug': 'salary-compensation-benchmarking-survey-2023',      'title': 'Salary & Compensation Benchmarking Survey 2023',        'category': 'Compensation',          'file_type': 'csv'},
    {'slug': 'annual-workforce-analytics-report-2023',            'title': 'Annual Workforce Analytics Report 2023',                'category': 'Workforce',             'file_type': 'pdf'},
    {'slug': 'executive-compensation-study-2023',                 'title': 'Executive Compensation Study 2023',                     'category': 'Compensation',          'file_type': 'csv'},
    {'slug': 'esg-maturity-benchmarks-2023',                      'title': 'ESG Maturity Benchmarks 2023',                          'category': 'ESG',                   'file_type': 'pdf'},
    {'slug': 'benefits-cost-trend-report-2023',                   'title': 'Benefits Cost Trend Report 2023',                       'category': 'Benefits',              'file_type': 'pdf'},
    {'slug': 'hybrid-work-compensation-study-2023',               'title': 'Hybrid Work Compensation Study 2023',                   'category': 'Workforce',             'file_type': 'csv'},
    {'slug': 'board-compensation-practices-2023',                 'title': 'Board Compensation Practices Report 2023',              'category': 'Governance',            'file_type': 'pdf'},
    {'slug': 'gender-pay-equity-audit-2023',                      'title': 'Gender Pay Equity Audit 2023',                          'category': 'Diversity & Inclusion', 'file_type': 'csv'},
    # 2022
    {'slug': 'salary-compensation-benchmarking-survey-2022',      'title': 'Salary & Compensation Benchmarking Survey 2022',        'category': 'Compensation',          'file_type': 'csv'},
    {'slug': 'workforce-resilience-report-2022',                  'title': 'Workforce Resilience Report 2022',                      'category': 'Workforce',             'file_type': 'pdf'},
    {'slug': 'executive-pay-ratio-study-2022',                    'title': 'Executive Pay Ratio Study 2022',                        'category': 'Compensation',          'file_type': 'csv'},
    {'slug': 'great-resignation-data-2022',                       'title': 'The Great Resignation: ACPWB Data Study 2022',          'category': 'Retention',             'file_type': 'pdf'},
    {'slug': 'benefits-benchmarking-survey-2022',                 'title': 'Benefits Benchmarking Survey 2022',                     'category': 'Benefits',              'file_type': 'csv'},
    # 2020–2021
    {'slug': 'pandemic-workforce-impact-study-2021',              'title': 'Pandemic Workforce Impact Study 2021',                  'category': 'Workforce',             'file_type': 'pdf'},
    {'slug': 'remote-work-compensation-framework-2020',           'title': 'Remote Work Compensation Framework 2020',               'category': 'Compensation',          'file_type': 'pdf'},
    {'slug': 'crisis-total-rewards-survey-2020',                  'title': 'Crisis-Period Total Rewards Survey 2020',               'category': 'Compensation',          'file_type': 'csv'},
    # 2016–2019
    {'slug': 'pay-equity-action-guide-2019',                      'title': 'Pay Equity Action Guide 2019',                          'category': 'Compensation',          'file_type': 'pdf'},
    {'slug': 'annual-hr-benchmarking-report-2018',                'title': 'Annual HR Benchmarking Report 2018',                    'category': 'Workforce',             'file_type': 'pdf'},
    {'slug': 'executive-compensation-governance-2017',            'title': 'Executive Compensation Governance Study 2017',          'category': 'Governance',            'file_type': 'pdf'},
    {'slug': 'compensation-trends-midmarket-2016',                'title': 'Compensation Trends: Mid-Market Employers 2016',        'category': 'Compensation',          'file_type': 'csv'},
    # Historical anchors
    {'slug': 'national-salary-survey-2001',                       'title': 'National Salary Survey 2001',                           'category': 'Compensation',          'file_type': 'pdf'},
    {'slug': 'corporate-pay-equity-study-1998',                   'title': 'Corporate Pay Equity Study 1998',                       'category': 'Compensation',          'file_type': 'pdf'},
    {'slug': 'workforce-demographics-report-2006',                'title': 'Workforce Demographics Report 2006',                    'category': 'Workforce',             'file_type': 'pdf'},
    {'slug': 'executive-benefits-survey-2003',                    'title': 'Executive Benefits Survey 2003',                        'category': 'Benefits',              'file_type': 'csv'},
    {'slug': 'governance-disclosure-standards-2009',              'title': 'Governance Disclosure Standards 2009',                  'category': 'Governance',            'file_type': 'pdf'},
    {'slug': 'compensation-benchmarking-methodology-1995',        'title': 'Compensation Benchmarking Methodology 1995',            'category': 'Compensation',          'file_type': 'pdf'},
    {'slug': 'annual-hr-metrics-report-2011',                     'title': 'Annual HR Metrics Report 2011',                         'category': 'Workforce',             'file_type': 'csv'},
    {'slug': 'job-grade-architecture-white-paper-2007',           'title': 'Job Grade Architecture White Paper 2007',               'category': 'Compensation',          'file_type': 'pdf'},
    {'slug': 'total-compensation-survey-2014',                    'title': 'Total Compensation Survey 2014',                        'category': 'Compensation',          'file_type': 'csv'},
    {'slug': 'board-effectiveness-study-2012',                    'title': 'Board Effectiveness Study 2012',                        'category': 'Governance',            'file_type': 'pdf'},
    {'slug': 'healthcare-cost-benchmarking-2010',                 'title': 'Healthcare Cost Benchmarking Study 2010',               'category': 'Benefits',              'file_type': 'csv'},
    {'slug': 'early-career-compensation-guide-2005',              'title': 'Early Career Compensation Guide 2005',                  'category': 'Compensation',          'file_type': 'pdf'},
    {'slug': 'stock-option-design-survey-1999',                   'title': 'Stock Option Design Survey 1999',                       'category': 'Compensation',          'file_type': 'pdf'},
    {'slug': 'nonprofit-sector-pay-study-2008',                   'title': 'Nonprofit Sector Pay Study 2008',                       'category': 'Compensation',          'file_type': 'pdf'},
    {'slug': 'regional-labor-market-analysis-2004',               'title': 'Regional Labor Market Analysis 2004',                   'category': 'Workforce',             'file_type': 'csv'},
    {'slug': 'acpwb-founding-compensation-principles-1996',       'title': 'ACPWB Founding Compensation Principles 1996',           'category': 'Compensation',          'file_type': 'pdf'},
]

REPORT_CATEGORIES = [
    'Compensation', 'Workforce', 'Governance', 'ESG', 'Benefits',
    'Economic', 'Compliance', 'Diversity & Inclusion', 'Technology',
    'Leadership', 'Talent Acquisition', 'Retention',
]

REPORT_ADJECTIVES = [
    'Annual', 'Quarterly', 'Biennial', 'Longitudinal', 'Comprehensive',
    'National', 'Regional', 'Cross-Industry', 'Mid-Year', 'Benchmark',
    'Baseline', 'Trend', 'Comparative', 'Strategic', 'Preliminary',
    'Final', 'Supplemental', 'Revised', 'Expanded', 'Integrated',
    'Inaugural', 'Landmark', 'Definitive', 'Sector-Wide', 'Rapid-Cycle',
    'Pilot', 'Multi-Year', 'Retrospective', 'Prospective', 'Rolling',
    'Deep-Dive', 'Exploratory', 'Validation', 'Cross-Sectional', 'Syndicated',
    'Independent', 'Peer-Reviewed', 'Practitioner-Focused', 'Evidence-Based',
]

REPORT_SUBJECTS = [
    'Compensation Benchmarking', 'Pay Equity', 'Workforce Planning',
    'Talent Retention', 'Benefits Utilization', 'Salary Transparency',
    'Board Effectiveness', 'Executive Pay', 'Compliance Risk',
    'Diversity Metrics', 'Turnover Analysis', 'Skills Gap Assessment',
    'Remote Work Productivity', 'Healthcare Cost Trends',
    'Retirement Plan Participation', 'Incentive Compensation',
    'Total Rewards Strategy', 'CEO Pay Ratio', 'Gender Pay Gap',
    'Geographic Pay Differentials', 'Wage Band Calibration',
    'Job Architecture Review', 'Market Pricing Methodology',
    'Employee Engagement Index', 'Hiring Cost Analysis',
    'Pay Transparency Compliance', 'Workforce Demographics',
    'Manager Effectiveness', 'Organizational Design', 'Succession Planning',
    'Learning & Development ROI', 'Contingent Workforce Costs',
    'Internal Mobility Trends', 'Onboarding Effectiveness',
    'Mental Health Benefits Adoption', 'Student Loan Repayment Benefits',
    'Equity Award Design', 'Long-Term Incentive Plan Prevalence',
    'Sales Compensation Effectiveness', 'Non-Compete Agreement Trends',
    'Pay Range Transparency', 'Job Evaluation Methodology',
    'Workforce Age Demographics', 'Generational Pay Patterns',
    'Profit-Sharing Plan Design', 'Variable Pay Program Analysis',
    'Compensation Governance Practices', 'Peer Group Selection Methodology',
    'Talent Acquisition Cost', 'Time-to-Hire Benchmarks', 'Quality of Hire Metrics',
    'Offer Acceptance Rates', 'Candidate Experience Scores', 'Recruiter Productivity',
    'Workforce Composition Analysis', 'Headcount Planning Practices',
    'Attrition Driver Analysis', 'Exit Interview Data Synthesis',
    'Stay Interview Outcomes', 'Boomerang Employee Trends',
    'Employee Net Promoter Score', 'Organizational Health Index',
    'Culture Survey Benchmarks', 'Psychological Safety Metrics',
    'Inclusion Index Trends', 'Belonging Score Analysis',
    'Pay Equity Remediation', 'Race Pay Gap Analysis', 'Disability Pay Analysis',
    'Board Compensation Practices', 'Director Pay Benchmarks',
    'Compensation Committee Governance', 'Say-on-Pay Vote Analysis',
    'Proxy Advisory Benchmarks', 'Shareholder Engagement Outcomes',
    'ESG Compensation Integration', 'Sustainability-Linked Pay',
    'Annual Incentive Plan Design', 'Performance Share Unit Prevalence',
    'Restricted Stock Unit Vesting Practices', 'Stock Option Prevalence',
    'Equity Plan Dilution Benchmarks', 'Clawback Policy Practices',
    'Deferred Compensation Plan Design', 'SERP Prevalence',
    'Benefits Cost Benchmarking', 'Medical Plan Cost-Sharing Trends',
    'Dental and Vision Benefit Practices', 'Wellness Program ROI',
    'Financial Wellness Benefit Adoption', 'Childcare Benefit Trends',
    'Paid Leave Policy Benchmarks', 'Parental Leave Practices',
    'Flexible Work Arrangement Policy', 'Four-Day Workweek Pilot Data',
    'HR Technology Adoption', 'Compensation Software Benchmarks',
    'People Analytics Maturity', 'Workforce Planning Technology',
    'AI in HR Decision-Making', 'Chatbot and Automation in HR',
    'Skills Taxonomy Development', 'Internal Talent Marketplace Adoption',
    'Career Framework Design', 'Job Level Harmonization',
    'Global Compensation Strategy', 'Expatriate Pay Practices',
    'International Benefits Benchmarks', 'Cross-Border Pay Equity',
]

REPORT_SUFFIXES = [
    'Report', 'Study', 'Survey', 'Analysis', 'Index', 'Guide',
    'Assessment', 'Framework', 'White Paper', 'Briefing',
    'Review', 'Outlook', 'Digest', 'Compendium', 'Data Set',
    'Pulse Survey', 'Field Report', 'Benchmark Report', 'Monograph',
]

SUMMARY_TEMPLATES = [
    "Drawing on proprietary survey data collected from {n} organizations across {states} states, this {adj} {subject} {suffix} documents {finding}. The analysis incorporates {years} years of longitudinal data and has been peer-reviewed by ACPWB's Research Advisory Panel.",
    "This {adj} {subject} {suffix} presents findings from ACPWB's national survey of {n} compensation professionals. Key findings include {finding}, with significant variation observed across industry sectors and geographic regions.",
    "Prepared by the ACPWB Research Division in collaboration with {n} contributing organizations, this {suffix} examines {finding} and provides actionable benchmarks for HR and finance professionals. Detailed methodology is available in Appendix B.",
    "An examination of {subject} trends across {n} participating organizations, this {suffix} draws on data collected between Q{q1} {year_from} and Q{q2} {year_to}. Findings are presented in aggregate to preserve respondent confidentiality.",
    "ACPWB's {adj} {suffix} synthesizes responses from {n} HR and finance leaders on the state of {subject} in American enterprise. The data reveals {finding}, underscoring the need for standardized benchmarking practices.",
    "Released by the ACPWB Research Division, this {adj} {suffix} on {subject} draws from a nationally representative sample of {n} employers. The report provides granular breakdowns by industry, geography, and company size.",
    "This {suffix} represents the {adj} installment of ACPWB's flagship {subject} research program, now in its {years}th year. Data was sourced from {n} voluntary respondents spanning all major industry sectors and organizational size bands.",
    "Commissioned by ACPWB's Research Advisory Panel, this {adj} {suffix} provides the most current data available on {subject}. With {n} respondents from {states} states, it is among the largest studies of its kind conducted in the United States.",
    "ACPWB's research into {subject} has produced this {adj} {suffix}, drawing on data contributed by {n} practitioners across industries. The findings have been validated against public datasets from the Bureau of Labor Statistics and the Department of Labor.",
    "The {adj} {subject} {suffix} is designed for HR and finance professionals who need reliable, current data on {finding}. The ACPWB Research Division updates this study on an {adj} basis to ensure continued relevance as labor market conditions evolve.",
    "For the {years}th consecutive year, ACPWB has published this {suffix} on {subject}. Participation from {n} organizations across {states} states makes this one of the most geographically representative studies in the field.",
    "Organizations navigating {subject} decisions require reliable external benchmarks. This {adj} {suffix}, produced by the ACPWB Research Division from a sample of {n} participating employers, provides those benchmarks with a level of precision and transparency unmatched in the field.",
    "The ACPWB Research Division's {adj} {suffix} on {subject} captures responses from {n} HR, finance, and legal professionals across {states} states. This edition highlights {finding}, a trend that has emerged consistently across three consecutive measurement cycles.",
    "ACPWB's {years}-year longitudinal study on {subject} continues with this {adj} {suffix}, which tracks outcomes for the same cohort of {n} organizations over time. Consistent with prior waves, the data confirms {finding} as a persistent and statistically significant pattern.",
    "This {adj} {suffix} is the product of ACPWB's annual collaboration with {n} corporate practitioners who contribute compensation and workforce data in exchange for customized benchmarking access. The {subject} findings presented here reflect the aggregate of that contribution.",
    "Produced in response to member demand for more granular {subject} data, this {adj} {suffix} covers {n} organizational respondents across {states} states and provides the first-ever breakdowns by revenue band, workforce size, and primary geographic market.",
    "The {year_to} edition of ACPWB's {subject} {suffix} reflects a {adj} expansion of the prior-year study, incorporating {n} new respondents and extending coverage to {states} additional states. Key new findings include {finding}.",
    "ACPWB's Research Division undertook this {adj} study of {subject} in response to a gap in publicly available benchmarking data identified by our member organizations. The {suffix} is based on {n} survey responses collected between Q{q1} {year_from} and Q{q2} {year_to}.",
    "With {n} respondents representing a combined workforce of more than two million employees across {states} states, this {adj} {subject} {suffix} is among the most statistically powerful studies ACPWB has published in its {years}-year history.",
    "This {suffix} presents a {adj} analysis of {subject} practices at {n} organizations, segmented by industry sector, company size, and geographic region. Highlights include {finding}, with detailed breakdowns available in the accompanying data appendix.",
    "Following three years of development, ACPWB's {adj} {subject} {suffix} delivers the most comprehensive examination of {finding} yet produced for HR practitioners. The study draws on {n} survey responses and incorporates qualitative insights from {states} executive focus groups.",
    "ACPWB's {subject} {suffix} provides {adj} benchmarks derived from data contributed by {n} organizations. This edition introduces a new scoring methodology that enables direct comparison across years — a capability our members have requested since the study's inaugural edition {years} years ago.",
]

FINDING_PHRASES = [
    "median base salary increases of {p}% across all job families",
    "a {p}% year-over-year rise in total compensation costs",
    "widening pay disparity between executive and non-executive roles",
    "accelerating adoption of performance-based incentive structures",
    "significant geographic variation in compensation for comparable roles",
    "declining employee satisfaction with current benefits packages",
    "increased regulatory scrutiny of pay practices across sectors",
    "a {p}% correlation between salary transparency and employee retention",
    "median total rewards gaps of {p}% between demographic cohorts",
    "growing adoption of pay band disclosure among Fortune 500 firms",
    "a {p}-point decline in employee confidence in compensation fairness",
    "rising benefits costs outpacing base salary growth by {p} percentage points",
    "a {p}% increase in organizations adopting formal pay equity review cycles",
    "accelerating convergence of remote and in-office compensation levels",
    "a {p}% drop in incentive compensation effectiveness ratings",
    "growing use of artificial intelligence in compensation benchmarking workflows",
    "a {p}% increase in organizations disclosing pay ranges in job postings",
    "sustained upward pressure on starting salaries for entry-level roles",
    "a {p}% rise in employee utilization of mental health benefit offerings",
    "declining confidence in traditional annual merit review cycles",
    "increased regulatory pressure around pay equity audit documentation",
    "a {p}% expansion in organizations offering student loan repayment benefits",
    "greater divergence between large-employer and small-employer pay practices",
    "a {p}% increase in the prevalence of spot-bonus and recognition programs",
    "a {p}% reduction in median time-to-fill for critical positions",
    "continued compression of pay ranges at professional grade levels",
    "a {p}% increase in CEO-to-median-worker pay ratios across surveyed firms",
    "growing prevalence of multi-year long-term incentive plan designs",
    "a {p}% decline in employee participation in voluntary benefits elections",
    "accelerating adoption of continuous performance management frameworks",
    "a {p}% increase in organizations with formal compensation philosophy documents",
    "widening gender pay gaps at senior director and above grade levels",
    "a {p}% rise in share-based compensation as a percentage of total pay",
    "increasing reliance on third-party compensation survey data for benchmarking",
    "a {p}% increase in organizations conducting annual pay equity audits",
    "declining effectiveness of annual merit cycles in driving retention",
    "a {p}% expansion in organizations offering childcare subsidy benefits",
    "sustained growth in pay transparency legislation at the state and local level",
    "a {p}-point increase in employee net promoter scores following compensation reform",
    "greater integration of total compensation data into recruiting conversations",
    "a {p}% increase in organizations using predictive analytics for attrition risk",
    "accelerating adoption of pay range midpoint management disciplines",
    "a {p}% rise in cash signing bonus prevalence for professional-level hires",
    "continued growth in four-day workweek pilot programs among surveyed employers",
    "a {p}% increase in organizations benchmarking benefits against peer groups annually",
    "growing concern among employees regarding retirement income adequacy",
    "a {p}% increase in organizations with formal DEI-linked incentive metrics",
    "rising adoption of skills-based pay frameworks in technology-intensive sectors",
    "a {p}% decline in organizations relying solely on traditional job evaluation methods",
]

DOCUMENT_PARAGRAPH_TEMPLATES = [
    "Based on {n} survey responses collected during {quarter} {year}, ACPWB's analysis reveals a median of ${val:,} — representing a {dir} of {p}% compared to the prior-year period. These findings are consistent with broader macroeconomic trends documented by the Bureau of Labor Statistics and peer research organizations.",
    "The {industry} sector demonstrated notable outcomes in the measured period, with the {percentile}th percentile of respondents reporting total compensation of ${val:,}. Geographic dispersion remains pronounced, with respondents in high-cost-of-living markets earning {p}% more than peers in comparable roles in lower-cost regions.",
    "Regression analysis controlling for years of experience, company size, and geographic market indicates that role-level classification accounts for approximately {p}% of variance in total compensation. This finding is consistent across all industry segments examined in this study.",
    "ACPWB's proprietary index for this metric — calculated using a weighting methodology refined over {years} years of continuous research — stands at {idx:.1f} as of {quarter} {year}, representing a movement of {pts} basis points from the prior measurement period.",
    "Respondents citing {factor} as a primary driver of compensation decisions increased by {p}% year-over-year, reflecting a broader shift in how organizations approach total rewards strategy. ACPWB anticipates this trend will continue through the next measurement cycle.",
    "This finding aligns with reporting from the WorldatWork Total Rewards Survey and the Mercer Compensation Planning Survey, both of which identified {factor} as a key determinant of compensation outcomes in the {industry} sector for this performance year.",
    "Organizations in the top quartile of {factor} implementation reported {p}% lower voluntary turnover than their peers, and ${val:,} lower average cost-per-hire. These findings suggest a meaningful return on investment for employers who prioritize systematic compensation strategy.",
    "Multivariate analysis of the {quarter} {year} dataset identified {factor} as the single strongest predictor of employee pay satisfaction, explaining {p}% of the observed variance. This relationship held across company size bands and geographic regions.",
    "Longitudinal comparison with ACPWB's prior {years} years of data reveals a consistent pattern: organizations that invest in {factor} outperform their sector peers on both retention and compensation competitiveness metrics. The {year} data reinforces this relationship with statistical significance at the 95% confidence level.",
    "The {industry} sector's median compensation for benchmark roles increased by {p}% in {year}, driven primarily by {factor}. This rate of increase is {p}% above the cross-industry average documented in this study, reflecting unique supply-demand dynamics within the sector.",
    "A subset analysis of {n} respondents who completed ACPWB's extended survey module on {factor} revealed that {p}% had implemented formal policies within the past 24 months — a figure that has nearly doubled since the {quarter} {year} measurement period.",
    "ACPWB's data shows that organizations with documented {factor} practices report median employee engagement scores {p} points higher than those without such practices. The economic value of this engagement differential — estimated at ${val:,} per employee per year — represents a compelling business case for investment.",
    "The {percentile}th percentile benchmark for this job family in the {industry} sector now stands at ${val:,} in total direct compensation, an increase of {p}% from the prior year. ACPWB recommends that organizations use this benchmark in conjunction with internal equity analysis before making compensation decisions.",
    "Cross-industry analysis reveals that {factor} practices have matured significantly since ACPWB first tracked this variable in {year}. What was once a differentiating practice among leading employers has become a baseline expectation, with {p}% of surveyed organizations now reporting formal programs.",
    "Geographic analysis of the {year} data reveals persistent and, in some cases, widening differentials. Roles in metropolitan markets with populations exceeding one million command a median premium of {p}% over comparable roles in smaller markets, after controlling for industry, company size, and role level.",
    "ACPWB's {industry} sector analysis confirms that {factor} is now a primary driver of compensation competitiveness, cited by {p}% of survey respondents as the single most important lever available to HR and finance leaders. Organizations that have invested in this capability report {p}% lower regrettable attrition.",
    "The {percentile}th-percentile benchmark for this role family now stands at ${val:,} in base salary, with total direct compensation reaching ${val:,} when annual incentives are included at target. These benchmarks represent a {p}% increase from the {year} baseline and are consistent with ACPWB's forward projection.",
    "Cluster analysis of the {n}-organization {year} dataset reveals four distinct compensation philosophy archetypes. Organizations in the 'market-leading' cluster outperform the overall median by {p}%, while those in the 'cost-constrained' cluster lag by a comparable margin. The {industry} sector is disproportionately represented in the market-leading group.",
    "The relationship between {factor} and employee retention has strengthened over ACPWB's {years}-year study horizon. The {year} data shows a correlation coefficient of {p}% — the highest recorded since the study's inception — suggesting that this practice has become a more decisive factor in employees' stay-or-leave decisions.",
    "Regression analysis controlling for company size, industry, and geographic market indicates that organizations with formal {factor} programs pay a median premium of {p}% for comparable roles compared to organizations without such programs. ACPWB interprets this premium as a market signal of the competitive value employees place on this practice.",
    "Year-over-year comparison of the {n}-organization panel shows that the {percentile}th percentile of total compensation increased by {p}% between {year} and the current measurement period, while the 10th-percentile benchmark increased by only {p}%, indicating growing dispersion in the distribution of compensation outcomes.",
    "A dedicated module added to the {year} survey instrument captured data on {factor} for the first time. Results show that {p}% of organizations have implemented formal practices in this area, with adoption highest among large employers and organizations in the {industry} sector.",
    "ACPWB's benefit valuation model estimates that a comprehensive benefits package is worth an average of ${val:,} per employee per year — representing {p}% of total direct compensation. This figure varies significantly by company size, with small employers averaging ${val:,} and large enterprises averaging ${val:,}.",
    "The {quarter} {year} pulse survey module captured sentiment data from {n} employees at participating organizations. Among respondents, {p}% expressed satisfaction with their total compensation package — a {p}-point decline from the prior measurement period — underscoring the growing importance of proactive communication around {factor}.",
    "Longitudinal analysis across {years} survey waves reveals that organizations with above-median {factor} investment consistently outperform their peers on retention, hiring efficiency, and employee satisfaction. The magnitude of this performance differential has grown by {p}% since {year}, suggesting that the competitive advantage associated with this practice is compounding over time.",
    "ACPWB's incentive compensation effectiveness score — a composite measure derived from {n} survey items — stands at {idx:.1f} for the {year} dataset, up {pts} basis points from the prior year. The {industry} sector records the highest average score among all sectors tracked, driven primarily by gains in {factor}.",
    "The {quarter} {year} regulatory compliance module of ACPWB's survey captured data on organizational responses to recent legislative developments affecting {factor}. Among respondents, {p}% have already implemented compliance programs, {p}% are in the process of doing so, and the remainder have not yet begun formal compliance planning.",
    "Data from {n} responding organizations indicates that the average HR technology investment per employee increased by {p}% in {year}, driven primarily by adoption of {factor}-related platforms. ACPWB anticipates continued acceleration of this trend as organizations seek to automate high-volume compensation and workforce analytics tasks.",
    "Benchmarking against ACPWB's prior {years} years of data reveals that the current environment — characterized by {factor} — represents the most significant structural shift in compensation practices recorded since the study began. Organizations that adapt quickly are positioned to outperform peers on talent acquisition and retention over the next three to five years.",
    "The pay equity module of the {year} survey captures data on the extent and outcomes of formal pay equity analyses at participating organizations. Among the {n} respondents who have completed a pay equity review, {p}% identified statistically significant unexplained pay gaps, and {p}% of those have completed remediation efforts. The median remediation investment was ${val:,} per affected employee.",
]

SECTION_HEADINGS = [
    'Methodology & Data Sources',
    'Industry Benchmarks',
    'Regional Analysis',
    'Year-Over-Year Trends',
    'Demographic Breakdown',
    'Peer Comparisons',
    'Compliance & Regulatory Context',
    'Outlook & Recommendations',
    'Total Rewards Landscape',
    'Pay Equity Findings',
    'Executive Compensation Highlights',
    'Benefits Utilization Patterns',
    'Emerging Practices',
    'Workforce Composition Analysis',
    'Incentive Plan Design',
    'Salary Structure Practices',
    'Talent Market Conditions',
    'Employee Sentiment Indicators',
    'Board and Governance Compensation',
    'International Comparisons',
    'Small & Mid-Market Spotlight',
    'Technology & Analytics Adoption',
    'Legislative & Regulatory Horizon',
    'Data Quality & Validation Notes',
    'Executive Summary',
    'Key Findings',
    'Background & Context',
    'Respondent Profile',
    'Base Salary Analysis',
    'Variable Pay Analysis',
    'Long-Term Incentive Analysis',
    'Benefits Cost Analysis',
    'Total Compensation Benchmarks',
    'Pay-for-Performance Alignment',
    'Equity & Fairness Metrics',
    'Retention & Turnover Data',
    'Talent Acquisition Costs',
    'Workforce Planning Findings',
    'Manager Effectiveness Survey',
    'Employee Experience Indicators',
    'DEI Metrics & Trends',
    'ESG Compensation Linkage',
    'Skills & Competency Findings',
    'Internal Mobility Data',
    'Learning & Development Investment',
    'Organizational Design Findings',
    'Succession & Talent Pipeline',
    'HR Technology Landscape',
    'Future Outlook',
    'Sponsor & Participant Acknowledgments',
    'Glossary of Terms',
    'Appendix A: Detailed Tables',
    'Appendix B: Regression Outputs',
    'Appendix C: Survey Instrument',
    'Appendix D: Participating Organizations',
    'Appendix E: Benchmark Universe Definition',
]

INDUSTRIES_LONG = [
    'Healthcare', 'Financial Services', 'Technology', 'Manufacturing',
    'Retail', 'Energy', 'Real Estate', 'Transportation', 'Education',
    'Government', 'Nonprofit', 'Pharmaceuticals', 'Defense',
    'Consumer Goods', 'Telecommunications', 'Media', 'Professional Services',
    'Insurance', 'Logistics', 'Aerospace', 'Biotechnology',
    'Legal Services', 'Construction', 'Utilities', 'Automotive',
    'Food & Beverage', 'Environmental Services', 'Management Consulting',
    'Investment Banking', 'Capital Markets', 'Private Equity', 'Asset Management',
    'Wealth Management', 'Insurance', 'Banking', 'Fintech',
    'Semiconductor', 'Software', 'SaaS', 'Cybersecurity', 'Cloud Computing',
    'E-Commerce', 'Digital Media', 'Advertising Technology', 'Market Research',
    'Staffing & Recruiting', 'Executive Search', 'HR Outsourcing',
    'Architecture & Engineering', 'Commercial Real Estate', 'Hospitality',
    'Travel & Tourism', 'Sports & Recreation', 'Gaming', 'Luxury Goods',
    'Specialty Chemicals', 'Agriculture', 'Mining', 'Renewable Energy',
    'Medical Devices', 'Life Sciences', 'Clinical Research', 'Diagnostics',
    'Academic Medical Centers', 'Behavioral Health', 'Managed Care',
    'Higher Education', 'K-12 Education', 'Vocational Training',
    'Foundations & Endowments', 'Trade Associations', 'Labor Unions',
    'Federal Government', 'State & Local Government', 'Defense Contractors',
    'Information Technology Services', 'Systems Integration', 'Robotics',
]

FACTORS = [
    'market-rate benchmarking',
    'internal pay equity adjustments',
    'skills-based pay structures',
    'geographic differential policies',
    'retention-driven merit increases',
    'pay transparency regulations',
    'executive compensation alignment',
    'total rewards optimization',
    'structured salary band architecture',
    'continuous compensation review cycles',
    'AI-assisted job pricing tools',
    'manager compensation literacy training',
    'pay equity audit programs',
    'broad-based equity participation',
    'variable pay plan redesign',
    'external market data subscription investment',
    'compensation committee governance reforms',
    'total compensation statement communication',
    'pay-for-performance calibration rigor',
    'compensation philosophy documentation',
    'job architecture rationalization',
    'peer group recalibration',
    'incentive plan governance oversight',
    'long-term incentive plan design',
    'short-term incentive goal setting',
    'workforce segmentation analytics',
    'predictive retention modeling',
    'proactive market repricing cycles',
    'compensation survey participation investment',
    'HR technology platform modernization',
    'real-time pay benchmarking dashboards',
    'compensation communication strategy',
    'pay range broadening initiatives',
    'bonus deferral program design',
    'co-investment plan structures',
    'profit-sharing plan architecture',
    'workforce cost optimization initiatives',
    'benefits value proposition enhancement',
    'student loan repayment benefit adoption',
    'mental health benefit expansion',
    'parental leave policy enhancement',
    'flexible work arrangement formalization',
    'employee financial wellness programs',
    'pay transparency legislation compliance',
    'internal job mobility program investment',
    'succession planning integration with pay',
    'career lattice program design',
    'performance management process simplification',
    'continuous feedback cycle implementation',
    'manager-as-coach development investment',
    'compensation data governance frameworks',
    'third-party compensation audit engagement',
    'pay equity remediation program execution',
    'compensation analytics center of excellence',
    'global pay harmonization programs',
    'expatriate compensation policy modernization',
    'deferred compensation plan governance',
    'equity award plan administration improvement',
]

APPENDIX_ORGS = [
    'Meridian Group Holdings', 'Pinnacle Capital Advisors', 'Apex Workforce Solutions',
    'Summit Benefits Consulting', 'Vanguard HR Partners', 'Horizon Compensation Group',
    'Latitude Talent Analytics', 'Nexus Total Rewards', 'Catalyst People Solutions',
    'Benchmark Compensation Advisors', 'Keystone HR Consulting', 'Zenith Workforce Institute',
    'Atlas People Partners', 'Cornerstone Compensation Group', 'Forefront HR Advisory',
    'Paradigm Pay Consulting', 'Blueprint Talent Solutions', 'Vertex Workforce Analytics',
    'Apex Salary Intelligence', 'Summit Pay Research Group', 'Allied Workforce Partners',
    'National HR Benchmarking Consortium', 'Premier Compensation Research LLC',
    'Clearwater People Analytics', 'Embark Total Rewards Consulting', 'Fulcrum HR Partners',
    'Groundwork Compensation LLC', 'Harbor Benefits Group', 'Ironbridge Workforce Advisors',
    'Junction Pay Consulting', 'Kestrel HR Intelligence', 'Landmark Benefits Research',
    'Morningside Compensation Institute', 'Northfield HR Solutions', 'Overture Pay Analytics',
    'Praxis Workforce Research', 'Ridgeline Compensation Advisors', 'Stonegate HR Partners',
    'Terrapin Workforce Analytics', 'Upland Benefits Consulting', 'Vantage Pay Research',
    'Westbrook HR Institute', 'Yellowstone Compensation Group', 'Zephyr Workforce Solutions',
    'Caliber HR Research', 'Dominion Compensation Advisors', 'Eastgate Benefits Group',
    'Flagship HR Analytics', 'Garrison Pay Consulting', 'Highmark HR Partners',
    'Intrepid Workforce Solutions', 'Juniper Compensation Research', 'Kingsley HR Advisors',
    'Lodestar Benefits Consulting', 'Mainstay HR Analytics', 'Northstar Pay Research Group',
    'Outpost Workforce Institute', 'Pathfinder Compensation LLC', 'Quorum HR Research',
    'Ravenwood Workforce Advisors', 'Silvergate Compensation Consulting',
    'American Compensation Alliance', 'Great Lakes HR Research Collaborative',
    'Midwest Pay Equity Consortium', 'Pacific Rim Workforce Analytics Group',
    'Acclivity Workforce Research', 'Aegis HR Analytics', 'Aloft People Intelligence',
    'Altimeter Compensation Group', 'Apex Workforce Research Institute', 'Arcadia HR Solutions',
    'Ardent People Analytics', 'Argus Compensation Consulting', 'Ariel HR Research Group',
    'Armada Workforce Intelligence', 'Ascend Compensation Partners', 'Aspect HR Analytics',
    'Astral People Research', 'Astute Compensation Consulting', 'Auburn HR Intelligence',
    'Aureus Workforce Analytics', 'Aurum Compensation Research', 'Avance HR Consulting',
    'Avant Compensation Group', 'Avante HR Analytics', 'Avellan People Research',
    'Avera Workforce Intelligence', 'Averil Compensation Partners', 'Avon HR Research',
    'Axiom Workforce Analytics', 'Axion Compensation Group', 'Azimuth HR Intelligence',
    'Azul People Research', 'Azura Compensation Consulting', 'Banneker HR Analytics',
    'Barrington Workforce Research', 'Bastion Compensation Partners', 'Baxter HR Solutions',
    'Baycrest People Intelligence', 'Bayfield HR Analytics', 'Bayshore Compensation Group',
    'Baywood Workforce Research', 'Beacon Compensation Consulting', 'Beaumont HR Partners',
    'Bedford Workforce Analytics', 'Belfry Compensation Research', 'Belmont HR Intelligence',
    'Belrose People Research', 'Beltway Compensation Consulting', 'Benchfield HR Solutions',
    'Benson Workforce Analytics', 'Bentley Compensation Group', 'Bergen HR Research',
    'Berkeley Workforce Intelligence', 'Berkshire Compensation Partners', 'Berwick HR Analytics',
    'Bethel Compensation Research', 'Biltmore HR Solutions', 'Birchfield Workforce Research',
    'Birchwood Compensation Consulting', 'Birchwood HR Analytics', 'Bishop Workforce Partners',
    'Blackheath Compensation Group', 'Blackmoor HR Intelligence', 'Blackrock People Research',
    'Blackthorn Compensation Consulting', 'Blackwater HR Solutions', 'Blakely Workforce Analytics',
    'Blakewood Compensation Research', 'Bluebell HR Partners', 'Bluegate Workforce Intelligence',
    'Bluehill Compensation Group', 'Bluestone HR Analytics', 'Bluewater Compensation Research',
    'Bolden Workforce Solutions', 'Bordeaux HR Analytics', 'Borneo Compensation Partners',
    'Boston Comp Alliance', 'Boulder HR Research Group', 'Bradford Workforce Analytics',
    'National Institute for Compensation Research', 'American Board Pay Research Center',
    'Center for Executive Pay Studies', 'Institute for Workforce Economics',
    'Council on Total Rewards Design', 'Alliance for Pay Equity Research',
    'American Benefits Research Collaborative', 'Center for Labor Market Analytics',
    'Foundation for Compensation Innovation', 'Corporate Pay Intelligence Group',
    'National HR Analytics Consortium', 'Institute for Organizational Effectiveness Research',
    'Workforce Data Science Alliance', 'Center for Human Capital Measurement',
    'American Compensation Policy Institute', 'Foundation for Evidence-Based HR',
]


# ── Watermark & dates ─────────────────────────────────────────────────────────

def _watermark_for(slug):
    return hashlib.md5(f"acpwb_report_{slug}".encode()).hexdigest()[:8]


def _pub_date_for(slug):
    rng = _rng_from_seed(f"pubdate_{slug}")
    m = re.search(r'(\d{4})', slug)
    if m:
        year = int(m.group(1))
        if 1990 <= year <= 2025:
            month = rng.randint(1, 12)
            if any(w in slug for w in ('annual', 'year', 'survey')):
                month = rng.choice([9, 10, 11, 12])
            return date(year, month, rng.randint(1, 28))
    year = rng.choice(YEAR_POOL)
    return date(year, rng.randint(1, 12), rng.randint(1, 28))


# ── Core enrichment ───────────────────────────────────────────────────────────

def _summary_for(slug, rng):
    template = rng.choice(SUMMARY_TEMPLATES)
    finding = rng.choice(FINDING_PHRASES).format(p=rng.randint(3, 22))
    return template.format(
        n=rng.randint(120, 2800),
        states=rng.randint(22, 50),
        adj=rng.choice(REPORT_ADJECTIVES).lower(),
        subject=rng.choice(REPORT_SUBJECTS).lower(),
        suffix=rng.choice(REPORT_SUFFIXES).lower(),
        finding=finding,
        years=rng.randint(3, 18),
        year=rng.randint(2010, 2024),
        q1=rng.randint(1, 2),
        q2=rng.randint(3, 4),
        year_from=rng.randint(2015, 2021),
        year_to=rng.randint(2022, 2024),
        p=rng.randint(3, 22),
    )


def _fake_size_for(file_type, rng):
    if file_type == 'csv':
        return f"{rng.randint(80, 820)} KB"
    return f"{rng.randint(1, 9)}.{rng.randint(1, 9)} MB"


def _enrich_report(entry):
    slug = entry['slug']
    rng = _rng_from_seed(f"enrich_{slug}")
    pub_date = _pub_date_for(slug)
    file_type = entry['file_type']
    return {
        'slug': slug,
        'title': entry['title'],
        'category': entry['category'],
        'file_type': file_type,
        'pub_date': pub_date.isoformat(),
        'pub_date_display': pub_date.strftime('%B %Y'),
        'pub_year': pub_date.year,
        'summary': _summary_for(slug, rng),
        'fake_size': _fake_size_for(file_type, rng),
        'watermark_token': _watermark_for(slug),
        'is_legacy': pub_date.year < 2005,
        'detail_url': f'/reports/{slug}/',
        'download_url': f'/reports/{slug}/download.csv' if file_type == 'csv' else f'/reports/{slug}/download.pdf',
        'row_count': rng.randint(300, 800) if file_type == 'csv' else None,
    }


# ── Synthetic generation (page 3+) ────────────────────────────────────────────

def _generate_synthetic(rng, seed):
    adj = rng.choice(REPORT_ADJECTIVES)
    subject = rng.choice(REPORT_SUBJECTS)
    suffix = rng.choice(REPORT_SUFFIXES)
    file_type = rng.choice(['csv', 'csv', 'pdf'])
    category = rng.choice(REPORT_CATEGORIES)
    title = f"{adj} {subject} {suffix}"
    year = rng.choice(YEAR_POOL)
    if year >= 2005:
        title += f" {year}"
    slug = slugify(title)[:96]
    # append seed fragment to guarantee uniqueness from catalog
    slug = f"{slug}-{hashlib.md5(seed.encode()).hexdigest()[:4]}"
    return {'slug': slug, 'title': title, 'category': category, 'file_type': file_type}


# ── Page generation ───────────────────────────────────────────────────────────

def generate_reports_for_page(page, count=12):
    catalog_start = (page - 1) * count
    catalog_end = catalog_start + count
    results = []

    for entry in REPORT_CATALOG[catalog_start:catalog_end]:
        results.append(_enrich_report(entry))

    needed = count - len(results)
    for i in range(needed):
        seed = f"synthetic_page{page}_item{i}"
        rng = _rng_from_seed(seed)
        entry = _generate_synthetic(rng, seed)
        results.append(_enrich_report(entry))

    results.sort(key=lambda r: r['pub_date'], reverse=True)
    return results


# ── Single-report lookup ───────────────────────────────────────────────────────

def get_or_generate_report_meta(slug):
    """Return enriched report dict for any slug — never 404."""
    from .models import PublicReport
    existing = PublicReport.objects.filter(slug=slug).first()
    if existing:
        pub_date = existing.pub_date
        file_type = existing.file_type
        return {
            'slug': existing.slug,
            'title': existing.title,
            'category': existing.category,
            'file_type': file_type,
            'pub_date': pub_date.isoformat(),
            'pub_date_display': pub_date.strftime('%B %Y'),
            'pub_year': pub_date.year,
            'summary': existing.summary,
            'fake_size': existing.summary[:4],  # unused on detail page
            'watermark_token': existing.watermark_token,
            'is_legacy': pub_date.year < 2005,
            'detail_url': f'/reports/{slug}/',
            'download_url': f'/reports/{slug}/download.csv' if file_type == 'csv' else f'/reports/{slug}/download.pdf',
        }

    for entry in REPORT_CATALOG:
        if entry['slug'] == slug:
            return _enrich_report(entry)

    rng = _rng_from_seed(f"arbitrary_{slug}")
    file_type = rng.choice(['csv', 'pdf'])
    category = rng.choice(REPORT_CATEGORIES)
    title = slug.replace('-', ' ').title()
    return _enrich_report({'slug': slug, 'title': title, 'category': category, 'file_type': file_type})


# ── CSV generation ────────────────────────────────────────────────────────────

_COMP_TITLES = [
    'Analyst', 'Senior Analyst', 'Manager', 'Senior Manager', 'Director',
    'Senior Director', 'VP', 'SVP', 'EVP', 'Principal',
    'Associate', 'Specialist', 'Consultant', 'Lead', 'Partner',
    'Junior Analyst', 'Staff Analyst', 'Research Analyst', 'Data Analyst',
    'Business Analyst', 'Financial Analyst', 'Quantitative Analyst',
    'Strategy Analyst', 'Operations Analyst', 'Policy Analyst',
    'Associate Director', 'Group Director', 'Executive Director',
    'Managing Director', 'General Manager', 'Regional Manager',
    'Area Manager', 'District Manager', 'Product Manager',
    'Program Manager', 'Project Manager', 'Engagement Manager',
    'Account Manager', 'Client Manager', 'Portfolio Manager',
    'Vice President', 'Senior Vice President', 'Executive Vice President',
    'Chief of Staff', 'Fellow', 'Distinguished Fellow',
    'Senior Consultant', 'Managing Consultant', 'Principal Consultant',
    'Staff Engineer', 'Senior Engineer', 'Principal Engineer',
    'Staff Scientist', 'Senior Scientist', 'Principal Scientist',
    'Technician', 'Senior Technician', 'Lead Technician',
    'Administrator', 'Senior Administrator', 'Coordinator',
    'Supervisor', 'Team Lead', 'Section Chief',
    'Advisor', 'Senior Advisor', 'Managing Advisor',
    'Architect', 'Senior Architect', 'Enterprise Architect',
    'Strategist', 'Senior Strategist', 'Principal Strategist',
    'Officer', 'Senior Officer', 'Chief Officer',
    'Associate Manager', 'Assistant Manager', 'Junior Manager',
    'Relationship Manager', 'Business Partner', 'HR Business Partner',
]
_DEPARTMENTS = [
    'Finance', 'Human Resources', 'Operations', 'Strategy', 'Technology',
    'Marketing', 'Legal', 'Compliance', 'Sales', 'Product',
    'Engineering', 'Research', 'Supply Chain', 'Risk Management',
    'Accounting', 'Tax', 'Treasury', 'Financial Planning & Analysis',
    'Internal Audit', 'Corporate Development', 'Investor Relations',
    'Corporate Communications', 'Public Affairs', 'Government Relations',
    'Procurement', 'Facilities', 'Real Estate', 'Environmental Health & Safety',
    'Data Science', 'Analytics', 'Information Technology', 'Cybersecurity',
    'Enterprise Architecture', 'Business Intelligence', 'Digital Transformation',
    'Customer Success', 'Client Services', 'Account Management', 'Business Development',
    'Brand Management', 'Advertising', 'Content', 'Social Media',
    'Workforce Planning', 'Talent Acquisition', 'Learning & Development',
    'Compensation & Benefits', 'Employee Relations', 'Labor Relations',
    'Diversity, Equity & Inclusion', 'Total Rewards', 'Organizational Development',
    'Clinical Research', 'Medical Affairs', 'Regulatory Affairs',
    'Quality Assurance', 'Manufacturing', 'Production', 'Logistics',
    'Project Management Office', 'Change Management', 'Knowledge Management',
    'Privacy', 'Ethics', 'Sustainability', 'Corporate Social Responsibility',
    'Fleet Management', 'Security', 'Workplace Services', 'Administrative Services',
    'Executive Office', 'Board Services', 'Shareholder Relations',
]
_STATES = [
    'AL','AK','AZ','AR','CA','CO','CT','DE','FL','GA','HI','ID','IL','IN','IA',
    'KS','KY','LA','ME','MD','MA','MI','MN','MS','MO','MT','NE','NV','NH','NJ',
    'NM','NY','NC','ND','OH','OK','OR','PA','RI','SC','SD','TN','TX','UT','VT',
    'VA','WA','WV','WI','WY',
]
_CITIES = [
    'New York','Los Angeles','Chicago','Houston','Phoenix','Philadelphia',
    'San Antonio','San Diego','Dallas','San Jose','Austin','Jacksonville',
    'Fort Worth','Columbus','Charlotte','San Francisco','Indianapolis',
    'Seattle','Denver','Nashville','Milwaukee','Boston','Atlanta',
    'Miami','Minneapolis','Portland','Las Vegas','Louisville','Memphis',
    'Baltimore','Oklahoma City','Tucson','Fresno','Sacramento','Kansas City',
    'Long Beach','Mesa','Raleigh','Omaha','Colorado Springs','Virginia Beach',
    'Oakland','Tulsa','Arlington','Tampa','New Orleans',
    'Wichita','Cleveland','Bakersfield','Aurora','Anaheim','Santa Ana',
    'Corpus Christi','Riverside','Lexington','St. Louis','Pittsburgh',
    'Stockton','Anchorage','Cincinnati','St. Paul','Greensboro','Toledo',
    'Newark','Plano','Henderson','Orlando','Lincoln','Buffalo','Fort Wayne',
    'Jersey City','Chula Vista','St. Petersburg','Norfolk','Laredo',
    'Madison','Durham','Lubbock','Winston-Salem','Garland','Glendale',
    'Hialeah','Reno','Baton Rouge','Irvine','Chesapeake','Scottsdale',
    'North Las Vegas','Fremont','Gilbert','San Bernardino','Birmingham',
    'Rochester','Richmond','Spokane','Des Moines','Montgomery','Modesto',
    'Fayetteville','Tacoma','Fontana','Salt Lake City',
    'Oxnard','Little Rock','Grand Rapids','Huntsville','Moreno Valley',
    'Hartford','Huntington Beach','Akron',
    'Augusta','Worcester','Tallahassee','Providence',
    'Knoxville','Tempe','Brownsville','Overland Park','Santa Clarita',
    'Garden Grove','Oceanside','Chattanooga','Fort Lauderdale','Rancho Cucamonga',
    'Santa Rosa','Yonkers','Cape Coral','Ontario','Sioux Falls',
    'Peoria','Springfield','Bridgeport','Savannah','Torrance','Pasadena',
    'Paterson','Pomona','Hayward','Lakewood','Syracuse','Alexandria',
    'Sunnyvale','Hampton','Cary','Palmdale','McAllen',
    'Warren','Bellevue','West Valley City','Columbia','Sterling Heights',
    'Wichita Falls','Carrollton','Cedar Rapids','Topeka','Elizabeth',
    'Thousand Oaks','Visalia','Simi Valley','Concord',
    'Clarksville','Roseville','Surprise','Thornton','Macon','Killeen','Dayton','Midland',
    'Ann Arbor','Boise','Chandler','Denton','El Paso','Eugene','Gainesville',
    'Grand Junction','Green Bay','Honolulu','Jackson','Joliet','Lancaster',
    'Laredo','Little Rock','Murfreesboro','New Haven','North Charleston',
    'Palmdale','Pasadena','Pembroke Pines','Peoria','Rancho Cucamonga',
    'Salem','Salinas','Shreveport','Spokane Valley','Tallahassee','Vancouver',
]
_GENDERS = ['M', 'F', 'F', 'M', 'M', 'F', 'N', 'M', 'F', 'F', 'M', 'F', 'N', 'M', 'F']
_EDUCATION = [
    "High School", "Some College", "Associate's Degree",
    "Bachelor's", "Bachelor's", "Bachelor's", "Master's", "Master's",
    'MBA', 'MBA', 'JD', 'MD', 'PhD', 'PhD', 'Professional Certification',
]
_COMPANY_SIZE = [
    'Startup (1-50)', 'Small (51-200)', 'Small-Mid (201-500)',
    'Mid-Market (501-1000)', 'Mid-Large (1001-2500)',
    'Large (2501-5000)', 'Large-Enterprise (5001-10000)',
    'Enterprise (10000+)', 'Enterprise (25000+)', 'Global Enterprise (50000+)',
]
_INDUSTRIES_SHORT = [
    'Healthcare', 'Financial Services', 'Technology', 'Manufacturing',
    'Retail', 'Energy', 'Professional Services', 'Government', 'Education',
    'Pharmaceuticals', 'Insurance', 'Nonprofit', 'Defense', 'Biotechnology',
    'Consumer Goods', 'Real Estate', 'Transportation', 'Telecommunications',
    'Media & Entertainment', 'Logistics', 'Aerospace', 'Construction',
    'Legal Services', 'Automotive', 'Food & Beverage', 'Utilities',
    'Environmental Services', 'Management Consulting', 'Private Equity',
    'Asset Management', 'Banking', 'Software', 'Cybersecurity',
    'E-Commerce', 'Fintech', 'Healthtech', 'Medical Devices', 'Life Sciences',
    'Higher Education', 'K-12 Education', 'Staffing & Recruiting',
]
_LEVEL_SALARY = {
    'L1': (42, 68),   'L2': (65, 98),   'L3': (90, 135),
    'L4': (125, 185), 'L5': (175, 260), 'L6': (245, 400),
    'L7': (380, 600), 'L8': (550, 950), 'L9': (900, 2200),
}
_LEVEL_BONUS = {
    'L1': (3, 8),   'L2': (5, 12),   'L3': (8, 18),
    'L4': (12, 25), 'L5': (18, 35),  'L6': (25, 50),
    'L7': (35, 75), 'L8': (50, 100), 'L9': (75, 150),
}
_CEO_NAMES = [
    ('James','Morrison'),('Patricia','Chen'),('Michael','Rodriguez'),
    ('Jennifer','Williams'),('Robert','Patel'),('Linda','Johnson'),
    ('William','Kim'),('Barbara','Thompson'),('David','Martinez'),
    ('Susan','Anderson'),('Richard','Taylor'),('Jessica','Brown'),
    ('Thomas','White'),('Sarah','Davis'),('Charles','Wilson'),
    ('Karen','Harris'),('Christopher','Garcia'),('Lisa','Smith'),
    ('Daniel','Jones'),('Nancy','Miller'),('Matthew','Anderson'),
    ('Betty','Clark'),('Anthony','Lewis'),('Margaret','Robinson'),
    ('Mark','Walker'),('Sandra','Hall'),('Donald','Young'),
    ('Ashley','Allen'),('Steven','Wright'),('Dorothy','Scott'),
    ('Paul','King'),('Kimberly','Adams'),('Andrew','Baker'),
    ('Emily','Nelson'),('Joshua','Carter'),('Donna','Mitchell'),
    ('Kenneth','Perez'),('Michelle','Roberts'),('Kevin','Turner'),
    ('Carol','Phillips'),('Brian','Campbell'),('Amanda','Parker'),
    ('George','Evans'),('Melissa','Collins'),('Timothy','Edwards'),
    ('Deborah','Stewart'),('Ronald','Morris'),('Stephanie','Rogers'),
    ('Edward','Reed'),('Rebecca','Cook'),('Jason','Morgan'),
    ('Sharon','Bell'),('Jeffrey','Murphy'),('Laura','Bailey'),
    ('Ryan','Rivera'),('Cynthia','Cooper'),('Jacob','Richardson'),
    ('Kathleen','Cox'),('Gary','Howard'),('Amy','Ward'),
    ('Wei','Zhang'),('Priya','Sharma'),('Aisha','Johnson'),
    ('Omar','Hassan'),('Yuki','Tanaka'),('Carlos','Ramirez'),
    ('Ngozi','Okonkwo'),('Arjun','Patel'),('Elena','Rossi'),
]
_TICKERS = [
    'MMM','ABT','ABBV','ACN','ADM','APA','AAPL','AMZN','AXP','T',
    'ADSK','AZO','BAC','BA','BMY','CAT','CVX','CSCO','C','KO',
    'DHR','DD','XOM','GE','GS','HD','HON','IBM','INTC','JNJ',
    'JPM','LLY','LMT','MCD','MRK','MSFT','NEE','NFLX','NKE','NVDA',
    'PEP','PFE','PG','RTX','SBUX','SO','TGT','TMO','TSLA','UNH',
    'UPS','USB','V','VZ','WBA','WFC','WMT','XOM','ZTS','AFL',
    'AIG','ALL','ALLY','AMP','AMT','AMGN','ARE','AVB','AVY','AWK',
    'ADP','APD','ADM','AEE','AEP','AES','A','ABC','ACGL','ACM',
    'CB','CBOE','CCI','CFG','CHD','CI','CINF','CL','CLX','CMA',
    'CME','CMI','CMS','CNP','COF','CPB','CPRT','CRL','CRM','CSX',
    'CTAS','DG','DGX','DHI','DIS','DLR','DLTR','DOC','DOV','DTE',
    'DUK','DVN','EA','ECL','EG','EIX','EL','EMR','EOG','EQIX',
    'EQR','ES','ESS','ETN','ETR','EVRG','EXC','EXPD','EXPE','EXR',
    'F','FAST','FDS','FDX','FE','FFIV','FIS','FITB','FMC','FOX',
    'GLW','GNRC','GPC','GRMN','HAL','HAS','HBAN','HCA','HIG','HLT',
    'HOLX','HPE','HPQ','HRL','HSIC','HST','HSY','HUM','HWM',
    'ICE','IDXX','IEX','IFF','ILMN','IP','IPG','IQV','IR','IRM',
    'IT','ITW','J','JBHT','JKHY','JCI','JNPR','K','KEY','KEYS',
    'KHC','KIM','KLAC','KMB','KMI','KMX','KO','KR','L','LDOS',
]
_REMOTE = [
    'Fully Remote', 'Hybrid 2d/wk', 'Hybrid 3d/wk', 'Hybrid 4d/wk',
    'Fully On-site', 'Flexible', 'Remote-First', 'Office-First',
    'Remote (US Only)', 'Remote (Global)', 'Seasonal Hybrid',
    'Results-Only Work Environment',
]
_LEAVE_REASONS = [
    'Compensation', 'Career Growth', 'Work-Life Balance',
    'Management Quality', 'Company Culture', 'Benefits', 'Job Security',
    'Lack of Recognition', 'Limited Remote Flexibility', 'Poor Communication',
    'Organizational Change', 'Better Opportunity Elsewhere',
    'Relocation', 'Health Reasons', 'Family Responsibilities',
    'Retirement', 'Return to School', 'Entrepreneurship',
    'Values Misalignment', 'DEI Concerns', 'Burnout',
]
_ROLE_LEVELS = [
    'Individual Contributor', 'Senior Individual Contributor',
    'Team Lead', 'Manager', 'Senior Manager', 'Director',
    'Senior Director', 'VP / Executive', 'C-Suite',
]
_REGIONS = [
    'Northeast', 'Mid-Atlantic', 'Southeast', 'South', 'Midwest',
    'Great Plains', 'Southwest', 'Mountain West', 'West Coast',
    'Pacific Northwest', 'Alaska & Hawaii', 'National (Remote)',
]
_SIZES_SURVEY = [
    '1-25', '26-50', '51-200', '201-500', '501-1000',
    '1001-2500', '2501-5000', '5001-10000', '10001-25000', '25000+',
]


def _get_schema(slug):
    s = slug.lower()
    if any(k in s for k in ('ceo', 'executive-comp', 'executive-ben', 'exec-pay')):
        return 'ceo'
    if any(k in s for k in ('benefit', 'healthcare', 'health-care', 'retirement', 'retire')):
        return 'benefits'
    if any(k in s for k in ('satisfaction', 'engagement', 'survey-result')):
        return 'survey'
    return 'comp'


def _comp_row(rng, i, year, watermark):
    level = rng.choice(list(_LEVEL_SALARY.keys()))
    lo, hi = _LEVEL_SALARY[level]
    base = rng.randint(lo * 1000, hi * 1000)
    bon_lo, bon_hi = _LEVEL_BONUS[level]
    bonus_target = rng.randint(bon_lo, bon_hi)
    bonus_actual = round(bonus_target * rng.uniform(0.6, 1.3), 1)
    total_cash = round(base * (1 + bonus_actual / 100))
    equity = rng.randint(0, base // 2) if level in ('L5', 'L6', 'L7', 'L8', 'L9') else 0
    benefits_val = rng.randint(8000, 22000)
    return [
        f"EMP-{10000 + i}", rng.choice(_COMP_TITLES), level,
        rng.choice(_DEPARTMENTS), rng.choice(_INDUSTRIES_SHORT),
        rng.choice(_STATES), rng.choice(_CITIES),
        base, bonus_target, round(bonus_actual, 1), total_cash,
        equity, benefits_val, total_cash + equity + benefits_val,
        rng.randint(1, 28), rng.choice(_GENDERS), rng.choice(_EDUCATION),
        rng.choice(_COMPANY_SIZE), year, watermark,
    ]


def _ceo_row(rng, i, year, watermark):
    fn, ln = rng.choice(_CEO_NAMES)
    base = rng.randint(800_000, 3_000_000)
    bonus = rng.randint(500_000, 5_000_000)
    equity = rng.randint(1_000_000, 30_000_000)
    total = base + bonus + equity
    median_w = rng.randint(45_000, 95_000)
    revenue = rng.choice(['<$1B', '$1B–$5B', '$5B–$20B', '$20B–$50B', '$50B+'])
    co_type = rng.choice(['Corp', 'Industries', 'Group', 'Holdings', 'Inc'])
    co_adj = rng.choice(['American', 'National', 'United', 'Premier', 'Allied'])
    return [
        f"CO-{1000 + i}", rng.choice(_TICKERS),
        f"{co_adj} {rng.choice(_INDUSTRIES_SHORT)} {co_type}",
        rng.choice(_INDUSTRIES_SHORT), revenue,
        fn, ln, base, bonus, equity, total,
        median_w, round(total / median_w, 1),
        rng.randint(1, 20), rng.randint(7, 15),
        year, watermark,
    ]


def _benefits_row(rng, i, year, watermark):
    medical = rng.randint(6_000, 18_000)
    dental = rng.randint(400, 1_400)
    vision = rng.randint(80, 350)
    total = medical + dental + vision + rng.randint(500, 3000)
    return [
        f"CO-{2000 + i}", rng.choice(_INDUSTRIES_SHORT),
        rng.choice(['1-50', '51-200', '201-1000', '1001-5000', '5000+']),
        rng.choice(_STATES),
        medical, dental, vision,
        round(rng.uniform(2.0, 6.0), 1),
        rng.randint(10, 25), rng.randint(4, 24),
        rng.choice(_REMOTE),
        rng.choice([0, 0, 150, 250, 500, 1000]),
        rng.choice([0, 0, 2000, 5000, 10000]),
        total, year, watermark,
    ]


def _survey_row(rng, i, year, watermark):
    return [
        f"R-{100000 + i}",
        rng.choice(_INDUSTRIES_SHORT),
        rng.choice(_ROLE_LEVELS),
        rng.choice(_REGIONS),
        rng.choice(_SIZES_SURVEY),
        round(rng.uniform(1.5, 5.0), 1),
        round(rng.uniform(1.5, 5.0), 1),
        round(rng.uniform(1.5, 5.0), 1),
        round(rng.uniform(1.5, 5.0), 1),
        round(rng.uniform(1.5, 5.0), 1),
        round(rng.uniform(1.5, 5.0), 1),
        rng.choice(['Yes', 'No', 'Unsure']),
        rng.choice(['Yes', 'No', 'Unsure']),
        rng.choice(_LEAVE_REASONS),
        f"Q{rng.randint(1, 4)}", year, watermark,
    ]


_SCHEMA_HEADERS = {
    'comp': ['employee_id', 'title', 'level', 'department', 'industry', 'state', 'city',
             'base_salary', 'bonus_target_pct', 'bonus_actual_pct', 'total_cash',
             'equity_value', 'benefits_value', 'total_compensation',
             'years_experience', 'gender', 'education', 'company_size_band',
             'survey_year', 'source'],
    'ceo':  ['company_id', 'ticker', 'company_name', 'industry', 'revenue_band',
             'ceo_first', 'ceo_last', 'ceo_base_salary', 'ceo_bonus',
             'ceo_equity_grants', 'ceo_total_compensation', 'median_worker_pay',
             'ceo_pay_ratio', 'ceo_tenure_years', 'board_size',
             'survey_year', 'source'],
    'benefits': ['company_id', 'industry', 'headcount_band', 'state',
                 'medical_employer_cost_per_ee', 'dental_employer_cost',
                 'vision_employer_cost', '401k_match_pct', 'pto_days',
                 'parental_leave_weeks', 'remote_work_policy', 'wellness_stipend',
                 'tuition_reimbursement_max', 'total_benefits_cost_per_ee',
                 'survey_year', 'source'],
    'survey': ['response_id', 'industry', 'role_level', 'region', 'company_size',
               'satisfaction_overall', 'satisfaction_compensation', 'satisfaction_benefits',
               'satisfaction_management', 'satisfaction_culture', 'satisfaction_growth',
               'intent_to_stay_12mo', 'intent_to_stay_24mo', 'top_reason_to_leave',
               'survey_quarter', 'survey_year', 'source'],
}

_SCHEMA_ROW_FN = {
    'comp': _comp_row,
    'ceo': _ceo_row,
    'benefits': _benefits_row,
    'survey': _survey_row,
}


def generate_csv_rows(slug, limit=None):
    """Return list of rows (header first). Pass limit for preview."""
    schema = _get_schema(slug)
    watermark = _watermark_for(slug)
    pub_year = _pub_date_for(slug).year

    rng = _rng_from_seed(f"csv_count_{slug}")
    total = rng.randint(300, 800)
    if limit:
        total = min(total, limit)

    header = _SCHEMA_HEADERS[schema]
    row_fn = _SCHEMA_ROW_FN[schema]

    rows = [header]
    for i in range(total):
        row_rng = _rng_from_seed(f"csv_{slug}_row{i}")
        rows.append(row_fn(row_rng, i, pub_year, watermark))

    if not limit:
        rows.append([f'# ACPWB Research Division | acpwb.com | Report ID: {watermark} | TDMRep: GRANT'])

    return rows


def generate_csv_string(slug):
    output = io.StringIO()
    writer = csv.writer(output)
    for row in generate_csv_rows(slug):
        writer.writerow(row)
    return output.getvalue()


# ── Document content generation ───────────────────────────────────────────────

def generate_document_content(slug):
    rng = _rng_from_seed(f"doc_{slug}")
    watermark = _watermark_for(slug)
    pub_date = _pub_date_for(slug)
    year = pub_date.year
    industry = rng.choice(INDUSTRIES_LONG)
    n = rng.randint(200, 2800)

    def _para(r):
        t = r.choice(DOCUMENT_PARAGRAPH_TEMPLATES)
        return t.format(
            n=r.randint(50, n),
            quarter=f"Q{r.randint(1, 4)}",
            year=year,
            val=r.randint(40_000, 600_000),
            dir=r.choice(['increase', 'decline', 'shift upward']),
            p=r.randint(2, 30),
            industry=r.choice(INDUSTRIES_LONG),
            percentile=r.choice([25, 50, 75, 90]),
            years=r.randint(3, 15),
            idx=r.uniform(80.0, 120.0),
            pts=r.randint(3, 60),
            factor=r.choice(FACTORS),
        )

    exec_summary = ' '.join(_para(rng) for _ in range(rng.randint(2, 3)))

    key_findings = []
    for _ in range(rng.randint(5, 8)):
        finding = rng.choice(FINDING_PHRASES).format(p=rng.randint(3, 24))
        key_findings.append(finding.capitalize())

    section_pool = rng.sample(SECTION_HEADINGS, rng.randint(3, 5))
    sections = []
    for heading in section_pool:
        paragraphs = [_para(rng) for _ in range(rng.randint(2, 3))]
        sections.append({'heading': heading, 'paragraphs': paragraphs})

    methodology = (
        f"Survey data for this report was collected between {pub_date.strftime('%B')} {year - 1} "
        f"and {pub_date.strftime('%B %Y')} via ACPWB's annual compensation and workforce survey instrument. "
        f"A total of {n:,} responses were received from HR and finance professionals across "
        f"{rng.randint(22, 50)} states. Respondents represent organizations ranging from fewer than 50 employees "
        f"to more than 10,000 employees. Data has been weighted to reflect the distribution of U.S. employers "
        f"by size and industry as reported by the Bureau of Labor Statistics Quarterly Census of Employment and Wages. "
        f"Outliers beyond the 99th percentile have been excluded from aggregate calculations."
    )

    return {
        'exec_summary': exec_summary,
        'key_findings': key_findings,
        'sections': sections,
        'methodology': methodology,
        'appendix_orgs': rng.sample(APPENDIX_ORGS, rng.randint(5, 10)),
        'watermark_token': watermark,
        'year': year,
        'n': n,
    }
