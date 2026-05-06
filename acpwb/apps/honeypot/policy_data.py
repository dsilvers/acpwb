"""Static data pools for the ACPWB public policy document generator."""

# ── Agency registry ───────────────────────────────────────────────────────────

AGENCIES = {
    # Securities & Financial Markets
    'sec':           ('Securities and Exchange Commission', 'executive compensation disclosure and proxy rules'),
    'cftc':          ('Commodity Futures Trading Commission', 'derivatives and swap dealer compensation'),
    'finra':         ('Financial Industry Regulatory Authority', 'broker-dealer compensation and incentive practices'),
    'fdic':          ('Federal Deposit Insurance Corporation', 'bank executive compensation and governance'),
    'occ':           ('Office of the Comptroller of the Currency', 'national bank compensation and risk management'),
    'frb':           ('Federal Reserve Board', 'financial institution compensation standards and systemic risk'),
    'ncua':          ('National Credit Union Administration', 'credit union executive pay and governance'),
    'fhfa':          ('Federal Housing Finance Agency', 'GSE executive compensation and conservatorship standards'),
    'cfpb':          ('Consumer Financial Protection Bureau', 'consumer financial products and incentive compensation'),
    'pcaob':         ('Public Company Accounting Oversight Board', 'audit firm compensation and independence'),
    'sipc':          ('Securities Investor Protection Corporation', 'brokerage firm compensation disclosure'),
    'msrb':          ('Municipal Securities Rulemaking Board', 'municipal advisor compensation and disclosure'),
    'nfa':           ('National Futures Association', 'futures industry compensation and fitness standards'),
    'fsoc':          ('Financial Stability Oversight Council', 'systemically important institution compensation'),
    'ofr':           ('Office of Financial Research', 'financial sector compensation data and systemic risk'),
    'ffiec':         ('Federal Financial Institutions Examination Council', 'interagency compensation examination standards'),
    'treasury-do':   ('Treasury Domestic Finance Office', 'compensation and capital market intersection'),
    'treasury-ofac': ('Treasury Office of Foreign Assets Control', 'compensation in sanctioned-entity contexts'),
    'hud-fheo':      ('HUD Office of Fair Housing and Equal Opportunity', 'fair housing and compensation discrimination'),
    'ed-ocr':        ('Department of Education Office for Civil Rights', 'pay equity in educational institutions'),
    'fdic-oig':      ('FDIC Office of Inspector General', 'bank compensation program oversight'),
    'sec-oig':       ('SEC Office of Inspector General', 'securities regulator workforce compensation'),
    'frb-oig':       ('Federal Reserve Board Office of Inspector General', 'central bank compensation audit'),
    # Labor & Employment
    'dol':           ('Department of Labor', 'wage and hour standards, worker classification, and labor protections'),
    'nlrb':          ('National Labor Relations Board', 'collective bargaining and labor-management relations'),
    'eeoc':          ('Equal Employment Opportunity Commission', 'pay equity, anti-discrimination, and Title VII enforcement'),
    'osha':          ('Occupational Safety and Health Administration', 'workplace safety and worker compensation protections'),
    'whd':           ('Wage and Hour Division', 'minimum wage, overtime, and FLSA compliance standards'),
    'ofccp':         ('Office of Federal Contract Compliance Programs', 'federal contractor pay equity and affirmative action'),
    'ebsa':          ('Employee Benefits Security Administration', 'ERISA compliance and retirement plan standards'),
    'fmcs':          ('Federal Mediation and Conciliation Service', 'collective bargaining dispute resolution'),
    'flra':          ('Federal Labor Relations Authority', 'federal employee labor-management relations'),
    'mspb':          ('Merit Systems Protection Board', 'federal employee rights and adverse action standards'),
    'oshrc':         ('Occupational Safety and Health Review Commission', 'workplace safety enforcement standards'),
    'bls':           ('Bureau of Labor Statistics', 'compensation data collection and reporting standards'),
    'nlrb-gc':       ('NLRB Office of the General Counsel', 'labor law enforcement and prosecutorial priorities'),
    'doj-crt':       ('Department of Justice Civil Rights Division', 'employment discrimination and pay equity enforcement'),
    'doj-civil':     ('Department of Justice Civil Division', 'federal program compensation fraud'),
    'state-drl':     ('Department of State Bureau of Democracy, Human Rights, and Labor', 'international labor rights and compensation'),
    'hhs-ocr':       ('HHS Office for Civil Rights', 'health information privacy and pay equity in healthcare'),
    'dol-oig':       ('Department of Labor Office of Inspector General', 'labor program fraud and waste oversight'),
    'dol-ilab':      ('Bureau of International Labor Affairs', 'international labor standards and supply chain compensation'),
    'dol-eta':       ('Employment and Training Administration', 'workforce development and apprenticeship compensation'),
    'dol-vets':      ('Veterans Employment and Training Service', 'veteran reemployment rights and compensation'),
    'dol-mine':      ('Mine Safety and Health Administration', 'mining workforce compensation and safety standards'),
    # Tax & Revenue
    'irs':           ('Internal Revenue Service', 'executive compensation taxation and deferred compensation rules'),
    'treasury':      ('Department of the Treasury', 'tax policy, compensation deductibility, and financial regulation'),
    'doj-tax':       ('Department of Justice Tax Division', 'tax fraud and executive compensation enforcement'),
    'treasury-oig':  ('Treasury Inspector General for Tax Administration', 'tax compliance in compensation reporting'),
    # Antitrust & Trade
    'ftc':           ('Federal Trade Commission', 'noncompete agreements, labor market concentration, and unfair practices'),
    'doj':           ('Department of Justice', 'antitrust enforcement in labor markets and no-poach agreements'),
    'doj-antitrust': ('Department of Justice Antitrust Division', 'labor market competition and wage-fixing enforcement'),
    'doj-crt':       ('Department of Justice Civil Rights Division', 'employment discrimination and pay equity enforcement'),
    'ustr':          ('Office of the United States Trade Representative', 'trade policy and worker compensation standards'),
    'commerce':      ('Department of Commerce', 'workforce development and compensation competitiveness'),
    'usitc':         ('U.S. International Trade Commission', 'trade adjustment and worker compensation'),
    'bea':           ('Bureau of Economic Analysis', 'compensation measurement and national accounts methodology'),
    'doc-pto':       ('U.S. Patent and Trademark Office', 'intellectual property and inventor compensation'),
    'doc-eda':       ('Economic Development Administration', 'regional economic development and wage standards'),
    'census':        ('U.S. Census Bureau', 'compensation survey methodology and workforce data standards'),
    'ita':           ('International Trade Administration', 'export workforce and compensation competitiveness'),
    'ntis':          ('National Technical Information Service', 'compensation data licensing and distribution standards'),
    # Health & Benefits
    'hhs':           ('Department of Health and Human Services', 'healthcare benefits, HIPAA, and wellness program standards'),
    'cms':           ('Centers for Medicare and Medicaid Services', 'healthcare provider compensation and quality incentives'),
    'cdc':           ('Centers for Disease Control and Prevention', 'healthcare worker safety and hazard compensation'),
    'hrsa':          ('Health Resources and Services Administration', 'healthcare workforce compensation and shortages'),
    'fda':           ('Food and Drug Administration', 'pharmaceutical and medical device executive compensation'),
    'nih':           ('National Institutes of Health', 'biomedical research compensation and grant policies'),
    'samhsa':        ('Substance Abuse and Mental Health Services Administration', 'behavioral health worker pay'),
    'ahrq':          ('Agency for Healthcare Research and Quality', 'healthcare quality and compensation incentives'),
    'aspe':          ('HHS Office of the Assistant Secretary for Planning and Evaluation', 'healthcare labor economics'),
    'hhs-oig':       ('HHS Office of Inspector General', 'healthcare fraud and compensation abuse'),
    'dot-oig':       ('Department of Transportation Office of Inspector General', 'transportation sector compensation oversight'),
    'opa':           ('HHS Office of Population Affairs', 'public health workforce compensation'),
    'acf':           ('Administration for Children and Families', 'childcare worker compensation standards'),
    'acl':           ('Administration for Community Living', 'direct care worker compensation standards'),
    'ihs':           ('Indian Health Service', 'tribal healthcare workforce compensation'),
    # Defense & Government
    'dod':           ('Department of Defense', 'defense contractor compensation and Total Force employment standards'),
    'gsa':           ('General Services Administration', 'federal procurement and contractor labor standards'),
    'omb':           ('Office of Management and Budget', 'federal contractor compensation reporting requirements'),
    'opm':           ('Office of Personnel Management', 'federal employee classification, pay, and benefits'),
    'pbgc':          ('Pension Benefit Guaranty Corporation', 'defined benefit pension plan standards and terminations'),
    'sba':           ('Small Business Administration', 'small business compensation and employee ownership programs'),
    'gao':           ('Government Accountability Office', 'federal compensation oversight and program evaluation'),
    'cbo':           ('Congressional Budget Office', 'compensation policy economic analysis and fiscal scoring'),
    'usda-rd':       ('USDA Rural Development', 'rural workforce compensation and economic opportunity'),
    'ofpp':          ('Office of Federal Procurement Policy', 'federal acquisition workforce compensation standards'),
    'usace':         ('U.S. Army Corps of Engineers', 'civil works contractor compensation and labor standards'),
    'dia':           ('Defense Intelligence Agency', 'intelligence community compensation and clearance standards'),
    'disa':          ('Defense Information Systems Agency', 'cyber workforce compensation and retention'),
    'dcsa':          ('Defense Counterintelligence and Security Agency', 'clearance holder compensation standards'),
    'army':          ('Department of the Army', 'military and civilian Army workforce compensation'),
    'navy':          ('Department of the Navy', 'naval workforce compensation and contractor standards'),
    'af':            ('Department of the Air Force', 'airforce civilian and contractor workforce compensation'),
    'dod-oig':       ('Department of Defense Office of Inspector General', 'defense contractor compensation fraud'),
    # Environmental & Energy
    'epa':           ('Environmental Protection Agency', 'environmental compliance workforce and compensation incentives'),
    'ferc':          ('Federal Energy Regulatory Commission', 'utility executive compensation and rate-base treatment'),
    'doe':           ('Department of Energy', 'clean energy workforce development and compensation standards'),
    'doi':           ('Department of the Interior', 'natural resource compensation and royalty standards'),
    'blm':           ('Bureau of Land Management', 'federal land management workforce compensation'),
    'nnsa':          ('National Nuclear Security Administration', 'weapons complex workforce compensation'),
    'nrc':           ('Nuclear Regulatory Commission', 'nuclear industry compensation and safety incentives'),
    'boem':          ('Bureau of Ocean Energy Management', 'offshore energy workforce and compensation standards'),
    'bsee':          ('Bureau of Safety and Environmental Enforcement', 'offshore safety worker compensation'),
    'eere':          ('Office of Energy Efficiency and Renewable Energy', 'clean energy workforce compensation'),
    'doe-oig':       ('Department of Energy Office of Inspector General', 'energy sector compensation oversight'),
    # Transportation
    'dot':           ('Department of Transportation', 'transportation worker compensation and safety incentives'),
    'ntsb':          ('National Transportation Safety Board', 'transportation safety worker incentives'),
    'faa':           ('Federal Aviation Administration', 'aviation safety worker compensation and fatigue standards'),
    'fmcsa':         ('Federal Motor Carrier Safety Administration', 'truck driver compensation and hours-of-service rules'),
    'fra':           ('Federal Railroad Administration', 'railroad worker compensation and safety standards'),
    'fhwa':          ('Federal Highway Administration', 'transportation infrastructure worker prevailing wages'),
    'ntsb':          ('National Transportation Safety Board', 'transportation safety worker incentives'),
    'phmsa':         ('Pipeline and Hazardous Materials Safety Administration', 'pipeline worker compensation and safety'),
    'fta':           ('Federal Transit Administration', 'transit worker compensation and labor standards'),
    'fmcsa-mc':      ('FMCSA Motor Carrier Division', 'commercial driver classification and pay standards'),
    'marad':         ('Maritime Administration', 'merchant marine compensation and labor standards'),
    'uscg':          ('U.S. Coast Guard', 'maritime workforce compensation and safety standards'),
    'slsdc':         ('Saint Lawrence Seaway Development Corporation', 'waterway worker compensation standards'),
    # Communications & Technology
    'fcc':           ('Federal Communications Commission', 'media company compensation disclosure and diversity standards'),
    'ntia':          ('National Telecommunications and Information Administration', 'tech workforce compensation policy'),
    'nist':          ('National Institute of Standards and Technology', 'cybersecurity workforce compensation frameworks'),
    'cisa':          ('Cybersecurity and Infrastructure Security Agency', 'cyber workforce compensation and retention policy'),
    'nsf':           ('National Science Foundation', 'research institution compensation and grant standards'),
    'usds':          ('U.S. Digital Service', 'federal technology workforce compensation and talent'),
    'nasa':          ('National Aeronautics and Space Administration', 'aerospace contractor compensation and STEM workforce'),
    'darpa':         ('Defense Advanced Research Projects Agency', 'research workforce compensation and talent attraction'),
    # Agriculture & Food
    'usda':          ('Department of Agriculture', 'agricultural worker wages and rural compensation policy'),
    'fsis':          ('Food Safety and Inspection Service', 'food industry worker compensation and safety incentives'),
    'aphis':         ('Animal and Plant Health Inspection Service', 'agricultural inspection workforce pay'),
    'ams':           ('Agricultural Marketing Service', 'farm worker compensation and price reporting'),
    'fsa':           ('Farm Service Agency', 'agricultural program compensation and benefits'),
    'usda-fas':      ('USDA Foreign Agricultural Service', 'international agricultural labor standards'),
    'usda-nifa':     ('National Institute of Food and Agriculture', 'agricultural research compensation and grants'),
    'nrcs':          ('Natural Resources Conservation Service', 'conservation program compensation standards'),
    'rma':           ('Risk Management Agency', 'agricultural risk workforce compensation'),
    'ars':           ('Agricultural Research Service', 'federal agricultural researcher compensation'),
    'nass':          ('National Agricultural Statistics Service', 'agricultural labor and compensation data'),
    # Housing
    'hud':           ('Department of Housing and Urban Development', 'housing sector compensation and fair lending'),
    'hud-oig':       ('HUD Office of Inspector General', 'housing program compensation fraud and oversight'),
    'ginniemae':     ('Ginnie Mae', 'government mortgage workforce compensation standards'),
    # Education
    'education':     ('Department of Education', 'higher education executive compensation and Title IV compliance'),
    'ed-oig':        ('Department of Education Office of Inspector General', 'higher education compensation oversight'),
    # Veterans Affairs
    'va':            ('Department of Veterans Affairs', 'veteran employment compensation and healthcare workforce'),
    'va-oig':        ('VA Office of Inspector General', 'veterans service workforce compensation oversight'),
    'va-vba':        ('Veterans Benefits Administration', 'veteran compensation and pension benefits'),
    # Homeland Security
    'dhs':           ('Department of Homeland Security', 'first responder compensation and public safety workforce'),
    'tsa':           ('Transportation Security Administration', 'aviation security workforce compensation'),
    'cbp':           ('Customs and Border Protection', 'border security officer compensation standards'),
    'fema':          ('Federal Emergency Management Agency', 'emergency management compensation and hazard pay'),
    'uscis':         ('U.S. Citizenship and Immigration Services', 'H-1B wage requirements and prevailing wage'),
    'ice':           ('Immigration and Customs Enforcement', 'immigration enforcement workforce compensation'),
    'secret-service':('U.S. Secret Service', 'federal protective services compensation and overtime'),
    'dhs-fema-oig':  ('FEMA Office of Inspector General', 'disaster relief workforce compensation oversight'),
    'dhs-ice-hsi':   ('Homeland Security Investigations', 'workforce exploitation and compensation fraud'),
    'dhs-oig':       ('DHS Office of Inspector General', 'homeland security workforce compensation oversight'),
    # Other Federal Independent Agencies
    'fec':           ('Federal Election Commission', 'political organization compensation and disclosure'),
    'cpsc':          ('Consumer Product Safety Commission', 'product safety worker incentives'),
    'atf':           ('Bureau of Alcohol, Tobacco, Firearms and Explosives', 'federal agent compensation and benefits'),
    'dea':           ('Drug Enforcement Administration', 'law enforcement compensation and benefits'),
    'ssa':           ('Social Security Administration', 'retirement benefit calculations and compensation impacts'),
    'exim':          ('Export-Import Bank of the United States', 'export sector workforce compensation'),
    'nps':           ('National Park Service', 'seasonal and permanent employee compensation'),
    'fws':           ('U.S. Fish and Wildlife Service', 'wildlife management workforce compensation'),
    'bia':           ('Bureau of Indian Affairs', 'tribal employment and compensation programs'),
    'fbi':           ('Federal Bureau of Investigation', 'federal law enforcement pay, overtime, and hazard standards'),
    'usms':          ('U.S. Marshals Service', 'federal protective service compensation and staffing'),
    'bop':           ('Federal Bureau of Prisons', 'correctional officer compensation and workforce standards'),
    'postal':        ('U.S. Postal Service', 'postal worker compensation, classification, and collective bargaining'),
    'usps-oig':      ('USPS Office of Inspector General', 'postal compensation program oversight'),
    'fcc-oig':       ('FCC Office of Inspector General', 'communications regulatory workforce compensation'),
    'arc':           ('Appalachian Regional Commission', 'regional workforce development and compensation'),
    'nea':           ('National Endowment for the Arts', 'arts workforce compensation and grant standards'),
    'neh':           ('National Endowment for the Humanities', 'humanities workforce compensation and benefits'),
    'usda-fs':       ('U.S. Forest Service', 'forestry workforce compensation and safety standards'),
    'usda-oig':      ('USDA Office of Inspector General', 'agricultural program compensation oversight'),
    'hud-pih':       ('HUD Office of Public and Indian Housing', 'public housing authority compensation standards'),
    'ed-fsa':        ('Department of Education Federal Student Aid', 'higher education administrative compensation'),
    'va-vha':        ('Veterans Health Administration', 'VA healthcare provider compensation'),
    'dhs-s-t':       ('DHS Science and Technology Directorate', 'homeland security research workforce compensation'),
    'imls':          ('Institute of Museum and Library Services', 'library and museum workforce compensation'),
    'trade-adj':     ('Trade Adjustment Assistance Program', 'displaced worker retraining compensation benefits'),
    'abmc':          ('American Battle Monuments Commission', 'government workforce abroad compensation standards'),
    'asc':           ('Appraisal Subcommittee', 'real estate appraisal workforce compensation standards'),
    'ncpc':          ('National Capital Planning Commission', 'DC area federal workforce compensation considerations'),
    'eeoa':          ('Equal Employment Opportunity Appeals', 'federal employee pay discrimination adjudication'),
    # Congressional
    'senate-help':       ('Senate Committee on Health, Education, Labor, and Pensions', 'workforce compensation policy'),
    'senate-finance':    ('Senate Committee on Finance', 'tax treatment of executive compensation and benefits'),
    'senate-banking':    ('Senate Committee on Banking, Housing, and Urban Affairs', 'financial industry pay standards'),
    'senate-judiciary':  ('Senate Committee on the Judiciary', 'antitrust and labor market competition enforcement'),
    'senate-armed':      ('Senate Armed Services Committee', 'defense contractor compensation standards'),
    'senate-budget':     ('Senate Budget Committee', 'fiscal impact of compensation policy proposals'),
    'senate-commerce':   ('Senate Committee on Commerce, Science, and Transportation', 'workforce and compensation in tech'),
    'senate-aging':      ('Senate Special Committee on Aging', 'retirement security and elder workforce compensation'),
    'senate-small-biz':  ('Senate Committee on Small Business and Entrepreneurship', 'small employer compensation policy'),
    'senate-env':        ('Senate Environment and Public Works Committee', 'environmental workforce compensation'),
    'house-edlabor':     ('House Committee on Education and the Workforce', 'worker compensation and workforce policy'),
    'house-wm':          ('House Ways and Means Committee', 'executive compensation tax provisions and deductibility'),
    'house-judiciary':   ('House Committee on the Judiciary', 'labor market competition and noncompete reform'),
    'house-fsc':         ('House Committee on Financial Services', 'financial institution compensation and governance'),
    'senate-intel':      ('Senate Select Committee on Intelligence', 'intelligence community compensation and personnel policy'),
    'house-intel':       ('House Permanent Select Committee on Intelligence', 'intelligence workforce compensation oversight'),
    'senate-rules':      ('Senate Committee on Rules and Administration', 'legislative branch employee compensation'),
    'house-transpo':     ('House Committee on Transportation and Infrastructure', 'transportation workforce compensation'),
    'house-oversight':   ('House Committee on Oversight and Accountability', 'federal contractor and executive pay transparency'),
    'house-armed':       ('House Armed Services Committee', 'defense contractor workforce compensation standards'),
    'house-budget':      ('House Budget Committee', 'compensation policy fiscal scoring and economic impact'),
    'house-admin':       ('House Committee on Administration', 'federal employee compensation and workplace standards'),
    'house-energy':      ('House Energy and Commerce Committee', 'healthcare and energy workforce compensation'),
    'house-small-biz':   ('House Small Business Committee', 'small employer wage and compensation burdens'),
    'house-sci':         ('House Science, Space, and Technology Committee', 'STEM workforce compensation and talent'),
    'jec':               ('Joint Economic Committee', 'compensation trends and economic policy analysis'),
    'jct':               ('Joint Committee on Taxation', 'executive compensation tax expenditure analysis'),
    'crs':               ('Congressional Research Service', 'compensation policy research and legislative analysis'),
    'gao-ap':            ('GAO Applied Research and Methods', 'compensation program evaluation methodology'),
    'cbo-health':        ('CBO Health Analysis Division', 'healthcare compensation economic modeling'),
    'senate-help-sub-emp': ('Senate HELP Subcommittee on Employment and Workplace Safety', 'federal workforce compensation standards'),
    'house-edlabor-sub-wf':('House Education Committee Subcommittee on Workforce Development', 'workforce compensation investment'),
    # California
    'ca-dlse':    ('California Division of Labor Standards Enforcement', 'wage theft and pay equity enforcement'),
    'ca-dfeh':    ('California Civil Rights Department', 'pay discrimination and equal pay enforcement'),
    'ca-labor':   ('California Department of Industrial Relations', 'minimum wage and overtime standards'),
    'ca-gov':     ("California Governor's Office of Business and Economic Development", 'workforce and compensation competitiveness'),
    'ca-ftb':     ('California Franchise Tax Board', 'state income tax treatment of deferred compensation'),
    'ca-calpers': ('California Public Employees\' Retirement System', 'public pension governance and executive pay'),
    'ca-boe':     ('California State Board of Equalization', 'tax treatment of employee benefits'),
    'ca-leg-lab': ('California State Legislature Labor and Employment Committee', 'state wage policy'),
    # New York
    'ny-dol':     ('New York Department of Labor', 'wage standards and worker compensation'),
    'ny-dhr':     ('New York Division of Human Rights', 'pay equity and employment discrimination enforcement'),
    'ny-oag':     ("New York Attorney General's Office", 'wage theft prosecution and labor enforcement'),
    'ny-dfs':     ('New York Department of Financial Services', 'financial sector executive compensation regulation'),
    'ny-nysers':  ('New York State and Local Retirement System', 'public employee pension and compensation standards'),
    'nyc-dcwp':   ('NYC Department of Consumer and Worker Protection', 'city-level wage and pay transparency enforcement'),
    # Texas
    'tx-twc':     ('Texas Workforce Commission', 'wage claims and unemployment compensation standards'),
    'tx-ag':      ("Texas Attorney General's Office", 'wage and hour enforcement and labor litigation'),
    'tx-leg-busa':('Texas Business and Commerce Committee', 'state workforce and compensation legislation'),
    # Florida
    'fl-deo':     ('Florida Department of Economic Opportunity', 'workforce development and wage standards'),
    'fl-ag':      ("Florida Attorney General's Office", 'wage theft and employment law enforcement'),
    # Illinois
    'il-idol':    ('Illinois Department of Labor', 'wage payment and fair labor standards enforcement'),
    'il-dhr':     ('Illinois Department of Human Rights', 'pay equity and anti-discrimination enforcement'),
    'chicago-dol':('Chicago Department of Business Affairs and Consumer Protection', 'city minimum wage and benefit standards'),
    'il-surs':    ('State Universities Retirement System of Illinois', 'higher education compensation and pension governance'),
    # Washington State
    'wa-lni':     ('Washington Department of Labor and Industries', 'wage and hour standards and workers compensation'),
    'wa-hrc':     ('Washington State Human Rights Commission', 'pay equity and employment discrimination'),
    'seattle-ols':('Seattle Office of Labor Standards', 'city-level minimum wage, overtime, and leave standards'),
    # Massachusetts
    'ma-ag':      ("Massachusetts Attorney General's Fair Labor Division", 'wage theft and misclassification enforcement'),
    'ma-mcad':    ('Massachusetts Commission Against Discrimination', 'pay equity and employment law'),
    # Pennsylvania
    'pa-dol':     ('Pennsylvania Department of Labor and Industry', 'minimum wage and overtime enforcement'),
    'pa-chr':     ('Pennsylvania Human Relations Commission', 'pay equity and employment discrimination'),
    'phila-oles': ('Philadelphia Office of Labor Relations', 'city wage and benefits standards'),
    # Ohio
    'oh-com':     ('Ohio Department of Commerce Division of Labor', 'wage and hour standards'),
    'oh-crc':     ('Ohio Civil Rights Commission', 'pay equity and employment discrimination'),
    # Colorado
    'co-dol':     ('Colorado Department of Labor and Employment', 'COMPS order and wage equity enforcement'),
    'co-crc':     ('Colorado Civil Rights Division', 'pay equity and equal pay for equal work'),
    # Michigan
    'mi-leo':     ('Michigan Department of Labor and Economic Opportunity', 'wage standards and enforcement'),
    'mi-doc':     ('Michigan Department of Civil Rights', 'pay equity and employment discrimination'),
    # New Jersey
    'nj-dol':     ('New Jersey Department of Labor and Workforce Development', 'pay equity and wage theft enforcement'),
    'nj-dcr':     ('New Jersey Division on Civil Rights', 'equal pay and employment discrimination'),
    # Minnesota
    'mn-dli':     ('Minnesota Department of Labor and Industry', 'prevailing wage and pay transparency standards'),
    'mn-mdhr':    ('Minnesota Department of Human Rights', 'pay equity and anti-discrimination enforcement'),
    # Oregon
    'or-boli':    ('Oregon Bureau of Labor and Industries', 'pay equity, wage enforcement, and worker rights'),
    # Virginia
    'va-dol':     ('Virginia Department of Labor and Industry', 'minimum wage and overtime enforcement'),
    'va-chr':     ('Virginia Council on Human Rights', 'pay equity and employment discrimination'),
    # Maryland
    'md-dol':     ('Maryland Department of Labor', 'wage payment and pay equity enforcement'),
    'md-chr':     ('Maryland Commission on Civil Rights', 'pay equity and anti-discrimination'),
    # Connecticut
    'ct-dol':     ('Connecticut Department of Labor', 'wage standards and pay equity enforcement'),
    'ct-chro':    ('Connecticut Commission on Human Rights and Opportunities', 'pay equity enforcement'),
    # Arizona
    'az-ica':     ('Arizona Industrial Commission', 'wage claims and workers compensation'),
    'az-ag':      ("Arizona Attorney General's Office", 'wage and employment law enforcement'),
    # North Carolina
    'nc-dol':     ('North Carolina Department of Labor', 'wage and hour standards enforcement'),
    # Georgia
    'ga-dol':     ('Georgia Department of Labor', 'unemployment and wage standards'),
    'ga-chrc':    ('Georgia Commission on Equal Opportunity', 'employment discrimination and pay equity'),
    # Wisconsin
    'wi-dwd-er':  ('Wisconsin DWD Equal Rights Division', 'state pay equity and discrimination enforcement'),
    'wi-dsps':    ('Wisconsin Department of Safety and Professional Services', 'professional licensing and compensation standards'),
    'wi-oci':     ('Wisconsin Office of the Commissioner of Insurance', 'insurance industry compensation governance'),
    'wi-dfi':     ('Wisconsin Department of Financial Institutions', 'state-chartered bank compensation standards'),
    'wi-dwd':     ('Wisconsin Department of Workforce Development', 'minimum wage, overtime, and equal pay'),
    # Nevada
    'nv-labor':   ('Nevada Office of the Labor Commissioner', 'wage standards and pay equity'),
    'nv-eeor':    ('Nevada Equal Rights Commission', 'pay equity and employment discrimination'),
    # District of Columbia
    'dc-does':    ('District of Columbia Department of Employment Services', 'wage theft and pay equity enforcement'),
    'dc-ohr':     ('DC Office of Human Rights', 'pay equity and anti-discrimination standards'),
    # Additional states
    'hi-dlir':    ('Hawaii Department of Labor and Industrial Relations', 'wage standards and pay equity'),
    'ri-dlt':     ('Rhode Island Department of Labor and Training', 'wage and hour standards'),
    'de-dol':     ('Delaware Department of Labor', 'wage payment and equal pay enforcement'),
    'vt-dol':     ('Vermont Department of Labor', 'wage standards and earned paid leave'),
    'nh-dol':     ('New Hampshire Department of Labor', 'wage and hour enforcement'),
    'me-dol':     ('Maine Department of Labor', 'minimum wage and overtime standards'),
    'wv-dol':     ('West Virginia Division of Labor', 'wage payment and compensation standards'),
    'ky-labor':   ('Kentucky Labor Cabinet', 'wage and hour standards enforcement'),
    'sc-llr':     ('South Carolina Department of Labor, Licensing, and Regulation', 'wage standards'),
    'la-lwc':     ('Louisiana Workforce Commission', 'wage standards and unemployment compensation'),
    'ok-dol':     ('Oklahoma Department of Labor', 'wage and hour standards enforcement'),
    'ia-dol':     ('Iowa Division of Labor', 'wage and hour and employment standards'),
    'ks-dol':     ('Kansas Department of Labor', 'wage payment and collection standards'),
    'ne-dol':     ('Nebraska Department of Labor', 'wage payment and collection enforcement'),
    'nm-dws':     ('New Mexico Department of Workforce Solutions', 'wage and employment standards'),
    'ak-dol':     ('Alaska Department of Labor and Workforce Development', 'prevailing wages and employment standards'),
    'mt-dol':     ('Montana Department of Labor and Industry', 'wage standards and human rights'),
    'id-labor':   ('Idaho Department of Labor', 'wage standards and workforce development'),
    'wy-dows':    ('Wyoming Department of Workforce Services', 'wage and hour standards'),
    'nd-labor':   ('North Dakota Department of Labor and Human Rights', 'wage and employment discrimination'),
    'sd-dol':     ('South Dakota Department of Labor and Regulation', 'wage and hour standards'),
    'ut-labor':   ('Utah Labor Commission', 'wage claims, anti-discrimination, and workers compensation'),
    'ar-dol':     ('Arkansas Department of Labor and Licensing', 'minimum wage and overtime standards'),
    'ms-decd':    ('Mississippi Department of Employment Security', 'unemployment and wage standards'),
    'al-dol':     ('Alabama Department of Labor', 'wage and hour enforcement'),
    'tn-dol':     ('Tennessee Department of Labor and Workforce Development', 'wage and employment standards'),
    'in-dol':     ('Indiana Department of Labor', 'minimum wage and workplace safety standards'),
    'mo-dol':     ('Missouri Department of Labor and Industrial Relations', 'minimum wage and wage theft'),
    # Self-regulatory organizations
    'nyse':       ('New York Stock Exchange', 'listed company compensation and governance standards'),
    'finra-oig':  ('FINRA Office of the Ombudsman', 'broker-dealer compensation dispute resolution'),
    'finra-enf':  ('FINRA Department of Enforcement', 'broker compensation rule enforcement'),
    'msrb-enf':   ('MSRB Enforcement Division', 'municipal advisor compensation rule enforcement'),
    'nfa-comp':   ('NFA Compliance Department', 'futures industry compensation rule compliance'),
    'nasdaq':     ('Nasdaq Stock Market', 'listed company executive pay disclosure and governance'),
    'cboe':       ('Chicago Board Options Exchange', 'equity compensation and derivatives governance'),
    'dtcc':       ('Depository Trust and Clearing Corporation', 'financial market infrastructure compensation'),
    'isda':       ('International Swaps and Derivatives Association', 'swap dealer compensation governance'),
    'sifma':      ('Securities Industry and Financial Markets Association', 'broker-dealer compensation regulatory standards'),
    'ici':        ('Investment Company Institute', 'fund manager compensation and governance'),
    'iia':        ('Institute of Internal Auditors', 'internal audit workforce compensation standards'),
    'aicpa':      ('American Institute of Certified Public Accountants', 'accounting profession compensation standards'),
    'bar-state':  ('National Conference of State Bar Associations', 'legal profession compensation and ethics'),
    'aba':        ('American Bar Association', 'attorney compensation and professional standards'),
    'fasb':       ('Financial Accounting Standards Board', 'stock compensation accounting standards (ASC 718)'),
    'gasb':       ('Governmental Accounting Standards Board', 'public sector pension and OPEB accounting'),
    'asb':        ('Actuarial Standards Board', 'actuarial standards for pension and benefit valuation'),
    'naic':       ('National Association of Insurance Commissioners', 'insurance executive compensation model laws'),
    'ama':        ('American Medical Association', 'physician compensation and healthcare workforce policy'),
    'shrm':       ('Society for Human Resource Management', 'HR compensation benchmarking and policy standards'),
    'wc-ncci':    ('National Council on Compensation Insurance', 'workers compensation rate standards'),
    # International
    'ilo':        ('International Labour Organization', 'global labor standards and wage policy'),
    'oecd':       ('Organisation for Economic Co-operation and Development', 'comparative executive pay and governance'),
    'imf':        ('International Monetary Fund', 'financial sector compensation and systemic risk'),
    'wto':        ('World Trade Organization', 'trade policy and international labor standards'),
    'worldbank':  ('World Bank', 'developing economy labor standards and compensation frameworks'),
    'bis':        ('Bank for International Settlements', 'global banking compensation and systemic risk'),
    'un-hrc':     ('UN Human Rights Council', 'business and human rights, including fair compensation'),
    'un-global':  ('UN Global Compact', 'corporate sustainability principles including labor standards'),
    'wef':        ('World Economic Forum', 'global competitiveness and human capital policy'),
    'gri':        ('Global Reporting Initiative', 'sustainability reporting standards for compensation'),
    'fst-board':  ('Financial Stability Board', 'global financial institution compensation standards'),
    'iosco':      ('International Organization of Securities Commissions', 'global securities compensation governance'),
    'iais':       ('International Association of Insurance Supervisors', 'insurance executive compensation standards'),
    'iops':       ('International Organisation of Pension Supervisors', 'pension fund governance and compensation'),
    # UK
    'uk-fca':     ('UK Financial Conduct Authority', 'financial services remuneration and conduct rules'),
    'uk-pra':     ('UK Prudential Regulation Authority', 'bank and insurer remuneration standards'),
    'uk-hmrc':    ('HM Revenue and Customs', 'executive compensation tax and benefits-in-kind treatment'),
    'uk-acas':    ('UK Advisory, Conciliation and Arbitration Service', 'employment dispute and pay equity standards'),
    'uk-frc':     ('UK Financial Reporting Council', 'corporate governance and stewardship code compensation'),
    'uk-pensions':('The Pensions Regulator (UK)', 'defined benefit pension funding and governance'),
    'uk-eq':      ('UK Equality and Human Rights Commission', 'gender pay gap reporting and enforcement'),
    'uk-tuc':     ('Trades Union Congress', 'collective bargaining and worker compensation standards'),
    # EU
    'eu-comm-empl':('European Commission DG Employment', 'EU labor standards and minimum wage directive'),
    'eu-parl-empl':('European Parliament EMPL Committee', 'EU social and compensation legislation'),
    'esma':        ('European Securities and Markets Authority', 'EU financial sector variable pay standards'),
    'eba':         ('European Banking Authority', 'EU bank remuneration governance and disclosure'),
    'ecb':         ('European Central Bank', 'Eurozone bank executive compensation and risk'),
    'eiopa':       ('European Insurance and Occupational Pensions Authority', 'insurer compensation governance'),
    # Canada
    'ca-esdc':    ('Employment and Social Development Canada', 'federal labor standards and employment insurance'),
    'ca-hrc':     ('Canadian Human Rights Commission', 'pay equity and employment discrimination'),
    'ca-osfi':    ('Office of the Superintendent of Financial Institutions Canada', 'financial institution compensation governance'),
    'ca-osc':     ('Ontario Securities Commission', 'executive compensation disclosure and governance'),
    # Australia
    'au-fwc':     ('Fair Work Commission', 'minimum wage and enterprise bargaining standards'),
    'au-asic':    ('Australian Securities and Investments Commission', 'executive remuneration governance'),
    'au-apra':    ('Australian Prudential Regulation Authority', 'financial institution remuneration standards'),
    # Germany
    'de-bafin':   ('German Federal Financial Supervisory Authority (BaFin)', 'German bank and insurer remuneration'),
    'de-bmas':    ('German Federal Ministry of Labour and Social Affairs (BMAS)', 'German labor and wage policy'),
    # France
    'fr-amf':     ('Autorité des marchés financiers (AMF)', 'French executive compensation disclosure'),
    'fr-acpr':    ('Autorité de contrôle prudentiel et de résolution (ACPR)', 'French bank and insurer remuneration'),
}

