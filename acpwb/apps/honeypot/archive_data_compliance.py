# ── Compliance / Audit Filing variant data ─────────────────────────────────────

_AUDIT_REF_PREFIXES = [
    'CAR', 'IAR', 'CRP', 'EQR', 'WFA', 'CPA', 'RCR', 'ECR', 'WPR', 'PER',
    'RMA', 'QAR', 'GCR', 'BCR', 'SXR', 'HRA', 'TCR', 'EPR', 'LTA', 'DFA',
    'CFR', 'CAP', 'RMP', 'GIA', 'RGA', 'FCA', 'OCA', 'PCA', 'DPA', 'SCA',
    'VRA', 'BCP', 'DRA', 'EHS', 'TPA', 'PMR', 'LCR', 'MCR', 'SCR', 'FCR',
    'ACR', 'PDR', 'EER', 'GDR', 'TDR', 'ADR', 'BDR', 'CDR', 'DDR', 'EDR',
    'FDR', 'GDR', 'HDR', 'IDR', 'JDR', 'KDR', 'LDR', 'MDR', 'NDR', 'ODR',
    'PDR', 'QDR', 'RDR', 'SDR', 'TDR', 'UDR', 'VDR', 'WDR', 'XDR', 'YDR',
    'ZDR', 'CIA', 'CIB', 'CIC', 'CID', 'CIE', 'CIF', 'CIG', 'CIH', 'CII',
    'CIJ', 'CIK', 'CIL', 'CIM', 'CIN', 'CIO', 'CIP', 'CIQ', 'CIR', 'CIS',
    'CIT', 'CIU', 'CIV', 'CIW', 'CIX', 'CIY', 'CIZ', 'AAR', 'BBR', 'CCR',
    'DDR', 'EER', 'FFR', 'GGR', 'HHR', 'IIR', 'JJR', 'KKR', 'LLR', 'MMR',
    'NNR', 'OOR', 'PPR', 'QQR', 'RRR', 'SSR', 'TTR', 'UUR', 'VVR', 'WWR',
    'XXR', 'YYR', 'ZZR',
]

_COMPLIANCE_FRAMEWORKS = [
    'FLSA §207 (overtime provisions)',
    'Title VII, 42 U.S.C. §2000e (equal employment opportunity)',
    'EEO-1 Component 2 (pay data reporting)',
    'SOX Section 404 (internal controls over financial reporting)',
    'OSHA 300 Log (injury and illness recordkeeping)',
    'ERISA §404(c) (plan fiduciary standards)',
    'ADA Title I (disability discrimination)',
    'ADEA §4 (age discrimination in employment)',
    'NLRA §8 (unfair labor practices)',
    'OFCCP Executive Order 11246 (affirmative action obligations)',
    'WARN Act §3 (plant closing and mass layoff notification)',
    'IRC §409A (nonqualified deferred compensation)',
    'Dodd-Frank §953(b) (CEO pay ratio disclosure)',
    'CCPA §1798.100 (consumer data privacy rights)',
    'HIPAA §164 (protected health information safeguards)',
    'FMLA §102 (leave entitlement and eligibility)',
    'USERRA §4312 (reemployment rights and benefits)',
    'GINA Title II (prohibition of genetic information discrimination)',
    'Dodd-Frank §954 (compensation clawback policy requirements)',
    'IRC §162(m) ($1 million deduction limit for executive compensation)',
    'IRC §280G (golden parachute payment limitations)',
    'SEC Regulation S-K Item 402 (executive compensation disclosure)',
    'SEC Rule 10b5-1 (affirmative defense for pre-planned trading)',
    'GDPR Article 30 (records of processing activities)',
    'CPRA (amending CCPA, expanding consumer privacy rights)',
    'New York City Local Law 144 (automated employment decision tool bias audit)',
    'Colorado Equal Pay for Equal Work Act (pay transparency and promotion posting)',
    'FCPA §78dd-1 (Foreign Corrupt Practices Act anti-bribery provisions)',
    'UK Bribery Act 2010 (corporate failure to prevent bribery offense)',
    'PCAOB Auditing Standard No. 18 (related party transactions)',
    # Pay Equity & Transparency
    'California Equal Pay Act (Labor Code §1197.5)',
    'New York Pay Equity Law (Labor Law §194)',
    'Illinois Equal Pay Act of 2003',
    'Washington Equal Pay and Opportunities Act',
    'Maryland Equal Pay for Equal Work Act',
    'New Jersey Diane B. Allen Equal Pay Act',
    'Massachusetts Equal Pay Act (MEPA)',
    'Paycheck Fairness Act (proposed federal)',
    'EU Pay Transparency Directive',
    'UK Gender Pay Gap Information Regulations 2017',
    'Australia Workplace Gender Equality Act 2012',
    'Canada Pay Equity Act',
    'OFCCP Directive 2022-01 (pay equity audits)',
    # Leave & Benefits
    'COBRA §601 (continuation of health coverage)',
    'ACA §1557 (nondiscrimination in health programs)',
    'MHPAEA (Mental Health Parity and Addiction Equity Act)',
    'Newborns\' and Mothers\' Health Protection Act',
    'Women\'s Health and Cancer Rights Act (WHCRA)',
    'Michelle\'s Law (dependent student health coverage)',
    'California Paid Family Leave (PFL)',
    'New York Paid Family Leave (PFL)',
    'Washington Paid Family & Medical Leave (PFML)',
    'Massachusetts Paid Family and Medical Leave (PFML)',
    'New Jersey Family Leave Insurance (FLI)',
    'San Francisco Paid Parental Leave Ordinance',
    # Discrimination & Harassment
    'Pregnancy Discrimination Act (PDA)',
    'Lilly Ledbetter Fair Pay Act of 2009',
    'Civil Rights Act of 1866, 42 U.S.C. §1981',
    'Rehabilitation Act of 1973, Section 503',
    'Vietnam Era Veterans\' Readjustment Assistance Act (VEVRAA)',
    'Title IX of the Education Amendments of 1972',
    'California Fair Employment and Housing Act (FEHA)',
    'New York State Human Rights Law (NYSHRL)',
    'New York City Human Rights Law (NYCHRL)',
    'Illinois Human Rights Act',
    # Labor & Employee Relations
    'Davis-Bacon Act (prevailing wage requirements)',
    'Service Contract Act (SCA)',
    'Walsh-Healey Public Contracts Act (PCA)',
    'Worker Adjustment and Retraining Notification (WARN) Act',
    'Employee Polygraph Protection Act (EPPA)',
    'Fair Credit Reporting Act (FCRA)',
    'Immigration Reform and Control Act (IRCA)',
    'E-Verify Program (federal contractor requirements)',
    'California Labor Code §226 (wage statement requirements)',
    'New York Labor Law §195 (wage notice requirements)',
    'Massachusetts Wage Act',
    # Financial & Corporate Governance
    'Sarbanes-Oxley Act of 2002 (SOX) Section 302',
    'Dodd-Frank Act Section 951 (Say-on-Pay)',
    'Dodd-Frank Act Section 952 (Compensation Committee Independence)',
    'Dodd-Frank Act Section 955 (Hedging Disclosure)',
    'Dodd-Frank Act Section 956 (Incentive Compensation at Financial Institutions)',
    'SEC Regulation S-K Item 407 (corporate governance disclosures)',
    'SEC Regulation S-K Item 404 (related party transactions)',
    'SEC Rule 14a-8 (shareholder proposals)',
    'NYSE Listed Company Manual Section 303A (corporate governance standards)',
    'NASDAQ Listing Rule 5600 Series (board and committee requirements)',
    'COSO Internal Control – Integrated Framework (2013)',
    'PCAOB Auditing Standard No. 5 (Audit of Internal Control Over Financial Reporting)',
    'PCAOB Auditing Standard No. 12 (Identifying and Assessing Risks of Material Misstatement)',
    'Hart-Scott-Rodino Antitrust Improvements Act of 1976',
    'Bank Secrecy Act (BSA) / Anti-Money Laundering (AML)',
    'Office of Foreign Assets Control (OFAC) Sanctions Programs',
    # Data Privacy
    'Virginia Consumer Data Protection Act (VCDPA)',
    'Colorado Privacy Act (CPA)',
    'Utah Consumer Privacy Act (UCPA)',
    'Connecticut Data Privacy Act (CTDPA)',
    'Children\'s Online Privacy Protection Act (COPPA)',
    'Gramm-Leach-Bliley Act (GLBA) Safeguards Rule',
    'Family Educational Rights and Privacy Act (FERPA)',
    'Brazil Lei Geral de Proteção de Dados (LGPD)',
    'Canada Personal Information Protection and Electronic Documents Act (PIPEDA)',
    'China Personal Information Protection Law (PIPL)',
    'ISO/IEC 27001 (Information Security Management)',
    'NIST Cybersecurity Framework',
    'NIST Privacy Framework',
    'PCI DSS (Payment Card Industry Data Security Standard)',
    # Health & Safety
    'OSHA General Duty Clause (Section 5(a)(1))',
    'OSHA Hazard Communication Standard (HCS)',
    'OSHA Bloodborne Pathogens Standard',
    'OSHA Emergency Action Plan Standard',
    'California Division of Occupational Safety and Health (Cal/OSHA) standards',
    'Federal Mine Safety and Health Act of 1977 (Mine Act)',
    # International & Cross-Border
    'EU Whistleblower Protection Directive',
    'UK Corporate Governance Code',
    'Germany Corporate Governance Code (Deutscher Corporate Governance Kodex)',
    'Japan Corporate Governance Code',
    'OECD Guidelines for Multinational Enterprises',
    'UN Guiding Principles on Business and Human Rights',
    'Modern Slavery Act 2015 (UK)',
    'California Transparency in Supply Chains Act',
    'EU Corporate Sustainability Reporting Directive (CSRD)',
    'EU Sustainable Finance Disclosure Regulation (SFDR)',
    # Additional Tax & Executive Comp
    'IRC §401(k) (qualified cash or deferred arrangements)',
    'IRC §403(b) (tax-sheltered annuity plans)',
    'IRC §457(b) (deferred compensation plans of state and local governments and tax-exempt organizations)',
    'IRC §83(b) (election to include property transferred for services in gross income)',
    'IRC §3121(v)(2) (FICA taxation of nonqualified deferred compensation)',
    'IRC §4999 (excise tax on excess parachute payments)',
    'IRC §4960 (excise tax on excess tax-exempt organization executive compensation)',
    'SEC Rule 16b-3 (transactions between an issuer and its officers or directors)',
    'SEC Regulation FD (Fair Disclosure)',
    'SEC Pay Versus Performance Disclosure Rules (Item 402(v) of Reg S-K)',
    # Emerging Tech & AI
    'Illinois Biometric Information Privacy Act (BIPA)',
    'EU AI Act (proposed)',
    'NIST AI Risk Management Framework',
    'Algorithm-Driven Decision Tools (ADDT) Law (DC)',
    'Automated Employment Decision Tools Law (NYC Local Law 144)',
]

