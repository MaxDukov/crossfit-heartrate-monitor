import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import Layout from "./components/Layout";
import Dashboard from "./pages/Dashboard";
import AnalyticsPage from "./pages/AnalyticsPage";
import SettingsPage from "./pages/SettingsPage";
import WodPage from "./pages/WodPage";
import CycleDetailPage from "./pages/CycleDetailPage";
import HistoryPage from "./pages/HistoryPage";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Dashboard />} />
          <Route path="/wod" element={<WodPage />} />
          <Route path="/wod/cycles/:cycleId" element={<CycleDetailPage />} />
          <Route path="/history" element={<HistoryPage />} />
          <Route path="/analytics/:id" element={<AnalyticsPage />} />
          <Route path="/settings" element={<SettingsPage />} />
          {/* Легаси-маршруты → новые места */}
          <Route path="/wod-builder" element={<Navigate to="/wod" replace />} />
          <Route path="/athletes" element={<Navigate to="/settings?tab=athletes" replace />} />
          <Route path="/sensors" element={<Navigate to="/settings?tab=sensors" replace />} />
          <Route path="/equipment" element={<Navigate to="/wod" replace />} />
          <Route path="/sessions" element={<Navigate to="/history" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
