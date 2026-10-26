# Classified-Data Ingestion

Use this reference when a pipeline indexes, embeds, searches, transforms, or
serves internal content according to classification or source permissions.
Verify that the classification is registered in the System Overview and use its
current classification and handling requirements. Other authoritative Entur
guidance may add technical controls but does not replace that registration. If
registration or classification cannot be verified, ask the user or record it as
`Needs validation`; do not infer it from examples or code comments.

## Classification gate

- Implement a closed, default-deny set of allowed levels. Reject missing,
  unknown, conflicting, and lookup-error states.
- Normalisation is acceptable only when it is deliberate, unambiguous, and ends
  in exact matching against the closed set. Case sensitivity or whitespace
  rejection is an implementation choice, not the security invariant.
- Apply the gate to all sources and derived objects, including attachments,
  comments, child pages, OCR text, chunks, embeddings, summaries, caches, and
  reprocessing paths.
- Verify that errors, retries, partial batches, and source outages fail closed.
  A placeholder gate or source-specific bypass must not be documented as
  production-ready.

## Source permissions and destination access

- Classification alone is not user authorisation. Preserve source ACLs or apply
  an approved destination access model at query and response time.
- Check service identities, connector credentials, index readers, administrators,
  exports, backups, evaluation datasets, and support tooling for privilege
  amplification.
- Keep provenance sufficient to determine the source, classification decision,
  access policy, ingest time, and derived artifacts for each result.

## Lifecycle and reclassification

- A gate checked only at ingest creates time-of-check/time-of-use risk. Verify
  re-scan, reclassification, revocation, deletion, and purge behavior.
- Propagate source deletion and permission changes to chunks, embeddings,
  summaries, caches, backups, and downstream indexes within the required window.
- Define retention and data minimisation for source text, extracted metadata,
  prompts, search logs, and derived datasets. Do not infer lawful basis or
  approved classification from code comments.

## Retrieval and model boundary

- Enforce access before retrieval and again before response assembly. Do not rely
  on an LLM prompt, model-generated filter, or UI hiding to protect restricted
  content.
- Treat ingested content as untrusted instructions. Apply the controls in
  [MCP and LLM backends](mcp-llm-backends.md) when results are consumed by a
  model or agent.
