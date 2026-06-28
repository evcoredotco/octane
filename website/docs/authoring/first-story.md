---
sidebar_position: 1
---

# Authoring Your First Story

A story starts with `Meta`, declares its stable ID and OCPP traceability,
then uses keyword steps to drive the CSMS.

```text
Meta
    Name:        Boot notification accepted
    Id:          boot_notification_accepted
    Spec-Ref:    OCPP-J 1.6 -6.5 BootNotification
    Tags:        boot, conformance
    Stations:    1

Scenario: station boots
    When station "CP01" sends BootNotification with reason "PowerUp"
    Then the CSMS responds with status "Accepted" within 30s
```

For authoring rules, see `docs/concepts/story-syntax.md`.

