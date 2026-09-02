import { useState, useEffect, useRef } from "react";
import { Link, useLocation } from "react-router-dom";
import { useAccount, useSignMessage, useDisconnect } from "wagmi";
import { ConnectButton } from "@rainbow-me/rainbowkit";
import { useAuth } from "../store/auth";
import { authApi, settingsApi } from "../services/api";
import { useNetwork } from "../hooks/useNetwork";

const nav = [
  { to: "/dashboard", label: "Dashboard" },
  { to: "/strategies", label: "Strategies" },
  { to: "/backtests", label: "Backtests" },
  { to: "/events", label: "Markets" },
  { to: "/settings", label: "Settings" },
];

export default function Layout({ children }: { children: React.ReactNode }) {
  const { pathname } = useLocation();
  const { user, logout, refreshUser } = useAuth();
  const { network, setNetwork } = useNetwork();
  const { address, isConnected } = useAccount();
  const { signMessageAsync } = useSignMessage();
  const { disconnect } = useDisconnect();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [profileOpen, setProfileOpen] = useState(false);
  const [linking, setLinking] = useState(false);
  const profileRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (profileRef.current && !profileRef.current.contains(e.target as Node)) setProfileOpen(false);
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  const handleNetworkSwitch = async (nw: string) => {
    setNetwork(nw);
    try { await settingsApi.update({ network: nw }); } catch { /* local only if not authed */ }
  };

  const handleLinkWallet = async () => {
    if (!address) return;
    setLinking(true);
    try {
      const { data } = await authApi.nonce(address);
      const sig = await signMessageAsync({ message: data.message });
      await authApi.walletLink({ address, message: data.message, signature: sig });
      await refreshUser();
    } catch { /* handled by UI */ } finally { setLinking(false); }
  };

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
            {/* Desktop auth — profile dropdown */}
            <div className="hidden sm:flex items-center gap-2">
              {user ? (
                <div className="relative" ref={profileRef}>
                  <button onClick={() => setProfileOpen(!profileOpen)} className="flex items-center gap-2 pl-1 pr-3 py-1 rounded-full bg-white/[0.06] border border-white/10 hover:bg-white/[0.10] transition-colors">
                    {user.avatar_url ? <img src={user.avatar_url} alt="" className="w-6 h-6 rounded-full" /> : <span className="w-6 h-6 rounded-full bg-blue-600 flex items-center justify-center text-white text-[10px]">{(user.username || user.email || "U")[0].toUpperCase()}</span>}
                    <span className="text-xs text-white max-w-[120px] truncate">{user.username || user.email}</span>
                    <span className="text-zinc-500 text-xs">▾</span>
                  </button>
                  {profileOpen && (
                    <div className="absolute right-0 mt-2 w-72 rounded-2xl bg-[#0f1420] border border-white/10 shadow-xl overflow-hidden z-50">
                      <div className="p-4 border-b border-white/5">
                        <div className="text-sm font-medium truncate">{user.username || "User"}</div>
                        <div className="text-xs text-zinc-500 truncate">{user.email}</div>
                        {user.wallet_address ? (
                          <div className="mt-2 flex items-center gap-2 text-xs font-mono bg-white/[0.06] border border-white/10 rounded-full px-3 py-1.5">
                            <span className="w-2 h-2 rounded-full bg-emerald-400" />
                            <span className="truncate">{user.wallet_address.slice(0, 6)}...{user.wallet_address.slice(-4)}</span>
                            <button onClick={() => { navigator.clipboard.writeText(user.wallet_address!); }} className="ml-auto text-zinc-400 hover:text-white">⧉</button>
                          </div>
                        ) : (
                          <div className="mt-3 space-y-2">
                            <div className="text-xs text-amber-300">No wallet linked</div>
                            <ConnectButton />
                            {isConnected && address && (
                              <button onClick={handleLinkWallet} disabled={linking} className="w-full text-xs bg-blue-600 hover:bg-blue-500 text-white py-2 rounded-full disabled:opacity-50">
                                {linking ? "Linking..." : "Sign & Link Wallet"}
                              </button>
                            )}
                          </div>
                        )}
                      </div>
                      <div className="p-3 space-y-2">
                        <div className="text-[11px] tracking-widest uppercase text-zinc-500">Network</div>
                        <div className="flex gap-2">
                          {(["testnet", "mainnet"] as const).map((nw) => (
                            <button key={nw} onClick={() => handleNetworkSwitch(nw)} className={`flex-1 px-3 py-2 rounded-full text-xs font-medium border transition-colors ${network === nw ? "bg-white text-zinc-900 border-white" : "bg-white/[0.06] text-zinc-400 border-white/10 hover:text-white"}`}>
                              {nw === "testnet" ? "Testnet" : "Mainnet"}
                            </button>
                          ))}
                        </div>
                      </div>
                      <div className="p-3 border-t border-white/5 flex gap-2">
                        <Link to="/settings" onClick={() => setProfileOpen(false)} className="flex-1 text-center text-xs bg-white/[0.06] border border-white/10 rounded-full py-2 hover:bg-white/[0.10]">Settings</Link>
                        <button onClick={() => { disconnect(); logout(); setProfileOpen(false); }} className="flex-1 text-xs bg-white text-zinc-900 rounded-full py-2 font-medium">Logout</button>
                      </div>
                    </div>
                  )}
                </div>
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
