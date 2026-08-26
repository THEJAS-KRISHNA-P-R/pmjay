import type { Metadata } from "next";
import { Header } from "../components/Header";
import { Footer } from "../components/landing/Footer";
import { IconShieldCheck, IconLock, IconCheck, IconTrash } from "../components/icons";
import Link from "next/link";

export const metadata: Metadata = {
  title: "Privacy Policy — PMJAY Advocate",
  description:
    "How PMJAY Advocate protects your privacy: zero accounts, no server-side identity tracking, local browser-only storage, and full DPDP Act 2023 compliance.",
};

export default function PrivacyPage() {
  return (
    <div className="min-h-screen flex flex-col bg-sand-50">
      <Header />

      <main className="mx-auto w-full max-w-4xl flex-1 px-4 sm:px-8 lg:px-10 pt-12 sm:pt-16 pb-16 space-y-10 animate-fade-in">
        {/* Header */}
        <div className="space-y-3">
          <div className="inline-flex items-center gap-2">
            <span className="text-xs sm:text-sm font-bold uppercase tracking-wider text-emerald-700">
              Data Protection &amp; Privacy
            </span>
          </div>
          <h1 className="font-display text-3xl sm:text-4xl lg:text-5xl font-bold tracking-tight-display text-ink-950">
            Privacy Policy
          </h1>
          <p className="text-sm sm:text-base text-sand-600 font-medium leading-relaxed">
            Last updated: August 2026. Built from the ground up for privacy-first, account-free public access in compliance with the Digital Personal Data Protection Act (DPDP Act 2023, India).
          </p>
        </div>

        {/* Highlight Banner */}
        <div className="card p-6 sm:p-7 border border-emerald-200/80 bg-emerald-50/50 space-y-3 shadow-xs">
          <div className="flex items-start gap-3">
            <span className="flex h-8 w-8 items-center justify-center rounded-xl bg-emerald-100 text-emerald-800 shrink-0 mt-0.5">
              <IconLock className="h-4.5 w-4.5" />
            </span>
            <div className="space-y-1">
              <h2 className="text-base font-bold text-emerald-950">
                Core Privacy Guarantee: No Accounts, No Tracking
              </h2>
              <p className="text-xs sm:text-sm text-emerald-900 leading-relaxed font-medium">
                PMJAY Advocate does not require a login, phone number, Aadhaar number, or email address to use. We do not build identity profiles, sell user data, or run third-party advertising tracking scripts.
              </p>
            </div>
          </div>
        </div>

        {/* Section 1: Information We Process */}
        <section className="card p-6 sm:p-8 space-y-4">
          <h2 className="text-lg sm:text-xl font-bold text-sand-900 flex items-center gap-2">
            <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-sand-100 text-sand-800 text-xs font-bold">
              1
            </span>
            Information We Process
          </h2>
          <div className="space-y-3 text-xs sm:text-sm text-sand-700 leading-relaxed font-medium">
            <p>
              When you use PMJAY Advocate to evaluate a hospital billing denial, you submit an unstructured text description of your medical situation and what the billing desk stated.
            </p>
            <ul className="space-y-2 pl-4 list-disc marker:text-emerald-700">
              <li>
                <strong>Case Situation Input:</strong> The narrative you type is transmitted securely via HTTPS to our backend evaluation engine to cross-reference against official NHA Health Benefit Package (HBP) master schedules.
              </li>
              <li>
                <strong>Unguessable Unique Case ID:</strong> Each evaluated case is assigned a cryptographically random UUIDv4. There is no public directory of cases; only someone holding the exact case URL can view that specific evaluation.
              </li>
              <li>
                <strong>Local Browser Storage:</strong> Your case history and optional user preferences (such as name, phone, or language preference) are saved exclusively in your browser&rsquo;s local storage (`localStorage`). They are never aggregated into a centralized user registry.
              </li>
            </ul>
          </div>
        </section>

        {/* Section 2: Ephemeral LLM Processing */}
        <section className="card p-6 sm:p-8 space-y-4">
          <h2 className="text-lg sm:text-xl font-bold text-sand-900 flex items-center gap-2">
            <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-sand-100 text-sand-800 text-xs font-bold">
              2
            </span>
            AI &amp; Large Language Model Processing
          </h2>
          <div className="space-y-3 text-xs sm:text-sm text-sand-700 leading-relaxed font-medium">
            <p>
              To translate complex medical narratives into structured PMJAY package codes, case descriptions are processed via secure enterprise AI API endpoints.
            </p>
            <div className="grid sm:grid-cols-2 gap-3 pt-1">
              <div className="card p-3.5 space-y-1">
                <p className="font-bold text-sand-900 flex items-center gap-1.5">
                  <IconCheck className="h-3.5 w-3.5 text-emerald-700" />
                  Zero Model Training
                </p>
                <p className="text-xs text-sand-600">
                  Data sent to API endpoints is strictly ephemeral and is not used to train public AI models.
                </p>
              </div>
              <div className="card p-3.5 space-y-1">
                <p className="font-bold text-sand-900 flex items-center gap-1.5">
                  <IconCheck className="h-3.5 w-3.5 text-emerald-700" />
                  Strict Transport Encryption
                </p>
                <p className="text-xs text-sand-600">
                  All requests use TLS 1.3 encryption in transit with restricted API gateway tokenization.
                </p>
              </div>
            </div>
          </div>
        </section>

        {/* Section 3: Cookies & Analytics */}
        <section className="card p-6 sm:p-8 space-y-4">
          <h2 className="text-lg sm:text-xl font-bold text-sand-900 flex items-center gap-2">
            <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-sand-100 text-sand-800 text-xs font-bold">
              3
            </span>
            Cookies and Telemetry
          </h2>
          <div className="space-y-3 text-xs sm:text-sm text-sand-700 leading-relaxed font-medium">
            <p>
              PMJAY Advocate uses <strong>zero third-party tracking cookies</strong>, zero advertising pixels (such as Meta Pixel or Google Ads tags), and no behavioral surveillance trackers.
            </p>
            <p>
              Only essential, anonymous server-side rate limiting headers and operational error monitoring are used to protect the public infrastructure against automated distributed denial-of-service (DDoS) abuse.
            </p>
          </div>
        </section>

        {/* Section 4: Data Retention & User Control */}
        <section className="card p-6 sm:p-8 space-y-4">
          <h2 className="text-lg sm:text-xl font-bold text-sand-900 flex items-center gap-2">
            <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-sand-100 text-sand-800 text-xs font-bold">
              4
            </span>
            Your Rights and Data Control
          </h2>
          <div className="space-y-3 text-xs sm:text-sm text-sand-700 leading-relaxed font-medium">
            <p>
              Under the Digital Personal Data Protection Act 2023, you retain absolute ownership and control over any data stored on your device:
            </p>
            <ul className="space-y-2 pl-4 list-disc marker:text-emerald-700">
              <li>
                <strong>Instant Data Purge:</strong> You can wipe all case history and saved preferences from your browser at any time via the <Link href="/settings" className="font-bold text-emerald-700 underline hover:text-emerald-900">Settings Page</Link>.
              </li>
              <li>
                <strong>Private Browsing:</strong> Using Incognito / Private mode automatically deletes all local session data as soon as you close your browser tab.
              </li>
            </ul>
          </div>
        </section>

        {/* Section 5: Contact & Inquiries */}
        <section className="card p-6 sm:p-8 space-y-3 border border-sand-200/80 bg-white">
          <h2 className="text-base sm:text-lg font-bold text-sand-900">
            Privacy Inquiries
          </h2>
          <p className="text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
            For questions regarding this privacy policy or our open-source data practices, please consult our <Link href="/about" className="text-emerald-700 font-bold underline">About Page</Link> or official project documentation.
          </p>
        </section>
      </main>

      <Footer />
    </div>
  );
}
