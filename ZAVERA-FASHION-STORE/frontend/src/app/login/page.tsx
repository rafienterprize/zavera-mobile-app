"use client";

import { useState, useEffect, Suspense, useCallback } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { motion } from "framer-motion";
import { useAuth } from "@/context/AuthContext";
import { useToast } from "@/components/ui/Toast";
import Script from "next/script";
import api from "@/lib/api";

// Admin email from env
const ADMIN_EMAIL = process.env.NEXT_PUBLIC_ADMIN_EMAIL || "pemberani073@gmail.com";

function LoginContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { login, loginWithGoogle, isAuthenticated, isLoading: authLoading, user } = useAuth();
  const { showToast } = useToast();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [googleLoaded, setGoogleLoaded] = useState(false);

  const redirectTo = searchParams.get("redirect") || "/";

  const handleGoogleCallback = useCallback(async (response: { credential: string }) => {
    try {
      setIsLoading(true);
      // Call Google login API directly to get user data
      const loginResponse = await api.post("/auth/google", { id_token: response.credential });
      const { access_token, user: userData } = loginResponse.data;
      localStorage.setItem("auth_token", access_token);
      
      showToast("Login berhasil!", "success");
      
      // Check if user is admin and redirect accordingly
      if (userData.email === ADMIN_EMAIL) {
        router.push("/admin/dashboard");
      } else {
        router.push(redirectTo);
      }
      
      // Refresh auth context
      window.location.reload();
    } catch (error: unknown) {
      const err = error as { response?: { data?: { message?: string } } };
      showToast(err.response?.data?.message || "Google login gagal", "error");
    } finally {
      setIsLoading(false);
    }
  }, [showToast, router, redirectTo]);

  useEffect(() => {
    if (!authLoading && isAuthenticated && user) {
      // If user is admin, redirect to admin dashboard
      if (user.email === ADMIN_EMAIL) {
        router.push("/admin/dashboard");
      } else {
        router.push(redirectTo);
      }
    }
  }, [isAuthenticated, authLoading, router, redirectTo, user]);

  useEffect(() => {
    if (googleLoaded && window.google) {
      window.google.accounts.id.initialize({
        client_id: process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID || "",
        callback: handleGoogleCallback,
      });

      const googleBtn = document.getElementById("google-signin-btn");
      if (googleBtn) {
        window.google.accounts.id.renderButton(googleBtn, {
          theme: "outline",
          size: "large",
          width: 400,
        });
      }
    }
  }, [googleLoaded, handleGoogleCallback]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      const userData = await login(email, password);
      showToast("Login berhasil!", "success");
      
      // Check if user is admin and redirect accordingly
      if (userData.email === ADMIN_EMAIL) {
        router.push("/admin/dashboard");
      } else {
        router.push(redirectTo);
      }
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string; message?: string } } };
      if (err.response?.data?.error === "email_not_verified") {
        showToast("Email belum diverifikasi. Silakan cek inbox Anda.", "warning");
      } else {
        showToast(err.response?.data?.message || "Login gagal", "error");
      }
    } finally {
      setIsLoading(false);
    }
  };

  if (authLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <>
      <Script
        src="https://accounts.google.com/gsi/client"
        onLoad={() => setGoogleLoaded(true)}
      />
      <div className="min-h-screen bg-gray-50 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="max-w-md w-full space-y-8"
        >
          {/* Header */}
          <div className="text-center">
            <Link href="/" className="inline-block">
              <h1 className="text-3xl font-serif font-bold tracking-[0.2em] text-gray-900">
                ZAVERA
              </h1>
            </Link>
            <h2 className="mt-6 text-2xl font-semibold text-gray-900">
              Masuk ke Akun Anda
            </h2>
            <p className="mt-2 text-sm text-gray-600">
              Belum punya akun?{" "}
              <Link href="/register" className="text-primary hover:underline font-medium">
                Daftar sekarang
              </Link>
            </p>
          </div>

          {/* Form */}
          <div className="bg-white py-8 px-6 shadow-lg rounded-xl">
            <form onSubmit={handleSubmit} className="space-y-6">
              {/* Email */}
              <div>
                <label htmlFor="email" className="block text-sm font-medium text-gray-700">
                  Email
                </label>
                <input
                  id="email"
                  type="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="mt-1 block w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
                  placeholder="nama@email.com"
                />
              </div>

              {/* Password */}
              <div>
                <label htmlFor="password" className="block text-sm font-medium text-gray-700">
                  Password
                </label>
                <div className="relative mt-1">
                  <input
                    id="password"
                    type={showPassword ? "text" : "password"}
                    required
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="block w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent transition-all pr-12"
                    placeholder="••••••••"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-700"
                  >
                    {showPassword ? (
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                      </svg>
                    ) : (
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                      </svg>
                    )}
                  </button>
                </div>
              </div>

              {/* Submit Button */}
              <button
                type="submit"
                disabled={isLoading}
                className="w-full flex justify-center py-3 px-4 border border-transparent rounded-lg shadow-sm text-sm font-medium text-white bg-primary hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary disabled:opacity-50 disabled:cursor-not-allowed transition-all"
              >
                {isLoading ? (
                  <svg className="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                ) : (
                  "Masuk"
                )}
              </button>
            </form>

            {/* Divider */}
            <div className="mt-6">
              <div className="relative">
                <div className="absolute inset-0 flex items-center">
                  <div className="w-full border-t border-gray-300" />
                </div>
                <div className="relative flex justify-center text-sm">
                  <span className="px-2 bg-white text-gray-500">atau masuk dengan</span>
                </div>
              </div>
            </div>

            {/* Google Sign In */}
            <div className="mt-6 flex justify-center">
              <div id="google-signin-btn"></div>
            </div>
          </div>
        </motion.div>
      </div>
    </>
  );
}

export default function LoginPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
      </div>
    }>
      <LoginContent />
    </Suspense>
  );
}
