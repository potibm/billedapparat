# ADR 0005: "Urgent" Handling via Virtual Playlists

**Status:** Accepted

## Context

Our digital signage system needs to handle "Urgent" slides (e.g., breaking news or emergency announcements) that must bypass the regular scheduled sequence and be displayed immediately.
Initially, this prioritization was handled imperatively within the core slideshow engine (`engineReducer` and `useSlideshowEngine`). This resulted in scattered conditional checks (`if (hasUrgent) { ... }`) throughout the state management and side effects. It led to high cyclomatic complexity, made the core engine difficult to test due to "in situ" special handling, and provoked bugs—especially when interacting with automated timeouts and manual pauses.

## Decision

We decided to adopt a declarative approach and strip the slideshow engine of all "Urgent" special logic. The engine will become "dumb" again, with a single responsibility: play whatever playlist it is given.

The conditional logic is moved one layer up into the data fetching layer (`useCurrentPlaylist`). This hook will now act as an interceptor:

1. It checks the `slideStore` for active urgent slides.
2. **If urgent slides exist:** It builds a "virtual playlist" object on the fly containing only these urgent slides and returns it.
3. **If NO urgent slides exist:** It returns the regular scheduled playlist.

### 3. The Interceptor Hook (`useCurrentPlaylist`)

`````typescript
if (urgentSlides.length > 0) {
  // Override regular playlist with a virtual urgent playlist
  return {
    id: -1, // Ensures the engine detects a playlist change
    name: "Urgent Override",
    steps: [
      {
        type: "urgent", // Special internal type mapped in the engine
        count: 1,       // Engine handles looping natively
        order: "desc",
        duration: 10,
      }
    ]
  };
}
````
`````
