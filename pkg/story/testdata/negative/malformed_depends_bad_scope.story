Meta
    Name:      Malformed Depends Bad Scope Fixture
    Id:        malformed_depends_bad_scope_fixture
    Spec-Ref:  OCPP 2.0.1 §B01 BootNotification
    Tags:      conformance
    Stations:  1
    Depends:
      - id:    some_dependency
        scope: once-per-universe

Scenario: Dummy scenario
    When  something happens
    Then  something is verified
