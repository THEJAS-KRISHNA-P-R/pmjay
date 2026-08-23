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
    <section className="rounded-xl border border-sand-200 bg-white p-5 sm:p-6">
      <h2 className="text-lg font-bold text-teal-800">Keep a record</h2>
      <p className="mt-1 text-sm text-sand-900/70">{prompt}</p>

      <form onSubmit={handleSubmit} className="mt-4 space-y-3">
        <div className="grid gap-3 sm:grid-cols-2">
          <div>
            <label htmlFor="staffName" className="block text-sm font-bold text-sand-900">
              Staff member&rsquo;s name
            </label>
            <input
              id="staffName"
              type="text"
              value={staffName}
              onChange={(e) => setStaffName(e.target.value)}
              className="mt-1 w-full rounded-lg border border-sand-200 px-3 py-2 text-sand-900"
              placeholder="If you know it"
            />
          </div>
          <div>
            <label htmlFor="approxTime" className="block text-sm font-bold text-sand-900">
              Approximate time
            </label>
            <input
              id="approxTime"
              type="text"
              value={approxTime}
              onChange={(e) => setApproxTime(e.target.value)}
              className="mt-1 w-full rounded-lg border border-sand-200 px-3 py-2 text-sand-900"
              placeholder="e.g. around 3:30 PM"
            />
          </div>
        </div>
        <div>
          <label htmlFor="note" className="block text-sm font-bold text-sand-900">
            Anything else worth noting
          </label>
          <textarea
            id="note"
            value={note}
            onChange={(e) => setNote(e.target.value)}
            rows={2}
            className="mt-1 w-full rounded-lg border border-sand-200 px-3 py-2 text-sand-900"
          />
        </div>

        {error && (
          <p role="alert" className="text-sm font-bold text-tier-red-strong">
            {error}
          </p>
        )}

        <button
          type="submit"
          disabled={isPending}
          className="rounded-lg bg-teal-700 px-5 py-2.5 font-bold text-white transition hover:bg-teal-800 disabled:opacity-60"
        >
          {isPending ? "Saving…" : "Save this"}
        </button>
      </form>

      {evidence.length > 0 && (
        <ul className="mt-5 space-y-2 border-t border-sand-200 pt-4">
          {evidence.map((e, i) => (
            <li key={i} className="text-sm text-sand-900/80">
              {[e.staff_name, e.approx_time, e.note].filter(Boolean).join(" — ")}
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
