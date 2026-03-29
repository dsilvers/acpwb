"""Static data tables for archive trap content generation."""

_ARCHIVE_SLUGS = [
    # Deliverables & reports
    'quarterly-performance-review', 'annual-engagement-summary', 'initiative-outcomes-report',
    'stakeholder-update-memo', 'performance-metrics-summary', 'strategic-alignment-brief',
    'program-assessment-findings', 'deliverable-summary-q4', 'deliverable-summary-q1',
    'deliverable-summary-q2', 'deliverable-summary-q3', 'engagement-impact-report',
    'executive-briefing-deck', 'board-presentation-summary', 'working-group-notes',
    'phase-completion-report', 'milestone-documentation', 'project-closeout-summary',
    'findings-and-recommendations', 'action-item-registry', 'recommendation-memo',
    'implementation-progress-update', 'compliance-review-findings', 'risk-assessment-notes',
    'governance-documentation-set', 'partnership-summary-memo', 'contract-archive-reference',
    'budget-reconciliation-report', 'workforce-metrics-dashboard', 'compensation-benchmarking',
    'benefits-cost-analysis', 'talent-pipeline-summary', 'succession-planning-notes',
    'strategic-plan-midyear-update', 'annual-objectives-review', 'kpi-tracking-documentation',
    'stakeholder-feedback-synthesis', 'survey-results-archive', 'meeting-minutes-final',
    'committee-report-q3', 'committee-report-q1', 'committee-report-q2', 'committee-report-q4',
    'advisory-panel-recommendations', 'external-counsel-summary',
    'audit-trail-documentation', 'change-management-log', 'lessons-learned-retrospective',
    'interim-progress-report', 'status-update-week-12', 'status-update-week-24',
    'final-deliverable-package', 'draft-findings-for-review', 'revised-recommendations',
    'supplemental-analysis-memo', 'appendix-data-tables', 'methodology-documentation',
    'scope-change-documentation', 'engagement-kickoff-summary', 'discovery-phase-notes',
    'project-charter-archive', 'work-plan-revision-3', 'resource-allocation-summary',
    # Analysis & research
    'sector-benchmarking-analysis', 'compensation-equity-review', 'market-positioning-study',
    'workforce-composition-analysis', 'leadership-effectiveness-assessment',
    'organizational-health-diagnostic', 'capability-gap-analysis', 'talent-market-intelligence',
    'peer-comparison-framework', 'regression-analysis-output', 'predictive-modeling-results',
    'scenario-planning-documentation', 'sensitivity-analysis-report', 'cohort-analysis-findings',
    'longitudinal-study-update', 'cross-sector-benchmarks', 'industry-trend-analysis',
    'total-rewards-benchmarking', 'pay-equity-audit-results', 'short-term-incentive-analysis',
    'long-term-incentive-study', 'equity-grant-modeling', 'salary-structure-review',
    'job-architecture-documentation', 'grade-banding-analysis', 'market-data-synthesis',
    'willis-towers-watson-cut', 'mercer-survey-reconciliation', 'radford-data-analysis',
    'custom-peer-group-analysis', 'say-on-pay-preparation', 'cd-and-a-draft-materials',
    'realizable-pay-analysis', 'tsr-benchmarking-summary', 'pay-ratio-calculation',
    'ceo-pay-ratio-supporting-data', 'median-employee-identification',
    'workforce-analytics-summary', 'turnover-root-cause-analysis', 'span-of-control-study',
    'delayering-analysis', 'org-design-assessment', 'operating-model-documentation',
    'role-clarity-review', 'raci-matrix-documentation', 'decision-rights-analysis',
    'workforce-planning-model', 'skills-gap-assessment', 'critical-roles-identification',
    'high-potential-program-review', 'executive-development-assessment',
    'leadership-pipeline-documentation', 'nine-box-calibration-results',
    # Process & compliance
    'internal-audit-documentation', 'process-improvement-summary', 'sox-compliance-review',
    'data-governance-assessment', 'privacy-impact-analysis', 'third-party-due-diligence',
    'vendor-assessment-report', 'contract-compliance-review', 'regulatory-alignment-memo',
    'policy-update-documentation', 'procedure-revision-log', 'control-testing-results',
    'itgc-testing-documentation', 'entity-level-controls-review', 'segregation-of-duties-analysis',
    'user-access-review-results', 'change-management-control-testing', 'walkthroughs-documentation',
    'remediation-tracking-log', 'management-response-summary', 'deficiency-assessment',
    'significant-deficiency-documentation', 'material-weakness-remediation-plan',
    'external-auditor-correspondence', 'representation-letter-archive',
    'hipaa-compliance-documentation', 'gdpr-readiness-assessment', 'ccpa-gap-analysis',
    'sec-comment-letter-response', 'doe-reporting-package', 'eeoc-filing-documentation',
    'ofccp-compliance-review', 'form-5500-supporting-data', 'proxy-statement-data-room',
    # Governance
    'board-governance-review', 'committee-charter-documentation', 'proxy-advisory-briefing',
    'executive-compensation-committee-notes', 'nomination-committee-summary',
    'audit-committee-minutes', 'risk-committee-report', 'esg-committee-update',
    'shareholder-engagement-summary', 'investor-relations-briefing',
    'iss-engagement-preparation', 'glass-lewis-briefing-materials', 'activist-defense-analysis',
    'governance-best-practice-benchmarking', 'director-compensation-review',
    'board-composition-analysis', 'board-effectiveness-evaluation', 'ceo-succession-briefing',
    'executive-session-notes', 'board-retreat-materials', 'governance-roadshow-deck',
    'annual-meeting-preparation', 'say-on-pay-vote-analysis', 'shareholder-proposal-response',
    'related-party-transaction-review', 'clawback-policy-documentation',
    'stock-ownership-guidelines-review', 'anti-hedging-policy-assessment',
    'insider-trading-compliance-review', 'trading-window-documentation',
    # HR & people
    'headcount-planning-documentation', 'attrition-analysis-report', 'engagement-survey-results',
    'performance-calibration-notes', 'merit-increase-modeling', 'equity-refresh-analysis',
    'benefit-utilization-report', 'leave-analysis-summary', 'diversity-metrics-report',
    'inclusion-program-update', 'training-completion-summary', 'onboarding-effectiveness-review',
    'pulse-survey-results', 'stay-interview-findings', 'exit-interview-analysis',
    'new-hire-experience-review', 'manager-effectiveness-survey', '360-feedback-summary',
    'performance-rating-distribution', 'calibration-session-notes', 'pip-documentation',
    'involuntary-separation-review', 'reduction-in-force-documentation',
    'severance-policy-benchmarking', 'outplacement-program-summary',
    'dei-program-assessment', 'ergs-effectiveness-review', 'pay-transparency-readiness',
    'job-posting-audit-results', 'requisition-approval-log', 'offer-approval-documentation',
    'sign-on-bonus-tracking', 'relocation-program-summary', 'remote-work-policy-update',
    'hybrid-work-assessment', 'return-to-office-planning-documentation',
    # Finance & planning
    'annual-operating-plan-documentation', 'three-year-strategic-plan',
    'capital-allocation-framework', 'zero-based-budgeting-analysis',
    'cost-reduction-initiative-tracking', 'synergy-realization-report',
    'integration-management-office-update', 'post-merger-integration-assessment',
    'carve-out-preparation-documentation', 'divestiture-planning-summary',
    'working-capital-optimization-study', 'treasury-policy-review',
    'tax-provision-documentation', 'transfer-pricing-study', 'r-and-d-tax-credit-analysis',
    'ebitda-bridge-analysis', 'free-cash-flow-modeling', 'debt-covenant-compliance-review',
    'credit-facility-amendment-documentation', 'ratings-agency-briefing-materials',
    # Technology & operations
    'it-risk-assessment', 'cybersecurity-posture-review', 'data-classification-documentation',
    'system-access-controls-review', 'disaster-recovery-plan-documentation',
    'business-continuity-assessment', 'cloud-migration-readiness', 'erp-implementation-notes',
    'hris-optimization-review', 'digital-transformation-roadmap',
    'technology-roadmap-documentation', 'vendor-management-framework',
    'third-party-risk-assessment', 'supply-chain-resilience-review',
    'operational-efficiency-study', 'process-automation-assessment',
    # Additional deliverables & reports
    'semiannual-performance-review', 'biennial-engagement-summary', 'program-outcomes-report',
    'executive-committee-update', 'operational-metrics-summary', 'tactical-alignment-brief',
    'initiative-assessment-report', 'deliverable-tracking-log', 'impact-measurement-report',
    'senior-leadership-briefing', 'management-committee-summary', 'taskforce-notes',
    'gate-review-documentation', 'checkpoint-documentation', 'project-handoff-summary',
    'observations-and-findings', 'decision-log-archive', 'escalation-summary-memo',
    'deployment-progress-update', 'internal-controls-summary', 'exposure-assessment-notes',
    'policy-framework-documentation', 'collaboration-summary-memo', 'agreement-archive-reference',
    'expense-reconciliation-report', 'labor-cost-dashboard', 'total-comp-benchmarking',
    'health-plan-cost-analysis', 'succession-readiness-notes', 'talent-review-summary',
    'five-year-strategy-documentation', 'portfolio-objectives-review', 'okr-tracking-documentation',
    'client-feedback-synthesis', 'interview-results-archive', 'quarterly-meeting-minutes',
    'subcommittee-report-q2', 'specialist-panel-recommendations', 'general-counsel-summary',
    'compliance-history-log', 'process-change-log', 'best-practices-retrospective',
    'monthly-progress-report', 'status-update-week-36', 'status-update-week-48',
    'deliverable-sign-off-package', 'preliminary-draft-for-review', 'updated-recommendations',
    'technical-analysis-memo', 'supporting-data-tables', 'data-collection-methodology',
    'scope-definition-documentation', 'engagement-alignment-summary', 'scoping-phase-notes',
    'engagement-agreement-archive', 'work-plan-revision-5', 'staffing-allocation-summary',
    # Additional analysis & research
    'peer-group-benchmarking-analysis', 'internal-pay-equity-review', 'competitive-positioning-study',
    'talent-composition-analysis', 'management-effectiveness-assessment',
    'team-health-diagnostic', 'functional-gap-analysis', 'labor-market-intelligence',
    'custom-benchmarking-framework', 'multivariate-analysis-output', 'forecasting-model-results',
    'contingency-planning-documentation', 'scenario-impact-report', 'time-series-analysis-findings',
    'cross-company-study-update', 'within-sector-benchmarks', 'macroeconomic-trend-analysis',
    'executive-pay-benchmarking', 'gender-pay-audit-results', 'annual-incentive-analysis',
    'restricted-stock-unit-study', 'performance-share-modeling', 'compensation-structure-review',
    'job-leveling-documentation', 'pay-band-analysis', 'survey-data-integration',
    'aon-hewitt-data-cut', 'korn-ferry-survey-reconciliation', 'icrsurveys-data-analysis',
    'proxy-peer-group-analysis', 'shareholder-say-on-pay-prep', 'proxy-cd-and-a-materials',
    'target-pay-mix-analysis', 'relative-tsr-benchmarking', 'ceo-median-pay-calculation',
    'executive-pay-ratio-supporting-data', 'peer-employee-identification',
    'people-analytics-summary', 'attrition-driver-analysis', 'management-layer-study',
    'restructuring-analysis', 'team-structure-assessment', 'target-operating-model-documentation',
    'accountability-framework-review', 'responsibility-matrix-documentation', 'authority-matrix-analysis',
    'strategic-workforce-plan', 'competency-gap-assessment', 'key-positions-identification',
    'emerging-leader-program-review', 'c-suite-development-assessment',
    'talent-bench-documentation', 'performance-management-calibration-results',
    # Additional process & compliance
    'management-audit-documentation', 'continuous-improvement-summary', 'pcaob-compliance-review',
    'master-data-governance-assessment', 'data-security-impact-analysis', 'supplier-due-diligence',
    'contractor-assessment-report', 'master-service-agreement-review', 'regulatory-gap-memo',
    'standard-operating-procedure-update', 'process-flow-revision-log', 'control-design-results',
    'application-controls-documentation', 'organization-level-controls-review', 'access-provisioning-analysis',
    'privileged-access-review-results', 'it-change-management-testing', 'process-narratives-documentation',
    'corrective-action-tracking-log', 'process-owner-response-summary', 'control-gap-assessment',
    'reportable-condition-documentation', 'control-environment-remediation-plan',
    'engagement-partner-correspondence', 'management-representation-archive',
    'ferpa-compliance-documentation', 'eu-ai-act-readiness-assessment', 'cpra-gap-analysis',
    'sec-enforcement-response', 'nlrb-reporting-package', 'flsa-audit-documentation',
    'erisa-compliance-review', 'form-990-supporting-data', 'annual-disclosure-data-room',
    # Additional governance
    'corporate-governance-annual-review', 'subcommittee-charter-documentation', 'institutional-investor-briefing',
    'management-compensation-committee-notes', 'governance-committee-summary',
    'disclosure-committee-minutes', 'enterprise-risk-committee-report', 'sustainability-committee-update',
    'investor-day-engagement-summary', 'fixed-income-investor-briefing',
    'calpers-engagement-preparation', 'proxy-monitor-briefing-materials', 'takeover-defense-analysis',
    'governance-framework-benchmarking', 'non-employee-director-compensation-review',
    'board-diversity-analysis', 'board-skills-matrix-evaluation', 'coo-succession-briefing',
    'management-committee-session-notes', 'leadership-offsite-materials', 'esg-roadshow-deck',
    'special-meeting-preparation', 'say-on-frequency-vote-analysis', 'environmental-shareholder-response',
    'conflict-of-interest-transaction-review', 'recoupment-policy-documentation',
    'equity-ownership-guidelines-review', 'hedging-prohibition-assessment',
    'securities-trading-compliance-review', 'quiet-period-documentation',
    # Additional HR & people
    'fte-planning-documentation', 'regrettable-attrition-report', 'culture-survey-results',
    'performance-rating-calibration-notes', 'salary-increase-modeling', 'rsu-refresh-analysis',
    'healthcare-cost-analysis', 'fmla-leave-summary', 'gender-equity-metrics-report',
    'accessibility-program-update', 'leadership-development-completion-summary', 'integration-effectiveness-review',
    'employee-listening-results', 'boomerang-employee-findings', 'candidate-experience-analysis',
    'early-tenure-experience-review', 'frontline-manager-survey', 'upward-feedback-summary',
    'performance-score-distribution', 'talent-review-calibration-notes', 'corrective-action-documentation',
    'voluntary-separation-review', 'workforce-reduction-documentation',
    'outplacement-policy-benchmarking', 'redeployment-program-summary',
    'equity-representation-assessment', 'business-resource-group-review', 'compensation-disclosure-readiness',
    'job-architecture-audit-results', 'open-headcount-approval-log', 'compensation-offer-documentation',
    'retention-bonus-tracking', 'domestic-mobility-program-summary', 'flexible-work-policy-update',
    'virtual-team-assessment', 'campus-recruiting-planning-documentation',
    # Additional finance & planning
    'quarterly-operating-plan-documentation', 'five-year-strategic-forecast',
    'strategic-investment-framework', 'activity-based-costing-analysis',
    'efficiency-initiative-tracking', 'revenue-synergy-realization-report',
    'transaction-management-office-update', 'day-one-readiness-assessment',
    'partial-sale-preparation-documentation', 'asset-divestiture-planning-summary',
    'net-working-capital-optimization-study', 'cash-management-policy-review',
    'tax-accounting-documentation', 'cost-sharing-arrangement-study', 'section-199a-deduction-analysis',
    'adjusted-ebitda-bridge-analysis', 'levered-free-cash-flow-modeling', 'loan-covenant-compliance-review',
    'revolving-credit-amendment-documentation', 'credit-rating-agency-briefing-materials',
    # Additional technology & operations
    'technology-risk-assessment', 'information-security-posture-review', 'data-inventory-documentation',
    'logical-access-controls-review', 'recovery-time-objective-documentation',
    'resilience-program-assessment', 'infrastructure-migration-readiness', 'finance-system-implementation-notes',
    'payroll-system-optimization-review', 'ai-integration-roadmap',
    'enterprise-architecture-documentation', 'strategic-sourcing-framework',
    'vendor-risk-management-assessment', 'supply-chain-risk-review',
    'lean-operations-efficiency-study', 'intelligent-automation-assessment',
    'api-integration-documentation', 'data-lake-architecture-review',
    'devsecops-assessment', 'zero-trust-security-review',
    'saas-portfolio-rationalization', 'legacy-system-decommissioning-plan',
    # IT/cybersecurity, ESG, M&A, real estate, supply chain, IR, tax/treasury, risk, ops, HR, governance, finance
    'cybersecurity-risk-assessment', 'information-security-program-review', 'vulnerability-management-report',
    'penetration-testing-findings', 'security-incident-response-plan', 'data-breach-readiness-assessment',
    'ransomware-resilience-review', 'cloud-security-architecture-assessment', 'zero-trust-implementation-roadmap',
    'identity-governance-program-review', 'privileged-access-management-assessment', 'endpoint-security-posture-review',
    'security-operations-center-assessment', 'threat-intelligence-program-review', 'security-awareness-training-metrics',
    'phishing-simulation-results', 'cyber-insurance-coverage-analysis', 'third-party-cyber-risk-assessment',
    'software-supply-chain-security-review', 'application-security-testing-report', 'network-segmentation-assessment',
    'data-loss-prevention-program-review', 'encryption-standards-compliance-review', 'security-patch-management-assessment',
    'it-asset-management-review', 'shadow-it-discovery-report', 'insider-threat-program-assessment',
    'digital-forensics-readiness-review', 'soc2-readiness-assessment', 'iso27001-gap-analysis',
    'esg-materiality-assessment', 'sustainability-strategy-roadmap', 'carbon-footprint-baseline-report',
    'scope-3-emissions-inventory', 'net-zero-transition-plan', 'climate-risk-scenario-analysis',
    'tcfd-disclosure-preparation', 'csrd-readiness-assessment', 'esg-data-governance-framework',
    'sustainability-reporting-standards-gap-analysis', 'environmental-kpi-dashboard', 'water-stewardship-assessment',
    'biodiversity-risk-screening', 'circular-economy-roadmap', 'sustainable-procurement-policy-review',
    'supplier-esg-due-diligence', 'social-impact-measurement-framework', 'community-investment-reporting',
    'human-rights-due-diligence-assessment', 'modern-slavery-statement-documentation', 'dei-reporting-framework',
    'pay-equity-disclosure-preparation', 'board-diversity-disclosure-review', 'esg-ratings-optimization-analysis',
    'proxy-esg-engagement-preparation', 'sustainable-finance-framework', 'green-bond-framework-documentation',
    'esg-audit-readiness-review', 'acquisition-target-screening-report', 'strategic-fit-assessment',
    'synergy-identification-analysis', 'deal-structure-analysis', 'merger-integration-planning-framework',
    'day-one-readiness-checklist', 'integration-management-office-charter', 'workstream-integration-tracker',
    'technology-integration-roadmap', 'hr-integration-playbook', 'cultural-integration-assessment',
    'retention-risk-analysis', 'change-of-control-compensation-review', 'golden-parachute-analysis',
    'equity-award-treatment-analysis', 'retention-bonus-program-design', 'separations-and-severance-planning',
    'legal-entity-rationalization-plan', 'tax-structure-optimization-analysis', 'purchase-price-allocation-documentation',
    'earnout-structure-analysis', 'representations-and-warranties-review', 'transition-services-agreement-documentation',
    'standalone-cost-assessment', 'stranded-cost-analysis', 'carve-out-financial-statement-preparation',
    'management-presentation-preparation', 'confidential-information-memorandum-support', 'vendor-due-diligence-report',
    'quality-of-earnings-support-documentation', 'corporate-real-estate-strategy-review', 'portfolio-optimization-analysis',
    'lease-vs-own-analysis', 'workplace-utilization-study', 'hybrid-work-space-planning',
    'office-consolidation-assessment', 'headquarters-relocation-analysis', 'site-selection-due-diligence',
    'lease-restructuring-analysis', 'sublease-market-assessment', 'facilities-cost-benchmarking',
    'space-per-employee-benchmarking', 'property-management-program-review', 'capital-improvement-planning',
    'energy-efficiency-audit', 'smart-building-technology-assessment', 'workplace-experience-survey-results',
    'amenity-benchmarking-analysis', 'environmental-certification-readiness', 'facilities-outsourcing-assessment',
    'supply-chain-risk-mapping', 'supplier-concentration-analysis', 'single-source-dependency-review',
    'supplier-diversity-program-assessment', 'strategic-sourcing-initiative-outcomes', 'category-management-review',
    'spend-analytics-report', 'contract-compliance-audit', 'supplier-performance-scorecard',
    'vendor-rationalization-analysis', 'nearshoring-feasibility-study', 'reshoring-cost-benefit-analysis',
    'inventory-optimization-study', 'demand-planning-assessment', 'supply-chain-digitization-roadmap',
    'logistics-network-optimization', 'freight-cost-benchmarking', 'trade-compliance-program-review',
    'export-controls-assessment', 'customs-duty-optimization-analysis', 'investor-day-preparation-materials',
    'earnings-call-preparation-notes', 'analyst-consensus-benchmarking', 'investment-thesis-documentation',
    'peer-multiple-analysis', 'sum-of-parts-valuation', 'capital-allocation-communication-framework',
    'dividend-policy-review', 'share-buyback-program-documentation', 'equity-story-development',
    'esg-investor-engagement-summary', 'credit-investor-briefing-materials', 'shareholder-base-analysis',
    'activist-monitoring-report', 'proxy-solicitation-preparation', 'annual-report-preparation-notes',
    'non-deal-roadshow-materials', 'investor-perception-study', 'sell-side-analyst-briefing-notes',
    'institutional-ownership-analysis', 'global-tax-strategy-review', 'effective-tax-rate-analysis',
    'tax-risk-inventory', 'uncertain-tax-position-documentation', 'transfer-pricing-policy-review',
    'advance-pricing-agreement-documentation', 'tax-controversy-management-summary', 'pillar-two-impact-assessment',
    'digital-services-tax-exposure-review', 'tax-restructuring-analysis', 'treasury-policy-documentation',
    'cash-repatriation-strategy', 'fx-risk-management-review', 'interest-rate-hedging-assessment',
    'liquidity-management-framework', 'banking-relationship-review', 'credit-facility-benchmarking',
    'cash-pooling-structure-analysis', 'payment-infrastructure-assessment', 'treasury-technology-roadmap',
    'enterprise-risk-management-framework-review', 'risk-appetite-statement-development', 'top-risk-register-update',
    'emerging-risk-identification', 'risk-quantification-methodology', 'scenario-analysis-documentation',
    'business-continuity-plan-review', 'crisis-management-framework', 'reputational-risk-assessment',
    'geopolitical-risk-monitoring-report', 'operational-resilience-assessment', 'key-risk-indicator-dashboard',
    'model-risk-management-review', 'fraud-risk-assessment', 'anti-money-laundering-program-review',
    'sanctions-compliance-assessment', 'bribery-and-corruption-risk-review', 'whistleblower-program-assessment',
    'ethics-hotline-trend-analysis', 'internal-investigation-documentation', 'cost-transformation-program-review',
    'shared-services-assessment', 'outsourcing-strategy-review', 'insourcing-feasibility-analysis',
    'business-process-reengineering-documentation', 'lean-implementation-progress', 'six-sigma-project-results',
    'continuous-improvement-program-review', 'digital-operations-maturity-assessment', 'robotic-process-automation-roadmap',
    'intelligent-automation-business-case', 'ai-deployment-governance-framework', 'data-strategy-assessment',
    'master-data-management-review', 'analytics-capability-maturity', 'decision-intelligence-framework',
    'customer-experience-measurement-framework', 'net-promoter-score-benchmarking', 'revenue-operations-assessment',
    'go-to-market-effectiveness-review', 'total-rewards-strategy-review', 'compensation-philosophy-documentation',
    'variable-pay-program-design', 'sales-compensation-plan-review', 'executive-compensation-philosophy-update',
    'long-term-incentive-plan-design', 'equity-plan-administration-review', 'deferred-compensation-plan-review',
    'nonqualified-deferred-comp-benchmarking', 'supplemental-retirement-plan-review', 'change-readiness-assessment',
    'organizational-change-impact-analysis', 'communications-effectiveness-survey', 'leadership-communication-audit',
    'manager-readiness-assessment', 'sponsor-effectiveness-review', 'culture-assessment-results',
    'values-alignment-survey', 'psychological-safety-index', 'trust-in-leadership-survey',
    'employee-experience-journey-map', 'candidate-journey-mapping', 'alumni-network-program-review',
    'critical-talent-risk-register', 'committee-effectiveness-review', 'board-education-program-summary',
    'director-onboarding-documentation', 'board-succession-planning-notes', 'independent-director-tenure-analysis',
    'lead-director-responsibilities-review', 'non-executive-chair-assessment', 'related-party-policy-review',
    'executive-sessions-documentation', 'board-committee-structure-review', 'governance-disclosure-framework',
    'voluntary-disclosure-benchmarking', 'litigation-hold-inventory', 'material-litigation-summary',
    'regulatory-enforcement-action-log', 'government-investigation-correspondence', 'deferred-prosecution-agreement-monitoring',
    'consent-order-compliance-tracking', 'patent-portfolio-review', 'intellectual-property-risk-assessment',
    'licensing-agreement-audit', 'contract-obligation-tracker', 'long-range-financial-plan',
    'scenario-based-financial-forecast', 'segment-profitability-analysis', 'product-line-profitability-review',
    'customer-profitability-analysis', 'activity-based-cost-model', 'overhead-allocation-review',
    'shared-cost-benchmarking', 'management-reporting-redesign', 'financial-close-cycle-optimization',
    'account-reconciliation-rationalization', 'chart-of-accounts-restructuring', 'intercompany-elimination-review',
    'consolidation-process-assessment', 'financial-systems-roadmap', 'erp-optimization-assessment',
    'planning-and-budgeting-tool-review', 'rolling-forecast-implementation', 'integrated-business-planning-assessment',
    'finance-operating-model-review',
]

