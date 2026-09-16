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
        temperature: 'rgb(var(--temperature) / <alpha-value>)',
        sleep: 'rgb(var(--sleep) / <alpha-value>)',
        outdoor: 'rgb(var(--outdoor) / <alpha-value>)',
        // 固定辅助色
        secondary: '#E7B7C2',
        success: '#43A585',
        // 暖调语义色（与高级主题统一，保证对比度）
        warning: '#D27836',
        danger: '#D5574B',
        'danger-light': '#F9E9E6',
        muted: '#F1EDEB',
        'bg-main': '#FDF7F5',
        'bg-secondary': '#F9EFEB',
        'text-primary': '#3A3330',
        'text-secondary': '#766B66',
        'border-color': '#F3E7E3',
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
        'card': '0 2px 12px rgba(110, 82, 72, 0.06)',
        'card-hover': '0 6px 20px rgba(110, 82, 72, 0.10)',
        'float': '0 8px 24px rgba(110, 82, 72, 0.14)',
      },
    },
  },
  plugins: [],
}
