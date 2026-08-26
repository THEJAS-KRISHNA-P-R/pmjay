"use client";

import { useState } from "react";
import { setComplaintStatus } from "@/lib/caseHistory";
import type { ComplaintStatus } from "@/lib/types";
import { IconCheck } from "./icons";

const STEPS: { value: ComplaintStatus; label: string }[] = [
  { value: "draft", label: "Still drafting" },
  { value: "ready_to_submit", label: "Ready to submit" },
  { value: "submitted", label: "Submitted" },
  { value: "awaiting_response", label: "Awaiting a response" },
  { value: "resolved", label: "Resolved" },
];

export function ComplaintStatusTracker({
  caseId,
  initialStatus,
}: {
  caseId: string;
  initialStatus: ComplaintStatus;
}) {
  const [status, setStatus] = useState<ComplaintStatus>(
    initialStatus === "not_started" ? "draft" : initialStatus,
  );

  function handleSelect(next: ComplaintStatus) {
    setStatus(next);
    setComplaintStatus(caseId, next);
  }

  return (
    <section aria-labelledby="complaint-status-heading" className="card p-6 sm:p-8 animate-fade-in-up">
      <h2 id="complaint-status-heading" className="text-lg sm:text-xl font-bold tracking-tight text-sand-900">
        Track your complaint
      </h2>
      <p className="mt-2 text-xs sm:text-sm leading-relaxed text-sand-600 max-w-md font-medium">
        Nothing here submits or checks on your complaint for you — you still send the draft above yourself
        through the official Ayushman App. This is just a place to keep track of where things stand.
      </p>

      <div
        role="group"
        aria-label="Complaint status"
        className="mt-5 flex flex-wrap gap-2"
      >
        {STEPS.map((step) => {
          const active = step.value === status;
          return (
            <button
              key={step.value}
              type="button"
              aria-pressed={active}
              onClick={() => handleSelect(step.value)}
              className={`tap-target inline-flex items-center gap-1.5 rounded-xl px-4 py-2.5 text-xs sm:text-sm font-bold transition-all active:scale-95 ${
                active
                  ? "bg-emerald-700 text-white border border-emerald-800 shadow-[var(--shadow-clay-btn-primary)]"
                  : "btn-secondary text-sand-700 hover:text-sand-900"
              }`}
            >
              {active && <IconCheck className="h-3.5 w-3.5" />}
              <span>{step.label}</span>
            </button>
          );
        })}
      </div>
    </section>
  );
}
