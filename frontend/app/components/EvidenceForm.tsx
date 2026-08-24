"use client";

import { useState, useTransition } from "react";
import { addEvidence, ApiError } from "@/lib/api";
import type { EvidenceEntry } from "@/lib/types";
import { IconPencilLine } from "./icons";

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
    <section className="card p-6 sm:p-8 animate-fade-in-up">
      <div className="flex items-center gap-2.5">
        <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-teal-100 text-teal-800">
          <IconPencilLine className="h-4 w-4" />
        </div>
        <h2 className="text-lg sm:text-xl font-bold tracking-tight text-sand-900">
          Keep a record
        </h2>
      </div>
      <p className="mt-2 text-xs sm:text-sm leading-relaxed text-sand-600">{prompt}</p>

      <form onSubmit={handleSubmit} className="mt-5 space-y-4">
        <div className="grid gap-3.5 sm:grid-cols-2">
          <div>
            <label htmlFor="staffName" className="block text-xs sm:text-sm font-bold text-sand-800">
              Staff member&rsquo;s name
            </label>
            <input
              id="staffName"
              type="text"
              value={staffName}
              onChange={(e) => setStaffName(e.target.value)}
              className="field mt-1.5"
              placeholder="If you know it"
            />
          </div>
          <div>
            <label htmlFor="approxTime" className="block text-xs sm:text-sm font-bold text-sand-800">
              Approximate time
            </label>
            <input
              id="approxTime"
              type="text"
              value={approxTime}
              onChange={(e) => setApproxTime(e.target.value)}
              className="field mt-1.5"
              placeholder="e.g. around 3:30 PM"
            />
          </div>
        </div>

        <div>
          <label htmlFor="note" className="block text-xs sm:text-sm font-bold text-sand-800">
            Anything else worth noting
          </label>
          <textarea
            id="note"
            value={note}
            onChange={(e) => setNote(e.target.value)}
            rows={2}
            className="field mt-1.5 resize-none"
            placeholder="e.g. said the pre-auth server was down, or gave no written explanation"
          />
        </div>

        {error && (
          <p role="alert" className="text-xs sm:text-sm font-bold text-tier-red-text">
            {error}
          </p>
        )}

        <div>
          <button
            type="submit"
            disabled={isPending}
            className="tap-target inline-flex items-center gap-2 rounded-2xl bg-teal-700 px-7 py-3 text-sm font-bold text-white transition-all hover:bg-teal-800 active:scale-95 disabled:opacity-60 disabled:active:scale-100"
          >
            {isPending ? "Saving…" : "Save this"}
          </button>
        </div>
      </form>

      {evidence.length > 0 && (
        <div className="mt-6 pt-5 border-t border-sand-200">
          <p className="text-xs font-bold uppercase tracking-wider text-sand-500 mb-3">
            Logged Details
          </p>
          <ul className="space-y-2.5">
            {evidence.map((e, i) => (
              <li
                key={i}
                className="flex items-center gap-2.5 rounded-xl bg-sand-50 border border-sand-200/70 px-4 py-2.5 text-xs sm:text-sm text-sand-800 font-medium"
              >
                <span className="h-1.5 w-1.5 rounded-full bg-teal-600 shrink-0" aria-hidden="true" />
                <span>{[e.staff_name, e.approx_time, e.note].filter(Boolean).join(" — ")}</span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </section>
  );
}