_COMPLIANCE_FINDING_TYPES = [
    'Compensation equity variance exceeds 15% threshold across protected class segments',
    'Job classification accuracy below 92% benchmark for exempt/non-exempt designations',
    'Benefit administration inconsistency identified across {regions} regional offices',
    'Performance rating distribution deviates from expected normal curve by >{pct}%',
    'Overtime eligibility misclassification affecting {n} employees in applicable job families',
    'Pay grade overlap exceeding policy threshold between consecutive salary bands',
    'Documentation gaps in personnel files for {n} employees — required for audit trail compliance',
    'Incentive compensation recoupment policy not applied consistently across business units',
    'Total rewards benchmarking data >24 months old — refresh required under {frameworks}',
    'Variable pay accrual methodology inconsistent with {frameworks} disclosure requirements',
    'Equity grant documentation incomplete for awards issued during the review period',
    'Employee acknowledgment records incomplete for revised compensation policy (v{doc_version})',
    'Offer letter language inconsistent with executed employment agreements for {n} hires',
    'Leave administration records incomplete or contradictory for FMLA-qualifying events',
    'Annual merit increase administration varied by >{pct}% across comparable job families',
    # Pay Equity & Transparency
    'Salary range disclosure missing from {pct}% of job postings in jurisdictions requiring it',
    'Salary history inquiries identified in {n} interview records from prohibited jurisdictions',
    'EEO-1 Component 2 data submission contains statistically improbable values for {n} job categories',
    'UK Gender Pay Gap report for {year} uses incorrect methodology for calculating quartiles',
    'OFCCP Directive 2022-01 pay equity audit not conducted for the {year} affirmative action plan cycle',
    'Pay transparency report for {regions} does not meet EU Pay Transparency Directive requirements',
    # Leave & Benefits
    'COBRA election notices for {n} qualifying events sent {pct}% later than the statutory 14-day deadline',
    'ACA Form 1095-C contains incorrect offer of coverage codes for {n} full-time employees',
    'Mental Health Parity (MHPAEA) non-quantitative treatment limitation (NQTL) analysis is incomplete for the current plan year',
    'FMLA eligibility calculation incorrectly excludes prior service for {n} re-hired employees',
    'USERRA reemployment rights not properly communicated to {n} employees on military leave',
    '401(k) plan loan administration not in compliance with plan document provisions for {n} participants',
    'Summary Plan Descriptions (SPDs) for {n} benefit plans have not been updated to reflect recent plan amendments',
    # Discrimination & Harassment
    'Mandatory anti-harassment training not completed by {pct}% of managers in {regions} as required by state law',
    'ADA interactive process not documented for {n} accommodation requests reviewed during the period',
    'Genetic information (GINA) requested on post-offer medical questionnaire for {n} new hires',
    'ADEA waiver and release agreements for {n} separating employees lack required OWBPA language',
    'Pregnancy Discrimination Act (PDA) accommodation process is not formally documented or consistently applied',
    # Labor & Employee Relations
    'I-9 form completion and verification process deviates from USCIS requirements for {n} new hires',
    'Fair Credit Reporting Act (FCRA) pre-adverse action notices not provided to {n} candidates',
    'Final pay for {n} terminated employees in California not provided within the statutory timeframe',
    'Meal and rest break premiums not paid for {n} non-exempt employees with missed breaks in {regions}',
    'Employee handbook contains policies that could be construed as chilling NLRA-protected concerted activity',
    'Wage statements for {n} employees in {regions} are missing required information under state law',
    'E-Verify case for {n} new hires not initiated within the three-day post-hire requirement',
    # Financial & Corporate Governance
    'SOX 302 sub-certification process for HR controls is not consistently documented or retained',
    'Dodd-Frank §954 clawback policy not triggered for {n} executives following a material financial restatement',
    '§162(m) covered employee list for {year} is incomplete, omitting {n} former executives',
    'Golden parachute payment calculations for {n} executives do not align with §280G regulations',
    'Proxy statement CD&A omits required discussion of {frameworks} performance metrics',
    '10b5-1 trading plan for {n} executives lacks required cooling-off period under amended rules',
    'NYSE/NASDAQ board independence requirements not met for {n} committee members',
    'PCAOB Auditing Standard No. 18 requires disclosure of related party transactions not found in the {year} 10-K',
    # Data Privacy
    'GDPR record of processing activities (Article 30) is incomplete for HR data systems',
    'Data Processing Agreement (DPA) not in place with {n} third-party HR vendors processing EU employee data',
    'HIPAA privacy notice not provided to {n} new enrollees in the group health plan',
    'CCPA/CPRA "Do Not Sell" link not conspicuously posted on internal and external career portals',
    'Data subject access requests under GDPR have an average response time of {n} days, exceeding the statutory limit',
    'Personal Health Information (PHI) transmitted via unencrypted email for {n} benefits claims',
    # Health & Safety
    'OSHA 300A summary not certified by a company executive for {regions} establishments',
    'OSHA Hazard Communication Standard (HCS) training records are incomplete for {n} employees',
    'Emergency Action Plan not reviewed or updated in the last 12 months for the {regions} facility',
    # International & Cross-Border
    'FCPA due diligence documentation missing for {n} third-party agents in high-risk jurisdictions',
    'UK Bribery Act "adequate procedures" defense not supported by current training records for {regions} UK-based staff',
    'Modern Slavery Act statement for {year} lacks sufficient detail on supply chain due diligence measures',
    'EU Whistleblower Protection Directive transposition into local country policies is incomplete for {regions} EU member states',
    # Emerging Tech & AI
    'Automated employment decision tool bias audit not conducted as required by NYC Local Law 144',
    'Illinois Biometric Information Privacy Act (BIPA) written consent not obtained from {n} employees prior to collection of fingerprints for timekeeping',
    'EU AI Act readiness assessment has not been initiated, despite the use of {n} AI-driven tools in EU hiring processes',
    # Additional General Compliance
    'Independent contractor classification for {n} workers does not meet the economic realities test under FLSA',
    'ERISA plan documents not furnished to {n} participants within 30 days of written request',
    '401(k) plan fiduciary committee has not met in {n} quarters, violating its charter',
    'OFCCP affirmative action plan goals not updated for the current plan year',
    'WARN Act notification timeline not met for {n} affected employees in {regions} reduction-in-force',
    'System access for {n} terminated employees not revoked within the 24-hour SLA defined by IT security policy',
    'Stock option grant approval dates do not align with board meeting minutes for {n} awards',
    'Non-compete agreements are in use for {n} employees in jurisdictions where they are unenforceable',
    'Expense reimbursement for {n} remote employees does not comply with state-specific requirements',
    'Commission plan agreements for {n} sales employees are not in writing as required by state law',
    'Internal controls over incentive compensation calculation have a material design deficiency',
    'Union information requests for bargaining unit data were not fulfilled within the required timeframe',
    'Succession planning materials for {n} leadership roles contain language suggesting potential age-based bias',
    'Reasonable accommodation for religious practices not documented for {n} employee requests',
    'Performance improvement plan documentation for {n} employees suggests potential disability-based bias',
]

