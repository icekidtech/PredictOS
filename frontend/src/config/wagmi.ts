import { http, createConfig } from "wagmi";
import { mainnet, sepolia } from "wagmi/chains";
import { injected, walletConnect } from "wagmi/connectors";

// Somnia testnet chain (add mainnet when available)
export const somniaTestnet = {
  id: 50312,
  name: "Somnia Testnet",
  network: "somnia-testnet",
  nativeCurrency: { name: "Somnia", symbol: "STT", decimals: 18 },
  rpcUrls: { default: { http: ["https://testnet.somnia.network"] } },
  blockExplorers: { default: { name: "Somnia Explorer", url: "https://testnet.somnia.network" } },
} as const;

const projectId = import.meta.env.VITE_WALLETCONNECT_PROJECT_ID || "";

export const wagmiConfig = createConfig({
  chains: [somniaTestnet, sepolia, mainnet],
  transports: {
    [somniaTestnet.id]: http(),
    [sepolia.id]: http(),
    [mainnet.id]: http(),
  },
  connectors: [
    injected(),
    ...(projectId ? [walletConnect({ projectId })] : []),
  ],
});
