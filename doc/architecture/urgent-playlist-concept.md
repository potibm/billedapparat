# Architectural Concept: "Urgent" as a Virtual Playlist

## 🚨 The Problem (Status Quo)
Currently, the prioritization of "Urgent" slides (breaking news) is handled imperatively within the core engine. Throughout the entire state management (`engineReducer`) and the side effects (`useSlideshowEngine`), there are scattered conditional checks like `if (hasUrgent) { ... }`. 
This leads to high cyclomatic complexity, makes the code difficult to test ("in situ" special handling), and provokes bugs, especially when interacting with timeouts and manual pauses.

## 💡 The Solution (Declarative Approach)
The slideshow engine (`useSlideshowEngine` & `engineReducer`) will be completely stripped of the "Urgent" special logic. The engine becomes "dumb" again and no longer needs to know what a breaking news slide is. Its single responsibility is to play whatever playlist it receives.

The conditional logic is moved one layer up: into the data fetching layer. The hook responsible for providing the current playlist (`useCurrentPlaylist`) will act as an interceptor.

### How the Interceptor works:
1. The hook checks the `slideStore` for active urgent slides.
2. **If urgent slides exist:** The hook builds a virtual playlist object "on the fly" containing only these urgent slides and returns it.
3. **If NO urgent slides exist:** The hook returns the regular scheduled playlist.

## 💻 Concept Code (`useCurrentPlaylist.ts`)

\`\`\`typescript
import { useSlideManager } from "./useSlideManager";
// ... other regular playlist imports

export const useCurrentPlaylist = () => {
  const { getUrgent } = useSlideManager();
  const urgentSlides = getUrgent();
  
  // The regular logic to determine the active scheduled playlist
  const regularPlaylist = useRegularPlaylistLogic(); 

  // --- The Interceptor ---
  if (urgentSlides.length > 0) {
    return {
      id: "virtual-urgent-playlist",
      name: "Urgent Override",
      steps: [
        {
          type: "news", // The type corresponding to urgent slides
          count: urgentSlides.length,
          order: "priority",
          duration: 10,
        }
      ]
    };
  }

  return regularPlaylist;
};
\`\`\`

## 🚀 Architectural Advantages
* **Massive Code Reduction in the Engine:** All `if (hasUrgent)` blocks are removed from the reducer (`NEXT` action) and the React hooks.
* **No Complex Payload Dependencies:** The engine no longer needs to pass `urgentSlides` into the reducer payload.
* **Automatic Fallback (Self-Healing):** When the last urgent slide is deactivated, `useCurrentPlaylist` automatically switches back to the regular playlist. Because this changes the `playlist.id`, the existing `useEffect` in the engine naturally triggers a clean `RESET_PLAYLIST` and resumes normal playback without requiring manual interventions in the timers.
* **Better Testability:** The logic is cleanly separated. We can test the playlist hook in isolation (does it return the virtual playlist correctly?) and the engine in isolation (does it play the given playlist correctly?).

## 🛠️ Next Refactoring Steps (Action Items)
1. Rewrite `useCurrentPlaylist` to implement the interceptor for the virtual playlist.
2. Clean up `engineReducer.ts`: Remove `hasUrgent` and `urgentSlides` from the payload. Delete the special urgent handling inside the `NEXT` action.
3. Clean up `useSlideshowEngine.ts`: Remove all derived urgent states and timeouts that specifically react to the urgent status.
