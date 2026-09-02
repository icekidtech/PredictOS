import { useEffect, useState } from "react";

type ToastProps = { message: string; onClose: () => void };

export function Toast({ message, onClose }: ToastProps) {
  useEffect(() => {
    const t = setTimeout(onClose, 2500);
    return () => clearTimeout(t);
  }, [onClose]);
  return (
    <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 px-4 py-2.5 rounded-full bg-white text-zinc-900 text-sm font-medium shadow-xl border border-zinc-200 flex items-center gap-2">
      <span className="w-2 h-2 rounded-full bg-emerald-500" />
      {message}
    </div>
  );
}

export function useToast() {
  const [msg, setMsg] = useState<string | null>(null);
  const show = (m: string) => setMsg(m);
  const toast = msg ? <Toast message={msg} onClose={() => setMsg(null)} /> : null;
  return { show, toast };
}
