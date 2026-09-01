import { useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { useAuth } from "../store/auth";

const nav = [
  { to: "/dashboard", label: "Dashboard" },
  { to: "/strategies", label: "Strategies" },
  { to: "/backtests", label: "Backtests" },
  { to: "/events", label: "Markets" },
  { to: "/settings", label: "Settings" },
];

export default function Layout({ children }: { children: React.ReactNode }) {
  const { pathname } = useLocation();
  const { user, logout } = useAuth();
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <div className="min-h-screen flex flex-col bg-[#060a14] text-white">
      <header className="sticky top-0 z-20 border-b border-white/5 bg-[#060a14]/80 backdrop-blur-xl">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 h-14 flex items-center justify-between gap-4">
          <div className="flex items-center gap-6">
            <Link to="/" className="flex items-center gap-2 font-bold text-[15px] tracking-tight shrink-0">
              <span className="w-7 h-7 rounded-lg bg-blue-600 flex items-center justify-center text-white text-xs">◈</span>
              Predict<span className="text-blue-400">OS</span>
            </Link>
            {/* Desktop nav */}
            <nav className="hidden md:flex items-center gap-1 p-1 rounded-full bg-white/[0.06] border border-white/10">
              {nav.map((n) => (
                <Link
                  key={n.to}
                  to={n.to}
                  className={`px-3.5 py-1.5 rounded-full text-xs font-medium transition-colors ${
                    pathname === n.to || pathname.startsWith(n.to + "/")
                      ? "bg-white text-zinc-900"
                      : "text-zinc-400 hover:text-white"
                  }`}
                >
                  {n.label}
                </Link>
              ))}
            </nav>
          </div>

          <div className="flex items-center gap-2 sm:gap-3">
            {/* Desktop auth */}
            <div className="hidden sm:flex items-center gap-3">
              {user ? (
                <>
                  <span className="text-xs text-zinc-400 max-w-[140px] truncate">{user.username || user.email}</span>
                  <button onClick={logout} className="text-xs text-zinc-400 hover:text-white px-3 py-1.5 rounded-full border border-white/10">
                    Logout
                  </button>
                </>
              ) : (
                <Link to="/login" className="text-xs font-medium bg-[#3b82f6] hover:bg-[#2563eb] text-white px-4 py-2 rounded-full transition-colors">
                  Connect Wallet
                </Link>
              )}
            </div>
            {/* Mobile menu button */}
            <button
              onClick={() => setMobileOpen(!mobileOpen)}
              className="md:hidden w-9 h-9 rounded-full bg-white/[0.06] border border-white/10 flex items-center justify-center text-zinc-300"
              aria-label="Menu"
            >
              <span className="text-lg leading-none">{mobileOpen ? "✕" : "☰"}</span>
            </button>
          </div>
        </div>

        {/* Mobile nav */}
        {mobileOpen && (
          <div className="md:hidden border-t border-white/5 bg-[#0a0e1a] px-4 py-4 space-y-3">
            <nav className="flex flex-col gap-1">
              {nav.map((n) => (
                <Link
                  key={n.to}
                  to={n.to}
                  onClick={() => setMobileOpen(false)}
                  className={`px-4 py-2.5 rounded-xl text-sm font-medium ${
                    pathname === n.to ? "bg-white text-zinc-900" : "text-zinc-400 bg-white/[0.04] border border-white/5"
                  }`}
                >
                  {n.label}
                </Link>
              ))}
            </nav>
            <div className="pt-3 border-t border-white/5">
              {user ? (
                <div className="flex items-center justify-between">
                  <span className="text-sm text-zinc-400 truncate">{user.username || user.email}</span>
                  <button onClick={logout} className="text-sm text-zinc-400 hover:text-white">Logout</button>
                </div>
              ) : (
                <Link to="/login" onClick={() => setMobileOpen(false)} className="block text-center bg-[#3b82f6] text-white py-3 rounded-full text-sm font-medium">
                  Connect Wallet
                </Link>
              )}
            </div>
          </div>
        )}
      </header>
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 py-6 sm:py-8">{children}</main>
    </div>
  );
}
