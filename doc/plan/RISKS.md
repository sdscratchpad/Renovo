# Risk Register (Demo Program)

## R1: Demo fails due to unstable environment
- Impact: High
- Mitigation: isolate demo cluster, seed deterministic data, run pre-demo health checklist.

## R1a: Laptop resource limits degrade demo performance
- Impact: High
- Mitigation: define minimum laptop profile, cap workload sizes, pre-pull images, and pre-warm services.

## R2: AI suggests incorrect remediation
- Impact: High
- Mitigation: constrain actions to approved runbooks, enforce manual approval by default.

## R3: Detection noise creates false alarms
- Impact: Medium
- Mitigation: tune thresholds with replay data and keep scenario-specific detection rules for MVP.

## R4: Remediation takes too long in live demo
- Impact: Medium
- Mitigation: pre-warm dependencies, choose fast-recovering scenarios, keep backup scenario.

## R4a: Local cluster drift across repeated runs
- Impact: Medium
- Mitigation: include deterministic reset script and rebuild `kind` cluster when baseline checks fail.

## R5: Scope creep delays delivery
- Impact: High
- Mitigation: strict 3-scenario scope lock and weekly scope review.

## R6: Customer asks production-readiness questions
- Impact: Medium
- Mitigation: prepare clear boundary statement and phased roadmap slide.
