import type { Metadata } from "next";
import { Header } from "../components/Header";
import { Footer } from "../components/landing/Footer";
import { GuideReader } from "./GuideReader";

export const metadata: Metadata = {
  title: "Your Rights — PMJAY Advocate",
  description:
    "A plain-language guide to what PMJAY covers, what hospitals aren't allowed to do, and what to do if your coverage is denied.",
};

export default function GuidePage() {
  return (
    <div className="min-h-screen flex flex-col bg-sand-50">
      <Header />

      <main className="mx-auto w-full max-w-6xl flex-1 px-4 sm:px-8 lg:px-10 pt-6 sm:pt-12 lg:pt-16 pb-12 sm:pb-16 lg:pb-24 animate-fade-in">
        <div className="max-w-3xl space-y-3.5 mb-8 sm:mb-10">
          <p className="text-xs font-bold uppercase tracking-wider text-sand-500">Know before you go</p>
          <h1 className="font-display text-3xl sm:text-4xl lg:text-5xl font-semibold leading-tight tracking-tight-display text-ink-950">
            Your Rights Under PMJAY
          </h1>
          <p className="text-base sm:text-lg leading-relaxed text-sand-700 font-medium">
            General guidance, written plainly (not a substitute for the official scheme rules), but enough to
            know when something at the billing desk doesn&rsquo;t add up.
          </p>
        </div>

        <GuideReader />
      </main>

      <Footer />
    </div>
  );
}
