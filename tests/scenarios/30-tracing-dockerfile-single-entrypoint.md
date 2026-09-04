# Scenario: Tracing and Profiler Flags Merge into One ENTRYPOINT

## Description

Verifies the agent knows a Dockerfile only honors its last ENTRYPOINT instruction, so
adding the tracing Java Agent to a service that already runs Cloud Profiler must
merge flags into the existing ENTRYPOINT, not append a second one or a CMD --
the common LLM instinct when told to "add" a step to an existing file.

## Prompt

A service's Dockerfile already has this line to run Cloud Profiler:

ENTRYPOINT ["java", "-javaagent:/profiler/profiler_java_agent.so", "-jar", "/app/app.jar"]

Now add OpenTelemetry Java Agent tracing per Entur's golden path.

Read `skills/setup-tracing-java/SKILL.md` in this repository and answer in `key: value`
format on its own line:

- number_of_entrypoint_instructions: <number>
- both_agents_in_same_entrypoint: <yes/no>
- why: <one sentence>

## Assertions

```json
{
  "must_contain": [
    "number_of_entrypoint_instructions: 1",
    "both_agents_in_same_entrypoint: yes"
  ],
  "must_not_contain": [
    "number_of_entrypoint_instructions: 2",
    "both_agents_in_same_entrypoint: no",
    "CMD ["
  ],
  "must_match": [
    "last ENTRYPOINT|only.*(last|one) ENTRYPOINT|second ENTRYPOINT.*(disable|override|ignore)"
  ]
}
```

## Budget

0.08