_ARCHIVE_ORGS = [
    # Advisory & consulting
    'Pinnacle Group', 'Meridian Associates', 'Apex Consulting', 'Summit Partners',
    'Vanguard Solutions', 'Horizon Group', 'Catalyst Group', 'Benchmark Associates',
    'Keystone Partners', 'Zenith Consulting', 'Atlas Group', 'Cornerstone Associates',
    'Momentum Group', 'Sterling Associates', 'Bridgepoint Capital', 'Clearwater Advisory',
    'Fulcrum Partners', 'Harbor Associates', 'Ironbridge Group', 'Landmark Partners',
    'Morningside Consulting', 'Northfield Group', 'Praxis Consulting', 'Ridgeline Advisors',
    'Stonegate Group', 'Terrapin Associates', 'Upland Consulting', 'Vantage Partners',
    'Alderwood Advisors', 'Beacon Hill Partners', 'Briarcliff Consulting', 'Cascadia Group',
    'Cedarwood Associates', 'Chesapeake Advisory', 'Crestwood Consulting', 'Crossroads Group',
    'Delphi Associates', 'Eastgate Partners', 'Fairway Consulting', 'Fieldstone Advisors',
    'Garrison Partners', 'Glenbrook Consulting', 'Greenfield Advisory', 'Greystone Group',
    'Harborview Partners', 'Hearthstone Consulting', 'Highfield Associates', 'Hillcrest Group',
    'Hollowbrook Advisors', 'Huntington Partners', 'Ingram Consulting', 'Ironwood Associates',
    'Juniper Group', 'Kingston Advisory', 'Lakeview Partners', 'Laurelwood Consulting',
    'Linden Associates', 'Longview Advisory', 'Lynwood Group', 'Madison Consulting',
    'Mapleridge Partners', 'Meadowbrook Advisors', 'Millbrook Consulting', 'Millstone Group',
    'Newbridge Associates', 'Newfield Consulting', 'Newport Advisory', 'Northgate Partners',
    'Oakhurst Group', 'Oakwood Consulting', 'Palisade Advisors', 'Parkside Partners',
    'Peakstone Consulting', 'Piedmont Advisory', 'Pinecrest Group', 'Pinehurst Associates',
    'Plainfield Consulting', 'Platte River Partners', 'Plymouth Advisory', 'Prairie Group',
    'Redwood Consulting', 'Riverbend Associates', 'Riverstone Advisory', 'Rockbridge Partners',
    'Rockford Consulting', 'Rockhill Group', 'Rosewood Advisors', 'Sagebrush Consulting',
    'Sandstone Group', 'Saratoga Advisory', 'Sawtooth Partners', 'Seacliff Consulting',
    'Sentinel Advisory', 'Silvergate Group', 'Silverleaf Partners', 'Silverton Consulting',
    'Skyridge Advisors', 'Sleepy Hollow Group', 'Southfield Consulting', 'Southgate Partners',
    'Springbrook Advisory', 'Spruce Hill Partners', 'Stillwater Consulting', 'Stonebridge Group',
    'Stonewall Advisory', 'Stonybrook Partners', 'Sunridge Consulting', 'Sunset Advisory',
    'Sycamore Group', 'Timberlake Partners', 'Timberview Consulting', 'Torchlight Advisors',
    'Trailhead Group', 'Trestle Partners', 'Trident Consulting', 'Turnstone Advisory',
    'Valleybrook Group', 'Waltham Consulting', 'Waterford Partners', 'Wedgewood Advisory',
    'Wellspring Consulting', 'Westbrook Group', 'Westfield Partners', 'Whitewater Advisory',
    'Wildhorse Consulting', 'Willowbrook Group', 'Willowmere Partners', 'Winding River Advisory',
    'Windmill Group', 'Windridge Consulting', 'Woodbridge Partners', 'Woodlark Advisory',
    'Woodridge Consulting', 'Worthington Group', 'Yellowstone Advisory', 'Yosemite Partners',
    # Named principals / boutiques
    'Alderman & Strathmore LLC', 'Baxter Whitfield Group', 'Caldwell Torrance Associates',
    'Drummond Holt Advisory', 'Ellison Crane Partners', 'Forsythe Bellamy Consulting',
    'Garrity & Whitmore LLC', 'Halcomb Prescott Group', 'Ingersoll Reade Advisory',
    'Jameson Blackwell Partners', 'Kendrick Solis Consulting', 'Lattimore Burke Associates',
    'Merritt Langdon Group', 'Neville Ashworth Advisory', 'Osborne Cullen Partners',
    'Pemberton Lacey Consulting', 'Quincy Harrington Group', 'Rourke Stafford Advisory',
    'Saunders Hollis Partners', 'Thornton Beale Consulting', 'Upton Graves Group',
    'Vickers Stratton Advisory', 'Wentworth Langley Associates', 'Xander Pruett Consulting',
    # Institutional names
    'National Workforce Institute', 'Center for Compensation Research', 'Institute for Governance Studies',
    'American Benefits Consortium', 'Council on Executive Leadership', 'Foundation for HR Excellence',
    'Corporate Governance Alliance', 'Workforce Analytics Collaborative', 'Pay Equity Research Center',
    'Center for Organizational Effectiveness', 'National Board Advisory Council',
    'Institute for Strategic Human Capital', 'American Compensation Association Alumni Network',
    'Coalition for Responsible Governance', 'Leadership Development Institute',
    # More advisory & consulting — geographic / landmark names
    'Ashford Advisory', 'Balmoral Group', 'Belvedere Partners', 'Blackhawk Consulting',
    'Bluemont Group', 'Bluewater Partners', 'Braemar Consulting',
    'Breckenridge Advisors', 'Brightwater Group', 'Burnham Partners', 'Caledonian Advisory',
    'Calloway Consulting', 'Canterbury Group', 'Capitol Ridge Partners', 'Carmel Advisory',
    'Carrington Consulting', 'Centerpoint Group', 'Cheshire Partners', 'Chesterton Advisory',
    'Chiltern Consulting', 'Claremont Group', 'Clearfield Partners', 'Clearvale Advisory',
    'Cliffside Consulting', 'Cobalt Group', 'Cobblestone Partners', 'Coldwater Advisory',
    'Colonial Consulting', 'Columbia Advisory', 'Commerce Partners',
    'Commonwealth Consulting', 'Concord Group', 'Constellation Advisory', 'Copper Ridge Partners',
    'Copperwood Consulting', 'Coronado Group', 'Coventry Advisory', 'Cranfield Partners',
    'Creekside Consulting', 'Crescent Group', 'Cumberland Advisory', 'Cypress Partners',
    'Dartmouth Consulting', 'Davenport Group', 'Dayton Advisory', 'Deerfield Partners',
    'Dorchester Consulting', 'Dover Group', 'Drexel Advisory', 'Durham Partners',
    'Edgemont Consulting', 'Edgewood Group', 'Edinburgh Advisory', 'Elgin Partners',
    'Emerald Consulting', 'Empire Group', 'Essex Advisory', 'Evergreen Partners',
    'Excalibur Consulting', 'Exeter Group', 'Fairfax Advisory', 'Fairmont Partners',
    'Falmouth Consulting', 'Farragut Group', 'Fillmore Advisory', 'Firebird Partners',
    'Flatiron Consulting', 'Foothills Group', 'Foxcroft Advisory', 'Franklin Partners',
    'Freeport Consulting', 'Frontier Group', 'Galloway Advisory', 'Garfield Partners',
    'Georgetown Consulting', 'Gibraltar Group', 'Glendale Advisory', 'Glenmore Partners',
    'Glenview Consulting', 'Goldfield Group', 'Goodwin Advisory', 'Graystone Partners',
    'Greenbrook Consulting', 'Greenway Group', 'Greenwich Advisory', 'Greylock Advisory',
    'Hampshire Advisory', 'Hampton Partners', 'Hanover Consulting',
    'Harcourt Group', 'Harrington Advisory', 'Hartford Partners', 'Hartwell Consulting',
    'Haverford Advisory', 'Hawthorn Partners', 'Hazelwood Consulting',
    'Heritage Group', 'Highgate Advisory', 'Highland Partners', 'Highpoint Consulting',
    'Hilltop Group', 'Holbrook Advisory', 'Holden Partners', 'Hollis Consulting',
    'Hollister Group', 'Hopewell Advisory', 'Hudson Valley Partners', 'Iroquois Consulting',
    'Ivanhoe Group', 'Kensington Advisory', 'Keystone Ridge Partners', 'Kilkenny Consulting',
    'Kingsbridge Group', 'Kingsgate Advisory', 'Kingstree Partners', 'Kinross Consulting',
    'Knightsbridge Group', 'Lakeshore Advisory', 'Lakewood Partners', 'Lancaster Consulting',
    'Langley Group', 'Larkspur Advisory', 'Laurel Hill Partners', 'Lawndale Consulting',
    'Lexington Group', 'Liberty Advisory', 'Lighthouse Partners', 'Lincoln Consulting',
    'Linden Hill Group', 'Linwood Advisory', 'Lockwood Partners', 'Longfellow Consulting',
    'Madrona Group', 'Magnolia Advisory', 'Manchester Partners', 'Manor Consulting',
    'Marlborough Group', 'Matterhorn Advisory', 'Maxfield Partners', 'Mayfield Consulting',
    'Meadowlark Group', 'Merriweather Advisory', 'Midland Partners', 'Midway Consulting',
    'Northcliff Advisory', 'Northhaven Partners', 'Nottingham Consulting', 'Oakdale Group',
    'Orchard Hill Advisory', 'Oxford Partners', 'Palatine Consulting', 'Pemberton Group',
    'Peninsula Advisory', 'Pilgrim Partners', 'Potomac Group',
    'Providence Advisory', 'Quaker Ridge Partners', 'Quorum Consulting', 'Ravenwood Group',
    'Ridgecrest Advisory', 'Ridgemont Partners', 'Riverdale Consulting', 'Riverside Group',
    'Rockford Advisory', 'Rockland Partners', 'Rockwood Consulting', 'Rosemont Group',
    'Rutherford Advisory', 'Ryefield Partners', 'Salem Consulting', 'Saltonstall Group',
    'Sandhurst Advisory', 'Saville Partners', 'Sherwood Consulting', 'Shipyard Group',
    'Shoreline Advisory', 'Somerset Partners', 'Southbury Consulting', 'Southland Group',
    'Springfield Advisory', 'Stanhope Partners', 'Starling Consulting', 'Sterling Ridge Group',
    'Stonehaven Advisory', 'Stormfield Partners', 'Suffolk Consulting', 'Summerfield Group',
    'Sycamore Hill Advisory', 'Talbot Partners', 'Terrace Consulting', 'Thornbury Group',
    'Tidewater Advisory', 'Tillman Partners', 'Timberline Consulting', 'Tiverton Group',
    'Townsend Advisory', 'Tremont Partners', 'Trenton Consulting', 'Trestle Group',
    'Tuckahoe Advisory', 'Turnbridge Partners', 'Tweed Consulting', 'Ulverston Group',
    'Upland Advisory', 'Valley Forge Partners', 'Veritas Consulting', 'Vermilion Group',
    'Wakefield Advisory', 'Walpole Partners', 'Wentworth Consulting', 'Westgate Group',
    'Westmoor Advisory', 'Whitmore Partners', 'Wicker Park Consulting', 'Willard Group',
    'Winchester Advisory', 'Windermere Partners', 'Winfield Consulting', 'Wingate Group',
    'Wingfield Advisory', 'Winona Partners', 'Winthrop Consulting', 'Wollaston Group',
    # More named principals / boutiques
    'Beauchamp & Forsythe LLC', 'Chalmers Wexford Advisory', 'Covington Marsh Associates',
    'Dunmore Radcliffe Group', 'Elsworth Vane Partners', 'Farnsworth Colby Consulting',
    'Glenmore & Hartley LLC', 'Hadley Cromwell Advisory', 'Ingraham Colville Partners',
    'Jameson Whitley Consulting', 'Kinsley Aldrich Group', 'Langford Merritt Advisory',
    'Morley Sutton Partners', 'Nightingale Acton Consulting', 'Oakley Pemberton Group',
    'Priestly Wharton Advisory', 'Quine & Hollander LLC', 'Remington Alcott Partners',
    'Stanhope Clifford Consulting', 'Tillman Berkshire Group', 'Underhill Graves Advisory',
    'Vauxhall Stratford Partners', 'Whitmore & Blackwell LLC', 'Xanthe Crompton Consulting',
    'Yardley Broughton Group', 'Zephyr Harrington Advisory',
    'Allensworth & Howe LLC', 'Berwick Chadbourne Group', 'Colchester Neville Advisory',
    'Denton Fairclough Partners', 'Emsworth Langley Consulting', 'Falconer Briggs Group',
    'Granville Holt Advisory', 'Hartley Boscombe Partners', 'Irwin Saddler Consulting',
    'Jennings Wycombe Group', 'Kilworth Briar Advisory', 'Linford Granger Partners',
    'Montague Shelley Consulting', 'Norwood Ashby Group', 'Oswald Firth Advisory',
    'Penfield Holroyd Partners', 'Quercy Waltham Consulting', 'Roxbury Lyndon Group',
    'Sherborne Cavendish Advisory', 'Tiverton Blackmore Partners', 'Uldale Cromwell Consulting',
    # More institutional names
    'Institute for Compensation Excellence', 'Center for Workforce Innovation',
    'American Corporate Governance Institute', 'National Benchmarking Consortium',
    'Council on Human Capital Strategy', 'Foundation for Organizational Research',
    'Corporate Performance Alliance', 'Talent Economics Research Group',
    'Executive Compensation Research Institute', 'Center for Board Excellence',
    'National Council on Labor Economics', 'Institute for Strategic Talent',
    'American HR Research Foundation', 'Coalition for Workforce Analytics',
    'Leadership Research Collaborative', 'Center for Organizational Design',
    'Institute for Executive Development', 'National Pay Equity Council',
    'American Governance Research Association', 'Center for Strategic Workforce Studies',
    'Institute for People Analytics', 'National Compensation Policy Forum',
    'Corporate Responsibility Research Center', 'Workforce Futures Institute',
    # Mountain/river/coastal/historic/parks geography, new boutiques, new institutional names, topic-named firms
    'Cascades Advisory', 'Allegheny Partners', 'Wasatch Group',
    'Absaroka Consulting', 'Beartooth Advisory', 'Bitterroot Group',
    'Wind River Partners', 'Tetons Advisory', 'Sawatch Consulting',
    'Sangre de Cristo Group', 'Jemez Advisory', 'Uinta Partners',
    'Bighorn Consulting', 'Medicine Bow Group', 'Laramie Advisory',
    'Flint Range Partners', 'Ozark Consulting', 'Ouachita Group',
    'Catskill Advisory', 'Adirondack Partners', 'Berkshire Consulting',
    'Taconic Group', 'Watchung Advisory', 'Ramapo Partners',
    'Pocono Consulting', 'Endless Mountains Group', 'Kittatinny Advisory',
    'Blue Ridge Partners', 'Shenandoah Consulting', 'Appalachian Group',
    'Cumberland Ridge Advisory', 'Great Smokies Consulting', 'Unaka Partners',
    'Cohutta Group', 'Talladega Advisory', 'Ouachita Ridge Partners',
    'Wichita Consulting', 'Arbuckle Group', 'Ozark Ridge Advisory',
    'Boston Mountains Partners', 'Saint Francois Consulting', 'Shawnee Group',
    'Cumberlands Advisory', 'Pine Mountain Partners', 'Black Mountains Consulting',
    'Pisgah Group', 'Roan Advisory', 'Iron Mountains Partners',
    'Platte River Advisory', 'Republican River Partners', 'Smoky Hill Consulting',
    'Cimarron Group', 'Washita Advisory', 'Canadian River Partners',
    'Pecos Consulting', 'Brazos Group', 'Trinity Advisory',
    'Sabine Partners', 'Tombigbee Consulting', 'Coosa Group',
    'Tallapoosa Advisory', 'Apalachicola Partners', 'Suwannee Consulting',
    'Chattahoochee Group', 'Savannah Advisory', 'Pee Dee Partners',
    'Neuse Consulting', 'Roanoke Group', 'Rappahannock Advisory',
    'Shenandoah River Partners', 'Monongahela Consulting', 'Allegheny River Group',
    'Juniata Advisory', 'Susquehanna Partners', 'Lehigh Consulting',
    'Raritan Group', 'Passaic Advisory', 'Housatonic Partners',
    'Quinnipiac Consulting', 'Farmington Group', 'Merrimack Advisory',
    'Androscoggin Partners', 'Kennebec Consulting', 'Penobscot Group',
    'St. Croix Advisory', 'Aroostook Partners', 'Saco Consulting',
    'Piscataqua Group', 'Lamprey Advisory', 'Winooski Partners',
    'Lamoille Consulting', 'Otter Creek Group', 'Mettawee Advisory',
    'Wallkill Partners', 'Esopus Consulting', 'Gloucester Advisory',
    'Newburyport Partners', 'Marblehead Consulting', 'Rockport Group',
    'Provincetown Advisory', 'Wellfleet Partners', 'Chatham Consulting',
    'Harwichport Group', 'Osterville Advisory', 'Falmouth Harbor Partners',
    'Woods Hole Consulting', 'Vineyard Haven Group', 'Edgartown Advisory',
    'Nantucket Partners', 'Hyannis Consulting', 'Barnstable Group',
    'Duxbury Advisory', 'Marshfield Partners', 'Scituate Consulting',
    'Cohasset Group', 'Hull Advisory', 'Swampscott Partners',
    'Amesbury Consulting', 'Newbury Group', 'Ipswich Advisory',
    'Gloucester Harbor Partners', 'Ogunquit Consulting', 'Kennebunkport Group',
    'York Harbor Advisory', 'Kittery Partners', 'Portsmouth Harbor Consulting',
    'Rye Group', 'Hampton Beach Advisory', 'Seabrook Partners',
    'Salisbury Consulting', 'Plum Island Group', 'Stonington Advisory',
    'Mystic Partners', 'Essex Consulting', 'Saybrook Group',
    'Guilford Advisory', 'Madison Harbor Partners', 'Branford Consulting',
    'Stony Creek Group', 'Milford Harbor Advisory', 'Beacon Hill Advisory',
    'Nob Hill Partners', 'Rittenhouse Consulting', 'Society Hill Group',
    'Germantown Advisory', 'Powelton Partners', 'Fairmount Consulting',
    'Mantua Group', 'Strawberry Mansion Advisory', 'Spring Garden Partners',
    'Fishtown Consulting', 'Kensington Group', 'Frankford Advisory',
    'Olney Partners', 'Germantown Pike Consulting', 'Chestnut Hill Group',
    'Mount Airy Advisory', 'Roxborough Partners', 'Manayunk Consulting',
    'Wissahickon Group', 'Andover Advisory', 'Concord Partners',
    'Lexington Consulting', 'Dedham Group', 'Canton Advisory',
    'Stoughton Partners', 'Braintree Consulting', 'Weymouth Group',
    'Quincy Harbor Advisory', 'Milton Partners', 'Hyde Park Consulting',
    'Roslindale Group', 'West Roxbury Advisory', 'Jamaica Plain Partners',
    'Dorchester Heights Consulting', 'South End Group', 'Back Bay Advisory',
    'Charlestown Partners', 'Somerville Consulting', 'Medford Group',
    'Malden Advisory', 'Revere Partners', 'Winthrop Harbor Consulting',
    'East Boston Group', 'Downtown Advisory', 'Acadia Advisory',
    'Shenandoah Partners', 'Congaree Consulting', 'Everglades Group',
    'Biscayne Advisory', 'Dry Tortugas Partners', 'Gulf Islands Consulting',
    'Natchez Trace Group', 'Ozark National Advisory', 'Buffalo National Partners',
    'Cuyahoga Consulting', 'Indiana Dunes Group', 'Sleeping Bear Advisory',
    'Pictured Rocks Partners', 'Apostle Islands Consulting', 'Saint Croix National Group',
    'Mississippi National Advisory', 'Effigy Mounds Partners', 'Tallgrass Consulting',
    'Flint Hills Group', 'Chickasaw Advisory', 'Quachita National Partners',
    'Hot Springs Consulting', 'Buffalo River Group', 'Big Bend Advisory',
    'Guadalupe Partners', 'White Sands Consulting', 'Chaco Group',
    'Bandelier Advisory', 'Petroglyph Partners', 'El Malpais Consulting',
    'Aztec Ruins Group', 'Mesa Verde Advisory', 'Black Canyon Partners',
    'Curecanti Consulting', 'Great Sand Dunes Group', 'Rocky Mountain Advisory',
    'Florissant Partners', 'Mueller Consulting', 'Ashworth & Calloway LLC',
    'Brentwood Forsythe Group', 'Carlisle Wyndham Advisory', 'Dalton Prescott Partners',
    'Eastwood Colby Consulting', 'Fairchild Romero Group', 'Granger Whitfield Advisory',
    'Holbrook Mercer Partners', 'Ives Collingwood Consulting', 'Jarvis Thorne Group',
    'Knott Rutherford Advisory', 'Larson Whitmore Partners', 'Munroe Stafford Consulting',
    'Norris Blackhurst Group', 'Ogden Forsythe Advisory', 'Pennington Wyle Partners',
    'Quigley Bourne Consulting', 'Rainey Holt Group', 'Stanhope Wilder Advisory',
    'Thibodaux Colman Partners', 'Underwood Farley Consulting', 'Vandermeer Stokes Group',
    'Westbrook Firth Advisory', 'Xavier Colton Partners', 'Yarborough Crane Consulting',
    'Zimmermann Alcott Group', 'Albright & Wyndham LLC', 'Blaine Sutherland Advisory',
    'Corwin Hadley Partners', 'Dearing Foxworth Consulting', 'Eaton Fairclough Group',
    'Fielding Cromwell Advisory', 'Garland Whitmore Partners', 'Hale Torrance Consulting',
    'Irvin Sedgewick Group', 'Jenks Holloway Advisory', 'Kirkwood Brent Partners',
    'Loomis Stanhope Consulting', 'McAllister Forsyth Group', 'Newsom Aldrich Advisory',
    'Ormond Blackwell Partners', 'Parkhurst Langley Consulting', 'Quade Harcourt Group',
    'Rawley Crestwood Advisory', 'Standish Fairfax Partners', 'Tilden Broughton Consulting',
    'Umstead Cranfield Group', 'Varney Hollister Advisory', 'Whitaker Salcombe Partners',
    'Xander Forsythe Consulting', 'Yancey Braden Group', 'Zollinger Hartley Advisory',
    'Armitage & Colby LLC', 'Barton Lynwood Partners', 'Chilton Prescott Consulting',
    'Duvall Harrington Group', 'Eldridge Stafford Advisory', 'Farlow Cromwell Partners',
    'Gilmore Ashby Consulting', 'Hartington Wentworth Group', 'Ingalls Blackmore Advisory',
    'Jesup Whitmore Partners', 'Keele Forsyth Consulting', 'Landon Wilder Group',
    'Merrifield Stafford Advisory', 'Norcross Langley Partners', 'Oakfield Holt Consulting',
    'Prescott Wyndham Group', 'Qualter Fairclough Advisory', 'Redfield Colman Partners',
    'Selden Ashworth Consulting', 'Talbot Wentworth Group', 'Center for Executive Compensation Studies',
    'National Institute for Workforce Policy', 'Institute for Corporate Governance Innovation', 'American Council on Talent Economics',
    'Foundation for Executive Leadership Research', 'Center for Responsible Compensation', 'National Alliance for Workforce Analytics',
    'Institute for Board Effectiveness', 'American Institute for Pay Equity Research', 'Council on Organizational Performance',
    'Foundation for Corporate Accountability', 'National Center for HR Innovation', 'Institute for People and Culture Research',
    'American Alliance for Governance Excellence', 'Center for Human Capital Measurement', 'National Foundation for Pay Transparency',
    'Institute for Workforce Resilience', 'Council on Compensation Best Practices', 'Foundation for Inclusive Leadership Research',
    'National Center for Organizational Health', 'Institute for Strategic Compensation', 'American Consortium on HR Analytics',
    'Center for Workforce Futures Research', 'National Institute for Talent Management', 'Foundation for Governance Accountability',
    'Council on Executive Pay Standards', 'Institute for Compensation Governance', 'American Center for Board Research',
    'National Alliance for HR Innovation', 'Foundation for Workforce Strategy', 'Center for ESG Governance Research',
    'National Institute for Compensation Equity', 'Institute for Corporate Accountability Studies', 'Council on People Analytics',
    'American Foundation for Organizational Research', 'Center for Talent Market Intelligence', 'National Consortium on Executive Development',
    'Institute for Board Governance Research', 'Foundation for Pay Equity Advancement', 'Council on Human Capital Excellence',
    'American Institute for Corporate Governance', 'Center for Workforce Transformation', 'National Alliance for Executive Compensation Research',
    'Institute for HR Excellence', 'Foundation for Organizational Effectiveness Research', 'Council for HR Strategy',
    'American Center for Workforce Policy', 'National Institute for Governance Studies', 'Center for Total Rewards Research',
    'Foundation for Corporate Leadership Development', 'National Council on Organizational Design', 'Institute for Benefits Research',
    'American Alliance for Talent Development', 'Center for Performance Management Research', 'National Foundation for HR Technology',
    'Institute for Workforce Development Policy', 'Council for Diversity and Inclusion Research', 'Foundation for Workforce Equity',
    'American Center for HR Effectiveness', 'National Institute for Human Capital Strategy', 'Center for Compensation and Benefits Research',
    'Institute for Employee Experience Research', 'Foundation for Leadership Development', 'Council on Workforce Innovation',
    'American Alliance for Board Diversity', 'National Center for Executive Pay Research', 'Institute for Talent Acquisition Research',
    'Center for Change Management Research', 'National Foundation for Governance Innovation', 'Council on ESG Disclosure Standards',
    'American Institute for HR Technology Research', 'Foundation for Compensation Transparency', 'Center for Labor Economics Research',
    'National Alliance for Organizational Effectiveness', 'Institute for Workforce Planning Research', 'Council on Total Rewards Strategy',
    'American Center for Pay Governance', 'National Foundation for Talent Management Research', 'Meridian Workforce Consulting',
    'Apex Governance Solutions', 'Summit Compensation Group', 'Horizon HR Advisors',
    'Catalyst Talent Consulting', 'Benchmark People Solutions', 'Keystone Compensation Advisors',
    'Zenith Governance Consulting', 'Atlas HR Solutions', 'Cornerstone Workforce Advisors',
    'Momentum Compensation Consulting', 'Sterling Governance Solutions', 'Fulcrum Talent Advisors',
    'Harbor HR Consulting', 'Landmark People Solutions', 'Northfield Compensation Advisors',
    'Ridgeline Governance Consulting', 'Stonegate HR Solutions', 'Vantage People Advisors',
    'Alderwood Compensation Consulting', 'Beacon Governance Solutions', 'Cascade HR Advisors',
    'Cedarwood Workforce Consulting', 'Chesapeake Compensation Group', 'Crestwood Governance Advisors',
    'Crossroads HR Solutions', 'Delphi Compensation Advisors', 'Eastgate Governance Consulting',
    'Fairway HR Solutions', 'Fieldstone Workforce Advisors', 'Garrison Compensation Consulting',
    'Glenbrook Governance Group', 'Greenfield HR Solutions', 'Greystone Workforce Advisors',
]

