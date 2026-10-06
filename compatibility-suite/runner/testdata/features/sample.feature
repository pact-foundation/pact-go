Feature: Sample
  Background:
    Given it passes

  Scenario: passes
    Then it passes

  Scenario: fails
    Then it fails

  Scenario: undefined
    Then nobody defined this step

  Scenario: unsupported
    Then it is unsupported

  Scenario: panics
    Then it panics

  Scenario: skipped
    Given it skips
    Then it fails
