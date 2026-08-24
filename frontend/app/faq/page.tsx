import type { Metadata } from "next";
import { Header } from "../components/Header";
import { Footer } from "../components/landing/Footer";
import { FaqSection } from "../components/landing/FaqSection";

export const metadata: Metadata = {
  title: "Frequently Asked Questions — PMJAY Advocate",
  description:
    "Common questions about PMJAY coverage, point-of-denial disputes, hospital obligations, and how PMJAY Advocate works.",
};

export default function FaqPage() {
  return (
    <div className="min-h-screen flex flex-col bg-sand-50">
      <Header />

      <main className="mx-auto w-full max-w-6xl flex-1 px-6 sm:px-8 lg:px-10 pt-8 sm:pt-12 lg:pt-16 pb-12 sm:pb-16 lg:pb-24 space-y-10 animate-fade-in">
        <div className="space-y-3.5 max-w-2xl">
          <p className="text-xs font-bold uppercase tracking-wider text-teal-700">Questions &amp; answers</p>
          <h1 className="font-display text-3xl sm:text-4xl lg:text-5xl font-semibold leading-tight tracking-tight-display text-teal-950">
            Frequently Asked Questions
          </h1>
          <p className="text-base sm:text-lg leading-relaxed text-sand-700 font-medium">
            Common questions about what PMJAY covers, what hospitals are allowed to do, and what to expect
            from this tool.
          </p>
        </div>

        <FaqSection hideHeader={true} />
      </main>

      <Footer />
    </div>
  );
}
