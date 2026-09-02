import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useAccount, useSignMessage, useBalance, useDisconnect } from "wagmi";
import { ConnectButton } from "@rainbow-me/rainbowkit";
import { portfolioApi, authApi } from "../services/api";
import { useAuth } from "../store/auth";
import { somniaTestnet, somniaMainnet } from "../config/wagmi";
import { useNetwork } from "../hooks/useNetwork";

export default function Dashboard() {
  const { user } = useAuth();
  const { network } = useNetwork();
  const { address, isConnected } = useAccount();
  const { signMessageAsync } = useSignMessage();
  const { disconnect } = useDisconnect();
  const { data: walletBalance } = useBalance({
    address: (user?.wallet_address as `0x${string}`) || address,
    chainId: network === "mainnet" ? somniaMainnet.id : somniaTestnet.id,
    query: { enabled: !!(user?.wallet_address || address) },
  });
  const [summary, setSummary] = useState<Record<string, unknown> | null>(null);
  const [positions, setPositions] = useState<unknown[]>([]);
  const [linking, setLinking] = useState(false);
  const [linkError, setLinkError] = useState("");

  const hasWallet = !!user?.wallet_address;
  const needsWallet = !hasWallet;

  useEffect(() => {
    portfolioApi.summary().then((r) => setSummary(r.data)).catch(() => {});
    portfolioApi.positions().then((r) => setPositions(r.data.positions ?? [])).catch(() => {});
  }, []);

  const { refreshUser } = useAuth();

  const handleLinkWallet = async () => {
    if (!address) return;
    setLinking(true);
    setLinkError("");
    try {
      const { data } = await authApi.nonce(address);
      const signature = await signMessageAsync({ message: data.message });
      await authApi.walletLink({ address, message: data.message, signature });
      await refreshUser();
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error || (e instanceof Error ? e.message : "Failed to link wallet");
      setLinkError(msg);
    } finally {
      setLinking(false);
    }
  };

  return (
    <div className="space-y-6 sm:space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl sm:text-2xl font-semibold tracking-tight">Dashboard</h1>
          <p className="text-sm text-zinc-400 mt-1">Track your autonomous trading performance</p>
        </div>
        <Link to="/strategies" className="inline-flex items-center justify-center px-5 py-2.5 rounded-full bg-[#3b82f6] hover:bg-[#2563eb] text-white text-sm font-medium transition-colors w-full sm:w-auto">
          New Strategy
        </Link>
      </div>

      {/* Wallet banner — Google users without wallet */}
      {needsWallet ? (
        <div className="rounded-2xl bg-amber-500/10 border border-amber-500/20 p-4 sm:p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div className="flex gap-3">
            <span className="w-8 h-8 rounded-full bg-amber-500/20 flex items-center justify-center text-amber-300 shrink-0">◈</span>
            <div>
              <div className="text-sm font-medium text-amber-200">Connect wallet to enable live trading</div>
              <div className="text-xs text-amber-200/70 mt-1">You signed in with Google. Link a wallet to deploy agents on DreamDEX.</div>
            </div>
          </div>
          <div className="flex flex-col gap-2 sm:items-end">
            <ConnectButton />
            {isConnected && address && (
              <button onClick={handleLinkWallet} disabled={linking} className="text-xs bg-amber-500 hover:bg-amber-400 text-zinc-900 px-4 py-2 rounded-full font-medium disabled:opacity-50">
                {linking ? "Linking..." : "Sign & Link Wallet"}
              </button>
            )}
            {linkError && <span className="text-xs text-red-400">{linkError}</span>}
          </div>
        </div>
      ) : (
        <div className="rounded-2xl bg-white/[0.04] border border-white/10 p-4 sm:p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <span className="w-8 h-8 rounded-full bg-emerald-500/20 flex items-center justify-center text-emerald-300 shrink-0">◈</span>
            <div>
              <div className="text-sm font-medium">Wallet Connected</div>
              <div className="text-xs font-mono text-zinc-400">{user.wallet_address?.slice(0, 6)}...{user.wallet_address?.slice(-4)} · {walletBalance ? `${Number(walletBalance.formatted).toFixed(4)} ${walletBalance.symbol}` : "—"} on {network}</div>
            </div>
          </div>
          <button onClick={() => disconnect()} className="text-xs bg-white/[0.06] border border-white/10 rounded-full px-4 py-2 hover:bg-white/[0.10]">Disconnect</button>
        </div>
      )}

      {/* Stats — real data, no mock */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
        <StatCard label="Total Value" value={summary ? `$${Number(summary.total_value).toFixed(2)}` : "—"} sub={summary && Number(summary.positions) === 0 ? "No trades yet" : "Portfolio value"} />
        <StatCard label="Unrealized P&L" value={summary ? `$${Number(summary.unrealized_pnl).toFixed(2)}` : "—"} sub={summary ? `${(Number(summary.unrealized_pnl_percent) * 100).toFixed(2)}%` : "—"} accent />
        <StatCard label="Cash Available" value={summary ? `$${Number(summary.cash_available).toFixed(2)}` : "—"} sub={needsWallet ? "Connect wallet for live" : "Ready to deploy"} />
        <StatCard label="Positions" value={summary ? String(summary.positions) : "—"} sub="Open positions" />
      </div>

      {/* Positions */}
      <div className="rounded-2xl bg-white/[0.04] border border-white/10 backdrop-blur p-4 sm:p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="font-medium text-sm sm:text-base">Open Positions</h2>
          <span className="text-xs px-2.5 py-1 rounded-full bg-white/10 border border-white/10 text-zinc-300">
            {positions.length} open
          </span>
        </div>
        {positions.length === 0 ? (
          <div className="text-center py-10 sm:py-12">
            <div className="w-12 h-12 rounded-2xl bg-white/5 border border-white/10 flex items-center justify-center mx-auto mb-3 text-zinc-500">◈</div>
            <p className="text-sm text-zinc-400">No open positions</p>
            <p className="text-xs text-zinc-500 mt-1 max-w-sm mx-auto">Deploy a strategy to start autonomous trading on DreamDEX.</p>
            <Link to="/strategies" className="inline-flex mt-4 px-5 py-2 rounded-full bg-white text-zinc-900 text-sm font-medium">
              Create Strategy
            </Link>
          </div>
        ) : (
          <div className="space-y-2">
            {(positions as Array<Record<string, unknown>>).map((p) => (
              <div key={String(p.id)} className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 p-3 sm:p-4 rounded-xl bg-white/[0.04] border border-white/5">
                <div className="min-w-0">
                  <div className="text-sm font-medium truncate">{String(p.side)} {String(p.outcome)} · {String(p.event_id).slice(0, 8)}</div>
                  <div className="text-xs text-zinc-500">Entry {Number(p.entry_price).toFixed(2)} · Qty {String(p.quantity)}</div>
                </div>
                <div className="text-sm font-mono font-medium">
                  {p.unrealized_pnl != null ? `$${Number(p.unrealized_pnl).toFixed(2)}` : "—"}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Quick actions */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 sm:gap-4">
        <QuickLink to="/strategies" title="Strategies" desc="Build with natural language" />
        <QuickLink to="/backtests" title="Backtests" desc="Validate before deploying" />
        <QuickLink to="/events" title="Markets" desc="Browse DreamDEX events" />
      </div>
    </div>
  );
}

function StatCard({ label, value, sub, accent }: { label: string; value: string; sub: string; accent?: boolean }) {
  return (
    <div className="rounded-2xl bg-gradient-to-br from-white/[0.07] to-white/[0.02] border border-white/10 backdrop-blur p-4 sm:p-5">
      <div className="text-[11px] tracking-widest uppercase text-zinc-500">{label}</div>
      <div className={`text-lg sm:text-xl font-semibold mt-2 tracking-tight ${accent ? "text-emerald-400" : "text-white"}`}>{value}</div>
      <div className="text-xs text-zinc-500 mt-1">{sub}</div>
    </div>
  );
}

function QuickLink({ to, title, desc }: { to: string; title: string; desc: string }) {
  return (
    <Link to={to} className="rounded-2xl bg-white/[0.04] hover:bg-white/[0.07] border border-white/10 p-4 sm:p-5 transition-colors group">
      <div className="text-sm font-medium group-hover:text-white">{title}</div>
      <div className="text-xs text-zinc-500 mt-1">{desc}</div>
    </Link>
  );
}
