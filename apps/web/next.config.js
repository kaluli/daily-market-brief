/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  async rewrites() {
    const backend =
      process.env.API_BACKEND_URL ||
      (process.env.VERCEL_URL ? null : 'http://localhost:3090');
    if (!backend) return [];
    // Nunca reescribir al mismo despliegue (evita infinite loop en Vercel).
    try {
      const backendHost = new URL(backend).hostname;
      if (process.env.VERCEL_URL && backendHost === process.env.VERCEL_URL) return [];
    } catch (_) {}
    return [
      { source: '/api/v1', destination: backend },
      { source: '/api/v1/:path*', destination: `${backend}/:path*` },
    ];
  },
};
module.exports = nextConfig;
