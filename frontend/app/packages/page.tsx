import type { Metadata } from "next";
import Link from "next/link";
import { Header } from "../components/Header";
import { Footer } from "../components/landing/Footer";
import { PackagesExplorer } from "./PackagesExplorer";
import { IconAlertTriangle, IconPhone, IconArrowRight } from "../components/icons";

export const metadata: Metadata = {
  title: "PM-JAY package search — verified HBP codes and rates",
  description:
    "Search 300 Ayushman Bharat (PM-JAY) Health Benefit Packages whose code, specialty, and indicative rate were checked against a published government schedule. Free, no login.",
  alternates: {
    canonical: "/packages",
  },
  openGraph: {
    url: "/packages",
    title: "PM-JAY package search — verified HBP codes and rates",
    description:
      "Search 300 verified Ayushman Bharat Health Benefit Packages by procedure, specialty, or HBP code. No login, no model, no invented numbers.",
  },
};

export default function PackagesPage() {
  return (
    <div className="min-h-screen flex flex-col bg-sand-50">
      <Header />

      <main className="mx-auto w-full max-w-5xl flex-1 px-4 sm:px-8 lg:px-10 pt-10 sm:pt-14 pb-16 space-y-8 sm:space-y-10 animate-fade-in">
        <div className="space-y-4">
          <p className="text-xs sm:text-sm font-bold uppercase tracking-wider text-emerald-700">
            Health Benefit Package index
          </p>
          <h1 className="font-display text-3xl sm:text-4xl lg:text-5xl font-bold tracking-tight-display text-ink-950 max-w-3xl">
            Find the package and rate behind a PM-JAY denial
          </h1>
          <p className="text-sm sm:text-base text-sand-700 font-medium leading-relaxed max-w-2xl">
            When a hospital says a treatment is &ldquo;not covered&rdquo;, the useful questions are
            concrete: is there a Health Benefit Package for this procedure, what is its code, and
            what is the published rate? Search it here before you argue about money.
          </p>
          <p className="text-sm sm:text-base text-sand-700 font-medium leading-relaxed max-w-2xl">
            This index is a plain database lookup. It does not use an AI model, needs no account,
            and returns nothing when nothing matches — it will never invent a package to fill the
            page.
          </p>
        </div>

        <div className="card p-4 sm:p-5 border border-emerald-200/80 bg-emerald-50/50">
          <div className="flex items-start gap-3">
            <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl bg-emerald-100 text-emerald-800 mt-0.5">
              <IconPhone className="h-4 w-4" />
            </span>
            <p className="text-xs sm:text-sm text-emerald-950 font-medium leading-relaxed">
              <strong>Care comes first.</strong> Do not delay treatment while you check a rate. If
              you are at a hospital now, call the PMJAY helpline on{" "}
              <a href="tel:14555" className="font-bold underline">
                14555
              </a>{" "}
              and ask them to confirm the package. A billing dispute can wait; a patient cannot.
            </p>
          </div>
        </div>

        <PackagesExplorer />

        <section className="grid gap-4 sm:grid-cols-2" aria-label="How to read these results">
          <div className="card p-5 sm:p-6 space-y-2">
            <h2 className="text-base font-bold text-sand-900">What &ldquo;verified&rdquo; means</h2>
            <p className="text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
              The package code, name, specialty, and indicative rate were checked against a
              published National Health Authority Health Benefit Package schedule. That is a
              citation check, not a guarantee: your hospital&rsquo;s empanelment, the clinical
              facts, and pre-authorisation decide what is actually paid.
            </p>
          </div>
          <div className="card p-5 sm:p-6 space-y-2">
            <h2 className="text-base font-bold text-sand-900">What &ldquo;name only&rdquo; means</h2>
            <p className="text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
              The procedure is a real PM-JAY category, but this index has not verified its code or
              rate. You get the name so you can recognise it and ask about it. Neither a code nor a
              rupee figure is shown, because showing an unchecked one would be worse than showing
              none.
            </p>
          </div>
        </section>

        <section className="card p-5 sm:p-6 space-y-3 border border-sand-200/80">
          <div className="flex items-start gap-3">
            <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl bg-tier-amber-icon text-tier-amber-text mt-0.5">
              <IconAlertTriangle className="h-4 w-4" />
            </span>
            <div className="space-y-2">
              <h2 className="text-base font-bold text-sand-900">Coverage, stated honestly</h2>
              <p className="text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                This index covers 315 procedure names out of roughly 1,900 in the full PM-JAY
                master list, and 300 of those carry a verified code and rate. General surgery,
                orthopaedics, obstetrics and gynaecology, ophthalmology, ENT, urology, and
                neurosurgery are the known thin areas. A missing result means &ldquo;not in this
                index&rdquo;, never &ldquo;not covered by PM-JAY&rdquo;.
              </p>
              <p className="text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                Need the full decision, not just a lookup?{" "}
                <Link href="/cases/new" className="font-bold text-emerald-800 underline">
                  Describe the denial
                </Link>{" "}
                and get a care-first read on whether it looks correct.
              </p>
            </div>
          </div>
        </section>

        <div className="pt-2">
          <Link href="/cases/new" className="btn-primary tap-target px-5 py-3 text-sm sm:text-base">
            <span>Check a specific denial</span>
            <IconArrowRight className="h-4 w-4" />
          </Link>
        </div>
      </main>

      <Footer />
    </div>
  );
}
