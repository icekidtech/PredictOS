import { useState } from "react";
import { authApi } from "../services/api";
import { useAuth } from "../store/auth";
import { useNavigate } from "react-router-dom";

export default function Login() {
  const { login } = useAuth();
  const nav = useNavigate();
  const [wallet, setWallet] = useState("");
  const [loading, setLoading] = useState(false);

  const handleWalletLogin = async () => {
    if (!wallet) return;
    setLoading(true);
    try {
      // Simple wallet_address login (dev); SIWE flow uses nonce+signature in production
      const r = await authApi.login({ wallet_address: wallet });
      login(r.data.token, r.data.user);
      nav("/");
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : "login failed");
    } finally { setLoading(false); }
  };

  const handleGoogle = () => {
    window.location.href = "/api/v1/auth/google/login";
  };

  return (
    <div className="max-w-md mx-auto mt-16 space-y-6">
      <h1 className="text-2xl font-semibold text-center">Sign in to PredictOS</h1>

      <div className="bg-zinc-900 rounded-xl border border-zinc-800 p-6 space-y-4">
        <button onClick={handleGoogle} className="w-full bg-white text-zinc-900 py-2.5 rounded-md text-sm font-medium hover:bg-zinc-100">
          Continue with Google
        </button>
        <div className="text-center text-xs text-zinc-500">or</div>
        <input
          value={wallet} onChange={(e) => setWallet(e.target.value)}
          placeholder="Wallet address (0x...)"
          className="w-full bg-zinc-800 border border-zinc-700 rounded-md px-3 py-2 text-sm"
        />
        <button onClick={handleWalletLogin} disabled={loading} className="w-full bg-indigo-600 hover:bg-indigo-500 py-2.5 rounded-md text-sm disabled:opacity-50">
          {loading ? "Signing in..." : "Sign in with Wallet"}
        </button>
        <p className="text-xs text-zinc-500 text-center">WalletConnect / SIWE signature flow will be added with wagmi + RainbowKit.</p>
      </div>
    </div>
  );
}
