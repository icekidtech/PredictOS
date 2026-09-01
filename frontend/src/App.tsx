import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider, useAuth } from "./store/auth";
import Layout from "./components/Layout";
import Landing from "./pages/Landing";
import Dashboard from "./pages/Dashboard";
import Strategies from "./pages/Strategies";
import Backtests from "./pages/Backtests";
import Events from "./pages/Events";
import Settings from "./pages/Settings";
import Login from "./pages/Login";

function Protected({ children }: { children: React.ReactNode }) {
  const { token } = useAuth();
  if (!token) return <Navigate to="/login" replace />;
  return <>{children}</>;
}

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          {/* Public landing */}
          <Route path="/" element={<Landing />} />
          <Route path="/login" element={<Layout><Login /></Layout>} />
          <Route path="/events" element={<Layout><Events /></Layout>} />
          {/* Protected */}
          <Route path="/dashboard" element={<Protected><Layout><Dashboard /></Layout></Protected>} />
          <Route path="/strategies" element={<Protected><Layout><Strategies /></Layout></Protected>} />
          <Route path="/backtests" element={<Protected><Layout><Backtests /></Layout></Protected>} />
          <Route path="/settings" element={<Protected><Layout><Settings /></Layout></Protected>} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
}
