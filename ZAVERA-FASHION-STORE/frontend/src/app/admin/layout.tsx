"use client";

import { useEffect, useState } from "react";
import { useRouter, usePathname } from "next/navigation";
import { useAuth } from "@/context/AuthContext";
import { NotificationBellSSE } from "@/components/admin/NotificationBellSSE";
import Link from "next/link";
import {
  LayoutDashboard,
  ShoppingBag,
  Package,
  Truck,
  RefreshCcw,
  AlertTriangle,
  FileText,
  LogOut,
  Menu,
  X,
  ChevronRight,
} from "lucide-react";

// Admin email - can be changed via env
const ADMIN_EMAIL = process.env.NEXT_PUBLIC_ADMIN_EMAIL || "pemberani073@gmail.com";

interface NavItem {
  name: string;
  href: string;
  icon: React.ReactNode;
  badge?: number;
}

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const { user, isLoading, isAuthenticated, logout } = useAuth();
  const router = useRouter();
  const pathname = usePathname();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [isAdmin, setIsAdmin] = useState(false);

  useEffect(() => {
    if (!isLoading) {
      if (!isAuthenticated) {
        router.push("/login?redirect=/admin/dashboard");
        return;
      }

      // Check if user is admin
      if (user?.email !== ADMIN_EMAIL) {
        router.push("/?error=unauthorized");
        return;
      }

      setIsAdmin(true);
    }
  }, [isLoading, isAuthenticated, user, router]);

  const navigation: NavItem[] = [
    { name: "Dashboard", href: "/admin/dashboard", icon: <LayoutDashboard size={20} /> },
    { name: "Orders", href: "/admin/orders", icon: <ShoppingBag size={20} /> },
    { name: "Products", href: "/admin/products", icon: <Package size={20} /> },
    { name: "Shipments", href: "/admin/shipments", icon: <Truck size={20} /> },
    { name: "Refunds", href: "/admin/refunds", icon: <RefreshCcw size={20} /> },
    { name: "Disputes", href: "/admin/disputes", icon: <AlertTriangle size={20} /> },
    { name: "Audit Log", href: "/admin/audit", icon: <FileText size={20} /> },
  ];

  if (isLoading || !isAdmin) {
    return (
      <div className="min-h-screen bg-neutral-950 flex items-center justify-center">
        <div className="text-center">
          <div className="w-12 h-12 border-2 border-white/20 border-t-white rounded-full animate-spin mx-auto mb-4" />
          <p className="text-white/60">Verifying admin access...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-neutral-950">
      {/* Mobile sidebar backdrop */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black/60 backdrop-blur-sm z-40 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside
        className={`fixed top-0 left-0 z-50 h-full w-72 bg-neutral-900 border-r border-white/10 transform transition-transform duration-300 ease-in-out lg:translate-x-0 ${
          sidebarOpen ? "translate-x-0" : "-translate-x-full"
        }`}
      >
        {/* Logo */}
        <div className="h-16 flex items-center justify-between px-6 border-b border-white/10">
          <Link href="/admin/dashboard" className="flex items-center gap-3">
            <div className="w-8 h-8 bg-white rounded-lg flex items-center justify-center">
              <span className="text-black font-bold text-sm">Z</span>
            </div>
            <span className="text-white font-semibold tracking-wide">ZAVERA ADMIN</span>
          </Link>
          <button
            onClick={() => setSidebarOpen(false)}
            className="lg:hidden text-white/60 hover:text-white"
          >
            <X size={20} />
          </button>
        </div>

        {/* Navigation */}
        <nav className="p-4 space-y-1">
          {navigation.map((item) => {
            const isActive = pathname === item.href;
            return (
              <Link
                key={item.name}
                href={item.href}
                className={`flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-200 group ${
                  isActive
                    ? "bg-white text-black"
                    : "text-white/60 hover:bg-white/5 hover:text-white"
                }`}
              >
                <span className={isActive ? "text-black" : "text-white/40 group-hover:text-white"}>
                  {item.icon}
                </span>
                <span className="font-medium">{item.name}</span>
                {item.badge && (
                  <span className="ml-auto bg-red-500 text-white text-xs px-2 py-0.5 rounded-full">
                    {item.badge}
                  </span>
                )}
                <ChevronRight
                  size={16}
                  className={`ml-auto transition-transform ${
                    isActive ? "text-black/40" : "text-white/20 group-hover:translate-x-1"
                  }`}
                />
              </Link>
            );
          })}
        </nav>

        {/* User info */}
        <div className="absolute bottom-0 left-0 right-0 p-4 border-t border-white/10">
          <div className="flex items-center gap-3 px-4 py-3 rounded-xl bg-white/5">
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center text-white font-semibold">
              {user?.first_name?.[0] || "A"}
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-white font-medium truncate">{user?.first_name || "Admin"}</p>
              <p className="text-white/40 text-sm truncate">{user?.email}</p>
            </div>
            <button
              onClick={logout}
              className="p-2 text-white/40 hover:text-red-400 transition-colors"
              title="Logout"
            >
              <LogOut size={18} />
            </button>
          </div>
        </div>
      </aside>

      {/* Main content */}
      <div className="lg:pl-72">
        {/* Top bar */}
        <header className="h-16 bg-neutral-900/50 backdrop-blur-xl border-b border-white/10 sticky top-0 z-30">
          <div className="h-full px-4 lg:px-8 flex items-center justify-between">
            <button
              onClick={() => setSidebarOpen(true)}
              className="lg:hidden p-2 text-white/60 hover:text-white"
            >
              <Menu size={24} />
            </button>

            <div className="flex items-center gap-2 text-sm">
              <span className="text-white/40">Admin Control Center</span>
              <ChevronRight size={14} className="text-white/20" />
              <span className="text-white font-medium">
                {navigation.find((n) => n.href === pathname)?.name || "Dashboard"}
              </span>
            </div>

            <div className="flex items-center gap-4">
              {/* Notification Bell */}
              <NotificationBellSSE />
              
              <div className="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-500/10 border border-emerald-500/20">
                <div className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
                <span className="text-emerald-400 text-sm font-medium">System Online</span>
              </div>
            </div>
          </div>
        </header>

        {/* Page content */}
        <main className="p-4 lg:p-8">{children}</main>
      </div>
    </div>
  );
}
