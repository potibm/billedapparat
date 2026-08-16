# ADR 0003: Playlist-Driven Content Organization & System Unification

**Status:** Accepted

## Context
Historically, the setup at the Evoke demoparty consisted of two separate physical screens running two distinct software solutions:
1. **Main Screen:** A slide rotation tool displaying Sponsor screens, News, and the Timetable.
2. **Secondary Screen:** A custom-built Social Media wall.

We were rather dissatisfied with the tooling for the main screen. While planning a rewrite, we realized the requirements for both systems were strikingly similar. At the same time, discussions arose about potentially dropping the secondary screen entirely and showing everything on the main screen, but no clear consensus was reached. We needed a system that could handle both scenarios without requiring code changes.

## Decision
We decided to unify both tools into a single application (Billedapparat) and handle the display topology by introducing declarative "Playlists". 

Instead of hardcoding sequences or building separate apps for different screen setups, we "embraced" the open physical decision. A Playlist is defined in the system's configuration and dictates a sequence of "Steps" based purely on content *types* (e.g., "sponsor", "social.media", "news"). 

This allows us to freely compose what is shown. We can seamlessly replicate the legacy setup (running one client with a "Main" playlist and another client with a "Social" playlist) or merge everything onto a single screen using a combined playlist.

### Configuration Example
Playlists are defined flexibly in the configuration:

```yaml
playlists:
    - id: 1
      name: "Default"
      steps:
        - type: "sponsor"
        - type: "social.media"
        - type: "sponsor"
        - type: "timetable"
          count: 2
          order: "asc"
        - type: "news"
        - type: "sponsor"
        - type: "social.media"
          count: 2
    - id: 2
      name: "Social Media"
      steps:
        - type: "social.media"
          duration: 2
```
        
## Consequences

* **Positive (Absolute Freedom):** The software does not dictate the physical hardware setup. Organizers can change their mind about the screen topology up to the last minute by simply switching the active playlist.
* **Positive (System Consolidation):** Maintaining one unified engine for all screen types heavily reduces development and maintenance overhead compared to the old two-tool setup.
* **Positive (On-the-fly Switching):** Operators can instantly switch between playlists (e.g., via keyboard shortcuts) to adapt to the current event situation (e.g., forcing a "Sponsors Only" or "Timetable" loop during a break).
* **Negative:** It adds a layer of abstraction. Developers and users need to understand the mental model that a Playlist does not contain specific slides, but rather rules for selecting slides at runtime.