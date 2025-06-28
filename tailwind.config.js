/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/web/templates/*.html",
    "./internal/web/templates/**/*.html",
  ],
  theme: {
    extend: {
      colors: {
        'todo': '#3B82F6',
        'wip': '#F59E0B',
        'wait': '#6B7280',
        'sche': '#8B5CF6',
        'done': '#10B981',
      },
    },
  },
  plugins: [],
}