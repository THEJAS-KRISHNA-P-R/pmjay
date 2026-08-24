import { caseDocumentUrl } from "@/lib/api";
import { IconFileText, IconDownload } from "./icons";

/**
 * CaseDocumentPanel provides a direct link to the backend's compiled PDF
 * document for the case.
 */
export function CaseDocumentPanel({ caseId }: { caseId: string }) {
  return (
    <section className="card p-6 sm:p-8 animate-fade-in-up">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-5">
        <div className="flex items-start gap-4">
          <div
            className="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-teal-100 text-teal-800"
            aria-hidden="true"
          >
            <IconFileText className="h-5 w-5" />
          </div>
          <div>
            <h2 className="text-lg sm:text-xl font-bold tracking-tight text-sand-900">
              Your case as one document
            </h2>
            <p className="mt-1 text-xs sm:text-sm leading-relaxed text-sand-600 max-w-md font-medium">
              Everything above, formatted as one PDF you can print, save, or hand to hospital
              staff — including the draft complaint and hospital script, if this case has them.
            </p>
          </div>
        </div>
        <a
          href={caseDocumentUrl(caseId)}
          target="_blank"
          rel="noopener noreferrer"
          className="tap-target inline-flex items-center justify-center gap-2 rounded-2xl bg-teal-700 px-6 py-3.5 text-sm font-bold text-white shadow-sm transition-all hover:bg-teal-800 hover:shadow-md active:scale-95 self-start sm:self-auto shrink-0"
        >
          <span>Download PDF</span>
          <IconDownload className="h-4 w-4" />
        </a>
      </div>
    </section>
  );
}