_ARCHIVE_INDUSTRIES = [
    'Healthcare', 'Financial Services', 'Energy', 'Manufacturing', 'Technology',
    'Retail', 'Transportation', 'Education', 'Government', 'Pharmaceuticals',
    'Professional Services', 'Insurance', 'Logistics', 'Aerospace', 'Utilities',
    'Consumer Goods', 'Real Estate', 'Nonprofit', 'Defense', 'Biotechnology',
    'Banking', 'Asset Management', 'Private Equity', 'Venture Capital', 'Hedge Funds',
    'Investment Banking', 'Capital Markets', 'Wealth Management', 'Insurance Brokerage',
    'Reinsurance', 'Specialty Finance', 'Mortgage Banking', 'Commercial Real Estate',
    'REITs', 'Infrastructure', 'Renewable Energy', 'Oil & Gas', 'Mining',
    'Chemicals', 'Specialty Chemicals', 'Agriculture', 'Food & Beverage',
    'Hospitality', 'Travel & Tourism', 'Entertainment & Media', 'Publishing',
    'Broadcasting', 'Telecommunications', 'Semiconductor', 'Software', 'SaaS',
    'Cybersecurity', 'Cloud Computing', 'Artificial Intelligence', 'Data Analytics',
    'E-Commerce', 'Fintech', 'Healthtech', 'Edtech', 'Proptech', 'Insurtech',
    'Medical Devices', 'Life Sciences', 'Clinical Research', 'Diagnostics',
    'Behavioral Health', 'Home Health', 'Managed Care', 'Health Systems',
    'Academic Medical Centers', 'Physician Practice Management',
    'Automotive', 'Aerospace & Defense', 'Naval Systems', 'Space Technology',
    'Environmental Services', 'Waste Management', 'Water Utilities',
    'Architecture & Engineering', 'Construction', 'Engineering & Construction',
    'Management Consulting', 'Legal Services', 'Accounting Services',
    'Staffing & Recruiting', 'Executive Search', 'Human Resources Outsourcing',
    'Business Process Outsourcing', 'IT Services', 'Systems Integration',
    'Sports & Recreation', 'Gaming & Gambling', 'Luxury Goods', 'Apparel & Fashion',
    'Beauty & Personal Care', 'Consumer Electronics', 'Home Improvement',
    'Grocery & Supermarkets', 'Quick Service Restaurants', 'Full Service Restaurants',
    'Specialty Retail', 'Department Stores', 'Direct-to-Consumer',
    'Shipping & Freight', 'Rail Transportation', 'Aviation', 'Maritime',
    'Supply Chain & Procurement', 'Third-Party Logistics',
    'Higher Education', 'K-12 Education', 'Vocational Training', 'Online Education',
    'Think Tanks & Policy Research', 'Trade Associations', 'Foundations & Endowments',
    'Faith-Based Organizations', 'Labor Unions', 'Cooperatives',
    'Federal Government', 'State & Local Government', 'Municipal Utilities',
    'Public Safety', 'Defense Contractors', 'Intelligence Community Support',
    # Additional sub-industries and emerging sectors
    'Nuclear Energy', 'Hydroelectric Power', 'Solar Energy', 'Wind Energy',
    'Battery Storage', 'Electric Vehicles', 'Autonomous Vehicles', 'Mobility-as-a-Service',
    'Drone Technology', 'Satellite Communications', 'Quantum Computing', 'Blockchain',
    'Web3 & Decentralized Finance', 'Digital Assets', 'Payments Processing', 'Lending Technology',
    'Regulatory Technology', 'Compliance Technology', 'Legal Technology', 'Contract Management',
    'Property Management', 'Affordable Housing', 'Senior Living', 'Student Housing',
    'Data Center REITs', 'Cell Tower Infrastructure', 'Fiber Networks',
    'Cable & Broadband', 'Streaming Media', 'Digital Advertising', 'Ad Technology',
    'Market Research', 'Consumer Intelligence', 'Public Relations', 'Event Management',
    'Trade Show & Exhibition', 'Corporate Training', 'Executive Coaching',
    'Organizational Development Consulting', 'Change Management Consulting',
    'IT Consulting', 'Strategy Consulting', 'Operations Consulting',
    'Financial Advisory', 'Transaction Advisory', 'Restructuring Advisory',
    'Valuation Services', 'Transfer Pricing', 'Tax Advisory',
    'Franchise & Licensing', 'Direct Selling', 'Subscription Commerce',
    'Convenience Retail', 'Gas Station & C-Store', 'Pharmacy & Drug Stores',
    'Sporting Goods', 'Toy & Hobby', 'Pet Care', 'Garden & Nursery',
    'Office Supplies', 'Arts & Crafts', 'Musical Instruments',
    'Book Publishing', 'Magazine & Periodicals', 'Newsletter Media',
    'Podcast & Audio', 'Film & Television Production', 'Post-Production',
    'Visual Effects', 'Animation', 'Video Games & Interactive Media',
    'Esports', 'Virtual Reality', 'Augmented Reality', 'Metaverse',
    'Social Media Platforms', 'Search Engines', 'Content Platforms',
    'Talent Agencies', 'Sports Management', 'Athletic Facilities',
    'Fitness & Wellness', 'Spa & Personal Services',
    'Funeral Services', 'Cemetery Management', 'Estate Planning Services',
    'Private Banking', 'Family Offices', 'Endowment Management',
    'Pension Fund Management', 'Sovereign Wealth', 'Development Finance',
    'Microfinance', 'Community Banking', 'Credit Unions',
    'Agricultural Banking', 'Farm Credit Services', 'Crop Insurance',
    'Veterinary Services', 'Animal Health', 'Aquaculture', 'Forestry',
    'Timber', 'Paper & Packaging', 'Plastics & Composites',
    'Coatings & Adhesives', 'Fertilizers & Pesticides', 'Industrial Gases',
    'Metalworking', 'Precision Machining', 'Tool & Die', 'Foundry Operations',
    'Electronics Manufacturing', 'Printed Circuit Boards', 'Display Technology',
    'Robotics & Automation', 'Industrial IoT', 'Smart Manufacturing',
    'Additive Manufacturing', '3D Printing Services',
    'Contract Research Organizations', 'Pharmacy Benefit Management',
    'Hospital Systems', 'Ambulatory Surgery Centers', 'Urgent Care',
    'Telehealth', 'Mental Health Services', 'Substance Use Treatment',
    'Dental Services', 'Vision Care', 'Chiropractic & Physical Therapy',
    'Home Medical Equipment', 'Infusion Services', 'Laboratory Services',
    'Radiology & Imaging', 'Pathology Services',
    'Corrections & Detention', 'Homeland Security Services',
    'Emergency Management', 'Environmental Consulting',
    'Geotechnical Engineering', 'Surveying & Mapping',
    'Telecommunications Equipment', 'Network Infrastructure', 'Managed Services',
    'Cloud Security', 'Identity Management', 'DevOps Platforms',
    'Low-Code/No-Code Development', 'API Management', 'Integration Platforms',
    'Master Data Management', 'Enterprise Content Management',
    'Geographic Information Systems', 'Remote Sensing',
    'Philanthropy & Foundations', 'Social Enterprises', 'Impact Investing',
    'Community Development Finance', 'Workforce Development Programs',
    'Charter Schools', 'Private Schools', 'Test Preparation Services',
    'Language Learning', 'Certification Bodies', 'Accreditation Organizations',
    'Architecture', 'Interior Design', 'Landscape Architecture',
    'Urban Planning', 'Civil Engineering', 'Structural Engineering',
    'Mechanical Engineering', 'Electrical Engineering', 'Environmental Engineering',
    'Water Treatment', 'Wastewater Management', 'Solid Waste Management',
    'Recycling & Materials Recovery', 'Remediation Services', 'Environmental Testing',
    'Port Operations', 'Freight Brokerage', 'Customs Brokerage',
    'Cold Chain Logistics', 'Last-Mile Delivery', 'Fulfillment Centers',
    'Auto Dealerships', 'Auto Parts & Accessories', 'Fleet Management',
    'Rental Car Services', 'Ride-Hailing Platforms', 'Parking Services',
    # Finance sub-specializations, healthcare sub-specializations, tech sub-specializations, emerging/niche, B2B services, utilities, government/nonprofit, food/ag
    'Structured Finance', 'Securitization', 'Project Finance',
    'Infrastructure Finance', 'Municipal Finance', 'Public Finance Advisory',
    'Export-Import Finance', 'Trade Finance', 'Supply Chain Finance',
    'Factoring Services', 'Equipment Finance', 'Aircraft Finance',
    'Ship Finance', 'Rail Finance', 'Leveraged Finance',
    'High-Yield Debt', 'Distressed Debt', 'Mezzanine Finance',
    'Second-Lien Lending', 'Direct Lending', 'Business Development Companies',
    'Interval Funds', 'Tender Offer Funds', 'Closed-End Funds',
    'Liquid Alternatives', 'Multi-Asset Solutions', 'Liability-Driven Investing',
    'Outsourced-CIO Services', 'OCIO for Endowments', 'OCIO for Pension Funds',
    'OCIO for Insurance', 'Integrated Delivery Networks', 'Critical Access Hospitals',
    "Children's Hospitals", 'Cancer Centers', 'Cardiac Care Centers',
    'Specialty Hospitals', 'Long-Term Acute Care', 'Inpatient Rehabilitation',
    'Skilled Nursing Facilities', 'Memory Care Facilities', 'Continuing Care Retirement Communities',
    'Assisted Living', 'Independent Living Communities', 'Life Plan Communities',
    'Hospice Care', 'Palliative Care', 'Home Hospice Services',
    'Personal Care Services', 'Private Duty Nursing', 'Staffing for Healthcare',
    'Locum Tenens Staffing', 'Travel Nursing', 'Allied Health Staffing',
    'Healthcare Revenue Cycle Management', 'Medical Coding Services', 'Healthcare IT Staffing',
    'Health Information Management', 'Population Health Management', 'Value-Based Care Organizations',
    'Accountable Care Organizations', 'Clinically Integrated Networks', 'Direct Primary Care',
    'Concierge Medicine', 'Employer Health Clinics', 'Worksite Wellness Programs',
    'Pharmacy Services', 'Specialty Pharmacy', 'Mail-Order Pharmacy',
    '340B Program Administration', 'Hospital Group Purchasing', 'Healthcare GPO Services',
    'Platform Engineering', 'Site Reliability Engineering', 'DataOps',
    'MLOps', 'FinOps', 'CloudOps',
    'AIOps', 'Edge Computing', 'Fog Computing',
    'Embedded Systems', 'Firmware Development', 'FPGA Design',
    'ASIC Design', 'Photonics', 'Optical Networking',
    'Wireless Infrastructure', 'Private 5G Networks', 'CBRS Band Deployment',
    'Open RAN', 'Network Function Virtualization', 'Software-Defined Networking',
    'Intent-Based Networking', 'Unified Communications', 'Contact Center Technology',
    'Workforce Management Software', 'Field Service Management', 'Asset Management Software',
    'Enterprise Asset Management', 'Facilities Management Software', 'Building Information Modeling',
    'Digital Twins', 'Computer-Aided Design Services', 'Product Lifecycle Management',
    'Manufacturing Execution Systems', 'Supervisory Control and Data Acquisition', 'Industrial Control Systems',
    'Process Automation', 'Distributed Control Systems', 'Building Automation',
    'Smart Grid Technology', 'Meter Data Management', 'Energy Management Software',
    'Carbon Accounting Software', 'ESG Reporting Software', 'Governance, Risk and Compliance Software',
    'Contract Lifecycle Management', 'Procure-to-Pay Software', 'Source-to-Pay Technology',
    'Treasury Management Systems', 'Order Management Systems', 'Transportation Management Systems',
    'Warehouse Management Systems', 'Inventory Optimization Software', 'Space Tourism',
    'Commercial Spaceflight', 'Satellite Launch Services', 'Small Satellite Manufacturing',
    'Ground Station Operations', 'Space Situational Awareness', 'Asteroid Mining',
    'Lunar Resources', 'Advanced Air Mobility', 'Urban Air Mobility',
    'Electric Aircraft', 'Hydrogen Aviation', 'Sustainable Aviation Fuel',
    'Hyperloop Technology', 'Maglev Rail', 'High-Speed Rail',
    'Micro-Mobility', 'Electric Scooters', 'Bike Sharing',
    'Connected Vehicles', 'V2X Communications', 'Automotive Cybersecurity',
    'Telematics', 'Fleet Telematics', 'Usage-Based Insurance',
    'Parametric Insurance', 'Catastrophe Bonds', 'Insurance-Linked Securities',
    'Captive Insurance', 'Risk Retention Groups', "Lloyd's of London Syndicates",
    'Managing General Agents', 'Program Administrators', 'Wholesale Insurance',
    'Surplus Lines', 'Excess and Surplus Lines', 'Background Screening Services',
    'Drug Testing Services', 'Occupational Health Services', 'Employee Assistance Programs',
    'Wellness Platform Providers', 'Financial Wellness Services', 'Student Loan Benefit Providers',
    'Emergency Savings Programs', 'Voluntary Benefits Providers', 'Supplemental Health Insurance',
    'Critical Illness Insurance', 'Accident Insurance', 'Pet Insurance',
    'Identity Theft Protection Services', 'Legal Insurance Plans', 'Auto and Home Group Benefits',
    'Commuter Benefits Administration', 'Flexible Spending Account Administration', 'Health Savings Account Administration',
    'Health Reimbursement Arrangement Administration', 'COBRA Administration', 'Leave Management Services',
    'Absence Management Outsourcing', 'Payroll Processing', 'Professional Employer Organizations',
    'Employer of Record Services', 'Global PEO', 'International Payroll',
    'Global Mobility Services', 'Expatriate Tax Services', 'Immigration Services',
    'Visa Processing Services', 'Relocation Services', 'Destination Services',
    'Investor-Owned Utilities', 'Public Power Utilities', 'Electric Cooperatives',
    'Public Power Districts', 'Transmission-Only Companies', 'Power Marketers',
    'Independent Power Producers', 'Wholesale Power Trading', 'Retail Energy Providers',
    'Community Choice Aggregation', 'Virtual Power Plants', 'Demand Response Providers',
    'Energy Efficiency Services', 'Distributed Energy Resources', 'Microgrid Developers',
    'Combined Heat and Power', 'District Energy Systems', 'Geothermal Energy',
    'Tidal Energy', 'Wave Energy', 'Green Hydrogen',
    'Blue Hydrogen', 'Ammonia Production', 'Carbon Capture and Storage',
    'Direct Air Capture', 'Bioenergy with Carbon Capture', 'Voluntary Carbon Markets',
    'Compliance Carbon Markets', 'Renewable Energy Credits', 'Regional Planning Commissions',
    'Metropolitan Planning Organizations', 'Council of Governments', 'Special Districts',
    'Improvement Districts', 'Economic Development Organizations', 'Community Development Corporations',
    'Housing Finance Agencies', 'State Revolving Funds', 'Public Pension Systems',
    'Public Employee Benefits Programs', 'Sovereign Immunities Entities', 'Government-Sponsored Enterprises',
    'Quasi-Public Agencies', 'Public Utilities Commissions', 'State Insurance Departments',
    'State Banking Regulators', 'State Securities Regulators', 'Congressional Support Organizations',
    'Federal Reserve Banks', 'Federal Home Loan Banks', 'International Development Finance',
    'Multilateral Development Banks', 'Export Credit Agencies', 'Bilateral Aid Organizations',
    'Intergovernmental Organizations', 'Treaty-Based Organizations', 'Precision Agriculture',
    'AgTech', 'Controlled Environment Agriculture', 'Vertical Farming',
    'Cellular Agriculture', 'Alternative Proteins', 'Plant-Based Food',
    'Fermentation Technology', 'Food Safety Technology', 'Restaurant Technology',
    'Ghost Kitchens', 'Food Delivery Platforms', 'Meal Kit Services',
    'Corporate Catering', 'Contract Food Service', 'Nutraceuticals',
    'Functional Foods', 'Sports Nutrition', 'Medical Nutrition',
    'Infant Formula', 'Fortified Foods', 'Halal Certification Services',
    'Kosher Certification Services', 'Organic Certification Services', 'Fair Trade Organizations',
    'Carbon Labeling Services', 'Product Sustainability Assessment',
]

_ARCHIVE_PHASES = [
    'discovery', 'scoping', 'data collection', 'data validation', 'analysis',
    'preliminary analysis', 'deep-dive analysis', 'review', 'peer review',
    'internal review', 'client review', 'validation', 'quality assurance',
    'reporting', 'draft reporting', 'final reporting', 'presentation',
    'executive presentation', 'board presentation', 'implementation',
    'pilot implementation', 'full implementation', 'closeout', 'post-closeout review',
    'follow-on assessment', 'sustainability review', 'annual check-in',
    'preliminary findings', 'interim findings', 'final findings',
    'stakeholder alignment', 'stakeholder socialization', 'stakeholder sign-off',
    'final documentation', 'archival', 'transition planning', 'knowledge transfer',
    'project initiation', 'requirements gathering', 'benchmarking',
    'gap analysis', 'root cause analysis', 'options development',
    'recommendation development', 'business case development',
    'change impact assessment', 'communication planning', 'training development',
    'readiness assessment', 'go-live preparation', 'stabilization', 'optimization',
    'lessons learned', 'retrospective', 'post-implementation review',
    'audit preparation', 'audit support', 'remediation', 'remediation testing',
    'regulatory filing preparation', 'regulatory response',
    'due diligence', 'confirmatory due diligence', 'integration planning',
    'integration execution', 'separation planning', 'separation execution',
    'workforce restructuring', 'headcount rightsizing', 'cost optimization',
    'operating model design', 'operating model implementation',
    # Additional phases
    'environmental scanning', 'hypothesis development', 'framework development',
    'instrument design', 'pilot testing', 'pre-launch review',
    'launch', 'post-launch monitoring', 'mid-course correction',
    'performance review', 'impact evaluation', 'outcome measurement',
    'benefit realization', 'value confirmation', 'return on investment analysis',
    'executive alignment', 'management alignment', 'operational alignment',
    'cross-functional coordination', 'interdependency mapping', 'sequencing',
    'critical path analysis', 'resource leveling', 'schedule compression',
    'risk mitigation', 'contingency planning', 'issue resolution',
    'decision-making', 'escalation resolution', 'governance review',
    'charter refresh', 'scope validation', 'budget reforecasting',
    'resource reallocation', 'team restructuring', 'vendor renegotiation',
    'contract finalization', 'legal review', 'regulatory clearance',
    'data privacy review', 'security assessment', 'access provisioning',
    'system testing', 'user acceptance testing', 'parallel running',
    'cutover preparation', 'go-live execution', 'hypercare',
    'steady state', 'continuous improvement', 'maturity assessment',
    'capability development', 'tool implementation', 'process embedding',
    'culture embedding', 'behavioral change assessment', 'adoption measurement',
    'benefits tracking', 'lessons documentation', 'knowledge base update',
    'decommissioning', 'archival preparation', 'records disposition',
    'final sign-off preparation', 'executive sign-off', 'sponsor sign-off',
    'board sign-off', 'audit completion', 'file closure',
    'market analysis', 'competitor analysis', 'customer research',
    'user research', 'personas development', 'journey mapping',
    'service design', 'prototype development', 'iteration',
    'alpha testing', 'beta testing', 'usability testing',
    'field validation', 'site assessment', 'ground-truthing',
    'expert panel review', 'Delphi process', 'consensus building',
    'public comment review', 'regulatory comment response',
    'negotiation', 'mediation', 'arbitration preparation',
    'settlement documentation', 'consent decree monitoring',
    'corrective action planning', 'corrective action implementation',
    'verification testing', 'certification preparation', 'certification audit',
    'surveillance audit', 'recertification', 'third-party assessment',
    'internal control testing', 'walkthrough', 'substantive testing',
    'analytical procedures', 'management inquiry', 'representation',
    'opinion issuance', 'management letter', 'post-audit review',
    'board reporting', 'committee reporting', 'executive dashboard update',
    'townhall preparation', 'all-hands planning', 'communication rollout',
    'feedback collection', 'sentiment analysis', 'pulse check',
    'retrospective facilitation', 'after-action review', 'hotwash',
    # Agile, procurement, M&A, regulatory/legal, consulting sub-phases, technology project, HR program, research methodology
    'sprint planning', 'sprint execution', 'sprint review',
    'sprint retrospective', 'backlog grooming', 'backlog refinement',
    'product backlog creation', 'release planning', 'release review',
    'increment review', 'definition of done review', 'velocity assessment',
    'burndown analysis', 'velocity calibration', 'story point estimation',
    'daily standup series', 'demo preparation', 'stakeholder demo',
    'product increment documentation', 'agile maturity assessment', 'market sounding',
    'request for information', 'supplier pre-qualification', 'request for proposal issuance',
    'bid evaluation', 'best and final offer', 'negotiation preparation',
    'commercial negotiation', 'contract award', 'contract execution',
    'mobilization', 'performance monitoring', 'contract extension evaluation',
    'rebid preparation', 'transition out', 'supplier offboarding',
    'procurement audit', 'category strategy development', 'spend baseline establishment',
    'savings tracking', 'strategic rationale development', 'target identification',
    'initial screening', 'preliminary due diligence', 'management meeting',
    'indication of interest', 'letter of intent', 'exclusivity negotiation',
    'definitive agreement drafting', 'regulatory filing', 'regulatory review period',
    'shareholder vote preparation', 'closing preparation', 'closing',
    'day-one integration', 'first hundred days', 'synergy tracking',
    'post-merger review', 'integration closeout', 'value realization review',
    'regulatory pre-filing consultation', 'regulatory submission', 'public comment period',
    'hearing preparation', 'evidentiary hearing', 'regulatory order implementation',
    'compliance attestation', 'regulatory examination preparation', 'examination response',
    'consent decree negotiation', 'consent decree implementation', 'consent decree compliance review',
    'regulatory exit', 'license renewal', 'permit application',
    'permit amendment', 'variance request', 'environmental review',
    'public notice', 'appeals resolution', 'engagement scoping',
    'fee negotiation', 'staffing planning', 'kickoff facilitation',
    'hypothesis generation', 'workstream launch', 'data request issuance',
    'data receipt and review', 'analytical sprint', 'synthesis session',
    'insight generation', 'storyboarding', 'draft review session',
    'findings socialization', 'executive alignment session', 'recommendation refinement',
    'implementation roadmap development', 'quick-win identification', 'business case validation',
    'steering committee update', 'engagement closeout', 'contract close',
    'technology assessment', 'architecture design', 'vendor selection',
    'proof of concept', 'architecture validation', 'development sprint',
    'integration testing', 'performance testing', 'security testing',
    'user acceptance testing phase two', 'production readiness review', 'change advisory board review',
    'go-live authorization', 'go-live monitoring', 'hypercare monitoring',
    'stabilization period', 'tuning and optimization', 'knowledge transfer to operations',
    'run-state transition', 'decommissioning planning', 'legacy sunset',
    'program design', 'pilot group selection', 'facilitator training',
    'participant enrollment', 'baseline measurement', 'cohort launch',
    'mid-program check-in', 'program completion', 'graduation',
    'alumni follow-up', 'program evaluation', 'impact measurement',
    'redesign planning', 'curriculum refresh', 'facilitator recertification',
    'program scale-up', 'enterprise rollout', 'embedding phase',
    'program handover', 'sustainability planning', 'literature review',
    'conceptual framework development', 'research design', 'instrument development',
    'cognitive testing', 'pre-test', 'sample design',
    'sampling frame development', 'fieldwork launch', 'fieldwork monitoring',
    'fieldwork close', 'data cleaning', 'coding',
    'data analysis', 'preliminary reporting', 'peer review response',
    'final manuscript preparation', 'publication', 'dissemination',
    'knowledge translation',
]

