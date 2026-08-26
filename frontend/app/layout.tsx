import type { Metadata, Viewport } from "next";
import localFont from "next/font/local";
import "./globals.css";

// Self-hosted local font files rather than next/font/google — this build
// environment's network sandbox doesn't reach fonts.googleapis.com (see
// ../../ARCHITECTURE.md for the parallel story on the Go backend's
// dependencies), so the actual OFL-licensed font files were fetched
// directly from Google's own font repository on GitHub and are hosted
// from this repo instead. Functionally identical to next/font/google's
// output — self-hosted, no runtime request to any Google CDN, zero
// layout shift — this just does the fetch at a different time. License
// files are alongside the font files in ./fonts.
//
// Atkinson Hyperlegible is used for all functional UI text — it's a
// typeface designed specifically for readers who find text difficult to
// read accurately (published by the Braille Institute), which is a
// functional requirement for this product's named user, not a stylistic
// pick. Fraunces is used only for the wordmark/hero title — everything a
// family actually has to read to understand their situation stays in the
// hyperlegible face.
const atkinson = localFont({
  src: [
    { path: "./fonts/AtkinsonHyperlegible-Regular.ttf", weight: "400", style: "normal" },
    { path: "./fonts/AtkinsonHyperlegible-Bold.ttf", weight: "700", style: "normal" },
  ],
  variable: "--font-atkinson",
  display: "swap",
});

const fraunces = localFont({
  src: [{ path: "./fonts/Fraunces-Variable.ttf", weight: "300 900", style: "normal" }],
  variable: "--font-fraunces",
  display: "swap",
});

export const metadata: Metadata = {
  title: "PMJAY Point-of-Denial Advocate",
  description:
    "Free help understanding whether a hospital's PMJAY coverage denial is correct, right when you're standing at the billing desk.",
  icons: {
    icon: [
      { url: "/favicon.ico" },
      { url: "/icon.png", type: "image/png" },
    ],
    apple: [
      { url: "/apple-icon.png" },
    ],
  },
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  maximumScale: 5,
  themeColor: "#faf8f3",
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" className="scroll-smooth">
      <head>
        {/* Synchronous and pre-paint on purpose — this is the same
            flash-of-wrong-appearance problem a dark-mode toggle has to
            solve, just for text size instead of color scheme. Reading
            localStorage from a client component (app/settings/page.tsx
            or elsewhere) would only run after first paint, so anyone
            with large-text mode on would see one frame (or, on a slow
            connection, much longer) of normal-size text first. Fails
            silently and simply does nothing if localStorage is
            unavailable (private browsing, storage disabled) — same
            posture as lib/caseHistory.ts's own defensive handling. */}
        <script
          dangerouslySetInnerHTML={{
            __html: `try{if(localStorage.getItem('pmjay-advocate:text-scale')==='large'){document.documentElement.setAttribute('data-text-scale','large');}}catch(e){}`,
          }}
        />
      </head>
      <body className={`${atkinson.variable} ${fraunces.variable} font-sans antialiased min-h-screen flex flex-col`}>
        {children}
      </body>
    </html>
  );
}
