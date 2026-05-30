TITLE_CASE_LOWER = frozenset([
    'a', 'an', 'the', 'and', 'but', 'or', 'nor', 'for', 'so',
    'at', 'by', 'in', 'of', 'on', 'to', 'up', 'as', 'vs',
])

ACRONYMS = {
    # C-suite / titles
    'ceo': 'CEO', 'cfo': 'CFO', 'cto': 'CTO', 'coo': 'COO',
    'chro': 'CHRO', 'clo': 'CLO', 'cio': 'CIO', 'ciso': 'CISO',
    'cro': 'CRO', 'cmo': 'CMO', 'ocio': 'OCIO',
    'vp': 'VP', 'svp': 'SVP', 'evp': 'EVP', 'md': 'MD',
    # Business / HR / finance
    'hr': 'HR', 'dei': 'DEI', 'esg': 'ESG', 'roi': 'ROI',
    'kpi': 'KPI', 'pmo': 'PMO', 'okr': 'OKR', 'mbo': 'MBO',
    'fte': 'FTE', 'ipo': 'IPO', 'pe': 'PE', 'vc': 'VC',
    'ebitda': 'EBITDA', 'cagr': 'CAGR', 'npv': 'NPV', 'irr': 'IRR',
    'nps': 'NPS', 'cac': 'CAC', 'ltv': 'LTV', 'arr': 'ARR',
    'mrr': 'MRR', 'arpu': 'ARPU', 'roic': 'ROIC', 'tsr': 'TSR',
    'lti': 'LTI', 'ltip': 'LTIP', 'spac': 'SPAC', 'adr': 'ADR',
    'raci': 'RACI', 'rfp': 'RFP', 'rfi': 'RFI', 'rfq': 'RFQ',
    'nda': 'NDA', 'sla': 'SLA', 'kpi': 'KPI', 'erp': 'ERP',
    'crm': 'CRM', 'hris': 'HRIS', 'lms': 'LMS', 'ats': 'ATS',
    'cogs': 'COGS', 'gpo': 'GPO', 'peo': 'PEO', 'erm': 'ERM',
    # Regulatory / standards
    'gdpr': 'GDPR', 'ccpa': 'CCPA', 'sec': 'SEC', 'iso': 'ISO',
    'osha': 'OSHA', 'eeoc': 'EEOC', 'cpa': 'CPA', 'irc': 'IRC',
    'gri': 'GRI', 'sasb': 'SASB', 'tcfd': 'TCFD', 'cdp': 'CDP',
    'cobra': 'COBRA', 'flsa': 'FLSA', 'ofccp': 'OFCCP',
    # Technology
    'it': 'IT', 'ai': 'AI', 'ml': 'ML', 'nlp': 'NLP', 'rpa': 'RPA',
    'api': 'API', 'erp': 'ERP', 'saas': 'SaaS', 'sql': 'SQL',
    'iot': 'IoT', 'gpu': 'GPU', 'fpga': 'FPGA', 'asic': 'ASIC',
    'cbrs': 'CBRS', 'ran': 'RAN', 'ui': 'UI', 'ux': 'UX',
    # Industries / general
    'ngo': 'NGO', 'oem': 'OEM', 'cpg': 'CPG', 'apac': 'APAC',
    'stem': 'STEM', 'covid': 'COVID', 'crispr': 'CRISPR',
    'tv': 'TV', 'seo': 'SEO', 'sem': 'SEM',
}
