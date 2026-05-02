import { ConfigProvider } from "@core/config/ConfigProvider";
import { Router } from "./router/Router";

function App() {
  return (
    <div className="App">
      <ConfigProvider>
        <Router />
      </ConfigProvider>
    </div>
  );
}

export default App;
