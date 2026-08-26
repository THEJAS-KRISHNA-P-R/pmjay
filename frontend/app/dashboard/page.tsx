"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { AppShell } from "@/app/components/AppShell";
import { CaseCard } from "@/app/components/CaseCard";
import { getCase } from "@/lib/api";
import { listCaseHistory, saveCaseToHistory, needsAttention } from "@/lib/caseHistory";
import type { LocalCaseRecord } from "@/lib/types";
import { IconPlus, IconClipboardList, IconAlertTriangle, IconCheck, IconPhone, IconScale } from "@/app/components/icons";

export default function DashboardPage() {
  const [cases, setCases] = useState<LocalCaseRecord[] | null>(null);
  const [filter, setFilter] = useState<"all" | "attention" | "resolved">("all");

  useEffect(() => {
    const local = listCaseHistory();
    setCases(local);

    // Best-effort refresh against the real backend so a case another
    // tab updated (e.g. new evidence added) doesn't look stale here.
    let cancelled = false;
    (async () => {
      for (const record of local) {
        try {
          const fresh = await getCase(record.id);
          if (cancelled) return;
          saveCaseToHistory(fresh);
        } catch {
          continue;
        }
      }
      if (!cancelled) setCases(listCaseHistory());
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  if (cases === null) {
    return (
      <AppShell>
        <div className="space-y-4" role="status" aria-label="Loading your cases">
          <div className="h-8 w-40 rounded-lg skeleton-shimmer" />
          <div className="h-28 rounded-2xl skeleton-shimmer" />
          <div className="h-28 rounded-2xl skeleton-shimmer" />
        </div>
      </AppShell>
    );
  }

  const attention = cases.filter(needsAttention);
  const resolved = cases.filter((c) => c.complaintStatus === "resolved");

  const displayedCases =
    filter === "attention"
      ? attention
      : filter === "resolved"
      ? resolved
      : cases;

  return (
    <AppShell>
      <div className="space-y-8 sm:space-y-10">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="space-y-1.5">
            <h1 className="font-display text-2xl sm:text-3xl font-semibold tracking-tight-display text-ink-950">
              Your cases
            </h1>
            <p className="text-sm sm:text-base text-sand-600 font-medium">
              What&rsquo;s happening, and what to do next.
            </p>
          </div>
          <Link href="/cases/new" className="btn-primary tap-target px-5 py-3 text-sm sm:text-base self-start sm:self-auto">
            <IconPlus className="h-4 w-4" />
            <span>Start a new case</span>
          </Link>
        </div>

        {cases.length === 0 ? (
          <EmptyDashboard />
        ) : (
          <>
            {/* Interactive Stat Filter Tiles */}
            <div className="grid grid-cols-3 gap-2.5 sm:gap-4" role="tablist" aria-label="Filter cases">
              <StatTile
                label="All Cases"
                value={cases.length}
                Icon={IconClipboardList}
                isActive={filter === "all"}
                onClick={() => setFilter("all")}
              />
              <StatTile
                label="Needs action"
                value={attention.length}
                Icon={IconAlertTriangle}
                accent={attention.length > 0}
                isActive={filter === "attention"}
                onClick={() => setFilter(filter === "attention" ? "all" : "attention")}
              />
              <StatTile
                label="Resolved"
                value={resolved.length}
                Icon={IconCheck}
                isActive={filter === "resolved"}
                onClick={() => setFilter(filter === "resolved" ? "all" : "resolved")}
              />
            </div>

            {/* Case List with Filter Status */}
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h2 className="text-lg sm:text-xl font-bold tracking-tight text-sand-900">
                  {filter === "attention"
                    ? `Cases needing attention (${attention.length})`
                    : filter === "resolved"
                    ? `Resolved cases (${resolved.length})`
                    : `All your cases (${cases.length})`}
                </h2>
                {filter !== "all" && (
                  <button
                    type="button"
                    onClick={() => setFilter("all")}
                    className="text-xs font-bold text-emerald-800 hover:text-emerald-950 underline tap-target"
                  >
                    Show all ({cases.length})
                  </button>
                )}
              </div>

              {displayedCases.length === 0 ? (
                <div className="card p-8 text-center space-y-2">
                  <p className="text-sm font-bold text-sand-800">
                    {filter === "attention"
                      ? "No cases currently need attention. All clear!"
                      : "No resolved cases yet."}
                  </p>
                  <button
                    type="button"
                    onClick={() => setFilter("all")}
                    className="text-xs font-bold text-emerald-800 hover:underline"
                  >
                    View all cases
                  </button>
                </div>
              ) : (
                <div className="space-y-3">
                  {displayedCases.map((record) => (
                    <CaseCard key={record.id} record={record} />
                  ))}
                </div>
              )}
            </div>

            {/* Emergency & Statutory Quick Access */}
            <div className="pt-4 border-t border-sand-200/70">
              <div className="card p-5 sm:p-6 bg-sand-25 border border-sand-200/80">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                  <div className="space-y-1">
                    <p className="text-xs font-bold uppercase tracking-wider text-sand-500">Need immediate help at the hospital?</p>
                    <p className="text-sm text-sand-800 font-medium">
                      Emergency treatment cannot be denied for want of payment or card clearance.
                    </p>
                  </div>
                  <div className="flex flex-wrap items-center gap-2.5">
                    <a
                      href="tel:14555"
                      className="btn-secondary tap-target px-3.5 py-2 text-xs text-emerald-900 bg-emerald-50 border-emerald-200 hover:bg-emerald-100"
                    >
                      <IconPhone className="h-3.5 w-3.5 text-emerald-700" />
                      <span>PMJAY: 14555</span>
                    </a>
                    <a
                      href="tel:15100"
                      className="btn-secondary tap-target px-3.5 py-2 text-xs"
                    >
                      <IconScale className="h-3.5 w-3.5 text-sand-700" />
                      <span>Legal Aid: 15100</span>
                    </a>
                    <a
                      href="tel:112"
                      className="btn-secondary tap-target px-3.5 py-2 text-xs text-tier-red-text bg-tier-red-bg border-tier-red-border hover:bg-tier-red-icon/40"
                    >
                      <IconPhone className="h-3.5 w-3.5 text-tier-red-text" />
                      <span>Emergency: 112</span>
                    </a>
                  </div>
                </div>
              </div>
            </div>

            <p className="text-xs leading-relaxed text-sand-500 font-medium max-w-md">
              This list lives in this browser only — there&rsquo;s no login, so a different device or a
              cleared browser won&rsquo;t show it. Save or print a case&rsquo;s PDF if you need it somewhere
              more permanent.
            </p>
          </>
        )}
      </div>
    </AppShell>
  );
}

function EmptyDashboard() {
  return (
    <div className="card flex flex-col items-center gap-4 px-6 py-14 sm:py-16 text-center">
      <span className="flex h-14 w-14 items-center justify-center rounded-2xl bg-sand-100 text-ink-700 shadow-xs" aria-hidden="true">
        <IconClipboardList className="h-6 w-6" />
      </span>
      <div className="space-y-1.5 max-w-sm">
        <p className="text-lg font-bold text-sand-900">Your cases will appear here.</p>
        <p className="text-sm text-sand-600 leading-relaxed font-medium">
          Describe what&rsquo;s happening at the hospital, in your own words, and we&rsquo;ll help you
          understand it — this becomes your first case.
        </p>
      </div>
      <Link href="/cases/new" className="btn-primary tap-target px-6 py-3.5 text-sm sm:text-base mt-2">
        <IconPlus className="h-4 w-4" />
        <span>Start a case</span>
      </Link>
    </div>
  );
}

function StatTile({
  label,
  value,
  Icon,
  accent = false,
  isActive = false,
  onClick,
}: {
  label: string;
  value: number;
  Icon: typeof IconClipboardList;
  accent?: boolean;
  isActive?: boolean;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      role="tab"
      aria-selected={isActive}
      className={`tile-soft p-3 sm:p-4 text-left cursor-pointer transition-all active:scale-[0.98] w-full ${
        isActive
          ? "ring-2 ring-emerald-700/80 border-emerald-300 bg-white shadow-[0_4px_14px_rgba(4,120,87,0.12),inset_0_1px_1px_rgba(255,255,255,1)]"
          : "hover:border-sand-300/80 bg-white"
      }`}
    >
      <div className="flex items-center gap-2 sm:gap-3">
        {/* Icon badge — left */}
        <div
          className={`flex h-9 w-9 sm:h-10 sm:w-10 shrink-0 items-center justify-center rounded-xl transition-colors ${
            accent
              ? "bg-tier-amber-icon text-tier-amber-text shadow-[inset_0_1px_1px_rgba(255,255,255,0.9)]"
              : isActive
              ? "bg-emerald-50 text-emerald-800 shadow-[inset_0_1px_1px_rgba(255,255,255,0.9)]"
              : "bg-sand-100 text-sand-700 shadow-[inset_0_1px_1px_rgba(255,255,255,0.8)]"
          }`}
          aria-hidden="true"
        >
          <Icon className="h-4.5 w-4.5 sm:h-5 sm:w-5" />
        </div>

        {/* Number + label — right */}
        <div className="min-w-0 flex-1">
          <p className="text-xl sm:text-2xl font-black tracking-tight text-sand-950 leading-none">{value}</p>
          <p className="text-[11px] sm:text-xs font-bold text-sand-600 leading-tight mt-1">{label}</p>
        </div>
      </div>
    </button>
  );
}
