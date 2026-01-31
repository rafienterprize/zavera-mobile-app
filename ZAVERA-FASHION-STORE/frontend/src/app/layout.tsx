import type { Metadata } from "next";
import { Inter, Playfair_Display } from "next/font/google";
import "./globals.css";
import { CartProvider } from "@/context/CartContext";
import { WishlistProvider } from "@/context/WishlistContext";
import { AuthProvider } from "@/context/AuthContext";
import { DialogProvider } from "@/context/DialogContext";
import { ToastProvider } from "@/components/ui/Toast";
import MidtransScript from "@/components/MidtransScript";
import LayoutWrapper from "@/components/LayoutWrapper";

const inter = Inter({ subsets: ["latin"], variable: "--font-inter" });
const playfair = Playfair_Display({ subsets: ["latin"], variable: "--font-playfair" });

export const metadata: Metadata = {
  title: "ZAVERA - Modern Online Fashion Store",
  description:
    "ZAVERA is a modern fashion e-commerce platform with secure online payments",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className={`${inter.variable} ${playfair.variable} font-sans`}>
        <MidtransScript />
        <DialogProvider>
          <ToastProvider>
            <AuthProvider>
              <CartProvider>
                <WishlistProvider>
                  <LayoutWrapper>{children}</LayoutWrapper>
                </WishlistProvider>
              </CartProvider>
            </AuthProvider>
          </ToastProvider>
        </DialogProvider>
      </body>
    </html>
  );
}
