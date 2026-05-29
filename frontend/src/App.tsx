import { ConfigProvider } from "@core/config/ConfigProvider";
import { Router } from "./router/Router";
import SentryInitializer from "@core/monitoring/SentryInitializer";

function App() {
  return (
    <div className="App">
      <ConfigProvider>
        <SentryInitializer>
          <Router />
        </SentryInitializer>
      </ConfigProvider>
    </div>
  );
}

export default App;