_COMPLIANCE_RISK_LEVELS = ['HIGH', 'MEDIUM', 'LOW', 'INFORMATIONAL']

_COMPLIANCE_STATUSES = ['OPEN', 'IN PROGRESS', 'REMEDIATED', 'DEFERRED', 'MONITORING']

_COMPLIANCE_SCOPE_TEMPLATES = [
    'This review was conducted at the request of {org} executive leadership to assess compliance '
    'with applicable federal and state employment regulations governing compensation practices '
    'across {regions} operating locations. The review period covers fiscal year {year} through '
    'Q{q} {endyear}, encompassing {n} active employee records across all job families.',

    'The scope of this engagement encompasses a systematic evaluation of {org}\'s compensation '
    'administration practices against regulatory benchmarks established under applicable law. '
    'Field work was conducted between {date} and {enddate}, with data collection spanning '
    '{n} employee records across {regions} reporting units in the {industry} sector.',

    'ACPWB was engaged by {org} to conduct an independent compliance assessment of workforce '
    'compensation programs following an internal audit exception identified in Q{q} {year}. '
    'This report summarizes findings, risk ratings, and corrective action requirements '
    'for {n} employee records reviewed across {regions} locations.',

    'At the direction of {org}\'s Board of Directors, ACPWB performed a comprehensive compliance '
    'review of total rewards and workforce administration practices. The assessment covered '
    '{n} active and {n2} inactive employee records across {regions} jurisdictions, with '
    'particular attention to the {industry} sector comparator population.',

    'The scope of this review is limited to an assessment of {org}\'s compliance with {frameworks}. '
    'The analysis covers the period from {date} to {enddate} and includes a sample of {n} employee records '
    'from the {industry} business unit. Findings are intended for internal management use only.',

    'This compliance assessment was initiated as part of {org}\'s pre-transaction due diligence for a planned '
    'acquisition in the {industry} sector. The scope covers {n} material compensation and benefits programs '
    'across {regions} legal entities, focusing on potential liabilities and integration risks.',

    'Pursuant to the terms of the {year} regulatory settlement agreement, ACPWB was engaged as an independent '
    'third-party monitor to assess {org}\'s compliance with the corrective action plan. This report covers '
    'the Q{q} {year} review period and evaluates progress on {n} open remediation items.',

    'This report details the findings of a proactive compliance audit focused on pay equity and transparency '
    'laws in {regions} key states of operation for {org}. The review analyzed compensation data for {n} employees '
    'and reviewed {n2} external job postings from the period {date} to {enddate}.',

    'The engagement scope was designed to provide {org}\'s Audit Committee with an independent perspective on '
    'the effectiveness of internal controls over financial reporting for compensation-related processes, '
    'specifically addressing risks related to {frameworks}. The review covered {n} key controls across {regions} process areas.',

    'This document constitutes the final report for the {year} annual HR compliance audit for {org}. The scope '
    'encompassed a review of {n} policies and procedures, a sample of {n2} personnel files, and interviews with '
    '{regions} HR and payroll staff members.',

    'ACPWB conducted a targeted compliance review of {org}\'s executive compensation programs, focusing on '
    'adherence to {frameworks}. The review covered all compensation actions for the top {n} executives '
    'for the fiscal year ending {date}.',

    'The scope of this review was to assess the design and operational effectiveness of {org}\'s compliance '
    'program related to anti-bribery and corruption regulations, including the FCPA and UK Bribery Act. '
    'The assessment included a review of policies, training records for {n} employees, and a sample of {n2} third-party vendor payments.',

    'This report summarizes the results of a data privacy compliance assessment for {org}\'s HR function. '
    'The scope included a data mapping exercise for {n} HR systems and an evaluation of compliance with '
    'GDPR, CCPA/CPRA, and other applicable privacy laws across {regions} jurisdictions.',

    'This compliance review focuses on wage and hour practices for {org}\'s non-exempt workforce in {regions} states. '
    'The assessment included a statistical analysis of timekeeping records for {n} employees and a review of '
    'meal and rest period policies and administration.',

    'As part of {org}\'s ongoing risk management program, ACPWB was engaged to perform a compliance assessment '
    'of the company\'s employee benefits plans under ERISA and ACA. The review covered plan documents, '
    'summary plan descriptions, and Form 5500 filings for the {year} plan year for {n} sponsored plans.',

    'The scope of this review was to validate the accuracy and completeness of {org}\'s EEO-1 Component 2 pay data '
    'submission for the {year} reporting cycle. The assessment involved reconciling source payroll and HRIS data for '
    '{n} employees against the final submission file.',

    'This report outlines findings from a compliance review of {org}\'s leave of absence administration processes. '
    'The scope covered FMLA, ADA, and state-specific leave laws in {regions} states, and included a review of {n} '
    'leave cases from the period {date} to {enddate}.',

    'This assessment was commissioned by {org}\'s legal department to review compliance with NLRA requirements '
    'in employee handbooks, policies, and communications. The review covered {n} policy documents and {n2} '
    'internal communications templates.',

    'The scope of this engagement was to perform an independent audit of {org}\'s automated employment decision tools '
    'for compliance with NYC Local Law 144. The assessment included a bias audit of {n} algorithms used in hiring '
    'and promotion decisions for roles based in New York City.',
]

