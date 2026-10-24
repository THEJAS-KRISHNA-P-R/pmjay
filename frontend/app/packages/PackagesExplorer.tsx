"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { searchPackages, ApiError } from "@/lib/api";
import type { PackageSearchResult, PackageSearchResponse } from "@/lib/types";
import { IconSearch, IconSpinner, IconAlertTriangle, IconPhone } from "@/app/components/icons";

const PAGE_SIZE = 25;
const DEBOUNCE_MS = 300;

/**
 * PackagesExplorer is deliberately client-side and deliberately dumb: it
 * renders exactly what GET /api/v1/packages returns, with no local
 * matching, no ranking, and no rate arithmetic. If the search endpoint is
 * down the page says so and points at 14555 rather than quietly showing
 * cached or invented data.
 */
export function PackagesExplorer() {
  const [input, setInput] = useState("");
  const [query, setQuery] = useState("");
  const [specialty, setSpecialty] = useState("");
  const [offset, setOffset] = useState(0);

  const [data, setData] = useState<PackageSearchResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Debounce the text box so a person typing "gallbladder" makes one
  // request, not eleven. The debounced value is the only thing that
  // triggers a fetch; `input` exists purely to keep the box responsive.
  useEffect(() => {
    const timer = setTimeout(() => {
      setQuery(input.trim());
      setOffset(0);
    }, DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [input]);

  const load = useCallback(
    (signal: AbortSignal) => {
      setLoading(true);
      setError(null);
      searchPackages({
        q: query || undefined,
        specialty: specialty || undefined,
        limit: PAGE_SIZE,
        offset,
      })
        .then((res) => {
          if (signal.aborted) return;
          setData(res);
        })
        .catch((err) => {
          if (signal.aborted) return;
          setError(
            err instanceof ApiError
              ? err.message
              : "Could not reach the package index. Check your connection and try again.",
          );
        })
        .finally(() => {
          if (signal.aborted) return;
          setLoading(false);
        });
    },
    [query, specialty, offset],
  );

  useEffect(() => {
    const controller = new AbortController();
    load(controller.signal);
    return () => controller.abort();
  }, [load]);

  const specialties = data?.specialties ?? [];

  const total = data?.total ?? 0;
  const shownFrom = total === 0 ? 0 : offset + 1;
  const shownTo = offset + (data?.count ?? 0);
  const hasPrev = offset > 0;
  const hasNext = shownTo < total;

  const coverageSummary = useMemo(() => {
    if (loading || !data) return null;
    return `${total.toLocaleString("en-IN")} ${total === 1 ? "package" : "packages"}`;
  }, [loading, data, total]);

  return (
    <section aria-labelledby="package-search-heading" className="space-y-5">
      <div className="card p-5 sm:p-6 space-y-4">
        <div className="space-y-1">
          <h2 id="package-search-heading" className="text-lg sm:text-xl font-bold tracking-tight text-sand-900">
            Search the package list
          </h2>
          <p className="text-xs sm:text-sm text-sand-600 font-medium">
            Plain English is fine. Try <em>dialysis</em>, <em>cataract</em>, or <em>gallbladder</em>.
          </p>
        </div>

        <div className="grid gap-3 sm:grid-cols-[1fr_auto]">
          <label className="relative block">
            <span className="sr-only">Search PM-JAY packages</span>
            <IconSearch className="pointer-events-none absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-sand-400" />
            <input
              type="search"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Procedure, package name, or HBP code"
              className="field pl-10"
              autoComplete="off"
            />
          </label>

          <label className="block">
            <span className="sr-only">Filter by specialty</span>
            <select
              value={specialty}
              onChange={(e) => {
                setSpecialty(e.target.value);
                setOffset(0);
              }}
              className="field sm:w-64"
            >
              <option value="">All specialties</option>
              {specialties.map((s) => (
                <option key={s} value={s}>
                  {s}
                </option>
              ))}
            </select>
          </label>
        </div>

        <p className="text-xs text-sand-500 font-medium" role="status" aria-live="polite">
          {loading
            ? "Searching…"
            : error
              ? "Search unavailable right now."
              : `${coverageSummary}${total > 0 ? ` · showing ${shownFrom}–${shownTo}` : ""}`}
        </p>
      </div>

      {loading && (
        <div className="card p-8 flex items-center justify-center gap-2.5 text-sm font-bold text-ink-800" role="status">
          <IconSpinner className="h-4 w-4 animate-spin text-ink-700" />
          <span>Searching the package list…</span>
        </div>
      )}

      {error && !loading && (
        <div className="rounded-2xl border-2 border-tier-red-border bg-tier-red-bg p-5 sm:p-6">
          <div className="flex items-start gap-3">
            <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-tier-red-icon text-tier-red-text">
              <IconAlertTriangle className="h-5 w-5" />
            </span>
            <div className="space-y-1.5">
              <p className="font-bold text-tier-red-text">{error}</p>
              <p className="text-sm text-tier-red-text/90 leading-relaxed">
                You do not need this page to get help. Call the PMJAY helpline at{" "}
                <a href="tel:14555" className="font-bold underline">
                  14555
                </a>{" "}
                and ask them to confirm the package and rate for your treatment.
              </p>
            </div>
          </div>
        </div>
      )}

      {!loading && !error && data && data.results.length === 0 && (
        <div className="card p-8 text-center space-y-3">
          <p className="text-base font-bold text-sand-900">Nothing matched that search.</p>
          <p className="text-sm text-sand-600 leading-relaxed max-w-lg mx-auto">
            That is not a denial and not proof that nothing is covered. The full PM-JAY master
            list has roughly 1,900 procedures and this index covers 315 of them. Ask the hospital
            for the exact procedure name and HBP code, then call{" "}
            <a href="tel:14555" className="font-bold text-emerald-800 underline">
              14555
            </a>
            .
          </p>
        </div>
      )}

      {!loading && !error && data && data.results.length > 0 && (
        <>
          <ul className="space-y-3">
            {data.results.map((pkg) => (
              <PackageRow
                key={pkg.package_code ?? `${pkg.package_name}-${pkg.specialty}`}
                pkg={pkg}
              />
            ))}
          </ul>

          {(hasPrev || hasNext) && (
            <div className="flex items-center justify-between gap-3 pt-1">
              <button
                type="button"
                onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
                disabled={!hasPrev}
                className="btn-secondary tap-target px-4 py-2 text-sm disabled:opacity-40 disabled:cursor-not-allowed"
              >
                Previous
              </button>
              <span className="text-xs font-bold text-sand-600">
                {shownFrom}–{shownTo} of {total.toLocaleString("en-IN")}
              </span>
              <button
                type="button"
                onClick={() => setOffset(offset + PAGE_SIZE)}
                disabled={!hasNext}
                className="btn-secondary tap-target px-4 py-2 text-sm disabled:opacity-40 disabled:cursor-not-allowed"
              >
                Next
              </button>
            </div>
          )}
        </>
      )}

      <div className="flex flex-wrap items-center gap-2 pt-1 text-xs font-bold">
        <span className="inline-flex items-center gap-1.5 rounded-full border border-emerald-200 bg-emerald-50 px-2.5 py-1 text-emerald-900">
          <span className="h-1.5 w-1.5 rounded-full bg-emerald-600" aria-hidden="true" />
          Verified code and rate
        </span>
        <span className="inline-flex items-center gap-1.5 rounded-full border border-sand-300 bg-sand-100 px-2.5 py-1 text-sand-700">
          <span className="h-1.5 w-1.5 rounded-full bg-sand-400" aria-hidden="true" />
          Name only — code and rate not verified
        </span>
      </div>
    </section>
  );
}

function PackageRow({ pkg }: { pkg: PackageSearchResult }) {
  return (
    <li className="card p-4 sm:p-5">
      <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3">
        <div className="min-w-0 space-y-1">
          <div className="flex flex-wrap items-center gap-2">
            <h3 className="text-sm sm:text-base font-bold text-sand-950 leading-snug">
              {pkg.package_name}
            </h3>
            {pkg.verified ? (
              <span className="inline-flex items-center rounded-full border border-emerald-200 bg-emerald-50 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider text-emerald-900">
                Verified
              </span>
            ) : (
              <span className="inline-flex items-center rounded-full border border-sand-300 bg-sand-100 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider text-sand-700">
                Name only
              </span>
            )}
          </div>
          <p className="text-xs text-sand-600 font-medium">
            {pkg.specialty}
            {pkg.package_code ? (
              <>
                {" · "}
                <span className="font-mono text-[11px] text-sand-700">{pkg.package_code}</span>
              </>
            ) : null}
            {pkg.requires_preauth ? " · pre-authorisation required" : ""}
          </p>
        </div>

        <div className="sm:text-right shrink-0">
          {pkg.verified ? <RateBlock pkg={pkg} /> : <UnverifiedNote />}
        </div>
      </div>
    </li>
  );
}

function RateBlock({ pkg }: { pkg: PackageSearchResult }) {
  const inr = (n: number) => `₹${n.toLocaleString("en-IN")}`;

  if (pkg.per_diem_rates) {
    const r = pkg.per_diem_rates;
    return (
      <div className="space-y-1">
        <p className="text-[10px] font-bold uppercase tracking-wider text-sand-500">Per day, by level of care</p>
        <dl className="grid grid-cols-2 gap-x-4 gap-y-0.5 text-xs">
          <dt className="text-sand-600">Ward</dt>
          <dd className="font-bold text-sand-900 text-right">{inr(r.routine_ward_inr)}</dd>
          <dt className="text-sand-600">HDU</dt>
          <dd className="font-bold text-sand-900 text-right">{inr(r.hdu_inr)}</dd>
          <dt className="text-sand-600">ICU</dt>
          <dd className="font-bold text-sand-900 text-right">{inr(r.icu_no_vent_inr)}</dd>
          <dt className="text-sand-600">ICU + vent</dt>
          <dd className="font-bold text-sand-900 text-right">{inr(r.icu_vent_inr)}</dd>
        </dl>
      </div>
    );
  }

  if (pkg.indicative_rate_inr) {
    return (
      <div className="space-y-0.5">
        <p className="text-[10px] font-bold uppercase tracking-wider text-sand-500">Indicative package rate</p>
        <p className="text-base sm:text-lg font-black text-sand-950">
          {inr(pkg.indicative_rate_inr)}
          {pkg.rate_max_inr && pkg.rate_max_inr > pkg.indicative_rate_inr ? (
            <span className="text-sand-600 font-bold"> – {inr(pkg.rate_max_inr)}</span>
          ) : null}
        </p>
      </div>
    );
  }

  return <UnverifiedNote />;
}

function UnverifiedNote() {
  return (
    <p className="max-w-[15rem] text-xs text-sand-600 leading-relaxed sm:text-right">
      No verified rate in this index.{" "}
      <a href="tel:14555" className="inline-flex items-center gap-1 font-bold text-emerald-800 underline">
        <IconPhone className="h-3 w-3" />
        14555
      </a>{" "}
      can confirm it.
    </p>
  );
}
