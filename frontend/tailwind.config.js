/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // 主题色（由 CSS 变量驱动，按宝宝性别切换）
        primary: 'rgb(var(--primary) / <alpha-value>)',
        'primary-light': 'rgb(var(--primary-light) / <alpha-value>)',
        diaper: 'rgb(var(--diaper) / <alpha-value>)',
        // 固定辅助色
        secondary: '#FFB6C1',
        accent: '#FFD93D',
        success: '#43C59E',
        warning: '#FF9800',
        'bg-main': '#F7F8FC',
        'bg-secondary': '#F1F3F9',
        'card-bg': '#FFFFFF',
        'text-primary': '#2D3436',
        'text-secondary': '#8A94A6',
        'border-color': '#EDEFF4',
      },
      fontFamily: {
        sans: ['"PingFang SC"', '"Hiragino Sans GB"', '"Microsoft YaHei"', 'sans-serif'],
        mono: ['"DIN Alternate"', '"Roboto Mono"', 'monospace'],
      },
      borderRadius: {
        'xl': '16px',
        '2xl': '20px',
        '3xl': '28px',
      },
      boxShadow: {
        'card': '0 2px 12px rgba(45, 52, 54, 0.06)',
        'card-hover': '0 6px 20px rgba(45, 52, 54, 0.10)',
        'float': '0 8px 24px rgba(45, 52, 54, 0.14)',
      },
    },
  },
  plugins: [],
}
