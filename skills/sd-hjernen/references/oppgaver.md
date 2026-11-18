# The task recipes

Each recipe says which pages to search, which to read, and how to shape
the answer.
Every heading gets its pages linked. Under a heading where you found
no pages, write one line that says so, and keep the heading, unless the
recipe says to leave that heading out. The page shapes are
in `sider.md`.

## A brief before a meeting with a partner

When the user is going to meet a partner and wants to prepare, or asks
what the knowledge base knows about one partner.

Search and read:

1. The partner page: search `space KI1 <partner> mandat prioriteringer`
   and read the page titled with the partner's name.
2. The problem pages that concern the partner: search `space KI1
   problemstilling Gjelder: <partner>` and read each problem page whose
   «Gjelder» line names the partner. Pages marked «Hele sektoren»
   concern every partner. Read those whose title touches the partner's
   own domain. List the others by title and link only, and say that you
   did not read them. State nothing about what a page says unless you
   read the page.
3. The data page, if the user asks about data: search `space KI1 Data
   fra <partner>`.

Answer under these headings, in this order. Link the partner page
under every heading that draws on it, even when that is the same page
three times.

- «Rolle og mandat»: what the partner is for and must do, from the
  partner page, with the source document named.
- «Mål og prioriteringer for <year>», with the year as the partner page
  has it: what the partner must achieve that year.
- «Problemstillinger som gjelder <partner>»: each problem page, with
  the finding in one or two sentences and the other partners it
  concerns.
- «Relevant for datadeling og Samferdselsdata»: the parts of the
  mandate and the priorities about data, from the partner page, and the
  data page if it was read.

If the user asks for questions to bring to the meeting, add «Forslag
til spørsmål» at the end and say that the questions are yours, built on
the pages above.

## What Nasjonal transportplan says

When the user asks what NTP prioritises, on a topic or in general, or
what the national guidelines say.

Search and read:

1. The NTP summary: search `space KI1 Oppsummering: NTP <topic>` and
   read the page titled «Oppsummering: NTP». It covers the plan for 2025
   to 2036.
2. The documents written for the next plan: search `space KI1
   Oppsummering ny NTP <topic>` and read each page whose title says it
   is an answer, a report, or a study for the next plan.
3. Other source summaries that the search finds on the topic, titled
   «Oppsummering: <document>».

Answer under these headings, in this order:

- «NTP 2025–2036»: the goals and priorities on the topic, with amounts
  where the summary gives them, and what the summary says under
  «Relevant for data og digitalisering».
- «Dokumenter til neste NTP»: what the documents written for the next
  plan say on the topic, each named. There is no new plan yet. Say so
  if the user seems to think there is.
- «Andre nasjonale føringer»: other source summaries on the topic, if
  any were found. Leave the heading out if none were.

## Which data exists on a topic

When the user asks what data the partners share on a topic, or what
data one partner shares.

Search and read:

1. Search `space KI1 Data fra <topic words>` and, for a named partner,
   `space KI1 Data fra <partner>`. Read each data page the hits point
   to.
2. If a problem page or a partner page mentions a data source on the
   topic, name it too, with its page.

Answer with one heading per publisher, «Data fra <publisher>», linked
to the data page. Under it, one line per dataset: the name, «ikke
verifisert» if the page marks the dataset so, the access as the page
states it (`åpen`, `krever registrering`, `begrenset`, or `uavklart`),
the access route, such as REST-API, GraphQL-API, SIRI, DATEX II,
filnedlasting, WFS, or WMS, the format, and one sentence about what it
holds. Where the page states no access route or no format, write «ikke
oppgitt». The owner of every dataset under a heading is that
publisher. Do not invent a dataset, an access, an access route, or a
format that the page does not state.

## Grounding an idea, a plan, or a draft

When the user has an idea, a plan, a draft, a presentation, or a
dashboard and wants to know how it sits with what the sector says, or
asks you to ground it.

Search and read:

1. Pick the topics in the user's text, two to four of them.
2. For each topic, search the problems (`space KI1 problemstilling
   <topic>`), the partners the text names or implies (`space KI1
   <partner> mandat prioriteringer`), and the national guidelines
   (`space KI1 Oppsummering: NTP <topic>`).
3. Read the pages that bear on the text. Skip the rest.

Answer under these headings, in this order:

- «Støtte i kunnskapsbasen»: each point in the user's text that a page
  supports, with the page linked and the finding or priority that
  supports it.
- «Motstand eller hull»: each point that a page contradicts, and each
  point that you found no pages on. Say which is which.
- «Problemstillinger som ideen adresserer»: the problem pages the text
  addresses, each with the partners it concerns.
- «Forslag til forankring»: the changes that would tie the text more
  closely to the problems, the mandates, and the priorities. Mark these
  as yours.

Keep the user's text as it is. You ground it, you do not rewrite it,
unless the user asks for a rewrite.

## Ideas tied to the knowledge base

When the user asks for ideas, suggestions, or opportunities on a topic
or for a partner.

Search and read as for an overview of the topic: the problems, the
partners, the national guidelines, and the data. Read the problem pages
and the data pages the ideas will rest on.

Give three to five ideas. For each, in this order: a title, two or
three sentences on what it is, then «Bygger på»: the problem page it
addresses, the partner whose mandate or priority it serves, the
national priority it follows, and the data it would use, each linked.
An idea with nothing under «Bygger på» is not an idea for this task.
Say which parts are yours: the idea itself is yours, the pages under
«Bygger på» are the knowledge base's.

## Content for a tool

When the user builds a dashboard, a presentation, a mockup, a report,
or another tool, in a chat or in a coding agent, and the tool shows
anything about the sector, the partners, the problems, the priorities,
or the data.

Before you put content into the tool, follow «Content that leaves the
chat» in `SKILL.md`.

Take the content from the knowledge base, not from your own knowledge.
Search and read as for an overview of each topic the tool covers. For
a list of partners, problems, priorities, or datasets, take the list
from the pages, as complete as the searches find it, and keep the page
titles as the names. Say in the tool if the list may be incomplete.

Put the source into the tool itself: every item that comes from a page
carries the page's title and URL, in a way that fits the tool, such as
a link, a footnote, a tooltip, or a source column. Where the tool needs
something the knowledge base does not have, such as a number the pages
do not give, say so in the tool and in the chat, and mark what you
filled in from elsewhere. Do not make up a figure to fill a chart.
