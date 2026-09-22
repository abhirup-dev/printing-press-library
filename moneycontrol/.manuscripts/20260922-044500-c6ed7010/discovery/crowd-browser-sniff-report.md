# Moneycontrol crowd-sniff report

## Result

The Printing Press crowd-sniff command was run with `moneycontrol` and base URL `https://www.moneycontrol.com`. It exited with code 4: `no endpoints discovered for "moneycontrol"`. No generated crowd spec is trusted or used.

## npm Packages Analyzed

No relevant npm SDK package was returned by crowd-sniff. No package endpoint evidence was accepted.

## GitHub Repos Searched

Crowd-sniff returned no endpoint-bearing GitHub evidence. Independently recovered PR #1701 (`mvanhorn/printing-press-library`, fork branch `abhirup-dev/printing-press-library:feat/moneycontrol`) remains historical implementation evidence, not crowd-sniff output.

## Endpoints Discovered

None by crowd-sniff. The build will use the live EgoBrowser discovery evidence and manually validated public routes documented in the research brief.

## Base URL Resolution

No community base URL was resolved. The build must preserve the separately validated hosts: `www.moneycontrol.com`, `priceapi.moneycontrol.com`, and `api.moneycontrol.com`, with the dedicated priceapi client behavior and encoded index keys.

## Auth Patterns Detected

None by crowd-sniff. Auth/Pro behavior was validated separately through sanitized page-generated traffic in EgoBrowser TaskSpace 7; no credentials, cookies, tokens, or session values were written.

## Parameter Name Evidence

None by crowd-sniff. Parameter evidence comes from the research brief and live browser capture.

## Coverage Summary

Crowd-sniff: 0 endpoints, 0 resources. Fallback: use the existing discovery report, page-generated browser evidence, and a hand-authored/reconciled spec rather than treating the empty crowd result as proof that Moneycontrol has no usable surfaces.
