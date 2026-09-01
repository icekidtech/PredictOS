/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      colors: {
        brand: { 500: "#3b82f6", 600: "#2563eb", 700: "#1d4ed8" },
        surface: { 900: "#0a0e1a", 800: "#111827", 700: "#1f2937" },
      },
      fontFamily: { mono: ["JetBrains Mono", "monospace"] },
    },
  },
  plugins: [],
};