# ── Policy slug vocabulary ────────────────────────────────────────────────────

POLICY_SLUGS = [
    # Executive compensation — disclosure
    'executive-compensation-disclosure-requirements',
    'ceo-pay-ratio-disclosure-rule',
    'clawback-policy-final-rule',
    'equity-compensation-disclosure-standards',
    'deferred-compensation-reporting-requirements',
    'incentive-based-compensation-arrangements',
    'say-on-pay-frequency-requirements',
    'golden-parachute-disclosure-rule',
    'proxy-advisory-firm-oversight-standards',
    'compensation-committee-independence-requirements',
    'pay-versus-performance-disclosure',
    'hedging-policy-disclosure-requirements',
    'stock-repurchase-disclosure-rule',
    'executive-compensation-tax-deductibility-limits',
    'supplemental-executive-retirement-plan-disclosure',
    'change-in-control-severance-standards',
    'clawback-policy-enforcement-guidance',
    'equity-award-acceleration-disclosure',
    'compensation-recovery-rule-implementation',
    'ceo-pay-ratio-methodology-standards',
    'named-executive-officer-compensation-tables',
    'annual-incentive-plan-disclosure-standards',
    'long-term-incentive-plan-proxy-disclosure',
    'compensation-discussion-and-analysis-guidance',
    'compensation-committee-report-disclosure',
    'director-compensation-disclosure-standards',
    'beneficial-ownership-reporting-compensation',
    'insider-trading-disclosure-compensation',
    'related-party-transaction-compensation-disclosure',
    'form-8k-executive-compensation-triggers',
    'perquisite-disclosure-threshold-standards',
    'pension-value-increase-disclosure-rule',
    'nonqualified-deferred-compensation-disclosure',
    'all-other-compensation-column-standards',
    'ceo-succession-compensation-disclosure',
    'relative-tsr-peer-group-methodology',
    # Pay equity
    'pay-equity-reporting-standards',
    'pay-equity-data-submission-format',
    'pay-equity-statistical-methodology-guidance',
    'pay-equity-attorney-client-privilege-guidelines',
    'pay-equity-remediation-budgeting-standards',
    'pay-equity-and-performance-ratings',
    'pay-equity-in-hiring-and-promotion-decisions',
    'pay-transparency-requirements',
    'gender-pay-gap-disclosure-rule',
    'racial-pay-gap-reporting-standards',
    'equal-pay-certification-requirements',
    'compensation-history-ban-enforcement',
    'pay-equity-audit-standards',
    'salary-range-disclosure-requirements',
    'comparable-worth-standards',
    'pay-equity-remediation-guidelines',
    'pay-data-reporting-methodology',
    'intersectional-pay-equity-analysis',
    'pay-equity-safe-harbor-standards',
    'eo-1-pay-data-collection-expansion',
    'pay-equity-regression-analysis-standards',
    'pay-gap-reporting-industry-disaggregation',
    'controlled-pay-gap-analytical-methods',
    'pay-equity-data-collection-requirements',
    'pay-equity-enforcement-priorities',
    'pay-equity-litigation-standards',
    'pay-equity-voluntary-compliance-programs',
    'pay-equity-and-ai-hiring-tools',
    'pay-equity-in-federal-contracting',
    'total-compensation-equity-assessment',
    'pay-equity-certification-renewal-standards',
    'proactive-pay-equity-audit-requirements',
    'racial-wage-gap-closing-initiatives',
    'pay-transparency-salary-posting-enforcement',
    'internal-pay-equity-review-standards',
    'pay-equity-class-certification-standards',
    'employer-pay-gap-remediation-timelines',
    'pay-equity-and-remote-work-compensation',
    'pay-equity-in-mergers-and-acquisitions',
    'pay-equity-and-job-architecture-design',
    'pay-equity-and-variable-pay-programs',
    'pay-equity-and-starting-salary-negotiation',
    'pay-equity-and-retention-bonus-allocation',
    # Worker classification
    'independent-contractor-classification-rule',
    'gig-worker-classification-standards',
    'abc-test-enforcement-guidance',
    'worker-misclassification-penalties',
    'platform-worker-compensation-requirements',
    'freelancer-classification-standards',
    'joint-employer-liability-rule',
    'staffing-agency-worker-standards',
    'employee-vs-contractor-economic-reality-test',
    'app-based-worker-benefits-portability',
    'on-demand-worker-labor-protections',
    'franchise-worker-classification-guidance',
    'owner-operator-driver-classification',
    'day-labor-worker-employment-standards',
    'leased-employee-compensation-standards',
    'temporary-worker-compensation-parity',
    'subcontractor-employee-classification',
    'franchisee-employee-joint-employer-status',
    'intern-and-trainee-compensation-rules',
    'volunteer-and-stipend-worker-classification',
    'worker-classification-safe-harbor-provisions',
    'professional-employer-organization-standards',
    'worker-classification-federal-preemption',
    'multi-party-employment-compensation-liability',
    'digital-platform-worker-earned-leave',
    'classification-audit-trigger-standards',
    # Minimum wage & overtime
    'minimum-wage-regional-adjustment-factors',
    'overtime-calculation-for-bonuses-and-commissions',
    'white-collar-exemption-duties-test-modernization',
    'highly-compensated-employee-exemption-threshold',
    'outside-sales-exemption-guidance',
    'creative-professional-exemption-standards',
    'overtime-threshold-adjustment',
    'minimum-wage-floor-increase',
    'exempt-employee-salary-threshold',
    'tipped-worker-minimum-wage-standards',
    'federal-contractor-minimum-wage',
    'subminimum-wage-elimination',
    'overtime-exemption-criteria',
    'fluctuating-workweek-compensation',
    'automatic-overtime-threshold-indexing',
    'dual-minimum-wage-small-employer-standards',
    'state-preemption-minimum-wage-standards',
    'living-wage-federal-contractor-rule',
    'tipped-credit-elimination-proposal',
    'youth-subminimum-wage-phase-out',
    'agricultural-worker-overtime-exemption-reform',
    'domestic-worker-flsa-coverage-expansion',
    'on-call-and-waiting-time-compensation',
    'travel-time-compensability-standards',
    'meal-and-rest-break-compensation-guidance',
    'workweek-definition-and-overtime-calculation',
    'regular-rate-of-pay-calculation-guidance',
    'compensatory-time-off-public-sector-rules',
    'wage-deduction-and-garnishment-limits',
    'final-paycheck-timing-requirements',
    'wage-theft-prevention-and-enforcement',
    'prevailing-wage-determination-process',
    'salary-basis-test-modifications',
    'computer-professional-exemption-standards',
    'administrative-exemption-primary-duty-test',
    'executive-exemption-supervision-requirement',
    'learned-professional-exemption-guidance',
    # Benefits & retirement
    'erisa-fiduciary-duty-modernization',
    'prohibited-transaction-exemption-process',
    'defined-benefit-plan-mortality-table-updates',
    '401k-hardship-withdrawal-rules',
    'top-hat-plan-definition-and-requirements',
    'cafeteria-plan-nondiscrimination-testing',
    'retirement-plan-fiduciary-standards',
    'defined-benefit-pension-funding-rules',
    'erisa-preemption-scope',
    'health-insurance-employer-mandate',
    'benefits-continuation-standards',
    'student-loan-repayment-benefit-guidance',
    'paid-family-leave-standards',
    'paid-sick-leave-requirements',
    'emergency-paid-leave-guidance',
    'mental-health-parity-enforcement',
    'wellness-program-incentive-limits',
    'retirement-contribution-limit-adjustment',
    '401k-automatic-enrollment-standards',
    'pension-reform-funding-relief',
    'retiree-health-benefit-standards',
    'erisa-section-409a-deferred-compensation',
    'nonqualified-deferred-compensation-rules',
    'executive-benefit-plan-disclosure',
    'pension-plan-de-risking-strategies',
    '401k-fee-disclosure-requirements',
    'retirement-plan-investment-advice-rules',
    'esg-investing-in-retirement-plans',
    'missing-participant-guidance-retirement-plans',
    'cybersecurity-for-benefit-plans',
    'supplemental-disability-insurance-standards',
    'long-term-care-benefit-treatment',
    'esop-valuation-standards',
    'profit-sharing-plan-contribution-limits',
    'erisa-plan-asset-definition',
    'multiemployer-plan-funding-standards',
    'cash-balance-plan-conversion-standards',
    'pension-benefit-guaranty-premium-increase',
    'retiree-medical-benefit-accounting-standards',
    'flexible-spending-account-reform',
    'health-savings-account-limit-adjustment',
    'qualified-transportation-benefit-standards',
    'dependent-care-assistance-program-limits',
    'educational-assistance-benefit-guidance',
    'adoption-assistance-program-standards',
    'employee-assistance-program-coverage',
    'group-term-life-insurance-taxation',
    'disability-benefit-taxation-rules',
    'cafeteria-plan-election-change-rules',
    'cobra-subsidy-administration-guidance',
    'health-reimbursement-arrangement-integration-rules',
    'qualified-small-employer-hra-qsehra-standards',
    # Labor relations
    'noncompete-agreement-enforcement-guidelines',
    'no-poach-agreement-enforcement-standards',
    'confidentiality-agreement-scope-limits',
    'collective-bargaining-unit-determination',
    'union-election-procedures-rule',
    'card-check-certification-standards',
    'labor-management-reporting-requirements',
    'right-to-organize-protections',
    'strike-replacement-worker-standards',
    'mandatory-arbitration-employment-limits',
    'protected-concerted-activity-standards',
    'anti-retaliation-wage-complaint-protections',
    'secondary-boycott-enforcement-guidance',
    'workplace-surveillance-and-privacy-rights',
    'social-media-policy-nlra-implications',
    'employee-handbook-rule-review',
    'unfair-labor-practice-remedies',
    'neutrality-agreement-enforceability',
    'access-to-company-property-for-organizing',
    'successorship-doctrine-compensation-obligations',
    'effects-bargaining-compensation-duty',
    'unilateral-change-doctrine-compensation',
    'grievance-arbitration-compensation-disputes',
    'interest-arbitration-public-sector-standards',
    'labor-management-cooperation-programs',
    'union-security-agreement-enforcement',
    'agency-shop-fee-calculation-standards',
    'bargaining-in-good-faith-compensation-context',
    'wage-reopener-bargaining-standards',
    'pattern-bargaining-antitrust-exemption',
    'workplace-civility-policy-nlra-review',
    'joint-labor-management-committee-guidance',
    'employee-representation-on-corporate-boards',
    'sectoral-bargaining-framework-proposal',
    'electronic-voting-in-union-elections',
    'permanent-strike-replacement-prohibition',
    'employee-voice-mechanism-alternatives',
    # Financial services compensation
    'incentive-compensation-recovery-rule',
    'banker-bonus-deferral-requirements',
    'material-risk-taker-compensation-standards',
    'trading-desk-compensation-governance',
    'hedge-fund-manager-compensation-disclosure',
    'private-equity-carried-interest-standards',
    'insurance-agent-commission-disclosure',
    'mortgage-originator-compensation-rule',
    'investment-adviser-compensation-standards',
    'broker-dealer-compensation-governance',
    'variable-annuity-compensation-disclosure',
    'robo-adviser-compensation-conflicts',
    'registered-investment-adviser-fee-disclosure',
    'fiduciary-duty-compensation-conflicts',
    'revenue-sharing-arrangement-disclosure',
    'wrap-fee-program-compensation-standards',
    'financial-advisor-recruiting-bonus-disclosure',
    'soft-dollar-arrangements-and-compensation',
    'payment-for-order-flow-compensation-conflicts',
    '12b-1-fee-compensation-disclosure',
    'mutual-fund-sub-advisor-compensation',
    'credit-rating-agency-compensation-conflicts',
    'performance-based-fee-standards',
    'banking-incentive-compensation-guidance',
    'insurance-executive-bonus-governance',
    'private-fund-adviser-compensation-disclosure',
    'clawback-financial-institution-standards',
    'risk-adjusted-compensation-methodology',
    'compliance-function-compensation-independence',
    'front-office-compensation-risk-governance',
    'fintech-compensation-and-regulatory-arbitrage',
    'digital-asset-trader-compensation-guidance',
    'algorithmic-trading-incentive-compensation',
    'investment-banker-deal-toy-disclosure',
    'research-analyst-compensation-and-conflicts',
    'insurance-underwriter-profit-sharing-rules',
    # Federal contracting
    'prevailing-wage-determination-standards',
    'service-contract-act-wage-rates',
    'davis-bacon-act-modernization',
    'federal-contractor-pay-transparency',
    'contractor-affirmative-action-compensation',
    'contractor-minimum-wage-enforcement',
    'subcontractor-compensation-standards',
    'defense-contractor-executive-pay-limits',
    'government-contractor-pay-equity',
    'section-503-contractor-disability-compensation',
    'vevraa-contractor-veteran-compensation',
    'federal-contractor-paid-leave-rule',
    'contractor-employee-classification-guidance',
    'cost-reimbursement-contract-compensation',
    'time-and-materials-labor-hour-standards',
    'fixed-price-contractor-labor-compliance',
    'bid-wage-certification-requirements',
    'labor-category-qualification-standards',
    'overtime-recovery-contractor-standards',
    'contractor-whistleblower-compensation-protections',
    'service-contract-act-compliance-audits',
    'davis-bacon-act-compliance-and-enforcement',
    'contractor-joint-employer-liability',
    'project-labor-agreement-compensation-standards',
    'cost-accounting-standards-for-compensation',
    'allowable-executive-compensation-caps',
    'debarment-wage-violation-standards',
    'contractor-sick-leave-accrual-rule',
    # Corporate governance
    'board-diversity-disclosure-requirements',
    'shareholder-nomination-of-directors-proxy-access',
    'corporate-political-spending-disclosure-compensation',
    'staggered-board-declassification-proposals',
    'poison-pill-shareholder-rights-plan-review',
    'special-meeting-and-written-consent-rights',
    'board-compensation-oversight-standards',
    'institutional-investor-engagement-policy',
    'shareholder-say-on-pay-governance',
    'proxy-voting-compensation-guidelines',
    'corporate-governance-best-practices',
    'director-independence-compensation-standards',
    'related-party-transaction-compensation',
    'dual-class-share-compensation-governance',
    'esg-linked-compensation-disclosure',
    'diversity-linked-compensation-standards',
    'shareholder-proposal-compensation-threshold',
    'activist-investor-compensation-engagement',
    'stock-ownership-guidelines-governance',
    'anti-pledging-and-hedging-policies',
    'compensation-committee-charter-standards',
    'independent-compensation-consultant-standards',
    'compensation-peer-group-selection-standards',
    'say-on-golden-parachutes-governance',
    'universal-proxy-compensation-implications',
    'majority-voting-director-compensation-impact',
    'board-refreshment-compensation-practices',
    'director-compensation-disclosure-standards',
    'board-skills-matrix-disclosure',
    'lead-independent-director-role-and-pay',
    'executive-sessions-of-the-board-disclosure',
    'corporate-charter-and-bylaw-amendments-compensation',
    'shareholder-rights-plan-adoption-and-renewal',
    'virtual-shareholder-meeting-governance',
    'officer-compensation-recovery-policy',
    # Healthcare & life sciences
    'physician-compensation-stark-law-compliance',
    'hospital-executive-compensation-disclosure',
    'nonprofit-hospital-ceo-pay-ratio',
    'healthcare-worker-hazard-pay-standards',
    'pharmaceutical-executive-compensation-disclosure',
    'physician-self-referral-law-compensation',
    'anti-kickback-statute-compensation-arrangements',
    'gainsharing-arrangements-in-healthcare',
    'medical-loss-ratio-and-compensation',
    '340b-drug-pricing-program-compensation',
    'sunshine-act-compensation-reporting',
    'clinical-trial-investigator-compensation',
    'medicare-physician-payment-reform',
    'value-based-care-compensation-incentives',
    'telehealth-provider-compensation-standards',
    'direct-care-worker-wage-floor',
    'nursing-home-staffing-compensation-minimums',
    'home-health-aide-wage-floor',
    'physician-noncompete-reform',
    'medical-residency-compensation-standards',
    'nurse-staffing-ratio-compensation-impact',
    'hospital-merger-compensation-governance',
    'biotech-startup-equity-compensation-disclosure',
    'group-purchasing-organization-compensation',
    'pharmacy-benefit-manager-fee-disclosure',
    'durable-medical-equipment-supplier-compensation',
    'telemedicine-cross-state-licensing-and-pay',
    'accountable-care-organization-shared-savings',
    'clinical-trial-diversity-and-investigator-pay',
    'drug-pricing-reform-and-executive-incentives',
    'pharmacy-benefit-manager-compensation-reform',
    'behavioral-health-parity-workforce-compensation',
    # Technology & innovation
    'tech-worker-visa-prevailing-wage',
    'h1b-wage-level-requirements',
    'tech-industry-noncompete-reform',
    'algorithmic-wage-setting-standards',
    'remote-work-compensation-standards',
    'technology-transfer-compensation-rights',
    'artificial-intelligence-workforce-displacement',
    'gig-platform-worker-benefits-portability',
    'software-engineer-overtime-exemption-review',
    'ai-system-bias-in-compensation-decisions',
    'automated-hiring-tool-compensation-equity',
    'data-scientist-classification-standards',
    'h-1b-dependent-employer-wage-requirements',
    'l-1-visa-specialized-knowledge-compensation',
    'stem-opt-training-plan-compensation',
    'ai-and-job-displacement-compensation-policy',
    'algorithmic-management-and-pay-setting',
    'digital-nomad-tax-and-compensation-issues',
    'cyber-workforce-compensation-retention',
    'tech-startup-equity-compensation-guidance',
    'blockchain-worker-compensation-standards',
    'remote-employee-jurisdiction-compensation',
    'work-from-home-expense-reimbursement',
    'monitoring-software-and-compensation-equity',
    'productivity-tracking-compensation-implications',
    'platform-worker-data-portability-compensation',
    'open-source-developer-compensation-models',
    'patent-inventor-royalty-sharing-standards',
    'data-privacy-professional-compensation-benchmarks',
    'agile-development-team-compensation-structures',
    'tech-ethics-officer-compensation-and-independence',
    'quantum-computing-talent-retention-strategies',
    # Climate & ESG
    'climate-risk-compensation-incentives',
    'esg-executive-pay-linkage-disclosure',
    'carbon-reduction-compensation-standards',
    'sustainability-linked-compensation-disclosure',
    'scope-3-compensation-accountability',
    'just-transition-worker-compensation',
    'clean-energy-workforce-compensation-standards',
    'green-jobs-wage-floor-standards',
    'environmental-justice-workforce-compensation',
    'diversity-equity-inclusion-pay-disclosure',
    'human-capital-disclosure-compensation',
    'social-impact-compensation-metric-standards',
    'racial-equity-audit-compensation-governance',
    'corporate-political-spending-and-compensation',
    'esg-rating-agency-influence-on-compensation',
    'shareholder-proposals-on-esg-compensation',
    'board-diversity-and-compensation-outcomes',
    'supply-chain-labor-standards-and-compensation',
    'circular-economy-and-workforce-compensation',
    'pay-equity-esg-integration-standards',
    'biodiversity-and-land-use-compensation-metrics',
    'water-stewardship-incentive-compensation',
    'esg-rating-agency-methodology-transparency',
    'greenwashing-and-executive-accountability',
    'sustainable-finance-compensation-rules',
    'natural-capital-accounting-and-compensation',
    'workforce-diversity-compensation-reporting',
    'fair-trade-supply-chain-compensation',
    # International & cross-border
    'cross-border-compensation-reporting',
    'expatriate-compensation-tax-standards',
    'foreign-corrupt-practices-compensation',
    'international-executive-pay-comparability',
    'transfer-pricing-compensation-standards',
    'global-minimum-tax-compensation-impacts',
    'country-by-country-compensation-reporting',
    'eu-pay-transparency-directive-compliance',
    'oecd-pillar-two-compensation-implications',
    'trade-agreement-labor-chapter-enforcement',
    'supply-chain-due-diligence-compensation',
    'offshore-tax-haven-deferred-compensation',
    'foreign-private-issuer-compensation-disclosure',
    'global-mobility-and-remote-work-taxation',
    'works-council-consultation-on-compensation',
    'eu-corporate-sustainability-due-diligence-directive',
    'international-labor-organization-conventions-compensation',
    'oecd-guidelines-for-multinational-enterprises-compensation',
    'un-guiding-principles-on-business-and-human-rights-compensation',
    'bilateral-totalization-agreement-compensation',
    # Enforcement & penalties
    'wage-theft-penalties-enhancement',
    'civil-money-penalty-compensation-violations',
    'willful-wage-violation-criminal-penalties',
    'enforcement-coordination-compensation-fraud',
    'private-right-of-action-pay-equity',
    'class-action-compensation-litigation-standards',
    'arbitration-clause-compensation-disputes',
    'wage-and-hour-class-action-certification',
    'flsa-collective-action-decertification',
    'attorney-fee-awards-in-wage-litigation',
    'department-of-labor-amicus-brief-program',
    'strategic-enforcement-initiatives-compensation',
    'corporate-monitoring-for-wage-violations',
    'liquidated-damages-flsa-enhancement',
    'back-pay-interest-calculation-standards',
    'multi-agency-wage-enforcement-coordination',
    'state-federal-enforcement-preemption-framework',
    'qui-tam-compensation-violation-provisions',
    'retaliation-damages-compensation-complaints',
    'corporate-probation-and-monitoring-for-wage-violations',
    'debarment-for-repeated-wage-and-hour-violations',
    'personal-liability-for-executives-in-wage-cases',
    'statute-of-limitations-for-pay-discrimination-claims',
    'discovery-rules-in-compensation-litigation',
    'use-of-statistical-sampling-in-wage-audits',
    # Transparency & reporting
    'compensation-data-collection-standards',
    'eo-1-pay-data-reporting',
    'erisa-form-5500-compensation-disclosure',
    'proxy-statement-compensation-narrative',
    'total-compensation-disclosure-standards',
    'annual-incentive-plan-disclosure',
    'long-term-incentive-disclosure-requirements',
    'all-in-compensation-reporting-standard',
    'pay-ratio-peer-benchmarking-disclosure',
    'tally-sheet-disclosure-requirements',
    'workforce-analytics-reporting-standards',
    'human-capital-metrics-compensation-disclosure',
    'ceo-to-median-worker-pay-calculation',
    'voluntary-compensation-disclosure-safe-harbor',
    'integrated-reporting-and-compensation',
    'sustainability-reporting-compensation-metrics',
    'human-capital-accounting-standards',
    'workforce-cost-and-productivity-reporting',
    'talent-retention-and-turnover-disclosure',
    'diversity-and-inclusion-data-reporting',
    'executive-pay-ratio-trending-disclosure',
    'workforce-cost-capitalization-vs-expensing',
    'employee-wellbeing-and-mental-health-reporting',
    'workforce-safety-data-and-compensation-reporting',
    'training-and-development-investment-disclosure',
    'internal-mobility-and-promotion-rate-disclosure',
    'contingent-worker-and-contractor-spending-disclosure',
    # Specific rulemaking actions
    'proposed-rulemaking-comment',
    'advance-notice-proposed-rulemaking-response',
    'request-for-information-response',
    'proposed-rule-amendment-comment',
    'final-rule-implementation-guidance',
    'interim-final-rule-comment',
    'rulemaking-petition-executive-pay',
    'no-action-letter-request',
    'interpretive-guidance-request',
    'exemption-application-compensation',
    'variance-request-compensation-standard',
    'supplemental-comments-compensation-rule',
    'reply-comments-compensation-rulemaking',
    'ex-parte-submission-compensation-rule',
    'rulemaking-petition-pay-equity',
    'rulemaking-petition-overtime-threshold',
    'request-for-extension-comment-period',
    'request-for-public-hearing',
    'petition-for-reconsideration-of-final-rule',
    'request-for-technical-correction',
    'comment-on-information-collection-request',
    'response-to-regulatory-flexibility-analysis',
    'comment-on-paperwork-reduction-act-submission',
    'request-for-rulemaking-worker-classification',
    'request-for-guidance-deferred-compensation',
    'petition-to-repeal-existing-rule',
    'request-for-negotiated-rulemaking',
    'comment-on-direct-final-rule',
    'response-to-agency-data-quality-challenge',
    'comment-on-significant-guidance-document',
    'request-for-stay-of-effective-date',
    'request-for-no-action-relief',
    # General position/comment
    'position-statement-pay-transparency',
    'position-noncompete-reform',
    'position-overtime-threshold',
    'position-pay-equity-audit',
    'position-clawback-policy',
    'position-say-on-pay-effectiveness',
    'position-board-compensation-oversight',
    'position-gig-worker-classification',
    'position-minimum-wage-indexing',
    'position-erisa-preemption',
    'position-pay-ratio-disclosure',
    'position-compensation-tax-deductibility',
    'position-independent-contractor-test',
    'position-living-wage-standards',
    'position-executive-pay-cap',
    'position-mandatory-arbitration-reform',
    'position-worker-misclassification-penalties',
    'position-ai-compensation-equity',
    'position-remote-work-pay-standards',
    'position-esg-pay-linkage',
    'comment-executive-pay-ratio-rule',
    'position-shareholder-proposal-reform',
    'position-proxy-plumbing-and-compensation',
    'position-universal-proxy-and-compensation',
    'position-financial-transaction-tax-compensation',
    'position-carried-interest-tax-reform',
    'position-corporate-tax-rate-and-compensation',
    'position-on-board-diversity-mandates',
    'position-on-esg-disclosure-standardization',
    'position-on-portable-benefits-framework',
    'position-on-federal-paid-leave-mandate',
    'position-on-right-to-disconnect-legislation',
    'position-on-ai-in-hiring-and-promotion-regulation',
    'comment-overtime-salary-level',
    'comment-pay-equity-reporting',
    'comment-noncompete-final-rule',
    'comment-clawback-rule-implementation',
    'comment-worker-classification-final-rule',
    'comment-fiduciary-duty-compensation',
    'comment-pay-transparency-rule',
    'comment-federal-contractor-minimum-wage',
    'comment-mental-health-parity-enforcement',
    'comment-retirement-plan-fiduciary-standards',
    'comment-incentive-compensation-rule',
    # Testimony
    'testimony-executive-pay-reform',
    'testimony-pay-equity-legislation',
    'testimony-noncompete-reform-act',
    'testimony-worker-classification-reform',
    'testimony-minimum-wage-increase',
    'testimony-erisa-modernization',
    'testimony-overtime-rule-economic-impact',
    'testimony-ceo-pay-ratio-effectiveness',
    'testimony-compensation-transparency-legislation',
    'testimony-labor-market-competition',
    'testimony-pay-data-collection-improvement',
    'testimony-worker-misclassification-costs',
    'testimony-gig-economy-worker-protections',
    'testimony-mandatory-arbitration-reform',
    'testimony-stock-buyback-compensation',
    'testimony-401k-reform',
    'testimony-ai-workforce-displacement',
    'testimony-supply-chain-labor-standards',
    'testimony-portable-benefits-for-gig-workers',
    'testimony-corporate-governance-reform',
    'testimony-financial-regulation-and-compensation',
    'testimony-tax-policy-and-executive-pay',
    'testimony-antitrust-enforcement-in-labor-markets',
    'testimony-pay-transparency-state-laws',
    'testimony-federal-paid-leave',
    'testimony-future-of-work-and-automation',
    'testimony-retirement-security-crisis',
    'testimony-healthcare-affordability-and-benefits',
    'testimony-corporate-short-termism-and-compensation',
    'testimony-esg-investing-and-fiduciary-duty',
    'testimony-data-privacy-and-employee-monitoring',
    'testimony-pension-modernization',
    # Amicus briefs
    'amicus-brief-equal-pay-act',
    'amicus-brief-title-vii-compensation',
    'amicus-brief-flsa-exemption',
    'amicus-brief-erisa-plan-assets',
    'amicus-brief-arbitration-wage-claims',
    'amicus-brief-nlra-protected-activity',
    'amicus-brief-noncompete-enforceability',
    'amicus-brief-flsa-overtime-exemption',
    'amicus-brief-class-action-certification-standards',
    'amicus-brief-arbitrability-of-wage-claims',
    'amicus-brief-statistical-evidence-in-pay-cases',
    'amicus-brief-corporate-veil-piercing-for-wages',
    'amicus-brief-fiduciary-duty-in-esop-valuation',
    'amicus-brief-preemption-of-state-labor-laws',
    'amicus-brief-erisa-fiduciary-duty',
    'amicus-brief-pay-equity-statute',
    'amicus-brief-worker-classification',
    'amicus-brief-section-162m-tax',
    'amicus-brief-eeoc-pay-data',
    'amicus-brief-flsa-joint-employer',
    'amicus-brief-erisa-preemption',
    'amicus-brief-first-amendment-and-compelled-disclosure',
    'amicus-brief-takings-clause-and-compensation-regulation',
    'amicus-brief-commerce-clause-and-labor-law',
    'amicus-brief-administrative-procedure-act-rulemaking',
    'amicus-brief-chevron-deference-in-labor-cases',
    'amicus-brief-standing-in-data-privacy-compensation-cases',
    # Other
    'human-capital-management-disclosure-standards',
    'workforce-data-privacy-and-security-rules',
    'algorithmic-hiring-and-promotion-bias-audits',
    'internal-talent-marketplace-governance',
    'skills-based-hiring-and-compensation-frameworks',
    'employee-data-portability-and-ownership-rights',
    'ai-in-performance-management-disclosure-requirements',
    'workforce-productivity-monitoring-and-privacy',
    'digital-upskilling-and-reskilling-investment-incentives',
    'remote-work-and-geographic-pay-equity-standards',
    'contingent-workforce-and-benefits-parity-rules',
    'esg-metrics-in-executive-compensation-guidance',
    'board-level-human-capital-governance-disclosure',
    'supply-chain-labor-and-human-rights-due-diligence',
    'climate-risk-and-workforce-transition-planning',
    'just-transition-and-green-jobs-compensation-policy',
    'corporate-political-spending-and-lobbying-transparency',
    'shareholder-proposal-reform-and-employee-voice',
    'dual-class-share-structures-and-executive-accountability',
    'board-refreshment-and-tenure-limit-proposals',
    'cybersecurity-risk-oversight-and-board-compensation',
    'ai-ethics-committee-and-governance-reporting',
    'stakeholder-capitalism-and-corporate-purpose-disclosure',
    'workforce-mental-health-and-wellbeing-reporting-standards',
    'employee-financial-wellness-and-benefits-design',
    'student-loan-repayment-and-employer-contribution-tax-treatment',
    'portable-benefits-for-gig-economy-workers-framework',
    'right-to-disconnect-and-after-hours-work-policy',
    'four-day-work-week-and-overtime-implications',
    'workplace-safety-and-incentive-compensation-linkage',
    'predictive-analytics-in-hiring-and-adverse-impact-rules',
    'employee-surveillance-and-electronic-monitoring-disclosure',
    'workforce-data-monetization-and-employee-consent',
    'generative-ai-and-intellectual-property-ownership-in-employment',
    'virtual-reality-and-workplace-training-standards',
    'neurodiversity-in-the-workplace-accommodation-guidance',
    'caregiver-leave-and-flexible-work-arrangement-mandates',
    'long-term-care-and-employer-sponsored-benefit-plans',
    'fertility-and-family-forming-benefits-nondiscrimination',
    'gender-affirming-care-and-health-plan-coverage-mandates',
    'executive-perquisite-and-personal-use-of-company-assets-disclosure',
    'corporate-jet-usage-and-shareholder-value-analysis',

    'esg-rating-agency-methodology-transparency',
    'human-capital-accounting-standards',
    'workforce-data-privacy-and-security-rules',
    'algorithmic-hiring-and-promotion-bias-audits',
    'internal-talent-marketplace-governance',
    'skills-based-hiring-and-compensation-frameworks',
    'employee-data-portability-and-ownership-rights',
    'ai-in-performance-management-disclosure-requirements',
    'workforce-productivity-monitoring-and-privacy',
    'digital-upskilling-and-reskilling-investment-incentives',
    'remote-work-and-geographic-pay-equity-standards',
    'contingent-workforce-and-benefits-parity-rules',
    'board-level-human-capital-governance-disclosure',
    'supply-chain-labor-and-human-rights-due-diligence',
    'climate-risk-and-workforce-transition-planning',
    'just-transition-and-green-jobs-compensation-policy',
    'corporate-political-spending-and-lobbying-transparency',
    'shareholder-proposal-reform-and-employee-voice',
    'dual-class-share-structures-and-executive-accountability',
    'board-refreshment-and-tenure-limit-proposals',
    'cybersecurity-risk-oversight-and-board-compensation',
    'ai-ethics-committee-and-governance-reporting',
    'stakeholder-capitalism-and-corporate-purpose-disclosure',
    'workforce-mental-health-and-wellbeing-reporting-standards',
    'employee-financial-wellness-and-benefits-design',
    'student-loan-repayment-and-employer-contribution-tax-treatment',
    'portable-benefits-for-gig-economy-workers-framework',
    'right-to-disconnect-and-after-hours-work-policy',
    'four-day-work-week-and-overtime-implications',
    'workplace-safety-and-incentive-compensation-linkage',
    'predictive-analytics-in-hiring-and-adverse-impact-rules',
    'employee-surveillance-and-electronic-monitoring-disclosure',
    'workforce-data-monetization-and-employee-consent',
    'generative-ai-and-intellectual-property-ownership-in-employment',
    'virtual-reality-and-workplace-training-standards',
    'neurodiversity-in-the-workplace-accommodation-guidance',
    'caregiver-leave-and-flexible-work-arrangement-mandates',
    'long-term-care-and-employer-sponsored-benefit-plans',
    'fertility-and-family-forming-benefits-nondiscrimination',
    'gender-affirming-care-and-health-plan-coverage-mandates',
    'executive-perquisite-and-personal-use-of-company-assets-disclosure',
    'corporate-jet-usage-and-shareholder-value-analysis',
    'physician-compensation-and-value-based-care-models',
    'hospital-price-transparency-and-executive-compensation',
    'pharmaceutical-drug-pricing-and-rd-incentives',
    'medical-device-sales-compensation-safe-harbors',
    'telehealth-reimbursement-and-provider-compensation',
    'nursing-staffing-ratios-and-wage-pass-through-rules',
    'direct-care-worker-compensation-and-medicaid-reimbursement',
    'physician-noncompete-reform-and-market-competition',
    'gme-funding-and-medical-resident-compensation',
    'biotech-ipo-and-executive-equity-arrangements',
    'pbm-rebate-reform-and-compensation-implications',
    'clinical-trial-diversity-and-investigator-compensation',
    'accountable-care-organization-gainsharing-rules',
    'behavioral-health-workforce-compensation-initiatives',
    'tech-talent-immigration-and-prevailing-wage-standards',
    'ai-engineer-compensation-and-retention-strategies',
    'remote-work-taxation-and-state-nexus-issues',
    'open-source-software-and-developer-compensation-models',
    'data-privacy-officer-compensation-and-independence',
    'quantum-computing-talent-and-national-security-implications',
    'esop-and-employee-ownership-tax-incentives',
    'multiemployer-pension-plan-reform-and-withdrawal-liability',
    '401k-leakage-and-emergency-savings-proposals',
    'automatic-ira-and-state-mandated-retirement-plans',
    'pension-de-risking-and-lump-sum-window-guidance',
    'retirement-plan-fee-disclosure-and-litigation-trends',
    'esg-investing-in-erisa-plans-fiduciary-guidance',
    'cybersecurity-for-benefit-plans-and-fiduciary-liability',
    'long-term-care-insurance-and-public-private-partnerships',
    'health-savings-account-expansion-and-reform-proposals',
    'union-organizing-and-card-check-neutrality-agreements',
    'joint-employer-standard-and-franchise-model-liability',
    'nlra-and-employee-use-of-corporate-email-systems',
    'mandatory-arbitration-and-class-action-waiver-enforceability',
    'sectoral-bargaining-and-industry-wide-wage-standards',
    'workplace-surveillance-and-employee-privacy-rights',
    'employee-handbook-rules-and-protected-concerted-activity',
    'unfair-labor-practice-remedies-and-monetary-penalties',
    'right-to-work-laws-and-union-security-agreements',
    'electronic-voting-in-union-representation-elections',
    'banker-bonus-caps-and-systemic-risk-mitigation',
    'hedge-fund-carried-interest-tax-treatment',
    'private-equity-fee-and-expense-disclosure-rules',
    'investment-adviser-fiduciary-duty-and-compensation-conflicts',
    'broker-dealer-regulation-best-interest-and-commissions',
    'insurance-agent-compensation-and-consumer-protection',
    'mortgage-originator-compensation-and-steering-incentives',
    'fintech-and-regulatory-arbitrage-in-compensation',
    'digital-asset-compensation-and-securities-law-implications',
    'risk-adjusted-remuneration-in-financial-institutions',
    'prevailing-wage-requirements-for-infrastructure-projects',
    'service-contract-act-and-federal-service-contractor-wages',
    'davis-bacon-act-and-federal-construction-project-wages',
    'federal-contractor-minimum-wage-and-executive-order-14026',
    'ofccp-pay-equity-audits-and-affirmative-action-compliance',
    'defense-contractor-executive-compensation-caps-and-allowability',
    'cost-accounting-standards-for-pension-and-benefit-costs',
    'project-labor-agreements-and-federal-construction-projects',
    'universal-proxy-and-contested-director-elections',
    'shareholder-proposals-on-executive-pay-and-governance',
    'proxy-advisor-regulation-and-influence-on-voting',
    'board-diversity-and-nasdaq-listing-rule-requirements',
    'corporate-political-spending-and-shareholder-disclosure',
    'esg-metrics-in-incentive-plans-and-greenwashing-risk',
    'stock-buybacks-and-executive-compensation-linkage',
    'activist-investor-campaigns-and-board-compensation-responses',
    'say-on-climate-and-executive-accountability-for-emissions',
    'human-capital-management-disclosure-and-sec-rulemaking',
    'wage-theft-and-criminal-liability-for-executives',
    'class-action-certification-in-wage-and-hour-litigation',
    'private-attorneys-general-act-paga-and-labor-code-enforcement',
    'flsa-liquidated-damages-and-good-faith-defense',
    'department-of-labor-strategic-enforcement-priorities',
    'amicus-brief-on-statistical-evidence-in-pay-discrimination-cases',
    'amicus-brief-on-arbitrability-of-erisa-fiduciary-claims',
    'amicus-brief-on-preemption-of-state-labor-laws',
    'amicus-brief-on-joint-employer-status-under-the-nlra',
    'amicus-brief-on-the-major-questions-doctrine-and-labor-regulation',
    'c-suite-perquisite-disclosure-reform',
    'executive-severance-tax-implications',
    'non-employee-director-stock-award-valuation',
    'insider-trading-policy-10b5-1-plan-reform',
    'compensation-consultant-independence-disclosure',
    'peer-group-selection-and-disclosure-best-practices',
    'say-on-pay-vote-outcome-and-board-responsiveness',
    'equity-granting-practices-and-timing-disclosure',
    'executive-compensation-in-bankruptcy-proceedings',
    'performance-metric-selection-and-rigor-disclosure',
    'pay-equity-and-ai-driven-compensation-tools',
    'intersectional-pay-gap-analysis-methodology',
    'pay-transparency-and-remote-work-geographic-pay',
    'opportunity-equity-and-promotion-velocity-metrics',
    'pay-equity-in-merit-and-bonus-allocation',
    'global-pay-equity-reporting-standards',
    'pay-equity-audit-privilege-and-waiver-risks',
    'salary-history-ban-impact-on-starting-pay',
    'pay-compression-analysis-and-remediation-strategies',
    'pay-equity-and-contingent-workforce-compensation',
    'portable-benefits-for-gig-workers-federal-framework',
    'algorithmic-management-and-worker-classification',
    'franchise-model-and-joint-employer-liability-reform',
    'digital-platform-worker-minimum-earnings-floor',
    'misclassification-and-state-unemployment-insurance',
    'sectoral-bargaining-for-app-based-workers',
    'right-to-disconnect-for-platform-workers',
    'data-portability-for-freelancers-and-gig-workers',
    'auto-portability-for-401k-plans-final-rule',
    'mental-health-parity-and-telehealth-coverage',
    'long-term-care-benefits-and-public-private-partnerships',
    'student-loan-matching-in-401k-plans-guidance',
    'emergency-savings-accounts-in-retirement-plans',
    'erisa-preemption-and-state-mandated-benefits',
    'cybersecurity-fiduciary-duty-for-benefit-plans',
    'esg-investing-and-erisa-fiduciary-duty-update',
    'pension-risk-transfer-and-annuity-provider-selection',
    'health-plan-price-transparency-and-fiduciary-duty',
    'ai-in-hiring-and-eeoc-compliance-guidance',
    'workplace-surveillance-and-employee-privacy-act',
    'generative-ai-and-intellectual-property-in-employment',
    'algorithmic-hiring-bias-audit-requirements',
    'digital-upskilling-and-workforce-investment-tax-credits',
    'human-capital-metrics-and-investor-disclosure',
    'board-level-cybersecurity-expertise-disclosure',
    'dual-class-share-structure-sunset-provisions',
    'shareholder-proposal-resubmission-thresholds',
    'universal-proxy-and-board-accountability',
    'climate-competency-on-corporate-boards',
    'corporate-political-spending-disclosure-act',
    'esg-rating-agency-regulation-and-transparency',
    'natural-capital-accounting-and-disclosure-standards',
    'just-transition-and-fossil-fuel-workforce-support',
    'supply-chain-human-rights-due-diligence-act',
    'circular-economy-and-extended-producer-responsibility',
    'sustainable-finance-disclosure-regulation-us-alignment',
    'drug-price-negotiation-and-pharma-executive-pay',
    'hospital-charity-care-and-executive-compensation',
    'physician-fee-schedule-and-value-based-care-incentives',
    'medical-loss-ratio-and-insurance-executive-bonuses',
    'biotech-rd-investment-and-executive-stock-sales',
    'telehealth-fraud-and-provider-compensation-schemes',
    'nursing-home-staffing-mandates-and-wage-pass-throughs',
    '340b-drug-pricing-and-hospital-executive-incentives',
    'comment-on-ai-and-algorithmic-fairness-in-hiring',
    'position-statement-on-human-capital-disclosure-mandates',
    'testimony-on-the-future-of-work-and-automation',
    'amicus-brief-on-ai-bias-in-employment-decisions',
    'rulemaking-petition-for-esg-disclosure-standardization',
    'comment-on-cybersecurity-governance-and-disclosure',
    'position-on-shareholder-rights-and-corporate-accountability',
    'testimony-on-climate-risk-and-financial-regulation',
    'amicus-brief-on-erisa-preemption-of-state-benefit-mandates',
    'rulemaking-petition-for-portable-benefits-framework',
    'comment-on-joint-employer-status-final-rule',
    'position-statement-on-sectoral-bargaining-proposals',
    'testimony-on-noncompete-reform-and-labor-mobility',
    'amicus-brief-on-the-major-questions-doctrine-in-labor-law',
    'rulemaking-petition-on-algorithmic-management-transparency',
    'comment-on-prevailing-wage-modernization-for-green-jobs',
    'position-on-carried-interest-tax-loophole-closure',
    'testimony-on-financial-transaction-taxes-and-compensation',
    'amicus-brief-on-fiduciary-duty-in-esg-investing',
    'rulemaking-petition-for-banker-bonus-deferral-rules',
    'comment-on-digital-asset-compensation-and-taxation',
    'position-on-risk-adjusted-remuneration-for-systemically-important-financial-institutions',
    'testimony-on-insurance-agent-commission-disclosure-and-conflicts',
    'amicus-brief-on-the-accreditation-of-esg-rating-agencies',
    'rulemaking-petition-for-corporate-water-stewardship-disclosure',
    'comment-on-biodiversity-risk-and-financial-stability',
    'position-on-scope-3-emissions-and-executive-accountability',
    'testimony-on-the-just-transition-for-fossil-fuel-workers',
    'amicus-brief-on-greenwashing-and-securities-fraud',
    'rulemaking-petition-for-supply-chain-labor-due-diligence',
    'comment-on-the-eu-us-data-privacy-framework-and-hr-data',
    'position-on-global-minimum-tax-and-executive-compensation',
    'testimony-on-trade-agreements-and-international-labor-standards',
    'amicus-brief-on-extraterritorial-application-of-us-employment-law',
    'rulemaking-petition-for-works-council-consultation-rights',

    # Executive Compensation & Governance (Advanced Topics)
    'c-suite-perquisite-disclosure-reform',
    'executive-severance-tax-implications',
    'non-employee-director-stock-award-valuation',
    'insider-trading-policy-10b5-1-plan-reform',
    'compensation-consultant-independence-disclosure',
    'peer-group-selection-and-disclosure-best-practices',
    'say-on-pay-vote-outcome-and-board-responsiveness',
    'equity-granting-practices-and-timing-disclosure',
    'executive-compensation-in-bankruptcy-proceedings',
    'performance-metric-selection-and-rigor-disclosure',
    'compensation-committee-risk-oversight-duties',
    'ceo-succession-planning-and-emergency-compensation',
    'shareholder-engagement-on-compensation-best-practices',
    'activist-investor-demands-on-executive-pay',
    'esg-metrics-in-executive-incentives-framework',

    # Pay Equity & Transparency (Emerging Issues)
    'pay-equity-and-ai-driven-compensation-tools',
    'intersectional-pay-gap-analysis-methodology',
    'pay-transparency-and-remote-work-geographic-pay',
    'opportunity-equity-and-promotion-velocity-metrics',
    'pay-equity-in-merit-and-bonus-allocation',
    'global-pay-equity-reporting-standards',
    'pay-equity-audit-privilege-and-waiver-risks',
    'salary-history-ban-impact-on-starting-pay',
    'pay-compression-analysis-and-remediation-strategies',
    'pay-equity-and-contingent-workforce-compensation',
    'pay-data-reporting-to-state-agencies',
    'pay-equity-in-mergers-and-acquisitions-due-diligence',
    'algorithmic-hiring-and-pay-discrimination-risk',
    'pay-equity-and-job-architecture-leveling',
    'pay-equity-in-sales-commission-plans',

    # Future of Work & Worker Classification
    'portable-benefits-for-gig-workers-federal-framework',
    'algorithmic-management-and-worker-classification',
    'franchise-model-and-joint-employer-liability-reform',
    'digital-platform-worker-minimum-earnings-floor',
    'misclassification-and-state-unemployment-insurance',
    'sectoral-bargaining-for-app-based-workers',
    'right-to-disconnect-for-remote-workers',
    'data-portability-for-freelancers-and-gig-workers',
    'ai-and-the-redefinition-of-professional-exemptions',
    'workforce-data-privacy-and-employee-monitoring-rules',
    'four-day-work-week-and-overtime-implications',
    'skills-based-hiring-and-credentialing-standards',
    'internal-talent-marketplace-and-pay-equity',
    'workforce-reskilling-and-displacement-support-policy',
    'digital-nomad-tax-and-employment-law-framework',

    # Retirement & Benefits (Modernization)
    'auto-portability-for-401k-plans-final-rule',
    'mental-health-parity-and-telehealth-coverage-mandates',
    'long-term-care-benefits-and-public-private-partnerships',
    'student-loan-matching-in-401k-plans-guidance',
    'emergency-savings-accounts-in-retirement-plans',
    'erisa-preemption-and-state-mandated-benefits-programs',
    'cybersecurity-fiduciary-duty-for-benefit-plans',
    'esg-investing-in-erisa-plans-fiduciary-guidance-update',
    'pension-risk-transfer-and-annuity-provider-selection-rules',
    'health-plan-price-transparency-and-fiduciary-duty',
    'pharmacogenomics-and-group-health-plan-coverage',
    'fertility-and-family-building-benefits-nondiscrimination',
    'paid-family-and-medical-leave-federal-framework',
    'retirement-plan-fee-disclosure-and-litigation-trends',
    'lifetime-income-disclosure-for-defined-contribution-plans',

    # Human Capital & ESG Disclosure
    'human-capital-metrics-and-investor-disclosure-rules',
    'board-level-cybersecurity-expertise-disclosure',
    'dual-class-share-structure-sunset-provisions',
    'shareholder-proposal-resubmission-thresholds-reform',
    'universal-proxy-and-board-accountability-impact',
    'climate-competency-on-corporate-boards-disclosure',
    'corporate-political-spending-disclosure-act-comment',
    'esg-rating-agency-regulation-and-transparency',
    'natural-capital-accounting-and-corporate-disclosure',
    'just-transition-and-fossil-fuel-workforce-support-policy',
    'supply-chain-human-rights-due-diligence-act-comment',
    'circular-economy-and-extended-producer-responsibility-policy',
    'sustainable-finance-disclosure-regulation-us-alignment',
    'workforce-mental-health-and-wellbeing-disclosure-standards',
    'racial-equity-audit-disclosure-and-governance',

    # Healthcare & Life Sciences (Specialized Topics)
    'drug-price-negotiation-and-pharma-executive-pay-linkage',
    'hospital-charity-care-obligations-and-executive-compensation',
    'physician-fee-schedule-and-value-based-care-incentives',
    'medical-loss-ratio-and-insurance-executive-bonuses',
    'biotech-rd-investment-and-executive-stock-sales-policy',
    'telehealth-fraud-and-provider-compensation-schemes',
    'nursing-home-staffing-mandates-and-wage-pass-throughs',
    '340b-drug-pricing-and-hospital-executive-incentives',
    'physician-owned-distributorships-and-anti-kickback-rules',
    'medical-device-sales-rep-compensation-and-sunshine-act',
    'value-based-purchasing-and-hospital-incentive-pools',
    'gme-funding-reform-and-medical-resident-compensation',
    'clinical-trial-data-sharing-and-researcher-incentives',
    'pbm-rebate-reform-and-impact-on-plan-sponsor-costs',
    'behavioral-health-workforce-shortage-and-compensation-policy',

    # Financial Services (Specialized Topics)
    'banker-bonus-caps-and-systemic-risk-mitigation',
    'hedge-fund-carried-interest-tax-treatment-reform',
    'private-equity-fee-and-expense-disclosure-rules',
    'investment-adviser-fiduciary-duty-and-compensation-conflicts',
    'broker-dealer-regulation-best-interest-and-commissions',
    'insurance-agent-compensation-and-consumer-protection-standards',
    'mortgage-originator-compensation-and-steering-incentives',
    'fintech-and-regulatory-arbitrage-in-compensation',
    'digital-asset-compensation-and-securities-law-implications',
    'risk-adjusted-remuneration-in-financial-institutions',
    'payment-for-order-flow-and-best-execution-conflicts',
    '12b-1-fee-reform-and-mutual-fund-distribution-costs',
    'credit-rating-agency-analyst-compensation-and-independence',
    'high-frequency-trading-and-incentive-structures',
    'robo-adviser-compensation-and-algorithmic-bias',

    # Labor & Antitrust
    'no-poach-and-wage-fixing-agreement-enforcement',
    'labor-market-concentration-and-monopsony-power',
    'antitrust-and-collective-bargaining-for-independent-contractors',
    'information-sharing-and-wage-surveys-antitrust-guidelines',
    'interlocking-directorates-and-labor-market-effects',
    'non-compete-clause-reform-and-labor-mobility',
    'franchise-no-poach-clause-enforcement',
    'merger-review-and-impact-on-labor-markets',
    'price-gouging-and-hazard-pay-during-emergencies',
    'class-action-waivers-in-employment-arbitration-agreements',
    'paga-reform-and-private-attorney-general-actions-in-california',
    'joint-employer-liability-and-supply-chain-responsibility',
    'nlra-and-the-use-of-workplace-surveillance-technology',
    'employee-data-and-privacy-rights-in-the-workplace',
    'captive-audience-meetings-and-nlra-protections',

    # Executive Compensation & Governance (Advanced Topics)
    'c-suite-perquisite-disclosure-reform',
    'executive-severance-tax-implications',
    'non-employee-director-stock-award-valuation',
    'insider-trading-policy-10b5-1-plan-reform',
    'compensation-consultant-independence-disclosure',
    'peer-group-selection-and-disclosure-best-practices',
    'say-on-pay-vote-outcome-and-board-responsiveness',
    'equity-granting-practices-and-timing-disclosure',
    'executive-compensation-in-bankruptcy-proceedings',
    'performance-metric-selection-and-rigor-disclosure',
    'compensation-committee-risk-oversight-duties',
    'ceo-succession-planning-and-emergency-compensation',
    'shareholder-engagement-on-compensation-best-practices',
    'activist-investor-demands-on-executive-pay',
    'esg-metrics-in-executive-incentives-framework',

    # Pay Equity & Transparency (Emerging Issues)
    'pay-equity-and-ai-driven-compensation-tools',
    'intersectional-pay-gap-analysis-methodology',
    'pay-transparency-and-remote-work-geographic-pay',
    'opportunity-equity-and-promotion-velocity-metrics',
    'pay-equity-in-merit-and-bonus-allocation',
    'global-pay-equity-reporting-standards',
    'pay-equity-audit-privilege-and-waiver-risks',
    'salary-history-ban-impact-on-starting-pay',
    'pay-compression-analysis-and-remediation-strategies',
    'pay-equity-and-contingent-workforce-compensation',
    'pay-data-reporting-to-state-agencies',
    'pay-equity-in-mergers-and-acquisitions-due-diligence',
    'algorithmic-hiring-and-pay-discrimination-risk',
    'pay-equity-and-job-architecture-leveling',
    'pay-equity-in-sales-commission-plans',

    # Future of Work & Worker Classification
    'portable-benefits-for-gig-workers-federal-framework',
    'algorithmic-management-and-worker-classification',
    'franchise-model-and-joint-employer-liability-reform',
    'digital-platform-worker-minimum-earnings-floor',
    'misclassification-and-state-unemployment-insurance',
    'sectoral-bargaining-for-app-based-workers',
    'right-to-disconnect-for-remote-workers',
    'data-portability-for-freelancers-and-gig-workers',
    'ai-and-the-redefinition-of-professional-exemptions',
    'workforce-data-privacy-and-employee-monitoring-rules',
    'four-day-work-week-and-overtime-implications',
    'skills-based-hiring-and-credentialing-standards',
    'internal-talent-marketplace-and-pay-equity',
    'workforce-reskilling-and-displacement-support-policy',
    'digital-nomad-tax-and-employment-law-framework',

    # Retirement & Benefits (Modernization)
    'auto-portability-for-401k-plans-final-rule',
    'mental-health-parity-and-telehealth-coverage-mandates',
    'long-term-care-benefits-and-public-private-partnerships',
    'student-loan-matching-in-401k-plans-guidance',
    'emergency-savings-accounts-in-retirement-plans',
    'erisa-preemption-and-state-mandated-benefits-programs',
    'cybersecurity-fiduciary-duty-for-benefit-plans',
    'esg-investing-in-erisa-plans-fiduciary-guidance-update',
    'pension-risk-transfer-and-annuity-provider-selection-rules',
    'health-plan-price-transparency-and-fiduciary-duty',
    'pharmacogenomics-and-group-health-plan-coverage',
    'fertility-and-family-building-benefits-nondiscrimination',
    'paid-family-and-medical-leave-federal-framework',
    'retirement-plan-fee-disclosure-and-litigation-trends',
    'lifetime-income-disclosure-for-defined-contribution-plans',

    # Human Capital & ESG Disclosure
    'human-capital-metrics-and-investor-disclosure-rules',
    'board-level-cybersecurity-expertise-disclosure',
    'dual-class-share-structure-sunset-provisions',
    'shareholder-proposal-resubmission-thresholds-reform',
    'universal-proxy-and-board-accountability-impact',
    'climate-competency-on-corporate-boards-disclosure',
    'corporate-political-spending-disclosure-act-comment',
    'esg-rating-agency-regulation-and-transparency',
    'natural-capital-accounting-and-corporate-disclosure',
    'just-transition-and-fossil-fuel-workforce-support-policy',
    'supply-chain-human-rights-due-diligence-act-comment',
    'circular-economy-and-extended-producer-responsibility-policy',
    'sustainable-finance-disclosure-regulation-us-alignment',
    'workforce-mental-health-and-wellbeing-disclosure-standards',
    'racial-equity-audit-disclosure-and-governance',

    # Healthcare & Life Sciences (Specialized Topics)
    'drug-price-negotiation-and-pharma-executive-pay-linkage',
    'hospital-charity-care-obligations-and-executive-compensation',
    'physician-fee-schedule-and-value-based-care-incentives',
    'medical-loss-ratio-and-insurance-executive-bonuses',
    'biotech-rd-investment-and-executive-stock-sales-policy',
    'telehealth-fraud-and-provider-compensation-schemes',
    'nursing-home-staffing-mandates-and-wage-pass-throughs',
    '340b-drug-pricing-and-hospital-executive-incentives',
    'physician-owned-distributorships-and-anti-kickback-rules',
    'medical-device-sales-rep-compensation-and-sunshine-act',
    'value-based-purchasing-and-hospital-incentive-pools',
    'gme-funding-reform-and-medical-resident-compensation',
    'clinical-trial-data-sharing-and-researcher-incentives',
    'pbm-rebate-reform-and-impact-on-plan-sponsor-costs',
    'behavioral-health-workforce-shortage-and-compensation-policy',

    # Financial Services (Specialized Topics)
    'banker-bonus-caps-and-systemic-risk-mitigation',
    'hedge-fund-carried-interest-tax-treatment-reform',
    'private-equity-fee-and-expense-disclosure-rules',
    'investment-adviser-fiduciary-duty-and-compensation-conflicts',
    'broker-dealer-regulation-best-interest-and-commissions',
    'insurance-agent-compensation-and-consumer-protection-standards',
    'mortgage-originator-compensation-and-steering-incentives',
    'fintech-and-regulatory-arbitrage-in-compensation',
    'digital-asset-compensation-and-securities-law-implications',
    'risk-adjusted-remuneration-in-financial-institutions',
    'payment-for-order-flow-and-best-execution-conflicts',
    '12b-1-fee-reform-and-mutual-fund-distribution-costs',
    'credit-rating-agency-analyst-compensation-and-independence',
    'high-frequency-trading-and-incentive-structures',
    'robo-adviser-compensation-and-algorithmic-bias',

    # Labor & Antitrust
    'no-poach-and-wage-fixing-agreement-enforcement',
    'labor-market-concentration-and-monopsony-power',
    'antitrust-and-collective-bargaining-for-independent-contractors',
    'information-sharing-and-wage-surveys-antitrust-guidelines',
    'interlocking-directorates-and-labor-market-effects',
    'non-compete-clause-reform-and-labor-mobility',
    'franchise-no-poach-clause-enforcement',
    'merger-review-and-impact-on-labor-markets',
    'price-gouging-and-hazard-pay-during-emergencies',
    'class-action-waivers-in-employment-arbitration-agreements',
    'paga-reform-and-private-attorney-general-actions-in-california',
    'joint-employer-liability-and-supply-chain-responsibility',
    'nlra-and-the-use-of-workplace-surveillance-technology',
    'employee-data-and-privacy-rights-in-the-workplace',
    'captive-audience-meetings-and-nlra-protections',
]

