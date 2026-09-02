import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { eventApi } from "../services/api";
import { useNetwork } from "../hooks/useNetwork";

type Market = {
  id: string;
  event_name: string;
  category: string;
  description?: string;
  current_yes_price?: number;
  current_no_price?: number;
  event_date?: string;
};

const CATEGORIES = ["All", "technology", "politics", "sports", "weather", "finance"] as const;

export default function Landing() {
  const { network } = useNetwork();
  const [markets, setMarkets] = useState<Market[]>([]);
  const [activeCat, setActiveCat] = useState("All");
  const [query, setQuery] = useState("");

  useEffect(() => {
    eventApi
      .list(activeCat !== "All" ? { category: activeCat } : undefined)
      .then((r) => setMarkets(r.data.events ?? []))
      .catch(() => {});
  }, [activeCat, network]);

  const filtered = query
    ? markets.filter((m) => m.event_name.toLowerCase().includes(query.toLowerCase()))
    : markets;

  return (
    <div className="min-h-screen bg-[#060a14] text-white">
      {/* Top nav — landing has its own minimal nav (not Layout) */}
      <header className="sticky top-0 z-20 border-b border-white/5 bg-[#060a14]/80 backdrop-blur-xl">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 h-14 flex items-center justify-between gap-4">
          <Link to="/" className="flex items-center gap-2 font-bold text-[15px] tracking-tight shrink-0">
            <span className="w-7 h-7 rounded-lg bg-blue-600 flex items-center justify-center text-white text-xs">◈</span>
            Predict<span className="text-blue-400">OS</span>
          </Link>

          {/* Search — desktop */}
          <div className="hidden sm:flex flex-1 max-w-md mx-4">
            <div className="relative w-full">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500 text-sm">⌕</span>
              <input
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Search markets..."
                className="w-full bg-white/[0.06] border border-white/10 rounded-full pl-9 pr-4 py-2 text-sm placeholder:text-zinc-500 focus:outline-none focus:border-white/20"
              />
            </div>
          </div>

          <div className="flex items-center gap-2">
            <span className={`hidden sm:inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[11px] font-medium border ${network === "mainnet" ? "bg-emerald-500/15 border-emerald-500/20 text-emerald-300" : "bg-amber-500/15 border-amber-500/20 text-amber-300"}`}>
              <span className={`w-1.5 h-1.5 rounded-full ${network === "mainnet" ? "bg-emerald-400" : "bg-amber-400"}`} />
              {network === "mainnet" ? "Mainnet" : "Testnet"}
            </span>
            <Link to="/dashboard" className="hidden sm:inline-flex text-xs text-zinc-400 hover:text-white px-3 py-2">
              Dashboard
            </Link>
            <Link to="/login" className="inline-flex items-center justify-center px-4 sm:px-5 py-2 rounded-full bg-[#3b82f6] hover:bg-[#2563eb] text-white text-xs sm:text-sm font-medium transition-colors">
              Connect Wallet
            </Link>
          </div>
        </div>

        {/* Mobile search */}
        <div className="sm:hidden px-4 pb-3">
          <div className="relative">
            <span className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500 text-sm">⌕</span>
            <input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search markets..."
              className="w-full bg-white/[0.06] border border-white/10 rounded-full pl-9 pr-4 py-2.5 text-sm placeholder:text-zinc-500 focus:outline-none focus:border-white/20"
            />
          </div>
        </div>
      </header>

      {/* Compact hero — not full-screen, market-first like Polymarket */}
      <section className="relative overflow-hidden border-b border-white/5">
        <div className="absolute inset-0 bg-gradient-to-br from-blue-600/10 via-transparent to-transparent pointer-events-none" />
        <div className="absolute -top-20 -right-20 w-[500px] h-[300px] bg-blue-500/10 rounded-full blur-[80px] pointer-events-none" />
        <div className="max-w-7xl mx-auto px-4 sm:px-6 py-8 sm:py-10 relative">
          <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-6">
            <div className="max-w-2xl">
              <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-white/[0.06] border border-white/10 text-[11px] tracking-wide text-zinc-300 mb-3">
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
                Autonomous Trading on Somnia × DreamDEX
              </div>
              <h1 className="text-[28px] sm:text-4xl font-semibold tracking-tight leading-tight">
                Predict markets.
                <br />
                <span className="text-zinc-400">Automate the trade.</span>
              </h1>
              <p className="mt-3 text-sm sm:text-[15px] text-zinc-400 leading-relaxed max-w-xl">
                Describe your strategy in plain English, backtest it against history, and deploy autonomous agents that trade 24/7 on zero-fee DreamDEX.
              </p>
              <div className="mt-5 flex flex-col sm:flex-row gap-3">
                <Link to="/strategies" className="inline-flex items-center justify-center px-6 py-3 rounded-full bg-[#3b82f6] hover:bg-[#2563eb] text-white text-sm font-medium transition-colors">
                  Build a Strategy
                </Link>
                <Link to="/events" className="inline-flex items-center justify-center px-6 py-3 rounded-full bg-white/[0.06] hover:bg-white/[0.10] border border-white/10 text-white text-sm font-medium transition-colors">
                  Explore Markets
                </Link>
              </div>
            </div>

            {/* Stats strip — compact, inline */}
            <div className="grid grid-cols-3 gap-3 sm:gap-4 lg:min-w-[380px]">
              <Stat label="Total Volume" value="$2.4M" />
              <Stat label="Active Traders" value="47,736" />
              <Stat label="Weekly Vol" value="$850K" />
            </div>
          </div>
        </div>
      </section>

      {/* Category pills — Polymarket-style */}
      <div className="sticky top-14 z-10 bg-[#060a14]/80 backdrop-blur-xl border-b border-white/5">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 py-3 flex items-center gap-2 overflow-x-auto scrollbar-none">
          {CATEGORIES.map((cat) => (
            <button
              key={cat}
              onClick={() => setActiveCat(cat)}
              className={`shrink-0 px-4 py-1.5 rounded-full text-xs font-medium border transition-colors ${
                activeCat === cat
                  ? "bg-white text-zinc-900 border-white"
                  : "bg-white/[0.06] text-zinc-400 border-white/10 hover:text-white hover:border-white/20"
              }`}
            >
              {cat === "All" ? "All Markets" : cat.charAt(0).toUpperCase() + cat.slice(1)}
            </button>
          ))}
          <span className="ml-auto hidden sm:inline-flex text-xs text-zinc-500 shrink-0">
            {filtered.length} markets
          </span>
        </div>
      </div>

      {/* Markets grid — Polymarket-style cards */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 py-6 sm:py-8">
        {filtered.length === 0 ? (
          <div className="text-center py-16 sm:py-20">
            <div className="w-12 h-12 rounded-2xl bg-white/5 border border-white/10 flex items-center justify-center mx-auto mb-3 text-zinc-500">◈</div>
            <p className="text-sm text-zinc-400">No markets found</p>
            <p className="text-xs text-zinc-500 mt-1">Try a different category or check back soon. Markets come from Somnia/DreamDEX.</p>
            <Link to="/strategies" className="inline-flex mt-4 px-5 py-2 rounded-full bg-white text-zinc-900 text-sm font-medium">
              Create a Strategy
            </Link>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-5">
            {filtered.map((m) => (
              <MarketCard key={m.id} market={m} />
            ))}
          </div>
        )}
      </section>

      {/* How it works — compact 3-step */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 pb-8 sm:pb-12">
        <div className="rounded-[24px] bg-white/[0.04] border border-white/10 p-6 sm:p-8">
          <h2 className="text-lg sm:text-xl font-semibold tracking-tight">How PredictOS works</h2>
          <p className="text-sm text-zinc-400 mt-1">From idea to autonomous execution in three steps.</p>
          <div className="mt-6 grid grid-cols-1 sm:grid-cols-3 gap-4 sm:gap-6">
            {[
              { step: "01", title: "Describe", desc: "Write your strategy in plain English. AI converts it to executable logic." },
              { step: "02", title: "Backtest", desc: "Replay against historical events. See win rate, Sharpe, and P&L before risking capital." },
              { step: "03", title: "Deploy", desc: "Launch an autonomous agent that monitors DreamDEX and trades 24/7." },
            ].map((s) => (
              <div key={s.step} className="rounded-2xl bg-white/[0.04] border border-white/5 p-5">
                <div className="text-xs font-mono text-blue-400">{s.step}</div>
                <h3 className="font-medium text-sm mt-1">{s.title}</h3>
                <p className="text-xs sm:text-sm text-zinc-400 mt-2 leading-relaxed">{s.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-white/5">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 py-6 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-zinc-500">
          <span>© 2026 PredictOS — Built on Somnia × DreamDEX</span>
          <span>Zero-fee • Self-custodial • Onchain • Testnet & Mainnet</span>
        </div>
      </footer>
    </div>
  );
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl bg-white/[0.04] border border-white/10 p-3 sm:p-4">
      <div className="text-[10px] sm:text-xs tracking-widest uppercase text-zinc-500">{label}</div>
      <div className="text-sm sm:text-lg font-semibold mt-1 tracking-tight">{value}</div>
    </div>
  );
}

function MarketCard({ market }: { market: Market }) {
  const yes = market.current_yes_price != null ? Math.round(market.current_yes_price * 100) : null;
  const no = market.current_no_price != null ? Math.round(market.current_no_price * 100) : null;

  return (
    <Link
      to={`/events`}
      className="group rounded-2xl bg-white/[0.04] hover:bg-white/[0.07] border border-white/10 hover:border-white/15 p-4 sm:p-5 transition-colors flex flex-col"
    >
      <div className="flex items-start justify-between gap-3">
        <span className="inline-flex px-2 py-1 rounded-full bg-white/10 border border-white/10 text-[10px] tracking-wide uppercase text-zinc-300">
          {market.category}
        </span>
        {yes != null && (
          <span className="text-xs font-mono text-zinc-400">
            {yes}¢ / {no}¢
          </span>
        )}
      </div>
      <h3 className="font-medium text-sm leading-snug mt-3 line-clamp-2 group-hover:text-white">
        {market.event_name}
      </h3>
      {market.description && (
        <p className="text-xs text-zinc-500 mt-1.5 line-clamp-2 leading-relaxed">{market.description}</p>
      )}
      <div className="mt-4 flex items-center gap-2">
        <span className="flex-1 inline-flex items-center justify-center px-3 py-2 rounded-full bg-emerald-500/15 border border-emerald-500/20 text-emerald-300 text-xs font-medium">
          YES {yes != null ? `${yes}¢` : "—"}
        </span>
        <span className="flex-1 inline-flex items-center justify-center px-3 py-2 rounded-full bg-red-500/10 border border-red-500/20 text-red-300 text-xs font-medium">
          NO {no != null ? `${no}¢` : "—"}
        </span>
      </div>
      <div className="mt-3 flex items-center justify-between text-[11px] text-zinc-500">
        <span>Trade on DreamDEX</span>
        <span className="group-hover:text-zinc-300">View →</span>
      </div>
    </Link>
  );
}
