"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { usePathname } from "next/navigation";
import { IconPhone, IconMenu, IconX } from "./icons";

const NAV_LINKS = [
  { href: "/how-it-works", label: "How It Works" },
  { href: "/guide", label: "Your Rights" },
  { href: "/faq", label: "FAQ" },
  { href: "/about", label: "About" },
];

export function Header() {
  const pathname = usePathname();
  const [menuOpen, setMenuOpen] = useState(false);

  useEffect(() => {
    setMenuOpen(false);
  }, [pathname]);

  useEffect(() => {
    if (!menuOpen) return;
    function handleKey(e: KeyboardEvent) {
      if (e.key === "Escape") setMenuOpen(false);
    }
    document.addEventListener("keydown", handleKey);
    return () => document.removeEventListener("keydown", handleKey);
  }, [menuOpen]);

  return (
    <header className="sticky top-0 z-50 w-full border-b border-sand-200 bg-sand-25/95 backdrop-blur-md shadow-[0_1px_2px_rgba(42,38,33,0.04)]">
      <div className="mx-auto flex w-full max-w-6xl items-center justify-between px-6 sm:px-8 lg:px-10 py-3.5 sm:py-4">
        {/* Brand logo & title — single line */}
        <Link
          href="/"
          className="group flex items-center gap-3 rounded-lg transition-opacity hover:opacity-85 shrink-0"
        >
          <Image
            src="/logo.svg"
            alt="PMJAY Logo"
            width={38}
            height={38}
            className="h-9 w-9 object-contain shrink-0"
            priority
          />
          <span className="font-display text-lg sm:text-xl font-semibold tracking-tight-display text-sand-900 whitespace-nowrap">
            PMJAY Advocate
          </span>
        </Link>

        {/* Desktop nav — single line */}
        <nav aria-label="Primary" className="hidden md:flex items-center gap-1 lg:gap-2 shrink-0">
          {NAV_LINKS.map((link) => {
            const active = pathname === link.href;
            return (
              <Link
                key={link.href}
                href={link.href}
                aria-current={active ? "page" : undefined}
                className={`whitespace-nowrap rounded-xl px-3.5 py-2 text-sm font-bold transition-colors ${
                  active
                    ? "bg-teal-50 text-teal-800"
                    : "text-sand-600 hover:bg-sand-100 hover:text-sand-900"
                }`}
              >
                {link.label}
              </Link>
            );
          })}
        </nav>

        {/* Helpline action & Mobile button — single line */}
        <div className="flex items-center gap-2 shrink-0">
          <a
            href="tel:14555"
            className="tap-target inline-flex items-center gap-2 rounded-xl bg-teal-50 px-3.5 sm:px-4 py-2 text-xs sm:text-sm font-bold text-teal-800 transition-colors hover:bg-teal-100 active:scale-95 whitespace-nowrap shrink-0"
          >
            <IconPhone className="h-4 w-4 shrink-0 text-teal-700" />
            <span className="whitespace-nowrap">Helpline: 14555</span>
          </a>

          <button
            type="button"
            onClick={() => setMenuOpen((v) => !v)}
            aria-expanded={menuOpen}
            aria-controls="mobile-nav"
            aria-label={menuOpen ? "Close menu" : "Open menu"}
            className="tap-target inline-flex items-center justify-center rounded-xl text-sand-700 hover:bg-sand-100 md:hidden"
          >
            {menuOpen ? <IconX className="h-5 w-5" /> : <IconMenu className="h-5 w-5" />}
          </button>
        </div>
      </div>

      {/* Mobile nav panel */}
      {menuOpen && (
        <div
          id="mobile-nav"
          className="md:hidden border-t border-sand-200 bg-sand-25 animate-fade-in"
        >
          <nav aria-label="Primary" className="mx-auto flex max-w-6xl flex-col gap-1 px-6 py-3">
            {NAV_LINKS.map((link) => {
              const active = pathname === link.href;
              return (
                <Link
                  key={link.href}
                  href={link.href}
                  aria-current={active ? "page" : undefined}
                  className={`tap-target flex items-center rounded-xl px-4 text-base font-bold transition-colors ${
                    active ? "bg-teal-50 text-teal-800" : "text-sand-700 hover:bg-sand-100"
                  }`}
                >
                  {link.label}
                </Link>
              );
            })}
          </nav>
        </div>
      )}
    </header>
  );
}