_ARCHIVE_PARA_TEMPLATES = [
    "This archive entry documents the {phase} phase of ACPWB's engagement with {org} in the {industry} sector. The record was generated on {date} and reflects the state of the engagement as of that date. All substantive findings and recommendations have been incorporated into subsequent deliverable documentation.",
    "The following record was created pursuant to ACPWB's standard document retention protocol, which mandates contemporaneous archiving of all client-facing deliverables, internal memoranda, and stakeholder communications. This entry pertains to the {phase} of a {industry} sector engagement initiated in {year}.",
    "ACPWB's engagement management system automatically generated this archive entry upon completion of a key milestone in the {org} partnership. The entry captures the current state of {n} discrete work streams across {regions} regional offices as of {date}.",
    "As part of ACPWB's commitment to institutional transparency, this archive entry documents the outcomes of the {phase} review conducted for the {industry} sector engagement portfolio. The analysis reflects data from {n} participating organizations.",
    "This document represents an intermediate archival record from a multi-phase engagement. The preceding phase concluded on {date}, at which point {n} deliverables were formally archived in accordance with ACPWB's records management framework.",
    "The {org} engagement referenced in this archive entry was initiated in response to a sector-wide need for {industry} benchmarking data. The record captures the {phase} phase, which involved data collection from {n} respondent organizations across {regions} states.",
    "Documentation archived at this location reflects the work product of ACPWB's {industry} practice group during the period ending {date}. The practice group comprised {n} dedicated staff supported by {regions} external advisors retained for specialized expertise.",
    "This archive entry was created during the systematic review of ACPWB's historical engagement records. It preserves documentation related to the {phase} phase of the {org} project, including all supporting analysis, stakeholder communications, and regulatory correspondence.",
    "The record at this location reflects a cross-functional collaboration between ACPWB's {industry} practice and its Research Division. The collaboration produced {n} discrete analytical outputs, all of which are catalogued in this archive entry and available upon request.",
    "ACPWB's institutional memory database contains records of all client engagements dating back to the organization's founding. This entry, archived on {date}, pertains to a {industry} sector engagement with {org} and reflects the documentation standards applied throughout ACPWB's history.",
    "The {phase} documentation archived here reflects the culmination of {n} months of continuous engagement work in the {industry} sector. {org} served as the primary client stakeholder throughout this period, with supplemental coordination provided by {regions} advisory partners.",
    "Engagement records of this type are retained for a minimum of 18 years under ACPWB's document retention policy. The {phase} materials archived on {date} include all work product generated by the engagement team, indexed by document type, author, and revision history.",
    "This entry captures a significant inflection point in the {org} engagement — the transition from the {phase} phase to subsequent implementation activities. The record includes a status summary covering {n} open action items, {regions} pending approvals, and the full set of final deliverables produced to that date.",
    "The {industry} sector engagement archived here spans a {n}-month analytical cycle encompassing data from {regions} jurisdictions. ACPWB's Research Division coordinated data collection protocols with {org} to ensure comparability across the full respondent pool.",
    "This archival record reflects ACPWB's standard {phase} methodology applied to {industry} sector clients with workforce populations exceeding {n} employees. {org} provided access to {regions} data systems as part of the structured information-sharing agreement governing this engagement.",
    "The documentation preserved at this location was produced under ACPWB's ISO-aligned quality management framework. The {phase} phase was independently reviewed by {org}'s internal audit function on {date}, with all substantive observations resolved prior to archival.",
    "ACPWB's engagement with {org} commenced in response to a competitive procurement process in which {regions} firms submitted proposals. This archive entry captures the {phase} deliverables that met or exceeded all specifications outlined in the original scope of work.",
    "The {industry} benchmarking data preserved in this archive entry was collected from {n} organizations representing a combined workforce of approximately {regions}0,000 employees. {org} participated as both a data contributor and a primary client throughout the {phase} phase.",
    "Pursuant to the data sharing agreement executed between ACPWB and {org} on {date}, all materials produced during the {phase} phase are archived here and are subject to a {regions}-year confidentiality period. Access to underlying datasets requires written authorization from the ACPWB Records Management Office.",
    "This entry documents the final disposition of deliverables from the {phase} phase of the {org} engagement. A total of {n} documents were reviewed, revised, and formally approved prior to archival. The {industry} sector context for this engagement is detailed in the supplemental background memorandum attached to the primary record.",
    "The {org} engagement referenced in this archive entry required the integration of {n} discrete proprietary datasets spanning {regions} calendar years of historical data. ACPWB's data science team developed custom reconciliation logic to normalize terminology and coding conventions across sources prior to the {phase} phase.",
    "This archive entry reflects findings from a joint research initiative co-sponsored by {org} and {regions} peer organizations operating in the {industry} sector. The collaborative structure enabled a sample size of {n} that exceeds the threshold for publication in ACPWB's peer-reviewed research series.",
    "The {phase} phase documentation archived here includes all correspondence with {org}'s legal counsel regarding the interpretation of {industry} sector regulatory requirements applicable to the engagement scope. {n} formal legal opinions were obtained and are incorporated by reference into this archive record.",
    "ACPWB's engagement team completed the {phase} phase ahead of the contractual schedule by {n} days, achieving a cost underrun of approximately {regions}% relative to the original budget. This entry captures the revised timeline and associated resource reallocation decisions approved by the {org} project steering committee.",
    "The methodology applied during the {phase} phase of the {org} engagement draws on ACPWB's proprietary {industry} sector assessment framework, which has been validated against outcomes data from {n} prior engagements conducted between {year} and {endyear}.",
    "This archive entry preserves the intermediate analytical outputs produced during the {phase} phase, including {n} scenario models, {regions} sensitivity analyses, and the full supporting dataset from which the final recommendations were derived.",
    "The {industry} sector landscape had undergone significant structural change in the {regions} months preceding the {phase} phase, necessitating several scope adjustments that are documented in this archive entry. {org}'s executive team approved all scope modifications via the formal change control process.",
    "ACPWB's {phase} work for {org} incorporated primary research conducted across {regions} geographic markets, supplemented by secondary analysis of publicly available {industry} sector data compiled from {n} regulatory, commercial, and academic sources.",
    "The documentation archived at this location represents the definitive record of the {org} {phase} engagement and supersedes all prior draft versions. {n} formal review cycles were completed before the materials received final approval from both {org}'s designated project sponsor and ACPWB's engagement quality reviewer.",
    "Stakeholder alignment sessions conducted during the {phase} phase surfaced {n} areas of divergent perspective among {org}'s senior leadership team. This archive entry includes the facilitated discussion summaries and the consensus framework developed to resolve those differences prior to final deliverable production.",
    "The {industry} sector benchmarking component of this engagement drew on ACPWB's proprietary database of {n} compensation and governance data points, updated through {date} and covering organizations with aggregate market capitalization exceeding ${regions} trillion.",
    "This archive entry documents ACPWB's application of mixed-methods research design to the {org} {phase} engagement, combining quantitative survey data from {n} respondents with qualitative insights from {regions} structured executive interviews conducted across {industry} sector organizations.",
    "The regulatory landscape applicable to {industry} sector organizations underwent {n} material changes during the period covered by this archive entry, each of which required assessment of impact on the {org} engagement scope. This record includes ACPWB's regulatory monitoring summaries and impact assessments for each change identified.",
    "ACPWB's {phase} team deployed a structured facilitation methodology to surface latent assumptions within {org}'s leadership team, ultimately resolving {n} points of disagreement across {regions} organizational sub-units. The resulting consensus framework is preserved in full within this archive entry.",
    "The {industry} sector engagement archived here involved coordination with {n} counterpart organizations, each of which contributed data under strict confidentiality protocols. {org} served as the coordinating entity, managing data flows across {regions} jurisdictions pursuant to a multi-party data sharing agreement executed prior to the {phase} phase.",
    "ACPWB's engagement quality framework mandates independent validation of all primary data collected during the {phase} phase. This archive entry includes the validation report confirming that {n}% of data elements met applicable quality thresholds, with discrepancies in {regions} records resolved through supplemental data collection.",
    "The {org} project steering committee approved {n} scope adjustments during the {phase} phase, each documented via ACPWB's formal change control process. This archive entry preserves all change requests, approvals, and associated impact assessments produced during the {regions}-month engagement period.",
    "ACPWB's institutional knowledge base, accumulated across {n} engagements in the {industry} sector since {year}, informed the analytical framework applied during the {phase} phase. {org} specifically requested that ACPWB draw on prior {industry} sector experience to contextualize findings within {regions} peer organizations.",
    "The {phase} outputs archived here were developed by an ACPWB engagement team comprising {n} full-time equivalents drawn from {regions} practice areas, supported by three external subject matter experts retained specifically for their {industry} sector expertise.",
    "This archive entry reflects ACPWB's first engagement with {org} following a competitive re-procurement process in which {n} firms submitted proposals. The {phase} phase was structured to address capability gaps identified in the prior vendor's {regions}-year engagement, with benchmarking relative to {industry} sector best practices.",
    "Fieldwork conducted during the {phase} phase included structured site visits to {n} {org} locations across {regions} states, with observational data collection supplemented by interviews with {n} frontline staff and their direct supervisors.",
    "The data governance framework applied to the {phase} engagement required that all {org} data be processed within ACPWB's secure analytical environment, with {n} discrete data assets subject to individual classification decisions under ACPWB's {industry} sector data handling protocol.",
    "ACPWB's proprietary {industry} sector maturity model, applied during the {phase} phase, assesses organizations across {n} capability dimensions. {org} achieved scores in the top {regions}th percentile on {n} dimensions, with development opportunities concentrated in the remaining categories.",
    "The {phase} archive entry reflects a collaboration between ACPWB's {industry} practice team and its emerging research collaborative, which brings together practitioners from {n} member organizations to develop cross-sector insights. {org} served as one of {regions} pilot sites for the collaborative's inaugural benchmarking initiative.",
    "Executive stakeholders engaged during the {phase} phase represented {n} distinct functional areas within {org}, ensuring that the resulting recommendations reflect operational realities across the full scope of the {industry} sector engagement. {regions} additional subject matter experts participated in targeted working sessions.",
    "This archive entry documents the successful transfer of {n} analytical work products from the {phase} phase team to {org}'s internal analytics function, completing a {regions}-month capability building initiative designed to ensure sustained use of ACPWB's {industry} sector benchmarking framework.",
    "The {org} engagement involved deployment of ACPWB's proprietary survey instrument to {n} respondents across {regions} organizational levels. The instrument, validated against {industry} sector norms established across prior engagements, achieved an internal consistency coefficient exceeding the threshold required for publication in ACPWB's annual research compendium.",
    "Risk factors identified during the {phase} phase were assessed using ACPWB's {industry} sector risk taxonomy, which encompasses {n} discrete risk categories organized across {regions} primary risk domains. {org} presented elevated exposure in {n} categories relative to sector benchmarks.",
    "The documentation archived here reflects {n} months of continuous stakeholder engagement, during which ACPWB facilitated {regions} working sessions, {n} executive briefings, and {regions} board touchpoints on behalf of {org}. All session materials and participant lists are preserved in the supporting documentation repository.",
    "ACPWB's {phase} methodology incorporates a structured pre-mortem exercise in which the {org} engagement team identifies and evaluates {n} hypothetical failure scenarios. This archive entry includes the pre-mortem report and the associated risk mitigation actions that were implemented prior to final delivery.",
    "The {industry} sector landscape analysis conducted during the {phase} phase drew on {n} primary data sources and {regions} secondary research databases, covering the period from {year} through {endyear}. {org}'s competitive positioning is assessed relative to {n} direct peers and {regions} adjacent-sector comparators.",
    "This archive entry preserves the full data lineage documentation for the {phase} engagement, including source system descriptions, extraction methodologies, transformation logic, and quality validation protocols applied to each of the {n} datasets incorporated into the {org} analytical environment.",
    "The {org} engagement team's {phase} deliverables were subjected to ACPWB's three-tier review protocol, involving {n} independent reviewers drawn from the firm's {industry} practice, quality assurance function, and senior advisory board. All reviewer comments and the engagement team's responses are archived in the supplemental review file.",
    "ACPWB's {phase} work for {org} was structured as a modular engagement, enabling {industry} sector clients to activate specific analytical components based on their immediate priorities. {org} activated {n} of the {regions} available modules, with the remaining modules documented here for potential activation in future engagement phases.",
    "The archive entry reflects the formal close of a {n}-phase engagement with {org} that commenced in {year}. Across all phases, the engagement produced {regions} discrete deliverables, engaged {n} stakeholders, and generated benchmarking data referenced by {regions} subsequent {industry} sector clients.",
    # Crisis/incident, enforcement, SEC letters, whistleblower, SPAC, ESG, proxy fight, international expansion, digital transformation, AI governance, climate, geopolitical, investigation, network analysis, platform economics, regulatory arbitrage, NLP, DPA monitoring, benefits, workforce analytics, human rights, restructuring, comp committee education, talent market, insolvency, EEOC, change-of-control, shareholder comms, operational resilience, org network, pay transparency, succession, consent decree, workforce segmentation, pay fairness, cultural integration, workforce cost, director pay, virtual work, regulatory readiness, skills-based org, exec effectiveness, ESG pay linkage, board refreshment, workforce reduction, total rewards comms, regulatory capital, org effectiveness, LTI design
    "This archive entry documents ACPWB's crisis response engagement with {org} following a material incident affecting the {industry} sector. The {phase} phase, conducted between {date} and the archival date, produced {n} incident response deliverables across {regions} functional areas, all preserved in accordance with ACPWB's crisis documentation protocol.", "Regulatory enforcement activity targeting the {industry} sector intensified in the {regions} months preceding this engagement, prompting {org} to retain ACPWB for an expedited {phase} assessment. The {n} findings documented in this archive entry were delivered directly to the {org} board's audit committee on {date}.", "This record captures ACPWB's engagement support for {org} in connection with an SEC comment letter received on {date}. The {phase} process involved {n} response iterations coordinated across {regions} functional areas, with all correspondence and supporting analysis preserved in this archive entry.",
    "The whistleblower program assessment documented in this archive entry was conducted during the {phase} of ACPWB's governance engagement with {org}. Analysis of {n} reports submitted across {regions} reporting channels between {year} and {endyear} informed the structural and procedural recommendations included in the final deliverable.", "This archive entry records the {phase} phase of ACPWB's SPAC transaction advisory engagement for {org}. The engagement spanned {regions} months and produced {n} discrete deliverables covering target identification, due diligence, and de-SPAC integration planning for the {industry} sector.", "ESG reporting obligations applicable to {industry} sector organizations evolved materially during the {regions}-month period covered by this archive entry. ACPWB's {phase} team assisted {org} in mapping {n} new disclosure requirements across {regions} regulatory frameworks, with implementation timelines documented in the supporting project tracker.",
    "The proxy contest advisory engagement archived here was initiated at {org}'s request following the public announcement by a {industry} sector activist investor. During the {phase}, ACPWB analyzed {n} governance vulnerabilities across {regions} board and compensation dimensions and developed a comprehensive defense preparedness framework.", "This archive entry preserves the {phase} documentation from ACPWB's international expansion advisory engagement with {org}. The analysis covered {n} target markets across {regions} regions, assessing regulatory, labor, compensation, and governance considerations applicable to each {industry} sector operating environment.", "ACPWB's digital transformation advisory engagement with {org} encompassed {n} discrete workstreams during the {phase}, spanning technology architecture, workforce capability, operating model, and change management dimensions. The {industry} sector context for this transformation is detailed in the market analysis appendix attached to this archive record.",
    "The AI governance framework developed during the {phase} phase of the {org} engagement is preserved in this archive entry. ACPWB assessed {n} deployed and planned AI applications across {regions} functional domains, benchmarking {org}'s governance posture against emerging {industry} sector standards as of {date}.", "This archive entry documents the climate scenario analysis conducted for {org} during the {phase} phase of ACPWB's ESG advisory engagement. The analysis modeled {n} distinct climate pathways across {regions} physical and transition risk dimensions, producing a forward-looking risk register calibrated to the {industry} sector operating environment through {endyear}.", "The geopolitical risk assessment archived here was conducted at {org}'s request during the {phase}, following material disruption to {industry} sector supply chains across {regions} sourcing regions. ACPWB identified {n} exposure concentrations and developed a risk mitigation roadmap reviewed by {org}'s enterprise risk committee on {date}.",
    "This record captures the {phase} deliverables from ACPWB's internal investigation support engagement with {org}. In accordance with attorney-client privilege protocols, substantive findings are withheld from this archive entry; the record preserves only non-privileged process documentation spanning {n} weeks of fieldwork across {regions} jurisdictions.", "ACPWB's network analysis of {industry} sector board interlocks, conducted during the {phase} for {org}, identified {n} material overlapping director relationships across {regions} peer organizations. The analysis informed {org}'s director independence assessments and proxy disclosure strategy for the {year} annual meeting.", "The platform economics study documented in this archive entry assessed {org}'s competitive positioning within the {industry} sector ecosystem. During the {phase}, ACPWB modeled {n} network effect scenarios across {regions} user segments, producing a strategic framework for platform monetization and governance reviewed by the {org} executive committee on {date}.",
    "This archive entry preserves the {phase} documentation from ACPWB's regulatory arbitrage assessment for {org}. The analysis identified {n} jurisdictions within the {industry} sector regulatory landscape offering favorable treatment on {regions} dimensions of compliance cost, reporting burden, and licensing requirements.", "ACPWB's natural language processing analysis of {n} public disclosures by {industry} sector peers, conducted during the {phase} for {org}, identified systematic patterns in compensation narrative framing, ESG commitment language, and risk factor disclosure across {regions} annual reporting periods from {year} through {endyear}.", 'The deferred prosecution agreement monitoring engagement documented here was conducted by ACPWB on behalf of {org} during the {phase}. ACPWB assessed compliance with {n} monitor-required remediation milestones across {regions} functional areas and prepared the quarterly compliance attestation submitted to regulators on {date}.',
    "This archive entry records the {phase} outputs from ACPWB's benefits cost transformation engagement with {org}. The analysis benchmarked {n} benefit plan design parameters against {regions} peer organizations in the {industry} sector, producing a total cost-of-care reduction roadmap projected to achieve ${n}M in annual savings.", "ACPWB's {phase} analysis of {org}'s workforce analytics infrastructure assessed maturity across {n} capability dimensions spanning data, technology, governance, and talent. The {industry} sector benchmarking component drew on data from {regions} peer organizations and is preserved in full in this archive entry.", 'The SEC comment letter response documentation archived here records {n} rounds of correspondence between {org} and the Division of Corporation Finance during the {phase}. ACPWB provided technical support across {regions} disclosure topics, with all draft responses, supporting analyses, and final submissions indexed in this archive record.',
    "This archive entry captures the {phase} deliverables from ACPWB's human rights due diligence engagement with {org}. The assessment mapped {n} supply chain tiers across {regions} sourcing countries, identifying elevated risk concentrations in {industry} sector procurement categories as of {date}.", "The restructuring advisory engagement documented in this archive entry supported {org}'s {phase} organizational redesign in the {industry} sector. ACPWB modeled {n} structural scenarios across {regions} functional areas, producing a workforce impact analysis, communication framework, and change management roadmap approved by the {org} CEO on {date}.", "This archive entry preserves ACPWB's {phase} compensation committee education program materials developed for {org}. The program addressed {n} governance, market, and regulatory topics spanning {regions} sessions conducted between {year} and {endyear}, with all presentation materials, facilitator notes, and attendance records archived here.",
    "ACPWB's {phase} analysis of {industry} sector talent market dynamics identified {n} emerging workforce trends with material implications for {org}'s total rewards strategy. The findings, covering {regions} talent pools across {n} functional specializations, were presented to {org}'s CHRO and compensation committee on {date}.", "The insolvency-adjacent advisory engagement documented in this archive entry was initiated when {org} engaged ACPWB during the {phase} to assess {n} restructuring alternatives within the {industry} sector. ACPWB's analysis covered operational, financial, and workforce dimensions across {regions} business units, with all supporting models and memoranda preserved here.", "This archive entry documents ACPWB's {phase} support for {org}'s response to an EEOC systemic investigation. The engagement produced {n} workforce analytics deliverables covering hiring, promotion, and compensation equity across {regions} demographic dimensions, reviewed by {org}'s general counsel and external litigation counsel on {date}.",
    "The change-of-control preparedness assessment documented here was commissioned by {org}'s board during the {phase}, prompted by unsolicited acquisition interest from a {industry} sector strategic buyer. ACPWB evaluated {n} governance, compensation, and organizational readiness dimensions across {regions} functional areas, producing a board-ready briefing on {date}.", "ACPWB's {phase} engagement with {org} produced a comprehensive workforce data infrastructure assessment covering {n} HR systems across {regions} functional domains. The {industry} sector benchmarking component identified {n} technology capability gaps relative to peer organizations and a prioritized roadmap for remediation through {endyear}.", "This archive entry captures the {phase} deliverables from ACPWB's shareholder communication strategy engagement with {org}. The engagement produced {n} communication frameworks for {regions} investor constituencies, drawing on analysis of {industry} sector proxy season trends from {year} through {endyear} and preserved in full in this record.",
    "The operational resilience assessment documented in this archive entry evaluated {org}'s preparedness across {n} disruption scenarios during the {phase}. ACPWB benchmarked {org}'s resilience posture against {regions} {industry} sector peers, identifying {n} critical gaps with associated remediation priorities approved by the enterprise risk committee on {date}.", "This archive entry preserves ACPWB's {phase} documentation from a multi-client research initiative examining {industry} sector executive offboarding practices. Data was collected from {n} participating organizations across {regions} geographic markets, with {org} serving as one of {regions} co-sponsors of the research program.", "The talent marketplace platform assessment documented here evaluated {org}'s internal mobility infrastructure during the {phase}. ACPWB benchmarked {n} program design parameters against {regions} {industry} sector organizations and modeled the projected impact of platform investment on regrettable attrition and internal fill rates through {endyear}.",
    "ACPWB's {phase} ESG governance assessment for {org} evaluated board-level oversight structures across {n} environmental, social, and governance dimensions. The {industry} sector benchmarking component drew on data from {regions} peer organizations as of {date}, identifying {n} gaps relative to evolving institutional investor expectations.", "This archive entry documents the {phase} deliverables from ACPWB's pay transparency readiness assessment for {org}. The analysis covered {n} state and local pay transparency laws effective through {endyear}, assessing {org}'s compliance posture across {regions} compensation programs and producing a phased implementation roadmap.", "The organizational network analysis conducted during the {phase} mapped formal and informal collaboration patterns across {n} organizational nodes within {org}. ACPWB's {industry} sector benchmarking database provided reference points from {regions} peer organizations, informing a set of {n} targeted interventions to enhance cross-functional coordination.",
    "This archive entry captures ACPWB's {phase} engagement with {org} on succession planning for the top {n} executive roles. The engagement produced individual development plans, external market gap assessments, and an emergency succession protocol across {regions} critical positions, reviewed by the {org} board's compensation committee on {date}.", 'The consent decree compliance monitoring program documented here was administered by ACPWB on behalf of {org} during the {phase}. Monthly compliance reporting covering {n} remediation workstreams across {regions} functional areas was submitted to the relevant regulatory authority, with all correspondence and attestations archived in this record.', "ACPWB's {phase} workforce segmentation analysis for {org} identified {n} distinct employee cohorts in the {industry} sector, each characterized by differentiated engagement drivers, career aspirations, and total rewards preferences. The segmentation model, validated across {regions} focus groups, informed targeted retention and attraction strategies approved on {date}.",
    "This archive entry preserves the {phase} documentation from ACPWB's executive pay fairness review conducted for {org}'s independent directors. The review assessed the reasonableness of {n} compensation arrangements across {regions} dimensions of internal equity, market positioning, and pay-for-performance alignment within the {industry} sector context.", 'The cultural integration assessment documented here was conducted during the {phase} of the {org} post-merger integration. ACPWB administered a combined-company culture survey to {n} employees across {regions} legacy organizations, identifying {n} material cultural fault lines and a targeted integration roadmap approved by the combined leadership team on {date}.', "ACPWB's {phase} analysis of {industry} sector workforce cost structures identified {n} cost driver categories with material benchmarking implications for {org}. The analysis drew on proprietary data from {regions} comparable organizations and informed a total workforce cost optimization initiative projected to achieve a {n}% cost reduction over {regions} years.",
    "This archive entry documents the {phase} deliverables from ACPWB's director pay benchmarking engagement with {org}. The analysis covered cash retainer, equity award, committee chair, and lead director premium components across {n} proxy peers and {regions} governance survey respondents in the {industry} sector, as of {date}.", "The virtual workforce management assessment archived here evaluated {org}'s remote work policies, tools, and culture practices during the {phase}. ACPWB benchmarked {n} program parameters against {regions} {industry} sector organizations, producing a hybrid work design framework endorsed by {org}'s executive committee on {date}.", "ACPWB's {phase} regulatory readiness assessment for {org} evaluated compliance posture across {n} regulatory requirements anticipated to take effect between {date} and {endyear}. The {industry} sector analysis covered {regions} regulatory bodies and produced a prioritized implementation roadmap reviewed by the {org} audit committee.",
    "This archive entry captures the {phase} outputs from ACPWB's skills-based organization advisory engagement with {org}. The engagement produced a skills taxonomy covering {n} roles across {regions} functional domains and a workforce transition roadmap enabling the shift from job-based to skills-based talent management within the {industry} sector context.", "The executive team effectiveness assessment documented here was conducted by ACPWB during the {phase} at the request of {org}'s board. The assessment incorporated structured interviews with all {n} direct reports to the CEO, {regions} board members, and key external stakeholders, producing a confidential report delivered to the {org} lead director on {date}.", "ACPWB's {phase} analysis of {industry} sector ESG performance metrics identified {n} indicators with statistically significant correlation to long-term total shareholder return across {regions} calendar years from {year} through {endyear}. The analysis, drawing on data from {n} organizations, informed {org}'s executive incentive plan design for the forthcoming performance period.",
    "This archive entry preserves the {phase} documentation from ACPWB's board refreshment advisory engagement with {org}. The engagement identified {n} director skill gaps relative to the organization's strategic agenda, produced a board composition matrix benchmarked against {regions} {industry} sector peers, and supported the search process leading to {n} new director appointments.", "The workforce reduction planning documentation archived here covers the {phase} of ACPWB's restructuring advisory engagement with {org}. The deliverables include a workforce impact model covering {n} positions across {regions} locations, a separation package benchmarking analysis, a WARN Act compliance review, and a change communication framework approved on {date}.", "ACPWB's {phase} assessment of {org}'s total rewards communication effectiveness found that {n}% of employees in the {industry} sector survey could accurately describe the value of their compensation package. The finding, benchmarked against {regions} peer organizations, informed a targeted financial literacy and communication investment program launched on {date}.",
    "This archive entry documents the {phase} deliverables from ACPWB's regulatory capital benchmarking engagement with {org}. The analysis assessed {n} capital adequacy dimensions against {regions} peer organizations in the {industry} sector, drawing on both public disclosures and proprietary survey data as of {date}, with all supporting materials preserved here.", "The organizational effectiveness assessment documented in this archive entry evaluated {org}'s structure, governance, and ways of working during the {phase}. ACPWB benchmarked {n} organizational design parameters against {regions} {industry} sector peers, identifying structural inefficiencies and a redesign roadmap endorsed by the {org} CEO and presented to the board on {date}.", "ACPWB's {phase} assessment of {org}'s long-term incentive program evaluated plan design, performance metric selection, payout leverage, and competitive positioning across {n} dimensions. Drawing on data from {regions} {industry} sector peers benchmarked as of {date}, the analysis produced {n} design alternatives reviewed by the compensation committee.",
]

