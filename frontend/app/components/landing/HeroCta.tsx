"use client";

import Link from "next/link";
import { IconArrowRight } from "../icons";

export function HeroCta() {
  function handleScrollToInput(e: React.MouseEvent) {
    e.preventDefault();
    const target = document.getElementById("check-coverage");
    if (target) {
      target.scrollIntoView({ behavior: "smooth", block: "start" });
    }
    const input = document.getElementById("description");
    if (input) {
      setTimeout(() => {
        input.focus();
      }, 400);
    }
  }

  return (
    <div className="flex flex-wrap items-center gap-3.5 pt-2">
      <button
        type="button"
        onClick={handleScrollToInput}
        className="btn-primary tap-target px-6 sm:px-8 py-3.5 text-sm sm:text-base"
      >
        <span>Check Your Situation Now</span>
        <span aria-hidden="true" className="text-lg">↓</span>
      </button>

      <Link
        href="/how-it-works"
        className="btn-secondary tap-target px-5 sm:px-6 py-3.5 text-sm sm:text-base font-bold text-sand-800 rounded-xl hover:text-sand-950 transition-colors"
      >
        <span>See How It Works</span>
        <IconArrowRight className="h-4 w-4" />
      </Link>
    </div>
  );
}
