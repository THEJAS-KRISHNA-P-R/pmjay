"use client";

import { useEffect, useState, useRef } from "react";
import Link from "next/link";
import Image from "next/image";
import { usePathname } from "next/navigation";
import { hasCaseHistory } from "@/lib/caseHistory";
import { useScrollDirection } from "@/lib/useScrollDirection";
import { IconPhone, IconMenu, IconX, IconClipboardList } from "./icons";

const NAV_LINKS = [
  { href: "/how-it-works", label: "How It Works" },
  { href: "/guide", label: "Your Rights" },
  { href: "/faq", label: "FAQ" },
  { href: "/about", label: "About" },
];

export function Header() {
  const pathname = usePathname();
  const [menuOpen, setMenuOpen] = useState(false);
  const [hasCases, setHasCases] = useState(false);
  const isScrollVisible = useScrollDirection();
  const headerRef = useRef<HTMLElement>(null);

  useEffect(() => {
    setMenuOpen(false);
  }, [pathname]);

  useEffect(() => {
    setHasCases(hasCaseHistory());
  }, [pathname]);

  // Auto-close on click outside, page scroll, or Escape key
  useEffect(() => {
    if (!menuOpen) return;

    function handlePointerDown(e: MouseEvent | TouchEvent) {
      if (headerRef.current && !headerRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    }

    function handleScroll() {
      setMenuOpen(false);
    }

    function handleKey(e: KeyboardEvent) {
      if (e.key === "Escape") setMenuOpen(false);
    }

    document.addEventListener("mousedown", handlePointerDown);
    document.addEventListener("touchstart", handlePointerDown);
    window.addEventListener("scroll", handleScroll, { passive: true });
    document.addEventListener("keydown", handleKey);

    return () => {
      document.removeEventListener("mousedown", handlePointerDown);
      document.removeEventListener("touchstart", handlePointerDown);
      window.removeEventListener("scroll", handleScroll);
      document.removeEventListener("keydown", handleKey);
    };
  }, [menuOpen]);

  return (
    <>
      <header
        ref={headerRef}
        className={`fixed top-0 inset-x-0 z-50 w-full border-b border-sand-200 bg-sand-25/90 backdrop-blur-md backdrop-saturate-150 shadow-[0_1px_2px_rgba(42,38,33,0.04)] transition-transform duration-300 ease-out will-change-transform ${
          isScrollVisible || menuOpen ? "translate-y-0" : "-translate-y-full"
        }`}
      >
        <div className="mx-auto flex w-full max-w-6xl items-center justify-between px-4 sm:px-8 lg:px-10 py-3 sm:py-4">
          {/* Brand logo & title */}
          <Link
            href="/"
            className="group flex items-center gap-2 sm:gap-3 rounded-lg transition-opacity hover:opacity-85 shrink-0"
          >
            <Image
              src="/logo.svg"
              alt="PMJAY Logo"
              width={36}
              height={36}
              className="h-8 w-8 sm:h-9 sm:w-9 object-contain shrink-0"
              priority
            />
            <span className="font-display text-base sm:text-lg lg:text-xl font-semibold tracking-tight-display text-sand-900 whitespace-nowrap">
              PMJAY Advocate
            </span>
          </Link>

          {/* Desktop nav */}
          <nav aria-label="Primary" className="hidden md:flex items-center gap-6 lg:gap-8 shrink-0">
            {NAV_LINKS.map((link) => {
              const active = pathname === link.href;
              return (
                <Link
                  key={link.href}
                  href={link.href}
                  aria-current={active ? "page" : undefined}
                  className={`inline-block py-1.5 px-0.5 text-base font-semibold tracking-tight transition-colors duration-150 border-b-2 whitespace-nowrap ${
                    active
                      ? "text-emerald-700 border-emerald-600"
                      : "text-sand-700 border-transparent hover:text-emerald-700 hover:border-emerald-600"
                  }`}
                >
                  {link.label}
                </Link>
              );
            })}
          </nav>

          {/* Helpline action & Mobile button */}
          <div className="flex items-center gap-1.5 sm:gap-2 shrink-0">
            {hasCases && (
              <Link
                href="/dashboard"
                className="btn-secondary hidden sm:inline-flex px-3 sm:px-3.5 py-1.5 sm:py-2 text-xs sm:text-sm"
              >
                <IconClipboardList className="h-4 w-4 shrink-0" />
                <span>My Cases</span>
              </Link>
            )}
            <a
              href="tel:14555"
              className="inline-flex items-center gap-1.5 sm:gap-2 rounded-xl px-3 py-1.5 sm:px-3.5 sm:py-2 text-xs sm:text-sm font-bold text-emerald-900 bg-emerald-50 border border-emerald-200/90 hover:bg-emerald-100/90 shadow-[inset_0_1px_1px_rgba(255,255,255,0.9),0_1px_2px_rgba(42,38,33,0.04)] active:scale-95 transition-all whitespace-nowrap shrink-0"
            >
              <IconPhone className="h-3.5 w-3.5 sm:h-4 sm:w-4 shrink-0 text-emerald-700" />
              <span className="hidden sm:inline whitespace-nowrap">Helpline: 14555</span>
              <span className="sm:hidden whitespace-nowrap font-extrabold">14555</span>
            </a>

            <button
              type="button"
              onClick={() => setMenuOpen((v) => !v)}
              aria-expanded={menuOpen}
              aria-controls="mobile-nav"
              aria-label={menuOpen ? "Close menu" : "Open menu"}
              className="tap-target inline-flex items-center justify-center rounded-xl text-sand-700 hover:bg-sand-100 md:hidden shrink-0"
            >
              {menuOpen ? <IconX className="h-5 w-5" /> : <IconMenu className="h-5 w-5" />}
            </button>
          </div>
        </div>

        {/* Mobile nav panel */}
        {menuOpen && (
          <div
            id="mobile-nav"
            className="md:hidden border-t border-sand-200 bg-sand-25 animate-fade-in shadow-lg"
          >
            <nav aria-label="Primary" className="mx-auto flex max-w-6xl flex-col gap-1 px-4 py-3">
              {hasCases && (
                <Link
                  href="/dashboard"
                  className="tap-target flex items-center gap-2.5 rounded-xl px-4 text-base font-bold text-ink-800 bg-ink-50 mb-1"
                >
                  <IconClipboardList className="h-4 w-4" />
                  <span>My Cases</span>
                </Link>
              )}
              {NAV_LINKS.map((link) => {
                const active = pathname === link.href;
                return (
                  <Link
                    key={link.href}
                    href={link.href}
                    aria-current={active ? "page" : undefined}
                    className={`tap-target flex items-center rounded-xl px-4 text-base font-semibold transition-colors ${
                      active ? "text-emerald-700 bg-emerald-50/60" : "text-sand-700 hover:bg-sand-100"
                    }`}
                  >
                    {link.label}
                  </Link>
                );
              })}
              <a
                href="tel:14555"
                className="tap-target flex items-center gap-2.5 rounded-xl px-4 text-base font-bold text-ink-800 bg-sand-100 mt-2"
              >
                <IconPhone className="h-4 w-4 text-ink-700" />
                <span>Helpline: 14555 (Toll-Free)</span>
              </a>
            </nav>
          </div>
        )}
      </header>

      {/* Static height compensation spacer for fixed header */}
      <div aria-hidden="true" className="h-16 sm:h-[70px] shrink-0" />

      {/* Backdrop overlay on mobile: click outside to dismiss */}
      {menuOpen && (
        <div
          onClick={() => setMenuOpen(false)}
          className="fixed inset-0 top-[57px] z-40 bg-sand-950/20 backdrop-blur-xs md:hidden animate-fade-in"
          aria-hidden="true"
        />
      )}
    </>
  );
}