# ── Document type definitions ─────────────────────────────────────────────────

DOCUMENT_TYPES = [
    ('comment-letter',          'Comment Letter'),
    ('position-statement',      'Position Statement'),
    ('policy-brief',            'Policy Brief'),
    ('legislative-testimony',   'Legislative Testimony'),
    ('amicus-brief',            'Amicus Curiae Brief'),
    ('white-paper',             'White Paper'),
    ('supplemental-comments',   'Supplemental Comments'),
    ('reply-comments',          'Reply Comments'),
    ('ex-parte-submission',     'Ex Parte Submission'),
    ('regulatory-petition',     'Petition for Rulemaking'),
    ('no-action-request',       'No-Action Request'),
    ('advisory-memorandum',     'Advisory Memorandum'),
    ('guidance-document',       'Interpretive Guidance'),
    ('enforcement-policy',      'Enforcement Policy Statement'),
    ('compliance-bulletin',     'Compliance Bulletin'),
    ('legal-analysis',          'Legal Analysis Memorandum'),
    ('economic-analysis',       'Economic Impact Analysis'),
    ('research-report',         'Research Report'),
    ('joint-comments',          'Joint Comments'),
    ('research-memorandum',     'Research Memorandum'),
    ('formal-objection',        'Formal Objection'),
    ('request-for-information-response', 'Response to Request for Information'),
    ('advance-notice-comment',  'Comment on Advance Notice of Proposed Rulemaking'),
    ('interim-final-rule-comment', 'Comment on Interim Final Rule'),
    ('petition-for-reconsideration', 'Petition for Reconsideration of Final Rule'),
    ('request-for-stay',        'Request for Stay of Effective Date'),
    ('request-for-exemption',   'Request for Exemption'),
    ('cost-benefit-analysis',   'Cost-Benefit Analysis'),
    ('implementation-guide',    'Implementation Guide'),
    ('best-practices-guide',    'Best Practices Guide'),
    ('coalition-letter',        'Coalition Letter'),
    # More specific regulatory/legal actions
    ('expert-declaration',      'Expert Declaration'),
    ('request-for-technical-correction', 'Request for Technical Correction'),
    ('comment-on-information-collection', 'Comment on Information Collection Request'),
    ('petition-to-repeal-rule', 'Petition to Repeal Existing Rule'),
    ('response-to-agency-data-quality-challenge', 'Response to Data Quality Challenge'),
    # Data & methodology focused
    ('data-submission',         'Data Submission'),
    ('methodology-white-paper', 'Methodology White Paper'),
    ('statistical-analysis-report', 'Statistical Analysis Report'),
    ('fact-sheet',              'Fact Sheet'),
    ('roundtable-summary',      'Public Roundtable Summary'),
]

# ── Signatory generation ──────────────────────────────────────────────────────

SIGNATORY_TITLES = [
    'Senior Policy Director',
    'Vice President, Government Relations',
    'Director of Regulatory Affairs',
    'Principal, Compensation Policy',
    'Senior Fellow, Labor Markets',
    'Director of Policy Research',
    'Vice President, Policy and Advocacy',
    'Chief Policy Officer',
    'Senior Policy Counsel',
    'Director, Public Affairs',
    'Director of Research and Policy',
    'General Counsel',
    'Chief Legal Officer',
    'Deputy General Counsel, Regulatory Affairs',
    'Associate General Counsel, Labor & Employment',
    'Senior Corporate Counsel, Policy',
    'Regulatory Counsel',
    'Principal, Economic Research',
    'Senior Director, Regulatory Strategy',
    'Vice President, External Affairs',
    'Managing Director, Policy',
    'Associate Director, Regulatory Affairs',
    'Fellow, Executive Compensation Research',
    'Senior Research Director',
    'Chief Economist',
    'Senior Vice President, Policy',
    'Director of Legislative Affairs',
    'Principal, Labor and Employment Policy',
    'Senior Economist, Labor Markets',
    'Principal Data Scientist, Policy Analytics',
    'Quantitative Analyst, Regulatory Economics',
    'Senior Research Scientist, Workforce Policy',
    'Director of Economic Analysis',
    'Head of Quantitative Research',
    'Research Director, Compensation Policy',
    'Senior Advisor, Government Affairs',
    'Vice President, Research',
    'Director, Workforce Policy',
    'Senior Principal, Regulatory Consulting',
    'Associate General Counsel, Policy',
    'Director, Policy Innovation',
    'Head of Policy Research',
    'Senior Policy Fellow',
    'Principal, Benefits and Retirement Policy',
    'Director of Federal Affairs',
    'Senior Director, Research',
    'Managing Principal, Government Affairs',
    'Vice President, Strategy and Policy',
    'Principal Economist',
    'Director of Compensation Research',
    'Senior Advisor, Regulatory Policy',
    'Vice President, Compensation Governance',
    'Director, Corporate Governance Policy',
    'Senior Fellow, Regulatory Affairs',
    'Principal, Pay Equity Research',
    'Director, Employer Policy',
    'Senior Director, Global Policy',
    'Director, State and Local Government Affairs',
    'Manager, Federal Government Relations',
    'Policy Analyst',
    'Legislative Analyst',
    'Regulatory Analyst',
    'Research Associate, Policy',
    'Legal Fellow, Regulatory Policy',
    'Economic Policy Fellow',
    'Government Affairs Specialist',
    'Public Policy Manager',
    'Senior Vice President, Research',
    'Managing Director, Government Affairs',
    'Public Policy Manager',
    'Senior Vice President, Research',
    'Managing Director, Government Affairs',
    # More senior/specialized roles
    'Chief Research Officer',
    'Head of Global Policy',
    'Distinguished Fellow, Corporate Governance',
    'Research Director, Labor Economics',
    'Lead Quantitative Researcher, Pay Equity',
    'Principal, Human Capital Research',
    'Senior Regulatory Counsel',
    'Chief Counsel for Policy',
    'Head of Legislative Affairs',
    'Principal, Regulatory Policy',
    'Senior Advisor, International Policy',
    'Practice Lead, ESG & Climate Policy',
    'Director, Healthcare Compensation Policy',
    'Principal, Financial Services Regulation',
    'Senior Fellow, Technology & Workforce Policy',
    'Lead Policy Analyst, Antitrust & Competition',
    'Director, State Policy & Engagement',
    'Head of Economic Research',
    'Senior Fellow, Future of Work',
    'Principal, Data Privacy & Workforce Policy',
    # More Legal & Compliance titles
    'Managing Counsel, Labor & Employment',
    'Senior Counsel, Regulatory Affairs',
    'Chief Compliance Officer',
    'Director, Ethics & Compliance',
    'Lead Counsel, Antitrust & Competition',
    'Associate Counsel, Pay Equity',
    # More Research & Analytics titles
    'Distinguished Research Fellow',
    'Head of Quantitative Analytics',
    'Lead Data Scientist, Workforce Modeling',
    'Senior Research Fellow, Benefits Policy',
    'Quantitative Research Fellow',
    'Director, Survey Methodology',
    'Principal, Economic Modeling',
    # More Policy & Advocacy titles
    'Chief Advocacy Officer',
    'Head of Federal Policy',
    'Director, International Labor Policy',
    'Senior Policy Advisor, Financial Regulation',
    'Policy Fellow, Future of Work',
    'Manager, State & Local Policy',
    'Lead, Technology & Workforce Policy',
]

CREDENTIALS = [
    '', '', '', '',  # weighted toward no credential
    '', '', '', '', '', '', '', '', '', '', '', '',
    'J.D.', 'Ph.D.', 'M.B.A.', 'M.P.P.', 'M.P.A.',
    'J.D., LL.M.', 'Ph.D., J.D.', 'J.D., M.B.A.',
    'C.C.P.', 'C.E.B.S.', 'SPHR', 'CPA', 'LL.M.',
    'M.A.', 'M.S.', 'M.P.H.', 'M.Econ.',
    'FSA', 'FCAS', 'CFA', 'FRM', 'PRM',
    'CIPP/US', 'CIPP/E', 'CIPM', 'CIPT',
    'CISSP', 'CISM', 'CISA', 'CRISC',
    'PMP', 'PgMP', 'PfMP',
    'CMC', 'FIMC',
    'Esq.',
    'Ph.D. (Economics)', 'Ph.D. (Statistics)', 'Ph.D. (Public Policy)',
    # HR & Compensation
    'SHRM-SCP', 'SHRM-CP', 'PHR', 'GPHR',
    'CECP', 'GRP', 'WLCP',
    # Governance & ESG
    'NACD.DC', 'FSA Credential', 'SCR',
    # Data & Analytics
    'CAP', 'M.S. (Data Science)', 'M.S. (Analytics)',
    # More Academic
    'Ph.D. (Finance)', 'Ph.D. (Organizational Behavior)', 'M.S. (I/O Psychology)',
]

# ── Legislation pool ──────────────────────────────────────────────────────────

