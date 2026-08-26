"use client";

import { useState, useEffect, useRef } from "react";

/**
 * Universal, high-performance scroll direction hook:
 * - Hides navbar on scroll DOWN by at least 8px
 * - Instantly pops navbar back into view on ANY scroll UP (>= 2px)
 * - Always keeps navbar visible when near the top of the page (<= 24px)
 * - Uses Math.max across all scroll properties for 100% cross-browser reliability
 */
export function useScrollDirection(downThreshold: number = 8, upThreshold: number = 2) {
  const [isVisible, setIsVisible] = useState(true);
  const lastScrollY = useRef(0);

  useEffect(() => {
    function getScrollPosition(): number {
      if (typeof window === "undefined") return 0;
      return Math.max(
        window.scrollY || 0,
        window.pageYOffset || 0,
        document.documentElement?.scrollTop || 0,
        document.body?.scrollTop || 0
      );
    }

    lastScrollY.current = getScrollPosition();

    function handleScroll() {
      const currentScrollY = getScrollPosition();

      // Always show when near the very top of the page
      if (currentScrollY <= 24) {
        setIsVisible(true);
        lastScrollY.current = Math.max(0, currentScrollY);
        return;
      }

      const diff = currentScrollY - lastScrollY.current;

      if (diff > downThreshold) {
        // Scrolling DOWN -> slide navbar away
        setIsVisible(false);
        lastScrollY.current = currentScrollY;
      } else if (diff < -upThreshold) {
        // Scrolling UP -> immediately pop navbar in
        setIsVisible(true);
        lastScrollY.current = currentScrollY;
      }
    }

    window.addEventListener("scroll", handleScroll, { passive: true });
    document.addEventListener("scroll", handleScroll, { passive: true });

    return () => {
      window.removeEventListener("scroll", handleScroll);
      document.removeEventListener("scroll", handleScroll);
    };
  }, [downThreshold, upThreshold]);

  return isVisible;
}
