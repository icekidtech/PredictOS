import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../store/auth";
import api from "../services/api";

export default function AuthCallback() {
  const nav = useNavigate();
  const { login } = useAuth();

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const token = params.get("token");
    if (!token) {
      nav("/login");
      return;
    }
    localStorage.setItem("token", token);
    // Fetch user profile
    api
      .get("/auth/me")
      .then((r) => {
        login(token, r.data.user ?? r.data);
        nav("/dashboard");
      })
      .catch(() => {
        // Fallback: store token and go to dashboard — backend will validate
        login(token, { id: "", username: "", email: "" });
        nav("/dashboard");
      });
  }, [login, nav]);

  return (
    <div className="min-h-[60vh] flex items-center justify-center">
      <div className="text-center">
        <div className="w-8 h-8 border-2 border-white/20 border-t-white rounded-full animate-spin mx-auto" />
        <p className="text-sm text-zinc-400 mt-3">Completing sign-in...</p>
      </div>
    </div>
  );
}
