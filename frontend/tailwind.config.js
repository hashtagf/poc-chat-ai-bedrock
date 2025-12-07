/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Chat UI custom theme colors
        'chat-bg': '#f9fafb',
        'chat-surface': '#ffffff',
        'chat-border': '#e5e7eb',
        'chat-user': '#3b82f6',
        'chat-agent': '#6b7280',
        'chat-accent': '#2563eb',
        'chat-error': '#ef4444',
        'chat-success': '#10b981',
      },
      spacing: {
        'chat-xs': '0.5rem',
        'chat-sm': '0.75rem',
        'chat-md': '1rem',
        'chat-lg': '1.5rem',
        'chat-xl': '2rem',
      },
      borderRadius: {
        'chat': '0.75rem',
      },
      boxShadow: {
        'chat': '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)',
        'chat-lg': '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
      },
    },
  },
  plugins: [],
}
