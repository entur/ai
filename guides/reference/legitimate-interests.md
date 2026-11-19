# Legitimate Interests Assessment (GDPR Art. 6 (1) f)

Require a documented legitimate interests assessment ("vurdering av berettiget interesse") whenever an Entur application, service, feature, or integration processes personal data on the basis of GDPR article 6 (1) (f).

- **Target audience**: developers, product owners, and AI agents designing or changing code, data flows, or AI usage that touches personal data.
- **Intent**: every processing that relies on legitimate interest has a reviewable, versioned assessment before processing starts, using a shared format across Entur repositories, services, and applications.
- **Scope**: when an assessment is required, how to carry out the three-part test, where to store the result, and a Markdown template. Choosing between other legal bases, full DPIAs, data processing agreements, and third-country transfers are out of scope -- this guide only says when to flag them.
- **Prerequisites**: basic familiarity with GDPR concepts (personal data, controller, data subject). Read [security.md](security.md) for technical controls and [it-systems-policy.md](../../it-systems-policy.md) for system registration and information classification.

## Why a Written Assessment Is Required

GDPR article 5 (2) makes Entur, as controller, accountable for demonstrating that every processing is lawful. Legitimate interest is the most flexible legal basis, and therefore the one that requires the most documentation: unlike consent or contract, it rests entirely on Entur's own balancing of interests. Without a written assessment Entur cannot show that the balancing took place.

Datatilsynet (the Norwegian Data Protection Authority) states in its guidance on legitimate interest that the controller must carry out and document the assessment before processing starts, and it publishes a template for this purpose. The European Data Protection Board (EDPB) sets out the same three cumulative conditions in its Guidelines 1/2024 on processing of personal data based on article 6 (1) (f), building on case law from the Court of Justice of the EU (among others C-252/21 *Meta Platforms v Bundeskartellamt* and C-621/22 *Koninklijke Nederlandse Lawn Tennisbond*). The template below follows the structure of Datatilsynet's template.

## Rules

These rules apply to all Entur repositories:

- **MUST** document a legitimate interests assessment before processing personal data on the basis of article 6 (1) (f). This includes new services, new data fields, new recipients, new purposes, AI and LLM usage, analytics, logging beyond operational need, and test or experimentation with real data.
- **MUST** update the assessment when purpose, data categories, data subjects, recipients, retention, or third-country transfer changes. A changed scope is a new assessment, not an implied extension of the old one.
- **MUST** obtain explicit approval of the drafted assessment from one or more human reviewers in the pull request before merging it into the repository. AI agents must not approve their own drafts or merge an assessment without recorded human approval.
- **MUST** keep missing evidence visible. Write "ikke dokumentert" or "ikke avklart" rather than inventing facts, and never let a gap become a positive conclusion.
- **MUST NOT** claim that a safeguard is in place, that a DPIA is done, or that the processing is approved unless the repository, the user, or an authoritative Entur source confirms it. Mark proposed safeguards as proposed.
- **MUST NOT** rely on legitimate interest alone for special categories of personal data (article 9) or data on criminal convictions (article 10). These require a separate exception; ask Personvernleder.
- **SHOULD** ask Personvernleder to review the assessment when the conclusion is uncertain, data subjects include children or other vulnerable groups, processing is large-scale or involves monitoring of employees, data leaves the EU/EEA, or the DPIA screening says a DPIA is required.

## When Legitimate Interest Is the Right Basis

Legitimate interest is one of six legal bases in article 6 (1). Check the others first -- if a more specific basis fits, use that instead and do not write this assessment:

| Basis | Typical Entur example |
|-------|-----------------------|
| (b) Contract | Selling and delivering a ticket the customer has bought |
| (c) Legal obligation | Bookkeeping records required by Norwegian accounting law |
| (a) Consent | Optional marketing the customer has actively opted in to |
| (e) Public task | Processing that a statute assigns to Entur (confirm with Personvernleder) |
| (f) Legitimate interest | Fraud prevention, network and information security, internal analytics, AI-assisted support tooling |

