"use client";

export default function GlobalError({
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <html lang="en">
      <body className="flex min-h-screen items-center justify-center bg-sand-50 p-6">
        <div className="max-w-md rounded-xl border border-tier-red-border bg-tier-red-bg p-6 text-center">
          <p className="font-bold text-tier-red-text">Something went wrong on this page.</p>
          <p className="mt-2 text-sm text-tier-red-text/90">
            If you&rsquo;re in the middle of something urgent, get treatment first — call the
            PMJAY helpline at <a href="tel:14555" className="underline">14555</a> any time.
          </p>
          <button
            onClick={reset}
            className="mt-4 rounded-lg bg-teal-700 px-4 py-2 font-bold text-white"
          >
            Try again
          </button>
        </div>
      </body>
    </html>
  );
}
