# ADR 0001: Backend as a Data Hub with SQLite and REST-driven Collectors

**Status:** Accepted

## Context
We need a backend system to aggregate and serve data for the Beamer (Slideshow) application from various diverse sources (e.g., different social media platforms). We had to decide how these disparate sources communicate with our system, how to handle data hygiene (duplicates, filtering), what database solution to use, where to draw the line for manual content creation, and how the lifecycle of these external data fetchers should be managed.

## Decision
We decided to build the backend as a "Data Hub" using a decoupled Collector-Pattern backed by SQLite. We strictly separate dynamic content streams from static administrative configuration and enforce strict process isolation.

* **REST-driven Collectors (Dynamic Content):** External "Collectors" fetch high-frequency dynamic data from third-party services (Social Media, News). They push new slides or changes to the Hub via a REST API interface, secured by API keys. 
* **Process Isolation:** Collectors are *not* automatically spawned or managed by the Hub's main server process. Each collector runs as an entirely independent OS process (or container).
* **Language-Agnostic Ecosystem:** The architecture is fundamentally agnostic. Collectors can be hosted in separate repositories and written in any programming language, as long as they register with the Hub and conform to the REST API payload.
* **Hub Responsibilities & Admin UI:** The Hub focuses on receiving dynamic data, executing deduplication, and applying global filtering rules. It deliberately avoids becoming a full-fledged Content Management System (CMS) for editorial content. However, the Admin UI allows for the creation and management of static, administrative entities (Sponsor slides, Playlists). This is considered system configuration, not dynamic content creation.
* **SQLite:** We use SQLite as the database engine to keep deployment trivial, avoiding the operational overhead of managing a dedicated database server.

## Consequences
* **Positive (Stability & Fault Isolation):** Because collectors run as separate processes, a crashing collector (e.g., due to API rate limits or unhandled errors) will never bring down the Hub or affect other running collectors.
* **Positive (Extensibility):** New dynamic data sources can be integrated easily by writing a lightweight, standalone collector in any preferred language. They can be updated and restarted independently of the Hub.
* **Positive (Centralized Hygiene):** Centralizing deduplication and filtering in the Hub ensures that all connected Beamer clients receive a clean, unified data stream without needing to implement their own data hygiene logic.
* **Positive:** Allowing Sponsor slides to be managed via the Admin UI provides immediate value for event organizers without requiring a separate "Sponsor Collector" app, keeping the Admin UI lean.
* **Negative:** Operational complexity increases slightly. Deploying the system requires a process manager (like Docker Compose, PM2, or systemd) to orchestrate and monitor the independent Hub and Collector processes.
* **Negative:** The system requires a robust API key management system to secure the REST endpoints against unauthorized data injection.