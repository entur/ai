---
name: hora
description: Use this skill whenever the user wants to create a meeting agenda, plan a meeting, structure a meeting, or prepare for a meeting. This skill applies the HØRA template — Hensikt, Ønsket resultat, Ressurser, og Agenda. The goal is to produce clear, purpose-driven meeting agendas. Trigger this skill for any request involving meeting preparation, agenda writing, or meeting structure, even if the user does not explicitly mention HØRA.
---
 
# HØRA Meeting Agenda Skill
 
Use the HØRA template to structure every meeting agenda. HØRA ensures meetings have a clear purpose, a defined desired outcome, and a concrete agenda.
 
## The HØRA Template
 
### Hensikt (Purpose)
State *why* this meeting is being held. One or two sentences max. Be specific — avoid vague phrases like "to discuss X". Instead: "To decide X" or "To align on X before Y deadline."
 
### Ønsket resultat (Desired outcome)
State what a successful meeting looks like when it ends. This is concrete and measurable:
- A decision made
- A plan approved
- A problem solved
- A shared understanding reached
### Ressurser (Resources / Participants)
List who needs to be present and any materials, documents, or preparation required before the meeting.
 
### Agenda
A numbered list of agenda items with time allocations. Each item should map directly back to the Ønsket resultat.
 
---
 
## Output Format
 
Always produce the agenda in Norwegian unless the user requests another language. Use this structure:
 
```
## Hensikt
[Én til to setninger om hvorfor møtet holdes]
 
## Ønsket resultat
[Konkret og målbart: hva skal være oppnådd når møtet er ferdig?]
 
## Ressurser
**Deltakere:** [Navn / roller]
**Forberedelser:** [Dokumenter, lesing, eller oppgaver som må gjøres på forhånd]
 
## Agenda
1. [Punkt] — [X min] - [Ansvarlig]
2. [Punkt] — [X min] - [Ansvarlig]
3. [Punkt] — [X min] - [Ansvarlig]
...
 
**Total tid:** [X min]
```
 
---
 
## Guidance
 
- **Hensikt** drives everything. If the purpose is unclear, ask the user before proceeding.
- **Ønsket resultat** must be a concrete deliverable or decision, not an activity ("we discussed") but an outcome ("we decided").
- **Agenda items** should be ordered logically and timed. Flag if total time seems unrealistic.
- Keep the agenda focused. If the user lists too many items for the time available, flag this and suggest prioritizing or splitting into multiple meetings.
- If the user provides partial information, fill in reasonable defaults and note any assumptions made.