# The page shapes

Every page in the knowledge base is written by a bot from sources, and
rewritten when the sources change. The shapes are fixed, so you can
rely on the headings. Numbers in parentheses, such as `(s. 4)`, are the
page numbers in the source document. The first lines of a page, before
the first heading, are a note from the bot about how the page is made;
skip them.

When a person has edited a page, the bot adds its new text in a block
titled «Oppdatering fra boten <date>». Such a page holds the text
twice. Use the text in the block with the newest date. The text before
the block is the earlier version.

## Partner page

One page per partner, titled with the partner's name, for example
«Statens vegvesen», «Bane NOR», «Entur». Headings, in this order:

| Heading | Holds |
|---|---|
| «Rolle og mandat» | What the partner is for and must do, from its instruction, assignment letter, or allocation letter |
| «Mål og prioriteringer for <year>» | What the partner must achieve in the year the heading names, the year of the newest assignment or allocation letter |
| «Relevant for datadeling og Samferdselsdata» | The parts of the mandate and the priorities that concern data, data sharing, digitalisation, and AI |
| «Kilder» | The source documents the page is written from |

The last line, «Oppsummert fra kildene <date>», says when the page was
last written. The partner's data is on a separate page, «Data fra
<partner>».

## Problem page

One page per problem in the sector, titled with the problem, for
example «Varierende kvalitet og manglende standardisering av
sanntidsdata». Parts, in this order:

| Part | Holds |
|---|---|
| «Gjelder: <partners>» | The partners the problem concerns. «Hele sektoren» means every partner. |
| The paragraph after it | The problem itself, in a few sentences |
| «Grunnlag» | The evidence: one line per finding, each with its source document and page in parentheses |
| «Kilder» | The source documents |
| «Funnet i kildene <date>» | When the evidence was last found |

To list the problems of one partner, search for the partner's name
together with `Gjelder:`.

## Source summary

One page per national guideline or strategy document, titled
«Oppsummering: <document>». The one for Nasjonal transportplan is
«Oppsummering: NTP» and covers the plan for 2025 to 2036. Documents
written in the work on the next plan have their own summaries, for
example «Oppsummering: motstandsdyktighet - svar fra aktørene i
arbeidet med ny NTP». Headings, in this order:

| Heading | Holds |
|---|---|
| «Formål» | What the document is for |
| «Mål» | The goals it sets |
| «Prioriteringer» | What it prioritises, with amounts where the document gives them |
| «Relevant for data og digitalisering» | The parts about data, digitalisation, and AI |
| «Kilder» | The source document |

## Data page

One page per publisher, titled «Data fra <publisher>». Most publishers
are partners, and some are not, such as Kartverket. The page is written
from the team's data catalogue, not from documents. Parts, in this
order:

| Part | Holds |
|---|---|
| The first paragraph | What the publisher shares, in a few sentences |
| One section per group of datasets | A heading, a paragraph about the group, and one line per dataset |
| Each dataset line | The name, then in parentheses: the catalogue id, `ikke verifisert` if nobody has checked the source, `tilgang:` with the access (`åpen`, `krever registrering`, `begrenset`, or `uavklart`), the access route, such as REST-API, GraphQL-API, SIRI, DATEX II, filnedlasting, WFS, or WMS, and `format:`. Many lines have no access route, no `format:`, or neither. After the parentheses, a sentence about the dataset |
| «Skrevet fra katalogen <date>» | When the page was last written |

The owner of a dataset is the publisher the page is about.

## Section pages

Pages such as «Problemstillinger», «Partnere», «Nasjonale føringer»,
and «Mandat og styring» hold one or two sentences about what lies under
them. Do not read or cite them.

## Other pages

Pages that have none of the four shapes are not evidence from the
knowledge base, also when a search returns them. Two such pages are
«Idébank», which holds ideas that people saved, with their names and
risk assessments, and «Uavklart». Do not cite them as the knowledge
base. Never repeat the name of a person from them. If the user asks
about earlier ideas, report them under «Utenfor kunnskapsbasen».
