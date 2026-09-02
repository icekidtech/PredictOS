import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { eventApi } from "../services/api";

type Market = {
  id: string;
  event_name: string;
  category: string;
  description?: string;
  current_yes_price?: number | null;
  current_no_price?: number | null;
};

const CATEGORIES = ["All", "crypto", "technology", "politics", "sports", "weather", "finance"] as const;

export default function Events() {
  const [events, setEvents] = useState<Market[]>([]);
  const [activeCat, setActiveCat] = useState("All");
  const [query, setQuery] = useState("");

  useEffect(() => {
    eventApi
      .list(activeCat !== "All" ? { category: activeCat } : undefined)
      .then((r) => setEvents(r.data.events ?? []))
      .catch(() => {});
  }, [activeCat]);

  const filtered = query
    ? events.filter((m) => m.event_name.toLowerCase().includes(query.toLowerCase()))
    : events;

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <h1 className="text-xl sm:text-2xl font-semibold tracking-tight">Markets</h1>
        <div className="relative w-full sm:w-72">
          <span className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500 text-sm">⌕</span>
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search markets..."
            className="w-full bg-white/[0.06] border border-white/10 rounded-full pl-9 pr-4 py-2 text-sm placeholder:text-zinc-500 focus:outline-none focus:border-white/20"
          />
        </div>
      </div>

      <div className="flex items-center gap-2 overflow-x-auto scrollbar-none pb-1">
        {CATEGORIES.map((cat) => (
          <button
            key={cat}
            onClick={() => setActiveCat(cat)}
            className={`shrink-0 px-4 py-1.5 rounded-full text-xs font-medium border transition-colors ${
              activeCat === cat
                ? "bg-white text-zinc-900 border-white"
                : "bg-white/[0.06] text-zinc-400 border-white/10 hover:text-white"
            }`}
          >
            {cat === "All" ? "All Markets" : cat.charAt(0).toUpperCase() + cat.slice(1)}
          </button>
        ))}
        <span className="ml-auto hidden sm:inline-flex text-xs text-zinc-500 shrink-0">{filtered.length} markets</span>
      </div>

      {filtered.length === 0 ? (
        <div className="text-center py-16">
          <div className="w-12 h-12 rounded-2xl bg-white/5 border border-white/10 flex items-center justify-center mx-auto mb-3 text-zinc-500">◈</div>
          <p className="text-sm text-zinc-400">No markets found</p>
          <p className="text-xs text-zinc-500 mt-1">Markets sync from DreamDEX every 30s.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-5">
          {filtered.map((m) => (
            <MarketCard key={m.id} market={m} />
          ))}
        </div>
      )}
    </div>
  );
}

function MarketCard({ market }: { market: Market }) {
  const yes = market.current_yes_price != null ? Math.round(market.current_yes_price * 100) : null;
  const no = market.current_no_price != null ? Math.round(market.current_no_price * 100) : null;
  return (
    <div className="rounded-2xl bg-white/[0.04] hover:bg-white/[0.07] border border-white/10 p-4 sm:p-5 transition-colors flex flex-col">
      <div className="flex items-start justify-between gap-3">
        <span className="inline-flex px-2 py-1 rounded-full bg-white/10 border border-white/10 text-[10px] tracking-wide uppercase text-zinc-300">
          {market.category}
        </span>
        {yes != null && <span className="text-xs font-mono text-zinc-400">{yes}¢ / {no}¢</span>}
      </div>
      <h3 className="font-medium text-sm leading-snug mt-3 line-clamp-2">{market.event_name}</h3>
      {market.description && <p className="text-xs text-zinc-500 mt-1.5 line-clamp-2 leading-relaxed">{market.description}</p>}
      <div className="mt-4 flex items-center gap-2">
        <span className="flex-1 inline-flex items-center justify-center px-3 py-2 rounded-full bg-emerald-500/15 border border-emerald-500/20 text-emerald-300 text-xs font-medium">
          YES {yes != null ? `${yes}¢` : "—"}
        </span>
        <span className="flex-1 inline-flex items-center justify-center px-3 py-2 rounded-full bg-red-500/10 border border-red-500/20 text-red-300 text-xs font-medium">
          NO {no != null ? `${no}¢` : "—"}
        </span>
      </div>
      <div className="mt-3 flex items-center justify-between text-[11px] text-zinc-500">
        <span>{market.id.slice(0, 8)} · DreamDEX</span>
        <Link to={`/events`} className="hover:text-zinc-300">View →</Link>
      </div>
    </div>
  );
}
