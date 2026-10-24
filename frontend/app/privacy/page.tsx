import type { Metadata } from "next";
import { Header } from "../components/Header";
import { Footer } from "../components/landing/Footer";
import { IconLock, IconCheck } from "../components/icons";
import Link from "next/link";

export const metadata: Metadata = {
  title: "Privacy Policy",
  description:
    "What PMJAY Advocate stores, who processes it, how long it is kept, and how to delete a case. No account, no Aadhaar, no advertising trackers.",
  alternates: {
    canonical: "/privacy",
  },
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
            Last updated: October 2026. This page describes what actually happens to what you
            type. It is written to be checked against the code, not to sound reassuring.
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
                The short version
              </h2>
              <p className="text-xs sm:text-sm text-emerald-900 leading-relaxed font-medium">
                No account, no phone number, no Aadhaar number, no advertising trackers. What you
                type is sent to an AI provider to be structured, and the resulting case is stored
                on our server so you can reopen it. It is deleted automatically after 30 days of
                inactivity, and you can delete it yourself at any time from the case page.
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
            What we collect, and why
          </h2>
          <div className="space-y-3 text-xs sm:text-sm text-sand-700 leading-relaxed font-medium">
            <ul className="space-y-2.5 pl-4 list-disc marker:text-emerald-700">
              <li>
                <strong>What you type about the denial.</strong> This is the only input. It is
                sent to the backend, where it is matched against the Health Benefit Package
                reference data and then processed by an AI model (see section 2).
              </li>
              <li>
                <strong>The case that results.</strong> The explanation, suggested steps, desk
                script, complaint draft, and any evidence you add are stored on the server under a
                random case ID, so the page still works when you come back to it.
              </li>
              <li>
                <strong>Your browser&rsquo;s local list.</strong> Your case history and preferences
                (such as text size) live in this browser&rsquo;s <code>localStorage</code>. They are
                never uploaded as a profile and never merged with anyone else&rsquo;s.
              </li>
              <li>
                <strong>Your IP address, briefly.</strong> It is used in memory to rate-limit
                requests and is not written to the case record. Server logs record the route and
                status, with case IDs removed.
              </li>
            </ul>
            <p className="rounded-xl bg-sand-50 border border-sand-200/70 p-3.5">
              <strong>We do not ask for, and do not want,</strong> your Aadhaar number, PM-JAY card
              number, hospital UHID, full medical records, photographs of documents, or your exact
              address. Describe the situation, not your identity.
            </p>
          </div>
        </section>

        {/* Section 2: AI processing */}
        <section className="card p-6 sm:p-8 space-y-4">
          <h2 className="text-lg sm:text-xl font-bold text-sand-900 flex items-center gap-2">
            <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-sand-100 text-sand-800 text-xs font-bold">
              2
            </span>
            AI processing, and who receives your words
          </h2>
          <div className="space-y-3 text-xs sm:text-sm text-sand-700 leading-relaxed font-medium">
            <p>
              To turn a plain-language description into a structured PM-JAY package match, the
              text you submit is sent to the AI provider this deployment is configured with
              (Anthropic Claude, Groq, or Google Gemini). That provider processes the text on its
              infrastructure under its own API terms.
            </p>
            <div className="grid sm:grid-cols-2 gap-3 pt-1">
              <div className="card p-3.5 space-y-1">
                <p className="font-bold text-sand-900 flex items-center gap-1.5">
                  <IconCheck className="h-3.5 w-3.5 text-emerald-700" />
                  Not used for training
                </p>
                <p className="text-xs text-sand-600">
                  We do not use your text to train any model, and the providers&rsquo; commercial
                  API terms do not permit them to train on API inputs by default. We cannot make
                  promises beyond those terms on the provider&rsquo;s behalf.
                </p>
              </div>
              <div className="card p-3.5 space-y-1">
                <p className="font-bold text-sand-900 flex items-center gap-1.5">
                  <IconCheck className="h-3.5 w-3.5 text-emerald-700" />
                  Encrypted in transit
                </p>
                <p className="text-xs text-sand-600">
                  Browser to server and server to provider both use HTTPS. Your text is not sold,
                  shared with advertisers, or used to build an identity profile.
                </p>
              </div>
            </div>
            <p>
              If you would rather not send anything to a model, use the{" "}
              <Link href="/packages" className="font-bold text-emerald-700 underline hover:text-emerald-900">
                package search
              </Link>
              , which performs a plain database lookup with no AI call at all.
            </p>
          </div>
        </section>

        {/* Section 3: Retention */}
        <section className="card p-6 sm:p-8 space-y-4">
          <h2 className="text-lg sm:text-xl font-bold text-sand-900 flex items-center gap-2">
            <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-sand-100 text-sand-800 text-xs font-bold">
              3
            </span>
            How long we keep it
          </h2>
          <div className="space-y-3 text-xs sm:text-sm text-sand-700 leading-relaxed font-medium">
            <ul className="space-y-2 pl-4 list-disc marker:text-emerald-700">
              <li>
                <strong>Server-side case: 30 days after the last change.</strong> An untouched case
                is deleted automatically. The retention window is configurable per deployment; a
                deployment may choose a shorter one.
              </li>
              <li>
                <strong>Your browser list: until you clear it.</strong> It stays in this browser
                until you delete a case, use &ldquo;clear all&rdquo; in{" "}
                <Link href="/settings" className="font-bold text-emerald-700 underline hover:text-emerald-900">
                  Settings
                </Link>
                , or clear site data. Private browsing discards it when the tab closes.
              </li>
              <li>
                <strong>No archive, no backup copy offered to you.</strong> Once a case is deleted
                or expires, it is gone. Print the PDF first if you need a permanent record for a
                complaint.
              </li>
            </ul>
          </div>
        </section>

        {/* Section 4: Access control and limits */}
        <section className="card p-6 sm:p-8 space-y-4">
          <h2 className="text-lg sm:text-xl font-bold text-sand-900 flex items-center gap-2">
            <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-sand-100 text-sand-800 text-xs font-bold">
              4
            </span>
            Who can open a case
          </h2>
          <div className="space-y-3 text-xs sm:text-sm text-sand-700 leading-relaxed font-medium">
            <p>
              There are no accounts. A case is reachable by anyone who has its link, because the
              random case ID <em>is</em> the access key. That is a deliberate trade — it means a
              family under stress never has to create a login — but it has a real consequence:
            </p>
            <p className="rounded-xl border border-tier-amber-border bg-tier-amber-bg p-3.5 text-tier-amber-text">
              Treat a case link like a private document. Do not post it in a public group. Anyone
              with the link can read it until it expires or you delete it.
            </p>
            <p>
              The ID is generated with a cryptographic random source, is not listed anywhere, and
              is stripped from server logs. There is no password reset and no account recovery,
              because there is no account.
            </p>
          </div>
        </section>

        {/* Section 5: Your choices */}
        <section className="card p-6 sm:p-8 space-y-4">
          <h2 className="text-lg sm:text-xl font-bold text-sand-900 flex items-center gap-2">
            <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-sand-100 text-sand-800 text-xs font-bold">
              5
            </span>
            Your choices and rights
          </h2>
          <div className="space-y-3 text-xs sm:text-sm text-sand-700 leading-relaxed font-medium">
            <p>
              The Digital Personal Data Protection Act, 2023 sets out rights to access, correction,
              and erasure of personal data. This project is an independent public-interest tool,
              not a company with a grievance officer, but the product is built so those actions are
              possible without asking anyone:
            </p>
            <ul className="space-y-2 pl-4 list-disc marker:text-emerald-700">
              <li>
                <strong>Access:</strong> the case page shows your original words and everything
                derived from them.
              </li>
              <li>
                <strong>Correction:</strong> add evidence to a case, or start a new case with a
                corrected description. Generated guidance is not hand-editable, because editing a
                generated explanation would make it no longer match the reasoning behind it.
              </li>
              <li>
                <strong>Erasure:</strong> use <em>Delete this case</em> at the bottom of any case
                page. That removes the server record and the browser entry. Clearing browser data
                removes the local entry but not the server record, so delete the case first if you
                want both gone.
              </li>
              <li>
                <strong>Complaints about this tool:</strong> open an issue on the project&rsquo;s
                public repository. Do not include a case link or medical details in a public issue.
              </li>
            </ul>
          </div>
        </section>

        {/* Section 6: Tracking */}
        <section className="card p-6 sm:p-8 space-y-4">
          <h2 className="text-lg sm:text-xl font-bold text-sand-900 flex items-center gap-2">
            <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-sand-100 text-sand-800 text-xs font-bold">
              6
            </span>
            Cookies, analytics, and advertising
          </h2>
          <div className="space-y-3 text-xs sm:text-sm text-sand-700 leading-relaxed font-medium">
            <p>
              No advertising pixels, no third-party behavioural analytics, no session recording, no
              data brokers. The only stored preferences are the ones you set on this device, and
              the only server-side counters are in-memory rate limits that reset when the process
              restarts.
            </p>
            <p>
              If a future deployment adds analytics, this page must be updated before that ships.
              The same applies to adding a database, an email service, or any other processor.
            </p>
          </div>
        </section>

        {/* Section 7: Contact */}
        <section className="card p-6 sm:p-8 space-y-3 border border-sand-200/80 bg-white">
          <h2 className="text-base sm:text-lg font-bold text-sand-900">
            Questions, or a correction to the data
          </h2>
          <p className="text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
            Rate and package corrections are genuinely useful and belong in the dataset, not in a
            private email. See the{" "}
            <Link href="/about" className="text-emerald-700 font-bold underline">
              About page
            </Link>{" "}
            for how the data is sourced, and the project repository for how to report a wrong
            figure. This tool is not affiliated with the National Health Authority, and it cannot
            change your PM-JAY record.
          </p>
        </section>
      </main>

      <Footer />
    </div>
  );
}
