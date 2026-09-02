import { http, createConfig } from "wagmi";
import { mainnet, sepolia } from "wagmi/chains";
import { injected, walletConnect } from "wagmi/connectors";

// Somnia chains
export const somniaTestnet = {
  id: 50312,
  name: "Somnia Testnet",
  network: "somnia-testnet",
  nativeCurrency: { name: "Somnia", symbol: "STT", decimals: 18 },
  rpcUrls: { default: { http: ["https://dream-rpc.somnia.network"] } },
  blockExplorers: { default: { name: "Somnia Explorer", url: "https://explorer.somnia.network" } },
} as const;

export const somniaMainnet = {
  id: 5031,
  name: "Somnia Mainnet",
  network: "somnia-mainnet",
  nativeCurrency: { name: "Somnia", symbol: "SOMI", decimals: 18 },
  rpcUrls: { default: { http: ["https://api.infra.mainnet.somnia.network"] } },
  blockExplorers: { default: { name: "Somnia Explorer", url: "https://explorer.somnia.network" } },
} as const;

const projectId = import.meta.env.VITE_WALLETCONNECT_PROJECT_ID || "";

export const wagmiConfig = createConfig({
  chains: [somniaTestnet, somniaMainnet, sepolia, mainnet],
  transports: {
    [somniaTestnet.id]: http(),
    [somniaMainnet.id]: http(),
    [sepolia.id]: http(),
    [mainnet.id]: http(),
  },
  connectors: [
    injected(),
    ...(projectId ? [walletConnect({ projectId })] : []),
  ],
});