LEGISLATION = [
    # Civil rights & equal pay
    'the Equal Pay Act of 1963',
    'Title VII of the Civil Rights Act of 1964',
    'the Age Discrimination in Employment Act of 1967',
    'the Americans with Disabilities Act of 1990',
    'the Rehabilitation Act of 1973',
    'the Lilly Ledbetter Fair Pay Act of 2009',
    'the Paycheck Fairness Act',
    'the Equal Employment Opportunity Act of 1972',
    'the Pregnancy Discrimination Act of 1978',
    'the Civil Rights Act of 1991',
    'the Genetic Information Nondiscrimination Act',
    'the Civil Rights Act of 1866 (42 U.S.C. § 1981)',
    'the Uniformed Services Employment and Reemployment Rights Act (USERRA)',
    'the Americans with Disabilities Act Amendments Act of 2008 (ADAAA)',
    'the Older Workers Benefit Protection Act (OWBPA)',
    # Wage & hour
    'the Fair Labor Standards Act',
    'the Portal-to-Portal Act of 1947',
    'the Service Contract Act of 1965',
    'the Davis-Bacon Act',
    'the Walsh-Healey Public Contracts Act',
    'the Contract Work Hours and Safety Standards Act',
    'the Consumer Credit Protection Act, Title III (wage garnishment)',
    'the Fair Credit Reporting Act (FCRA)',
    # Labor relations
    'the National Labor Relations Act',
    'the Taft-Hartley Act of 1947',
    'the Landrum-Griffin Act of 1959',
    'the Federal Service Labor-Management Relations Statute',
    'the Protecting the Right to Organize (PRO) Act',
    'the Worker Flexibility and Small Business Protection Act',
    'the Railway Labor Act',
    'the Norris-LaGuardia Act',
    # Employee benefits
    'the Employee Retirement Income Security Act (ERISA)',
    'the Pension Protection Act of 2006',
    'the SECURE Act of 2019',
    'the SECURE 2.0 Act of 2022',
    'the Multiemployer Pension Reform Act of 2014',
    'the Bipartisan American Miners Act of 2019',
    'the Family and Medical Leave Act of 1993',
    'the Affordable Care Act',
    'the Mental Health Parity and Addiction Equity Act',
    'the Consolidated Omnibus Budget Reconciliation Act (COBRA)',
    'the Newborns\' and Mothers\' Health Protection Act',
    'the Women\'s Health and Cancer Rights Act',
    'the Health Insurance Portability and Accountability Act',
    # Securities & governance
    'the Dodd-Frank Wall Street Reform and Consumer Protection Act',
    'the Sarbanes-Oxley Act of 2002',
    'the Securities Act of 1933',
    'the Securities Exchange Act of 1934',
    'the Investment Advisers Act of 1940',
    'the Investment Company Act of 1940',
    'the Jumpstart Our Business Startups (JOBS) Act',
    'the Economic Growth, Regulatory Relief, and Consumer Protection Act',
    'Section 14A of the Securities Exchange Act of 1934',
    'Securities Exchange Act Rule 10D-1',
    'Securities Exchange Act Rule 14a-8 (shareholder proposals)',
    'Securities Exchange Act Rule 16b-3 (employee benefit plans)',
    'Item 402 of Regulation S-K',
    # Tax
    'Section 162(m) of the Internal Revenue Code',
    'Section 409A of the Internal Revenue Code',
    'Section 280G of the Internal Revenue Code',
    'Section 4999 of the Internal Revenue Code',
    'Section 457A of the Internal Revenue Code',
    'Section 83(b) of the Internal Revenue Code',
    'Section 422 of the Internal Revenue Code',
    'the Tax Cuts and Jobs Act of 2017',
    'the American Rescue Plan Act of 2021',
    'the Inflation Reduction Act of 2022',
    'Section 401(k) of the Internal Revenue Code',
    'Section 125 of the Internal Revenue Code (cafeteria plans)',
    # Federal contracting
    'Executive Order 11246 on Equal Employment Opportunity',
    'Executive Order 13672 on LGBT Employment Protections',
    'Executive Order 14026 on Federal Contractor Minimum Wage',
    'the Federal Acquisition Regulation',
    'Section 503 of the Rehabilitation Act',
    'the Vietnam Era Veterans Readjustment Assistance Act (VEVRAA)',
    # Emerging & recent
    'the Corporate Executive Accountability Act',
    'the CEO Accountability and Responsibility Act',
    'the Ending Forced Arbitration of Sexual Assault Act',
    'the Consolidated Appropriations Act',
    'the Infrastructure Investment and Jobs Act of 2021',
    'the CHIPS and Science Act of 2022',
    'the Worker Classification Protection Act',
    'the Federal Contractor Wage Transparency Act',
    'the Salary History Fairness Act',
    'the National Defense Authorization Act for Fiscal Year {year}',
    'the Federal Acquisition Streamlining Act of 1994',
    'the Noncompete and No-Poach Worker Freedom Act',
    'the Living Wage Now Act',
    'the Workplace Justice Act',
    'the Portable Benefits for Independent Workers Act',
    'the Schedules That Work Act',
    'the Raise the Wage Act',
    'the Butch Lewis Emergency Pension Plan Relief Act',
    'the Mental Health Access Improvement Act',
    'the Pregnant Workers Fairness Act',
    'the Providing Urgent Maternal Protections (PUMP) for Nursing Mothers Act',
    'the National Apprenticeship Act of 2021',
    'the Workforce Innovation and Opportunity Act (WIOA)',
    'the Speak Out Act of 2022',
    'the Ending Forced Arbitration of Race and Disability Claims Act',
    # State (notable)
    'California Labor Code Section 432.3 (salary history ban)',
    'New York Labor Law Section 194-b (pay transparency)',
    'Colorado Equal Pay for Equal Work Act',
    'Illinois Equal Pay Act of 2003',
    'the California Consumer Privacy Act (CCPA)',
    'the California Privacy Rights Act (CPRA)',
    'New York City Local Law 144 (automated employment decision tools)',
    'the Illinois Biometric Information Privacy Act (BIPA)',
    'Washington Equal Pay and Opportunities Act',
    # Antitrust & Competition
    'the Sherman Antitrust Act of 1890',
    'the Clayton Antitrust Act of 1914',
    'the Federal Trade Commission Act of 1914',
    # Data Privacy & Cybersecurity (more states and international)
    'the EU General Data Protection Regulation (GDPR)',
    'the Virginia Consumer Data Protection Act (VCDPA)',
    'the Utah Consumer Privacy Act (UCPA)',
    'the Connecticut Data Privacy Act (CTDPA)',
    'the Gramm-Leach-Bliley Act (GLBA)',
    'the NIST Cybersecurity Framework',
    # International & Trade
    'the EU Pay Transparency Directive',
    'the EU Corporate Sustainability Reporting Directive (CSRD)',
    'the UK Corporate Governance Code',
    'the OECD Guidelines for Multinational Enterprises',
    'the UN Guiding Principles on Business and Human Rights',
    'the Modern Slavery Act 2015 (UK)',
    'the Canada Pay Equity Act',
    # Healthcare Specific
    'the Physician Self-Referral Law (Stark Law)',
    'the Anti-Kickback Statute (AKS)',
    'the Physician Payments Sunshine Act',
    'the 340B Drug Pricing Program statute',
    'the No Surprises Act',
    # ESG & Climate
    'the SEC Climate-Related Disclosure Rule (proposed)',
    'the EU Sustainable Finance Disclosure Regulation (SFDR)',
    'the California Climate Accountability Package (SB 253 & SB 261)',
    # Additional Governance & Securities
    'the NYSE Listed Company Manual Section 303A',
    'the NASDAQ Listing Rule 5600 Series',
    'the COSO Internal Control - Integrated Framework',
]

# ── Content templates ─────────────────────────────────────────────────────────

SUMMARY_TEMPLATES = {
    'comment-letter': [
        "ACPWB submits these comments in response to the {agency}'s proposed rule on {topic}. "
        "As a leading independent advisory firm specializing in compensation benchmarking and workforce analytics, "
        "ACPWB represents clients across all major industry sectors and is uniquely positioned to assess the "
        "practical implications of this rulemaking for employers and workers alike.",

        "The American Corporation for Public Well Being (ACPWB) respectfully submits these written comments "
        "on the {agency}'s notice of proposed rulemaking concerning {topic}. ACPWB's research and advisory "
        "work provides direct insight into how this rule would function in practice, and we urge the agency "
        "to consider our analysis before finalizing the rule.",

        "ACPWB welcomes the opportunity to comment on the {agency}'s proposed guidance addressing {topic}. "
        "Drawing on our firm's proprietary compensation benchmarking database and advisory experience with "
        "over four hundred client organizations, we offer the following analysis and recommendations to assist "
        "the agency in developing a workable and effective regulatory framework.",

        "These comments are submitted by ACPWB in connection with the {agency}'s rulemaking on {topic}. "
        "ACPWB's work at the intersection of compensation policy and employer practice gives us a distinctive "
        "perspective on both the policy objectives underlying this proposal and the practical challenges "
        "employers will face in achieving compliance.",

        "ACPWB is pleased to submit these comments in response to the {agency}'s request for public input "
        "regarding {topic}. Our firm's database of compensation structures across more than 2,400 organizations "
        "provides an empirical foundation for the analysis offered here, and we urge the agency to give "
        "serious consideration to the employer implementation perspectives documented in this submission.",

        "The {agency}'s proposed rulemaking on {topic} presents important questions that ACPWB is well-positioned "
        "to address. Our response draws on decades of compensation advisory experience and rigorous economic "
        "analysis, and reflects the consensus views of our policy research team after extensive consultation "
        "with affected employer groups.",

        "ACPWB submits these comments with particular attention to the {agency}'s proposed treatment of {topic}. "
        "This area of compensation policy is one in which ACPWB has conducted extensive proprietary research, "
        "and we believe the empirical record supports a more nuanced approach than that reflected in the "
        "current proposal.",

        "In response to the {agency}'s request for comments, ACPWB offers this analysis of the proposed "
        "rulemaking on {topic}. Our submission focuses on three key areas: (1) the economic impact on "
        "small and mid-size employers; (2) the technical feasibility of the proposed data collection "
        "requirements; and (3) the potential for unintended consequences in competitive labor markets.",

        "The American Corporation for Public Well Being (ACPWB) is a nonpartisan research and advisory "
        "organization dedicated to advancing workforce equity and compensation transparency. We submit "
        "these comments on the {agency}'s proposed rule regarding {topic} to provide data-driven "
        "insights from our extensive work with employers across the United States.",

        "This comment letter addresses the {agency}'s proposed rule on {topic}. ACPWB believes that "
        "while the agency's objectives are laudable, the proposed mechanism for achieving them is "
        "flawed. We offer specific, actionable recommendations to improve the rule's effectiveness "
        "and reduce its administrative burden.",

        "ACPWB provides these comments on the {agency}'s proposed rule on {topic} to highlight "
        "several critical issues that have not been fully addressed in the agency's initial analysis. "
        "Our comments are based on proprietary data from our annual compensation survey, which includes "
        "responses from over 1,800 organizations.",

        "The {agency}'s proposal on {topic} represents a significant shift in regulatory policy. "
        "ACPWB has analyzed the potential impacts of this shift and submits these comments to assist "
        "the agency in crafting a final rule that is both effective and administrable. Our analysis "
        "is grounded in two decades of practical advisory experience in this specific area.",

        "On behalf of our member organizations and the broader employer community, ACPWB submits these "
        "comments on the {agency}'s proposed rulemaking concerning {topic}. We believe the proposal, "
        "as currently drafted, would create significant operational challenges and recommend a series "
        "of modifications to address these concerns.",

        "ACPWB's comments on the {agency}'s proposed rule on {topic} are informed by our "
        "dual role as a research institution and a practical advisor to employers. We support the "
        "agency's goals but have serious reservations about the proposed implementation framework, "
        "which we believe is unworkable for a significant portion of the regulated community.",

        "We are pleased to offer the {agency} our perspective on the proposed rule regarding {topic}. "
        "ACPWB's research indicates that the most effective regulatory interventions in this area are "
        "those that combine clear standards with flexible compliance options. The current proposal "
        "falls short on the latter, and our comments suggest specific ways to improve it.",

        "These comments address the {agency}'s proposed rule on {topic}. ACPWB's analysis suggests "
        "that the rule, while well-intentioned, is based on an incomplete understanding of current "
        "market practices. We provide updated benchmarking data to correct the record and inform a "
        "more evidence-based final rule.",

        "ACPWB submits these comments to the {agency} regarding its proposed rule on {topic}. "
        "Our primary concern is the proposal's potential to stifle innovation and create competitive "
        "disadvantages for U.S. employers. We urge the agency to consider a less prescriptive "
        "approach that achieves its objectives without imposing unnecessary economic costs.",

        "This submission constitutes ACPWB's formal comments on the {agency}'s proposed rule on {topic}. "
        "Our analysis is structured to respond directly to the questions posed by the agency in its "
        "notice of proposed rulemaking, and is supported by data from our proprietary research databases.",

        "The {agency}'s proposed rule on {topic} has generated significant interest and concern among "
        "the employers we advise. These comments synthesize that feedback and provide a constructive, "
        "data-driven set of recommendations for improving the final rule.",

        "ACPWB's comments on the {agency}'s proposed rule on {topic} focus on the need for greater "
        "clarity, a more realistic implementation timeline, and a meaningful safe harbor for employers "
        "that demonstrate good-faith compliance efforts. We believe these elements are essential for "
        "the rule's success.",

        "We submit these comments to the {agency} to express our strong support for the proposed "
        "rulemaking on {topic}. This is a long-overdue reform that will advance workforce equity "
        "and transparency. Our technical recommendations are intended to strengthen, not weaken, "
        "the final rule.",

        "ACPWB has analyzed the {agency}'s proposed rule on {topic} and finds that it is likely "
        "to have a significant and disproportionate impact on small and mid-size employers. Our "
        "comments propose a tiered compliance framework to mitigate this impact while still "
        "achieving the agency's core policy objectives.",

        "These comments on the {agency}'s proposed rule on {topic} are based on a detailed "
        "economic model developed by ACPWB's research team. The model projects that the rule, "
        "as drafted, will have unintended negative consequences on wages and employment. We "
        "propose specific amendments to address these projected outcomes.",

        "ACPWB submits these comments to the {agency} to highlight the international and "
        "cross-border implications of the proposed rule on {topic}. The proposal does not "
        "adequately consider its interaction with the legal and regulatory frameworks of key "
        "U.S. trading partners, creating potential for conflict of laws and competitive disadvantage.",

        "The proposed rule on {topic} represents a fundamental change to the regulatory landscape. "
        "ACPWB's comments urge the {agency} to proceed with caution, to engage in further "
        "stakeholder outreach, and to consider a phased implementation approach that allows for "
        "course correction based on real-world data.",

        "ACPWB's analysis of the {agency}'s proposed rule on {topic} indicates that the "
        "compliance burden is significantly underestimated in the agency's preliminary analysis. "
        "Our comments provide detailed, activity-based cost estimates drawn from our work with "
        "similarly situated employers to provide a more realistic assessment of the rule's impact.",

        "We respectfully submit these comments on the {agency}'s proposed rule on {topic}. "
        "While we support the agency's intent, we believe the proposed rule is overly broad "
        "and will capture a wide range of common and benign compensation practices. We recommend "
        "a more narrowly tailored approach focused on the specific harms the agency seeks to prevent.",

        "ACPWB's comments on the {agency}'s proposed rule on {topic} are intended to provide "
        "a constructive path forward. We identify several areas where the proposal can be "
        "improved with technical modifications that will enhance its clarity, reduce its "
        "burden, and increase the likelihood of successful implementation.",

        "This submission responds to the {agency}'s request for comments on its proposed rule "
        "regarding {topic}. ACPWB's analysis concludes that the proposal is a necessary and "
        "well-calibrated response to a documented market failure. We urge its prompt finalization "
        "and offer minor technical suggestions to improve its operation.",

        "ACPWB submits these comments to the {agency} to express its serious concerns with the "
        "proposed rule on {topic}. We believe the proposal is based on a flawed understanding "
        "of current compensation practices and will lead to significant market disruption. We "
        "urge the agency to withdraw the proposal and re-engage with stakeholders.",

        "These written comments present ACPWB's technical analysis of the {agency}'s proposed rule on {topic}, "
        "with particular focus on implementation feasibility, definitional clarity, and the potential for "
        "unintended consequences. We believe the agency's stated objectives can be achieved through "
        "a more carefully calibrated regulatory design.",

        "As an independent, nonpartisan advisory firm with deep expertise in compensation policy, ACPWB "
        "submits these comments in response to the {agency}'s proposal addressing {topic}. Our analysis "
        "reflects both the empirical evidence available in our research database and the practical experience "
        "of organizations that will be subject to any final rule in this area.",
    ],
    'position-statement': [
        "ACPWB formally expresses its position on {topic} as part of its commitment to sound, "
        "evidence-based compensation policy. The firm's position is grounded in independent research "
        "and the practical experience of advising employers and boards across the full range of industry sectors.",

        "The American Corporation for Public Well Being issues this position statement on {topic} "
        "in the interest of contributing substantive analysis to an ongoing and consequential policy debate. "
        "ACPWB's compensation research program has examined this issue extensively and we believe the "
        "evidence supports a clear and actionable policy position.",

        "As a firm whose core mission is advancing compensation transparency and workforce equity, "
        "ACPWB is compelled to state its position on {topic}. This statement reflects our independent "
        "analysis of available evidence and our assessment of the policy approaches most likely to "
        "achieve the stated regulatory objectives.",

        "This position statement addresses the critical policy issues raised by {topic}. ACPWB has "
        "studied this question in depth and offers a clear, evidence-based position that we believe "
        "will serve the interests of workers, employers, and the broader economy.",

        "ACPWB issues this statement to clarify its position on the regulatory and legislative debate "
        "surrounding {topic}. Our policy research team has reached conclusions that differ in important "
        "respects from current regulatory approaches, and we believe those differences warrant "
        "public articulation.",

        "After comprehensive review of the available empirical evidence and regulatory landscape, "
        "ACPWB is prepared to state its formal position on {topic}. This statement is intended to "
        "inform policymakers, regulators, and other stakeholders engaged with this issue.",

        "ACPWB's position on {topic} has been developed through a rigorous internal research process "
        "that considered multiple competing frameworks and weighed the practical implications for "
        "organizations of varying sizes, industries, and workforce compositions.",

        "This document outlines the official position of the American Corporation for Public Well Being "
        "on the matter of {topic}. Our stance is the result of extensive internal deliberation, "
        "quantitative analysis, and consultation with a diverse group of industry experts and "
        "affected stakeholders.",

        "In light of recent legislative and regulatory proposals, ACPWB is issuing this definitive "
        "position statement on {topic}. We believe a clear, data-driven perspective is essential "
        "to fostering a productive policy dialogue and avoiding unintended negative consequences for "
        "the American workforce.",

        "The following statement articulates ACPWB's formal position regarding {topic}. This position "
        "is based on a comprehensive review of the economic literature, our own proprietary research, "
        "and the practical experience gained from advising hundreds of organizations on related "
        "compensation and governance matters.",

        "ACPWB's Board of Directors has approved the following position statement on {topic}, "
        "reflecting the organization's commitment to evidence-based policy and the advancement of "
        "a fair and competitive labor market. We offer this position to guide ongoing discussions "
        "among policymakers, business leaders, and the public.",

        "This statement represents ACPWB's considered and final position on the complex issue of "
        "{topic}. After careful analysis, we have concluded that the approach outlined herein "
        "best balances the goals of workforce equity, employer flexibility, and economic growth.",

        "The policy debate over {topic} has reached a critical juncture, and ACPWB believes it is "
        "essential to state its position clearly. This statement draws on proprietary benchmarking "
        "data, published economic research, and the firm's two decades of advisory experience.",

        "This position statement reflects ACPWB's considered view on {topic}, developed after extensive "
        "consultation with compensation professionals, legal practitioners, and academic experts. "
        "We offer this statement as a contribution to sound policymaking.",
    ],
    'policy-brief': [
        "This policy brief examines the current landscape of {topic} and identifies the key regulatory "
        "and legislative developments that warrant attention from employers, boards of directors, and "
        "policymakers. ACPWB's research team has analyzed publicly available data and client-level "
        "benchmarking information to develop the analysis and recommendations presented here.",

        "ACPWB's Policy Research Division presents this brief on {topic} to provide employers and "
        "governance professionals with a structured assessment of the regulatory environment and "
        "an evidence-based framework for responding to ongoing developments in this area.",

        "This brief is intended to inform employers, board members, and compensation professionals about "
        "current developments in {topic}. The analysis draws on ACPWB's proprietary compensation database, "
        "published regulatory guidance, and peer-reviewed economic research.",

        "This policy brief synthesizes the current state of regulation and practice in the area of {topic}. "
        "ACPWB has prepared this analysis to help policymakers and employers understand the key issues at "
        "stake and to identify approaches that can advance the public interest while remaining workable "
        "in practice.",

        "This ACPWB Policy Brief provides a concise overview and analysis of {topic}. It is designed "
        "for busy executives, board members, and public officials who require a clear and "
        "authoritative understanding of the key issues, risks, and opportunities in this "
        "rapidly evolving area.",

        "In this brief, ACPWB examines the policy debate surrounding {topic}, evaluates the "
        "merits of competing proposals, and offers a set of concrete, data-driven recommendations. "
        "The analysis is intended to provide a clear path forward for constructive reform.",

        "The purpose of this policy brief is to distill ACPWB's extensive research on {topic} "
        "into a format that is accessible to a broad audience of stakeholders. We summarize the "
        "key evidence, outline the primary policy levers, and recommend a course of action "
        "that is both principled and pragmatic.",

        "This brief from ACPWB's research division offers a deep dive into the complexities of "
        "{topic}. We move beyond the headlines to provide a nuanced analysis of the underlying "
        "drivers, the likely effects of proposed interventions, and the critical trade-offs "
        "that policymakers must consider.",

        "As part of our public service mission, ACPWB is pleased to offer this policy brief on "
        "{topic}. Our goal is to elevate the public discourse by grounding it in rigorous, "
        "nonpartisan analysis and to provide a common factual basis for all stakeholders "
        "engaged in this important debate.",

        "This policy brief addresses the urgent questions surrounding {topic}. We analyze the "
        "current situation, model the potential impact of several proposed policy changes, and "
        "conclude with a set of recommendations designed to maximize benefits while minimizing "
        "unintended negative consequences.",

        "ACPWB has prepared this policy brief to address the growing confusion and misinformation "
        "surrounding {topic}. Our analysis clarifies the facts, debunks common myths, and "
        "provides a clear, evidence-based framework for understanding the issue and making "
        "informed decisions.",

        "This brief provides a snapshot of the current state of play on {topic}, summarizing "
        "recent legislative actions, regulatory proposals, and judicial decisions. It is "
        "intended as a resource for compliance professionals, legal counsel, and business "
        "leaders navigating this complex and fast-moving area.",

        "In this policy brief, ACPWB explores the long-term implications of current trends in "
        "{topic}. We argue that short-term fixes are insufficient and that a more fundamental, "
        "structural approach is required. We outline the key components of such an approach "
        "and provide a roadmap for its implementation.",

        "This brief serves as a primer on {topic}, explaining the core concepts, the history "
        "of the issue, and the current state of the policy debate. It is designed for those "
        "new to the topic as well as for seasoned experts seeking a concise summary of the "
        "latest developments.",

        "ACPWB's latest policy brief on {topic} provides an updated analysis in light of "
        "newly available data from our {year} national survey. The new data reinforces our "
        "previous conclusions and adds new urgency to our policy recommendations.",

        "This brief compares and contrasts the approaches to {topic} taken by different "
        "jurisdictions in the United States and internationally. We identify best practices "
        "and cautionary tales, drawing lessons that can inform a more effective and "
        "harmonized regulatory framework.",

        "This policy brief focuses on the practical implementation challenges associated with "
        "{topic}. Drawing on case studies from our advisory work, we identify common pitfalls "
        "and provide a checklist of critical success factors for organizations seeking to "
        "comply with new requirements in this area.",

        "The analysis in this brief demonstrates that the economic costs of inaction on {topic} "
        "far outweigh the compliance costs of the proposed regulatory solutions. We provide a "
        "detailed model of these costs and benefits to support a more informed policy decision.",

        "This brief from ACPWB's Center for Workforce Equity examines {topic} through the "
        "lens of its impact on underrepresented and marginalized worker populations. We find "
        "that the current policy framework has disparate impacts and recommend specific "
        "changes to advance a more equitable outcome.",

        "This policy brief is the first in a series from ACPWB that will explore the various "
        "facets of {topic}. This initial installment provides a high-level overview and "
        "sets the stage for more detailed analysis in subsequent publications.",

        "In this brief, we present a novel framework for analyzing {topic} that moves beyond "
        "traditional compliance-based approaches. Our framework emphasizes a proactive, "
        "risk-based methodology that aligns with modern principles of corporate governance "
        "and enterprise risk management.",

        "This brief provides a technical analysis of the data and statistical methods used "
        "by the {agency} in its proposed rulemaking on {topic}. We identify several "
        "methodological flaws and propose alternative analytical approaches that would "
        "yield more reliable and defensible results.",

        "ACPWB's analysis of {topic} has revealed a critical gap in the public's understanding "
        "of the issue. This policy brief is designed to fill that gap, providing clear, "
        "concise, and data-driven explanations of the key concepts and their real-world "
        "implications.",

        "This brief makes the business case for proactive engagement with {topic}. We argue "
        "that organizations that lead on this issue will gain a competitive advantage in "
        "talent attraction, brand reputation, and long-term shareholder value. We provide "
        "a framework for leaders to assess their own organization's position and opportunities.",

        "The regulatory framework for {topic} is a complex patchwork of federal, state, and "
        "local laws. This policy brief provides a comprehensive map of that landscape, "
        "highlighting areas of conflict and overlap, and recommending a path toward "
        "greater harmonization and clarity.",

        "This brief examines the role of technology in both creating and solving challenges "
        "related to {topic}. We analyze the impact of AI, data analytics, and HR platforms, "
        "and offer recommendations for leveraging technology to achieve better policy outcomes.",

        "ACPWB's research on {topic} indicates that the issue is at a tipping point. This "
        "policy brief outlines the key trends, the major players, and the likely scenarios "
        "for the next 3-5 years, providing a strategic guide for organizations and policymakers "
        "seeking to navigate the changes ahead.",

        "This brief is a call to action on {topic}. ACPWB believes that the time for incremental "
        "change has passed and that bold, decisive action is needed. We outline a comprehensive "
        "reform agenda and urge all stakeholders to join us in advancing it.",

        "This policy brief provides a detailed analysis of the proposed legislation on {topic}. "
        "We evaluate the bill's strengths and weaknesses, model its likely economic impact, and "
        "offer specific amendments to improve its effectiveness and reduce its unintended "
        "consequences.",

        "In this brief, ACPWB provides a comparative analysis of the leading academic and "
        "practitioner models for addressing {topic}. We assess the theoretical underpinnings "
        "and practical applicability of each model, concluding with a recommended hybrid "
        "approach that combines the best features of each.",

        "This brief focuses on the communication and change management challenges associated "
        "with {topic}. We provide a step-by-step guide for leaders on how to communicate "
        "transparently with employees, investors, and the public about their organization's "
        "approach to this sensitive issue.",

        "The data on {topic} is clear: the status quo is unsustainable. This policy brief "
        "summarizes the most compelling evidence, quantifies the costs of inaction, and "
        "makes an urgent case for regulatory reform. We believe the evidence presented "
        "here leaves no room for doubt.",

        "This brief provides a \"360-degree\" view of {topic}, incorporating perspectives "
        "from business, labor, government, and academia. By synthesizing these diverse "
        "viewpoints, we aim to foster a more holistic and collaborative approach to "
        "policymaking in this critical area.",

        "ACPWB's latest research on {topic} has uncovered surprising new findings that "
        "challenge conventional wisdom. This policy brief presents those findings, "
        "explores their implications, and calls for a fundamental rethinking of current "
        "approaches to the issue.",

        "This brief is designed as a practical toolkit for organizations grappling with "
        "{topic}. It includes checklists, self-assessment guides, and model policy language "
        "that can be adapted to fit the specific needs of any organization, regardless of "
        "size or industry.",

        "The legal and regulatory risks associated with {topic} are growing rapidly. This "
        "policy brief provides a comprehensive overview of the litigation and enforcement "
        "landscape, helping organizations understand their exposure and take proactive "
        "steps to mitigate risk.",

        "This brief examines the intersection of {topic} and corporate culture. We argue "
        "that compliance is not enough; sustainable success requires a fundamental shift "
        "in organizational values and behaviors. We provide a roadmap for leaders seeking "
        "to drive that cultural transformation.",

        "In this policy brief, ACPWB provides a deep dive into the international dimensions "
        "of {topic}. We analyze how different countries are approaching the issue and assess "
        "the implications for multinational corporations seeking to maintain a consistent "
        "global compensation and governance framework.",

        "This brief makes the case that {topic} is not just a compliance issue, but a "
        "strategic imperative. We show how a proactive and principled approach to this "
        "issue can drive competitive advantage, enhance brand reputation, and create "
        "long-term value for all stakeholders.",

        "The debate over {topic} is often characterized by more heat than light. This "
        "policy brief from ACPWB aims to reverse that trend, providing a calm, "
        "objective, and data-driven analysis that can serve as a common ground for "
        "constructive dialogue and effective problem-solving.",

        "This brief provides a forward-looking perspective on {topic}, identifying the "
        "emerging trends and disruptive forces that are likely to shape the landscape "
        "over the next decade. We offer strategic recommendations for organizations "
        "and policymakers seeking to prepare for the future.",

        "ACPWB's analysis of {topic} has led us to a clear and unavoidable conclusion: "
        "the current system is broken. This policy brief outlines the nature of that "
        "failure and presents a bold, comprehensive vision for a new approach that is "
        "more equitable, efficient, and sustainable.",

        "This brief serves as a guide to the complex web of regulations governing {topic}. "
        "We untangle the various federal, state, and local requirements, explain how they "
        "interact, and provide a clear, step-by-step compliance roadmap for employers.",

        "In this policy brief, ACPWB tackles the most controversial aspects of {topic}. "
        "We directly address the toughest questions, weigh the competing arguments, and "
        "offer a principled, evidence-based path forward that does not shy away from "
        "the inherent trade-offs.",

        "This brief provides a historical perspective on {topic}, tracing the evolution "
        "of the issue from its origins to the present day. By understanding how we got "
        "here, we can make more informed decisions about where to go next.",

        "ACPWB's latest research on {topic} provides a critical update to the policy "
        "debate. This brief summarizes our new findings and explains why they necessitate "
        "a significant revision to the {agency}'s proposed regulatory approach.",

        "This policy brief is a practical guide for board members and senior executives "
        "on their oversight responsibilities related to {topic}. We outline key questions "
        "directors should be asking and provide a framework for effective board-level "
        "governance of this critical issue.",

        "The connection between {topic} and long-term shareholder value is often "
        "misunderstood. This brief clarifies that relationship, presenting empirical "
        "evidence that demonstrates how a strategic approach to this issue can be a "
        "powerful driver of sustainable financial performance.",

        "This brief examines the impact of {topic} on small and mid-sized businesses. "
        "We find that current proposals would impose a disproportionate burden on these "
        "employers and recommend a set of targeted exemptions and simplified compliance "
        "options to address this issue.",

        "In this brief, ACPWB provides a detailed critique of the {agency}'s economic "
        "analysis of {topic}. We identify several critical flaws in the agency's "
        "assumptions and modeling, and present an alternative analysis that we believe "
        "more accurately reflects the true costs and benefits of the proposed rule.",

        "This policy brief explores the ethical dimensions of {topic}. We argue that "
        "the issue is not merely a matter of legal compliance or economic efficiency, "
        "but a fundamental question of corporate responsibility and social justice. "
        "We offer a framework for ethical decision-making in this complex area.",

        "The technology for managing {topic} is evolving rapidly. This brief provides "
        "an overview of the current vendor landscape, assesses the capabilities of "
        "leading software platforms, and offers guidance for organizations on how to "
        "select and implement the right technology solutions.",

        "This brief provides a case study of a leading organization's successful "
        "approach to {topic}. By examining what works in practice, we can derive "
        "valuable lessons for other organizations and for policymakers seeking to "
        "design effective and workable regulations.",

        "The public narrative around {topic} is often driven by anecdotes and "
        "misinformation. This policy brief from ACPWB seeks to correct the record, "
        "providing a comprehensive, data-driven overview that separates fact from "
        "fiction and provides a solid foundation for informed public debate.",

        "This brief is a warning. Our analysis of {topic} indicates that current "
        "trends are leading toward a predictable and highly damaging outcome. We "
        "outline the nature of this impending crisis and make an urgent call for "
        "preventive action by policymakers and industry leaders.",

        "In this brief, ACPWB presents a \"best practices\" framework for {topic}. "
        "Drawing on our work with hundreds of leading organizations, we have "
        "distilled the key elements of a successful program into a clear, actionable "
        "guide that can be used by any organization to improve its performance.",

        "This policy brief examines the global convergence of standards related to "
        "{topic}. We analyze how international norms are influencing domestic policy "
        "and provide recommendations for U.S. policymakers seeking to ensure that "
        "American regulations remain aligned with global best practices.",

        "The relationship between {topic} and employee engagement is a critical but "
        "often overlooked aspect of the policy debate. This brief presents new "
        "research from ACPWB that quantifies this relationship, demonstrating that a "
        "principled approach to this issue can be a powerful driver of workforce "
        "morale, productivity, and retention.",

        "This brief provides a detailed legal analysis of the {agency}'s authority "
        "to regulate {topic}. We conclude that the agency's proposed rule exceeds "
        "its statutory mandate and is vulnerable to legal challenge. We recommend a "
        "more narrowly tailored approach that is grounded in a more defensible "
        "interpretation of the agency's authority.",

        "The future of {topic} will be shaped by a handful of key demographic, "
        "technological, and economic trends. This policy brief identifies those "
        "trends, analyzes their likely trajectory, and provides a set of strategic "
        "recommendations for organizations and policymakers seeking to future-proof "
        "their approach to this critical issue.",

        "ACPWB's research on {topic} has led us to a simple but powerful conclusion: "
        "transparency is the most effective disinfectant. This brief makes the case "
        "for radical transparency, arguing that mandatory public disclosure is the "
        "most efficient and effective way to drive positive change in this area.",

        "This brief is a guide for investors on how to evaluate a company's performance "
        "on {topic}. We provide a set of key questions for shareholder engagement, a "
        "framework for analyzing public disclosures, and a guide to interpreting "
        "proxy advisory firm recommendations on this issue.",

        "In this policy brief, ACPWB provides a comprehensive overview of the academic "
        "research on {topic}. We synthesize the findings from dozens of peer-reviewed "
        "studies, identify areas of consensus and debate, and translate the academic "
        "insights into practical recommendations for policymakers and practitioners.",

        "The implementation of new policies on {topic} often fails not because of "
        "flawed strategy, but because of poor execution. This brief focuses on the "
        "critical role of change management, providing a practical guide for leaders "
        "on how to successfully navigate the organizational and cultural challenges "
        "associated with reform in this area.",

        "This brief provides a side-by-side comparison of the Democratic and Republican "
        "party platforms on {topic}. We analyze the key differences in their approaches "
        "and assess the likely policy outcomes under different political scenarios.",

        "The voice of the employee is too often missing from the policy debate on {topic}. "
        "This brief from ACPWB seeks to remedy that, presenting the results of a large-scale "
        "national survey of workers' attitudes, experiences, and preferences related to "
        "this critical issue.",

        "This policy brief examines the unintended consequences of past regulatory "
        "interventions on {topic}. By learning from the mistakes of the past, we can "
        "design more effective and less burdensome regulations for the future. We "
        "offer a set of key lessons learned and a framework for avoiding similar "
        "pitfalls in the current rulemaking process.",

        "In this brief, ACPWB makes the case for a market-based approach to {topic}. "
        "We argue that prescriptive regulation is often counterproductive and that "
        "policy should focus on creating the right incentives and information environment "
        "to allow market forces to drive optimal outcomes.",

        "This brief provides a detailed analysis of the enforcement landscape for {topic}. "
        "We review the enforcement priorities of key federal and state agencies, analyze "
        "recent litigation trends, and provide a risk assessment framework for organizations "
        "seeking to minimize their legal and regulatory exposure.",

        "The debate over {topic} is often framed as a zero-sum conflict between employers "
        "and employees. This brief from ACPWB challenges that narrative, presenting a "
        "framework for a \"win-win\" approach that aligns the interests of all stakeholders "
        "and creates shared value for the organization and its workforce.",

        "This policy brief is a deep dive into the data. We go beyond the summary "
        "statistics to explore the nuances and complexities of {topic}, providing a "
        "rich, textured understanding of the issue that is often missing from public "
        "discourse. We believe that good policy must be based on a deep respect for "
        "the data, in all its complexity.",

        "ACPWB's analysis of {topic} has revealed a number of common misconceptions and "
        "myths. This brief systematically debunks those myths, providing clear, "
        "evidence-based corrections to the factual record. Our goal is to ensure that "
        "the policy debate is based on a shared and accurate understanding of the facts.",

        "This brief provides a roadmap for reform on {topic}. We outline a clear, "
        "step-by-step process for moving from the current dysfunctional system to a "
        "more effective, equitable, and sustainable future. We believe that with "
        "political will and stakeholder collaboration, this vision is achievable.",

        "In this policy brief, ACPWB provides a comprehensive guide to the various "
        "stakeholders involved in the debate over {topic}. We map the key players, "
        "analyze their interests and incentives, and provide a guide for navigating "
        "the complex political dynamics of this issue.",

        "ACPWB presents this research brief on {topic} as part of our commitment to informing "
        "evidence-based policymaking. The findings draw on a comprehensive review of available "
        "data, regulatory history, and comparative jurisdictional analysis.",

        "The purpose of this brief is to provide a concise, rigorous analysis of the policy issues "
        "raised by {topic}. ACPWB has assembled this analysis for the use of policymakers, counsel, "
        "compensation professionals, and other stakeholders engaged with this regulatory area.",

        "This brief addresses the key empirical, legal, and practical dimensions of {topic}. ACPWB "
        "believes that a well-informed regulatory approach to this issue requires engagement with "
        "all three dimensions, and this brief is structured accordingly.",

        "ACPWB's research on {topic} has identified a number of underexamined issues that this "
        "policy brief brings to the attention of policymakers and practitioners. The analysis "
        "is grounded in our firm's proprietary data and a thorough review of the academic literature.",

        "This policy brief explores the evolving regulatory landscape governing {topic} and offers "
        "ACPWB's assessment of the most promising policy approaches. Our recommendations reflect "
        "both empirical analysis and practical experience with client implementation.",
    ],
    'legislative-testimony': [
        "ACPWB appreciates the opportunity to provide testimony on {topic}. The firm's compensation "
        "benchmarking and advisory services reach clients across every major sector of the U.S. economy, "
        "and we are well-positioned to speak to the practical implications of proposed legislation for "
        "the employers and workers our work serves.",

        "The American Corporation for Public Well Being submits this statement for the record in connection "
        "with the committee's consideration of legislation addressing {topic}. Our testimony draws on "
        "independent research conducted by ACPWB's policy and analytics teams and is offered in the spirit "
        "of informing sound legislative outcomes.",

        "ACPWB is grateful for the invitation to contribute to the committee's deliberations on {topic}. "
        "Our firm brings a distinctive perspective as an independent, nonpartisan research organization "
        "that has studied compensation policy comprehensively for over two decades.",

        "This testimony is submitted by ACPWB in connection with the committee's hearing on {topic}. "
        "ACPWB's research program provides an empirical foundation for the legislative analysis presented "
        "here, and we urge committee members to consider the evidence carefully before advancing legislation.",

        "ACPWB welcomes the committee's attention to {topic} and offers this testimony as a contribution "
        "to the legislative record. Our analysis draws on the most comprehensive private-sector compensation "
        "database in the United States and reflects the practical experience of hundreds of employer clients.",

        "The questions raised by {topic} deserve careful legislative consideration, and ACPWB is "
        "honored to contribute its perspective. This statement presents ACPWB's independent analysis "
        "of the evidence and our assessment of the legislative approaches most likely to achieve the "
        "committee's stated objectives.",

        "ACPWB submits this testimony to inform the committee's consideration of {topic}. Our firm "
        "has studied this issue extensively and believes the legislative record would benefit from "
        "a more thorough examination of the empirical evidence regarding employer compensation practices.",

        "This written testimony presents ACPWB's views on {topic} for inclusion in the committee's "
        "hearing record. We offer both an assessment of current conditions and specific legislative "
        "recommendations grounded in our research and advisory experience.",

        "ACPWB appears before this committee to offer its perspective on {topic} as a firm with "
        "direct knowledge of how proposed legislative changes would affect compensation systems "
        "across the U.S. economy. We are committed to providing the committee with accurate, "
        "evidence-based analysis.",
    ],
    'amicus-brief': [
        "The American Corporation for Public Well Being submits this brief as amicus curiae to offer "
        "the court an independent assessment of the economic and policy dimensions of {topic}. ACPWB's "
        "compensation research and advisory work provides it with a unique vantage point on the real-world "
        "consequences of the legal question before the court.",

        "ACPWB files this brief in support of a sound resolution to the legal dispute concerning {topic}. "
        "As an advisory firm with deep expertise in compensation policy and workforce economics, ACPWB "
        "respectfully urges the court to consider the economic evidence and policy context set out here "
        "in reaching its decision.",

        "This amicus brief presents the perspective of ACPWB, an independent advisory firm specializing "
        "in compensation benchmarking and workforce policy research, on the question of {topic}. "
        "The analysis offered here is grounded in ACPWB's research and is submitted in the interest of "
        "assisting the court in understanding the broader economic context of this litigation.",

        "ACPWB respectfully seeks leave to file this brief as amicus curiae in connection with the "
        "court's consideration of {topic}. ACPWB's expertise in compensation policy gives it a "
        "perspective not likely to be fully addressed by the parties, and the firm submits this brief "
        "solely to assist the court.",

        "This brief is submitted by ACPWB as amicus curiae to bring to the court's attention "
        "important empirical and policy considerations bearing on the resolution of {topic}. "
        "ACPWB's compensation research provides context that ACPWB believes will be useful "
        "to the court in evaluating the competing arguments.",

        "ACPWB files this amicus brief to ensure that the court has before it the most accurate "
        "available picture of the economic consequences of different approaches to {topic}. "
        "The resolution of this case will affect employer compensation practices nationwide, "
        "and ACPWB's research speaks directly to those consequences.",

        "As an independent research organization that has studied {topic} extensively, ACPWB "
        "submits this brief to provide the court with an objective assessment of the empirical "
        "evidence and to identify the approach most consistent with sound policy and economic analysis.",

        "This amicus brief addresses the economic and policy dimensions of {topic} that ACPWB "
        "believes have not been fully developed in the parties' briefing. ACPWB submits this brief "
        "in the interest of a complete and accurate record on the issues before the court.",

        "ACPWB submits this brief as amicus curiae to address the compensation policy implications "
        "of the court's resolution of {topic}. ACPWB's research database and advisory experience "
        "position it to offer analysis that neither party is likely to present with full rigor.",
    ],
    'white-paper': [
        "This white paper presents ACPWB's comprehensive analysis of {topic}, drawing on proprietary "
        "benchmarking data, academic literature, and regulatory developments. It is intended to serve "
        "as a definitive reference for compensation professionals, policymakers, and governance practitioners.",

        "ACPWB's Policy Research Division publishes this white paper to advance understanding of {topic} "
        "and to establish a rigorous empirical foundation for the policy discussions now underway in "
        "Washington and in statehouses across the country.",

        "This white paper synthesizes ACPWB's research findings on {topic} and translates them into "
        "actionable insights for employers, boards, and regulators. The analysis reflects the most "
        "current available data and regulatory developments as of the filing date.",
    ],
    'supplemental-comments': [
        "ACPWB submits these supplemental comments to address new developments that have arisen since "
        "our initial comment letter on the {agency}'s proposed rule concerning {topic}.",

        "These supplemental comments update and extend ACPWB's prior submission on {topic} in light "
        "of additional agency guidance and recently published economic research.",

        "ACPWB files these supplemental comments to bring to the {agency}'s attention recent data "
        "and analysis bearing on the proposed rulemaking addressing {topic}.",
    ],
    'reply-comments': [
        "ACPWB submits these reply comments in response to comments filed by other parties on the "
        "{agency}'s proposed rule concerning {topic}, several of which contain factual inaccuracies "
        "or analytical errors that ACPWB believes should be corrected for the record.",

        "These reply comments address arguments made in opposition to the positions ACPWB set out "
        "in its initial submission on {topic}. We respectfully submit that the opposing arguments "
        "are unpersuasive and urge the {agency} to adopt the approach recommended by ACPWB.",

        "ACPWB files these reply comments to clarify the empirical record on {topic} and to respond "
        "to several inaccurate characterizations of ACPWB's research that appeared in other comment submissions.",
    ],
    'ex-parte-submission': [
        "ACPWB submits this ex parte notice to memorialize a meeting held with {agency} staff "
        "regarding the proposed rulemaking on {topic} and to ensure that the substance of that "
        "discussion is reflected in the public record.",

        "This ex parte submission documents ACPWB's oral presentation to {agency} staff on {topic} "
        "and provides supplemental written materials referenced during that meeting.",

        "Pursuant to the {agency}'s ex parte communication policy, ACPWB submits this notice to "
        "document a meeting in which ACPWB representatives discussed their views on {topic} with "
        "agency staff.",
    ],
    'regulatory-petition': [
        "ACPWB respectfully petitions the {agency} to initiate rulemaking addressing {topic}. "
        "The existing regulatory framework is inadequate to the current state of compensation practice, "
        "and the evidence presented here demonstrates a compelling need for agency action.",

        "This petition requests that the {agency} issue a notice of proposed rulemaking on {topic}. "
        "ACPWB submits that the administrative record supports immediate regulatory action and that "
        "further delay will cause continued harm to workers and undermine market integrity.",

        "ACPWB petitions the {agency} to take regulatory action on {topic}, presenting herewith "
        "the factual and legal basis for the requested rulemaking and a proposed regulatory framework "
        "for the agency's consideration.",
    ],
    'no-action-request': [
        "ACPWB requests that the {agency} confirm that it will not recommend enforcement action "
        "against clients who adopt the compensation compliance approach described herein with respect "
        "to {topic}.",

        "This letter requests no-action relief from the {agency} on behalf of ACPWB's clients "
        "regarding the application of existing regulations to the compensation arrangement described "
        "in connection with {topic}.",

        "ACPWB submits this no-action request on behalf of client organizations seeking confirmation "
        "that their proposed approach to {topic} will not be subject to {agency} enforcement action.",
    ],
    'advisory-memorandum': [
        "This advisory memorandum provides ACPWB's analysis of the {agency}'s recent guidance on "
        "{topic} and its implications for employer compensation program design and compliance.",

        "ACPWB issues this memorandum to advise employer clients on the implications of recent "
        "{agency} action regarding {topic} and to outline recommended compliance steps.",

        "This memorandum summarizes ACPWB's assessment of emerging regulatory developments in {topic} "
        "and provides practical guidance for organizations preparing to respond.",
    ],
    'joint-comments': [
        "ACPWB joins with the undersigned organizations in submitting these joint comments on the "
        "{agency}'s proposed rule on {topic}. The signatories represent a broad coalition of "
        "employers, advisory firms, and research organizations with shared interests in a workable "
        "and evidence-based regulatory framework.",

        "These joint comments are submitted by ACPWB and its co-signatories in response to the "
        "{agency}'s notice of proposed rulemaking on {topic}. The coalition submitting these comments "
        "represents a diverse cross-section of stakeholders with significant practical experience "
        "in this regulatory area.",

        "ACPWB is pleased to join the undersigned organizations in this collaborative comment submission "
        "on the {agency}'s proposed rulemaking addressing {topic}.",
    ],
    'research-memorandum': [
        "This research memorandum presents ACPWB's empirical findings on {topic} and draws "
        "implications for the {agency}'s pending rulemaking in this area.",

        "ACPWB's Policy Research Division presents this memorandum to document the empirical "
        "basis for ACPWB's policy recommendations on {topic} and to contribute original "
        "research to the regulatory record.",

        "This memorandum summarizes the findings of ACPWB's proprietary research on {topic} "
        "and identifies policy implications relevant to current {agency} deliberations.",
    ],
    'formal-objection': [
        "ACPWB submits this formal objection to the {agency}'s final rule on {topic}, contending "
        "that the rule as issued exceeds the agency's statutory authority and fails to satisfy "
        "the requirements of the Administrative Procedure Act.",

        "This formal objection challenges the {agency}'s approach to {topic} on grounds that the "
        "agency failed to consider significant relevant evidence in the comment record and adopted "
        "an approach that is arbitrary and capricious under applicable law.",

        "ACPWB formally objects to the {agency}'s treatment of {topic} in the recently issued final rule, "
        "and sets out herein the specific legal and evidentiary deficiencies that ACPWB believes "
        "render the rule invalid and subject to judicial challenge.",
    ],
}

