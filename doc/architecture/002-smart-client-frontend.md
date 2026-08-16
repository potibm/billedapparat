# ADR 0002: Smart Client Architecture & DOM-based Rendering

**Status:** Accepted

## Context

A digital signage/slideshow system requires precise timing, smooth animations, and the ability to react to sudden high-priority events (e.g., breaking news). We had to decide on the core technology stack for the frontend, how playback logic is distributed between client and server, and how the slides are visually rendered. Past systems often relied on the backend generating static graphics/images for each slide, which presented challenges in terms of payload size and adaptability.

## Decision

We decided on a "Smart Client" architecture built with React, utilizing dynamic HTML/CSS rendering for the visual output.

- **React as the Core Framework:** We chose React to build the frontend engine. This decision was primarily driven by existing team experience and expertise, allowing for rapid development, as well as React's robust state management capabilities which are ideal for a complex state machine like a slideshow engine.
- **Frontend-Heavy Logic (Smart Client):** The entire Slideshow Engine, playlist resolution, and timing logic reside strictly in the React frontend. The backend only provides raw data payloads (via SSE) and abstract playlist definitions. The frontend evaluates rules, computes next steps, and orchestrates animations autonomously.
- **DOM-based Rendering (HTML/CSS):** Instead of serving pre-rendered static graphics, the backend sends lightweight JSON payloads. The React frontend renders these payloads natively into HTML DOM elements for the different slide types (e.g., news, timetable, sponsor).
- **Event-Specific Theming:** To allow for high visual customization (e.g., for different events like the Evoke demoparty), the system supports loading a custom "user stylesheet".
- **BEM Methodology:** To make this theming robust and predictable for designers, the core components strictly use the BEM (Block Element Modifier) CSS naming convention (e.g., `slide-news`, `slide-news__title`, `slide-news__content`).

## Consequences

- **Positive (Performance & Bandwidth):** Transmitting raw JSON and rendering DOM elements is significantly more lightweight and faster than generating and downloading large image files over the network.
- **Positive (Customization):** The combination of DOM rendering, BEM classes, and user stylesheets provides ultimate flexibility. Event organizers can completely rebrand the Beamer output without altering the core React application code.
- **Positive (Resilience):** Playback is highly responsive and immune to minor network latency since the client orchestrates the timeline independently.
- **Negative (Hardware Dependency):** Rendering complex DOM structures and CSS animations requires slightly more processing power on the client device (e.g., a Raspberry Pi) compared to merely displaying static images.
- **Negative (Theming Curve):** Designers creating custom event themes must understand and adhere to the established BEM class structures to style the components successfully.
