# PHASE 9: Documentation Focus Plan (Lean and Practical)

เอกสารนี้เป็นแผนควบคุม scope งานเอกสาร เพื่อให้ทีมโฟกัสเฉพาะสิ่งที่กระทบการใช้งานจริงก่อน และหยุดการขยายเอกสารที่ยังไม่จำเป็นในรอบนี้

## TL;DR
- เป้าหมายรอบนี้: ลดความซับซ้อนของเอกสาร และเพิ่มความพร้อมใช้งานจริง (operations-ready)
- หลักคิด: Write less, decide faster, execute clearer
- ผลลัพธ์ที่ต้องได้: มีลิสต์ Now/Next/Later ชัดเจน พร้อม owner, Definition of Done, และ dependency

## Scope Boundary

### In Scope (รอบนี้)
- จัดลำดับเอกสารที่ต้องทำก่อนตามผลกระทบทางธุรกิจและปฏิบัติการ
- กำหนด owner role ต่อเอกสารแต่ละชิ้น
- กำหนด Definition of Done (DoD) และเกณฑ์ตรวจรับ
- เพิ่มจุดเข้าถึงแผนจากเอกสารหลัก

### Out of Scope (รอบนี้)
- เขียนเอกสารทุกหัวข้อให้ครบทั้งหมดในครั้งเดียว
- แตกเอกสารย่อยใหม่จำนวนมากโดยยังไม่มี owner และ deadline
- ทำเอกสารเชิงทฤษฎีที่ไม่เชื่อมกับการ deploy/run/support จริง

## Prioritized Backlog

## Now (Must-Have, Execution Critical)

| Priority | Document | Why Now | Owner Role | Definition of Done |
|---|---|---|---|---|
| P1 | Deployment and Environment Runbook | ทำให้ deploy dev/staging/prod ซ้ำได้และ rollback ได้ | Platform/Ops Lead | มีขั้นตอน deploy, promote, rollback, smoke check และตัวอย่างค่าคอนฟิกแต่ละ environment |
| P1 | Database Migration Strategy | ลดความเสี่ยง schema drift และ deploy fail | DB Owner/Backend Lead | มี versioning, apply order, validation checklist, rollback pattern, และ pre-release check |
| P1 | Config and Secrets Management | ลด incident จาก config mismatch และ secret leakage | Security/Platform | มี env matrix, secret source-of-truth, rotation policy, และ onboarding checklist |
| P1 | Incident Response and Troubleshooting Runbook | ลด MTTR ตอนระบบล่มหรือ alert ขึ้น | SRE/Ops On-call | มี alert-to-action mapping, triage steps, escalation path, และ post-incident template |
| P2 | CI/CD and Integration Testing Map | ลด regression และเพิ่ม release reliability | Dev Productivity/QA | มี pipeline stages, quality gates, required checks, และ integration test ownership |

## Next (Risk Reduction)

| Priority | Document | Why Next | Owner Role | Definition of Done |
|---|---|---|---|---|
| P2 | Service API Contract Baseline | ป้องกัน interface drift ระหว่างบริการ | Backend Leads | ระบุ contract source, versioning rule, deprecation policy, และ change process |
| P2 | Performance Baseline and Capacity Guide | วางเกณฑ์ scale จากข้อมูลจริง | Platform + Service Owners | มี p50/p95/p99 เป้าหมายหลัก, throughput baseline, scaling triggers |
| P2 | Security Hardening Checklist | ลดช่องโหว่พื้นฐานก่อนขยายระบบ | Security Lead | มี checklist ครอบคลุม auth, transport, secrets, dependency, audit logging |

## Later (Governance Maturity)

| Priority | Document | Why Later | Owner Role | Definition of Done |
|---|---|---|---|---|
| P3 | Data Retention and Privacy Policy | ใช้เมื่อเริ่ม governance/compliance เต็มรูปแบบ | Data Governance | ระบุ data class, retention windows, purge rules, audit evidence |
| P3 | Multi-Region and DR Strategy | ใช้เมื่อมี workload และงบพร้อมขยาย region | Platform Architecture | ระบุ RTO/RPO targets, failover model, DR drill cadence |

## Dependency Order

1. Deployment and Environment Runbook
2. Database Migration Strategy
3. Config and Secrets Management
4. CI/CD and Integration Testing Map
5. Incident Response and Troubleshooting Runbook
6. API Contract Baseline / Performance / Security (parallel)

## Stop-Doing List (Anti-Sprawl Rules)
- หยุดเพิ่มเอกสาร phase ใหม่ที่ยังไม่ผูกกับงาน deploy/run จริง
- หยุดแปลเอกสารทุกไฟล์แบบสองภาษาโดยไม่มี stakeholder ใช้งาน
- หยุดแตกเอกสารใหม่ถ้ายัง merge เข้าเอกสารเดิมได้โดยไม่เสียความชัดเจน
- หยุดเพิ่ม dashboard doc ซ้ำ ถ้ามี source-of-truth อยู่แล้วใน KPI docs

## Review Cadence
- รอบทบทวน: ทุก 2 สัปดาห์ในช่วง 2 เดือนแรก
- วาระทบทวน: completed items, blockers, re-prioritization, archive candidates
- กติกาเลื่อนงาน: ถ้าไม่มี owner และ DoD ชัดเจน ให้เลื่อนไป Later โดยอัตโนมัติ

## Success Metrics
- ลดจำนวนเอกสารใหม่ที่ไม่ถูกใช้งานจริง
- เพิ่มความเร็ว onboarding ทีมใหม่ในเรื่อง deploy/runbook
- ลดเวลาเฉลี่ยการแก้ incident (MTTR)
- เพิ่มความสม่ำเสมอของ release quality gate

## Exit Criteria For Phase 9
- มีเอกสาร Now เสร็จอย่างน้อย 3 จาก 5 หัวข้อ
- เอกสารทุกหัวข้อใน Now มี owner role และ DoD ครบ
- ลิงก์เอกสารแผนเข้าถึงได้จาก README หลักและ docs/systems hub
