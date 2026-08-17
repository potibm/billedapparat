# ADR 0004: Ubiquitous Language (Domain Terminology)

**Status:** Accepted

## Context

Our digital signage ecosystem spans multiple boundaries: independent data fetchers, a central backend, external content management apps, and a smart frontend client. In such a highly decoupled architecture, ambiguous terminology (e.g., using "screen", "page", or "loop" interchangeably) leads to bugs and communication overhead. We needed to establish a "Ubiquitous Language" (a core concept from Domain-Driven Design) to be used consistently across all codebases, variable names, and documentation.

## Decision

We standardize on the following core entities, grouped by their architectural domain:

### 1. Ecosystem Components

- **Hub:** The central backend system. It receives data via REST, deduplicates/filters it, persists it to SQLite, and broadcasts it to clients via Server-Sent Events (SSE).
- **Collector:** An independent, external process (often language-agnostic) that fetches data from third-party services (e.g., Social Media, Redis Streams) and pushes it to the Hub.
- **Beamer:** The smart client frontend application (built with React). It runs on the physical display hardware, rendering the slides and executing the complex playback logic.
- **Admin:** The web-based management UI used for system configuration (managing Sponsors, Playlists) and providing deep-links to external domain applications (like Funkapparat or Tidsapparat).

### 2. Content & Playback Engine

- **Slide:** The smallest visual unit of content. Contains raw data (title, body, type). It does _not_ know when or how it is displayed.
- **Playlist:** A named collection of rules (defined in the configuration) detailing the sequence of content.
- **Step:** A single rule within a Playlist (e.g., "Show 3 slides of type 'news' for 10 seconds each").
- **Engine:** The state machine residing in the Beamer frontend that orchestrates the playback based on Playlists and incoming Slides.

### 3. Content Types & Modifiers

- **Sponsor Slide:** A static slide used not only for commercial sponsors but also for community or scene announcements. Instead of a strict chronological sequence, its appearance frequency is governed by a configurable priority weight, determining how likely it is to be selected during a playlist step.
- **Social Text Slide:** A social media post containing only text/markdown. These are _not_ rendered as standalone slides in the playlist sequence. Instead, they are rendered as a **Toast (or Overlay)** — a non-blocking, floating notification displayed on top of the currently active slide.
- **Social Media Slide:** A social media post containing visual media (images or videos). These are rendered as regular slides within the playlist flow. They do not render the media fullscreen; instead, they display the image/video alongside the user's text message, avatar, and username.
- **Urgent:** A state/flag indicating that a slide contains breaking news or an emergency announcement. It must bypass all regular playlist rules and be shown immediately on the Beamer.

## Consequences

- **Positive:** Code becomes self-documenting. A variable named `currentStep` or a repository named `collector-mastodon` is immediately understood by any developer on the project without further explanation.
- **Positive:** Clarifies the exact rendering expectations and scheduling rules for frontend developers (e.g., Sponsor slides rely on priority weighting, not just simple ordering).
- **Negative:** Requires strict discipline during code reviews to ensure legacy or ambiguous terms are not accidentally introduced.
