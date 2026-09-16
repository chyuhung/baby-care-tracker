/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // 主题色（由 CSS 变量驱动，按宝宝性别切换；清透 iOS 风）
        primary: 'rgb(var(--primary) / <alpha-value>)',
        'primary-deep': 'rgb(var(--primary-deep) / <alpha-value>)',
        diaper: 'rgb(var(--diaper) / <alpha-value>)',
        'diaper-deep': 'rgb(var(--diaper-deep) / <alpha-value>)',
        temperature: 'rgb(var(--temperature) / <alpha-value>)',
        'temperature-deep': 'rgb(var(--temperature-deep) / <alpha-value>)',
        sleep: 'rgb(var(--sleep) / <alpha-value>)',
        'sleep-deep': 'rgb(var(--sleep-deep) / <alpha-value>)',
        outdoor: 'rgb(var(--outdoor) / <alpha-value>)',
        'outdoor-deep': 'rgb(var(--outdoor-deep) / <alpha-value>)',
        // 语义色（深档由 CSS 变量驱动，白字清晰）
        success: 'rgb(var(--success-deep) / <alpha-value>)',
        warning: 'rgb(var(--warning-deep) / <alpha-value>)',
        danger: 'rgb(var(--danger-deep) / <alpha-value>)',
        'danger-light': '#FFECEB',
        muted: '#F1F3F7',
        'bg-main': 'rgb(var(--bg-main) / <alpha-value>)',
        'bg-secondary': '#F1F3F7',
        'text-primary': '#1C1C1E',
        'text-secondary': '#7C838F',
        'border-color': '#E8EAF0',
      },
      fontFamily: {
        sans: ['-apple-system', 'BlinkMacSystemFont', '"PingFang SC"', '"Hiragino Sans GB"', '"Microsoft YaHei"', 'sans-serif'],
      },
      borderRadius: {
        'xl': '16px',
        '2xl': '20px',
      },
      boxShadow: {
        'card': '0 2px 12px rgba(24, 34, 58, 0.05)',
        'card-hover': '0 8px 22px rgba(24, 34, 58, 0.09)',
      },
    },
  },
  plugins: [],
}
