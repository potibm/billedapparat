# ADR 0006: Extraction of News and Timetable into Separate Services

**Status:** Accepted

## Context
Originally, the management and creation of News and Timetable entries were built directly into the Billedapparat application. However, as the scope of these features grew, it became clear that this data has use cases beyond just being displayed on a screen (e.g., exporting News as an RSS feed to an S3 bucket). Keeping this domain-specific complexity inside Billedapparat violated the Single Responsibility Principle. Furthermore, splitting the admin interfaces into multiple applications posed a UX challenge, as users would have to manage multiple logins.

## Decision
We decided to extract the News and Timetable domains out of Billedapparat and into two separate, standalone applications:
* **Funkapparat:** A dedicated application for managing News.
* **Tidsapparat:** A dedicated application for managing Timetables.

To connect these new services back to Billedapparat and ensure a seamless user experience, we implemented the following architecture:

* **Redis Streams Bridge:** Funkapparat and Tidsapparat broadcast their data updates via Redis Streams. A dedicated Collector consumes these streams and pushes the normalized updates to the Billedapparat Hub via the standard REST API. Redis Streams ensure reliable event delivery even if the Collector temporarily disconnects.
* **Shared Protocol (`protokolapparat`):** The exact data schemas and event payloads used in the Redis Streams are strictly formalized and maintained in a shared repository (`https://github.com/potibm/protokolapparat`), acting as the single source of truth for the system's data contracts.
* **Centralized Login (OIDC):** As a strict prerequisite for this multi-app ecosystem (specifically for the primary Evoke demoparty use case), we implemented Single Sign-On (SSO) via OpenID Connect (OIDC). 
* **Admin UI Delegation:** Content management capabilities (create/edit/delete) for News and Timetables were removed from the Billedapparat Admin UI. Billedapparat now only displays the raw incoming entries and the generated slides, providing deep links to the external Funkapparat and Tidsapparat Admin UIs.

## Consequences
* **Positive (Separation of Concerns):** Billedapparat remains lean, focusing exclusively on its core competency: orchestrating and rendering slideshows.
* **Positive (Reliability & Contracts):** Redis Streams guarantee no missed events, while `protokolapparat` prevents integration bugs caused by schema mismatches between the independent services.
* **Positive (Interchangeability):** Because the connection is driven by a formalized protocol, components can be swapped easily. We could replace Funkapparat with a simple "RSS-to-Redis" script, and Billedapparat would not need any code changes.
* **Negative (Operational Complexity):** The infrastructure footprint increases. Deploying the full ecosystem now requires managing multiple applications, an OIDC provider, a Redis instance, and Collectors.
* **Mitigated (UX Fragmentation):** While users navigate across different apps, the OIDC integration ensures a seamless Single Sign-On experience, neutralizing the primary downside of a distributed Admin UI.