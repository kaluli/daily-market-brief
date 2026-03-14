/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  async rewrites() {
    // En prod: /api/v1 se proxy a la API Go. En Vercel sin API_BACKEND_URL usamos la misma URL (evita DNS_HOSTNAME_RESOLVED_PRIVATE).
    const backend =
      process.env.API_BACKEND_URL ||
      (process.env.VERCEL_URL ? `https://${process.env.VERCEL_URL}` : null) ||
      'http://localhost:3090';
    return [
      { source: '/api/v1', destination: backend },
      { source: '/api/v1/:path*', destination: `${backend}/:path*` },
    ];
  },
};
module.exports = nextConfig;
