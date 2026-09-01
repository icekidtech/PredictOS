import { useEffect, useState } from "react";
import { strategyApi } from "../services/api";

export default function Strategies() {
  const [list, setList] = useState<unknown[]>([]);
  const [name, setName] = useState("");
  const [nl, setNl] = useState("");
  const [loading, setLoading] = useState(false);

  const refresh = () => strategyApi.list().then((r) => setList(r.data.strategies ?? [])).catch(() => {});
  useEffect(() => { refresh(); }, []);

  const create = async () => {
    if (!name) return;
    setLoading(true);
    try {
      await strategyApi.create({ name, natural_language: nl || undefined });
      setName(""); setNl("");
      await refresh();
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "failed";
      alert(msg);
    } finally { setLoading(false); }
  };

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Strategies</h1>

      <div className="bg-zinc-900 rounded-xl border border-zinc-800 p-6 space-y-4">
        <h2 className="font-medium">New Strategy</h2>
        <input
          value={name} onChange={(e) => setName(e.target.value)}
          placeholder="Strategy name (e.g. Tech IPO Bullish)"
          className="w-full bg-zinc-800 border border-zinc-700 rounded-md px-3 py-2 text-sm"
        />
        <textarea
          value={nl} onChange={(e) => setNl(e.target.value)}
          placeholder="Describe in plain English: Buy YES on tech IPO events if confidence >80%..."
          rows={3}
          className="w-full bg-zinc-800 border border-zinc-700 rounded-md px-3 py-2 text-sm"
        />
        <button onClick={create} disabled={loading} className="bg-indigo-600 hover:bg-indigo-500 px-4 py-2 rounded-md text-sm disabled:opacity-50">
          {loading ? "Creating..." : "Create Strategy"}
        </button>
      </div>

      <div className="bg-zinc-900 rounded-xl border border-zinc-800 p-6">
        <h2 className="font-medium mb-4">Your Strategies</h2>
        {list.length === 0 ? <p className="text-sm text-zinc-500">No strategies yet.</p> :
          <div className="space-y-2">
            {(list as Array<Record<string, unknown>>).map((s) => (
              <div key={String(s.id)} className="flex items-center justify-between bg-zinc-800 rounded-lg px-4 py-3">
                <div>
                  <div className="font-medium text-sm">{String(s.name)}</div>
                  <div className="text-xs text-zinc-500">{String(s.status)} · {String(s.id).slice(0, 8)}</div>
                </div>
                <span className="text-xs bg-zinc-700 px-2 py-1 rounded">{String(s.status)}</span>
              </div>
            ))}
          </div>
        }
      </div>
    </div>
  );
}