_COMPLIANCE_METHODOLOGY_TEMPLATES = [
    'Field work was conducted using a stratified random sample of {n} employee records, '
    'selected to ensure representation across job family, pay grade, tenure band, and '
    'protected class category. Statistical analysis employed a 95% confidence interval '
    'with \xb1{pct}% margin of error. All findings are expressed relative to peer benchmarks '
    'drawn from {industry} sector comparators.',

    'Findings were assessed against regulatory thresholds established under applicable law '
    'and ACPWB\'s internal compliance framework (version {doc_version}). Risk ratings reflect '
    'both the likelihood of regulatory examination and the potential financial and reputational '
    'exposure to {org}. Management responses were solicited for all HIGH and MEDIUM findings '
    'and are reproduced verbatim in Section 5.',

    'All data was sourced directly from {org}\'s HRIS and payroll systems as of the review '
    'date. ACPWB performed independent verification of a {pct}% stratified sub-sample. '
    'Findings marked REMEDIATED were confirmed through documented evidence submitted by '
    '{org} prior to report issuance. Findings marked OPEN have not been independently '
    'verified and are reported as described by {org} management.',
    'The methodology for this review was aligned with the PCAOB auditing standards for an audit of internal control over financial reporting. Testing included design effectiveness assessments and operating effectiveness testing for {n} key controls.',
    'A "desk audit" methodology was employed, consisting of a review of policies, procedures, and a sample of {n} transactional records provided by {org}. No on-site fieldwork or employee interviews were conducted as part of this limited-scope engagement.',
    'ACPWB conducted a series of structured interviews with {n} key process owners across HR, Payroll, and Legal to map the end-to-end {frameworks} compliance process. Process maps were validated with management and used to identify potential control gaps.',
    'The assessment utilized a gap analysis methodology, comparing {org}\'s current practices against the requirements of {frameworks} and ACPWB\'s proprietary {industry} sector maturity model. Gaps were categorized by risk level and remediation complexity.',
    'A statistical sampling approach was used to select {n} employee records for detailed testing. The sample size was calculated to provide a 95% confidence level with a 5% margin of error for projecting the error rate across the entire population.',
    'The review of automated decision tools included an algorithmic bias assessment performed on a sandboxed version of the hiring tool. Test data consisting of {n} synthetic candidate profiles was used to evaluate outcomes across demographic subgroups.',
    'ACPWB performed a document review of {n} plan documents, summary plan descriptions, trust agreements, and administrative service agreements to assess compliance with ERISA\'s documentary requirements. Findings were cross-referenced with operational practices.',
    'The methodology for this review included a physical walkthrough of {n} worksites in {regions} to observe safety practices and verify the posting of required OSHA and state-level labor law notices. Observations were documented using a standardized checklist.',
    'A comparative analysis was conducted, benchmarking {org}\'s pay practices for {n} job families against data from {n2} peer companies in the {industry} sector. Peer data was sourced from public proxy filings and proprietary ACPWB compensation surveys.',
    'The review of {org}\'s I-9 compliance included a re-verification of a sample of {n} forms using the E-Verify system and a review of internal audit procedures for form completion and retention.',
    'ACPWB\'s methodology for assessing data privacy compliance included the use of automated data discovery tools to scan {n} network shares and databases for unstructured sensitive employee data. Findings were classified based on the NIST Privacy Framework.',
    'The effectiveness of compliance training was assessed through a combination of training record review, knowledge assessments administered to a sample of {n} employees, and interviews with {n2} managers regarding their understanding of key policies.',
    'A "red team" exercise was conducted to test the effectiveness of controls designed to prevent unauthorized access to sensitive compensation data. The exercise simulated {n} common attack vectors against {n2} target systems.',
    'The methodology for this review was based on the "Three Lines of Defense" model, assessing the roles and effectiveness of business operations (first line), compliance and risk functions (second line), and internal audit (third line) in managing {frameworks} risk.',
    'ACPWB performed a forensic analysis of payroll and timekeeping data for {n} employees to identify anomalies and patterns indicative of non-compliance with FLSA overtime and meal/rest break requirements.',
    'The review included an analysis of {org}\'s governance, risk, and compliance (GRC) tool configuration and usage. The assessment evaluated the effectiveness of the tool in tracking compliance obligations, control testing, and issue remediation for {frameworks}.',
    'A benchmarking methodology was used to compare {org}\'s affirmative action plan goals and outreach efforts against those of {n} peer federal contractors in the {industry} sector and {regions} labor market.',
    'The assessment of §409A compliance included a detailed review of plan documents for documentary compliance and transactional testing of {n} deferral elections and distributions for operational compliance.',
    'ACPWB conducted a simulation of a mass layoff event to test the operational readiness of {org}\'s HR and legal teams to comply with WARN Act notification requirements under a compressed timeline.',
    'The methodology for this review included a "mock audit" simulating an OFCCP compliance evaluation. The process included a request for documents, an on-site investigation (simulated), and interviews with a sample of {n} managers and HR staff.',
    'A "look-back" analysis was performed on {n} terminated employee files from the past 24 months to assess compliance with state-specific final pay timing requirements and the proper handling of accrued vacation payouts.',
    'The review of {org}\'s HIPAA compliance included a security risk analysis based on the NIST SP 800-30 framework, covering {n} systems that create, receive, maintain, or transmit electronic PHI.',
    'ACPWB utilized a decision-tree analysis to evaluate {n} complex employee leave scenarios, testing the consistency and accuracy of FMLA eligibility and entitlement determinations made by the leave administration team.',
    'The assessment of FCPA compliance included a transactional review of a sample of {n} payments to third-party agents in high-risk jurisdictions, tracing payments from invoice to bank records and cross-referencing with due diligence documentation.',
    'A "policy-to-procedure-to-practice" tracing methodology was used to assess the implementation of {org}\'s code of conduct. The review traced {n} policy requirements through their corresponding procedures down to evidence of practical application in a sample of business transactions.',
    'The methodology for this review was aligned with the ISO 37001 standard for anti-bribery management systems. The assessment evaluated {org}\'s program against the standard\'s requirements for leadership, planning, support, operation, and performance evaluation.',
    'ACPWB performed a "stress test" of {org}\'s data subject access request (DSAR) process under CCPA/GDPR, submitting {n} test requests to evaluate response time, completeness, and accuracy.',
    'The review of {org}\'s EEO-1 reporting process included a reconciliation of HRIS headcount and demographic data to the final submitted report for {n} establishments, identifying any discrepancies in data aggregation or mapping.',
    'A "what-if" scenario analysis was conducted to model the potential financial impact of misclassifying {n} independent contractors as employees, including estimates for back taxes, benefits, and potential penalties.',
    'The assessment of {org}\'s compliance with pay transparency laws included a programmatic scan of {n} of the company\'s career pages and third-party job boards to verify the presence and accuracy of salary range disclosures.',
    'ACPWB conducted a "culture of compliance" assessment using a confidential survey administered to {n} employees, supplemented by {n2} leadership interviews. The survey measured perceptions of ethical tone at the top, fear of retaliation, and understanding of compliance resources.',
    'The methodology for reviewing {org}\'s §280G calculations involved building an independent model to replicate the "base amount" and "parachute payment" calculations for the top {n} executives, comparing the results to the company\'s analysis.',
    'A "time and motion" study methodology was used to analyze the work activities of a sample of {n} employees in roles with borderline FLSA exemption status, providing empirical data to support the classification decision.',
    'The review of {org}\'s drug testing program included an analysis of state-specific laws in {regions} states to ensure compliance with regulations regarding marijuana, both medical and recreational, and its impact on workplace policies.',
    'ACPWB performed a "horizontal audit" of the employee onboarding process, tracing a cohort of {n} new hires from offer letter to 90 days of employment to assess the consistency and compliance of all onboarding steps.',
    'The assessment of {org}\'s compliance with the Americans with Disabilities Act (ADA) included a review of {n} job descriptions to ensure essential functions were clearly defined and not unnecessarily exclusionary.',
    'A "document-request-list" simulation was used, mirroring a typical request from a regulatory agency like the DOL or EEOC. {org}\'s ability to produce the requested {n} documents within a 14-day timeframe was assessed.',
    'The methodology for this review included an analysis of {org}\'s insurance coverage for employment-related liabilities (EPLI), assessing policy limits and exclusions in the context of the compliance risks identified in this report.',
    'ACPWB performed a "peer benchmark" of {org}\'s compliance function, comparing its staffing levels, budget, and technology stack to those of {n} similarly sized companies in the {industry} sector.',
    'The review of {org}\'s political law compliance included a reconciliation of lobbying-related expenses reported under the LDA with internal accounting records for the {year} reporting period.',
    'A "root cause analysis" was performed for all HIGH and CRITICAL risk findings. The analysis employed the "5 Whys" technique during interviews with {n} process owners to identify underlying systemic '
    'issues rather than isolated symptoms.',

    'The assessment methodology included: (1) review of {n} relevant policy and procedure documents; '
    '(2) structured interviews with {n2} key personnel in HR, Legal, and Finance; and (3) transactional '
    'testing of a judgmental sample of {n} employee records selected based on risk criteria.',

    'A multiple regression analysis was performed on compensation data for {n} employees to identify '
    'statistically significant pay disparities based on gender and race/ethnicity, controlling for '
    'legitimate, non-discriminatory factors such as job level, tenure, and performance.',

    'Control testing was conducted using a combination of inquiry, observation, inspection of documents, '
    'and re-performance. For each of the {n} key controls in scope, a sample of {n2} transactions was '
    'selected to test operating effectiveness throughout the review period.',

    'The compliance framework used for this assessment is a composite of federal law, state regulations '
    'in {regions} key jurisdictions, and leading industry practices as documented in ACPWB\'s proprietary '
    'compliance database. Each finding is mapped to the specific requirement it contravenes.',

    'ACPWB utilized its proprietary automated compliance testing tool to scan {n} job postings for '
    'adherence to pay transparency disclosure requirements. Manual review was performed on a {pct}% '
    'sample of exceptions to validate tool accuracy.',

    'The methodology for this review was aligned with the COSO Internal Control – Integrated Framework. '
    'Each of the {n} in-scope business processes was evaluated across the five COSO components: Control '
    'Environment, Risk Assessment, Control Activities, Information & Communication, and Monitoring Activities.',

    'A root cause analysis was performed for all HIGH and CRITICAL risk findings. The analysis employed '
    'the "5 Whys" technique during interviews with {n} process owners to identify underlying systemic '
    'issues rather than isolated symptoms.',

    'The review methodology was designed to meet the standards of an independent audit that could be '
    'relied upon by external stakeholders. All workpapers, evidence, and conclusions are maintained '
    'by ACPWB and are available for review by {org}\'s external auditors upon request.',

    'Data analysis was performed using a combination of descriptive statistics, comparative analysis against '
    '{industry} sector benchmarks, and predictive modeling to identify outlier populations. All statistical '
    'tests were conducted at a 95% confidence level.',

    'The assessment included a review of {org}\'s governance and oversight structure for compliance. This '
    'involved reviewing the charters and meeting minutes of the Audit Committee and Compliance Committee '
    'for the {year} fiscal year.',

    'To assess compliance with data privacy regulations, ACPWB conducted a data discovery and mapping exercise, '
    'followed by a gap analysis against the requirements of GDPR and CCPA/CPRA. The review covered {n} '
    'systems containing employee personal data.',

    'The methodology for assessing {frameworks} compliance involved a document review of the plan text and '
    'summary plan description, followed by transactional testing of {n} participant events (e.g., distributions, '
    'loans, and hardship withdrawals) to confirm operational compliance.',
]