The last subparagraph of article 6 (1) says that public authorities cannot use legitimate interest for processing carried out in the performance of their tasks. Entur is a state-owned company; whether a given processing is part of a public task is a legal question. If the processing is required by a contract with a public authority or by statute, ask Personvernleder before choosing (f).

## Carry Out the Three-Part Test

All three conditions must be met. If one fails, legitimate interest cannot be the basis.

### 1. Purpose -- is the interest legitimate?

Describe the concrete interest Entur (or a third party) pursues. The interest must be lawful, clearly and precisely articulated, and real and present -- not speculative. Commercial interests can be legitimate, as the Court of Justice confirmed in the tennis federation case above. Recitals 47 and 49 of GDPR give fraud prevention and network and information security as examples.

Write the interest as something a reviewer can test: "Detect fraudulent ticket refunds within 24 hours" rather than "improve the service".

### 2. Necessity -- is the processing necessary for that interest?

Show that the purpose cannot reasonably be achieved with less personal data or less intrusive means. This is where most assessments fail in practice. Answer at least:

- Can the purpose be met with anonymised, pseudonymised, aggregated, or synthetic data?
- Is each data field needed, or only convenient?
- Is the retention period the shortest that serves the purpose?
- For AI usage: can identifiers be removed before the model call?

Data minimisation (article 5 (1) (c)) and necessity are assessed together.

### 3. Balancing -- do the data subjects' interests override?

Weigh Entur's interest against the data subjects' interests, rights, and freedoms. Consider:

- **Reasonable expectations** (recital 47): would the data subject expect this processing, given their relationship to Entur?
- **Nature of the data**: sensitive, financial, location, or travel-pattern data weighs heavier.
- **Data subjects**: children are given special weight in article 6 (1) (f) itself. Employees are in a position of dependency; employee monitoring must also satisfy the control-measure rules in chapter 9 of the Norwegian Working Environment Act (arbeidsmiljøloven).
- **Scale and combination**: many data subjects, or combining datasets, increases the impact.
- **Consequences**: profiling, automated decisions, exclusion, or loss of control over the data.
- **Safeguards**: pseudonymisation, access control, short retention, opt-out, transparency. Safeguards can tip the balance, but only if they are actually implemented.

## Duties That Follow From Choosing Legitimate Interest

Choosing article 6 (1) (f) triggers obligations that the assessment should point to:

- **Information** (articles 13 (1) (d) and 14 (2) (b)): the privacy notice must state which legitimate interests Entur pursues.
- **Right to object** (article 21): data subjects can object, and Entur must then stop unless it demonstrates compelling legitimate grounds. The service must be able to honour an objection.
- **Record of processing** (article 30): the processing must be in Entur's record of processing activities (behandlingsprotokoll). Ask Personvernleder how to register it.
- **DPIA** (article 35): screen for DPIA need. A legitimate interests assessment does not replace a DPIA.

## Where to Store the Assessment

Store the assessment as Markdown in the owning repository at `docs/privacy/legitimate-interests/<processing-id>.md` (default) for every Entur service, application, integration, batch job, or analytics pipeline. This keeps it next to what it documents, versioned and reviewed in a pull request. The PR contains the draft assessment. One or more human reviewers must explicitly approve the assessment in the PR before it is merged. The approved, merged file is the adopted assessment.

Use one file per processing. Use a short kebab-case `<processing-id>` that names the processing, not the technology: `refund-fraud-detection`, not `kafka-consumer`.

## Record the Conclusion

Every assessment ends with exactly one status:

| Status | Meaning |
|--------|---------|
| `supported` | Interest, necessity, and balancing are documented and support the processing, with any conditions listed. |
| `not-supported` | Legitimate interest cannot support the processing as described. Choose another basis or change the processing. |
| `insufficient-information` | Material evidence is missing. The conclusion lists what is needed. Processing must not start on this basis. |

A `not-supported` or `insufficient-information` conclusion is a valid, useful outcome -- it shows what to fix. Do not soften it to get a PR merged.