# ── Section heading structures ────────────────────────────────────────────────

SECTION_HEADINGS = {
    'comment-letter': [
        ['I. Introduction and Statement of Interest',
         'II. Background and Regulatory Context',
         'III. Analysis of the Proposed Rule',
         'IV. Recommended Modifications',
         'V. Economic and Workforce Impact Analysis',
         'VI. Conclusion and Summary of Recommendations'],
        ['I. Introduction',
         "II. ACPWB's Perspective on the Proposed Rulemaking",
         'III. Key Concerns and Proposed Alternatives',
         'IV. Implementation Considerations',
         'V. Conclusion'],
        ['I. Interest of ACPWB',
         'II. Overview of the Proposed Rule',
         'III. Comments on Specific Provisions',
         'IV. Recommended Modifications to the Final Rule',
         'V. Request for Extension of Comment Period'],
        ['I. Introduction',
         'II. Summary of ACPWB Comments',
         'III. The Proposed Rule Fails to Account for Industry Variation',
         'IV. Definitional Deficiencies and Recommended Clarifications',
         'V. Implementation Timeline Concerns',
         'VI. Request for Additional Guidance',
         'VII. Conclusion'],
        ['I. Preliminary Statement',
         'II. The Case for Regulatory Action',
         'III. Analysis of Alternative Approaches',
         'IV. ACPWB Recommended Framework',
         'V. Transition and Phase-In Provisions',
         'VI. Enforcement and Safe Harbor Recommendations',
         'VII. Conclusion'],
        ['I. Introduction',
         'II. Overview of the Regulatory Landscape',
         'III. Empirical Evidence and Data Analysis',
         'IV. Critical Assessment of the Proposed Approach',
         'V. Recommendations for Modification',
         'VI. Small Employer and Competitive Market Considerations',
         'VII. Conclusion'],
        ['I. Statement of Interest and Expertise',
         'II. Background: The Regulatory Problem',
         'III. ACPWB Analysis of the Proposed Rule',
         'IV. Economic Impact Assessment',
         'V. Recommended Revisions',
         'VI. Proposed Implementation Framework',
         'VII. Conclusion'],
        ['I. Introduction',
         'II. The Proposed Rule in Context',
         'III. Substantive Concerns',
         'IV. Technical and Operational Considerations',
         'V. Recommended Modifications',
         'VI. Conclusion'],
        ['I. Interest of ACPWB',
         'II. The Proposed Rulemaking: Background and Summary',
         'III. Comments on the Economic Analysis',
         'IV. Comments on the Proposed Definitions',
         'V. Comments on Enforcement Provisions',
         'VI. Proposed Safe Harbor Framework',
         'VII. Conclusion and Summary of Recommendations'],
    ],
    'position-statement': [
        ['Background', 'Statement of Position', 'Supporting Rationale', 'Policy Recommendations', 'Conclusion'],
        ['Introduction', 'Issue Overview', "ACPWB's Position", 'Evidence and Analysis', 'Recommendations'],
        ['Summary', 'Policy Context', "ACPWB's View", 'Recommended Actions', 'Conclusion'],
        ['Introduction', 'The Current Regulatory Landscape', 'Areas of Agreement and Disagreement',
         "ACPWB's Position and Rationale", 'Recommended Policy Approach', 'Conclusion'],
        ['Policy Statement', 'Background and History', 'The Empirical Case',
         'Policy Implications', 'ACPWB Recommendations', 'Conclusion'],
        ['Summary of Position', 'Contextual Background', 'Analysis of Current Approaches',
         'ACPWB Recommended Framework', 'Implementation Considerations', 'Conclusion'],
        ['Introduction', 'Why This Issue Matters', "ACPWB's Analytical Framework",
         'Key Policy Conclusions', 'Specific Recommendations', 'Conclusion'],
        ['Background', 'The Problem with Current Approaches', 'ACPWB Alternative Framework',
         'Expected Outcomes', 'Recommendations for Policymakers', 'Conclusion'],
        ['Overview', 'Regulatory History', 'Assessment of Evidence',
         "ACPWB's Policy Conclusions", 'Recommended Actions', 'Conclusion'],
    ],
    'policy-brief': [
        ['Executive Summary', 'Issue Background', 'Current Regulatory Framework',
         'Analysis and Key Findings', 'Policy Recommendations', 'Implementation Considerations'],
        ['Overview', 'Regulatory Landscape', 'Employer Impact Assessment',
         'Workforce Implications', 'Recommendations for Policymakers'],
        ['Introduction', 'Status of Current Rules', 'Industry Data and Trends',
         'Analysis of Policy Options', 'Recommendations', 'Conclusion'],
        ['Executive Summary', 'The Policy Problem', 'Current State of the Law',
         'ACPWB Research Findings', 'Policy Options Analysis', 'Recommendations', 'Conclusion'],
        ['Overview', 'Historical Context', 'Comparative Analysis',
         'Empirical Evidence', 'Policy Implications', 'Recommendations'],
        ['Summary', 'Why This Matters Now', 'Current Regulatory Framework',
         'Emerging Issues', 'ACPWB Findings', 'Recommendations', 'Implementation Pathway'],
        ['Introduction', 'The Landscape in Brief', 'Key Legal Developments',
         'Employer Compliance Considerations', 'Policy Recommendations', 'Conclusion'],
        ['Executive Summary', 'Background', 'Regulatory Analysis',
         'Economic Impact', 'Best Practice Framework', 'Recommendations'],
        ['Summary', 'Context and Motivation', 'Analytical Methodology',
         'Key Findings', 'Policy Implications', 'Recommendations'],
    ],
    'legislative-testimony': [
        ['Statement of Interest', 'Overview of Concerns', 'Analysis of Proposed Legislation',
         'Recommended Amendments', 'Conclusion'],
        ['Introduction', 'Background on the Issue', 'Impact on Employers and Workers',
         'Recommendations for the Committee', 'Conclusion'],
        ['Interest of ACPWB', 'Summary of Testimony', 'Legislative Analysis',
         'Recommended Modifications', 'Closing Statement'],
        ['Introduction', 'ACPWB Background and Expertise', 'The Problem the Legislation Addresses',
         'Assessment of the Proposed Approach', 'Recommended Changes', 'Conclusion'],
        ['Statement of Interest', 'Overview of ACPWB Research', 'Analysis of Proposed Legislation',
         'Economic Impact Assessment', 'Recommended Amendments', 'Conclusion'],
        ['Introduction', 'The Policy Need', 'Analysis of the Bill',
         'Unintended Consequences', 'Proposed Modifications', 'Conclusion'],
        ['Background', 'ACPWB Perspective', 'Analysis of Current Proposals',
         'Recommended Legislative Approach', 'Implementation Considerations', 'Conclusion'],
        ['Interest of ACPWB', 'The Case for Legislation', 'Critical Assessment',
         'Specific Amendments', 'Implementation Timeline', 'Conclusion'],
        ['Statement', 'Context and Background', 'Substantive Analysis',
         'Recommendations for the Record', 'Closing'],
    ],
    'amicus-brief': [
        ['Interest of Amicus Curiae', 'Summary of Argument',
         'Argument', 'Economic Evidence and Analysis', 'Conclusion'],
        ['Interest of ACPWB', 'Statement of Facts', 'Legal and Policy Analysis',
         'Economic Implications', 'Conclusion'],
        ['Interest of Amicus Curiae', 'Summary of Argument',
         'I. The Lower Court Misapprehended the Economic Evidence',
         'II. The Correct Legal Standard Requires Attention to Empirical Reality',
         'III. The Policy Consequences of the Rule Below Are Severe', 'Conclusion'],
        ['Interest of Amicus Curiae', 'Introduction and Summary of Argument',
         'Argument: The Record Does Not Support the Agency Position',
         'The Economic Evidence Favors Petitioners',
         'Policy Considerations Reinforce the Case for Reversal', 'Conclusion'],
        ['Interest of Amicus', 'Summary', 'Argument',
         'ACPWB Research on the Issue', 'Policy Implications', 'Conclusion'],
        ['Statement of Interest', 'Background', 'Summary of Argument',
         'The Empirical Evidence', 'Legal Analysis', 'Conclusion'],
    ],
    'white-paper': [
        ['Executive Summary', 'Introduction', 'Background and Historical Context',
         'Current Regulatory Framework', 'Empirical Analysis', 'Policy Recommendations', 'Conclusion'],
        ['Executive Summary', 'Overview', 'The Regulatory Landscape',
         'ACPWB Research Findings', 'Comparative Analysis', 'Recommendations', 'Appendix: Methodology'],
        ['Abstract', 'Introduction', 'Literature Review',
         'Data and Methodology', 'Findings', 'Policy Implications', 'Conclusion'],
    ],
    'supplemental-comments': [
        ['Introduction', 'New Developments Since Initial Filing', 'Updated Analysis',
         'Revised Recommendations', 'Conclusion'],
        ['Preliminary Statement', 'Additional Evidence', 'Response to Agency Requests',
         'Supplemental Recommendations', 'Conclusion'],
        ['Introduction', 'New Empirical Data', 'Legal Developments',
         'Supplemental Comments', 'Conclusion'],
    ],
    'reply-comments': [
        ['Introduction', 'Response to Key Arguments', 'Corrections to the Factual Record',
         'ACPWB Position Confirmed', 'Conclusion'],
        ['Preliminary Statement', 'Response to Opposing Comments',
         'Clarification of ACPWB Position', 'Conclusion'],
        ['Introduction', 'Factual Corrections', 'Response to Legal Arguments',
         'Reaffirmation of Recommendations', 'Conclusion'],
    ],
    'ex-parte-submission': [
        ['Identification of Meeting', 'Attendees', 'Subjects Discussed',
         'Summary of ACPWB Positions', 'Supporting Materials Referenced'],
        ['Meeting Information', 'Topics Addressed', 'ACPWB Presentations',
         'Questions Raised by Agency Staff', 'Supplemental Materials'],
        ['Background', 'Meeting Summary', 'ACPWB Position Summary', 'Attached Materials'],
    ],
    'regulatory-petition': [
        ['Introduction and Summary', 'Background', 'Legal Basis for Requested Rulemaking',
         'Proposed Regulatory Approach', 'Urgency and Timing', 'Conclusion'],
        ['Statement of Petitioner', 'Description of Regulatory Need', 'Proposed Rulemaking',
         'Evidence in Support', 'Conclusion'],
        ['Introduction', 'Factual and Legal Basis', 'Proposed Rule Text',
         'Economic Justification', 'Conclusion'],
    ],
    'no-action-request': [
        ['Introduction', 'Factual Background', 'Legal Analysis',
         'Basis for No-Action Relief', 'Conclusion'],
        ['Background', 'Description of the Arrangement', 'Applicable Regulatory Framework',
         'Request for Relief', 'Conclusion'],
        ['Introduction', 'Facts and Circumstances', 'Regulatory Analysis',
         'No-Action Request', 'Conclusion'],
    ],
    'advisory-memorandum': [
        ['Introduction', 'Summary of Recent Agency Action', 'Implications for Employers',
         'Recommended Compliance Steps', 'Conclusion'],
        ['Overview', 'Regulatory Background', 'Key Issues', 'Employer Action Items', 'Conclusion'],
        ['Summary', 'Background', 'Analysis', 'Practical Guidance', 'Conclusion'],
    ],
    'joint-comments': [
        ['Introduction and Statement of Interest', 'Background',
         'Joint Analysis of the Proposed Rule', 'Joint Recommendations', 'Conclusion'],
        ['Introduction', 'Summary of Joint Position', 'Analysis', 'Recommendations', 'Conclusion'],
        ['Statement of the Coalition', 'Overview', 'Key Concerns',
         'Recommended Modifications', 'Conclusion'],
    ],
    'research-memorandum': [
        ['Introduction', 'Research Question', 'Data and Methodology',
         'Findings', 'Policy Implications', 'Conclusion'],
        ['Summary', 'Background', 'Analytical Approach', 'Results',
         'Discussion', 'Recommendations'],
        ['Executive Summary', 'Introduction', 'Empirical Analysis',
         'Regulatory Implications', 'Recommendations', 'Conclusion'],
    ],
    'formal-objection': [
        ['Introduction', 'The Agency Exceeded Its Statutory Authority',
         'The Final Rule Is Arbitrary and Capricious', 'The Agency Failed to Consider Significant Evidence',
         'Relief Requested', 'Conclusion'],
        ['Overview', 'Legal Basis for Objection', 'Factual Grounds',
         'Procedural Deficiencies', 'Requested Remedy', 'Conclusion'],
        ['Introduction', 'Statutory Authority Limitations', 'Failure to Address Key Comments',
         'Economic Analysis Deficiencies', 'Request for Reconsideration', 'Conclusion'],
    ],
}

# Optional extra sections to inject into document structure for variety
_OPTIONAL_SECTION_POOL = [
    'Impact on Small Employers',
    'International Comparisons',
    'Enforcement and Compliance Considerations',
    'Economic and Workforce Impact Analysis',
    'Data and Methodology Concerns',
    'Request for Additional Guidance',
    'Implementation Timeline Concerns',
    'Definitional Issues and Recommended Clarifications',
    'Safe Harbor Framework Proposal',
    'Effect on Collective Bargaining Relationships',
    'Interaction with Existing Regulatory Frameworks',
    'Technology and Automation Considerations',
    'Small Business and Competitive Market Effects',
    'State Law Preemption Considerations',
    'Phase-In and Transition Rule Recommendations',
    'Guidance on Edge Cases and Fact Patterns',
    'Alternative Regulatory Approaches Considered',
    'Stakeholder Engagement Process Review',
    'Coordination with Other Federal Agencies',
    'Equity and Disparate Impact Analysis',
    'Sunset and Review Provisions',
    'Cost-Benefit Analysis Supplementation',
    'Employee and Worker Advocacy Perspectives',
    'Board and Shareholder Perspectives',
    'Industry-Specific Application Notes',
    # Technology & Data
    'Impact of Artificial Intelligence on Compliance',
    'Algorithmic Bias and Fairness Considerations',
    'Data Infrastructure and System Requirements',
    'Cybersecurity Implications of Data Collection',
    'Data Privacy and Employee Trust',
    'Cross-Border Data Transfer Limitations',
    'Data Minimization and Retention Policies',
    'Anonymization and De-identification Techniques',
    # Legal & Risk
    'Litigation Risk and Class Action Vulnerability',
    'Attorney-Client Privilege Considerations',
    'Private Right of Action Analysis',
    'Statutory Damages and Penalty Structures',
    # Economic & Financial
    'Impact on Capital Formation and Investment',
    'Analysis of Second-Order Economic Effects',
    'Impact on Labor Market Liquidity',
    'Cost-Benefit Analysis for Small Entities',
    # Workforce & Governance
    'Impact on Talent Acquisition and Retention',
    'Workforce Training and Reskilling Requirements',
    'Role of the Compensation Committee',
    'Board-Level Oversight and Fiduciary Duties',
    'Implications for Director Independence',
    'Interaction with Proxy Advisor Policies',
]

_EXPERT_TYPES = [
    'compensation attorneys', 'benefits counsel', 'HR executives and total rewards professionals',
    'compensation consultants and advisors', 'proxy advisory firm analysts',
    'benefits administrators and third-party recordkeepers', 'payroll professionals',
    'corporate governance specialists', 'ERISA practitioners', 'employment law attorneys',
    'actuaries and plan administrators', 'in-house employment counsel',
    'independent compensation committee advisors', 'workforce analytics professionals',
    'executive compensation consultants', 'labor and employment litigators',
    'retirement plan advisors', 'equity plan administrators',
    'chief financial officers and corporate controllers',
    'tax advisors and accounting professionals',
    'data scientists and quantitative researchers',
    'chief risk officers and compliance professionals',
    'corporate secretaries and governance officers',
    'investor relations professionals',
    'labor economists and industrial relations scholars',
    'union representatives and collective bargaining experts',
    'global mobility and expatriate tax specialists',
    'HRIS and people analytics leaders',
    'cybersecurity and data privacy counsel',
    'M&A due diligence advisors',
    'restructuring and turnaround consultants',
    'shareholder activists and engagement specialists',
    'ESG and sustainability reporting experts',
]

_INDUSTRY_SECTORS = [
    'financial services', 'healthcare and life sciences', 'technology and software',
    'manufacturing and industrial', 'retail and consumer goods', 'professional services',
    'transportation and logistics', 'energy and utilities', 'higher education',
    'hospitality and food service', 'media and entertainment', 'nonprofit and mission-driven',
    'government contracting', 'agriculture and food production', 'construction and real estate',
    'insurance', 'pharmaceutical', 'biotechnology',
    'asset management', 'private equity', 'venture capital', 'investment banking',
    'wealth management', 'commercial real estate', 'REITs', 'infrastructure', 'renewable energy',
    'oil and gas', 'mining', 'chemicals', 'specialty chemicals', 'food and beverage',
    'travel and tourism', 'entertainment and media', 'publishing',
    'broadcasting', 'telecommunications', 'semiconductors', 'software-as-a-service (SaaS)',
    'cybersecurity', 'cloud computing', 'artificial intelligence', 'data analytics',
    'e-commerce', 'fintech', 'healthtech', 'edtech', 'proptech', 'insurtech',
    'medical devices', 'life sciences', 'clinical research', 'diagnostics',
    'behavioral health', 'home health', 'managed care', 'health systems',
    'academic medical centers', 'physician practice management', 'automotive',
    'aerospace and defense', 'naval systems', 'space technology', 'environmental services',
    'waste management', 'water utilities', 'architecture and engineering',
    'engineering and construction', 'management consulting', 'legal services',
    'accounting services', 'staffing and recruiting', 'executive search',
    'human resources outsourcing', 'business process outsourcing', 'it services',
    'systems integration', 'sports and recreation', 'gaming and gambling', 'luxury goods',
    'apparel and fashion', 'beauty and personal care', 'consumer electronics',
    'home improvement', 'grocery & supermarkets', 'quick service restaurants',
    'full service restaurants', 'specialty retail', 'department stores',
    'direct-to-consumer', 'shipping and freight', 'rail transportation', 'aviation',
    'maritime', 'supply chain and procurement', 'third-party logistics',
    'K-12 education', 'vocational training', 'online education',
    'think tanks and policy research', 'trade associations', 'foundations and endowments',
    'faith-based organizations', 'labor unions', 'cooperatives',
    'federal government', 'state and local government', 'municipal utilities',
    'public safety', 'defense contractors', 'intelligence community support',
    'nuclear energy', 'battery storage', 'electric vehicles', 'autonomous vehicles',
    'drone technology', 'satellite communications', 'quantum computing', 'blockchain',
    'web3 and decentralized finance', 'digital assets', 'payments processing',
    'data center REITs', 'cell tower infrastructure', 'fiber networks',
    'streaming media', 'digital advertising', 'ad technology', 'market research', 'corporate training',
    'executive coaching', 'franchise and licensing',
    'subscription commerce', 'veterinary services', 'animal health', 'aquaculture',
    # More Finance
    'credit unions', 'community banking', 'mortgage lending', 'payment processing',
    'insurance brokerage', 'reinsurance', 'actuarial services', 'claims processing',
    # More Tech
    'robotics', 'agritech', 'cleantech', 'legaltech', 'regtech', 'govtech',
    'enterprise software', 'mobile applications', 'gaming', 'virtual reality', 'augmented reality',
    'internet of things (IoT)', 'edge computing', '5g technology', 'fiber optics',
    # More Healthcare
    'hospital systems', 'specialty clinics', 'long-term care', 'telemedicine',
    'dental services', 'vision care', 'mental health services', 'physical therapy',
    'contract research organizations (CROs)', 'pharmacy benefit management (PBM)',
    # More Manufacturing/Industrial
    'heavy machinery', 'industrial automation', 'robotics manufacturing', '3d printing',
    'semiconductor manufacturing', 'electronics manufacturing', 'automotive manufacturing',
    'aerospace manufacturing', 'defense manufacturing', 'shipbuilding', 'textiles',
    # More Consumer/Retail
    'fast-moving consumer goods (FMCG)', 'cosmetics', 'sporting goods', 'home furnishings',
    'convenience stores', 'online marketplaces', 'subscription boxes', 'luxury retail',
    # More Professional Services
    'environmental consulting', 'it consulting', 'strategy consulting', 'hr consulting',
    'marketing and advertising', 'public relations', 'architectural services', 'engineering services',
    # More Energy/Utilities
    'solar energy', 'wind energy', 'hydroelectric power', 'geothermal energy',
    'nuclear power', 'power transmission', 'natural gas distribution', 'water treatment',
    # More Media/Entertainment
    'film and television production', 'music industry', 'video game development',
    'esports', 'live events', 'theatrical production', 'book publishing', 'news media',
    # More Transportation/Logistics
    'air cargo', 'ocean freight', 'trucking', 'warehousing', 'last-mile delivery',
    'ride-sharing', 'public transit', 'airport operations', 'port operations',
    # Niche & Emerging
    'space exploration', 'commercial spaceflight', 'satellite services', 'drone services',
    'carbon capture', 'sustainable agriculture', 'plant-based foods', 'cellular agriculture',
    'cannabis industry', 'psychedelic therapeutics', 'longevity research', 'bioinformatics',
    'nanotechnology', 'advanced materials', 'synthetic biology', 'gene editing',
]

_TIMEFRAMES = [
    'eighteen months', 'two full fiscal years', 'a minimum of twenty-four months',
    'at least three annual reporting cycles', 'no fewer than eighteen months',
    'a period of not less than two years', 'at least one full plan year',
    'twenty-four to thirty-six months', 'a transition period of no less than eighteen months',
    'multiple plan years', 'at least two full compliance cycles',
    'the next fiscal year',
    'a 36-month transition period',
    'the subsequent 24-month period',
    'a five-year planning horizon',
    'a three-year look-back period',
    'no less than six quarters',
    'the upcoming calendar year',
    'a period of 12 to 18 months',
    'the next two reporting cycles',
    'a full business cycle',
    'the duration of the current collective bargaining agreement',
    'a multi-year implementation window',
]

_COMPARISON_GROUPS = [
    'large public companies', 'privately held employers', 'nonprofit organizations',
    'small and mid-sized employers', 'multistate employers', 'federal government contractors',
    'financial institutions subject to enhanced oversight', 'closely held family businesses',
    'portfolio companies of private equity sponsors', 'not-for-profit health systems',
    'cooperatives and employee-owned enterprises', 'early-stage and venture-backed companies',
    'publicly traded companies subject to SEC disclosure', 'employers in heavily unionized industries',
    'large research universities and academic medical centers',
    'state and local government agencies',
    'SaaS and enterprise software companies',
    'biotechnology and pharmaceutical companies',
    'aerospace and defense prime contractors',
    'automotive and industrial manufacturers',
    'consumer packaged goods companies',
    'hospitality and lodging operators',
    'regulated utilities and energy producers',
    'founder-led technology companies',
    'foreign-owned U.S. subsidiaries',
    'organizations with a majority-remote workforce',
    'recently-public (post-IPO) companies',
    'B-corporations and social enterprises',
]

