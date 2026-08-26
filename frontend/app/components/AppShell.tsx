"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { usePathname } from "next/navigation";
import { hasCaseHistory } from "@/lib/caseHistory";
import { IconHome, IconPlus, IconSettings, IconPhone, IconChevronLeft } from "./icons";

const NAV_ITEMS = [
  { href: "/dashboard", label: "Dashboard", Icon: IconHome },
  { href: "/cases/new", label: "New Case", Icon: IconPlus },
  { href: "/settings", label: "Settings", Icon: IconSettings },
] as const;

function isActive(pathname: string, href: string): boolean {
  if (href === "/dashboard") return pathname === "/dashboard";
  return pathname === href || pathname.startsWith(`${href}/`);
}

export function AppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const [hasHistory, setHasHistory] = useState(false);

  useEffect(() => {
    setHasHistory(hasCaseHistory());
  }, [pathname]);

  return (
    <div className="min-h-screen flex flex-col bg-sand-50">
      {/* Top bar — permanently locked to the top with exact 60px height */}
      <header className="sticky top-0 z-50 w-full h-[60px] border-b border-sand-200 bg-sand-25/95 backdrop-blur-md backdrop-saturate-150 shadow-[0_1px_2px_rgba(42,38,33,0.04)] flex items-center">
        <div className="flex w-full items-center justify-between gap-4 px-4 sm:px-6 lg:px-8">
          <div className="flex items-center gap-3 min-w-0">
            <Link
              href="/"
              className="tap-target inline-flex items-center justify-center rounded-lg text-sand-500 transition-colors hover:bg-sand-100 hover:text-sand-800 shrink-0"
              aria-label="Back to the PMJAY Advocate site"
            >
              <IconChevronLeft className="h-4 w-4" />
            </Link>
            <Link href="/dashboard" className="group flex items-center gap-3 rounded-lg transition-opacity hover:opacity-85 min-w-0">
              <Image
                src="/logo.svg"
                alt="PMJAY Logo"
                width={32}
                height={32}
                className="h-8 w-8 object-contain shrink-0"
                priority
              />
              <span className="font-display text-base sm:text-lg font-semibold tracking-tight-display text-sand-900 truncate">
                PMJAY Advocate
              </span>
            </Link>
          </div>

          <a
            href="tel:14555"
            className="inline-flex items-center gap-1.5 sm:gap-2 rounded-xl px-3 py-1.5 sm:px-3.5 sm:py-2 text-xs sm:text-sm font-bold text-emerald-900 bg-emerald-50 border border-emerald-200/90 hover:bg-emerald-100/90 shadow-[inset_0_1px_1px_rgba(255,255,255,0.9),0_1px_2px_rgba(42,38,33,0.04)] active:scale-95 transition-all whitespace-nowrap shrink-0"
          >
            <IconPhone className="h-3.5 w-3.5 sm:h-4 sm:w-4 shrink-0 text-emerald-700" />
            <span className="hidden sm:inline">Helpline: 14555</span>
            <span className="sm:hidden font-extrabold">14555</span>
          </a>
        </div>
      </header>

      {/* Main app body: docked locked sidebar on the left, main content on the right */}
      <div className="flex-1 flex w-full items-start">
        {/* Left docked sidebar — 100% fixed with zero jitter or scroll movement */}
        <nav
          aria-label="Product"
          className="hidden md:flex md:flex-col md:justify-between md:w-60 lg:w-64 shrink-0 fixed top-[60px] left-0 bottom-0 px-4 py-6 border-r border-sand-200 bg-sand-25/90 backdrop-blur-md z-30 overflow-y-auto"
        >
          <div className="space-y-1.5">
            {NAV_ITEMS.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                data-active={isActive(pathname, item.href)}
                aria-current={isActive(pathname, item.href) ? "page" : undefined}
                className="app-nav-item"
              >
                <item.Icon className="h-4.5 w-4.5 shrink-0" />
                <span>{item.label}</span>
              </Link>
            ))}
          </div>

          <div className="space-y-3 pt-4 border-t border-sand-200/60">
            {!hasHistory && (
              <p className="px-2 text-xs leading-relaxed text-sand-500 font-medium">
                Cases you start are listed here, in this browser, since there&rsquo;s no login.
              </p>
            )}
            <Link href="/how-it-works" className="app-nav-item text-xs">
              <span>How this tool works</span>
            </Link>
          </div>
        </nav>

        {/* Workspace content area — padded left on desktop to account for fixed sidebar */}
        <main className="flex-1 min-w-0 md:pl-60 lg:pl-64 pb-24 md:pb-12">
          <div className="w-full max-w-6xl px-4 sm:px-8 lg:px-10 py-6 sm:py-8 animate-fade-in">
            {children}
          </div>
        </main>
      </div>

      {/* Mobile bottom tab navigation — elevated, standout Apple-style bar */}
      <nav
        aria-label="Product"
        className="md:hidden fixed bottom-0 inset-x-0 z-50 pb-[env(safe-area-inset-bottom)]"
      >
        <div className="relative mx-auto max-w-md">
          {/* Navigation items — these determine the bar height */}
          <div className="flex items-center justify-around gap-1 px-3 py-3 sm:py-4 relative z-10">
            {NAV_ITEMS.map((item) => {
              const active = isActive(pathname, item.href);
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  aria-current={active ? "page" : undefined}
                  className={`tap-target relative flex-1 flex flex-col items-center justify-center gap-1.5 py-2 px-2 rounded-2xl text-[11px] font-bold transition-all duration-200 ${
                    active
                      ? "text-emerald-700"
                      : "text-sand-500 hover:text-sand-800"
                  }`}
                >
                  <span className="flex h-10 w-10 items-center justify-center" aria-hidden="true">
                    <item.Icon className={`h-6.5 w-6.5 transition-colors ${
                      active ? "text-emerald-700" : "text-sand-500"
                    }`} />
                  </span>
                  <span className="hidden sm:block">{item.label}</span>
                </Link>
              );
            })}
          </div>

          {/* Clay/neo-morphic background bar — sits behind, matches content height */}
          <div className="absolute inset-0 -inset-y-0.5 -inset-x-1.5 mx-0 rounded-3xl border border-sand-200 bg-white/95 backdrop-blur-lg shadow-clay-card-hover" aria-hidden="true" />
        </div>
      </nav>
    </div>
  );
}
