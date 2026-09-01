import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAccount, useSignMessage } from "wagmi";
import { ConnectButton } from "@rainbow-me/rainbowkit";
import { authApi } from "../services/api";
import { useAuth } from "../store/auth";

export default function Login() {
  const { login } = useAuth();
  const nav = useNavigate();
  const { address, isConnected } = useAccount();
  const { signMessageAsync } = useSignMessage();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const googleConfigured = true; // backend returns 500 if not configured — we handle gracefully

  const handleGoogle = () => {
    window.location.href = "/api/v1/auth/google/login";
  };

  const handleSiweLogin = async () => {
    if (!address) return;
    setLoading(true);
    setError("");
    try {
      const { data } = await authApi.nonce(address);
      const message: string = data.message;
      const signature = await signMessageAsync({ message });
      const res = await authApi.walletVerify({ address, message, signature });
      login(res.data.token, res.data.user);
      nav("/dashboard");
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "SIWE login failed";
      // Try to extract axios error
      const axiosMsg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      setError(axiosMsg || msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-md mx-auto mt-8 sm:mt-16 px-4 sm:px-0 space-y-6">
      <div className="text-center">
        <h1 className="text-xl sm:text-2xl font-semibold tracking-tight">Sign in to PredictOS</h1>
        <p className="text-sm text-zinc-400 mt-2">Choose your preferred sign-in method</p>
      </div>

      <div className="rounded-2xl bg-white/[0.04] border border-white/10 backdrop-blur p-5 sm:p-6 space-y-4">
        {/* Google */}
        <button
          onClick={handleGoogle}
          className="w-full flex items-center justify-center gap-2 bg-white text-zinc-900 py-3 rounded-full text-sm font-medium hover:bg-zinc-100 transition-colors"
        >
          <span className="w-5 h-5 rounded-full bg-white border border-zinc-200 flex items-center justify-center text-xs">G</span>
          Continue with Google
        </button>

        <div className="flex items-center gap-3">
          <div className="flex-1 h-px bg-white/10" />
          <span className="text-xs text-zinc-500">or</span>
          <div className="flex-1 h-px bg-white/10" />
        </div>

        {/* WalletConnect via RainbowKit */}
        <div className="space-y-3">
          <div className="flex justify-center">
            <ConnectButton />
          </div>

          {isConnected && address && (
            <button
              onClick={handleSiweLogin}
              disabled={loading}
              className="w-full bg-[#3b82f6] hover:bg-[#2563eb] text-white py-3 rounded-full text-sm font-medium disabled:opacity-50 transition-colors"
            >
              {loading ? "Signing..." : "Sign message to login"}
            </button>
          )}

          {!isConnected && (
            <p className="text-xs text-zinc-500 text-center">
              Connect wallet above, then sign a message to authenticate (SIWE).
            </p>
          )}

          {error && (
            <p className="text-xs text-red-400 text-center bg-red-500/10 border border-red-500/20 rounded-lg px-3 py-2">
              {error}
            </p>
          )}

          {!googleConfigured && (
            <p className="text-xs text-amber-400 text-center">Google OAuth not configured — set GOOGLE_CLIENT_ID in backend .env</p>
          )}
        </div>

        <p className="text-[11px] text-zinc-500 text-center leading-relaxed">
          By signing in, you agree to our Terms. Wallet signature proves ownership — no gas required.
        </p>
      </div>

      <p className="text-xs text-zinc-500 text-center">
        No wallet? <span className="text-zinc-300">Sign in with Google</span> and connect a wallet later for live trading.
      </p>
    </div>
  );
}
