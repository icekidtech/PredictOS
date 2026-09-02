import { useEffect, useState, useCallback } from "react";

export function useNetwork() {
  const [network, setNetwork] = useState<string>(() => localStorage.getItem("network") || "testnet");

  const updateNetwork = useCallback((nw: string) => {
    localStorage.setItem("network", nw);
    setNetwork(nw);
    window.dispatchEvent(new CustomEvent("network-change", { detail: nw }));
  }, []);

  useEffect(() => {
    const handler = (e: Event) => {
      const nw = (e as CustomEvent).detail as string;
      setNetwork(nw);
    };
    window.addEventListener("network-change", handler);
    const onStorage = (e: StorageEvent) => {
      if (e.key === "network" && e.newValue) setNetwork(e.newValue);
    };
    window.addEventListener("storage", onStorage);
    return () => {
      window.removeEventListener("network-change", handler);
      window.removeEventListener("storage", onStorage);
    };
  }, []);

  return { network, setNetwork: updateNetwork };
}
