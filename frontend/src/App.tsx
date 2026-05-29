import { ConfigProvider } from "@core/config/ConfigProvider";
import { Router } from "./router/Router";
import * as Sentry from "@sentry/react";

function App() {
  return (
    <div className="App">
      <ConfigProvider>
        <Sentry.ErrorBoundary
          fallback={
            <p>
              A serious error has occurred. Please restart the beamer
              application.
            </p>
          }
        >
          <Router />
        </Sentry.ErrorBoundary>
      </ConfigProvider>
    </div>
  );
}

export default App;