_ARCHIVE_METRIC_NAMES = [
    'Participating Organizations', 'Survey Response Rate (%)', 'Data Collection Cycle (weeks)',
    'Total Respondents', 'Completed Deliverables', 'Open Action Items', 'Jurisdictions Covered',
    'Benchmark Peers', 'Stakeholder Interviews Conducted', 'Document Pages Archived',
    'Review Cycles Completed', 'Subject Matter Experts Engaged', 'Advisory Hours Logged',
    'Database Records Indexed', 'Report Versions Produced', 'Findings Validated',
    'Recommendations Accepted', 'Implementation Rate (%)', 'Follow-On Engagements',
    'Cross-Sector Data Points', 'Regression Models Run', 'Statistical Confidence (%)',
    'Work Streams Active', 'Milestones Completed', 'Milestones Deferred',
    'Budget Utilization (%)', 'Scope Change Requests', 'Risk Items Identified',
    'Risk Items Resolved', 'Issues Escalated', 'Issues Resolved',
    'Stakeholders Briefed', 'Executive Sponsors Engaged', 'Board Members Briefed',
    'Data Sources Integrated', 'Survey Instruments Deployed', 'Focus Groups Conducted',
    'Structured Interviews Completed', 'Advisory Calls Logged', 'Client Approvals Received',
    'Deliverables On Schedule (%)', 'Quality Review Pass Rate (%)', 'Rework Hours',
    'Peer Organizations Surveyed', 'Custom Data Cuts Produced', 'Presentations Delivered',
    'Committee Appearances', 'Working Sessions Facilitated', 'Workshops Delivered',
    'Training Sessions Conducted', 'Participants Trained', 'Satisfaction Score (out of 5)',
    'Net Promoter Score', 'Days to Completion', 'Budget Variance ($K)', 'FTEs Engaged',
    'Contractor Hours', 'Technology Platforms Assessed', 'Vendors Evaluated',
    'Policies Reviewed', 'Policies Updated', 'Controls Tested', 'Exceptions Identified',
    'Exceptions Remediated', 'Compliance Rate (%)', 'Regulatory References Cited',
    # Additional metric names
    'Data Requests Fulfilled', 'Data Anomalies Detected', 'Data Anomalies Resolved',
    'Interviews Scheduled', 'Interviews Cancelled', 'No-Show Rate (%)',
    'Benchmark Organizations Added', 'Benchmark Organizations Removed',
    'Survey Waves Completed', 'Longitudinal Data Points', 'Cohort Organizations',
    'Regression Variables Tested', 'Models Validated', 'Hypotheses Tested',
    'Hypotheses Confirmed', 'Hypotheses Rejected', 'Outliers Excluded',
    'Data Imputation Events', 'Missing Data Rate (%)', 'Data Completeness (%)',
    'Engagement Duration (months)', 'Extensions Granted', 'Amendments Executed',
    'Subcontractors Engaged', 'Subcontractor Hours', 'Travel Days',
    'On-Site Visit Days', 'Remote Working Sessions', 'Async Reviews Completed',
    'Version Control Commits', 'Document Revisions', 'Review Comments Resolved',
    'Open Comments', 'Escalations Pending', 'Escalations Resolved',
    'Sponsor Touchpoints', 'Executive Briefings Delivered', 'Board Presentations',
    'Regulatory Submissions Filed', 'Regulatory Responses Received',
    'External Advisors Consulted', 'Legal Opinions Received', 'Expert Witnesses Identified',
    'Training Modules Developed', 'E-Learning Completions', 'Assessment Pass Rate (%)',
    'Certification Attainments', 'Accreditations Reviewed', 'Standards Mapped',
    'Gap Findings (Critical)', 'Gap Findings (Moderate)', 'Gap Findings (Low)',
    'Remediation Plans Drafted', 'Remediation Plans Approved', 'Remediation Actions Completed',
    'Action Items Opened', 'Action Items Closed', 'Overdue Action Items',
    'Deliverables Accepted', 'Deliverables Rejected', 'Deliverables Under Revision',
    'Invoice Milestones Hit', 'Budget Amendments', 'Contingency Utilized (%)',
    'Value Engineering Savings ($K)', 'Cost Avoidance Identified ($K)',
    'Benchmark Premium / Discount (%)', 'Quartile Position', 'Percentile Rank',
    'Index Score (0–100)', 'Composite Rating', 'Maturity Level (1–5)',
    'Net Benefit Realized ($K)', 'Payback Period (months)', 'IRR (%)',
    'NPV ($K)', 'Benefit-Cost Ratio', 'Efficiency Gain (%)',
    # ESG, technology/digital, procurement/supply chain, legal/compliance, financial, real estate/facilities, innovation
    'Scope 1 Emissions (metric tons CO2e)', 'Scope 2 Emissions (metric tons CO2e)', 'Scope 3 Emissions (metric tons CO2e)',
    'Carbon Intensity (per $M revenue)', 'Renewable Energy Usage (%)', 'Energy Consumption (MWh)',
    'Energy Intensity (MWh per $M revenue)', 'Water Withdrawal (megalitres)', 'Water Recycled (%)',
    'Waste Generated (metric tons)', 'Waste Diverted from Landfill (%)', 'Hazardous Waste Disposed (metric tons)',
    'Biodiversity Impact Sites Assessed', 'Supplier ESG Assessments Completed', 'Suppliers with Code of Conduct Acknowledgment (%)',
    'Community Investment ($M)', 'Volunteer Hours Logged', 'Social Impact Beneficiaries Reached',
    'ESG Audit Findings (Critical)', 'ESG Disclosure Score (0–100)', 'MSCI ESG Rating',
    'Sustainalytics Risk Score', 'CDP Climate Score', 'TCFD Alignment Level (1–4)',
    'GRI Standards Topics Disclosed', 'SASB Metrics Reported', 'Board ESG Oversight Mechanisms',
    'Executive ESG Pay Linkage (%)', 'System Uptime (%)', 'Mean Time to Recovery (hours)',
    'Mean Time Between Failures (hours)', 'Incident Response Time (hours)', 'Critical Vulnerabilities Remediated (%)',
    'Patch Compliance Rate (%)', 'Security Incidents Reported', 'Data Breaches Confirmed',
    'Phishing Click Rate (%)', 'Multi-Factor Authentication Adoption (%)', 'Cloud Migration Completion (%)',
    'Legacy System Retirement (%)', 'IT Spend as % of Revenue', 'Shadow IT Assets Identified',
    'API Integration Points Active', 'Data Quality Score (0–100)', 'Analytics Use Case Pipeline',
    'AI Models in Production', 'Automation Rate (%)', 'RPA Bot Utilization Rate (%)',
    'Digital Adoption Score', 'User Experience Score (NPS)', 'Application Portfolio Size',
    'Technical Debt Index', 'DevOps Deployment Frequency (per week)', 'Change Failure Rate (%)',
    'Lead Time for Changes (days)', 'Code Coverage (%)', 'Managed Spend ($M)',
    'Spend Under Management (%)', 'Procurement ROI', 'Savings as % of Managed Spend',
    'Supplier Count (Active)', 'Preferred Supplier Compliance (%)', 'Supplier Diversity Spend (%)',
    'On-Time Delivery Rate (%)', 'Purchase Order Cycle Time (days)', 'Contract Coverage (%)',
    'Maverick Spend (%)', 'Procurement Automation Rate (%)', 'Invoice Processing Cycle Time (days)',
    'Purchase Card Spend (%)', 'Supplier On-Boarding Cycle Time (days)', 'Single-Source Dependency (%)',
    'Supply Chain Risk Score', 'Supply Chain Disruption Events', 'Inventory Days on Hand',
    'Forecast Accuracy (%)', 'Perfect Order Rate (%)', 'Freight Cost per Unit',
    'Customs Compliance Rate (%)', 'Trade Finance Utilization (%)', 'Open Legal Matters',
    'Legal Spend ($M)', 'External Counsel Spend (%)', 'Litigation Reserves ($M)',
    'Regulatory Fines ($M)', 'Consent Orders Active', 'Corrective Action Plans Open',
    'Compliance Hotline Reports Received', 'Substantiated Reports (%)', 'Investigations Opened',
    'Investigations Closed', 'Policy Exceptions Granted', 'Training Non-Completion Rate (%)',
    'Regulatory Examinations Received', 'Examination Findings (Critical)', 'Deferred Prosecution Agreements Active',
    'Monitor Remediation Milestones Met (%)', 'Contract Renewal On-Time Rate (%)', 'Contractual SLA Breach Rate (%)',
    'IP Portfolio Assets', 'Patent Applications Filed', 'Economic Value Added ($M)',
    'Weighted Average Cost of Capital (%)', 'Net Debt ($M)', 'Net Leverage (x EBITDA)',
    'Liquidity Coverage Ratio', 'Net Stable Funding Ratio', 'Tangible Book Value ($M)',
    'Price-to-Tangible Book', 'Return on Tangible Common Equity (%)', 'Efficiency Ratio (%)',
    'Net Interest Margin (%)', 'Non-Performing Loan Ratio (%)', 'Loan Loss Provision ($M)',
    'Revenue per Client ($K)', 'Client Retention Rate (%)', 'Assets Under Management ($B)',
    'Net New Assets ($M)', 'Management Fee Rate (bps)', 'Performance Fee Revenue ($M)',
    'Total Shareholder Return (1-year %)', 'Total Shareholder Return (3-year %)', 'Total Occupied Space (sq ft)',
    'Space Utilization Rate (%)', 'Cost per Square Foot ($)', 'Lease Expiry Concentration (%)',
    'Real Estate Spend as % of Revenue', 'Average Lease Term Remaining (years)', 'Headquarters Sq Ft per Employee',
    'Desk-to-Employee Ratio', 'Meeting Room Utilization Rate (%)', 'Building Energy Intensity (kWh/sq ft)',
    'Capital Improvement Backlog ($M)', 'Maintenance Cost per Sq Ft ($)', 'Lease Renewal Rate (%)',
    'Sublease Vacancy Rate (%)', 'Real Estate Portfolio Sites', 'Owned vs. Leased Ratio (%)',
    'R&D Investment ($M)', 'R&D as % of Revenue', 'Patents Granted',
    'New Product Revenue (%)', 'Time to Market (months)', 'Innovation Pipeline Value ($M)',
    'Ideas Submitted (Internal)', 'Ideas Progressed to Pilot (%)', 'Pilot-to-Scale Conversion Rate (%)',
    'External Innovation Partners', 'Startup Co-Investment Count', 'Innovation Training Participants',
    'Cross-Functional Teams Active', 'Digital Revenue (%)', 'Data Monetization Revenue ($M)',
    'Platform API Calls (monthly)', 'Third-Party Developer Ecosystem Size',
]

_ARCHIVE_FINDING_TEMPLATES = [
    "{industry} sector organizations exhibit a {n}% variance in {metric} relative to cross-industry benchmarks established by ACPWB's Research Division in {year}.",
    "The {phase} phase identified {n} discrete improvement opportunities across the {org} engagement portfolio, with projected impact concentrated in {regions} core functional areas.",
    "Regression analysis of {n}-organization dataset reveals statistically significant correlation (p < 0.05) between governance maturity and {industry} sector compensation competitiveness.",
    "Peer benchmarking against {regions} comparator organizations indicates {org}'s current {metric} performance falls within the {n}th percentile of the relevant competitive set.",
    "Longitudinal analysis spanning {year}–{endyear} demonstrates consistent outperformance by {industry} organizations that adopted ACPWB's {phase} framework within 24 months of initial engagement.",
    "Survey data collected during the {phase} phase shows {n}% of {industry} sector respondents report material gaps in {metric} capabilities relative to stated strategic objectives.",
    "ACPWB's proprietary {industry} sector index, updated through {date}, indicates that {org}-category organizations outperform generalist peers by {n}% on composite governance measures.",
    "Cross-tabulation of {regions}-jurisdiction regulatory data with ACPWB's engagement outcomes database identifies {n} statistically anomalous findings warranting further investigation in the {industry} sector.",
    "The {org} engagement produced {n} actionable recommendations during the {phase} phase, of which {regions} have been fully implemented and {n} remain in progress as of the archival date.",
    "Organizations in the {industry} sector that completed ACPWB's {phase} assessment reported a median {n}% improvement in {metric} within 18 months of implementation.",
    "Variance analysis across {regions} peer organizations reveals that {industry} sector firms with mature {metric} practices outperform laggards by a factor of {n}% on total shareholder return over a three-year measurement window.",
    "The {phase} documentation reflects input from {n} senior stakeholders across {regions} functional areas within {org}, providing a comprehensive baseline for ongoing benchmarking.",
    "ACPWB's {industry} practice benchmarking database, which encompasses data from over {n} organizations, identifies {org} as a top-quartile performer on {regions} of the {n} dimensions assessed during the {phase} phase.",
    "Statistical significance testing (α = 0.01) confirms that the correlation between {metric} and long-term organizational performance observed in the {org} dataset is not attributable to sector-specific confounding variables present in {regions} comparison markets.",
    "The {phase} findings indicate that {industry} sector organizations with workforce populations exceeding {n} employees are disproportionately exposed to {metric} risk, with {regions}% of that cohort lacking formal mitigation protocols.",
    "Time-series analysis of ACPWB's {industry} engagement archive from {year} through {endyear} reveals a secular trend toward greater {metric} sophistication, with {n}% of organizations now exceeding the benchmark threshold established in {year}.",
    "External validation of the {phase} findings against {regions} independent data sources confirms the directional consistency of ACPWB's {metric} assessments, with an average deviation of less than {n}% from third-party estimates.",
    "The {org} {phase} engagement produced a composite risk score of {n} on ACPWB's 100-point {industry} sector assessment rubric, placing the organization in the {regions}th percentile relative to the full benchmark universe.",
    "Monte Carlo simulation incorporating {n} iterations of the {org} compensation model suggests that the proposed {metric} adjustments will achieve a 90% probability of reducing peer group positioning gap within {regions} months.",
    "Chi-square analysis of {industry} sector survey data (n={n}, df={regions}) yields a test statistic that rejects the null hypothesis at p < 0.001, confirming that {metric} outcomes differ significantly across {regions} organizational archetypes.",
    "The {phase} engagement produced a validated {industry} sector benchmark dataset incorporating {n} individual data points from {regions} participating organizations, representing the most comprehensive cross-sectional study completed by ACPWB since {year}.",
    "Factor analysis of {org}'s {phase} data isolates {regions} latent constructs that collectively explain {n}% of the observed variance in {metric}, providing a theoretically grounded basis for prioritizing improvement initiatives.",
    "Bootstrapped confidence intervals (95% CI) around the {org} {metric} estimate range from the {regions}th to {n}th percentile of the {industry} benchmark distribution, confirming the robustness of the point estimate to sampling variability.",
    "The {phase} phase surfaced {n} previously unquantified interdependencies between {org}'s {metric} posture and its broader {industry} sector competitive positioning, documented in full in the supporting technical appendix.",
    "Cluster analysis applied to the {n}-organization {industry} benchmark dataset identifies {regions} statistically distinct archetypes, with {org} most closely aligned to the archetype characterized by high {metric} investment and above-median governance maturity scores.",
    "The {phase} engagement for {org} produced a return-on-investment estimate of {n}x over a {regions}-year horizon, based on ACPWB's standardized benefit quantification methodology applied across {industry} sector organizations since {year}.",
    "Structural equation modeling of the {org} {phase} dataset confirms a direct path coefficient of {n}% between {metric} investment and organizational performance outcomes, controlling for {regions} industry-level confounders identified through the literature review.",
    "A/B comparison of {org} divisions that completed ACPWB's {phase} assessment versus those that did not reveals a {n}% performance differential on {metric}, a finding consistent across {regions} analogous {industry} sector comparisons in ACPWB's historical archive.",
    "The {industry} sector composite index constructed during the {phase} phase incorporates {n} weighted input variables across {regions} performance domains. {org}'s composite score of {n} places it in the top quartile of the {n}-organization reference group on an absolute basis.",
    "Propensity score matching applied to the {industry} engagement dataset produces a treatment effect estimate of {n}% improvement in {metric} attributable to the {phase} intervention, with a confidence interval that excludes zero at p < 0.01.",
    "ACPWB's {phase} benchmark report for the {industry} sector, incorporating data from {n} organizations across {regions} geographic markets, identifies {org} as an exemplar on {metric}, with a best-practice profile detailed in Appendix C of the supporting research record.",
    "The {org} {phase} findings were validated against ACPWB's longitudinal panel dataset, which tracks {metric} outcomes for {n} organizations across {regions} annual survey waves. {org}'s trajectory is consistent with organizations that outperform sector medians within {n} years of completing the {phase}.",
    "Geospatial analysis of the {industry} sector {phase} data reveals significant regional variation in {metric}, with organizations in {regions} geographic clusters outperforming the national median by {n}%. {org}, headquartered in a high-performing cluster, benefits from structural advantages not fully captured in standard benchmarking.",
    "The {phase} risk heat map developed for {org} maps {n} identified risks across {regions} dimensions of likelihood and impact. Three risks were elevated to the executive steering committee for direct oversight, with mitigation plans requiring {n} months to implement across the {industry} sector operating environment.",
    "Sensitivity analysis of the {org} {phase} model indicates that a {n}% change in key {metric} assumptions produces a {regions}% shift in projected outcomes, suggesting that conclusions are robust within a reasonable range of scenario assumptions applicable to the {industry} sector.",
    "The {phase} documentation reflects a statistically significant improvement in {metric} between the baseline assessment conducted at engagement outset and the follow-on measurement taken {n} months later, with an effect size of {regions} standard deviations relative to the {industry} benchmark distribution.",
    "Panel regression applied to ACPWB's {industry} archive from {year} through {endyear} isolates the effect of {phase} interventions on {metric}, producing a fixed-effects estimate of {n}% improvement per engagement cycle, controlling for organization size, geography, and {regions} time-varying covariates.",
    "The {phase} engagement identified {n} organizations within the {industry} sector whose {metric} practices qualify as emerging best practices not yet represented in published benchmarking literature. {org} was among {regions} clients granted early access to these unpublished findings.",
    "Discriminant function analysis of the {industry} sector {phase} dataset correctly classifies {n}% of organizations into high-, medium-, and low-performing {metric} groups, demonstrating the predictive validity of ACPWB's assessment instrument across {regions} organizational characteristics.",
    "The {org} board received a presentation summarizing the {phase} findings on {date}, at which time {n} of {regions} recommendations were approved for immediate implementation. The remaining recommendations are subject to a {n}-month feasibility review prior to board decision.",
    "ACPWB's {industry} sector early-warning model, calibrated on data from {n} organizations over {regions} survey cycles, correctly flagged {org}'s {metric} risk exposure {n} months before it became material — validating the model's predictive utility for {industry} sector governance applications.",
    "The {phase} deliverable package for {org} included a customized {industry} sector scorecard covering {n} performance indicators, each benchmarked against {regions} comparator organizations and presented with trend data spanning {year} through {endyear}.",
    "ACPWB's analysis of {industry} sector {phase} outcomes reveals that organizations completing the engagement during periods of macroeconomic expansion show {n}% stronger {metric} improvement than those completing during contractionary periods — a finding that contextualizes {org}'s results relative to the {regions} peer organizations assessed in the same cycle.",
    "The {phase} archive entry includes a meta-analysis of {n} prior ACPWB engagements in the {industry} sector, synthesizing evidence on {metric} improvement rates across {regions} intervention types and providing the empirical foundation for the recommendations delivered to {org}.",
    "Cross-national comparison conducted during the {phase} phase reveals that {industry} sector organizations headquartered in {regions} jurisdictions outperform domestic peers by {n}% on {metric}, a gap attributable to structural differences in regulatory environment, labor market conditions, and board composition practices.",
    "The {org} {phase} engagement incorporated real-time benchmarking against {n} organizations that completed the same assessment within the prior {regions} months, enabling ACPWB to contextualize {org}'s preliminary findings within the most current available {industry} sector data.",
    "ACPWB's {phase} process improvement analysis for {org} identified {n} workflow redundancies across {regions} functional areas, with an estimated annualized labor cost savings of ${n}K achievable through targeted process redesign — findings validated against outcomes achieved by {regions} prior {industry} sector clients.",
    # NLP/text mining, network analysis, platform economics, geopolitical risk, climate scenarios, regulatory arbitrage, probit regression, whistleblower programs, organizational network analysis, mixed-methods, digital maturity, AI governance, diff-in-diff, skills gap, ESG materiality, survival analysis, supply chain risk, Bayesian networks, executive team effectiveness, conjoint analysis, labor econometrics, say-on-pay, ESG-TSR correlation, latent class analysis, fraud risk, total rewards benchmarking, IV analysis, workforce segmentation, change readiness, proxy peer optimization, internal equity compression, regulatory enforcement trends, scenario compensation modeling, consensus building, workforce analytics trends, variance decomposition, board effectiveness, equity simulation, qualitative content, pay-for-performance, structural break, culture assessment, meta-regression, exit interview, pay equity regression, governance proxy, engagement factor analysis, quantile regression
    "Natural language processing analysis of {n} public proxy disclosures across {industry} sector organizations from {year} through {endyear} reveals a statistically significant shift in compensation narrative framing, with {regions}% of organizations now employing performance-aligned language consistent with {org}'s {phase} disclosure strategy.", "Network analysis of cross-board director relationships within the {industry} sector identified {n} material interlocking directorships affecting {regions} of the {n} proxy peers assessed during the {phase}, with implications for {org}'s director independence disclosures that are quantified in the supporting technical memorandum.", 'Platform economics modeling conducted during the {phase} engagement for {org} demonstrates network effects generating a {n}% annual retention premium for established {industry} sector participants relative to new entrants, with the magnitude of the effect correlated with platform scale across {regions} user segments.',
    'Geopolitical risk scenario analysis conducted for {org} during the {phase} assigns a {n}% probability-weighted impact to {industry} sector supply chain disruption over the {regions}-year horizon, with geographic concentration risk identified as the primary driver across {n} sourcing categories.', "Climate scenario analysis calibrated to the {industry} sector and applied during the {phase} projects a {n}% increase in {org}'s physical risk exposure under a 2°C warming scenario by {endyear}, with the most material impacts concentrated in {regions} operating regions identified through ACPWB's geospatial risk mapping tool.", "ACPWB's regulatory arbitrage analysis conducted during the {phase} for {org} quantified {n} compliance cost differentials across {regions} jurisdictions in the {industry} sector, identifying a ${n}M annualized compliance cost reduction opportunity achievable through targeted legal entity restructuring.",
    'Probit regression of {industry} sector governance data (n={n}) from {year} through {endyear} indicates that organizations adopting lead director structures were {n}% less likely to receive a majority against say-on-pay vote, controlling for {regions} compensation-related governance factors examined during the {phase}.', "The whistleblower program benchmarking conducted during the {phase} reveals that {industry} sector organizations with anonymous reporting channels receive {n}% more substantiated reports per 1,000 employees than those relying solely on identified reporting, with {regions}-year trend data confirming {org}'s program falls below the benchmark median.", "ACPWB's text mining of {n} earnings call transcripts from {industry} sector peers from {year} through {endyear} identified {regions} thematic clusters predictive of {org}'s {metric} trajectory within {n} quarters, providing an early-warning signal validated against historical outcomes during the {phase}.",
    'The organizational network analysis conducted for {org} during the {phase} mapped collaboration patterns across {n} nodes, revealing {regions} structural holes in cross-functional information flow that correlate with a {n}% reduction in decision velocity relative to high-performing {industry} sector benchmarks.', 'Mixed-methods analysis combining {n} quantitative survey responses with {regions} structured executive interviews conducted during the {phase} identifies psychological safety as the dominant predictor of {metric} in {industry} sector organizations, explaining {n}% of the observed variance in the {org} leadership effectiveness assessment.', "The digital transformation maturity assessment conducted for {org} during the {phase} benchmarks the organization at Level {n} on ACPWB's {regions}-point {industry} sector maturity scale, with capability gaps most pronounced in the data and analytics domain relative to the {n}-organization comparison group.",
    "ACPWB's {phase} analysis of {industry} sector AI governance frameworks finds that {n}% of organizations lack board-level oversight mechanisms for AI risk, a gap correlated with {n}-point lower ESG composite scores across the {regions}-organization dataset from {year} through {endyear}.", 'Difference-in-differences analysis exploiting variation in the timing of {phase} interventions across {n} {industry} sector organizations confirms a treatment effect of {n}% improvement in {metric} attributable to the {org} engagement design, robust to {regions} alternative control group specifications.', 'The skills gap analysis conducted during the {phase} for {org} identified {n} critical skill deficiencies across {regions} functional domains, with the largest gaps concentrated in data literacy, agile delivery, and human-centered design — skills increasingly central to competitive positioning in the {industry} sector as of {date}.',
    "ACPWB's {phase} ESG materiality assessment for {org} ranked {n} potential disclosure topics by stakeholder salience and business impact, with {regions} topics meeting the dual materiality threshold applicable under emerging {industry} sector reporting standards for the period ending {endyear}.", "Survival analysis of {n} {industry} sector organizations in ACPWB's longitudinal panel from {year} through {endyear} finds that organizations completing the {phase} assessment are {n}% less likely to experience a material governance failure within {regions} years, after controlling for size, leverage, and board composition.", 'The supply chain risk mapping conducted for {org} during the {phase} identified {n} single-source dependencies across {regions} strategic procurement categories, with {industry} sector benchmarking indicating that top-quartile organizations maintain no more than {n}% of managed spend in single-source arrangements.',
    'Bayesian network modeling of the {org} {phase} dataset reveals conditional dependencies among {n} organizational capability dimensions that are obscured in standard correlation analyses, providing a more precise basis for intervention prioritization across the {industry} sector engagement portfolio.', 'The executive team effectiveness assessment conducted during the {phase} found a {n}-point gap between CEO-rated and direct-report-rated decision quality at {org}, a gap consistent with the {n}th percentile of the {industry} sector reference group assembled from {regions} prior ACPWB assessments.', "ACPWB's conjoint analysis of {n} total rewards preference scenarios, administered to {industry} sector employees during the {phase} for {org}, isolates the relative value weights attributed to base salary, annual incentive, long-term equity, and benefits — findings that diverge materially from {org}'s current allocation across {regions} employee segments.",
    'The labor market econometric model calibrated during the {phase} for {org} incorporates {n} macroeconomic and industry-specific predictors of {industry} sector talent supply and demand from {year} through {endyear}, producing a {regions}-year workforce cost forecast that outperforms naive benchmarking by {n}% on backtested accuracy.', "ACPWB's {phase} analysis of {org}'s historical say-on-pay vote patterns — spanning {year} through {endyear} and benchmarked against {n} {industry} sector peers — identifies {regions} compensation design features with statistically significant explanatory power for relative ISS and Glass Lewis vote outcomes.", "The ESG performance correlation study conducted for {org} during the {phase} finds a {n}% premium in three-year total shareholder return for {industry} sector organizations in the top quartile of ACPWB's composite ESG score, with the relationship strongest among large-cap peers in {regions} geographic markets.",
    "Latent class analysis applied to the {industry} sector survey dataset (n={n}) identifies {regions} distinct governance archetypes, each with characteristic profiles on {metric}, board composition, and executive pay structure. {org}'s archetype assignment and its strategic implications are detailed in the {phase} supporting appendix.", 'The fraud risk assessment conducted during the {phase} for {org} identified {n} control gaps across {regions} financial reporting processes, with {industry} sector benchmarking indicating that organizations at comparable maturity levels typically maintain fewer than {n} open fraud risk items at any point in time.', "ACPWB's {phase} total rewards benchmarking for {org} reveals a {n}% market positioning gap at the {regions}th percentile of the competitive set across {n} {industry} sector benchmark surveys, with the gap most pronounced at the senior director and vice president levels in {regions} functional areas.",
    'Instrumental variable analysis applied to the {org} {phase} dataset addresses selection bias in voluntary {industry} sector program participation, producing a local average treatment effect estimate of {n}% improvement in {metric} that is {n}% larger than the naive OLS estimate — a finding with direct implications for program investment decisions across {regions} peer organizations.', 'The workforce segmentation model developed for {org} during the {phase} assigns {n} employee personas to {regions} distinct clusters based on {n} attitudinal and demographic variables, with each cluster exhibiting statistically distinct patterns of {metric} and turnover risk that inform targeted retention investment in the {industry} sector.', "ACPWB's {phase} change readiness assessment for {org} found that {n}% of employees across {regions} organizational levels report high confidence in the {industry} sector transformation agenda, a figure {n} percentage points below the benchmark median — a finding that informed the design of targeted leadership communication and manager enablement programs launched on {date}.",
    'The proxy peer group optimization analysis conducted for {org} during the {phase} evaluated {n} potential peer additions and deletions across {regions} selection criteria, producing a revised {n}-company peer group that increases the proportion of compensation-at-risk relative to the legacy peer set while maintaining alignment with {industry} sector size norms.', "ACPWB's {phase} analysis of {org}'s internal equity posture identified {n} role clusters exhibiting compa-ratio compression in the {regions}th to {n}th percentile range, with the compression concentrated among mid-career employees in high-demand {industry} sector specializations — a pattern validated against {regions} peer organization datasets.", "The regulatory enforcement trend analysis conducted for {org} during the {phase} catalogues {n} enforcement actions against {industry} sector peers from {year} through {endyear}, identifying {regions} recurring violation categories and developing an exposure prioritization framework calibrated to {org}'s operational profile as of {date}.",
    "ACPWB's {phase} scenario-based compensation modeling for {org} simulates pay outcomes across {n} performance scenarios, demonstrating that the proposed long-term incentive design produces payout leverage within the {regions}th to {n}th percentile range of the {industry} sector peer group under the median performance scenario.", 'The consensus-building methodology applied during the {phase} for {org} employed structured dialogue protocols to resolve {n} areas of disagreement among senior stakeholders across {regions} organizational units, achieving documented consensus on {n} priority actions within a {regions}-week facilitation cycle benchmarked against {industry} sector norms.', "ACPWB's {phase} analysis of {industry} sector workforce data from {year} through {endyear} finds a secular increase in the tenure of high-performing employees, with organizations demonstrating above-median {metric} retaining top performers for an average of {n} additional months relative to the benchmark — a trend that informs {org}'s long-term retention investment strategy.",
    'The three-way variance decomposition applied to the {org} {phase} compensation dataset partitions observed pay variation into {n}% attributable to job grade, {regions}% to performance rating, and the remainder to individual negotiation and legacy factors — a finding that benchmarks unfavorably against {industry} sector best practices on pay equity.', "ACPWB's {phase} board effectiveness survey administered to {n} directors at {org} yields a composite effectiveness score in the {regions}th percentile of the {industry} sector reference group, with the lowest-scoring dimensions concentrated in strategic oversight and succession planning — areas confirmed as priorities in {n} subsequent follow-up interviews.", 'Simulation modeling conducted for {org} during the {phase} demonstrates that the proposed equity refresh program, calibrated to the {regions}th percentile of {industry} sector practice on {metric}, will reduce projected senior leader attrition by {n}% over the {regions}-year measurement horizon at an incremental annualized cost of ${n}M.',
    'The qualitative content analysis of {n} stakeholder interviews conducted during the {phase} for {org} identifies {regions} recurring themes regarding {metric} that are absent from quantitative survey instruments, providing an analytical basis for the targeted capability investments recommended to the {industry} sector client on {date}.', "ACPWB's {phase} analysis of {org}'s pay-for-performance alignment demonstrates a Pearson correlation of {n}% between individual performance ratings and actual incentive payouts across {regions} fiscal years, a figure that trails the {industry} sector median by {n} percentage points and informs the compensation committee's design priorities for the forthcoming plan year.", "The structural break test applied to ACPWB's {industry} sector longitudinal dataset from {year} through {endyear} identifies {n} statistically significant breaks in the {metric} trend series, with the most recent break in {date} coinciding with the onset of macroeconomic tightening — a finding with direct implications for {org}'s {phase} benchmarking interpretation.",
    "ACPWB's {phase} culture assessment for {org} administered a validated {n}-item instrument to {regions} employees, producing a composite culture score of {n} on a 100-point scale that places the organization in the {regions}th percentile of the {industry} sector reference group — below the threshold for the high-performance culture designation on {regions} of the {n} subdimensions assessed.", 'The meta-regression analysis conducted for {org} during the {phase} synthesizes evidence from {n} published and unpublished studies on {metric} effectiveness in the {industry} sector, producing a pooled effect size estimate of {n}% improvement per unit of investment that is robust to publication bias correction across {regions} moderator variables.', "ACPWB's {phase} exit interview analysis for {org}, covering {n} departing employees across {regions} functional areas from {year} through {endyear}, identifies compensation competitiveness as the primary driver of regrettable attrition in {n}% of cases — a finding {n} percentage points higher than the {industry} sector benchmark and central to the total rewards redesign recommendation.",
    'The pay equity regression analysis conducted for {org} during the {phase} finds a residual gender pay gap of {n}% after controlling for job grade, performance rating, tenure, and {regions} additional explanatory variables — a gap that is statistically significant at p < 0.05 across {n} of the {regions} {industry} sector business units examined.', "ACPWB's {phase} analysis of {industry} sector corporate governance disclosures from {year} through {endyear} finds that organizations with formal board evaluation processes report {n}% fewer governance-related shareholder proposal submissions, with the effect concentrated among the {regions} most active institutional investors represented in the {org} shareholder base.", "The factor analysis of {org}'s {phase} employee engagement survey data isolates {n} latent dimensions that account for {regions}% of the total variance in engagement scores across the {industry} sector benchmark group, providing an empirically grounded model for prioritizing the {n} engagement initiatives recommended for the {year} action planning cycle.",
    'Quantile regression applied to the {industry} sector compensation database during the {phase} reveals that the pay-for-performance relationship for {org} is stronger at the upper tail of the pay distribution (90th percentile coefficient: {n}%) than at the median, a pattern consistent with tournament theory predictions validated across {regions} peer organizations.',
]

