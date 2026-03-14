/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  async rewrites() {
    // En prod: todo lo que pida bajo /api/v1 se proxy a la API Go (buena práctica: versionado de API).
    // Admin: https://tudominio.com/api/v1/admin
    const backend =
      process.env.API_BACKEND_URL || 'http://localhost:3090';
    return [
      { source: '/api/v1', destination: backend },
      { source: '/api/v1/:path*', destination: `${backend}/:path*` },
    ];
  },
};
module.exports = nextConfig;