## Template

Copy the template below into the file location above and replace every `<...>` placeholder. Keep all headings, even when the answer is "ikke dokumentert". The template is in Norwegian to follow Datatilsynet's terminology.

```markdown
# Vurdering av berettiget interesse: <behandling-id>

*Vurdering av lovlig grunnlag for behandling av personopplysninger etter personvernforordningen (GDPR) artikkel 6 (1) bokstav f («berettiget interesse»).*

Denne filen dokumenterer vurderingen ut fra kildene som er oppgitt nederst. Konklusjonen angir
om berettiget interesse er dokumentert for det beskrevne omfanget. Én eller flere personer må
gjennomgå og uttrykkelig godkjenne vurderingen i PR-en før den merges til kodelageret.
KI-agenter kan ikke godkjenne egne utkast. Den godkjente, mergede filen er den vedtatte vurderingen.

- Behandling: <tjeneste, applikasjon eller integrasjon som behandler personopplysningene>
- Personopplysninger: <ja — kort beskrivelse av hvilke>
- DPIA-status: <ikke vurdert | screening gjort, DPIA ikke påkrevd | DPIA påkrevd | DPIA gjennomført — henvisning>
- Informasjonsklasser: <klassifisering registrert i Systemoversikten, eller «ikke registrert»>
- Eier: <team>
- Vurderingsdato: <ÅÅÅÅ-MM-DD>
- Beslutnings-PR: PR-en som sist endrer vurderingen (se filhistorikken)
- Status: <supported | not-supported | insufficient-information> — <én setning om utfallet>

## Behandlingen som vurderes

<Hva gjøres med personopplysningene, i hvilken tjeneste, og hvorfor. Beskriv omfanget presist,
inkludert hva som er utenfor omfanget.>

## Eier som følger opp vurderingen

<Team> følger opp vurderingen som eier av behandlingen. Beslutningen spores i PR-review.
<Angi om Personvernleder har gjennomgått vurderingen. Ikke angi en personvernfaglig signering
som ikke har skjedd.>

## Dato for vurderingen

<ÅÅÅÅ-MM-DD>. Datoen gjelder vurderingen av det dokumenterte kildegrunnlaget; vedtaket spores i
beslutnings-PR-en.

## Berettiget interesse

**Interessen av behandlingen:** <Den konkrete interessen Entur eller en tredjepart ivaretar.>

**Om interessen er berettiget:** <Er interessen lovlig, klart formulert, reell og aktuell?>

## Nødvendighet

**Om behandlingen er nødvendig for formål knyttet til den berettigede interessen:** <Kan formålet
nås med færre opplysninger, anonymiserte, pseudonymiserte, aggregerte eller syntetiske data, eller
mindre inngripende midler? Begrunn hvert opplysningsfelt og oppbevaringstiden.>

## Behandlingen

**Antall registrerte:** <Omtrentlig antall, eller «ikke dokumentert».>

**Type registrerte:** <F.eks. kunder, reisende, Entur-ansatte, ansatte hos partnere, barn.>

**Type opplysninger:** <Opplysningskategorier, inkludert avledede opplysninger. Angi eksplisitt
at artikkel 9- og 10-opplysninger ikke behandles, eller hvilket unntak som gjelder.>

**Tidsperioden for behandlingen:** Behandlingsperiode: <start og slutt, eller løpende>.
Frekvens: <f.eks. sanntid, daglig>. Oppbevaring: <oppbevaringstid og sletterutine>.

**Forholdet mellom behandlingsansvarlig og registrerte:** <F.eks. kundeforhold, arbeidsforhold,
ingen direkte relasjon.>

**Annet av betydning for behandlingen:** Mottakere: <databehandlere og andre mottakere>.
Sammenstilling av datasett: <ja/nei, hvilke>. Overføring utenfor EU/EØS: <nei, eller land og
overføringsgrunnlag>. <Status for databehandleravtale, automatiserte beslutninger og annet.>

## Personvernet til de registrerte

**Behandlingens betydning for personvernet til de registrerte (knyttet til omfanget av behandlingen, se over):**
<Rimelige forventninger, mulige konsekvenser for de registrerte, sårbare grupper, kontroll over
egne opplysninger, risiko for profilering eller kobling til enkeltpersoner.>

## Tiltak for å ivareta eller bedre personvernet til de registrerte

**Følgende tiltak skal ivareta eller bedre personvernet til de registrerte** (foreslått her — må bekreftes og faktisk gjennomføres, ikke en påstand om at de allerede er på plass):

- <Gjennomført: tiltak som er bekreftet i kode, konfigurasjon eller avtale, med henvisning.>
- <Foreslått: tiltak som ikke er dokumentert gjennomført ennå.>
- <Informasjon til de registrerte og mulighet for å protestere etter artikkel 21.>

## Interesseavveining

**De nødvendige berettigede interessene er:** <Oppsummering av interessen og nødvendigheten.>

**Behandlingen vil omfatte:** <Registrerte, omfang og opplysningskategorier.>

**Behandlingen vil ha følgende betydning for de registrertes personvern:** <Oppsummering.>

**Tiltak for å ivareta eller bedre personvernet til de registrerte:** <Oppsummering, med skille
mellom gjennomførte og foreslåtte tiltak.>

**Konklusjon:** <Begrunnet konklusjon som samsvarer med statusfeltet øverst.>

**Vilkår for konklusjonen:**

- <Vilkår som må være oppfylt for at konklusjonen skal gjelde, eller «Ingen særskilte vilkår.»>

**Oppfølging for å dokumentere grunnlaget:**

- <Ved insufficient-information: hva som må dokumenteres, og av hvem.>

## Vurdering av riktig behandlingsgrunnlag

<Hvorfor bokstav f er valgt fremfor avtale, rettslig forpliktelse, samtykke eller offentlig
oppgave. Angi om behandlingen kan være del av en offentlig oppgave, og om formålet er forenlig med
formålet opplysningene opprinnelig ble samlet inn for (artikkel 6 (4)).>

## Kildegrunnlag og usikkerhet

- Kilder: <filer, avtaler og beslutninger vurderingen bygger på>.
- Ikke dokumentert: <opplysninger som mangler og som kan påvirke konklusjonen>.
- Åpne spørsmål: <spørsmål som ikke kan avgjøres i denne vurderingen>.
```

