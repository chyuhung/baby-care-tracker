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
        'primary-deep': 'rgb(var(--primary-deep) / <alpha-value>)',
        'primary-light': 'rgb(var(--primary-light) / <alpha-value>)',
        diaper: 'rgb(var(--diaper) / <alpha-value>)',
        'diaper-deep': 'rgb(var(--diaper-deep) / <alpha-value>)',
        temperature: 'rgb(var(--temperature) / <alpha-value>)',
        'temperature-deep': 'rgb(var(--temperature-deep) / <alpha-value>)',
        sleep: 'rgb(var(--sleep) / <alpha-value>)',
        'sleep-deep': 'rgb(var(--sleep-deep) / <alpha-value>)',
        outdoor: 'rgb(var(--outdoor) / <alpha-value>)',
        'outdoor-deep': 'rgb(var(--outdoor-deep) / <alpha-value>)',
        // 暖调语义色（深档由 CSS 变量驱动，白字 ≥4.5:1）
        success: 'rgb(var(--success-deep) / <alpha-value>)',
        warning: 'rgb(var(--warning-deep) / <alpha-value>)',
        danger: 'rgb(var(--danger-deep) / <alpha-value>)',
        'danger-light': '#F9E9E6',
        muted: '#F1EDEB',
        'bg-main': '#FDF7F5',
        'bg-secondary': '#F9EFEB',
        'text-primary': '#3A3330',
        'text-secondary': '#766B66',
        'border-color': '#F3E7E3',
      },
      fontFamily: {
        sans: ['-apple-system', 'BlinkMacSystemFont', '"PingFang SC"', '"Hiragino Sans GB"', '"Microsoft YaHei"', 'sans-serif'],
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
