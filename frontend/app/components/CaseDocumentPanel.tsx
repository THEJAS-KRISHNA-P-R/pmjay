import { caseDocumentUrl } from "@/lib/api";

/**
 * CaseDocumentPanel provides a direct link to the backend's compiled PDF
 * document for the case.
 */
export function CaseDocumentPanel({ caseId }: { caseId: string }) {
  return (
    <section className="glass-panel p-6 sm:p-8 animate-enter">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-5">
        <div className="flex items-start gap-4">
          <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-teal-100 text-teal-800 shadow-xs" aria-hidden="true">
            <svg
              className="h-5 w-5"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z" />
              <polyline points="14 2 14 8 20 8" />
              <line x1="16" y1="13" x2="8" y2="13" />
              <line x1="16" y1="17" x2="8" y2="17" />
              <line x1="10" y1="9" x2="8" y2="9" />
            </svg>
          </div>
          <div>
            <h2 className="text-lg sm:text-xl font-bold tracking-tight text-teal-950">
              Your case as one document
            </h2>
            <p className="mt-1 text-xs sm:text-sm leading-relaxed text-sand-900/75 max-w-md font-medium">
              Everything above, formatted as one PDF you can print, save, or hand to hospital
              staff — including the draft complaint and hospital script, if this case has them.
            </p>
          </div>
        </div>
        <a
          href={caseDocumentUrl(caseId)}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center justify-center gap-2 rounded-2xl bg-teal-800 px-6 py-3.5 text-sm font-bold text-white shadow-md transition touch-spring hover:bg-teal-900 hover:shadow-lg active:scale-95 focus-visible:outline-teal-600 self-start sm:self-auto shrink-0"
        >
          <span>Download PDF</span>
          <span aria-hidden="true">↓</span>
        </a>
      </div>
    </section>
  );
}
