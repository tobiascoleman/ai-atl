/** @type {import('next').NextConfig} */
const nextConfig = {
  images: {
    domains: ['a.espncdn.com'], // For NFL team logos
  },
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  },
}

module.exports = nextConfig

