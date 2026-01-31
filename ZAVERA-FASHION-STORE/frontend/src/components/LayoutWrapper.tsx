"use client";

import { usePathname } from "next/navigation";
import Header from "@/components/Header";
import Footer from "@/components/Footer";

export default function LayoutWrapper({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  
  // Hide header/footer on admin pages, login/register pages, and checkout pages
  const isAdminPage = pathname?.startsWith("/admin");
  const isAuthPage = pathname === "/login" || pathname === "/register";
  const isCheckoutPage = pathname?.startsWith("/checkout");
  
  if (isAdminPage) {
    // Admin pages have their own layout
    return <>{children}</>;
  }
  
  // Checkout pages - no header/footer, just the content
  if (isCheckoutPage) {
    return <main className="min-h-screen">{children}</main>;
  }
  
  return (
    <>
      {!isAuthPage && <Header />}
      <main className="min-h-screen">{children}</main>
      {!isAuthPage && <Footer />}
    </>
  );
}
