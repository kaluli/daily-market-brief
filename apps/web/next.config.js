/** @type {import('next').NextConfig} */
const nextConfig = {
  // standalone solo para Docker; en Vercel causa problemas con .next en monorepos
  ...(process.env.VERCEL ? {} : { output: 'standalone' }),
  async rewrites() {
    const rewritesList = [];

    // Solo en local: /admin → admin de la API (3090). En producción no se añade.
    if (process.env.NODE_ENV === 'development') {
      rewritesList.push({ source: '/admin', destination: 'http://localhost:3090/admin' });
    }

    const backend =
      process.env.API_BACKEND_URL ||
      (process.env.VERCEL_URL ? null : 'http://localhost:3090');
    if (backend) {
      try {
        const backendHost = new URL(backend).hostname;
        if (process.env.VERCEL_URL && backendHost === process.env.VERCEL_URL) return rewritesList;
      } catch (_) {}
      rewritesList.push(
        { source: '/api/v1', destination: backend },
        { source: '/api/v1/:path*', destination: `${backend}/:path*` },
      );
    }
    return rewritesList;
  },
};
module.exports = nextConfig;
