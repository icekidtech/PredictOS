import { useEffect, useState } from "react";
import { backtestApi, strategyApi } from "../services/api";

export default function Backtests() {
  const [strategies, setStrategies] = useState<Array<{ id: string; name: string }>>([]);
  const [selected, setSelected] = useState("");
  const [result, setResult] = useState<Record<string, unknown> | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    strategyApi.list().then((r) => setStrategies(r.data.strategies ?? [])).catch(() => {});
  }, []);

  const run = async () => {
    if (!selected) return;
    setLoading(true);
    try {
      const r = await backtestApi.create({ strategy_id: selected, initial_capital: 10000 });
      setResult(r.data);
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : "failed");
    } finally { setLoading(false); }
  };

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Backtests</h1>
      <div className="bg-zinc-900 rounded-xl border border-zinc-800 p-6 space-y-4">
        <select value={selected} onChange={(e) => setSelected(e.target.value)} className="w-full bg-zinc-800 border border-zinc-700 rounded-md px-3 py-2 text-sm">
          <option value="">Select strategy</option>
          {strategies.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
        </select>
        <button onClick={run} disabled={!selected || loading} className="bg-indigo-600 hover:bg-indigo-500 px-4 py-2 rounded-md text-sm disabled:opacity-50">
          {loading ? "Running..." : "Run Backtest"}
        </button>
      </div>
      {result && (
        <div className="bg-zinc-900 rounded-xl border border-zinc-800 p-6">
          <h2 className="font-medium mb-4">Result</h2>
          <pre className="text-xs text-zinc-400 overflow-auto">{JSON.stringify(result, null, 2)}</pre>
        </div>
      )}
    </div>
  );
}
