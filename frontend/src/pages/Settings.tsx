import { useEffect, useState } from "react";
import { settingsApi } from "../services/api";

export default function Settings() {
  const [provider, setProvider] = useState("openai");
  const [model, setModel] = useState("gpt-4o-mini");
  const [apiKey, setApiKey] = useState("");
  const [network, setNetwork] = useState("testnet");
  const [hasKey, setHasKey] = useState(false);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    settingsApi.get().then((r) => {
      setProvider(r.data.ai_provider ?? "openai");
      setModel(r.data.ai_model ?? "gpt-4o-mini");
      setHasKey(!!r.data.has_api_key);
      const nw = r.data.network ?? "testnet";
      setNetwork(nw);
      localStorage.setItem("network", nw);
    }).catch(() => {});
  }, []);

  const save = async () => {
    setSaving(true);
    try {
      await settingsApi.update({
        ai_provider: provider,
        ai_model: model,
        api_key: apiKey || undefined,
        network,
      });
      localStorage.setItem("network", network);
      setApiKey("");
      alert("Settings saved");
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : "failed");
    } finally { setSaving(false); }
  };

  return (
    <div className="space-y-6 max-w-2xl">
      <h1 className="text-2xl font-semibold">Settings</h1>

      <div className="bg-zinc-900 rounded-xl border border-zinc-800 p-6 space-y-4">
        <h2 className="font-medium">AI Provider</h2>
        <div className="grid grid-cols-2 gap-4">
          <label className="space-y-1">
            <span className="text-xs text-zinc-500">Provider</span>
            <select value={provider} onChange={(e) => setProvider(e.target.value)} className="w-full bg-zinc-800 border border-zinc-700 rounded-md px-3 py-2 text-sm">
              <option value="openai">OpenAI</option>
              <option value="anthropic">Anthropic</option>
            </select>
          </label>
          <label className="space-y-1">
            <span className="text-xs text-zinc-500">Model</span>
            <select value={model} onChange={(e) => setModel(e.target.value)} className="w-full bg-zinc-800 border border-zinc-700 rounded-md px-3 py-2 text-sm">
              {provider === "openai" ? (
                <>
                  <option value="gpt-4o-mini">gpt-4o-mini</option>
                  <option value="gpt-4o">gpt-4o</option>
                  <option value="gpt-4-turbo">gpt-4-turbo</option>
                </>
              ) : (
                <>
                  <option value="claude-3-5-sonnet-20241022">claude-3-5-sonnet</option>
                  <option value="claude-3-haiku-20240307">claude-3-haiku</option>
                </>
              )}
            </select>
          </label>
        </div>
        <label className="space-y-1 block">
          <span className="text-xs text-zinc-500">API Key {hasKey && "(already set — leave blank to keep)"}</span>
          <input value={apiKey} onChange={(e) => setApiKey(e.target.value)} type="password" placeholder="sk-..." className="w-full bg-zinc-800 border border-zinc-700 rounded-md px-3 py-2 text-sm" />
        </label>
      </div>

      <div className="bg-zinc-900 rounded-xl border border-zinc-800 p-6 space-y-4">
        <h2 className="font-medium">Network</h2>
        <select value={network} onChange={(e) => setNetwork(e.target.value)} className="w-full bg-zinc-800 border border-zinc-700 rounded-md px-3 py-2 text-sm">
          <option value="testnet">Somnia Testnet</option>
          <option value="mainnet">Somnia Mainnet</option>
        </select>
        <p className="text-xs text-zinc-500">Toggling switches the RPC used for DreamDEX data and trading.</p>
      </div>

      <button onClick={save} disabled={saving} className="bg-indigo-600 hover:bg-indigo-500 px-6 py-2 rounded-md text-sm disabled:opacity-50">
        {saving ? "Saving..." : "Save Settings"}
      </button>
    </div>
  );
}