_FINDINGS_BRIEF = [
    'material compliance challenges', 'significant implementation costs',
    'widespread uncertainty about applicable standards', 'inconsistent enforcement outcomes',
    'substantial variance in industry practice', 'a notable compliance gap',
    'meaningful divergence between regulatory intent and practical outcomes',
    'a pattern of good-faith compliance efforts falling short of regulatory expectations',
    'persistent ambiguity in how rules apply to non-traditional arrangements',
    'a clear need for additional regulatory guidance', 'substantial unmet compliance need',
    'a disconnect between the rule\'s stated objectives and observed employer behavior',
]

PARAGRAPH_TEMPLATES = [
    "ACPWB's proprietary compensation benchmarking database — which includes data from more than "
    "2,400 organizations across 47 industries — provides a unique empirical foundation for assessing "
    "the likely impact of regulatory changes in this area. Our analysis consistently demonstrates that "
    "{topic} has material consequences for employer compensation design, talent retention, and workforce equity.",

    "Employers across the spectrum of ACPWB's client base have raised consistent concerns about the "
    "practical implementation of requirements related to {topic}. These concerns are not theoretical: "
    "they reflect the operational reality of organizations managing complex compensation programs "
    "in a rapidly evolving regulatory environment.",

    "The economic literature on {topic} supports a measured and evidence-based approach to regulatory "
    "intervention. ACPWB's own research, consistent with peer-reviewed findings, suggests that "
    "overly prescriptive rules in this area can produce unintended consequences for the very workers "
    "and stakeholders that the regulatory framework is designed to protect.",

    "From a benchmarking perspective, ACPWB has observed significant variation in how organizations "
    "approach {topic}, reflecting differences in industry sector, organizational scale, ownership structure, "
    "and workforce composition. A one-size-fits-all regulatory approach risks imposing costs and "
    "compliance burdens on organizations for whom the proposed rule was not primarily designed.",

    "The proposed framework for {topic} raises important questions about data collection, reporting "
    "methodology, and enforcement consistency. ACPWB urges the agency to engage in a thorough "
    "cost-benefit analysis and to consult with compensation professionals and employers before "
    "finalizing any rule in this area.",

    "ACPWB supports the underlying goals of the regulatory effort in the area of {topic} and "
    "recognizes the legitimate interest of policymakers in ensuring that compensation practices are "
    "transparent, equitable, and consistent with the public interest. Our comments are offered in "
    "the spirit of helping the agency achieve those goals through a workable and appropriately "
    "calibrated regulatory framework.",

    "International comparisons are instructive. Jurisdictions that have adopted strong standards "
    "in the area of {topic} have generally done so through collaborative processes that involved "
    "meaningful engagement with employers, labor organizations, and independent research institutions. "
    "ACPWB recommends a similar approach in the domestic regulatory context.",

    "ACPWB's survey research on {topic} reveals that a substantial majority of employers — "
    "across industries and organization sizes — are supportive of clear and consistent standards "
    "in this area, provided that those standards are developed with appropriate input from affected "
    "parties and are implemented with adequate transition timelines.",

    "The proposed rule's treatment of {topic} reflects a regulatory approach that, while well-intentioned, "
    "may require modification to account for the diversity of compensation structures and employment "
    "arrangements present in the modern economy. ACPWB offers specific recommendations in this regard "
    "in the sections that follow.",

    "Historical experience with similar regulatory interventions in the area of {topic} is instructive. "
    "Effective rules have typically combined clear standards with flexible implementation mechanisms, "
    "robust enforcement infrastructure, and meaningful safe harbors for good-faith compliance efforts.",

    "ACPWB's analysis of publicly available enforcement data suggests that existing rules in the "
    "area of {topic} have not achieved their intended outcomes at scale. Additional guidance, clearer "
    "definitions, and stronger safe harbor provisions would materially improve compliance rates "
    "without imposing undue burdens on good-faith actors.",

    "The competitive dynamics of the labor market interact with regulation in {topic} in ways that "
    "are not always appreciated by regulators. ACPWB's research shows that overly prescriptive "
    "rules in this area can disadvantage smaller employers who lack the compliance infrastructure "
    "of larger organizations, potentially distorting labor market outcomes.",

    "We note that the agency's economic analysis underlying the proposed rule on {topic} relies on "
    "data that ACPWB believes significantly underestimates compliance costs. Our proprietary survey "
    "of HR and compensation professionals indicates that the administrative burden of the proposed "
    "rule is substantially higher than the agency's estimates suggest.",

    "Definitional precision is paramount in any regulatory framework addressing {topic}. ACPWB has "
    "identified at least three key terms in the proposed rule that lack the clarity necessary for "
    "consistent application across the range of compensation structures and employment arrangements "
    "present in the modern economy.",

    "The agency should not underestimate the significance of the transition challenge presented by "
    "{topic}. ACPWB's experience advising organizations through prior regulatory changes in this "
    "area suggests that eighteen to twenty-four months is the minimum realistic implementation timeline "
    "for organizations of moderate complexity.",

    "ACPWB's research on employer compensation practices in the area of {topic} reveals a significant "
    "divergence between large public companies and smaller private and nonprofit employers. "
    "Regulatory design that ignores this heterogeneity risks being either over-inclusive or "
    "under-inclusive in ways that compromise the rule's effectiveness.",

    "From the perspective of institutional investors who rely on compensation disclosures to evaluate "
    "governance quality, the current regulatory framework for {topic} leaves significant gaps. "
    "ACPWB has engaged with a number of major institutional investors in developing these comments, "
    "and their perspectives are reflected in our analysis.",

    "The enforcement record in the area of {topic} demonstrates that civil penalties alone are an "
    "insufficient deterrent to non-compliance. ACPWB recommends that any final rule include "
    "both enhanced civil penalties and a meaningful private right of action to ensure that "
    "the regulatory objectives are actually achieved.",

    "ACPWB has reviewed the academic literature on {topic} comprehensively. The weight of the "
    "evidence supports the conclusion that effective regulatory intervention in this area requires "
    "a combination of mandatory disclosure, audit requirements, and meaningful consequences for "
    "organizations that fall short of the regulatory standard.",

    "The intersection of {topic} with collective bargaining creates additional complexity that the "
    "proposed rule does not adequately address. ACPWB recommends that the agency provide specific "
    "guidance on how the proposed requirements interact with existing collective bargaining agreements "
    "and the duty to bargain over mandatory subjects of bargaining.",

    "Technology has materially changed the landscape for {topic} in ways that existing regulatory "
    "frameworks were not designed to address. ACPWB urges the agency to give particular attention "
    "to algorithmic compensation-setting, remote work dynamics, and platform employment structures "
    "in developing any final rule in this area.",

    "ACPWB's work with compensation committees at public companies has given us direct insight into "
    "the practical challenges of board-level governance in the area of {topic}. The perspectives "
    "of compensation committee chairs and independent directors are critically important to the "
    "regulatory design process and have been underrepresented in the public record.",

    "The proposed approach to {topic} may interact adversely with existing requirements under "
    "other regulatory frameworks, creating duplicative or inconsistent obligations for affected "
    "employers. ACPWB strongly recommends that the agency undertake a comprehensive harmonization "
    "analysis before issuing a final rule.",

    "ACPWB is concerned that the agency's proposed timeline for implementation of requirements "
    "related to {topic} does not adequately account for the systems changes, training, and legal "
    "review processes that organizations will need to undertake. A realistic transition period "
    "is essential to successful implementation.",

    "The evidence from early-adopter jurisdictions that have implemented requirements related to "
    "{topic} provides valuable lessons for the federal regulatory effort. ACPWB has reviewed "
    "publicly available compliance and enforcement data from these jurisdictions and draws on "
    "those findings in developing our recommendations.",

    "State law developments in the area of {topic} have created a complex patchwork of requirements "
    "that federal regulation has the potential to rationalize. ACPWB urges the agency to develop "
    "a framework that provides a clear federal floor while preserving the ability of states to "
    "innovate in ways that serve their unique labor market conditions.",

    "The small business implications of the proposed rule on {topic} deserve more careful attention. "
    "ACPWB's research indicates that compliance costs as a percentage of payroll are significantly "
    "higher for small employers than for large ones, and that the proposed exemption thresholds "
    "are insufficient to address this disparity.",

    "ACPWB's engagement with legal practitioners who advise employers on {topic} has reinforced our "
    "view that the current regulatory guidance is inadequate. Practitioners report significant "
    "uncertainty about how existing rules apply to novel fact patterns, and this uncertainty "
    "generates unnecessary compliance costs and litigation risk.",

    "The workforce equity dimensions of {topic} are central to ACPWB's analysis. Our research "
    "consistently shows that well-designed regulatory standards in this area can narrow persistent "
    "pay gaps without imposing disproportionate administrative burdens, but that design quality "
    "is critical to achieving this balance.",

    "ACPWB has studied the compliance experience of organizations that have proactively adopted "
    "best practices in the area of {topic} ahead of formal regulation. These early movers report "
    "that the benefits of proactive compliance — reduced legal risk, improved employee relations, "
    "better talent retention — typically outweigh the implementation costs.",

    "The proposed rule on {topic} appears to be based on an assumption that all employers possess a high degree "
    "of sophistication in their compensation data infrastructure. Our research indicates this is not the case. "
    "A significant portion of small and mid-sized employers rely on manual processes and less-structured data, "
    "which would make compliance with the proposed rule's reporting requirements exceptionally burdensome.",

    "We believe the {agency} has underestimated the potential for regulatory arbitrage under the proposed "
    "framework for {topic}. The rule's bright-line tests and narrow definitions may encourage employers "
    "to restructure compensation arrangements to fall just outside the rule's scope, undermining the "
    "intended policy objectives without changing the underlying economic substance of the arrangements.",

    "The interaction between the proposed rule on {topic} and existing tax law, particularly with respect to "
    "deferred compensation and equity awards, is not adequately addressed in the agency's analysis. "
    "ACPWB recommends that the {agency} consult with the Internal Revenue Service to ensure a harmonized "
    "approach that does not create conflicting incentives or compliance obligations for employers.",

    "Our analysis of the proposed rule on {topic} indicates that it may have a disproportionate impact on "
    "certain industries, particularly those with highly variable, performance-based compensation structures. "
    "The {agency} should consider providing industry-specific guidance or safe harbors to account for these "
    "legitimate business model differences.",

    "The proposed rule on {topic} would benefit from a more detailed discussion of its application to "
    "non-traditional work arrangements, including gig economy workers, independent contractors, and employees "
    "of professional employer organizations (PEOs). The current draft leaves significant ambiguity in these "
    "areas, which are a growing segment of the American workforce.",

    "ACPWB's research on employee sentiment suggests that while workers value transparency in {topic}, "
    "they also value clarity and simplicity. A regulatory framework that produces overly complex or "
    "confusing disclosures may not achieve the goal of empowering workers if the information provided "
    "is not readily understandable.",

    "The proposed rule on {topic} does not sufficiently account for the role of third-party administrators "
    "and technology vendors in the compensation and benefits ecosystem. Many employers rely on these "
    "vendors for data management and reporting, and the final rule should clarify the respective "
    "responsibilities of employers and their service providers.",

    "We are concerned that the proposed rule on {topic} may inadvertently stifle innovation in compensation "
    "design. By creating rigid compliance requirements, the rule may discourage employers from experimenting "
    "with novel pay structures that could be beneficial for both workers and the organization.",

    "The {agency}'s proposal on {topic} is a commendable effort to address a complex issue, but it "
    "relies on a static view of the labor market. ACPWB's research indicates that compensation practices "
    "are evolving rapidly, and any final rule must be flexible enough to accommodate future innovations "
    "and changing market dynamics.",

    "The compliance costs associated with the proposed rule on {topic} are not trivial. Our economic "
    "modeling suggests that for a mid-sized employer, first-year compliance costs could amount to "
    "a significant percentage of the total HR budget, diverting resources from other critical "
    "talent management initiatives.",

    "ACPWB believes that a principles-based approach to regulating {topic} would be more effective "
    "and less burdensome than the prescriptive, rules-based approach taken in the current proposal. "
    "A principles-based framework would allow employers to achieve the agency's objectives in a "
    "manner that is tailored to their specific circumstances.",

    "The proposed rule on {topic} would benefit from the inclusion of a quantitative materiality "
    "threshold. As currently drafted, the rule could be interpreted to apply to even the most "
    "minor and inconsequential compensation arrangements, creating unnecessary compliance burdens "
    "without advancing the core policy goals.",

    "The public record on {topic} would be improved by a more thorough analysis of the rule's "
    "potential impact on collective bargaining. The {agency} should explicitly address how the "
    "proposed requirements will interact with the duty to bargain and the terms of existing "
    "collective bargaining agreements.",

    "ACPWB's analysis indicates that the proposed rule on {topic} could create a competitive "
    "disadvantage for U.S.-based multinational corporations. The rule's requirements may conflict "
    "with the legal and cultural norms of other jurisdictions, creating a complex and costly "
    "compliance challenge for global employers.",

    "The proposed rule on {topic} is silent on the issue of remedies for non-compliance. ACPWB "
    "recommends that the final rule include a clear and tiered enforcement framework, with "
    "opportunities for employers to cure deficiencies before the imposition of significant penalties.",

    "We urge the {agency} to consider the role of board oversight in the context of {topic}. "
    "A final rule that empowers and relies on independent compensation committees to ensure "
    "compliance may be more effective than a one-size-fits-all prescriptive mandate.",

    "The data requirements of the proposed rule on {topic} are substantial. Many employers, "
    "particularly those with legacy HR systems, will face significant challenges in collecting, "
    "validating, and reporting the required data. The {agency} should provide a longer implementation "
    "period to allow for necessary system upgrades.",

    "ACPWB's research on corporate governance indicates that shareholder engagement is a powerful "
    "driver of change in the area of {topic}. The {agency} should consider how the proposed rule "
    "can be designed to facilitate, rather than supplant, constructive dialogue between companies "
    "and their investors on these critical issues.",

    "The proposed rule on {topic} does not adequately distinguish between different types of "
    "compensation. A final rule should include separate provisions for base salary, annual incentives, "
    "long-term equity awards, and benefits, as each of these components raises distinct policy "
    "and implementation considerations.",

    "We believe the {agency} should conduct a series of public roundtables with employers, worker "
    "advocates, and technical experts before finalizing the rule on {topic}. The complexity of the "
    "issue warrants a more deliberative and collaborative rulemaking process.",

    # Varied-opener templates
    "Survey data gathered from {n_orgs} organizations across ACPWB's compensation benchmarking "
    "network reveals {finding} in the context of {topic}. These results are consistent with findings "
    "from prior studies and reinforce the view that a thoughtful, evidence-based regulatory approach "
    "is essential to achieving the policy objectives underlying this rulemaking.",

    "A {pct}% majority of respondents to our most recent national compensation survey indicated that "
    "requirements related to {topic} present significant operational challenges. Notably, this concern "
    "was expressed across organization sizes and industries — it is not limited to smaller employers "
    "with constrained compliance resources.",

    "Practitioners advising {compare_group} on {topic} have consistently flagged the same core issues: "
    "definitional ambiguity, inconsistent enforcement posture, and inadequate guidance on how proposed "
    "requirements interact with state law. Addressing these concerns in the final rule would materially "
    "reduce the cost and uncertainty of compliance.",

    "Evidence from jurisdictions that have already implemented requirements analogous to those proposed "
    "here provides a useful preview of the compliance landscape. Early experience in those jurisdictions "
    "suggests that robust employer education campaigns and extended safe harbor periods are critical "
    "to achieving meaningful compliance in the first regulatory cycle.",

    "For organizations in the {industry} sector, requirements related to {topic} present challenges "
    "that are distinct from those faced by employers in other industries. Sector-specific guidance "
    "or safe harbors would allow affected employers to comply in a manner that reflects the practical "
    "realities of their compensation structures and workforce arrangements.",

    "Stakeholders across the {industry} sector have expressed consistent support for clear and "
    "enforceable standards in the area of {topic}, provided that those standards are accompanied "
    "by adequate guidance, reasonable transition timelines, and a good-faith compliance program "
    "that protects employers who make reasonable efforts to comply.",

    "{expert_type} who participated in ACPWB's roundtable discussions consistently identified "
    "{finding} as the most significant obstacle to effective compliance with proposed requirements "
    "for {topic}. Their perspectives, drawn from day-to-day advisory experience with affected "
    "organizations, provide a ground-level view that supplements the empirical data in our submission.",

    "Research consistently demonstrates that well-designed standards in the area of {topic} can "
    "achieve meaningful policy outcomes without imposing disproportionate administrative burdens. "
    "The key design variables — scope, reporting methodology, enforcement posture, and transition "
    "timelines — warrant careful attention in the development of any final rule.",

    "The regulatory history of {topic} at the federal level reflects a pattern of incremental "
    "rulemaking, informal guidance, and enforcement discretion that has produced inconsistent outcomes "
    "across industries and organization types. A comprehensive final rule with clear definitions "
    "and enforcement standards would provide greater certainty and reduce compliance costs.",

    "Three recurring challenges consistently emerge when organizations work through the compliance "
    "implications of {topic}: first, the difficulty of adapting legacy data systems to new reporting "
    "requirements; second, the ambiguity of key definitional terms; and third, the interaction of "
    "federal requirements with a complex and sometimes conflicting body of state law.",

    "Empirical evidence on {topic} suggests that the relationship between regulatory stringency and "
    "policy outcomes is not linear. Beyond a certain threshold of prescription, additional requirements "
    "tend to generate compliance costs without producing proportionate improvements in the "
    "substantive outcomes that regulators are seeking to advance.",

    "Comment letters filed with {agency} in prior related rulemakings reveal a consistent theme: "
    "employers are generally supportive of the underlying policy goals but are concerned about "
    "implementation details that, if not addressed, would transform a workable rule into an "
    "administratively burdensome compliance exercise.",

    "Board-level governance practices in the context of {topic} have evolved substantially over "
    "the past decade, driven in part by investor pressure, proxy advisor guidelines, and voluntary "
    "disclosure commitments. A regulatory framework that builds on these existing governance "
    "structures will likely achieve better outcomes than one that ignores them.",

    "Structural features of the modern labor market — including the growth of distributed work, "
    "platform employment, and project-based staffing — complicate the application of {topic} "
    "requirements in ways that existing regulatory frameworks were not designed to address. "
    "We encourage {agency} to explicitly engage with these structural realities in the final rule.",

    "Competitive dynamics in the {industry} sector create specific pressures on compensation "
    "design that have direct implications for how {topic} requirements would operate in practice. "
    "An analysis that ignores sector-specific context risks producing a regulatory framework "
    "that achieves outcomes different from those intended.",

    "A closer examination of the public record on {topic} reveals that the concerns raised by "
    "employers are not merely abstract objections to regulatory oversight. They reflect genuine "
    "operational challenges that, if not addressed in the final rule, will produce compliance "
    "failures and enforcement actions that ultimately serve no one's interests.",

    "In our experience advising organizations through prior regulatory transitions involving {topic}, "
    "the single greatest predictor of successful compliance is the adequacy of the implementation "
    "timeline. A transition period of {timeframe} is the minimum necessary for organizations of "
    "average complexity to implement the required changes without significant disruption.",

    "Our engagement with {expert_type} who advise clients on {topic} has produced a consistent "
    "finding: the gap between regulatory intent and practical implementation is widest in the first "
    "compliance cycle and narrows over time as enforcement practice develops. Adequate safe harbor "
    "protections during this initial period are therefore essential.",

    "Multiple stakeholders — including labor organizations, employer groups, and independent "
    "researchers — have called for a more collaborative approach to developing standards in the "
    "area of {topic}. We share this view and urge {agency} to convene additional public "
    "forums before issuing a final rule.",

    "Organizations operating across multiple state jurisdictions face a particular compliance "
    "challenge when federal and state requirements related to {topic} diverge. A federal rule "
    "that establishes a clear floor while providing certainty regarding federal preemption would "
    "reduce compliance costs and improve predictability for multistate employers.",

    "The interaction between requirements for {topic} and collectively bargained compensation "
    "structures raises questions that the proposed rule does not adequately address. Organizations "
    "with significant unionized workforces need specific guidance on how the proposed requirements "
    "interact with their existing bargaining obligations and agreement terms.",

    "Recent developments in algorithmic compensation-setting, AI-assisted pay decisions, and "
    "workforce analytics platforms have materially altered the operational landscape for {topic}. "
    "Standards developed without reference to these technological realities will rapidly become "
    "obsolete and will create compliance uncertainties for employers adopting current best practices.",

    "Enforcement trends in the area of {topic} over the past {n_years} years reveal a pattern "
    "of increasing agency activity, rising penalty levels, and growing private litigation risk. "
    "This trajectory makes clear that the stakes associated with compliance in this area are "
    "substantial and that employers need unambiguous regulatory guidance.",

    "At its core, effective regulation of {topic} requires a framework that is clear enough to "
    "enable consistent compliance, flexible enough to accommodate legitimate business variation, "
    "and enforced in a manner that targets bad actors rather than penalizing good-faith efforts. "
    "The current proposal does not fully achieve all three of these objectives.",

    "Independent analysis of the regulatory burden associated with {topic} requirements consistently "
    "indicates that first-year compliance costs are underestimated in agency economic analyses. "
    "This pattern reflects the difficulty of anticipating the full range of systems, process, "
    "and legal review costs that organizations must incur to come into initial compliance.",

    "Workers' perspectives on {topic} deserve more weight in the regulatory analysis than they "
    "have typically received. Where workers have been asked directly, they consistently express "
    "a preference for standards that are transparent and consistently enforced — not merely "
    "for more regulation in the abstract.",

    "International best practices in the area of {topic} offer instructive models. Jurisdictions "
    "that have achieved strong compliance outcomes have generally relied on a combination of clear "
    "disclosure standards, robust agency guidance, and collaborative enforcement that rewards "
    "good-faith efforts and focuses sanctions on deliberate non-compliance.",

    "From the perspective of {compare_group}, the proposed approach to {topic} raises concerns "
    "about both cost and feasibility. Survey data indicates that {pct}% of organizations in "
    "this category lack the internal infrastructure to comply within the proposed timeline "
    "without significant external assistance.",

    "Prior rulemakings in closely related areas provide a useful benchmark for evaluating the "
    "proposed approach to {topic}. Where those prior rules achieved strong compliance rates, the "
    "common factors were clear definitions, advance guidance, and extended implementation periods — "
    "all of which we recommend for any final rule in this area.",

    "The private sector compliance community — including {expert_type} who advise affected "
    "organizations — has developed a substantial body of practice in the area of {topic} that "
    "the regulatory framework should seek to build on rather than displace. Regulatory approaches "
    "that work with established compliance practices tend to produce better outcomes at lower cost.",

    "Economic modeling of the proposed regulatory requirements for {topic} must account for both "
    "direct compliance costs and behavioral responses that may alter the compensation practices "
    "being regulated. An analysis that focuses solely on direct costs will systematically "
    "underestimate the full economic impact of the proposed rule.",

    "Many of the compliance challenges associated with {topic} are concentrated in the transition "
    "period immediately following the rule's effective date. Front-loading of compliance resources, "
    "system upgrades, and training programs during this period creates significant disruption "
    "that is rarely fully captured in agency cost-benefit analyses.",

    "The distinction between large and small employers is not the only dimension of heterogeneity "
    "relevant to requirements for {topic}. Industry sector, ownership structure, workforce "
    "composition, and geographic footprint all influence how organizations experience the regulatory "
    "burden — and all of these dimensions should inform the agency's regulatory design choices.",

    "Benchmarking data drawn from {n_orgs} participating organizations reveals {finding} when "
    "organizations are asked to characterize their current state of readiness to comply with "
    "requirements similar to those proposed for {topic}. This readiness gap should inform "
    "the agency's decisions about transition timelines and enforcement prioritization.",

    "Regulatory design choices that may appear minor — the scope of a definition, the threshold "
    "for a reporting obligation, the treatment of a specific compensation element — can have "
    "outsized practical effects in the context of {topic}. Our comments reflect close attention "
    "to these details, which determine whether the rule will function as intended.",

    "Voluntary compliance programs and best-practice guidelines developed by industry associations "
    "have demonstrated that meaningful progress on {topic} is achievable without mandatory regulation. "
    "Rather than supplanting these efforts, a well-designed regulatory framework should build "
    "on and reinforce the momentum that voluntary programs have generated.",

    "The {n_years}-year implementation record of analogous requirements in other regulatory "
    "contexts provides a rich empirical foundation for evaluating the proposed approach to {topic}. "
    "Lessons learned from those prior implementations — particularly with respect to definitional "
    "gaps and enforcement ambiguities — should be incorporated into the final rule.",

    "Smaller employers — those with fewer than {n_orgs} employees — face a structurally different "
    "compliance environment than the large organizations whose compensation practices typically "
    "dominate the public record in rulemakings of this type. Requirements calibrated to large-company "
    "infrastructure will impose disproportionate burdens on smaller organizations.",

    "Cross-jurisdictional comparisons in the area of {topic} consistently reveal that the United "
    "States regulatory framework, while broadly sound in its objectives, lags behind peer jurisdictions "
    "in the clarity and specificity of its implementation guidance. Closing this gap should be a "
    "priority for any final rulemaking in this area.",

    "Shareholder advisory firms and institutional investors have increasingly incorporated {topic} "
    "considerations into their proxy voting and engagement frameworks. A regulatory floor that "
    "formalizes and clarifies minimum standards in this area would complement — rather than "
    "substitute for — the market pressures that are already driving compensation practice.",
]

RECOMMENDATION_TEMPLATES = [
    "The {agency} should extend the comment period by no fewer than sixty days to allow for adequate "
    "stakeholder engagement and empirical analysis.",

    "The agency should conduct and publish a rigorous cost-benefit analysis of the proposed rule's "
    "impact on employers of varying sizes and industries before proceeding to finalization.",

    "ACPWB recommends that the final rule include a phase-in period of not less than eighteen months "
    "to allow organizations adequate time for system implementation, training, and compliance preparation.",

    "The {agency} should establish a safe harbor for good-faith compliance efforts and provide "
    "substantive informal guidance before initiating enforcement actions in this area.",

    "We recommend that the agency convene an advisory committee of compensation professionals, "
    "legal practitioners, and employer representatives to provide ongoing technical input on "
    "implementation and enforcement matters.",

    "The final rule should include a de minimis exception for small employers and should define "
    "the applicability thresholds with sufficient clarity to permit compliance planning.",

    "ACPWB urges the {agency} to align its requirements in this area with existing reporting "
    "obligations to minimize duplicative compliance burdens on affected employers.",

    "The agency should clarify the definition of key terms in the proposed rule and provide "
    "illustrative examples of compliant and non-compliant practices to support consistent implementation.",

    "We recommend that the {agency} establish a formal feedback mechanism through which employers "
    "can report implementation challenges and obtain informal guidance without triggering enforcement.",

    "ACPWB requests that the agency publish interim final guidance on at least three priority issues "
    "identified in public comments before the effective date of any final rule.",

    "The final rule should expressly preserve the ability of employers to design compensation programs "
    "that reflect industry-specific norms and competitive market conditions, subject to the baseline "
    "protections established by the regulatory framework.",

    "ACPWB recommends that the {agency} coordinate with relevant sister agencies to ensure that "
    "reporting requirements are harmonized and that employers are not subject to inconsistent "
    "or contradictory regulatory mandates on the same underlying compensation practices.",

    "The agency should publish a clear enforcement prioritization policy that enables employers to "
    "focus compliance resources on the most significant risk areas and that limits enforcement "
    "discretion in ways that promote predictability.",

    "ACPWB recommends that the {agency} establish a standing technical advisory committee with "
    "representation from industry, labor, and academia to provide ongoing input on the operation "
    "and effectiveness of the final rule.",

    "The final rule should include an explicit sunset and review provision, requiring the agency "
    "to assess the rule's effectiveness and economic impact no later than five years after the "
    "effective date, and to publish the results of that assessment.",

    "The {agency} should develop model compliance programs and template disclosures that smaller "
    "employers can use as a starting point, reducing the cost of compliance for organizations "
    "without dedicated HR and legal teams.",

    "We recommend that the agency conduct robust outreach to small and mid-size employers during "
    "the transition period, providing workshops, webinars, and written guidance specifically "
    "tailored to the compliance challenges of these organizations.",

    "The final rule should establish clear and objective criteria for the agency's exercise of "
    "enforcement discretion, with published guidance on how the agency will prioritize cases "
    "and assess penalties.",

    "ACPWB recommends that the agency adopt a pilot program approach, implementing the proposed "
    "requirements on a voluntary basis for a two-year period before mandating compliance, to "
    "generate real-world implementation data that can inform the final rule design.",

    "The {agency} should establish an expedited process for issuing informal guidance on novel "
    "fact patterns arising during the transition period, with a commitment to respond to written "
    "requests within sixty days.",

    "We urge the agency to provide clear, specific guidance on how the proposed rule interacts "
    "with state law requirements, and to identify areas where federal preemption applies and "
    "where state law remains operative.",

    "ACPWB recommends that the final rule include an express provision requiring the agency to "
    "update its guidance materials at least annually to reflect enforcement experience, court "
    "decisions, and changes in compensation practice.",

    "The agency should develop a robust enforcement tracking system that makes aggregate data "
    "on enforcement actions, penalties, and compliance rates publicly available on at least an "
    "annual basis.",

    "We recommend that the {agency} establish a dedicated ombudsman function to assist employers "
    "with compliance questions and to serve as a channel for reporting implementation difficulties "
    "that do not rise to the level of formal complaints.",

    "The final rule should adopt the more employer-friendly of competing interpretations wherever "
    "the statutory text is ambiguous, consistent with the principles of fair notice and the "
    "constitutional due process requirements applicable to regulatory enforcement.",

    "ACPWB recommends that the agency provide a structured process for organizations to obtain "
    "binding advance determinations on proposed compensation arrangements, to reduce legal "
    "uncertainty and promote compliance.",

    "The {agency} should require that any third-party auditor or compliance assessor used in "
    "connection with the final rule meet minimum qualification standards, and should publish "
    "those standards with sufficient clarity to enable employers to vet potential service providers.",

    "We urge the agency to consider the cumulative compliance burden imposed by this rulemaking "
    "in combination with other recent regulatory actions, and to establish a mechanism for "
    "coordinating the timing and scope of compliance requirements to avoid regulatory overload.",

    "The final rule should preserve maximum flexibility for parties to address compliance issues "
    "through negotiated resolution, and should not require formal adjudication of matters that "
    "can be resolved more efficiently through alternative dispute resolution.",

    "ACPWB recommends that the agency require covered employers to designate a responsible "
    "compliance officer with defined duties and accountability for the requirements of the "
    "final rule, consistent with best practices in compliance program design.",

    "The {agency} should publish detailed economic analysis of the distributional effects "
    "of the proposed rule, including its impact on workers at different income levels and "
    "in different demographic groups, before finalizing the rule.",

    "We recommend that the agency establish a working group comprising large employers, small "
    "employers, worker advocates, and independent experts to develop consensus guidance on "
    "the most difficult implementation questions raised by the final rule.",

    "The final rule should include a transition rule providing that actions taken in reliance "
    "on existing guidance or industry practice prior to the rule's effective date will not "
    "be subject to retroactive enforcement.",

    "ACPWB urges the {agency} to revisit the proposed recordkeeping requirements and to limit "
    "them to information that is directly relevant to the rule's enforcement objectives, "
    "eliminating provisions that impose cost without commensurate benefit.",

    "The agency should establish a clear remediation pathway for organizations that self-identify "
    "compliance deficiencies, providing reduced penalties and no-action protections for good-faith "
    "remediation efforts that meet objective criteria.",

    "We recommend that the {agency} publish a detailed compliance calendar aligned with the rule's "
    "phase-in timeline, identifying specific milestones and deliverables to help employers "
    "structure their compliance programs effectively.",

    "The {agency} should specify data security and privacy standards for any information collected under this rule, consistent with the NIST Cybersecurity Framework, to protect sensitive employee data.",

    "ACPWB recommends that the final rule permit the use of statistical sampling for data collection where appropriate, to reduce the compliance burden on employers with large and complex workforces.",

    "The agency should provide a standardized data submission format and a public testing environment to allow employers to validate their reporting files prior to the compliance deadline.",

    "We urge the {agency} to consider the role of third-party HR technology providers and clarify their responsibilities versus those of the employer under the proposed rule on {topic}.",

    "The final rule should adopt a tiered implementation timeline, with large employers (over 500 employees) complying within 24 months and smaller employers complying within 36 months.",

    "ACPWB recommends a 'compliance grace period' for the first 12 months following the effective date, during which the {agency} would focus on education and technical assistance rather than punitive enforcement.",

    "The agency should commit to publishing a comprehensive set of Frequently Asked Questions (FAQs) and implementation guides at least six months prior to the first compliance deadline.",

    "We recommend the {agency} narrow the definition of '{topic}' to focus on arrangements that pose a systemic risk, thereby exempting routine and non-problematic compensation practices.",

    "The final rule should explicitly exempt employers with fewer than 100 employees from the more burdensome reporting requirements to mitigate the disproportionate impact on small businesses.",

    "ACPWB urges the agency to clarify the application of the proposed rule to non-U.S. employees of multinational corporations to avoid extraterritorial overreach and conflicts with foreign law.",

    "The {agency} should conduct a more thorough analysis of the rule's potential impact on labor market competition and wage inflation before finalization.",

    "We recommend the agency include provisions to ensure the rule does not inadvertently stifle innovation in compensation and benefits design by being overly prescriptive.",

    "The final rule should be reviewed by the Small Business Administration's Office of Advocacy to ensure its impact on small entities is fully understood and mitigated.",

    "The agency should clarify the statute of limitations for violations under the proposed rule and, where possible, align it with existing federal employment laws for consistency.",

    "ACPWB recommends that the final rule include a provision for confidential, binding arbitration to resolve disputes over technical compliance matters, promoting efficient resolution.",

    "The final rule should specify that penalties will be assessed based on the severity and willfulness of the violation, with clear distinctions between substantive failures and good-faith administrative errors.",

    "The {agency} should commit to a post-implementation review of the rule's effectiveness and economic impact within three years of the final compliance deadline.",

    "We recommend the agency provide a clear and simple process for employers to request advisory opinions on the application of the rule to novel compensation arrangements.",

    "The final rule should include a provision allowing for the correction of inadvertent errors within a 90-day cure period without penalty.",

    "ACPWB urges the {agency} to publish aggregated, anonymized data collected under this rule to provide valuable public benchmarks on {topic}.",

    "The agency should provide specific guidance on the interaction between the proposed rule and existing collective bargaining agreements, particularly concerning the duty to bargain over changes to compensation.",

    "We recommend that the final rule's recordkeeping requirements be limited to a three-year look-back period to align with standard business practices and reduce long-term data storage burdens.",

    "The {agency} should create a dedicated small business compliance assistance office to provide direct support and technical guidance to smaller employers navigating the rule on {topic}.",

    "The final rule should clarify that compliance with substantially similar state-level laws on {topic} will be deemed sufficient for federal compliance to avoid duplicative burdens.",

    "ACPWB recommends that the agency conduct a series of industry-specific roundtables to gather targeted feedback on the unique challenges the proposed rule presents for different sectors.",

    "The agency should provide a clear definition of 'control group' for the purpose of determining employer size and applicability, especially in complex corporate structures.",

    "We urge the {agency} to incorporate a materiality qualifier for disclosure requirements, focusing on compensation elements that have a significant impact on employee incentives and corporate risk.",
]

