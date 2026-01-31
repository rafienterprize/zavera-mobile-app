/** @type {import('next').NextConfig} */
const nextConfig = {
  // Use standard Next.js build (not static export)
  // Cloudflare Pages supports Next.js SSR via @cloudflare/next-on-pages
  // For now, we'll use standard build which works fine on Cloudflare
  
  images: {
    // Keep unoptimized for better Cloudflare compatibility
    unoptimized: true,
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'images.unsplash.com',
      },
      {
        protocol: 'https',
        hostname: 'upload.wikimedia.org',
      },
      {
        protocol: 'https',
        hostname: 'api.sandbox.midtrans.com',
      },
      {
        protocol: 'https',
        hostname: 'api.midtrans.com',
      },
      {
        protocol: 'https',
        hostname: 'res.cloudinary.com',
      },
    ],
  },
};

module.exports = nextConfig;
