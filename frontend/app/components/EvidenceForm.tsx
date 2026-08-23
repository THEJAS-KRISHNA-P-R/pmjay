"use client";

import { useState, useTransition } from "react";
import { addEvidence, ApiError } from "@/lib/api";
import type { EvidenceEntry } from "@/lib/types";

export function EvidenceForm({
  caseId,
  prompt,
  initialEvidence,
}: {
  caseId: string;
  prompt: string;
  initialEvidence: EvidenceEntry[];
}) {
  const [evidence, setEvidence] = useState(initialEvidence);
  const [staffName, setStaffName] = useState("");
  const [approxTime, setApproxTime] = useState("");
  const [note, setNote] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);

    if (!staffName.trim() && !approxTime.trim() && !note.trim()) {
      setError("Add at least one detail before saving.");
      return;
    }

    startTransition(async () => {
      try {
        const updated = await addEvidence(caseId, {
          staff_name: staffName.trim() || undefined,
          approx_time: approxTime.trim() || undefined,
          note: note.trim() || undefined,
        });
        setEvidence(updated.evidence ?? []);
        setStaffName("");
        setApproxTime("");
        setNote("");
      } catch (err) {
        setError(err instanceof ApiError ? err.message : "Could not save this right now.");
      }
    });
  }

  return (
    <section className="glass-panel p-6 sm:p-8 animate-enter">
      <div className="flex items-center gap-2.5">
        <div className="flex h-6 w-6 items-center justify-center rounded-lg bg-teal-100 text-teal-800">
          <svg
            className="h-3.5 w-3.5"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2.5"
            strokeLinecap="round"
            strokeLinejoin="round"
            aria-hidden="true"
          >
            <path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z" />
            <path d="m15 5 4 4" />
          </svg>
        </div>
        <h2 className="text-lg sm:text-xl font-bold tracking-tight text-teal-950">
          Keep a record
        </h2>
      </div>
      <p className="mt-2 text-xs sm:text-sm leading-relaxed text-sand-900/75">{prompt}</p>

      <form onSubmit={handleSubmit} className="mt-5 space-y-4">
        <div className="grid gap-3.5 sm:grid-cols-2">
          <div>
            <label htmlFor="staffName" className="block text-xs sm:text-sm font-bold text-sand-900">
              Staff member&rsquo;s name
            </label>
            <input
              id="staffName"
              type="text"
              value={staffName}
              onChange={(e) => setStaffName(e.target.value)}
              className="mt-1.5 w-full rounded-2xl bg-sand-100/60 px-4 py-3 text-sm text-sand-900 shadow-[inset_0_2px_4px_rgba(0,0,0,0.03)] placeholder:text-sand-900/40 transition focus:bg-white focus:shadow-[0_0_0_3px_rgba(47,127,115,0.25)] focus:outline-none"
              placeholder="If you know it"
            />
          </div>
          <div>
            <label htmlFor="approxTime" className="block text-xs sm:text-sm font-bold text-sand-900">
              Approximate time
            </label>
            <input
              id="approxTime"
              type="text"
              value={approxTime}
              onChange={(e) => setApproxTime(e.target.value)}
              className="mt-1.5 w-full rounded-2xl bg-sand-100/60 px-4 py-3 text-sm text-sand-900 shadow-[inset_0_2px_4px_rgba(0,0,0,0.03)] placeholder:text-sand-900/40 transition focus:bg-white focus:shadow-[0_0_0_3px_rgba(47,127,115,0.25)] focus:outline-none"
              placeholder="e.g. around 3:30 PM"
            />
          </div>
        </div>
        
        <div>
          <label htmlFor="note" className="block text-xs sm:text-sm font-bold text-sand-900">
            Anything else worth noting
          </label>
          <textarea
            id="note"
            value={note}
            onChange={(e) => setNote(e.target.value)}
            rows={2}
            className="mt-1.5 w-full rounded-2xl bg-sand-100/60 px-4 py-3 text-sm text-sand-900 shadow-[inset_0_2px_4px_rgba(0,0,0,0.03)] placeholder:text-sand-900/40 transition focus:bg-white focus:shadow-[0_0_0_3px_rgba(47,127,115,0.25)] focus:outline-none"
            placeholder="e.g. said the pre-auth server was down, or gave no written explanation"
          />
        </div>

        {error && (
          <p role="alert" className="text-xs sm:text-sm font-bold text-tier-red-strong">
            {error}
          </p>
        )}

        <div>
          <button
            type="submit"
            disabled={isPending}
            className="inline-flex items-center gap-2 rounded-2xl bg-teal-800 px-7 py-3 text-sm font-bold text-white shadow-sm transition touch-spring hover:bg-teal-900 active:scale-95 disabled:opacity-60 focus-visible:outline-teal-600"
          >
            {isPending ? "Saving…" : "Save this"}
          </button>
        </div>
      </form>

      {evidence.length > 0 && (
        <div className="mt-6 pt-5">
          <p className="text-xs font-bold uppercase tracking-wider text-sand-900/60 mb-3">
            Logged Details
          </p>
          <ul className="space-y-2.5">
            {evidence.map((e, i) => (
              <li 
                key={i} 
                className="flex items-center gap-2.5 rounded-xl bg-sand-100/60 px-4 py-2.5 text-xs sm:text-sm text-sand-900 font-medium"
              >
                <span className="h-2 w-2 rounded-full bg-teal-600 shrink-0" aria-hidden="true" />
                <span>{[e.staff_name, e.approx_time, e.note].filter(Boolean).join(" — ")}</span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </section>
  );
}
