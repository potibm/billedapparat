# ADR 0005: "Urgent" Handling via Virtual Playlists

**Status:** Accepted

## Context

Our digital signage system needs to handle "Urgent" slides (e.g., breaking news or emergency announcements) that must bypass the regular scheduled sequence and be displayed immediately.
Initially, this prioritization was handled imperatively within the core slideshow engine (`engineReducer` and `useSlideshowEngine`). This resulted in scattered conditional checks (`if (hasUrgent) { ... }`) throughout the state management and side effects. It led to high cyclomatic complexity, made the core engine difficult to test due to "in situ" special handling, and provoked bugs—especially when interacting with automated timeouts and manual pauses.

## Decision

We decided to adopt a declarative approach and move as much of the "Urgent" special logic as possible out of the slideshow engine and into the data-fetching layer (`useCurrentPlaylist`). The engine retains a single, narrowly scoped data-filtering exception so that urgent and non-urgent news slides can be cleanly separated. See _The Engine Pseudo-Type Exception_ below for the rationale.

The conditional logic is moved one layer up into the data fetching layer (`useCurrentPlaylist`). This hook will now act as an interceptor:

1. It checks the `slideStore` for active urgent slides.
2. **If urgent slides exist:** It builds a "virtual playlist" object on the fly containing only these urgent slides and returns it.
3. **If NO urgent slides exist:** It returns the regular scheduled playlist.

### The Interceptor Hook (`useCurrentPlaylist`)

```typescript
if (urgentSlides.length > 0) {
  // Override regular playlist with a virtual urgent playlist
  return {
    id: -1, // Ensures the engine detects a playlist change
    name: "Urgent Override",
    steps: [
      {
        type: "urgent", // Special internal type mapped in the engine (see below)
        count: 1, // Step advances through the urgent pool one slide per TICK
        order: "desc",
        duration: 10,
      },
    ],
  };
}
```

### The Engine Pseudo-Type Exception (Architectural Compromise)

The initial goal was to make the engine 100% "dumb" and unaware of priorities. However, mapping the override playlist to `type: "news"` caused urgent slides to leak into regular news rotation steps, and vice versa.

To solve this, the engine remains agnostic to _playback_ interruptions, but contains exactly one data-filtering exception in its pure reducer:
When asked for the pseudo-type `"urgent"`, it ignores `content.type` and instead resolves all active slides where `display_options.is_urgent === true`. Sorting by priority is handled internally within this specific filter branch, satisfying the rigid `PlaylistStepSchema` for the order property `("asc" | "desc" | "random")`.

### Considered Alternatives

We briefly considered expressing the urgent override purely as data by extending `PlaylistStep` with an optional `display_options_filter` field, for example:

```ts
{
  type: "news",
  display_options_filter: { is_urgent: true },
  count: 1,
  order: "desc",
  duration: 10,
}
```

Rejected because:

- It would spread urgent-handling knowledge across the runtime schema, the user-facing config files (`config.yaml`), and the admin UI.
- It would force every consumer of `PlaylistStep` (including the regular news rotation) to grow a second filter dimension, even though most steps only care about `content.type`.
- The chosen approach keeps "urgency" as exactly one well-tested special branch in the engine, which is consistent with the spirit of the ADR (declarative selection, well-isolated imperative resolution).