_CORRECTIVE_ACTION_TEMPLATES = [
    '{org} shall conduct a full pay equity analysis across all protected class segments '
    'within 45 days of report issuance and submit a remediation plan to ACPWB for review.',

    'Implement a centralized job classification review process with documented approval '
    'workflow. All reclassification decisions must be signed off by the designated owner '
    'and retained for a minimum of 3 years per applicable recordkeeping requirements.',

    'Engage external legal counsel to review variable pay documentation for identified '
    'employees and issue corrected award agreements within 60 days of this report.',

    'Update total rewards benchmarking data using a survey source current within 12 months. '
    'Recalibrate salary ranges against updated market data prior to the next annual '
    'compensation cycle and document the methodology applied.',

    'Conduct manager training on overtime eligibility determination for all supervisors '
    'in affected divisions. Certify completion records and retain for audit trail.',

    'Reconcile benefit enrollment records against payroll deduction data for all '
    '{regions} regional offices. Identify and correct discrepancies within 30 days. '
    'Implement monthly automated reconciliation going forward.',

    'Revise offer letter templates and employment agreement language to ensure consistency. '
    'All active employees with inconsistent documentation should receive amended agreements '
    'within 90 days of this report. Legal sign-off required before distribution.',

    'Establish a formal policy recoupment tracking log and assign ownership to a '
    'designated compliance officer. All recoupment events must be documented within '
    '30 days of trigger event and reported quarterly to the Audit Committee.',

    'Revise the employee handbook to remove language that could be interpreted as restricting employees\' Section 7 rights under the NLRA. Submit revised handbook to labor counsel for review within 60 days.',
    'Implement a mandatory annual training program for all supervisors on the proper classification of employees under the FLSA, with a focus on the duties tests for executive, administrative, and professional exemptions.',
    'Develop and implement a centralized process for creating and posting job requisitions to ensure '
    'all external postings in covered jurisdictions include the required salary range disclosures.',

    'Immediately cease all inquiries into candidate salary history during the hiring process. Update all '
    'recruiter training materials and interview guides to reflect this prohibition.',

    'Perform a root cause analysis of the EEO-1 Component 2 data errors and implement validation controls '
    'in the data preparation process to prevent recurrence in future reporting cycles.',

    'Revise the reduction-in-force (RIF) checklist to include a multi-level legal review of WARN Act '
    'notification triggers and timelines prior to any future mass layoff event.',

    'Conduct a self-audit of all I-9 forms for employees hired in the past 12 months. Correct all identified '
    'errors and provide remedial training to all staff responsible for I-9 verification.',

    'Correct the {n} non-compliant 401(k) plan loans and implement automated controls within the payroll '
    'system to ensure future loan repayments are processed in accordance with the plan document.',

    'Distribute updated HIPAA privacy notices to all new and existing health plan participants within 30 days. '
    'Incorporate notice distribution into the standard new hire and open enrollment onboarding process.',

    'Automate the COBRA notification process to ensure qualifying event notices are generated and mailed '
    'within 5 business days of the triggering event.',

    'Develop and distribute a manager and HR guide on USERRA reemployment rights and obligations. '
    'Conduct mandatory training for all hiring managers and HR business partners.',

    'The Compensation Committee shall convene a special meeting to review the identified financial '
    'restatement and initiate clawback procedures for all applicable executive incentive compensation.',

    'The Tax department shall perform a look-back analysis to identify all non-deductible compensation '
    'under §162(m) for the past three fiscal years and file amended returns if necessary.',

    'Engage a qualified third-party valuation firm to perform §280G calculations for all executives with '
    'change-in-control agreements to ensure accurate accounting for potential excise taxes.',

    'Amend the upcoming proxy statement to include the required disclosures under {frameworks}. '
    'Implement a proxy disclosure checklist to be reviewed by legal counsel prior to all future filings.',

    'All 10b5-1 trading plans for Section 16 officers must be reviewed and pre-cleared by the General Counsel '
    'to ensure compliance with all current SEC requirements, including cooling-off periods.',

    'The HRIS team, in partnership with Legal, shall create and maintain a comprehensive Record of Processing '
    'Activities for all employee data, to be reviewed and updated quarterly.',

    'Conduct a privileged pay equity analysis under the direction of outside counsel to identify and remediate any statistically significant pay disparities based on gender or race/ethnicity.',
    'Update the 401(k) plan\'s investment policy statement (IPS) to reflect current fiduciary best practices and document the committee\'s review of investment performance against the IPS on a quarterly basis.',
    'Immediately suspend the use of the identified automated employment decision tool in New York City '
    'pending the completion of a compliant bias audit and public disclosure of the results.',

    'Conduct a market pricing review for all roles based in Colorado to establish and document '
    'equitable pay ranges for compliance with the Equal Pay for Equal Work Act.',

    'Implement a risk-based third-party due diligence program for all new international agents and '
    'consultants, including mandatory FCPA compliance certifications and background checks.',

    'The designated management official at each of the {regions} establishments shall personally certify '
    'the OSHA 300A summary by the February 1st deadline annually.',

    'The 401(k) plan fiduciary committee shall formalize its vendor review process, including the '
    'issuance of RFPs and documentation of the selection rationale for all plan service providers.',

    'Mandatory ADA interactive process training shall be assigned to all managers. The HR team will '
    'implement a centralized case management system to document all accommodation requests and outcomes.',

    'All severance and release agreements for employees aged 40 and over must be reviewed by legal '
    'counsel to ensure compliance with all OWBPA requirements, including consideration periods.',

    'An audit of all {org} facilities shall be conducted to ensure the current, compliant version '
    'of the NLRA rights poster is displayed in a conspicuous location accessible to all employees.',

    'The annual affirmative action plan shall be updated to include specific, measurable, and time-bound '
    'goals for all job groups with identified underutilization. Progress will be reviewed quarterly.',

    'All nonqualified deferred compensation plan documents shall be amended to conform with the '
    'distribution and election timing rules under IRC §409A, effective the next plan year.',

    'The methodology for identifying the median employee for the CEO pay ratio disclosure shall be '
    'formally documented and reviewed by the Compensation Committee annually.',

    'Implement a "Do Not Sell or Share My Personal Information" link on all internal and external '
    'career pages and establish a process for responding to employee data requests under CCPA/CPRA.',

    'All HR and benefits personnel shall complete mandatory HIPAA training. The IT department will '
    'implement technical safeguards (e.g., email encryption) to protect all outbound transmissions of PHI.',

    'The HRIS system logic for FMLA eligibility shall be updated to correctly account for hours worked '
    'in the 12 months preceding the leave request, including for re-hired employees.',

    'Establish a formal process for reviewing and responding to all accommodation requests under the ADA and PWFA, including standardized forms for documenting the interactive process.',
    'Perform a comprehensive audit of all benefit plan documents to ensure they are up-to-date, signed, and compliant with the latest ERISA and IRS requirements. Distribute updated SPDs to all participants.',
    'Review and revise all pre-employment and post-offer medical questionnaires to remove any '
    'questions related to family medical history or genetic information.',

    'The Compensation Committee shall establish a formal, documented process for reviewing and approving '
    'all incentive plan payouts for executive officers prior to payment.',

    'The Total Rewards team will conduct a comprehensive review of all job postings to ensure salary '
    'range information is included where required by law, with a target of 100% compliance within 60 days.',

    'Recruiting and HR staff will undergo mandatory training on prohibited salary history inquiries, '
    'with training completion tracked and certified.',

    'A data validation and scrubbing process will be implemented for all future EEO-1 data submissions, '
    'including outlier analysis and logical checks for all reported pay data.',

    'The Legal and HR departments will jointly develop a RIF playbook that includes a detailed '
    'WARN Act compliance checklist and a multi-level review process.',

    'A quarterly self-audit of I-9 forms for all new hires will be instituted, with results reported '
    'to the Chief Compliance Officer.',

    'The benefits administration platform will be reconfigured to prevent plan loans that do not '
    'conform to the specific terms of the 401(k) plan document.',

    'The annual open enrollment communications package will be updated to include the most current '
    'HIPAA privacy notice as a mandatory acknowledgment item for all enrolling employees.',

    'The COBRA administration vendor\'s performance will be audited quarterly to ensure compliance '
    'with all statutory notification deadlines, with financial penalties for non-compliance.',

    'A formal USERRA compliance protocol will be integrated into the military leave of absence process, '
    'including pre-leave counseling and post-leave reemployment planning.',

    'The Compensation Committee charter will be amended to explicitly assign responsibility for '
    'overseeing the enforcement of the company\'s executive compensation clawback policy.',

    'The annual proxy disclosure process will be updated to include a specific §162(m) compliance '
    'review step, to be signed off by both the Tax and Legal departments.',

    'A third-party specialist will be engaged annually to provide §280G "golden parachute" calculations '
    'for all named executive officers in connection with any potential change-in-control scenario.',

    'The proxy drafting process will include a "Disclosure Checklist" to ensure all required tables '
    'and narrative discussions under SEC Regulation S-K Item 402 are included and accurate.',

    'The General Counsel\'s office will maintain a log of all 10b5-1 trading plans, including adoption '
    'dates, cooling-off periods, and termination dates, to ensure ongoing compliance.',

    'A cross-functional team from HR, IT, and Legal will be established to own and maintain the '
    'company\'s GDPR Record of Processing Activities, with updates required for any new system or process.',

    'The company will engage an independent auditor to conduct a bias audit of its automated hiring '
    'tools before the statutory deadline for NYC Local Law 144.',

    'A project will be initiated to map all job titles to the new job architecture and establish '
    'compliant pay ranges for all roles, with a priority on roles based in Colorado.',

    'The vendor onboarding process will be updated to include a mandatory, risk-based anti-corruption '
    'due diligence step for all international third-party intermediaries.',

    'A senior executive will be formally designated at each company establishment to be responsible '
    'for the review and certification of the annual OSHA 300A summary.',

    'The retirement plan committee will adopt a formal investment policy statement (IPS) and document '
    'its review of plan investment performance against the IPS on a quarterly basis.',

    'All managers will be required to complete a biennial training course on disability awareness and '
    'the ADA interactive process, with completion rates reported to the Chief Diversity Officer.',

    'Succession planning materials and talent review discussions will be reviewed by legal counsel '
    'to identify and mitigate any potential risk of age-based discrimination.',

    'A standardized process will be implemented to respond to all formal union information requests '
    'within 10 business days, with all responses reviewed by labor relations counsel.',

    'The company will partner with three new community organizations to enhance its recruitment '
    'outreach for underutilized job groups identified in the affirmative action plan.',

    'The deferred compensation plan committee will conduct a full review of all plan operations to '
    'ensure strict compliance with the documentary and operational requirements of IRC §409A.',

    'The process for identifying the median employee will be enhanced to include a statistical sampling '
    'approach, with the methodology and results documented and presented to the Audit Committee.',

    'The company website and all subsidiary career sites will be updated to include a compliant '
    '"Your Privacy Choices" link, and a dedicated internal team will be trained to handle data subject requests.',

    'The IT department will roll out mandatory end-to-end encryption for all emails containing '
    'sensitive employee data, including benefits and health information.',

    'The HRIS will be audited and reconfigured to ensure FMLA leave entitlement calculations are '
    '100% compliant with the "rolling 12-month period measured backward" method.',

    'The company will implement a centralized system for tracking all employee complaints of harassment, discrimination, and retaliation, ensuring all investigations are conducted and documented in a timely and consistent manner.',
    'The company shall engage a third-party firm to conduct an independent, privileged assessment of its corporate safety culture, focusing on leadership commitment, employee involvement, and hazard identification.',
    'The compliance department will develop and implement a risk-based monitoring and testing program for key HR compliance areas, with results reported to the Audit Committee on a quarterly basis.',
    'All managers and HR personnel involved in the hiring process must complete training on compliant interviewing techniques, including prohibitions on questions related to age, disability, national origin, and other protected characteristics.',
    'All hiring-related forms and questionnaires will be audited to ensure no questions solicit '
    'genetic information, in compliance with GINA.',
]

