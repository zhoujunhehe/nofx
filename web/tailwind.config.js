/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Figma 设计色卡
        vergex: {
          // 主色 - 紫色系
          primary: '#998cff',
          'primary-light': '#a397ff',

          // 背景色系
          'bg-primary': '#080808',
          'bg-secondary': 'rgba(255, 255, 255, 0.04)',
          'bg-tertiary': 'rgba(255, 255, 255, 0.06)',

          // 边框色系
          border: '#353535',
          'border-light': 'rgba(255, 255, 255, 0.1)',

          // 文字色系
          'text-primary': '#ffffff',
          'text-secondary': 'rgba(255, 255, 255, 0.6)',
          'text-tertiary': 'rgba(255, 255, 255, 0.4)',
          'text-muted': 'rgba(255, 255, 255, 0.1)',
        },
      },
      backgroundColor: {
        'vergex-gradient-start': 'rgba(255, 255, 255, 0.06)',
        'vergex-gradient-end': 'rgba(153, 153, 153, 0.06)',
      },
      backdropBlur: {
        'vergex': '20px',
      },
      fontFamily: {
        'vergex-heading': ['Poppins', 'sans-serif'],
        'vergex-body': ['Red Hat Text', 'Noto Sans SC', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
