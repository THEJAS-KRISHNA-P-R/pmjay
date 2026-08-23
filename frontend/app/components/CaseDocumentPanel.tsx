import { caseDocumentUrl } from "@/lib/api";

/**
 * CaseDocumentPanel is the one clearly-actionable "take this with you"
 * artifact on the case result page: a link to the same case content
 * (backend/internal/document's BuildCase), formatted as one downloadable,
 * printable PDF instead of the individual on-screen sections above and
 * below it. See internal/document/README.md for why this exists — in
 * short, "copy this text" (CopyableTextBox) still leaves a family
 * needing somewhere to paste it before they can hand it to anyone;
 * this gives them something to hand over directly.
 *
 * A plain anchor tag with target="_blank", not a click handler that
 * fetches and saves a blob — letting the browser navigate to the PDF
 * directly means its native viewer (present on every modern desktop and
 * mobile browser) provides print/save/share controls for free, and
 * degrades gracefully on however this ends up being opened (in-app
 * browser, older Android WebView, and so on) in a way a hand-rolled
 * download flow wouldn't reliably match.
 */
export function CaseDocumentPanel({ caseId }: { caseId: string }) {
  return (
    <section className="rounded-xl border border-sand-200 bg-white p-5 sm:p-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h2 className="text-lg font-bold text-teal-800">Your case as one document</h2>
          <p className="mt-1 text-sm text-sand-900/70">
            Everything above, formatted as one PDF you can print, save, or hand to hospital
            staff — including the draft complaint and hospital script, if this case has them.
          </p>
        </div>
        <a
          href={caseDocumentUrl(caseId)}
          target="_blank"
          rel="noopener noreferrer"
          className="shrink-0 rounded-lg border-2 border-teal-700 bg-teal-700 px-4 py-2 text-sm font-bold text-white transition hover:bg-teal-800 active:bg-teal-900"
        >
          Download PDF
        </a>
      </div>
    </section>
  );
}