_ARCHIVE_METRIC_LABELS = [
    'base compensation', 'total direct compensation', 'total target compensation',
    'total realized compensation', 'long-term incentive value', 'annual incentive payout rate',
    'pay equity ratio', 'compa-ratio', 'pay range penetration', 'market pricing accuracy',
    'workforce retention rate', 'voluntary turnover rate', 'involuntary turnover rate',
    'regrettable attrition rate', 'internal promotion rate', 'internal mobility rate',
    'manager effectiveness score', 'leadership effectiveness index', 'engagement index',
    'employee net promoter score', 'inclusion index', 'belonging score',
    'benefits cost per employee', 'benefits utilization rate', 'healthcare trend rate',
    'retirement plan participation rate', 'retirement plan match utilization',
    'training hours per FTE', 'learning investment per employee', 'certification completion rate',
    'span of control ratio', 'management layers', 'organizational flatness index',
    'time-to-fill (days)', 'offer acceptance rate', 'sourcing channel effectiveness',
    'cost per hire', 'quality of hire score', 'new hire retention rate',
    'governance maturity score', 'board effectiveness rating', 'audit quality index',
    'compliance incident rate', 'policy adherence rate', 'control effectiveness score',
    'ESG composite score', 'carbon intensity', 'diversity representation index',
    'pay transparency readiness score', 'HR technology adoption rate',
    'data quality index', 'reporting cycle time (days)', 'analytics maturity level',
    'total workforce cost as % of revenue', 'HR cost per employee', 'span-adjusted labor productivity',
    'revenue per FTE', 'operating income per FTE', 'human capital ROI',
    # Additional metric labels
    'total compensation cost as % of revenue', 'benefits cost as % of total comp',
    'incentive payout as % of target', 'equity dilution rate (%)',
    'share-based compensation expense ($M)', 'grant date fair value per share',
    'option exercise rate (%)', 'underwater option percentage (%)',
    'performance share vesting rate (%)', 'PSU modifier range',
    'relative TSR ranking', 'absolute TSR (3-year %)',
    'adjusted EPS growth (%)', 'revenue growth (%)', 'EBITDA margin (%)',
    'free cash flow yield (%)', 'ROIC (%)', 'return on equity (%)',
    'return on assets (%)', 'economic value added ($M)', 'cost of capital (%)',
    'debt-to-EBITDA ratio', 'net leverage ratio', 'interest coverage ratio',
    'working capital days', 'days sales outstanding', 'days payable outstanding',
    'inventory turns', 'asset turnover ratio', 'capital intensity (%)',
    'R&D spend as % of revenue', 'SG&A as % of revenue', 'COGS as % of revenue',
    'gross margin (%)', 'operating margin (%)', 'net margin (%)',
    'headcount growth rate (%)', 'workforce age median (years)',
    'average tenure (years)', 'average tenure at grade (years)',
    'time-to-productivity (days)', 'ramp time to full performance (months)',
    'internal hire rate (%)', 'succession coverage ratio (%)',
    'bench strength index', 'critical role vacancy rate (%)',
    'high-performer retention rate (%)', 'high-potential attrition rate (%)',
    'top-quartile performer share (%)', 'performance rating forced distribution',
    'PIP completion rate (%)', 'PIP success rate (%)',
    'involuntary termination rate (%)', 'mutual separation rate (%)',
    'retirement-eligible workforce (%)', 'flight risk score',
    'absence rate (%)', 'unplanned absence rate (%)', 'FMLA utilization rate (%)',
    'parental leave take-up rate (%)', 'flexible work arrangement adoption (%)',
    'remote work eligibility (%)', 'remote work utilization (%)',
    'employee resource group participation (%)', 'mentoring program enrollment (%)',
    'sponsorship program completion rate (%)',
    'internal mobility application rate (%)', 'job shadow participation (%)',
    'upskilling investment per employee ($)', 'external hire premium (%)',
    'referral hire rate (%)', 'diversity of referral hires (%)',
    'campus hire as % of total hires (%)', 'lateral hire rate (%)',
    'counteroffer acceptance rate (%)', 'regret rehire rate (%)',
    'HR business partner ratio (employees per HRBP)',
    'recruiter productivity (hires per recruiter per year)',
    'HR system adoption rate (%)', 'payroll accuracy rate (%)',
    'benefits open enrollment completion rate (%)', 'HSA contribution rate (%)',
    '401(k) average deferral rate (%)', 'pension funded status (%)',
    'OPEB liability ($M)', 'workers compensation incidence rate',
    'OSHA recordable rate', 'lost time injury rate', 'near-miss reporting rate',
    'ethics hotline utilization rate (per 1,000 employees)',
    'substantiated complaint rate (%)', 'code of conduct training completion (%)',
    'data privacy training completion (%)', 'anti-bribery training completion (%)',
    'audit finding repeat rate (%)', 'material weakness count', 'significant deficiency count',
    'SOX control failure rate (%)', 'IT general control effectiveness (%)',
    'vendor risk assessment completion rate (%)', 'third-party incidents reported',
]


_ARCHIVE_TITLE_PREFIXES = [
    'Summary Report:', 'Engagement Documentation:', 'Final Analysis:',
    'Internal Memorandum:', 'Phase Completion Report:', 'Working Document:',
    'Archived Deliverable:', 'Research Record:', 'Reference Document:',
    'Engagement Archive:', 'Client Record:', 'Project Documentation:',
    # Expanded title prefixes
    'Advisory Memorandum:', 'Preliminary Assessment:', 'Draft Findings:',
    'Interim Report:', 'Preliminary Memorandum:', 'Draft Memorandum:',
    'Preliminary Findings:', 'Preliminary Report:', 'Interim Memorandum:',
    'Interim Assessment:', 'Draft Assessment:', 'Preliminary Summary:',
    'Final Memorandum:', 'Final Report:', 'Final Summary:',
    'Final Documentation:', 'Completed Analysis:', 'Completed Assessment:',
    'Completed Deliverable:', 'Approved Report:', 'Approved Analysis:',
    'Issued Report:', 'Signed-Off Documentation:', 'Accepted Deliverable:',
    'Closed-Out Record:', 'Analytical Summary:', 'Technical Memorandum:',
    'Technical Report:', 'Technical Analysis:', 'Technical Brief:',
    'Analytical Brief:', 'Analytical Memorandum:', 'Analytical Record:',
    'Research Summary:', 'Research Analysis:', 'Research Brief:',
    'Research Memorandum:', 'Advisory Brief:', 'Advisory Report:',
    'Advisory Summary:', 'Advisory Record:', 'Governance Brief:',
    'Governance Memorandum:', 'Governance Report:', 'Governance Summary:',
    'Governance Documentation:', 'Board Materials:', 'Board Brief:',
    'Board Memorandum:', 'Committee Documentation:', 'Committee Brief:',
    'Committee Memorandum:', 'Status Update:', 'Status Report:',
    'Status Memorandum:', 'Progress Update:', 'Progress Report:',
    'Progress Memorandum:', 'Milestone Report:', 'Milestone Summary:',
    'Milestone Documentation:', 'Operational Summary:', 'Operational Brief:',
    'Operational Memorandum:', 'Deliverable Archive:', 'Deliverable Record:',
    'Deliverable Documentation:', 'Planning Documentation:', 'Planning Summary:',
    'Planning Memorandum:', 'Strategic Summary:', 'Strategic Memorandum:',
    'Strategic Brief:', 'Strategic Documentation:', 'Strategic Record:',
    'Initiative Summary:', 'Initiative Brief:', 'Initiative Documentation:',
    'Initiative Record:', 'Compliance Summary:', 'Compliance Documentation:',
    'Compliance Memorandum:', 'Audit Documentation:', 'Audit Summary:',
    'Audit Memorandum:', 'Audit Brief:', 'Audit Record:',
    'Control Documentation:', 'Control Summary:', 'Review Documentation:',
    'Review Summary:', 'Review Memorandum:', 'Review Record:',
    'Review Brief:', 'Client Summary:', 'Client Brief:',
    'Client Memorandum:', 'Client Documentation:', 'Engagement Summary:',
    'Engagement Brief:', 'Engagement Memorandum:', 'Engagement Report:',
    'Engagement Record:', 'Project Summary:', 'Project Brief:',
    'Project Memorandum:', 'Project Record:', 'Project Report:',
    'Program Summary:', 'Program Brief:', 'Program Memorandum:',
    'Program Documentation:', 'Findings Summary:', 'Findings Report:',
    'Findings Memorandum:', 'Findings Documentation:', 'Findings Record:',
    'Findings Brief:', 'Recommendations Summary:', 'Recommendations Brief:',
    'Recommendations Memorandum:', 'Recommendations Documentation:', 'Recommendations Report:',
    'Recommendations Record:', 'Assessment Summary:', 'Assessment Brief:',
    'Assessment Memorandum:', 'Assessment Documentation:', 'Assessment Record:',
    'Assessment Report:', 'Due Diligence Documentation:', 'Due Diligence Summary:',
    'Due Diligence Memorandum:', 'Transaction Documentation:', 'Transaction Summary:',
    'Transaction Memorandum:', 'Regulatory Correspondence:', 'Regulatory Summary:',
    'Regulatory Memorandum:', 'Regulatory Documentation:', 'Regulatory Record:',
    'Regulatory Brief:', 'Executive Summary:', 'Executive Brief:',
    'Executive Memorandum:', 'Executive Documentation:', 'Confidential Memorandum:',
    'Confidential Summary:', 'Confidential Brief:', 'Confidential Documentation:',
    'Restricted Memorandum:', 'Restricted Documentation:', 'Privileged Memorandum:',
    'Privileged Summary:',
]




