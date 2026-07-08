import { BrowserRouter, Routes, Route } from "react-router-dom";
import Layout from "./components/Layout";
import Dashboard from "./pages/Dashboard";
import AthletesPage from "./pages/AthletesPage";
import SensorsPage from "./pages/SensorsPage";
import SessionsPage from "./pages/SessionsPage";
import AnalyticsPage from "./pages/AnalyticsPage";
import EquipmentPage from "./pages/EquipmentPage";
import WodBuilderPage from "./pages/WodBuilderPage";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Dashboard />} />
          <Route path="/athletes" element={<AthletesPage />} />
          <Route path="/sensors" element={<SensorsPage />} />
          <Route path="/sessions" element={<SessionsPage />} />
          <Route path="/analytics/:id" element={<AnalyticsPage />} />
          <Route path="/equipment" element={<EquipmentPage />} />
          <Route path="/wod-builder" element={<WodBuilderPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