_MGMT_RESPONSE_TEMPLATES = [
    '{org} accepts this finding. A corrective action plan has been developed and '
    'assigned to the designated owner. Implementation is targeted for completion '
    'by the due date indicated. Progress will be reported to ACPWB on a monthly basis '
    'until formal closure.',

    'Management acknowledges the finding and has initiated a preliminary review. '
    'The designated owner will oversee remediation. {org} requests a 30-day extension '
    'to the original due date to allow for thorough analysis across all affected locations.',

    'The condition identified has been partially remediated. {org} has addressed '
    'the majority of the affected population and is on track for full closure '
    'by the due date. A status update will be provided at the next scheduled review.',

    'Management disputes the risk classification of this finding. While acknowledging '
    'the documentation gap, {org} believes the underlying practices are compliant with '
    'applicable requirements. A formal response memorandum has been filed with the '
    'engagement team under separate cover.',

    'This finding has been remediated as of the report date. Supporting documentation, '
    'including revised policy language and employee acknowledgment records, has been '
    'provided to ACPWB under separate cover. {org} requests formal closure of this item.',

    'Management concurs with the finding and has already initiated corrective action '
    'independent of this review. The designated owner will provide written confirmation '
    'of full remediation no later than the due date specified above.',

    'Management has reviewed the finding and determined that the current practice, while inconsistent with the recommendation, is based on a reasonable interpretation of an ambiguous regulation. We will continue to monitor regulatory developments but plan no immediate changes.',
    'The finding is acknowledged. This issue is part of a larger, previously identified process deficiency that is being addressed through a company-wide transformation project (Project {project_name}). We request this finding be tracked as part of that larger initiative.',
    'The finding is acknowledged. A cross-functional task force has been assembled to address '
    'the root cause. An initial project plan and timeline will be submitted to ACPWB within 15 days.',

    'While we accept the factual basis of the finding, we have performed a risk assessment and '
    'formally accepted the residual risk. This decision has been documented and approved by the '
    'Chief Risk Officer. No further corrective action is planned at this time.',

    'This issue has been escalated to the executive steering committee for prioritization and '
    'resource allocation. A formal response will be provided following the committee\'s next '
    'scheduled meeting on {date}.',

    'The identified control weakness has been contained through the implementation of a temporary '
    'manual workaround. A permanent, automated control is in development with a target '
    'implementation date of Q{q} {endyear}.',

    'Remediation is in progress. The initial phase, covering {pct}% of the affected population, '
    'is complete. The remaining population will be addressed in Phase 2, scheduled to '
    'conclude by the due date.',

    'We are awaiting final evidence of remediation from the third-party vendor responsible for '
    'this process. We have received their commitment to provide the necessary documentation '
    'by {date} and will forward it to ACPWB upon receipt.',

    'Management has reviewed the finding and concurs with the recommended corrective action. '
    'The action has been entered into our GRC system, and automated reminders have been '
    'assigned to the owner.',

    'The finding is valid. A budget of ${n}K has been approved to procure the necessary '
    'technology to automate this control. Implementation is projected to take {n2} weeks.',

    'This finding relates to a legacy system that is scheduled for decommissioning on {date}. '
    'Given the impending system retirement, management has accepted the risk for the interim '
    'period and will not invest in remediation for the legacy system.',

    'Corrective action is complete. The revised policy was approved by the Board on {date} '
    'and has been communicated to all employees. A copy of the communication and the '
    'approved policy is attached as evidence.',

    'We agree with the finding. A request for proposal (RFP) has been issued to select a '
    'vendor to assist with the implementation of the corrective action. The vendor selection '
    'is expected to be complete by {date}.',

    'The finding is noted. This issue stems from a lack of clarity in the regulatory guidance. '
    'We have submitted a formal inquiry to the regulatory agency and are awaiting their '
    'response before proceeding with a final corrective action.',

    'Management has implemented the primary corrective action. We are now in a monitoring phase '
    'to ensure the new process is operating effectively. We will provide a final report '
    'after 90 days of monitoring.',

    'This finding has been superseded by a subsequent internal audit finding which addresses '
    'the same root cause with a more comprehensive corrective action plan. We request this '
    'finding be closed and tracked under the new internal audit reference number.',

    'The finding is accepted. The required training has been developed and assigned to all '
    '{n} affected employees with a mandatory completion date of {date}. Completion rates '
    'will be tracked and reported to the compliance department.',

    'Management has performed a cost-benefit analysis and determined that the cost of full '
    'remediation exceeds the potential risk exposure. We have implemented mitigating controls '
    'and formally accepted the residual risk.',

    'The finding is valid and has been added to our compliance remediation backlog. It has been '
    'prioritized for action in the next fiscal quarter, subject to resource availability.',

    'We concur with the finding. The process has been redesigned, and user acceptance testing '
    'is scheduled for the week of {date}. Full implementation will follow immediately after '
    'successful testing.',

    'The finding is acknowledged. This issue was caused by a temporary system outage on {date}. '
    'The system has been restored, and the backlog of affected transactions has been cleared. '
    'We consider this matter closed.',

    'Management agrees with the assessment. A new hire has been approved to serve as the '
    'dedicated process owner for this compliance area, with an expected start date of {date}.',

    'The finding is valid. We have engaged our external auditors to perform an independent '
    'review and provide recommendations for a permanent solution. Their report is expected '
    'within {n} days.',

    'This finding is a duplicate of a previously identified issue that is already being tracked '
    'under reference number {audit_ref}. We request this finding be merged with the existing one.',

    'Management accepts the finding. The relevant system configuration has been updated, and '
    'the change will be deployed to production during the next scheduled release on {date}.',

    'The finding is acknowledged. The issue affects a business unit that is being divested. '
    'The corrective action will be the responsibility of the new owner post-transaction close, '
    'which is anticipated on or before {date}.',
    'We concur with the finding but disagree with the "HIGH" risk rating. Based on our internal analysis, the likelihood of occurrence is low, and the potential impact is minimal. We have re-rated this finding as "LOW" in our internal GRC system and will remediate it as resources permit.',
    'Management has implemented a compensating control that reduces the residual risk of this finding to an acceptable level. The original control deficiency will be addressed as part of the Q{q} {endyear} system upgrade. We request this finding be downgraded to "INFORMATIONAL".',

    'We concur. A project charter has been approved to address this finding. The project manager '
    'has been assigned, and a kickoff meeting is scheduled for {date}.',

    'The finding is valid. The required report has been developed and is now generated and '
    'distributed automatically on a weekly basis. See attached sample report as evidence.',

    'Management agrees with the finding. The issue has been added to the agenda for the next '
    'quarterly business review with the responsible process owners to ensure accountability.',

    'The finding is accepted. We have implemented a manual control as an interim solution while '
    'a permanent system fix is being developed. The manual control is performed daily by the '
    'compliance team.',
]
