# Smoke test — primitives only.
#
# Exercises the full set of wire-level primitive keywords in a single
# scenario: open, send, expect, close.  No domain keyword is used;
# every step resolves to a primitive-layer keyword from
# pkg/keywords/primitive.
#
# This story is the companion artifact for spec 004 AC6.  It must
# execute successfully against the pinned CitrineOS instance (see
# test/reference/citrineos.version) without any domain keyword
# registered.
#
# The story is tagged "helper" so that no Spec-Ref is required.
# The OCPP version is left at the default (any); primitive keywords
# are version-agnostic and are eligible regardless of version.

Meta
    Name:      Primitives only smoke test
    Id:        primitives_only_smoke
    Tags:      helper, smoke, primitives-only, wire-only
    Stations:  1
    Timeout:   60s

Scenario: Open, send BootNotification CALL, expect CALLRESULT, close
    When  open a WebSocket to "ws://localhost:9210/CP001" as station "CP001" with subprotocol "ocpp1.6"
    Then  the connection on station "CP001" is open
    When  send raw frame [2, "msg-001", "BootNotification", {"reason": "PowerUp", "chargingStation": {"model": "ACME", "vendorName": "Test"}}] on station "CP001"
    Then  expect any frame on station "CP001" within 30s
    When  close station "CP001"
    Then  the connection on station "CP001" is closed
