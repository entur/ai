# Guidelines for IT Systems and Software

- **Target audience**: AI agents and Entur contributors working with IT systems or software.
- **Intent**: keep system ownership, access control, information classification, licensing, and software lifecycle decisions compliant with Entur policy.
- **Scope**: SaaS products, internal tools, desktop and mobile apps, browser extensions, and developer tools introduced, recommended, configured, documented, or automated for Entur.

Entur's preferred model is Software as a Service (SaaS) connected to Entur user accounts through Microsoft Entra ID.

## Agent Checklist

When a task involves IT systems or software, AI agents must apply these checks before making or recommending a change:

- Prefer SaaS that supports Microsoft Entra ID and named Entur user accounts.
- Check whether the system or software is registered in the System Overview on the DAP portal.
- Identify the System Owner. Do not assume ownership when it is not documented.
- Require Single Sign-On (SSO) with Microsoft Entra ID for internal systems.
- Prefer automatic user provisioning when the product supports it.
- Ensure the data used by the system is classified in the System Overview.
- Avoid adding software, apps, or browser extensions without a clear work-related need.
- Avoid unused licences. Remove or reassign inactive licences instead of buying more.
- Do not claim that a system has been registered, classified, or approved unless the repository, user, or DAP documentation confirms it.

If you cannot verify these points from the repository or available documentation, call out the missing information and ask the user or relevant Entur team to verify it.

## System Overview

All IT systems and software used internally at Entur must be registered in the System Overview on the DAP portal.

The System Overview records basic information about each system, including who holds each role. The most important management role is System Owner. Role responsibilities are described on the DAP portal.

Contact Digital Workplace in the `#open-internsupport` Slack channel if you notice missing or incorrect information in the System Overview. Everyone must help keep the overview accurate and useful.

## User Administration

Internal IT systems must use SSO with Microsoft Entra ID as the identity provider. Enable automatic user provisioning where the product supports it.

If a system does not use SSO, the System Owner must contact DAP. AI agents must not propose bypasses, shared accounts, or local user administration as a default solution.

## Information Classification

Data used in an IT system must be classified so Entur can choose the correct security measures.

The classification must be registered in the System Overview. If the classification is unknown, treat it as missing required information and ask for clarification before recommending security controls.

## Licence Management

Licences that are not actively used must be removed and reassigned when needed. Do not recommend buying additional licences until inactive licences have been checked.

## Work-Related Use

Entur's IT systems and software must only be used for work-related activity.

Do not install, recommend, or automate arbitrary software on Entur computers, mobile phones, or tablets. Software should have a clear business need, a System Owner where relevant, and a path for support and lifecycle management.

## Software Updates

Everyone at Entur is responsible for keeping their computer, mobile phone, browsers, and installed software up to date with security updates.

Regular restarts are necessary because some updates are only installed after a restart. Guides for keeping Windows, macOS, and browsers such as Chrome, Firefox, and Edge up to date are available on the DAP portal.
