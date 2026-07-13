/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: '#6C63FF',
        secondary: '#FFB6C1',
        accent: '#FFD93D',
        success: '#4CAF50',
        warning: '#FF9800',
        'bg-main': '#FAFAFA',
        'card-bg': '#FFFFFF',
        'text-primary': '#2D3436',
        'text-secondary': '#636E72',
        'border-color': '#E0E0E0',
      },
      fontFamily: {
        sans: ['"PingFang SC"', '"Hiragino Sans GB"', '"Microsoft YaHei"', 'sans-serif'],
        mono: ['"DIN Alternate"', '"Roboto Mono"', 'monospace'],
      },
      borderRadius: {
        'xl': '16px',
        '2xl': '20px',
      },
      boxShadow: {
        'card': '0 2px 8px rgba(0, 0, 0, 0.06)',
        'card-hover': '0 4px 16px rgba(0, 0, 0, 0.1)',
      },
    },
  },
  plugins: [],
}