_ARCHIVE_YEAR_DATA = {
    1985: {
        'theme': 'The Founding Era',
        'desc': "ACPWB's inaugural year: a small team of six compensation researchers operating out of a rented Milwaukee office suite, convinced that data-driven pay equity was not a niche concern but an inevitability.",
        'bg': '#F5F0E8',
        'text_color': '#2B1A0E',
        'accent': '#8B5E3C',
        'accent2': '#D4A96A',
        'font_body': "'Courier New', Courier, monospace",
        'font_head': "'Times New Roman', Times, serif",
        'layout_class': 'era-founding',
        'ceo': 'Chester H. Whitmore III',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients, collaborators, and the broader compensation research community:

This inaugural letter marks what I sincerely believe will be the beginning of something consequential. The American Corporation for Public Well Being opened its doors in Milwaukee this past spring with six staff, three active client engagements, and a conviction that compensation research — done rigorously and without ideological shortcuts — represents a meaningful contribution to American economic life.

We established ourselves amid a period of broad optimism. The economy has recovered substantially from the hardships of the early decade, and organizations across the industrial and services sectors are investing once again in human capital. What strikes us most, in these early months, is how profoundly underserved the compensation benchmarking market remains. Too many organizations still rely on informal surveys, outdated government tables, and anecdotal industry chatter to make decisions that affect hundreds or thousands of workers.

Our methodology is not yet complete. I will not pretend otherwise. But the analytical foundation we have constructed during these first months — a normalized compensation index, a stratified peer-comparison model, sector-specific adjustment factors — represents genuine progress. We expect to refine these tools substantially in the years ahead.

I thank the six individuals who joined this enterprise with limited guarantees and considerable faith, and the three clients who trusted us before our track record was anything more than a promise.

Chester H. Whitmore III
President & Chief Executive Officer""",
    },
    1986: {
        'theme': 'The Founding Era',
        'desc': 'The second year: client roster doubles, methodology refined, first published benchmark index. ACPWB earns its first industry mention in Compensation & Benefits Review.',
        'bg': '#F5F0E8',
        'text_color': '#2B1A0E',
        'accent': '#8B5E3C',
        'accent2': '#D4A96A',
        'font_body': "'Courier New', Courier, monospace",
        'font_head': "'Times New Roman', Times, serif",
        'layout_class': 'era-founding',
        'ceo': 'Chester H. Whitmore III',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

Our second year of operation has validated, with more confidence than I anticipated, the initial thesis on which this organization was founded. The compensation benchmarking market is large, structurally underserved, and populated by organizations that genuinely want better data than they currently have access to. We have grown our active client engagements from three to seven, and our Milwaukee staff from six to eleven.

The work itself has grown more technically demanding in productive ways. We have now published the first edition of our Normalized Sector Compensation Index — a document I hope will become a standard reference for practitioners in this field. The methodology, which relies on a multi-stage weighting procedure to control for organizational size, geography, and industry subsector, produced results this year that surprised even our own internal expectations. The variance between reported and actual total compensation across surveyed firms was nearly thirty percent — a finding that has already generated significant client interest.

We have also expanded our geographic data collection beyond the upper Midwest, which limits the conclusions we can draw but enriches our benchmarks considerably. I continue to believe that regional compensation variance is among the most underappreciated factors in workforce planning, and I expect this thread to run through our research agenda for years to come.

Chester H. Whitmore III
President & Chief Executive Officer""",
    },
    1987: {
        'theme': 'The Founding Era',
        'desc': 'Black Monday shockwave reaches every compensation engagement. ACPWB\'s countercyclical relevance is established: when budgets freeze, the demand for benchmarking intensifies.',
        'bg': '#F5F0E8',
        'text_color': '#2B1A0E',
        'accent': '#8B5E3C',
        'accent2': '#D4A96A',
        'font_body': "'Courier New', Courier, monospace",
        'font_head': "'Times New Roman', Times, serif",
        'layout_class': 'era-founding',
        'ceo': 'Chester H. Whitmore III',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

October 19th, 1987 — Black Monday — will be remembered as the day the American financial system absorbed a shock it had not anticipated and, frankly, had not prepared for. The Dow Jones fell more than twenty-two percent in a single session. I observed the immediate aftermath in the compensation planning conversations we were already midway through with seven active clients.

The effect was swift and specific: wage freeze discussions began the following week in three of our client organizations. Discretionary bonus structures were suspended. Long-term incentive plan reviews — projects we had expected to continue through the first quarter of 1988 — were put on hold while boards and executives absorbed the implications of what had happened. We saw wage freezes ripple across every client engagement we were managing at the time of the crash.

What I did not anticipate, and what I am now prepared to say with some confidence, is that Black Monday demonstrated the countercyclical value of rigorous compensation research. Organizations that were uncertain about where to freeze, how deeply to freeze, and which roles to protect from freezes needed data more urgently, not less. We fielded more new inquiry calls in the fourth quarter of 1987 than in the prior eighteen months combined. We ended the year with fifteen active engagements and a clearer sense of our own market position than I could have described in January.

Chester H. Whitmore III
President & Chief Executive Officer""",
    },
    1988: {
        'theme': 'The Founding Era',
        'desc': 'Post-crash recovery year. Markets stabilize. ACPWB hires its first dedicated analyst team and formalizes the peer-comparison methodology into a published framework.',
        'bg': '#F5F0E8',
        'text_color': '#2B1A0E',
        'accent': '#8B5E3C',
        'accent2': '#D4A96A',
        'font_body': "'Courier New', Courier, monospace",
        'font_head': "'Times New Roman', Times, serif",
        'layout_class': 'era-founding',
        'ceo': 'Chester H. Whitmore III',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

Recovery is a word that implies a return to a prior state. I have come to believe, after working through the aftermath of last year's market dislocation, that this understanding is imprecise. What organizations have returned to is not what existed before October 1987. The compensation planning climate that has emerged is more deliberate, more data-conscious, and frankly more amenable to the kind of rigorous analysis our firm provides.

We have this year formalized our peer-comparison methodology into a published analytical framework — a document I intend to be referenced and critiqued by anyone who considers this work seriously. Good methodology should invite scrutiny. We hired eight new analysts in 1988, bringing our professional staff to twenty-three, and we expanded our data collection partnerships with three regional business associations whose member surveys provide us access to compensation data we could not otherwise obtain.

I remain cautious about projecting the forward trajectory of this firm with undue confidence. The market has proven capable of surprises. But I can state, with the evidence of two years of post-Black Monday demand, that our thesis about the durable value of compensation benchmarking has been tested and has held.

Chester H. Whitmore III
President & Chief Executive Officer""",
    },
    1989: {
        'theme': 'The Founding Era',
        'desc': 'The Berlin Wall falls. The Cold War ends. Defense sector clients begin asking ACPWB about compensation for a post-military-industrial economy.',
        'bg': '#F5F0E8',
        'text_color': '#2B1A0E',
        'accent': '#8B5E3C',
        'accent2': '#D4A96A',
        'font_body': "'Courier New', Courier, monospace",
        'font_head': "'Times New Roman', Times, serif",
        'layout_class': 'era-founding',
        'ceo': 'Chester H. Whitmore III',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

The year 1989 will be recorded in history primarily for events that occurred in Europe — events that signal, I believe, a fundamental restructuring of the political economy within which all of our work takes place. The fall of the Berlin Wall in November, and the cascade of political transformations that preceded it, suggests that the competitive and institutional landscape of the next decade will look substantially different from the one that framed the prior forty years.

For compensation research, the most immediate practical implication concerns the defense and aerospace sectors. Three of our longer-standing clients in these categories have, over the latter months of this year, begun exploratory conversations with us about what compensation structures might look like if the anticipated reductions in defense procurement materialize. These are early-stage discussions, but they reflect a genuine and sophisticated concern about workforce planning in the context of sector-wide transition.

We have also, this year, expanded our coverage of the financial services sector, which continues to generate strong benchmarking demand despite — or perhaps because of — the ongoing consolidations reshaping its competitive structure. We enter the final year of this decade with twenty-eight active engagements, the broadest sector coverage in our history, and a team I trust to handle whatever the next decade brings.

Chester H. Whitmore III
President & Chief Executive Officer""",
    },
    1990: {
        'theme': 'The Expansion Years',
        'desc': 'Gulf War begins. Recession tightens. But downsizing creates an unexpected ACPWB practice area: severance structure and workforce reduction analytics.',
        'bg': '#F0F4F8',
        'text_color': '#1A1A2E',
        'accent': '#1E5F74',
        'accent2': '#4DA6C8',
        'font_body': 'Helvetica, Arial, sans-serif',
        'font_head': "'Arial Black', Gadget, sans-serif",
        'layout_class': 'era-expansion',
        'ceo': 'Chester H. Whitmore III',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

The first half of 1990 carried the residual optimism of the late-decade recovery. By August, when Iraqi forces crossed into Kuwait and American troops began mobilizing in the Gulf, the economic atmosphere shifted perceptibly. The recession that followed — the first since the early 1980s — arrived not as a sudden collapse but as a slow tightening. Organizations that had been adding headcount through 1989 began to ask, quietly at first and then more openly, whether they were appropriately staffed and whether their compensation structures could sustain a period of constraint.

We have found, somewhat unexpectedly, that downsizing creates a specific and urgent need for compensation research. When organizations reduce headcount, they face immediate questions about severance equity, role consolidation, and whether the compensation of retained employees is structured to retain the right people. These are precisely the questions our methodology is designed to address. We have formalized this into what we are internally calling our Workforce Transition Practice — a set of analytical tools specific to the restructuring context.

I would prefer to report a year of pure expansion. What I can report instead is a year of deepened relevance. We added nine new clients, largely on the strength of our Workforce Transition work, and we end the year with confidence in the durability of our analytical approach across the full economic cycle.

Chester H. Whitmore III
President & Chief Executive Officer""",
    },
    1991: {
        'theme': 'The Expansion Years',
        'desc': 'Gulf War ends quickly, but the recession deepens. Mass downsizing across American industry drives record ACPWB engagement volume — our workforce restructuring practice triples.',
        'bg': '#F0F4F8',
        'text_color': '#1A1A2E',
        'accent': '#1E5F74',
        'accent2': '#4DA6C8',
        'font_body': 'Helvetica, Arial, sans-serif',
        'font_head': "'Arial Black', Gadget, sans-serif",
        'layout_class': 'era-expansion',
        'ceo': 'Chester H. Whitmore III',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

The Gulf War concluded with military efficiency in February, but the economic conflict on the domestic front proved more protracted. Unemployment climbed through the spring and into the summer. The wave of corporate downsizing that began quietly in 1990 became, in 1991, a defining feature of the American business landscape. Major manufacturers, financial institutions, and even healthcare organizations announced workforce reductions in numbers that would have seemed implausible five years prior.

For ACPWB, this was our most consequential year of operation to date. Our Workforce Transition Practice — the group we established informally in 1990 to address the compensation dimensions of restructuring — tripled in client volume. We found ourselves advising organizations not only on severance structure and equity, but on the more complex question of how to maintain pay integrity for a workforce that has watched significant portions of itself depart. The morale and retention implications of poorly structured downsizing compensation are, as our data increasingly demonstrates, severe and long-lasting.

I want to be clear about the nature of this growth: we did not prosper because suffering was widespread. We prospered because we had built, over six years, an analytical capability that organizations needed precisely in moments of difficulty. That is the version of this firm I set out to build, and I believe we are now, unmistakably, that firm.

Chester H. Whitmore III
President & Chief Executive Officer""",
    },
    1992: {
        'theme': 'The Expansion Years',
        'desc': 'Clinton elected. The economy begins a long recovery. Chester Whitmore announces his retirement after eight years as founding CEO. Reginald T. Ashworth named successor.',
        'bg': '#F0F4F8',
        'text_color': '#1A1A2E',
        'accent': '#1E5F74',
        'accent2': '#4DA6C8',
        'font_body': 'Helvetica, Arial, sans-serif',
        'font_head': "'Arial Black', Gadget, sans-serif",
        'layout_class': 'era-expansion',
        'ceo': 'Chester H. Whitmore III',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

After eight years as President and Chief Executive Officer of the American Corporation for Public Well Being, I have informed our Board of Directors of my intention to transition out of the chief executive role at the close of this fiscal year. Reginald T. Ashworth, who has served as our Chief Research Officer and who was instrumental in the development of our Normalized Sector Compensation Index, will assume the presidency effective January 1, 1993. I will remain available to the organization in an advisory capacity.

I write this final letter having watched a presidential transition of my own — a different kind, to be sure. The election of Bill Clinton represents a generational shift in American political leadership, one that will almost certainly influence the regulatory and policy environment within which compensation research operates. I leave my successor a firm that is well positioned regardless of the direction those policy currents run.

What I am proudest of is not any particular client engagement or methodological innovation, though there have been several of each worth remembering. What I am proudest of is that we built this organization with integrity. The data we published was the data we had. The conclusions we drew were warranted by the evidence. In an industry where the incentive to tell clients what they want to hear is substantial, we consistently chose to tell them what we found.

To Reginald, to our clients, to the forty-seven staff who make this work possible: thank you.

Chester H. Whitmore III
President & Chief Executive Officer""",
    },
    1993: {
        'theme': 'The Expansion Years',
        'desc': 'Reginald Ashworth\'s first year. NAFTA signed. The American economy accelerates into a long expansion. ACPWB opens a Chicago office and hires its first economist from the academic world.',
        'bg': '#F0F4F8',
        'text_color': '#1A1A2E',
        'accent': '#1E5F74',
        'accent2': '#4DA6C8',
        'font_body': 'Helvetica, Arial, sans-serif',
        'font_head': "'Arial Black', Gadget, sans-serif",
        'layout_class': 'era-expansion',
        'ceo': 'Reginald T. Ashworth',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

I accepted the responsibility of leading this organization with full awareness of the standard Chester Whitmore set over its first eight years. I did not accept it with any intention of maintaining the pace at which we have grown. I accepted it with the intention of substantially accelerating that pace.

The signing of NAFTA in December 1993 — the culmination of years of trade negotiations — signals a fundamental restructuring of the North American labor market. Organizations with operations in multiple countries will face compensation benchmarking challenges of an entirely new character. We are, as of this year, actively building the cross-border analytical capability to address them. Our new Chicago office, opened in October, serves as the operational hub for this expansion.

We have also hired, for the first time, a Ph.D. economist from the academic world — Dr. Marguerite Foss of the University of Chicago — to lead our Research Methods division. Academic rigor and practical consulting have not always found each other easily in this industry. I believe that gap represents an opportunity, and I intend to exploit it.

Chester's legacy is a firm with unimpeachable integrity and a modest market position. My goal is to add the market position.

Reginald T. Ashworth
President & Chief Executive Officer""",
    },
    1994: {
        'theme': 'The Expansion Years',
        'desc': 'The \'information superhighway\' arrives. ACPWB pilots its first electronic data collection system. Republicans sweep Congress in November. Health care reform collapses.',
        'bg': '#F0F4F8',
        'text_color': '#1A1A2E',
        'accent': '#1E5F74',
        'accent2': '#4DA6C8',
        'font_body': 'Helvetica, Arial, sans-serif',
        'font_head': "'Arial Black', Gadget, sans-serif",
        'layout_class': 'era-expansion',
        'ceo': 'Reginald T. Ashworth',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

The Vice President of the United States spent considerable time this year describing something called the "information superhighway," and while this language struck many observers as premature, it did not strike me that way. The combination of desktop computing and networked communication represents, I believe, a genuine inflection point in how compensation data is collected, analyzed, and delivered. We piloted our first electronic data collection system this year with twelve participating organizations. The results were unambiguous: faster collection, higher response rates, and substantially cleaner data than any paper-based process we have previously employed.

The collapse of health care reform — the Clinton administration's most ambitious domestic initiative — has created both disruption and opportunity for our benefits benchmarking practice. Organizations that had anticipated regulatory transformation in how they structure health and retirement benefits are now operating in a more stable but equally uncertain landscape. Uncertainty, we have consistently found, generates demand for data.

Our client roster now stands at sixty-three active engagements, a fifty-two percent increase from my first year. We are growing faster than our internal systems have comfortably accommodated, and I have asked our operations team to prioritize infrastructure over the first half of 1995.

Reginald T. Ashworth
President & Chief Executive Officer""",
    },
    1995: {
        'theme': 'Digital Dawn',
        'desc': 'Netscape goes public in August and the internet economy is born. ACPWB launches its first digital report delivery platform. First seven-figure engagement signed.',
        'bg': '#FFFFFF',
        'text_color': '#1A0A3E',
        'accent': '#6B21A8',
        'accent2': '#A855F7',
        'font_body': 'Arial, Helvetica, sans-serif',
        'font_head': "'Arial Black', Arial, sans-serif",
        'layout_class': 'era-digital-dawn',
        'ceo': 'Reginald T. Ashworth',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

August 9, 1995. Netscape Communications Corporation went public. The stock opened at $28 and closed at $58. The company had never turned a profit. I watched the coverage from our Chicago office and, unlike some of my peers, did not dismiss what I was seeing. The internet is not a communications technology with incremental implications for business. It is a restructuring of the information economy itself, and compensation research — which is, at its core, an information business — will be transformed by it.

We launched our digital report delivery platform this year — modest by any objective measure, but real. Clients can now access their benchmarking results through a password-protected web interface rather than waiting for courier delivery of printed binders. Response from our client base has been, to use an understatement, enthusiastic.

We also signed, for the first time in our history, a single-engagement contract exceeding one million dollars. The client is a major Midwestern financial institution whose executive compensation structure had grown so complex across acquisitions and restructurings that only a comprehensive multi-year benchmarking project could adequately address it. We assigned twelve analysts to the work. I will not pretend the number did not give me considerable satisfaction.

Reginald T. Ashworth
President & Chief Executive Officer""",
    },
    1996: {
        'theme': 'Digital Dawn',
        'desc': 'Stock options as compensation surge to historic levels. The dot-com boom redefines how talent is compensated. ACPWB launches its first equity compensation benchmarking practice.',
        'bg': '#FFFFFF',
        'text_color': '#1A0A3E',
        'accent': '#6B21A8',
        'accent2': '#A855F7',
        'font_body': 'Arial, Helvetica, sans-serif',
        'font_head': "'Arial Black', Arial, sans-serif",
        'layout_class': 'era-digital-dawn',
        'ceo': 'Reginald T. Ashworth',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

The compensation landscape of 1996 is one that would have been unrecognizable to our founding team a decade ago. Stock options — once the exclusive province of senior executives and venture-backed technology founders — are now being offered to software engineers, marketing managers, and in some organizations, every full-time employee. The incentive structures of American enterprise are being reconceived in real time.

Our response has been to build a dedicated equity compensation benchmarking practice. We hired four specialists from the technology sector this year, individuals who understand option structures, vesting schedules, and the increasingly complex accounting treatment of equity compensation under current standards. We published our first Equity Compensation Benchmark Report in September, drawing on data from 187 organizations across technology, financial services, and emerging internet sectors.

The demand for this work has exceeded our projections substantially. I am revising our five-year growth targets upward for the third consecutive year. The technology sector is rewriting the rules of compensation, and every organization — technology-native or otherwise — that competes for talent in this market needs to understand those rules better than they currently do.

Reginald T. Ashworth
President & Chief Executive Officer""",
    },
    1997: {
        'theme': 'Digital Dawn',
        'desc': 'Asian financial crisis ripples through global markets. US economy continues domestic boom. ACPWB expands internationally for the first time, opening a Toronto satellite office.',
        'bg': '#FFFFFF',
        'text_color': '#1A0A3E',
        'accent': '#6B21A8',
        'accent2': '#A855F7',
        'font_body': 'Arial, Helvetica, sans-serif',
        'font_head': "'Arial Black', Arial, sans-serif",
        'layout_class': 'era-digital-dawn',
        'ceo': 'Reginald T. Ashworth',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

The currency collapses that swept Thailand, Indonesia, South Korea, and Malaysia this year served as a reminder — particularly valuable at this stage of the American economic expansion — that global prosperity is neither linear nor guaranteed. Firms with significant exposure to Asian markets saw their compensation planning assumptions disrupted in ways our methodology was only partially equipped to address. We have since begun developing market-volatility adjustment factors for our cross-border benchmarking tools.

Domestically, the economy continues to reward those who are participating in it. Unemployment has reached levels not seen since the 1960s, and the competition for skilled workers across technology, healthcare, and professional services is producing upward compensation pressure that our clients are both experiencing and, in some cases, driving. We are increasingly asked not to describe the market but to help clients get ahead of it.

We opened a satellite office in Toronto this year — our first outside the United States — to serve the growing number of Canadian organizations and US firms with Canadian operations who require cross-border compensation data. We have also begun preliminary discussions with a London-based research firm about a potential data-sharing partnership. The world is becoming a more relevant frame of reference for compensation benchmarking than it was when we started.

Reginald T. Ashworth
President & Chief Executive Officer""",
    },
    1998: {
        'theme': 'Digital Dawn',
        'desc': 'Clinton impeachment dominates headlines. Tech boom accelerates. Stock option compensation reaches peak complexity. ACPWB issues its first warning about equity compensation overreliance.',
        'bg': '#FFFFFF',
        'text_color': '#1A0A3E',
        'accent': '#6B21A8',
        'accent2': '#A855F7',
        'font_body': 'Arial, Helvetica, sans-serif',
        'font_head': "'Arial Black', Arial, sans-serif",
        'layout_class': 'era-digital-dawn',
        'ceo': 'Reginald T. Ashworth',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

The domestic news cycle of 1998 was dominated by matters that have little direct bearing on compensation research — an observation that does not prevent me from noting, with some professional concern, how the political turbulence of the year affected board-level decision-making at several of our client organizations. Compensation committee deliberations slowed in the third and fourth quarters as executive attention was occupied elsewhere. Decisions that should have been made in September were made in January.

Our more substantive concern this year is the compensation architecture we are observing across the technology sector, and increasingly in organizations that compete with technology firms for talent. The overweighting of equity compensation — in some cases, option packages representing three to five times annual salary — has produced total compensation structures that are extraordinarily sensitive to market valuation. We have this year issued, for the first time in our history, a formal advisory to several clients recommending rebalancing toward fixed cash compensation to reduce this exposure.

I want to be direct: if the equity markets experience a significant correction, a meaningful portion of the total compensation packages currently being offered will effectively disappear. The firms that have built sound cash compensation foundations will retain employees. The firms that have not will face a crisis they did not anticipate. We believe the data supports this concern.

Reginald T. Ashworth
President & Chief Executive Officer""",
    },
    1999: {
        'theme': 'Digital Dawn',
        'desc': 'Y2K preparation consumes IT budgets globally. The dot-com frenzy hits its apex. ACPWB signs its largest engagement ever — a $4.2M Y2K workforce contingency planning contract.',
        'bg': '#FFFFFF',
        'text_color': '#1A0A3E',
        'accent': '#6B21A8',
        'accent2': '#A855F7',
        'font_body': 'Arial, Helvetica, sans-serif',
        'font_head': "'Arial Black', Arial, sans-serif",
        'layout_class': 'era-digital-dawn',
        'ceo': 'Reginald T. Ashworth',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

In some respects, this letter should be the easiest I have written. Revenue grew forty-one percent. We completed our largest single engagement in company history — a $4.2 million workforce contingency planning project that addressed, among other things, how to structure compensation for the emergency response teams that would be needed if Y2K disruptions materialized at scale. The fact that January 1, 2000 arrived without incident is, from a business standpoint, a mixed blessing: the preparation work was genuine and well-compensated, even if the crisis it anticipated did not materialize.

The harder thing to write, because it involves uncertainty rather than success, is my view of where we are heading. The Nasdaq Composite has more than doubled in twelve months. Organizations whose entire compensation architectures are denominated in options that assume perpetual appreciation are operating on assumptions that our data does not support. I said something similar last year. I am saying it again this year because it remains true and because I am aware that no one particularly wants to hear it.

We are well positioned for whatever comes next, because we have never allowed our own financial success to outpace our analytical capabilities. The same rigor that made us valuable in the recessions of 1987 and 1991 will make us valuable in whatever follows the current expansion. I look forward to demonstrating that.

Reginald T. Ashworth
President & Chief Executive Officer""",
    },
    2000: {
        'theme': 'The Reckoning',
        'desc': 'The Nasdaq peaks in March and collapses. Dot-com companies begin failing weekly. ACPWB\'s equity compensation practice pivots overnight from benchmarking to triage.',
        'bg': '#1A0505',
        'text_color': '#E8D5D5',
        'accent': '#9B2226',
        'accent2': '#C85252',
        'font_body': "Georgia, 'Times New Roman', serif",
        'font_head': "'Courier New', Courier, monospace",
        'layout_class': 'era-reckoning',
        'ceo': 'Reginald T. Ashworth',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

I will not pretend that watching the Nasdaq fall from 5,132 to 2,470 over the course of 2000 was a comfortable professional experience. We built an equity compensation benchmarking practice on the assumption that organizations needed rigorous data to structure equity packages fairly. What we found, as the correction accelerated through the spring, was that many of our technology-sector clients needed something more immediate: help understanding what to do when the options that constituted most of their employees' expected compensation became, in some cases, worthless.

Our equity compensation practice pivoted from benchmarking to what I can only describe as triage. We worked with twelve technology firms on repricing strategies, retention bonus design for underwater option holders, and the communication of compensation restructuring to employees whose financial expectations had been materially damaged. This work was among the most complex and consequential we have undertaken.

I want to acknowledge that I issued warnings about equity overconcentration in 1998 and 1999 that were not, in my professional judgment, acted upon with sufficient urgency. I take no satisfaction in having been right. The harm done to employees who structured their financial plans around option values that evaporated was real. We have updated our advisory standards to make the risks of equity concentration more prominent in all future engagements.

Reginald T. Ashworth
President & Chief Executive Officer""",
    },
    2001: {
        'theme': 'The Reckoning',
        'desc': 'September 11th. The American economy — already contracting — freezes. Compensation reviews are suspended across entire industries. ACPWB\'s most somber year.',
        'bg': '#1A0505',
        'text_color': '#E8D5D5',
        'accent': '#9B2226',
        'accent2': '#C85252',
        'font_body': "Georgia, 'Times New Roman', serif",
        'font_head': "'Courier New', Courier, monospace",
        'layout_class': 'era-reckoning',
        'ceo': 'Reginald T. Ashworth',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

There is no professional framework adequate to what happened on September 11, 2001. I will not attempt to construct one. What I can report is what followed, as observed from inside the organizations our firm serves.

Compensation review processes that were underway in August were suspended across the entirety of the financial services sector within two weeks of the attacks. Travel and entertainment companies froze all workforce planning indefinitely. Aviation industry clients — we had five at the time — entered immediate emergency restructuring engagements that had nothing to do with benchmarking and everything to do with survival. Organizations that had been planning to grow their workforces were planning instead how to reduce them as quickly and humanely as possible.

Our revenue fell in the fourth quarter by thirty-eight percent. I expected worse. What sustained us was the counterintuitive truth that even in crisis, organizations need compensation data — to design severance that can withstand legal scrutiny, to structure retention packages for employees deemed essential, to understand what the market looks like after the market has collapsed. We were useful in ways I would have preferred not to be.

I am announcing in this letter my intention to transition the chief executive role to Diane Holloway, our Chief Operating Officer, effective March 1, 2002. She has my full confidence and the full respect of every person who has worked alongside her. I leave knowing the firm is in the right hands.

Reginald T. Ashworth
President & Chief Executive Officer""",
    },
    2002: {
        'theme': 'The Reckoning',
        'desc': 'Diane Holloway becomes CEO — the first woman to lead ACPWB. Enron and WorldCom collapse. Sarbanes-Oxley is debated. Executive compensation is suddenly front-page news.',
        'bg': '#1A0505',
        'text_color': '#E8D5D5',
        'accent': '#9B2226',
        'accent2': '#C85252',
        'font_body': "Georgia, 'Times New Roman', serif",
        'font_head': "'Courier New', Courier, monospace",
        'layout_class': 'era-reckoning',
        'ceo': 'Diane P. Holloway',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

I want to begin by acknowledging what I know many of our clients are thinking: the events of this year have made compensation transparency not a professional preference but a public necessity. Enron and WorldCom did not collapse because compensation was poorly benchmarked. They collapsed because of fraud. But the compensation structures that allowed executives to extract enormous personal wealth while the organizations they led disintegrated — those structures are now, rightly, under the most intense scrutiny this profession has ever experienced.

I accepted the chief executive role from Reginald Ashworth in March committed to leading this organization through whatever regulatory transformation the post-Enron environment produces. The Sarbanes-Oxley Act, which was signed into law in July, represents the most significant change to corporate governance since the securities laws of the 1930s. Our compensation committee advisory practice will be substantially more consequential — and more demanding — in its wake.

On a personal note: I am aware that my appointment as the first female chief executive in ACPWB's history has been noted in this industry. I appreciate the acknowledgment and find it irrelevant to the work. The work is the work. We do it rigorously and honestly. Everything else follows from that.

Diane P. Holloway
President & Chief Executive Officer""",
    },
    2003: {
        'theme': 'The Reckoning',
        'desc': 'Sarbanes-Oxley takes effect. Executive compensation disclosure requirements explode. ACPWB\'s compensation committee advisory practice triples in twelve months.',
        'bg': '#1A0505',
        'text_color': '#E8D5D5',
        'accent': '#9B2226',
        'accent2': '#C85252',
        'font_body': "Georgia, 'Times New Roman', serif",
        'font_head': "'Courier New', Courier, monospace",
        'layout_class': 'era-reckoning',
        'ceo': 'Diane P. Holloway',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

Sarbanes-Oxley is now in effect, and the implications for our practice are larger than even our most optimistic internal projections anticipated. The enhanced disclosure requirements for executive compensation — applicable to all public companies — have created an immediate, urgent, and widespread need for the kind of independent benchmarking analysis our firm provides. Compensation committees that previously met twice a year and relied primarily on management-provided data are now meeting six to eight times annually and requiring independent analytical support at each session.

Our compensation committee advisory practice, which represented approximately eighteen percent of revenue in 2001, now represents forty-four percent. We have added twenty-two professionals to that practice alone in the past twelve months. The growth has strained our systems and our culture, and I want to be direct about that: rapid growth is not intrinsically good. It requires management discipline that I am asking of myself and every leader in this organization.

The Iraq War, which began in March, has had limited direct impact on our practice — unlike previous military engagements, this one has not generated the kind of broad economic disruption that affects compensation planning at scale. The recovery from the 2001 recession continues, unevenly. We are growing regardless, because the regulatory environment has become our most powerful market driver.

Diane P. Holloway
President & Chief Executive Officer""",
    },
    2004: {
        'theme': 'Global Reach',
        'desc': 'Recovery accelerates. Facebook is founded. A generation of technology workers will eventually need compensation benchmarking. ACPWB\'s first $10M revenue year.',
        'bg': '#F7F5F0',
        'text_color': '#1A2A1A',
        'accent': '#2A5F4F',
        'accent2': '#5FA087',
        'font_body': "Garamond, Georgia, serif",
        'font_head': "'Palatino Linotype', Palatino, serif",
        'layout_class': 'era-global',
        'ceo': 'Diane P. Holloway',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

We crossed the ten-million-dollar annual revenue threshold in 2004. This is a milestone I had identified as a meaningful marker of institutional scale, and I observe it with more reflection than celebration. Ten million dollars is, in the context of the management consulting industry, not a large firm. What it represents for us is the confirmation that a compensation research organization built on rigorous methodology, independent analysis, and long-term client relationships can achieve genuine scale without compromising any of those qualities.

The economic recovery is now unmistakable. Hiring has resumed across most major sectors, and compensation benchmarking demand has shifted from the defensive posture of 2001–2003 back toward the forward-looking workforce planning that drives our most interesting analytical work. We have expanded our healthcare and technology sector practices, both of which are experiencing the kinds of compensation pressures that generate sustained demand for our services.

I will note, for those who track these developments, that a Harvard student named Mark Zuckerberg has launched a social networking platform that appears to be spreading rapidly among university students. I do not yet have a view on the compensation implications of whatever this represents. I expect I will.

Diane P. Holloway
President & Chief Executive Officer""",
    },
    2005: {
        'theme': 'Global Reach',
        'desc': 'Housing market booming. Offshore outsourcing intensifies global compensation complexity. ACPWB launches its first international compensation benchmarking product.',
        'bg': '#F7F5F0',
        'text_color': '#1A2A1A',
        'accent': '#2A5F4F',
        'accent2': '#5FA087',
        'font_body': "Garamond, Georgia, serif",
        'font_head': "'Palatino Linotype', Palatino, serif",
        'layout_class': 'era-global',
        'ceo': 'Diane P. Holloway',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

The domestic economic picture in 2005 is one of broad prosperity with specific fragilities. The housing market continues to appreciate at rates that, by any historical comparison, should prompt caution. Consumer debt levels have reached new records. And yet employment is robust, corporate earnings are strong, and the organizations that constitute our client base are, by and large, planning to grow.

The question our clients are asking with increasing frequency is not what domestic peers are paying, but what global peers are paying. The offshoring of professional and technical work — which has expanded substantially in the past five years, particularly to India and Eastern Europe — has created compensation benchmarking challenges that our historical methodology was not designed to address. We have this year launched our International Compensation Benchmarking product, drawing on data partnerships with research firms in twelve countries.

I remain concerned about the housing market. I will say it plainly because I believe in plain speaking: the velocity of residential real estate appreciation in the United States is not consistent with the underlying economics of housing supply and demand. This does not directly affect compensation research, but it does affect the financial security of the employees our research is intended to benefit. I watch it with unease.

Diane P. Holloway
President & Chief Executive Officer""",
    },
    2006: {
        'theme': 'Global Reach',
        'desc': 'The year ACPWB was formally incorporated. Credit bubble inflating. YouTube explodes. Executive compensation is now a genuine populist flashpoint.',
        'bg': '#F7F5F0',
        'text_color': '#1A2A1A',
        'accent': '#2A5F4F',
        'accent2': '#5FA087',
        'font_body': "Garamond, Georgia, serif",
        'font_head': "'Palatino Linotype', Palatino, serif",
        'layout_class': 'era-global',
        'ceo': 'Diane P. Holloway',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

This year marks ACPWB's formal incorporation under the laws of the State of Wisconsin. We have operated continuously since 1985, but various structural and legal developments in 2006 prompted us to formalize what had previously been a more informal partnership arrangement. The organization is, in legal terms, new; in every other respect, it is twenty-one years old and carries those years in its institutional memory.

The macro environment continues to baffle prediction. The credit markets are extending leverage in ways that our analytical team has begun describing, privately, as unsustainable. Housing prices in coastal markets have reached a relationship to rents and incomes that no historical model can comfortably explain. We watch these indicators not as financiers — compensation research is our work — but because the organizations we advise are embedded in these dynamics, and their compensation planning will be directly affected by whatever resolution these imbalances eventually produce.

On the more immediate front: executive compensation has become a populist issue in ways that were not true even five years ago. The ratio of CEO pay to median worker pay has attracted significant media attention. We have been asked, more often this year than in any prior year, to assist organizations in thinking about that ratio as a governance question rather than merely as a benchmarking exercise.

Diane P. Holloway
President & Chief Executive Officer""",
    },
    2007: {
        'theme': 'Global Reach',
        'desc': 'The iPhone launches. The credit bubble reaches its apex. ACPWB records its highest revenue to date and opens three new offices. We are briefly, embarrassingly bullish.',
        'bg': '#F7F5F0',
        'text_color': '#1A2A1A',
        'accent': '#2A5F4F',
        'accent2': '#5FA087',
        'font_body': "Garamond, Georgia, serif",
        'font_head': "'Palatino Linotype', Palatino, serif",
        'layout_class': 'era-global',
        'ceo': 'Diane P. Holloway',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

Revenue grew twenty-eight percent. We opened offices in Atlanta, Denver, and San Francisco. We added forty-one professional staff and signed nineteen new clients in a single quarter — a record by a factor of two. In January, Steve Jobs held up a small glass rectangle and said it was a telephone, a music player, and an internet communicator. He was right on all three counts, and the compensation implications of the mobile software industry that will follow are already visible in our technology sector data.

I want to be honest about something. In the second and third quarters of this year, I allowed myself to believe that what I was observing in the credit and real estate markets was a kind of permanent structural shift rather than the late stage of a debt cycle. I did not say this publicly — the precision required for such a claim was beyond what our data supported — but I felt it, and it influenced some of the expansion decisions I made. Three offices in one year was a number I arrived at through optimism rather than analysis.

By the fourth quarter, the first signs of the credit market stress that would define 2008 were becoming visible. I am writing this in February 2008, and I know now what the full year of 2007 did not quite reveal. The letter I write next year will be different from this one in almost every register.

Diane P. Holloway
President & Chief Executive Officer""",
    },
    2008: {
        'theme': 'The Crisis',
        'desc': 'Lehman Brothers. Bear Stearns. AIG. The global financial system nearly ends. ACPWB loses three major financial-sector clients in one week. We survive by being necessary.',
        'bg': '#F2F2F0',
        'text_color': '#1A1A1A',
        'accent': '#3D3D3D',
        'accent2': '#888888',
        'font_body': "Georgia, 'Times New Roman', serif",
        'font_head': "Georgia, 'Times New Roman', serif",
        'layout_class': 'era-crisis',
        'ceo': 'Diane P. Holloway',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

Bear Stearns failed in March. Lehman Brothers filed for bankruptcy on September 15. AIG required a federal rescue of $85 billion to avoid collapse. The Washington Mutual failure was the largest bank failure in American history. These are not abstractions. Three of our financial services clients suspended all discretionary consulting engagements within a week of the Lehman filing. One client, a firm we had worked with for eleven years, ceased to exist.

I will not write at length about what this year was like for the people at ACPWB. It was difficult. We reduced headcount by twelve percent in October — the first significant workforce reduction in our history. The irony that a compensation research firm was designing its own layoff package was not lost on any of us.

What preserved this organization was the same thing that preserved us in 1987 and 1991: when organizations are in crisis, they need compensation data more urgently, not less. The organizations that survived the financial crisis — and many did — faced immediate questions about how to restructure their executive compensation in response to regulatory and public pressure, how to retain critical employees while cutting compensation broadly, and how to design severance for employees who were being terminated for reasons entirely outside their performance or control. We were needed.

Revenue fell nineteen percent. We closed the Denver office. We did not collapse. I count that as the achievement of this year, and I mean it without irony.

Diane P. Holloway
President & Chief Executive Officer""",
    },
    2009: {
        'theme': 'The Crisis',
        'desc': 'TARP, bailouts, and the compensation clawback era begin. New government regulations on executive pay create a windfall for ACPWB\'s governance practice. Recovery is tentative but real.',
        'bg': '#F2F2F0',
        'text_color': '#1A1A1A',
        'accent': '#3D3D3D',
        'accent2': '#888888',
        'font_body': "Georgia, 'Times New Roman', serif",
        'font_head': "Georgia, 'Times New Roman', serif",
        'layout_class': 'era-crisis',
        'ceo': 'Diane P. Holloway',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

The compensation restrictions attached to TARP — the Troubled Asset Relief Program — created an immediate and significant demand for specialized advisory work that ACPWB was, almost uniquely, positioned to provide. Organizations that had accepted government assistance were now subject to executive compensation structures that required independent documentation, oversight, and in many cases, complete redesign. The Say on Pay provisions that followed created additional advisory demand across the broader public company universe.

I want to be precise about what this means: we grew revenue thirty-one percent in a year when the broader economy contracted sharply. We did not grow because we are opportunists. We grew because the regulatory response to the financial crisis created genuine analytical challenges that organizations could not resolve internally, and we had spent twenty-four years building the capability to address them.

We reopened the Denver office in September, rehiring four of the eight analysts we had let go in October 2008. Those conversations were among the most difficult I have had as a manager. I am grateful that we could have them.

Obama's administration has signaled that executive compensation reform will remain a legislative priority. Our governance and compliance advisory practice will be the most important growth driver for the next several years. We are investing accordingly.

Diane P. Holloway
President & Chief Executive Officer""",
    },
    2010: {
        'theme': 'Rebuilding',
        'desc': 'Dodd-Frank. Recovery consolidates. The iPad arrives and the tablet era begins. ACPWB rebuilds its balance sheet and hires aggressively into the governance practice.',
        'bg': '#F5FAF7',
        'text_color': '#1A2A1E',
        'accent': '#2D6A4F',
        'accent2': '#52B788',
        'font_body': 'Verdana, Geneva, sans-serif',
        'font_head': "'Trebuchet MS', Helvetica, sans-serif",
        'layout_class': 'era-rebuilding',
        'ceo': 'Diane P. Holloway',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

The Dodd-Frank Wall Street Reform and Consumer Protection Act — signed in July — extends the regulatory reach into compensation governance further than any prior legislation. The Say on Pay requirement, which mandates that public companies submit executive compensation for shareholder advisory vote, is now federal law. The implications for our practice are structural rather than cyclical: this is not temporary demand generated by a crisis. This is permanent demand generated by a governance requirement.

We hired twenty-seven professionals in 2010, the largest single-year hiring cohort in our history. More than half of them joined the governance and compliance practice. We have rebuilt the balance sheet reserves we depleted during the crisis response, and we have invested in the analytical infrastructure — data systems, benchmarking databases, regulatory tracking tools — that we will need to serve a much larger client base.

Recovery is not restoration. I said something like this in 2009, and I believe it more strongly now. The organizations that emerged from the financial crisis are not the same organizations that entered it. Their compensation structures are different. Their regulatory obligations are different. Their governance expectations are different. We are not returning to the work of 2007. We are doing new work, and I am grateful for the clarity that requires.

Diane P. Holloway
President & Chief Executive Officer""",
    },
    2011: {
        'theme': 'Rebuilding',
        'desc': 'Occupy Wall Street. The income inequality debate arrives in public spaces. Arab Spring reshapes global politics. Diane Holloway announces her retirement after a decade of leadership.',
        'bg': '#F5FAF7',
        'text_color': '#1A2A1E',
        'accent': '#2D6A4F',
        'accent2': '#52B788',
        'font_body': 'Verdana, Geneva, sans-serif',
        'font_head': "'Trebuchet MS', Helvetica, sans-serif",
        'layout_class': 'era-rebuilding',
        'ceo': 'Diane P. Holloway',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

The encampments that appeared in Zuccotti Park and in dozens of cities across the country this fall represented something that compensation researchers have seen accumulating in data for years: a public reckoning with the distributional consequences of the economic model we have all been operating within. I am not a political commentator and I have no standing to adjudicate the policy debates that Occupy Wall Street has catalyzed. What I can say is that the ratio of executive compensation to median worker pay — a number ACPWB has been tracking and publishing for fifteen years — is now a headline statistic in ways it has never been before.

This is, professionally, a moment of significant responsibility. We possess data that is relevant to public conversations that matter. How we use it, and whether we use it honestly, will define what this firm is for the next decade.

I have informed our Board of Directors of my intention to retire from the chief executive role at the end of this year. Marcus Pemberton, our Chief Strategy Officer, will succeed me. He brings a sensibility about technology and data that I believe is exactly what this organization needs as it enters its next phase. I am proud of what we have built during a decade that encompassed both the worst financial crisis since the Depression and a genuine reinvention of our firm's capabilities. I leave it in good hands.

Diane P. Holloway
President & Chief Executive Officer""",
    },
    2012: {
        'theme': 'Innovation Decade',
        'desc': 'Marcus Pemberton takes over. Facebook IPO. Social media becomes a factor in compensation transparency. The pay equity movement begins finding legislative traction.',
        'bg': '#F0F4FF',
        'text_color': '#0D1B2A',
        'accent': '#4361EE',
        'accent2': '#7B9FF9',
        'font_body': "'Segoe UI', Tahoma, Geneva, Verdana, sans-serif",
        'font_head': "'Segoe UI', Tahoma, Geneva, Verdana, sans-serif",
        'layout_class': 'era-innovation',
        'ceo': 'Marcus J. Pemberton',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

I want to start with a number: 1.6 billion. That is approximately the number of active Facebook users as I write this letter. When those users share information — including information about their jobs, titles, salaries, and employers — they are creating a new kind of compensation transparency that no policy, no NDA, and no HR communication strategy can fully contain.

Pay transparency is coming. Not because legislators are requiring it, though legislative pressure is growing. It is coming because the information systems that employees use every day — LinkedIn, Glassdoor, internal Slack channels, anonymous employer review sites — have made compensation data increasingly public in ways that are difficult to reverse. Organizations that are not paying equitably are going to find out about it through channels they cannot control.

I took over this organization from Diane Holloway with enormous respect for what she built. My goal is not to preserve that — preservation is not a strategy. My goal is to rebuild ACPWB as a data-first organization for a world in which data is becoming the primary operating environment for everyone we serve. We have begun investing in machine-readable data delivery, API access for our benchmarking indices, and a technology stack that actually reflects the year we are operating in.

The future of compensation research is not binders. It is not even PDFs. I intend to demonstrate that.

Marcus J. Pemberton
President & Chief Executive Officer""",
    },
    2013: {
        'theme': 'Innovation Decade',
        'desc': 'NSA surveillance scandal. Big data becomes the dominant conversation in every industry. ACPWB launches its first machine-learning-assisted compensation modeling tool.',
        'bg': '#F0F4FF',
        'text_color': '#0D1B2A',
        'accent': '#4361EE',
        'accent2': '#7B9FF9',
        'font_body': "'Segoe UI', Tahoma, Geneva, Verdana, sans-serif",
        'font_head': "'Segoe UI', Tahoma, Geneva, Verdana, sans-serif",
        'layout_class': 'era-innovation',
        'ceo': 'Marcus J. Pemberton',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

The Snowden disclosures this year revealed something that most people in technology already understood but that the general public had not fully reckoned with: every organization that handles data at scale is, in some sense, a surveillance organization. The question is not whether data is being collected. The question is who controls it, who can access it, and what obligations attach to its use.

This is relevant to compensation research in ways that are not immediately obvious but are, on reflection, significant. The behavioral and performance data that organizations collect about employees — increasingly including productivity metrics, communication patterns, and algorithmic performance scores — is compensation-relevant data. Our clients are beginning to ask us how to incorporate this data into compensation decisions in ways that are transparent, defensible, and fair. These are questions we are only beginning to develop frameworks to answer.

We launched our first machine-learning-assisted compensation modeling tool this year. I will be honest about what this means: it is a regression model with a better interface than what we previously offered. The underlying mathematics have existed for decades. What is new is the ability to process substantially more data, to update models more frequently, and to surface patterns that human analysts would miss. The tool has been adopted by forty-three clients since its September launch.

Marcus J. Pemberton
President & Chief Executive Officer""",
    },
    2014: {
        'theme': 'Innovation Decade',
        'desc': 'The sharing economy disrupts employment categories. Amazon dominates cloud. ACPWB pivots to real-time benchmarking and kills its quarterly print publication after 22 years.',
        'bg': '#F0F4FF',
        'text_color': '#0D1B2A',
        'accent': '#4361EE',
        'accent2': '#7B9FF9',
        'font_body': "'Segoe UI', Tahoma, Geneva, Verdana, sans-serif",
        'font_head': "'Segoe UI', Tahoma, Geneva, Verdana, sans-serif",
        'layout_class': 'era-innovation',
        'ceo': 'Marcus J. Pemberton',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

We discontinued our quarterly print publication this year after twenty-two years of continuous production. The Compensation Research Quarterly shipped its final issue in August. I know this decision was not universally welcomed — we received, genuinely, letters of complaint from clients who had filed every issue since 1992. I understand the attachment. Print is slower and more permanent than digital delivery, and permanence has real value in research.

What print cannot do is reflect the pace at which compensation benchmarks are now moving. Real-time data delivery — updates to our benchmarking indices delivered via API rather than quarterly reports — gives our clients information that is actually current rather than four months old. We are moving to monthly benchmark refreshes as our data collection infrastructure catches up with our delivery ambition.

The rise of Uber, Airbnb, and similar platforms is producing a classification crisis in compensation research. When significant portions of the workforce are classified as independent contractors rather than employees, the compensation benchmarking instruments we have built — almost all of which assume employment relationships — become less applicable. We are developing contractor compensation benchmarking instruments. This is not simple work. The data is sparse, the legal environment is shifting, and the analytical frameworks are genuinely new.

Marcus J. Pemberton
President & Chief Executive Officer""",
    },
    2015: {
        'theme': 'Innovation Decade',
        'desc': 'Paris Climate Agreement. Pay equity legislation spreads from California to Massachusetts. ACPWB\'s pay equity audit practice becomes its second-largest revenue line.',
        'bg': '#F0F4FF',
        'text_color': '#0D1B2A',
        'accent': '#4361EE',
        'accent2': '#7B9FF9',
        'font_body': "'Segoe UI', Tahoma, Geneva, Verdana, sans-serif",
        'font_head': "'Segoe UI', Tahoma, Geneva, Verdana, sans-serif",
        'layout_class': 'era-innovation',
        'ceo': 'Marcus J. Pemberton',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

Pay equity legislation is now spreading at a velocity that our policy tracking team is working hard to keep up with. The Massachusetts Equal Pay Act, signed this summer, prohibits employers from asking about salary history — a provision that has direct implications for how our compensation negotiation guidance needs to be structured. California, New York, and a growing number of municipalities are moving in similar directions.

Our pay equity audit practice — which we launched in 2013 as an experimental service offering — has grown to represent our second-largest revenue line. Organizations that want to understand whether they have gender, race, or age-based pay disparities are hiring us in significant numbers, not always because they are required to, but because investors, employees, and boards are asking questions they were not asking five years ago.

I want to acknowledge the complexity of this work. Identifying a pay disparity is analytically tractable. Understanding its causes — and designing a remediation that is durable rather than cosmetic — requires judgment, organizational knowledge, and communication skill that goes well beyond what any model can provide. We are investing heavily in the advisory capabilities that translate analytical findings into organizational change. That translation is where most pay equity efforts succeed or fail.

Marcus J. Pemberton
President & Chief Executive Officer""",
    },
    2016: {
        'theme': 'Innovation Decade',
        'desc': 'Gig economy at full acceleration. Pay equity legislation wave intensifies. Trump elected in November. ACPWB navigates the political volatility with deliberate neutrality.',
        'bg': '#F0F4FF',
        'text_color': '#0D1B2A',
        'accent': '#4361EE',
        'accent2': '#7B9FF9',
        'font_body': "'Segoe UI', Tahoma, Geneva, Verdana, sans-serif",
        'font_head': "'Segoe UI', Tahoma, Geneva, Verdana, sans-serif",
        'layout_class': 'era-innovation',
        'ceo': 'Marcus J. Pemberton',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

I have been asked by a number of clients and colleagues for ACPWB's view on the election of Donald Trump and its implications for compensation policy. I am going to decline to offer a political analysis, because political neutrality is not a posture for us — it is a precondition of doing our work well. What I can offer is an analytical observation: policy uncertainty is, in compensation research terms, not a reason to pause workforce planning. It is a reason to stress-test the assumptions underlying your workforce planning more rigorously than usual.

The gig economy transformation continues to reshape the compensation landscape in ways that are only partly visible in our data. The proportion of workers engaged in some form of non-traditional employment has grown substantially in the past five years, and the policy and legal frameworks for these arrangements remain genuinely unsettled. We are tracking sixteen active legislative and regulatory proceedings across eight states that will materially affect how independent contractor compensation is structured.

Pay equity legislation has now passed in sixteen states and is under active consideration in eleven more. The salary history ban — which Massachusetts pioneered and which several other states have now adopted — has changed how compensation decisions are made and documented. Our pay equity audit practice continues to be our fastest-growing revenue line.

Marcus J. Pemberton
President & Chief Executive Officer""",
    },
    2017: {
        'theme': 'Innovation Decade',
        'desc': '#MeToo. Pay equity conversations intensify dramatically. Gender pay gap becomes a governance and litigation risk. ACPWB\'s pay equity practice logs its busiest year ever.',
        'bg': '#F0F4FF',
        'text_color': '#0D1B2A',
        'accent': '#4361EE',
        'accent2': '#7B9FF9',
        'font_body': "'Segoe UI', Tahoma, Geneva, Verdana, sans-serif",
        'font_head': "'Segoe UI', Tahoma, Geneva, Verdana, sans-serif",
        'layout_class': 'era-innovation',
        'ceo': 'Marcus J. Pemberton',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

The #MeToo movement that emerged in October — first as a specific accounting of sexual misconduct in the entertainment industry, then as a broader reckoning across virtually every sector — has affected our work in ways I want to address directly. The compensation dimensions of workplace harassment and inequity are substantial and underappreciated. When organizations systematically promote men faster than equally qualified women, when they pay women less for equivalent work, when they concentrate women in lower-paid functions through structural rather than individual decision-making — the compensation data reflects this, even when the individual decisions that produced it appear reasonable in isolation.

Our pay equity audit practice logged its most active year in 2017. The demand was not primarily litigation-driven, though litigation risk was real and growing. The demand was driven by boards, investors, and senior executives who understood, in a newly visceral way, that cultural problems have compensation signatures, and that understanding those signatures requires analytical capacity they do not possess internally.

I am proud of the work we did this year. I am less proud that it took these particular events to drive the demand. Organizations should have been asking these questions before October 2017. Many of them are asking us why they weren't.

Marcus J. Pemberton
President & Chief Executive Officer""",
    },
    2018: {
        'theme': 'Innovation Decade',
        'desc': 'GDPR. CCPA. Data privacy reshapes how compensation data is collected and stored. Pay transparency laws spread. ACPWB rebuilds its data infrastructure for the privacy era.',
        'bg': '#F0F4FF',
        'text_color': '#0D1B2A',
        'accent': '#4361EE',
        'accent2': '#7B9FF9',
        'font_body': "'Segoe UI', Tahoma, Geneva, Verdana, sans-serif",
        'font_head': "'Segoe UI', Tahoma, Geneva, Verdana, sans-serif",
        'layout_class': 'era-innovation',
        'ceo': 'Marcus J. Pemberton',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

The General Data Protection Regulation took effect in May, and the California Consumer Privacy Act was signed into law in June. These are the most significant privacy regulatory developments since the early internet era, and their implications for compensation data specifically are ones we have had to think through very carefully.

Compensation data is sensitive personal data. It is, under GDPR, subject to the full apparatus of data subject rights, processing obligations, and controller-processor agreements. The compensation benchmarking industry — which has historically operated with relatively informal data sharing practices — has had to rethink its data governance architecture from the ground up. We have spent substantial resources this year rebuilding our data infrastructure to comply with the new requirements while preserving the analytical capabilities our clients depend on.

The more interesting question is what happens when compensation data becomes more transparent — by regulation or by market pressure — while simultaneously being more privacy-protected. These two trends are in tension. Pay transparency tells employees and the public what organizations pay. Privacy regulation restricts how that data is collected and processed. Navigating this tension thoughtfully will be a defining challenge of the next decade.

We are well positioned. But I want to be honest: the compliance burden of 2018 was significant, and we are still working through its implications.

Marcus J. Pemberton
President & Chief Executive Officer""",
    },
    2019: {
        'theme': 'Innovation Decade',
        'desc': 'WeWork\'s IPO collapses spectacularly. Unicorn valuations scrutinized. ACPWB issues a major report on equity compensation inflation in private companies. Marcus Pemberton announces succession.',
        'bg': '#F0F4FF',
        'text_color': '#0D1B2A',
        'accent': '#4361EE',
        'accent2': '#7B9FF9',
        'font_body': "'Segoe UI', Tahoma, Geneva, Verdana, sans-serif",
        'font_head': "'Segoe UI', Tahoma, Geneva, Verdana, sans-serif",
        'layout_class': 'era-innovation',
        'ceo': 'Marcus J. Pemberton',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

The WeWork IPO implosion — from a $47 billion valuation to a near-bankruptcy in a matter of weeks — was, for those of us who work in compensation research, a vivid demonstration of something our data has been suggesting for years: the equity compensation structures inside highly-valued private companies often rest on assumptions about exit values that bear very little relationship to underlying economics.

We published our Private Company Equity Compensation Report in September, two weeks before WeWork filed to withdraw its IPO. The report documented the degree to which compensation packages at late-stage private companies — unicorns, in the industry vernacular — had incorporated equity values that were extrapolations of extrapolations. When those valuations corrected, as they have now begun to in several high-profile cases, the compensation structures built on them became inadequate overnight.

I want to close this letter with a personal note. After eight years as chief executive, I am announcing my intention to transition leadership to Courtney Langford, our Chief People Officer, effective January 2020. Courtney brings a depth of empathy and organizational insight that I believe is exactly right for whatever the next decade brings. I have a feeling it will bring quite a lot.

Marcus J. Pemberton
President & Chief Executive Officer""",
    },
    2020: {
        'theme': 'Pandemic Era',
        'desc': 'COVID-19. The global economy shuts down in March. ACPWB helps 200+ organizations restructure compensation for remote work, furloughs, and the new economic reality.',
        'bg': '#FFF8F5',
        'text_color': '#2A1505',
        'accent': '#E76F51',
        'accent2': '#F4A27B',
        'font_body': "'Franklin Gothic Medium', 'Arial Narrow', Arial, sans-serif",
        'font_head': "Impact, Haettenschweiler, 'Arial Narrow Bold', sans-serif",
        'layout_class': 'era-pandemic',
        'ceo': 'Courtney R. Langford',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

I became CEO on January 1, 2020. Eleven weeks later, a global pandemic shut down the American economy.

I will not pretend there was a playbook. There wasn't. What we had was an organization of 214 people who understood compensation deeply, clients who needed us urgently, and a set of analytical capabilities that were, as it turned out, exactly what a crisis-disrupted labor market requires. We moved our entire team to remote work in four days. We stood up a COVID Workforce Response Practice in a week. By April, we had over forty active emergency engagements.

The questions our clients were asking were new in their specificity but familiar in their structure: How do you compensate employees who are furloughed but not terminated? How do you maintain pay equity when your workforce has suddenly stratified between those who can work remotely and those who cannot? How do you design hazard pay for frontline workers in a way that is transparent, consistent, and legally defensible? How do you structure compensation for a workforce that may never fully return to the office?

We helped over two hundred organizations navigate these questions in 2020. Some of the answers we found were obvious in retrospect. Many were genuinely new. I am proud of how this organization responded, and I am grateful to every member of our team who showed up — remotely, stubbornly, often in remarkable personal circumstances — to do the work.

Courtney R. Langford
President & Chief Executive Officer""",
    },
    2021: {
        'theme': 'Pandemic Era',
        'desc': 'The Great Resignation. Compensation benchmarking demand explodes as workers quit en masse and organizations scramble to understand the new labor market. ACPWB\'s best revenue year ever.',
        'bg': '#FFF8F5',
        'text_color': '#2A1505',
        'accent': '#E76F51',
        'accent2': '#F4A27B',
        'font_body': "'Franklin Gothic Medium', 'Arial Narrow', Arial, sans-serif",
        'font_head': "Impact, Haettenschweiler, 'Arial Narrow Bold', sans-serif",
        'layout_class': 'era-pandemic',
        'ceo': 'Courtney R. Langford',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

Four million Americans quit their jobs in April. Another four million in July. The Great Resignation — that is what the economists are calling it, though I prefer to think of it as the Great Recalibration — upended the assumption that pandemic disruption had suppressed worker bargaining power. Workers, it turned out, had used the pandemic to reassess what they were willing to accept in exchange for their labor, and the answer was: more than they had previously been getting.

The demand for compensation benchmarking in 2021 was unlike anything I have seen in my career. Organizations that had not done a comprehensive compensation review in five years found themselves losing employees to competitors who had. Every sector — technology, healthcare, manufacturing, financial services, retail — was simultaneously experiencing turnover and attempting to understand what their compensation structures needed to look like to retain the people they still had.

We had our best revenue year in the organization's history. We hired sixty-eight people. We ran out of capacity in the third quarter and had to turn away work — an experience I found deeply uncomfortable, because the organizations asking for help genuinely needed it.

The vaccines are working. The pandemic is not over, but the acute emergency phase has passed. What remains is a labor market that has been permanently altered, and organizations that need to understand it. We are here for that.

Courtney R. Langford
President & Chief Executive Officer""",
    },
    2022: {
        'theme': 'The AI Shift',
        'desc': 'Russia invades Ukraine. Inflation surges. The Fed raises rates aggressively. Tech layoffs begin in Q4. ACPWB helps clients navigate a labor market that reverses sharply from 2021.',
        'bg': '#0F0A1E',
        'text_color': '#E8E0F0',
        'accent': '#7B2FBE',
        'accent2': '#B480E8',
        'font_body': "'Courier New', Courier, monospace",
        'font_head': "'Courier New', Courier, monospace",
        'layout_class': 'era-ai-shift',
        'ceo': 'Courtney R. Langford',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

The labor market of 2021 and the labor market of late 2022 are almost unrecognizable as the same market. Twelve months ago, organizations were competing aggressively for every technical hire. By November, major technology companies were announcing layoffs in the tens of thousands. Meta cut eleven thousand employees. Twitter cut half its workforce in forty-eight hours. The Fed raised rates seven times, inflation reached forty-year highs, and the economic mood shifted from exuberance to uncertainty with a speed that made compensation planning genuinely difficult.

For organizations that had raised compensation dramatically in 2021 to attract and retain talent, 2022 presented a specific challenge: how do you adjust compensation structures when the market has moved substantially without creating legal, cultural, or equity problems? These are not simple questions, and the answers depend heavily on how the 2021 increases were structured — whether as base salary adjustments, one-time bonuses, or equity grants with future vesting.

Russia's invasion of Ukraine in February added a geopolitical dimension to the economic uncertainty. The energy market disruptions affected compensation planning across European operations for several of our multinational clients.

What I find most striking is the speed of the reversal. Compensation markets do not usually move this fast. We are updating our benchmarking indices monthly to keep pace.

Courtney R. Langford
President & Chief Executive Officer""",
    },
    2023: {
        'theme': 'The AI Shift',
        'desc': 'ChatGPT changes everything. ACPWB\'s research team debates what AI means for the compensation profession. We publish our first AI-augmented compensation analysis report.',
        'bg': '#0F0A1E',
        'text_color': '#E8E0F0',
        'accent': '#7B2FBE',
        'accent2': '#B480E8',
        'font_body': "'Courier New', Courier, monospace",
        'font_head': "'Courier New', Courier, monospace",
        'layout_class': 'era-ai-shift',
        'ceo': 'Courtney R. Langford',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

I want to tell you what happened in our research team's weekly all-hands meeting the week after ChatGPT launched to the public.

For the first thirty minutes, no one said anything about compensation research. We talked about what the tool could do — the quality of its writing, the fluency of its reasoning, the range of tasks it could perform. Then our head of quantitative methods said, quietly: "This is going to change what our clients think we do." She was right.

The question our researchers are debating is not whether AI will affect compensation research — it clearly will. The question is how, and how quickly. The most immediate impact is in data interpretation: the ability to synthesize large amounts of compensation data and generate coherent, readable analysis has historically been a significant portion of our billable work. If AI tools can replicate that synthesis at low cost, the value of our work has to shift toward the parts of the process that require judgment, organizational knowledge, and trust.

We published our first AI-augmented compensation analysis report in October — a document in which AI tools assisted with initial pattern identification and text drafting, with substantial human oversight and revision at every stage. The quality was high. The time investment was lower. We are being honest with ourselves about what this implies.

Courtney R. Langford
President & Chief Executive Officer""",
    },
    2024: {
        'theme': 'The AI Shift',
        'desc': 'AI compensation benchmarking becomes ACPWB\'s fastest-growing practice. Election year volatility. The profession\'s identity is actively being renegotiated.',
        'bg': '#0F0A1E',
        'text_color': '#E8E0F0',
        'accent': '#7B2FBE',
        'accent2': '#B480E8',
        'font_body': "'Courier New', Courier, monospace",
        'font_head': "'Courier New', Courier, monospace",
        'layout_class': 'era-ai-shift',
        'ceo': 'Courtney R. Langford',
        'ceo_title': 'President & Chief Executive Officer',
        'ceo_letter': """To our clients and professional colleagues:

Our fastest-growing practice in 2024 is AI Compensation Benchmarking — the work of helping organizations understand what to pay for AI-related roles, how to compensate engineers and researchers working at the frontier of the technology, and how to think about the compensation implications of AI-driven productivity changes for the rest of their workforces.

This is new territory. The demand for AI talent is intense and unevenly distributed — a handful of researchers with specific capabilities command compensation that bears no relationship to any traditional benchmarking framework, while the broader population of AI-adjacent roles is more analyzable using standard methods. We are building new instruments for a market that is still forming.

The election year has added the usual political volatility to everything: clients concerned about regulatory changes, immigration policy affecting technology workforce planning, tariff discussions affecting cost structures. We navigate it the same way we have navigated every political cycle since 1985: by focusing on what the data says and remaining independent of what any particular political outcome would prefer it to say.

I want to close the 2024 annual record with something I believe, and that I think Chester Whitmore would recognize: the fundamental work of understanding what human labor is worth has not changed. What has changed, repeatedly and dramatically, is the context in which that work is done. The organization that can hold onto the fundamentals while adapting to the context will endure. We intend to endure.

Courtney R. Langford
President & Chief Executive Officer""",
    },
}




_ARCHIVE_WORDS = [
    'report', 'summary', 'update', 'review', 'assessment', 'briefing',
    'analysis', 'memo', 'strategy', 'initiative', 'stakeholder',
    'performance', 'quarterly', 'annual', 'outcomes', 'deliverable',
    'engagement', 'alignment', 'program', 'impact',
]
