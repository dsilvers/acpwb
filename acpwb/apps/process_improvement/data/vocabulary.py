METHODOLOGIES = [
    'Lean', 'Six Sigma', 'Agile', 'Kaizen', 'DMAIC',
    'BPR', 'PDCA', 'Value Stream', 'TQM', 'Scrum',
]

SUFFIXES = [
    'Optimization', 'Redesign', 'Automation Initiative', 'Streamlining Project',
    'Transformation', 'Enhancement Program', 'Improvement Plan', 'Modernization',
    'Efficiency Study', 'Review', 'Overhaul', 'Implementation',
]

STATUSES = [
    ('proposed',    'Proposed',     30),
    ('in-progress', 'In Progress',  35),
    ('completed',   'Completed',    25),
    ('on-hold',     'On Hold',      10),
]

STATUS_WEIGHTS = [s[2] for s in STATUSES]
STATUS_VALUES  = [(s[0], s[1]) for s in STATUSES]

METRICS = [
    ('Cycle Time',                    'days',   'reduced'),
    ('Error Rate',                    '%',      'reduced'),
    ('FTE Hours per Transaction',     'hours',  'reduced'),
    ('Cost per Transaction',          '$',      'reduced'),
    ('Approval Wait Time',            'days',   'reduced'),
    ('Customer Satisfaction Score',   '/10',    'increased'),
    ('Compliance Rate',               '%',      'increased'),
    ('Rework Rate',                   '%',      'reduced'),
    ('SLA Adherence',                 '%',      'increased'),
    ('Throughput',                    'units/day', 'increased'),
    ('First-Pass Yield',              '%',      'increased'),
    ('On-Time Delivery Rate',         '%',      'increased'),
    ('Vendor Invoice Accuracy',       '%',      'increased'),
    ('Employee Satisfaction (eNPS)',  'pts',    'increased'),
    ('Processing Time',               'hours',  'reduced'),
    ('Exception Rate',                '%',      'reduced'),
    ('Audit Finding Rate',            'per quarter', 'reduced'),
    ('System Downtime',               'hours/month', 'reduced'),
    ('Training Completion Rate',      '%',      'increased'),
    ('Onboarding Time-to-Productivity', 'days', 'reduced'),
]

PHASES = [
    ('Discovery',       'Assess current state, document pain points, identify stakeholders'),
    ('Analysis',        'Root cause analysis, data collection, gap assessment'),
    ('Design',          'Develop future-state process, define success metrics'),
    ('Pilot',           'Test new process with limited scope, gather feedback'),
    ('Implementation',  'Deploy process changes organization-wide'),
    ('Monitoring',      'Track KPIs, address exceptions, sustain gains'),
    ('Closure',         'Final measurement, lessons learned, hand-off to operations'),
]

RISK_LEVELS = [
    ('Low',    'Low probability, low impact. Monitor quarterly.'),
    ('Medium', 'Moderate probability or impact. Develop mitigation plan.'),
    ('High',   'High probability or significant impact. Immediate mitigation required.'),
]

OWNER_FIRST_NAMES = [
    'James', 'Sarah', 'Michael', 'Jennifer', 'Robert', 'Patricia', 'William',
    'Linda', 'David', 'Barbara', 'Richard', 'Susan', 'Thomas', 'Jessica',
    'Charles', 'Karen', 'Christopher', 'Nancy', 'Daniel', 'Lisa',
]

OWNER_LAST_NAMES = [
    'Anderson', 'Hernandez', 'Williams', 'Brown', 'Davis', 'Miller',
    'Wilson', 'Moore', 'Taylor', 'Jackson', 'Martin', 'Lee', 'Thompson',
    'White', 'Harris', 'Sanchez', 'Clark', 'Ramirez', 'Lewis', 'Robinson',
]

OWNER_TITLES = [
    'Director of Operations', 'VP of Process Excellence', 'Senior Manager, Business Transformation',
    'Director of Continuous Improvement', 'Chief Operating Officer', 'VP of Operational Efficiency',
    'Senior Director, Enterprise Operations', 'Manager, Process Engineering', 'Director of Quality',
    'SVP Operations', 'Head of Business Process Management', 'Director, Organizational Effectiveness',
]
