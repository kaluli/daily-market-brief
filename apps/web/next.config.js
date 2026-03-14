/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  async rewrites() {
    // Solo reescribir a un backend externo. En Vercel sin API_BACKEND_URL no reescribir (evita bucle infinito).
    const backend =
      process.env.API_BACKEND_URL ||
      (process.env.VERCEL_URL ? null : 'http://localhost:3090');
    if (!backend) return [];
    return [
      { source: '/api/v1', destination: backend },
      { source: '/api/v1/:path*', destination: `${backend}/:path*` },
    ];
  },
};
module.exports = nextConfig;