# ── Position vocabulary ───────────────────────────────────────────────────────

POSITIONS = [
    ('supports', 'ACPWB supports the overall objectives of this regulatory action and urges the agency '
     'to adopt the final rule with the modifications recommended in this filing.'),
    ('supports', 'ACPWB commends the agency for its work on this important issue and supports the proposed rule. Our filing provides technical suggestions to enhance the rule\'s clarity and effectiveness.'),
    ('supports', 'ACPWB strongly endorses the proposed rule and urges the agency to finalize it '
     'expeditiously. This filing offers technical comments intended to improve implementation clarity '
     'without altering the rule\'s core approach.'),
    ('supports', 'The proposed rule represents a critical advancement in compensation policy. ACPWB strongly supports its adoption and offers the following analysis to reinforce the agency\'s rationale.'),
    ('supports', 'ACPWB views the proposed rulemaking as a meaningful and necessary step toward '
     'addressing long-standing gaps in compensation regulation. We offer our support and identify '
     'specific modifications that would strengthen the rule\'s effectiveness.'),
    ('supports', 'ACPWB believes the proposed rule is a well-calibrated and necessary response to documented market failures. We support its prompt finalization and urge the agency to resist efforts to weaken its core provisions.'),
    ('supports', 'The evidence presented in the agency\'s proposal is compelling. ACPWB supports the rule and believes its benefits will substantially outweigh its costs. Our comments focus on minor technical refinements.'),
    ('supports', 'ACPWB is pleased to offer its strong support for this proposed rulemaking, which aligns with our long-standing research on the need for greater transparency and accountability in this area.'),
    ('supports', 'This is a timely and well-considered proposal. ACPWB supports the rule and encourages the agency to maintain the rigor of the current draft in the final version.'),
    ('supports', 'ACPWB supports the proposed rule and believes it will create a more level playing field for employers and workers alike. We offer our expertise to assist the agency in finalizing a robust and defensible rule.'),

    ('opposes', 'ACPWB opposes the proposed rule as currently drafted and urges the agency to withdraw '
     'or substantially revise it in light of the concerns identified in this filing.'),
    ('opposes', 'ACPWB believes the proposed rule is fundamentally flawed and will create significant, unnecessary burdens on employers. We urge the agency to withdraw the proposal and pursue a more targeted, evidence-based approach.'),
    ('opposes', 'ACPWB respectfully but firmly opposes the agency\'s approach in this rulemaking. '
     'The proposed rule will produce significant unintended consequences and should be substantially '
     'revised before finalization.'),
    ('opposes', 'The proposed rule is based on a misunderstanding of current market practices and will lead to perverse outcomes. ACPWB cannot support this proposal and recommends a complete re-evaluation of the underlying issues.'),
    ('opposes', 'ACPWB urges the agency to withdraw this proposal and undertake a more thorough '
     'economic analysis before re-proposing. The current record does not support the regulatory '
     'action contemplated, and finalization would be premature.'),
    ('opposes', 'While we share the agency\'s stated goals, the proposed rule is an unworkable and overly broad solution. ACPWB strongly opposes its adoption in its current form.'),
    ('opposes', 'The agency\'s economic analysis is deficient, its legal authority is questionable, and the operational burdens are unacceptably high. For these reasons, ACPWB opposes the proposed rule and urges its withdrawal.'),
    ('opposes', 'This proposal represents a significant overreach of the agency\'s statutory authority and imposes an unworkable mandate on employers. ACPWB urges the agency to abandon this approach.'),
    ('opposes', 'The proposed rule is a solution in search of a problem. The agency has not provided sufficient evidence of a market failure to justify this level of intervention. ACPWB opposes the rule.'),
    ('opposes', 'ACPWB believes the proposed rule will stifle innovation and harm competition without achieving its stated objectives. We respectfully request that the agency withdraw the proposal and engage in a new round of stakeholder outreach.'),

    ('supports-modifications', 'ACPWB supports the underlying goals of this regulatory action but urges '
     'the agency to adopt significant modifications before finalizing the rule.'),
    ('supports-modifications', 'ACPWB is in general agreement with the direction of the proposed rule but has identified several critical flaws that must be addressed. We offer our conditional support, pending the adoption of our recommended changes.'),
    ('supports-modifications', 'ACPWB offers conditional support for the proposed approach, contingent '
     'on the agency\'s adoption of the modifications and clarifications recommended in this filing.'),
    ('supports-modifications', 'The proposed rule is a step in the right direction, but it is not yet ready for finalization. ACPWB supports the rule\'s intent but strongly recommends the modifications detailed in this submission to ensure it is workable in practice.'),
    ('supports-modifications', 'While ACPWB is broadly supportive of the policy direction reflected '
     'in this proposed rule, we have identified several provisions that require significant revision '
     'to achieve the intended regulatory outcomes without imposing disproportionate compliance burdens.'),
    ('supports-modifications', 'ACPWB can support the final rule only if the agency incorporates significant changes to address the operational challenges and unintended consequences identified in our analysis.'),
    ('supports-modifications', 'The proposal has merit, but its implementation details are deeply problematic. ACPWB supports the rule in principle but cannot endorse it in its current form without substantial amendment.'),
    ('supports-modifications', 'We believe a workable final rule is achievable, but not without the significant modifications outlined in our comments. ACPWB offers its expertise to help the agency refine this proposal.'),
    ('supports-modifications', 'The agency has correctly identified the problem, but the proposed solution is flawed. ACPWB supports the agency\'s goal but urges a revised approach that incorporates the practical feedback provided in this submission.'),
    ('supports-modifications', 'ACPWB supports the spirit of the proposed rule but has serious reservations about its prescriptive nature. We recommend a more principles-based approach and offer specific language to that effect.'),
]


# ── Footnote templates ────────────────────────────────────────────────────────

_MONTHS_LONG = ['January', 'February', 'March', 'April', 'May', 'June',
                'July', 'August', 'September', 'October', 'November', 'December']

FOOTNOTE_TEMPLATES = [
    "See {agency} Notice of Proposed Rulemaking, Docket No. {docket}, {month} {year}.",
    "ACPWB Compensation Benchmarking Survey, {year} Edition (N={n} organizations across 47 industry sectors). "
    "Methodology available at acpwb.com/research.",
    "Bureau of Labor Statistics, Occupational Employment and Wage Statistics (OEWS), {year} Annual Release.",
    "See {act}. ACPWB's analysis of this authority is set forth in Section II of this filing.",
    "{agency} Annual Report, Fiscal Year {year}, at p. {page}.",
    "Harvard Kennedy School Program on Corporate Governance, Discussion Paper No. {paper_num} ({year}). "
    "Cited for empirical context on board-level compensation governance.",
    "ACPWB Internal Policy Brief No. {brief_num}-{year_short} (on file with ACPWB Policy Division).",
    "U.S. Government Accountability Office, Report No. GAO-{year_short}-{seq:03d} ({month} {year}).",
    "See generally {act} and implementing regulations at {cfr_title} C.F.R. Part {cfr_part}.",
    "Congressional Budget Office, Cost Estimate for the {topic_short} rule ({year}). "
    "Estimated 10-year budgetary impact: ${b} billion.",
    "OECD, \"Compensation Governance: Trends and Regulatory Developments\" ({year}). "
    "Paris: OECD Publishing.",
    "ACPWB Employer Survey, {year}: {pct}% of respondents reported compliance uncertainty "
    "in this regulatory area.",
    "National Bureau of Economic Research Working Paper No. {paper_num} ({year}). "
    "Cited for empirical findings on labor market effects; ACPWB does not endorse all conclusions.",
    "{agency} Enforcement Statistics, Fiscal Year {year}. "
    "Figures reflect administrative complaints filed, not adjudicated outcomes.",
    "World Economic Forum, \"Future of Jobs Report {year}.\" "
    "Cited for international labor market benchmarking context.",
]

# ── Document stub (lightweight — title + metadata only) ───────────────────────

_STUB_TITLE_PREFIXES = {
    'comment-letter': ['Comment Letter on', 'Comments of ACPWB Regarding', 'Written Comments on'],
    'position-statement': ['ACPWB Position Statement:', 'Statement of Position:'],
    'policy-brief': ['Policy Brief:', 'ACPWB Policy Brief:'],
    'legislative-testimony': ['Testimony of ACPWB on', 'Statement for the Record:'],
    'amicus-brief': ['Brief of ACPWB as Amicus Curiae:', 'Amicus Curiae Brief on'],
    'white-paper': ['White Paper:', 'ACPWB White Paper:'],
    'supplemental-comments': ['Supplemental Comments of ACPWB on', 'ACPWB Supplemental Submission on'],
    'reply-comments': ['Reply Comments of ACPWB on', 'ACPWB Reply Comments:'],
    'ex-parte-submission': ['Ex Parte Notice:', 'ACPWB Ex Parte Submission on'],
    'regulatory-petition': ['Petition for Rulemaking:', 'ACPWB Rulemaking Petition on'],
    'no-action-request': ['No-Action Request:', 'Request for No-Action Relief on'],
    'advisory-memorandum': ['Advisory Memorandum:', 'ACPWB Advisory Memorandum on'],
    'joint-comments': ['Joint Comments of ACPWB on', 'Joint Submission Regarding'],
    'research-memorandum': ['Research Memorandum:', 'ACPWB Research Note:'],
    'formal-objection': ['Formal Objection to', 'ACPWB Formal Objection:'],
}

# ── Featured filings for index page ──────────────────────────────────────────

_FEATURED_SEEDS = [
    (2024,  3, 15, 'sec',            'executive-compensation-disclosure-requirements'),
    (2023, 11,  2, 'dol',            'overtime-threshold-adjustment'),
    (2024,  1, 19, 'ftc',            'noncompete-agreement-enforcement-guidelines'),
    (2022,  8,  4, 'eeoc',           'pay-equity-reporting-standards'),
    (2023,  5, 22, 'nlrb',           'collective-bargaining-unit-determination'),
    (2024,  2,  8, 'cfpb',           'incentive-compensation-recovery-rule'),
    (2021, 10, 14, 'senate-help',    'testimony-executive-pay-reform'),
    (2023,  7, 31, 'irs',            'deferred-compensation-reporting-requirements'),
    (2022,  4, 11, 'osha',           'healthcare-worker-hazard-pay-standards'),
    (2024,  6,  3, 'doj',            'no-poach-agreement-enforcement-standards'),
    (2021,  9, 27, 'hhs',            'healthcare-worker-hazard-pay-standards'),
    (2023,  3,  6, 'ny-dol',         'pay-transparency-requirements'),
    (2022, 12, 13, 'ofccp',          'federal-contractor-pay-transparency'),
    (2024,  4, 22, 'finra',          'broker-dealer-compensation-governance'),
    (2020,  6,  8, 'house-edlabor',  'testimony-worker-classification-reform'),
    (2023,  8, 17, 'treasury',       'executive-compensation-tax-deductibility-limits'),
    (2021,  2, 25, 'nlrb',           'right-to-organize-protections'),
    (2022,  9, 29, 'sec',            'pay-versus-performance-disclosure'),
    (2024,  1,  5, 'dol',            'independent-contractor-classification-rule'),
    (2023,  6, 20, 'eeoc',           'gender-pay-gap-disclosure-rule'),
    (2022,  3, 14, 'ftc',            'amicus-brief-noncompete-enforceability'),
    (2021,  7,  8, 'ebsa',           'retirement-plan-fiduciary-standards'),
    (2023, 10, 31, 'occ',            'banker-bonus-deferral-requirements'),
    (2024,  5, 16, 'whd',            'tipped-worker-minimum-wage-standards'),
]

# ── CEO era registry ──────────────────────────────────────────────────────────

_CEO_NAMES = [
    (1993, 2001, "Richard A. Harmon", "Chief Executive Officer"),
    (2002, 2010, "Margaret S. Ellison", "President & Chief Executive Officer"),
    (2011, 2018, "Thomas J. Whitfield", "Chief Executive Officer"),
    (2019, 2025, "Catherine E. Voss", "President & Chief Executive Officer"),
]

# ── Year era themes ───────────────────────────────────────────────────────────

_YEAR_ERA_THEMES = {
    'early': [
        "executive compensation disclosure", "pay equity and comparable worth standards",
        "FMLA compliance and leave administration", "stock option accounting transparency",
        "proxy statement reform", "employment classification standards",
    ],
    'post_sox': [
        "Sarbanes-Oxley compliance frameworks", "executive pay recovery policies",
        "say-on-pay shareholder advisory votes", "board compensation committee independence",
        "deferred compensation tax treatment", "severance and golden parachute reform",
    ],
    'dodd_frank': [
        "Dodd-Frank pay ratio disclosure implementation", "CEO-to-median worker pay reporting",
        "clawback policy development and enforcement", "incentive compensation risk alignment",
        "pay-versus-performance disclosure", "hedging and pledging prohibitions",
    ],
    'recent': [
        "pay transparency legislation", "ESG-linked executive compensation disclosures",
        "remote work compensation benchmarking", "AI-assisted pay equity analysis",
        "pay data reporting to federal agencies", "noncompete agreement reform",
    ],
}

# ── CEO annual message templates ──────────────────────────────────────────────

_CEO_MESSAGE_TEMPLATES = [
    (
        "The {year} policy year marked a period of significant regulatory activity for "
        "compensation professionals across every sector. ACPWB submitted {total} formal "
        "regulatory filings, spanning comment letters, position statements, and legislative "
        "testimony before federal and state bodies. Our focus this year centered on {theme}, "
        "where we believe evidence-based employer guidance is most urgently needed."
    ),
    (
        "In {year}, ACPWB continued its three-decade commitment to independent, nonpartisan "
        "engagement with the agencies and legislative committees that shape American compensation "
        "policy. This annual summary reflects {total} individual submissions and positions taken "
        "on behalf of the employer community. The defining policy question of {year} was, in "
        "our assessment, {theme} — an area where ACPWB's proprietary data gave us a meaningful "
        "empirical voice in an otherwise contentious debate."
    ),
    (
        "Reflecting on {year}, I am proud of the depth and rigor our policy team brought to "
        "{theme} and related regulatory matters. Our {total} submissions this year were informed "
        "by survey data representing thousands of employers and millions of covered workers. "
        "We engaged constructively with agency staff, testified before two congressional "
        "subcommittees, and filed detailed economic analyses that shaped the final regulatory "
        "record in several important proceedings."
    ),
    (
        "The regulatory calendar in {year} was unusually full, and ACPWB rose to meet it. "
        "Across {total} filings, we addressed {theme} and adjacent issues that directly affect "
        "how American employers attract, retain, and compensate their workforces. Our comments "
        "drew on our flagship compensation benchmarking survey, independent economic research, "
        "and direct input from our employer advisory network. We remain committed to translating "
        "practitioner knowledge into credible policy recommendations."
    ),
    (
        "ACPWB's {year} policy engagement was shaped by a rapidly evolving regulatory "
        "environment. The {total} submissions catalogued here represent our team's work across "
        "multiple agencies, legislative venues, and industry forums. {theme} emerged as the "
        "dominant policy concern of the year, and we committed substantial analytical resources "
        "to ensure that the employer perspective — grounded in data, not ideology — was well "
        "represented in every proceeding we entered."
    ),
]

# ── Year-specific annual CEO letters (4–6 paragraphs each) ───────────────────