## Agent Checklist

When an AI agent writes or changes code that processes personal data, it must:

- Identify the legal basis from the repository or ask the user. Do not assume one.
- If the basis is legitimate interest, check for an existing assessment in the owning repository at `docs/privacy/legitimate-interests/`.
- If the change widens the scope (new fields, purposes, recipients, retention, or transfers), update the assessment in the same PR or flag that it must be updated.
- When drafting an assessment, fill only what the sources confirm, mark everything else as not documented, and set the status accordingly.
- Submit the draft assessment in a PR and obtain explicit approval from one or more human reviewers before merging it into the repository. Do not approve your own draft or merge without recorded human approval.
- Never state that Personvernleder, Datatilsynet, or any reviewer has approved the processing unless the user or a document confirms it.

## Further Reading

These sources explain the rules behind this guide:

- Datatilsynet's guidance on legitimate interest ("berettiget interesse") and its template for the assessment -- the authoritative Norwegian source and the basis for the template above.
- Datatilsynet's guidance on DPIA ("vurdering av personvernkonsekvenser") and its list of processing that always requires a DPIA.
- EDPB Guidelines 1/2024 on processing of personal data based on article 6 (1) (f) GDPR -- the European interpretation of the three-part test, with examples.
- GDPR articles 5, 6, 13, 14, 21, 30, and 35, and recitals 47 to 49, as incorporated into Norwegian law by the Personal Data Act (personopplysningsloven).
- [security.md](security.md) -- technical controls such as access control and secret handling that serve as safeguards.
- [incident-response.md](../playbooks/incident-response.md) -- escalation to Personvernleder for personal data breaches.
