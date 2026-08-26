import Link from "next/link";
import type { LocalCaseRecord } from "@/lib/types";
import { formatRelativeTime, snippet, COMPLAINT_STATUS_LABEL } from "@/lib/format";
import { TIER_STYLES } from "./TierBadge";
import { IconChevronRight } from "./icons";

export function CaseCard({ record }: { record: LocalCaseRecord }) {
  const style = TIER_STYLES[record.outcome];
  const { Icon } = style;
  const showComplaintPill = record.complaintStatus !== "not_started";

  return (
    <Link
      href={`/cases/${record.id}`}
      className="card card-interactive group flex items-center gap-4 p-4 sm:p-5"
    >
      <span
        aria-hidden="true"
        className={`flex h-10 w-10 sm:h-11 sm:w-11 shrink-0 items-center justify-center rounded-2xl ${style.iconBg} ${style.iconText}`}
      >
        <Icon className="h-5 w-5" />
      </span>

      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2 flex-wrap">
          <p className={`text-sm font-bold ${style.iconText}`}>{style.label}</p>
          {showComplaintPill && (
            <span className="badge bg-sand-100 border-sand-200 text-sand-700 normal-case tracking-normal font-bold text-[10px]">
              {COMPLAINT_STATUS_LABEL[record.complaintStatus]}
            </span>
          )}
        </div>
        <p className="mt-1 text-sm text-sand-700 leading-snug line-clamp-2 font-medium">
          {record.description ? snippet(record.description, 160) : record.tierMessage}
        </p>
        <p className="mt-1.5 text-xs text-sand-500 font-bold">
          Last viewed {formatRelativeTime(record.lastViewedAt)}
        </p>
      </div>

      <IconChevronRight className="h-4 w-4 shrink-0 text-sand-400 transition-transform group-hover:translate-x-0.5 group-hover:text-ink-600" />
    </Link>
  );
}
