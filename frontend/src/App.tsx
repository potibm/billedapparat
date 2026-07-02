import * as Sentry from "@sentry/react";
import Router from "./router/Router";

function App() {
  return (
    <div className="App">
      <Sentry.ErrorBoundary
        fallback={
          <p>
            A serious error has occurred. Please restart the beamer application.
          </p>
        }
      >
        <Router />
      </Sentry.ErrorBoundary>
    </div>
  );
}

export default App;