_YEAR_ANNUAL_LETTERS = {
    1993: [
        "The arrival of the Clinton administration in January set an immediate and decisive tone for compensation regulation. Congress moved on two fronts that defined our policy calendar: the Family and Medical Leave Act was signed in February, and Section 162(m) — the $1 million deductibility cap on executive pay — was enacted through the Omnibus Budget Reconciliation Act in August. Together these two actions generated more formal regulatory comment opportunity than any single year in ACPWB's operational history to that point, and our team submitted {total} filings across Treasury, the IRS, and the Department of Labor.",
        "Section 162(m) commanded the largest share of our analytical resources. The statute's performance-based compensation exemption raised immediate questions about compensation committee independence standards, the treatment of existing equity programs, and transitional relief for arrangements already in place. Our written comments to the IRS on the proposed regulations were cited in the final rule preamble — a meaningful validation of the technical depth our research team had developed.",
        "FMLA implementation presented a parallel compliance emergency. The statute required operational readiness in fewer than sixty days from enactment, and the DOL's initial guidance left substantive administration questions unresolved around intermittent leave, serious health condition definitions, and equivalent reinstatement rights. ACPWB submitted technical comments, published employer guidance, and participated in three DOL public meetings on implementation challenges.",
        "Our focus on {theme} reflected the broader administration interest in making pay practices more transparent and more directly tied to outcomes. Properly administered, performance linkage requirements create the kind of discipline that benefits well-run organizations. Our role is to ensure that regulatory implementation reflects operational realities, not idealized assumptions about how compensation programs actually work. The advisory relationships we built with IRS and DOL staff during {year} would prove durable through the decade that followed.",
    ],
    1994: [
        "The midterm elections of November altered the federal legislative and regulatory landscape more abruptly than any comparable event since 1980. The Republican capture of both chambers produced an immediate posture shift in the agencies: rulemaking timelines lengthened, proposals that had been advancing steadily were withdrawn or substantially rewritten, and the appetite for new employer mandates — already diminished by the collapse of the Clinton health care proposal — effectively disappeared. ACPWB's {total} filings reflected this transitional environment, with a larger share devoted to technical comment on existing regulations than to engagement with new proposals.",
        "The implementation of NAFTA on January 1 raised compensation questions that our member employers had not fully anticipated. Cross-border staffing arrangements, compensation benchmarking in a trinational labor market, and the treatment of expatriate assignments under the new trade framework generated demand for policy guidance that our research team worked to address throughout the year.",
        "The failure of the Clinton health care proposal — which had consumed enormous employer compliance preparation resources through mid-year before its collapse in September — left the benefits and compensation community in an unusual position, anticipating regulatory requirements that would never arrive. Our guidance efforts shifted toward the underlying employment cost questions that health reform had surfaced, including actuarial modeling of self-insured plan costs and the compensation implications of workforce restructuring driven by rising benefit expenses.",
        "Against this backdrop of regulatory retreat and legislative transition, {theme} remained a technically active area. The IRS had finalized the Section 162(m) regulations in October of the prior year, and the first full fiscal years subject to those rules were now generating real-world compliance questions. We convened a working group on deductibility planning for equity-based compensation and provided technical guidance to the many compensation committees navigating their first performance-based pay certifications.",
        "The year reinforced for us the institutional value of engagement across the political cycle. Organizations that participate constructively in rulemaking regardless of which party controls the executive or legislative branch accumulate a form of regulatory credibility that cannot be quickly rebuilt. ACPWB's sustained engagement — across administrations, across Congresses, across regulatory philosophies — is the foundation of our effectiveness as an advocate for the employer community.",
    ],
    1995: [
        "The 104th Congress arrived with an explicit mandate to curtail federal regulatory activity, and the compensation policy environment reflected that mandate directly. The Private Securities Litigation Reform Act, enacted over a presidential veto, reshaped the risk calculus around proxy-related disclosures. The Unfunded Mandates Reform Act imposed new analytical requirements on agency rulemaking, raising the bar for regulatory impact assessments that ACPWB and other employer groups had long argued were inadequate. Across {total} filings, our team engaged with a regulatory calendar defined less by new initiatives than by the reconsideration of existing ones.",
        "The congressional scrutiny of federal rulemaking gave our policy team unusual access to senior agency officials who were under pressure to justify regulatory costs with greater rigor. We submitted detailed economic analyses of several pending labor and compensation regulations, using our proprietary survey data to quantify employer compliance costs in ways that agency economists found credible and that congressional oversight staff found useful. This positioned ACPWB as a resource not merely for affected employers but for the regulatory institutions themselves.",
        "The focus on {theme} in {year} was shaped by the broader conversation about regulatory burden and cost-benefit accountability. Where prior years had produced rapid-fire proposals, {year} produced more deliberate agency engagement — longer comment periods, more substantive responses to commenters, more willingness to modify proposed rules in response to technical feedback. We found this environment productive, and our engagement with both the IRS on deferred compensation guidance and the SEC on executive compensation disclosure refinements yielded constructive outcomes.",
        "The institutional lesson of {year} was that regulatory rollback creates its own form of uncertainty. Employers who had structured programs in anticipation of requirements that were subsequently withdrawn faced transition costs that were real even if they attracted no public attention. Our guidance function — helping employers navigate the gap between what the law requires and what sound practice recommends — was, if anything, more valuable in a period of regulatory contraction than in one of regulatory expansion.",
    ],
    1996: [
        "The legislative output of {year} was unusually broad for a divided government. Minimum wage legislation, welfare reform, and the Health Insurance Portability and Accountability Act all enacted within a twelve-month window, each carrying compensation implications that our member employers were unprepared to address in isolation. ACPWB's {total} submissions spanned the DOL, the IRS, Treasury, and the newly constituted agencies responsible for HIPAA portability implementation, reflecting the range of policy fronts our team was required to engage simultaneously.",
        "The minimum wage increase — from $4.25 to $4.75 effective October 1, with a further increase to $5.15 scheduled for September 1997 — generated the most direct member inquiries. Employers with large hourly workforces needed analysis of wage structure compression effects and the implications for incentive pay programs tied to base wage levels. Our compensation benchmarking surveys provided real-time data on how employers were absorbing the cost increase and restructuring entry-level wage bands, data we shared with the DOL as it monitored aggregate employment effects.",
        "HIPAA's portability provisions introduced a new complexity into the relationship between benefit plan design and compensation strategy. Preexisting condition limitations on group health coverage had functioned as an implicit retention mechanism for a substantial number of employers; HIPAA's restrictions on those limitations altered the competitive landscape for talent in ways that were not immediately obvious. Our research team modeled these second-order effects and published guidance that many member employers incorporated into their workforce planning analyses.",
        "President Clinton's re-election in November provided continuity of regulatory direction that simplified our planning horizon considerably. The prospect of a second term without the distraction of a major transition allowed our team to focus on multi-year regulatory projects — particularly the ongoing dialogue with Treasury and the IRS on nonqualified deferred compensation frameworks — that benefit from sustained engagement rather than episodic intervention.",
    ],
    1997: [
        "The economic expansion of {year} created compensation pressures that the regulatory frameworks of the early decade had not been designed to address. Equity compensation usage reached historic levels, option grants were flowing to employees well below the executive level, and the question of how to account for and disclose these awards was moving from accounting theory into operational urgency. ACPWB's {total} filings reflected this shift, with a larger share than in prior years devoted to equity compensation design, disclosure, and tax treatment.",
        "IRS Notice 97-9 on split-dollar life insurance arrangements required immediate technical response from the compensation community. The notice signaled a fundamental reconsideration of the tax treatment of executive benefit arrangements that had been in place at hundreds of companies, and ACPWB submitted detailed comments on the potential regulatory approaches and their implications for existing programs.",
        "The Taxpayer Relief Act introduced modifications to capital gains rates and holding period requirements that had direct implications for option exercise strategies and the compensation benchmarking assumptions underlying long-term incentive design. Our technical guidance on the act's equity compensation provisions was among the most frequently requested content our team produced during the year.",
        "Our engagement with {theme} in the context of a prolonged bull market required a form of analytical discipline that is easy to lose when markets are rising. The structural questions about compensation design — pay mix, performance metric selection, time horizon alignment — are most consequential when addressed prospectively, before market conditions change. ACPWB's work in {year} on these structural questions would prove its value when market conditions eventually shifted.",
    ],
    1998: [
        "The Clinton impeachment proceedings consumed Washington's legislative attention through most of {year}, creating a regulatory environment in which the agencies operated with unusual autonomy from political direction. ACPWB's {total} filings addressed a calendar that was driven more by the organic pace of rulemaking than by political priority-setting, which proved congenial to the kind of technical, empirically grounded engagement at which our team excels.",
        "The equity markets reached extraordinary valuations through the spring, and the compensation implications of those valuations — option overhang, repricing pressure, the disconnect between option grant-date values and realized gain expectations — were beginning to create governance questions that institutional investors were not yet equipped to evaluate systematically. ACPWB's research on the economic characteristics of broad-based option programs was designed to give both employers and their shareholders a more rigorous analytical foundation for these discussions.",
        "Y2K compliance obligations intersected with compensation administration in ways that received insufficient attention in most organizational planning. The remediation of payroll systems, benefits administration platforms, and deferred compensation tracking tools had direct cost implications, and the liability questions around Y2K-driven benefit payment failures were genuinely unresolved. Our technical analysis of these questions was among the more practically valuable publications of {year}.",
        "The foundational work on {theme} that our team undertook in {year} reflected a recognition that the legal and regulatory framework for nonqualified deferred compensation had grown too complex and internally inconsistent to be managed without a comprehensive overhaul. The groundwork laid in this period would inform the Section 409A reform process that began in earnest after the Enron collapse brought deferred compensation governance to national attention.",
    ],
    1999: [
        "Y2K compliance preparation dominated organizational attention through the first three quarters of the year, but the regulatory calendar was nonetheless substantial — and our team submitted {total} filings while simultaneously managing the Y2K guidance demands that consumed significant staff capacity. The combination was a stress test of organizational discipline that we passed, and the Y2K remediation work itself provided valuable insight into the fragility of compensation administration infrastructure at many member employers.",
        "The Gramm-Leach-Bliley Act's modernization of the financial services regulatory framework had direct implications for compensation design at financial institutions, where the convergence of banking, insurance, and securities activities created new questions about the scope and comparability of executive compensation benchmarking. ACPWB's cross-sector survey capabilities were particularly valuable in this context.",
        "The SEC's ongoing refinement of executive compensation proxy disclosure requirements — including updated tabular formats and enhanced narrative disclosure expectations — occupied a significant share of our engagement with the commission. Our technical comments on the SEC's disclosure proposals addressed the tension between standardization, which facilitates investor comparison, and flexibility, which allows companies to present compensation programs in their proper context.",
        "The dot-com valuation environment was producing equity compensation design choices that, even at the time, raised questions about sustainability and alignment. Companies were granting options at strike prices that reflected market valuations with limited connection to underlying business fundamentals, and the governance framework for evaluating these grants had not kept pace with the scale of the awards. ACPWB's work on {theme} in {year} was in part an attempt to build the analytical infrastructure for the reckoning that we anticipated — correctly, as it turned out — was coming.",
    ],
    2000: [
        "The dot-com market collapse in March fundamentally altered compensation design discussions for the remainder of the year. Options that had been the centerpiece of competitive pay packages at technology and technology-adjacent companies were suddenly deeply underwater, and the governance questions around repricing, exchange programs, and underwater option replacement generated the most active member inquiry volume in our experience to that point. ACPWB's {total} filings addressed both the technical treatment of these programs and the governance frameworks for compensation committee decision-making in distressed equity conditions.",
        "SEC Regulation FD — adopted in August to restrict the selective disclosure of material nonpublic information — created immediate questions about how compensation committee discussions of performance targets and pay program design could be managed within the new disclosure framework. ACPWB submitted technical comments on the scope of Regulation FD's application to compensation-related communications and developed practical guidance for compensation committees managing the intersection of FD compliance and ongoing proxy season engagement.",
        "The presidential election's extended uncertainty through November and December created unusual regulatory planning challenges. With the outcome of the election genuinely unresolved for weeks, and with two candidates offering sharply different postures on executive compensation and labor regulation, our member employers faced the practical difficulty of anticipating regulatory requirements in a political environment that was literally undecided.",
        "The broader lesson of {year} for the compensation policy community was that the equity compensation model of the 1990s — option-heavy, grant-date-value-focused, relatively insensitive to downside risk — had been stress-tested by market conditions and found to have significant structural weaknesses. ACPWB's engagement with {theme} in this period was aimed at building the analytical and governance infrastructure for a more durable approach to equity-based pay that would perform across market cycles rather than only in a sustained bull market.",
    ],
    2001: [
        "September 11 interrupted a policy calendar that had already been shaped by the economic recession that began in March. Workforce reduction, benefit continuation under COBRA, the emergency modification of deferred compensation schedules to address liquidity needs, and the compensation implications of military deployment all generated immediate guidance demand. ACPWB's {total} filings spanned this emergency response work and the longer-term regulatory engagement that continued even as the country absorbed the shock of the attacks.",
        "The Enron collapse in October — and the revelation of the executive compensation and governance practices that had preceded it — transformed the policy environment in ways that would not be fully apparent until the following year's legislative response. The specific practices at issue: accelerated vesting triggered by the change-in-control mechanism, the insider trading that allowed executives to sell before the stock's collapse, the nonqualified deferred compensation losses suffered by employees who had no such exit — became the template for a decade of governance reform.",
        "The Economic Growth and Tax Relief Reconciliation Act addressed retirement savings in ways that our member employers had sought for years: higher contribution limits, catch-up contributions for workers over fifty, expanded Roth conversion opportunities, and a more generous deductibility framework for employer contributions. ACPWB's technical analysis of EGTRRA's retirement provisions was the most widely distributed research publication we produced in {year}.",
        "The focus on {theme} in {year} suddenly intersected with existential questions about governance and fiduciary responsibility that Enron had forced onto the national agenda. The technical policy questions we had been engaging for years — deferred compensation design, executive benefit governance, performance metric selection — were no longer abstract regulatory matters. They were the subject of congressional hearings, front-page coverage, and the direct concern of millions of workers who had watched their retirement savings disappear. ACPWB's technical engagement in this environment required a combination of analytical rigor and institutional humility that defined our approach to the regulatory reform process that followed.",
    ],
    2002: [
        "Sarbanes-Oxley was enacted in July under emergency legislative conditions unlike anything the regulatory community had experienced in decades. The executive compensation provisions it contained — Section 304 clawbacks, Section 402 loan prohibitions, Section 906 personal certifications — required immediate compliance assessment across every public company in the country. ACPWB's {total} filings in {year} were submitted against a background of institutional crisis that made technical precision both more difficult and more consequential than in ordinary regulatory cycles.",
        "WorldCom's collapse in July added urgency to what was already a legislative sprint. The scale of the accounting fraud, and the executive compensation practices that had accompanied it, reinforced congressional determination to enact comprehensive governance reform on an accelerated timeline that did not allow for the kind of deliberative technical comment process that produces the best regulatory outcomes. ACPWB engaged actively with committee staff throughout the legislative drafting process, seeking to ensure that the statutory text was technically workable even when it could not be fully deliberated.",
        "The SEC's accelerated rulemaking schedule to implement SOX's numerous executive compensation and governance provisions gave our team limited time to prepare substantive comments. We nonetheless submitted technical analysis on Section 404 internal controls and their intersection with incentive compensation design, on the scope of the Section 402 loan prohibition and its application to existing compensation arrangements, and on the standards for Section 304 clawback triggers. Each submission was the product of compressed but rigorous analysis.",
        "{theme} took on new meaning in {year}: the question was no longer whether executive pay required greater discipline, but whether legislated discipline could be structured to avoid unintended consequences for the much larger population of well-governed companies and non-executive officers whose programs were structurally sound. That remains the correct framing, and it is the one ACPWB brought to every regulatory engagement in the post-SOX period.",
    ],
    2003: [
        "The first full year of Sarbanes-Oxley compliance revealed implementation gaps that had not been visible in the legislative sprint of the prior year. ACPWB's {total} filings addressed the operational consequences — the scope ambiguities, the transition period questions, and the interaction of the new SOX requirements with compensation and governance frameworks that had been established before the statute's enactment.",
        "Section 404's internal control requirements generated the most acute compliance cost questions, and ACPWB's technical comments to the SEC on audit committee-compensation committee overlap and on materiality standards for incentive compensation controls were substantive contributions to the ongoing guidance dialogue. The economic burden of Section 404 compliance was distributed unevenly across the market, and our empirical analysis of that distribution supported the case for recalibration that the SEC would eventually address.",
        "The Jobs and Growth Tax Relief Reconciliation Act addressed capital formation and investment recovery, but its treatment of deferred compensation remained incomplete — a policy gap our team pressed Treasury to address directly. The groundwork laid in our technical submissions on nonqualified deferred compensation reform during {year} informed the American Jobs Creation Act provisions that would follow.",
        "Relations between the employer community and the SEC improved modestly through the year as rule implementation settled into a more deliberate pace following the emergency sprint of 2002. The career staff at the SEC's Division of Corporation Finance with whom we had built technical relationships over prior years remained in place, and our engagement with them on compensation disclosure refinements — including the early discussions that would eventually produce the 2006 comprehensive disclosure rewrite — was substantive and productive.",
    ],
    2004: [
        "The FASB's proposed Statement 123R — requiring expensing of all stock option grants at fair value — was the dominant technical discussion in executive compensation for {year}. ACPWB submitted extensive comments on the proposal's implications for broad-based equity programs, which had expanded dramatically during the 1990s and were now facing an accounting treatment change that would materially alter their apparent cost. Our {total} filings in {year} were anchored by this engagement and by the American Jobs Creation Act provisions that foreshadowed the Section 409A nonqualified deferred compensation reform.",
        "The AJCA's deferred compensation provisions — enacted in October — introduced an initial framework for the comprehensive reform that would become effective under Section 409A the following year. The statutory text raised more questions than it resolved, and ACPWB's preliminary technical analysis of the new framework was among the first substantive guidance the compensation community received on what 409A would require.",
        "The SEC's deliberative review of executive compensation disclosure requirements continued through {year}, with ACPWB participating in roundtable discussions and submitting technical comments that contributed to the comprehensive disclosure reforms that would follow in 2006. The seeds of what became the Compensation Discussion and Analysis requirement and the revised summary compensation table were planted in the technical exchanges of this period.",
        "President Bush's re-election in November provided regulatory continuity at the SEC and Treasury, maintaining the reform trajectory that had been established in the post-Enron environment. The second-term regulatory agenda — focused on disclosure quality rather than new substantive mandates — was well suited to the kind of technical engagement at which ACPWB excels. Our work on {theme} in {year} reflected the conviction that better disclosure produces better governance outcomes more reliably than any statutory mandate can.",
    ],
    2005: [
        "The effective date of FAS 123R for large accelerated filers arrived in the second quarter, and the practical implementation questions — fair value methodology selection, lattice model versus Black-Scholes, disclosure of key assumptions, the treatment of employee stock purchase plans — occupied significant ACPWB research capacity as part of our {total} filings. Option expensing had been debated for more than a decade; its arrival in the financial statements of the largest public companies produced a reconsideration of equity compensation program design that was swift and consequential.",
        "Section 409A became effective on January 1, and the IRS guidance process that was already underway would require sustained technical engagement through 2007. The scope of 409A's reach — covering virtually every arrangement that provided compensation in a year after the year in which it was earned — was broader than most practitioners had anticipated, and the transition period questions were numerous and genuinely complex. ACPWB submitted detailed technical comments on the initial proposed regulations and participated in IRS practitioner roundtables on implementation priorities.",
        "Hurricane Katrina in August raised compensation administration questions that had no established regulatory framework: emergency pay continuation, displacement-related benefit qualification, the tax treatment of employer assistance to affected employees, and the implications for deferred compensation plans of employees who could not access plan records. Our technical analysis of these questions — developed in collaboration with the IRS, DOL, and Treasury — was among the more practically valuable guidance we have produced.",
        "The year demonstrated the value of organizational investment in technical depth. The simultaneous demands of FAS 123R implementation, Section 409A transition planning, SOX continued compliance, and the Katrina-related emergency work would have overwhelmed an organization that had not built the analytical capacity and regulatory relationships required to address multiple complex technical issues in parallel. {theme} remained at the center of our regulatory engagement, and our ability to integrate tax, accounting, disclosure, and governance into a single analytical framework proved its value.",
    ],
    2006: [
        "The SEC's adoption of comprehensive executive compensation disclosure regulations in August — the most significant rewrite of the proxy disclosure rules in more than a decade — defined {year} for compensation professionals and generated more ACPWB filing activity than any regulatory event since Sarbanes-Oxley. Our {total} submissions addressed both the proposed and final rules, as well as the implementation questions that emerged in the months following adoption. The new Compensation Discussion and Analysis requirement, the revised summary compensation table, and the total compensation single-figure mandate each required detailed implementation guidance that our team developed through the final quarter.",
        "The stock option backdating scandals that emerged throughout {year} — ultimately affecting more than 130 publicly traded companies — cast a retrospective shadow over equity compensation practices that had been widely considered routine. The specific practice at issue — selecting grant dates retrospectively to capture favorable strike prices — was clearly improper, but the investigation process surfaced related questions about option grant procedures, documentation standards, and the governance role of compensation committees that affected far more companies than those directly implicated in misconduct.",
        "FAS 123R became fully effective for all public companies during {year}, resolving a period of accounting uncertainty that had persisted since the FASB finalized the standard. The combination of new accounting requirements and new SEC disclosure obligations created a moment of genuine transparency in executive pay that the institutional investor community welcomed and many issuers found challenging. ACPWB's integrated guidance — connecting accounting treatment, SEC disclosure requirements, and compensation committee governance — helped member employers navigate this convergence.",
        "{theme} in {year} was no longer a policy aspiration but a legal requirement with enforcement consequences. Our institutional role shifted accordingly — from advocacy for better disclosure to guidance on implementation, and from comment letters to technical workshops for compensation committee chairs and general counsel. The quality of implementation would determine whether the new disclosure framework produced the governance improvements the SEC had intended, and ACPWB invested heavily in ensuring that the technical standards for implementation were set as high as the regulatory text permitted.",
    ],
    2007: [
        "Section 409A deferred compensation regulations became effective on January 1, and the implementation questions that had been building through two years of guidance filings converged into a single compliance deadline. The scope of 409A's application — reaching defined benefit supplemental plans, employment agreements, bonus programs, and virtually every other arrangement providing deferred compensation — meant that compliance review was required across the full range of executive pay programs at every public company. Our {total} filings in the first half of the year addressed the concentrated guidance demand this deadline created.",
        "The SEC's proxy access debate began in earnest, with significant implications for compensation committee accountability. ACPWB submitted technical comments on the proposed proxy access framework, arguing that the governance benefits of direct shareholder nomination must be balanced against the operational costs of contested director elections and the risk that short-term activist pressures would distort the long-term performance orientation that sound compensation design requires.",
        "The early signals of stress in subprime mortgage markets — the Bear Stearns hedge fund failures in June, the broader credit market seizure in August — began to reshape the conversation about incentive compensation and risk alignment in the financial services sector. The questions about pay structure and systemic risk that would dominate 2008 and 2009 were already visible in these early months, and ACPWB's preliminary analysis of the relationship between compensation design and risk-taking at financial institutions reflected our recognition that a significant policy challenge was forming.",
        "{theme} dominated the regulatory agenda through {year}, and ACPWB's technical contributions on this front were substantive. We were called to testify before a Senate Finance subcommittee on the tax treatment of performance-based compensation and submitted companion written testimony addressing the empirical evidence on incentive compensation effectiveness. Patient, sustained, technically grounded engagement produces better regulatory outcomes — that is the institutional commitment ACPWB brings to every proceeding, and it was demonstrated with particular clarity in the Section 409A guidance process that concluded in {year}.",
    ],
    2008: [
        "The September collapse of Lehman Brothers, the passage of TARP, and the election of Barack Obama all occurred within a seven-week span that fundamentally altered the regulatory environment for executive compensation. The policy response was faster and more sweeping than any we had encountered in our organizational history. ACPWB's {total} submissions in {year} included emergency comment letters, technical analyses of proposed TARP compensation restrictions, and formal testimony before three congressional committees — the heaviest engagement workload our team had ever sustained.",
        "TARP's executive compensation conditions — the prohibition on golden parachutes for senior executives of recipient firms, limits on incentive pay, and the appointment of a Special Master for compensation review — created a compliance framework with no precedent in American corporate law. ACPWB submitted detailed technical comments on the Treasury's implementing regulations, with particular focus on the scope of the 'senior executive officer' definition, the treatment of compensation arrangements already in place at the time of TARP participation, and the interaction of the TARP conditions with existing employment contracts and severance obligations.",
        "The broader economy's distress surfaced compensation questions well beyond the financial sector: severance plan funding and priority in bankruptcy proceedings, the tax treatment of forgiven debt as compensation income, and the implications of workforce reductions on deferred compensation distribution schedules. Our guidance on these questions was developed under conditions of significant legal uncertainty, and we were careful to frame our analysis as technical assessment rather than definitive compliance guidance in areas where the law was genuinely unsettled.",
        "The elections of November produced a political transition of unusual significance for compensation policy. The incoming administration's stated priorities — stronger executive pay oversight, mandatory shareholder advisory votes, and a substantially expanded financial regulatory apparatus — would define the regulatory calendar for the following several years. ACPWB began its transition planning immediately after the election results were known, seeking to ensure that our technical engagement with the new administration's agenda would be timely, credible, and grounded in the kind of empirical analysis that career agency staff find most useful.",
    ],
    2009: [
        "The new administration moved on executive compensation within its first weeks. The American Recovery and Reinvestment Act imposed sweeping restrictions on TARP recipient firms, and Treasury's appointment of Kenneth Feinberg as Special Master for TARP Executive Compensation introduced a form of government pay determination with no close precedent in American corporate governance. ACPWB submitted {total} filings across the year addressing TARP conditions, the proposed Dodd-Frank framework, and the related agency guidance that accumulated with unusual speed through the legislative drafting process.",
        "The Special Master's review process — covering senior executive compensation at the seven largest TARP recipients — required the kind of technical benchmarking analysis that our team was uniquely positioned to provide. We submitted voluntary technical comments to the Special Master's office on the methodological challenges of benchmarking compensation for executives whose roles had no direct market comparables, and we testified before the House Financial Services Committee on the policy risks of government-determined pay rates for entire industries.",
        "The legislative drafting of what would become Dodd-Frank consumed much of the second half of the year. Multiple competing proposals circulated in the House and Senate banking committees, and ACPWB engaged actively with committee staff on the say-on-pay provision, the pay ratio disclosure requirement, and the clawback mandate — in each case arguing for technically workable implementations that would achieve legitimate governance objectives without creating unmanageable compliance burdens.",
        "The systemic risk framing of {theme} in {year} required our team to engage with questions that had not previously been central to our work: the relationship between compensation design and institutional risk appetite, the governance mechanisms for compensation committee oversight of risk-taking incentives, and the empirical evidence on whether executive pay structure had contributed to the financial sector risk concentration that produced the crisis. Our analysis of these questions — cautious where the evidence was limited, direct where it was clear — defined our engagement with the Dodd-Frank drafting throughout.",
    ],
    2010: [
        "Dodd-Frank was signed into law on July 21, and the compensation provisions it contained — mandatory say-on-pay votes, CEO pay ratio disclosure, clawback requirements, pay-versus-performance reporting, and proxy access — represented the most comprehensive legislative intervention in executive compensation since Sarbanes-Oxley. The statute's implementation timeline was compressed, and ACPWB immediately shifted to the engagement mode that major new legislation requires: rapid technical assessment, preliminary implementation guidance, and sustained participation in the rulemaking processes that would give the statute its operational content. Our {total} filings in {year} reflected this intensity.",
        "The say-on-pay mandate drew the most immediate attention. Mandatory advisory votes on executive compensation would apply to all public companies beginning in 2011, and the governance implications for compensation committee chairs — who would now be directly accountable to a shareholder majority — were profound. ACPWB developed initial guidance on say-on-pay board engagement practices, analytical frameworks for evaluating likely vote outcomes, and technical comments to the SEC on the implementing regulations.",
        "The Affordable Care Act, enacted in March, introduced a new dimension to total compensation analysis that required integration with existing health benefit benchmarking. The employer shared responsibility provisions, the large employer reporting requirements, and the excise tax on high-cost benefit plans all had compensation implications that our team worked to address in the context of the full compensation package rather than in isolation.",
        "{theme} became, after Dodd-Frank, not merely a policy priority but a statutory obligation. The challenge our member employers now faced was not whether to engage with these issues but how to do so in ways that were technically defensible, shareholder-responsive, and operationally sustainable. ACPWB's implementation guidance work — practical, technically grounded, and updated as the agencies issued implementing rules — became the central contribution of our policy function for the years ahead.",
    ],
    2011: [
        "The first year of mandatory say-on-pay votes produced results that surprised many observers. Shareholder support for executive compensation programs averaged above 90 percent across the Russell 3000, and the minority of firms that received low votes were concentrated in a relatively small number of sectors and pay structure types. ACPWB analyzed the first-year results in detail, identifying the program design features most predictive of shareholder opposition, and submitted {total} filings that included both the aggregate analysis and technical comments on the Dodd-Frank implementation rulemakings that continued throughout the year.",
        "The SEC's pay ratio rulemaking drew particular attention. The statutory requirement to disclose the ratio of CEO compensation to median employee compensation raised methodological questions of genuine complexity — how to define 'employee' for ratio purposes, whether to annualize compensation for part-time and seasonal workers, how to handle global workforces subject to privacy laws that restricted personal data collection. Our technical comment letter on the proposed rule addressed each of these questions with the analytical rigor they required.",
        "The debt ceiling crisis of July and August, and the broader pattern of political dysfunction in Washington that characterized {year}, did not materially slow the SEC's Dodd-Frank implementation schedule. The agency's institutional capacity to continue rulemaking work independent of the legislative environment — a product of its career staff and its established procedural infrastructure — was an important stabilizing factor during a period of considerable political turbulence.",
        "The year reinforced a pattern we had observed consistently: major statutory compensation requirements tend to produce their most significant governance changes not in the first year of application but in the two or three years of anticipatory preparation that precede the formal compliance deadline. The companies that were best positioned for say-on-pay were those that had begun substantive shareholder engagement on compensation program design in 2009 and 2010. Early engagement with {theme} produces governance quality that reactive compliance cannot replicate.",
    ],
    2012: [
        "The JOBS Act, signed in April, reduced reporting burdens for emerging growth companies in ways that touched several of our core policy areas, including executive compensation disclosure, say-on-pay frequency, and equity compensation plan disclosures. ACPWB submitted technical comments on the implementing regulations and practical guidance for the many member employers seeking to understand the emerging growth company framework — contributions that were part of our {total} filings for the year, which also included sustained engagement with the SEC's continuing Dodd-Frank implementation work.",
        "The second year of mandatory say-on-pay votes produced a more nuanced pattern than the first. Shareholder engagement practices had improved significantly at most large-cap companies, but the mid- and small-cap segments showed more variation, and the firms that received low votes in year two were meaningfully different in profile from those criticized in year one. ACPWB's longitudinal analysis of the two-year pattern identified program design features, disclosure quality indicators, and shareholder engagement practices that predicted vote outcomes with increasing precision.",
        "The fiscal cliff negotiations that consumed the final months of {year} created substantial uncertainty around capital gains rates, dividend taxation, and the treatment of equity compensation — particularly for executives and employees with large unrealized option positions. Our guidance on compensation implications of the various cliff scenarios was among the most frequently requested content our team produced, and the accelerated vesting and option exercise decisions that many companies made in the final weeks of the year generated significant compliance and disclosure questions that we addressed through emergency guidance publications.",
        "{theme} in {year} operated in the context of continuing economic uncertainty and the shareholder governance expectations that the post-financial-crisis environment had solidified. Compensation programs that were defensible in 2007 often required substantial restructuring to meet the governance standards that institutional investors and proxy advisors were now applying systematically. ACPWB's role in helping member employers navigate this recalibration — through both technical analysis and practical communication frameworks — was among our most consequential contributions of the year.",
    ],
    2013: [
        "The federal government shutdown in October — sixteen days during which the SEC and other regulatory agencies halted most rulemaking activity — was a visible reminder that compensation policy does not advance on a predictable calendar. ACPWB submitted {total} filings in {year}, a number that reflected the genuine regulatory activity outside the shutdown window as much as the disruption within it. Sequestration's across-the-board spending cuts affected agency staffing and rulemaking capacity throughout the year, producing longer response times, deferred comment periods, and an overall reduction in regulatory output.",
        "The SEC's pay ratio rulemaking continued to slip from its original post-Dodd-Frank timeline without a finalized proposed rule. ACPWB maintained its engagement with the commission's staff on the methodological questions that remained unresolved, with particular attention to the statistical sampling approaches that would determine compliance costs for large and complex global workforces. Our technical work during this period of regulatory delay was not wasted: it formed the foundation of the detailed comment letter we submitted when the proposed rule was eventually released.",
        "The institutional investor community continued to refine its executive compensation voting policies, and the proxy advisory firms updated their methodologies in ways with direct implications for compensation committee decision-making. ACPWB published detailed analysis of the updated proxy advisor policies and submitted technical comments to the SEC on the proper regulatory treatment of proxy advisory services — arguing that services that have developed functional regulatory authority through their influence on institutional voting deserve a commensurate level of regulatory accountability.",
        "The ACA rollout difficulties — the healthcare.gov launch failure and the enrollment problems that followed — created compensation implications for employers managing benefit plan design choices under the employer shared responsibility provisions. ACPWB's health benefit and total compensation teams worked through these questions together, producing integrated guidance that addressed the interaction between ACA cost obligations and competitive compensation positioning. Our engagement with {theme} in {year} was shaped by this recognition that total compensation analysis cannot compartmentalize benefits and pay.",
    ],
    2014: [
        "The SEC's September release of the proposed pay ratio rule — nearly four years after Dodd-Frank mandated it — generated more public comment letters than any SEC proposal in recent memory. ACPWB's submission was among the most technically substantive in the record, addressing the statistical methodology for median employee identification, the treatment of non-U.S. employees under the proposed exemption framework, and the interaction of the ratio disclosure with the existing executive compensation tables. Our {total} filings for the year were anchored by this engagement, which drew on two years of preliminary technical work we had undertaken in anticipation of the proposal.",
        "The midterm elections returned control of the Senate to Republicans, creating a divided government configuration that predictably slowed the legislative pipeline but had more limited effect on agency rulemaking than the political calendar might suggest. The SEC under Chair White continued its Dodd-Frank implementation schedule, and the DOL's initial steps toward overtime rule revision were announced before November — a signal of regulatory ambition that would be confirmed when the proposed rule was released the following year.",
        "The DOL's overtime threshold review — the first since 2004 — signaled an intention to substantially increase the salary level below which employees are automatically entitled to overtime pay. ACPWB engaged early in the agency's preliminary information-gathering process, submitting data on the distribution of exempt employees by salary level across industries and occupations, and providing empirical analysis of the likely employment effects of various threshold increase scenarios. The compensation structure implications of reclassification at scale were among the most consequential questions our research team had worked through in years.",
        "The capacity for integrated, multi-agency policy engagement that ACPWB had built over two decades was tested in {year} as our member employers faced prospective requirements from the SEC, DOL, and IRS on overlapping timelines. {theme} remained a leading regulatory priority in this environment, and ACPWB's technical contributions on this front — empirically grounded, practically oriented, and clearly communicated to both agency staff and our member community — demonstrated the value of sustained institutional engagement over reactive compliance responses.",
    ],
    2015: [
        "The SEC finalized the pay ratio disclosure rule in August — the last major Dodd-Frank executive compensation rulemaking to be completed — and ACPWB immediately shifted from advocacy to implementation guidance. The final rule's flexible methodologies for median employee identification, including statistical sampling and consistently applied compensation measures, reflected many of the technical points our comment letters had advanced. Our {total} filings in {year} included post-adoption technical analysis, guidance on compensation committee communication responsibilities, and detailed implementation frameworks that member employers incorporated into their proxy preparation processes.",
        "The DOL's July release of the proposed overtime rule — raising the salary threshold from $455 to $970 per week — was the most consequential labor cost proposal in a decade. ACPWB submitted one of the most detailed economic analyses in the comment record, drawing on proprietary survey data representing over eight thousand employers. Our modeling of the employment effects, compensation structure impacts, and reclassification costs was cited by the DOL's own economists in the regulatory impact analysis and informed both the formal comment record and informal technical assistance to agency staff throughout the rulemaking process.",
        "The DOL's proposed fiduciary rule — expanding the definition of investment advice fiduciary under ERISA — had significant implications for the advice that compensation consultants and retirement plan advisors could provide to plan participants. ACPWB submitted technical comments addressing the rule's impact on compensation consulting engagements that touched retirement plan design, and we worked with our consulting member firms to develop compliance frameworks for the new fiduciary standard.",
        "{theme} reached an inflection in {year}: the statutory framework had been set by Dodd-Frank, the major SEC implementing rules were now finalized, and the compliance challenge had shifted from 'what will be required' to 'how to implement it well.' ACPWB's technical guidance on the mechanics of pay ratio computation — including statistical sampling methodology for global workforces, the treatment of part-time and seasonal compensation, and the narrative disclosure expectations — was adopted by a substantial share of the public company community.",
    ],
    2016: [
        "The November election of Donald Trump produced an immediate recalibration of the regulatory outlook that defined compensation policy planning for the following several years. ACPWB's {total} filings in {year} were submitted against a backdrop of genuine regulatory uncertainty: a federal district court's preliminary injunction blocking the DOL overtime rule days before its December 1 effective date, the incoming administration's stated intention to revisit multiple Dodd-Frank provisions, and the first fiscal years subject to the pay ratio rule approaching their close.",
        "The DOL overtime rule's November injunction left employers who had already restructured classification and compensation programs in an uncertain position. ACPWB provided immediate guidance on the legal status of employer elections made in anticipation of the effective date, the implications for exempt status determinations, and the options for reversing or maintaining changes that had already been communicated to employees. The guidance was among the most operationally consequential work our team produced in {year}, and the volume of member inquiries in the days following the injunction reflected the scale of the compliance preparation that had been underway.",
        "Pay ratio compliance preparation continued through {year} for calendar-year issuers, whose first disclosures would appear in spring 2018 proxy statements. ACPWB's technical workshops on median employee identification methodology, statistical sampling design, and the treatment of compensation elements in the annual total compensation calculation were among the most attended events in our organizational history. The practical complexity of the rule — underestimated by many observers when it was finalized — was becoming apparent as employers worked through their first census and identification exercises.",
        "{theme} was a genuine area of constructive progress in a year otherwise defined by regulatory uncertainty and political transition. ACPWB's technical contributions on this front were substantive and recognized by the relevant agencies, and the institutional relationships we maintained with career staff at the SEC and DOL through the political transition would prove valuable as the new administration began its regulatory review processes.",
    ],
    2017: [
        "The Tax Cuts and Jobs Act — enacted in December after accelerated legislative consideration — was the most consequential compensation tax legislation since the Omnibus Budget Reconciliation Act of 1993. ACPWB's {total} filings in {year} were dominated by engagement with the TCJA's executive compensation provisions: the modification of Section 162(m) to expand the definition of covered employee to include the CFO and to eliminate the performance-based compensation exception, the new excise tax on excess parachute payments from tax-exempt organizations, and the transition rules governing written binding contracts in place before November 2, 2017.",
        "Section 162(m)'s modification effectively ended the compensation planning strategy that had dominated proxy season discussions for twenty-four years. Long-term incentive plans structured to qualify for the performance-based exception — with all their attendant governance requirements — required fundamental reconsideration. ACPWB's technical guidance on the transition rules, the grandfather relief standards, and the redesign implications for existing equity programs was our most consequential analytical contribution of the year, and our comments on the preliminary IRS guidance were among the most technically detailed in the regulatory record.",
        "The DOL's effective abandonment of the fiduciary rule — first delayed by administrative action in April and subsequently vacated by the Fifth Circuit — resolved a source of compliance uncertainty that had occupied our member employers for two years. ACPWB's position throughout had been that the rule's expansion of fiduciary status was insufficiently calibrated to the compensation and advisory relationships it was intended to address, and the legal proceedings confirmed the regulatory overreach we had documented in our technical comments.",
        "The #MeToo movement's arrival in corporate governance by year-end created new compensation committee questions: the clawback implications of misconduct-related terminations, the disclosure obligations when settlements affected compensation programs, and the governance process for addressing executive misconduct short of termination. ACPWB provided practical guidance on each of these questions in the final weeks of the year, drawing on the existing clawback and forfeiture frameworks that Dodd-Frank had established and connecting them to the more discretionary governance decisions that boards were being asked to make.",
    ],
    2018: [
        "The first wave of pay ratio disclosures arrived in spring proxy statements, and the results revealed both the technical diversity of the disclosures and their substantive significance. CEO-to-median-employee ratios across the Russell 1000 ranged from below 50:1 to above 5,000:1, reflecting differences in industry, workforce structure, business model, and compensation program design that a single-figure metric inevitably compresses. ACPWB published the most detailed cross-industry analysis of first-year pay ratio data, drawing on our proprietary survey resources and providing context that the bare ratios could not convey. Our {total} filings in {year} were anchored by this disclosure analysis and by continued TCJA implementation guidance.",
        "The TCJA's Section 162(m) modifications continued to generate implementation questions throughout {year}, particularly around the grandfather relief standards for written binding contracts and the application of the new covered employee definition to executives below the CEO and CFO level. IRS Notice 2018-68, issued in August, provided initial guidance on grandfathering and ACPWB submitted technical comments on the issues the notice had left unresolved — a pattern of engagement that reflects the iterative nature of major tax reform implementation.",
        "The #MeToo movement had moved fully into corporate boardrooms by {year}, and the governance response — clawback policy reviews, severance agreement revisions, miscellaneous compensation disclosures related to settled misconduct claims — created new technical questions for compensation committees at a pace that legislative and regulatory frameworks had not anticipated. ACPWB published guidance on the relationship between Dodd-Frank clawback requirements and the discretionary recoupment policies that boards were adopting, and we provided technical assistance to compensation committee chairs working through specific factual situations.",
        "{theme} in {year} required our team to navigate a compressed redesign cycle created by the convergence of TCJA restructuring, pay ratio disclosure implementation, and heightened governance scrutiny. Our integrated analytical approach — connecting tax, accounting, disclosure, and governance — proved particularly valuable in this environment. The proxy season of {year} reinforced that executive compensation disclosure had reached a new level of institutional sophistication: institutional investors, proxy advisors, and the financial press were all more analytically capable than they had been five years earlier, and compensation committees that had not invested in disclosure quality faced material shareholder risk.",
    ],
    2019: [
        "The Business Roundtable's August Statement on the Purpose of a Corporation — abandoning the shareholder primacy doctrine in favor of a multi-stakeholder model — arrived at a moment when ACPWB was already engaged with the policy implications of purpose-driven governance for compensation design. The statement committed its signatories to investing in their employees, which raised direct questions about the metrics and disclosures that should accompany that commitment. Our {total} filings in {year} addressed the translation of stakeholder governance commitments into concrete compensation program features, including long-term incentive metrics beyond total shareholder return and the treatment of workforce investment in pay-versus-performance analysis.",
        "The SECURE Act — signed in December — represented the most significant retirement plan legislation since the Pension Protection Act of 2006. Its implications for employer-sponsored plan design were substantial: expanded access requirements for part-time employees, new lifetime income illustration requirements for participant statements, and modifications to the rules governing multiple employer plans all required integration with existing compensation benchmarking frameworks. ACPWB's technical implementation guidance became a primary reference for the member community as it worked through the SECURE Act's requirements.",
        "The Federal Trade Commission signaled early interest in a potential rulemaking on noncompete agreements — a policy area that intersects directly with compensation program design through its implications for retention incentive structures, severance conditionality, and equity vesting terms. ACPWB submitted technical comments in the FTC's preliminary information-gathering process, presenting empirical evidence on the relationship between noncompete enforceability and compensation levels in affected labor markets and arguing for a targeted approach that distinguished between legitimate protection of proprietary information and anticompetitive labor market restriction.",
        "{theme} in {year} was engaged at a genuine inflection. The post-financial-crisis regulatory architecture was in place, ESG expectations were creating new governance requirements that statutory frameworks had not yet addressed, and the FTC's early signals on noncompetes suggested that labor market competition was about to receive sustained federal attention. ACPWB entered the following year with clear analytical priorities, strong institutional relationships, and a member community whose sophistication on compensation governance had grown substantially over the preceding decade.",
    ],
    2020: [
        "The COVID-19 pandemic produced a compensation policy emergency unlike anything in ACPWB's organizational history. The shift from ordinary regulatory engagement to crisis guidance — beginning in March and sustaining through year-end — produced our highest member inquiry volume in three decades of operation. Our {total} submissions addressed CARES Act executive compensation conditions, furlough and pay reduction governance, FFCRA paid leave requirements, PPP forgiveness interactions with executive compensation, and the design of COVID-related retention programs for organizations managing simultaneous workforce reduction and talent retention challenges.",
        "The CARES Act's restrictions on executive compensation at companies receiving direct loans created an asymmetric compliance landscape: firms that accepted CARES assistance faced legacy compensation commitments made before the pandemic that could not be honored without regulatory consequence. ACPWB submitted technical comments to Treasury and the SBA on the interpretation of these restrictions and provided practical implementation guidance to affected employers — work that required rapid legal analysis under conditions of significant statutory ambiguity.",
        "The shift to remote work at scale created compensation benchmarking challenges that existing analytical frameworks were not designed to address. Geographic pay differentiation — historically tied to cost-of-living and local market conditions — was suddenly complicated by employees working in locations outside their employer's established market geography. ACPWB launched an emergency benchmarking study on remote work compensation practices and submitted a white paper to the DOL on the geographic pay differential implications of extended remote work arrangements.",
        "The economic disruption of the pandemic intersected with the pay equity movement in ways that sharpened the urgency of our work on {theme}. The disproportionate economic impact on lower-wage workers — concentrated in industries without the ability to shift to remote work — put pay equity questions at the center of public discourse with direct policy consequences. Compensation committees in {year} faced a genuine reckoning about the relationship between executive and workforce compensation, driven not only by regulatory requirements but by the reputational and talent market consequences of visible pay disparities in a period of acute economic hardship.",
    ],
    2021: [
        "The American Rescue Plan's passage in March — followed by a sequence of Biden administration executive orders on federal contractor compensation, worker classification, and pay equity reporting — established the compensation policy agenda for {year} with unusual clarity. ACPWB's {total} filings addressed the implementing regulations for the federal contractor minimum wage increase, the pay equity reporting requirements signaled in executive orders, and the compensation implications of expanded paid leave proposals that circulated through the Build Back Better negotiations.",
        "The Great Resignation — the sustained wave of voluntary job separations that began in spring and intensified through the remainder of {year} — created compensation benchmarking pressures that our member employers had not encountered since the late-1990s technology labor market. ACPWB's quarterly compensation movement surveys became the most closely tracked publication in our research portfolio as employers scrambled to assess whether their pay structures were competitive in a market where voluntary attrition had reached historic levels.",
        "The FTC's accelerating interest in noncompete agreements — culminating in a request for information issued in January 2022 but seeded with research and advocacy work throughout {year} — occupied a significant share of our policy team's engagement capacity. ACPWB submitted preliminary technical comments addressing the relationship between noncompete enforceability and compensation program design, arguing that the empirical evidence supported targeted reforms rather than the comprehensive prohibition that advocacy groups were pressing the agency to adopt.",
        "{theme} became a lived operational reality rather than a policy discussion point in {year}. Employers managing intense wage pressure, retention competition, and the recruitment dynamics of a radically mobile labor market discovered that compensation program infrastructure designed for stable conditions was inadequate for the environment they were navigating. ACPWB's rapid guidance on market-adjustment pay actions, equity refresh programs, and retention incentive design was among our most practically consequential work in years.",
    ],
    2022: [
        "The SEC's August adoption of the pay-versus-performance disclosure rule — requiring companies to report 'compensation actually paid' alongside TSR and net income benchmarks — completed the core Dodd-Frank executive compensation rulemaking agenda more than a decade after the statute's enactment. ACPWB's {total} filings in {year} included substantive technical comments on the implementing guidance, detailed analysis of the compensation actually paid calculation methodology, and implementation frameworks that our member employers incorporated into their proxy preparation processes.",
        "The SEC's October adoption of the final clawback listing standards — requiring mandatory recoupment of erroneously awarded compensation without fault or misconduct — produced a new layer of governance obligation for all listed companies. ACPWB submitted technical comments addressing the clawback trigger definition, the interaction with existing voluntary clawback policies, and the administrative processes that compensation committees would need to establish. The December 2023 compliance deadline gave member employers fourteen months to prepare, and ACPWB's implementation guidance began immediately.",
        "The Federal Reserve's aggressive rate hike cycle — the fastest monetary tightening in four decades — was the dominant macroeconomic event of {year}, with direct consequences for compensation program design. Underwater equity awards returned after years of absence, interest rate assumptions in pension and deferred compensation calculations shifted dramatically, and the cost of capital benchmarks underlying long-term incentive design required fundamental revision. ACPWB's inflation-impact compensation benchmarking analysis — drawing on data from over four thousand participating employers — was the most widely cited research publication in our organizational history.",
        "{theme} in {year} was engaged within a macro environment that tested compensation program assumptions at their foundations. The combination of regulatory finalization — pay-versus-performance, mandatory clawbacks — and market disruption — inflation at forty-year highs, equity volatility, rate increases — created a compressed redesign cycle that demanded both technical precision and strategic adaptability from compensation committees. ACPWB's integrated support across both regulatory and market dimensions defined the character of our {year} engagement.",
    ],
    2023: [
        "The Federal Trade Commission's January release of its proposed rule banning virtually all noncompete agreements — covering an estimated thirty million workers — was the most consequential competition-policy intervention in workforce compensation in decades. ACPWB submitted one of the most technically detailed comments in the record, drawing on empirical research linking noncompete enforceability to compensation levels, retention program design, and innovation outcomes. Our {total} filings in {year} were anchored by this engagement and by the parallel SEC clawback implementation work that reached its compliance deadline in November.",
        "The Silicon Valley Bank collapse in March, and the subsequent stress in the regional banking sector, revived debates about incentive compensation and risk alignment that had last been this acute during the 2008-2009 financial crisis. Congressional examination of the pay structures at failed institutions, and the questions about insider stock sales before the collapse, created both regulatory risk and governance opportunity for financial sector employers. ACPWB provided technical analysis of the risk-alignment questions and submitted preliminary comments to the banking regulators on the compensation governance frameworks applicable to covered institutions.",
        "The SEC clawback listing standard's November 28 effective date required all listed companies to adopt a compliant mandatory recoupment policy by December 1. ACPWB's implementation guidance — developed through the preceding eighteen months — addressed the mechanics of mandatory recoupment calculations, the governance procedures for trigger determinations, and the interaction of the mandatory clawback with existing discretionary forfeiture provisions. The compliance quality we observed across the member community reflected the sustained guidance investment our team had made.",
        "Generative AI arrived in the compensation profession with unusual speed, and the governance questions it raised became pressing before the regulatory framework had formed. When AI tools are used to establish pay ranges, evaluate job architecture, or conduct pay equity analysis, what human oversight is required and what are the employment discrimination law implications of algorithmic compensation decisions? ACPWB convened a technical working group on AI in compensation governance and submitted a white paper to the EEOC addressing these questions — establishing an analytical foundation for the regulatory engagement that {theme} in the coming years would require.",
    ],
    2024: [
        "The federal district court's August ruling striking down the FTC's noncompete rule — and the subsequent appellate stay preventing enforcement pending further proceedings — produced one of the more unusual policy outcomes in our organizational experience: a regulation proposed, finalized, litigated, and vacated within eighteen months, leaving the underlying policy questions entirely unresolved at the federal level. ACPWB's {total} filings in {year} addressed both the legal proceedings and the substantive noncompete policy alternatives that Congress and state legislatures might pursue in the absence of a viable FTC rulemaking pathway.",
        "State pay transparency legislation continued to proliferate, with California, Colorado, New York, Washington, Illinois, and several additional states implementing or materially expanding requirements to disclose pay ranges in job postings and to provide compensation information to current employees upon request. ACPWB submitted technical comments to state agencies in multiple jurisdictions and published a comprehensive multi-state compliance guide addressing the interaction of overlapping requirements for employers with workforces across jurisdictions.",
        "The SEC's pay-versus-performance disclosure framework — in its first full proxy season — produced disclosure patterns that revealed the limitations of the compensation actually paid metric. The mark-to-market equity valuation approach produced volatile figures that neither reflected compensation committee intent nor correlated reliably with underlying business performance in the year of measurement. ACPWB submitted a technical comment letter to the commission recommending modifications to the CAP calculation methodology, and we presented the analysis directly to SEC staff.",
        "The November election returned a Republican administration that had signaled intentions to revisit multiple Biden-era labor regulations, to limit the reach of the FTC's competition authority, and to reduce the federal regulatory footprint in employment matters. ACPWB's transition analysis — prepared before the election and updated in its aftermath — gave member employers a clear-eyed assessment of which regulatory requirements were likely to be modified, which were statutory and immune to executive revision, and which carried sufficient institutional support to survive a change in administration. Sustained technical engagement with career agency staff — which ACPWB maintained throughout the political transition — is the most reliable hedge against policy uncertainty.",
    ],
    2025: [
        "The new administration's executive orders on diversity, equity, and inclusion programs — issued in the opening weeks of {year} — created immediate compliance uncertainty for federal contractors and broader governance questions for all public companies that had adopted formal DEI commitments with compensation metric linkages. ACPWB's {total} filings addressed the scope of the federal contractor orders, the Title VII implications of voluntary DEI program modifications, and the disclosure obligations for companies revising ESG-linked incentive metrics in response to the changed federal posture.",
        "The broad mandate to reduce federal regulatory output — and the accompanying reduction in force at multiple agencies — materially altered the rulemaking environment. ACPWB maintained active engagement with career staff at the SEC, DOL, and IRS throughout the transition, recognizing that the technical relationships built over decades of sustained engagement represent institutional infrastructure that survives changes in political direction. Our {total} filings reflect that the underlying regulatory agenda did not stop; its pacing and priority-setting changed.",
        "Artificial intelligence integration into compensation decision-making continued to accelerate, and the governance questions it raised became pressing before the regulatory framework had formed. ACPWB convened a technical working group on AI in compensation governance, engaging with the EEOC and DOL on the employment discrimination law implications of algorithmic pay decisions and publishing practical governance guidance for compensation committees that are evaluating or already deploying AI-assisted pay equity analysis, job architecture tools, and benchmarking platforms.",
        "{theme} remained a priority engagement area in a policy environment that had shifted dramatically toward state-level activity. With federal regulatory contraction, state pay transparency laws continued to expand — reaching new states and adding new requirements in existing jurisdictions — and pay equity litigation increased as plaintiff firms developed more sophisticated statistical methodologies. ACPWB's multi-state policy engagement capacity, built over the prior decade, proved its value precisely when the center of regulatory gravity moved away from Washington.",
        "The year presented ACPWB with the institutional challenge that defines genuinely enduring advocacy organizations: maintaining technical credibility and member trust through a period of significant policy reversal, without allowing analytical independence to be compromised by the political dynamics of the moment. Our commitment — to evidence-based, nonpartisan, technically rigorous engagement with the agencies and legislative bodies that shape American compensation policy — is unchanged by changes in administration, and it is the source of whatever influence we have earned over the thirty-plus years of our operation.",
    ],
}
