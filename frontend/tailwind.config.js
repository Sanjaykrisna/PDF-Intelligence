/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        "google-blue": "#1a73e8",
        "google-blue-hover": "#1765cc",
        "google-blue-light": "#e8f0fe",
        "google-red": "#ea4335",
        "google-red-light": "#fce8e6",
        "google-yellow": "#fbbc04",
        "google-yellow-dark": "#f9ab00",
        "google-yellow-light": "#fef7e0",
        "google-green": "#34a853",
        "google-green-light": "#e6f4ea",
        "google-gray-900": "#202124",
        "google-gray-700": "#5f6368",
        "google-gray-500": "#80868b",
        "google-gray-300": "#dadce0",
        "google-gray-200": "#e8eaed",
        "google-gray-100": "#f1f3f4",
        "google-gray-50": "#f8f9fa",
        "text-primary": "#202124",
        "text-secondary": "#5f6368",
        "text-muted": "#80868b",
        "border-color": "#dadce0",
        "border-subtle": "#e8eaed",
        "bg-surface": "#ffffff",
        "bg-page": "#f8f9fa",
        // Semantic Stitch tokens
        "primary": "#1a73e8",
        "primary-container": "#1a73e8",
        "on-surface": "#202124",
        "outline-variant": "#dadce0",
        "outline": "#727785",
        "surface": "#ffffff",
        "surface-container-low": "#f8f9fa",
        "surface-container": "#f1f3f4",
        "surface-container-high": "#e8eaed",
        "surface-container-highest": "#dfe3e8",
        "error": "#d93025",
        "error-container": "#fce8e6",
        "secondary": "#1e8e3e",
        "secondary-container": "#e6f4ea",
      },
      borderRadius: {
        "DEFAULT": "0.25rem",
        "xs": "0.125rem",
        "lg": "0.5rem",
        "xl": "0.75rem",
        "2xl": "1rem",
        "full": "9999px"
      },
      spacing: {
        "space-2xs": "0.125rem",
        "space-xs": "0.25rem",
        "space-sm": "0.5rem",
        "space-md": "1rem",
        "space-lg": "1.5rem",
        "space-xl": "2rem",
        "space-2xl": "3rem",
        "margin-desktop": "2rem",
        "gutter-desktop": "1.5rem"
      },
      fontFamily: {
        "headline-lg": ["Geist", "sans-serif"],
        "headline-md": ["Geist", "sans-serif"],
        "title-sm": ["Inter", "sans-serif"],
        "body-lg": ["Inter", "sans-serif"],
        "body-md": ["Inter", "sans-serif"],
        "body-sm": ["Inter", "sans-serif"],
        "label-md": ["Inter", "sans-serif"],
        "label-sm": ["JetBrains Mono", "monospace"],
        "citation-ref": ["JetBrains Mono", "monospace"]
      },
      fontSize: {
        "body-md": ["14px", { "lineHeight": "20px", "letterSpacing": "0.01em", "fontWeight": "400" }],
        "citation-ref": ["10px", { "lineHeight": "13px", "letterSpacing": "0.05em", "fontWeight": "600" }],
        "label-sm": ["11px", { "lineHeight": "14px", "letterSpacing": "0.02em", "fontWeight": "500" }],
        "label-md": ["13px", { "lineHeight": "18px", "letterSpacing": "0.01em", "fontWeight": "500" }],
        "headline-md": ["22px", { "lineHeight": "30px", "letterSpacing": "-0.015em", "fontWeight": "600" }],
        "headline-lg": ["28px", { "lineHeight": "36px", "letterSpacing": "-0.02em", "fontWeight": "600" }],
        "title-sm": ["15px", { "lineHeight": "22px", "letterSpacing": "-0.01em", "fontWeight": "500" }],
        "body-sm": ["12px", { "lineHeight": "18px", "letterSpacing": "0em", "fontWeight": "400" }]
      },
      boxShadow: {
        "google-card": "0 1px 3px rgba(60,64,67,0.1), 0 4px 8px rgba(60,64,67,0.06)",
        "google-sm": "0 1px 2px 0 rgba(60, 64, 67, 0.3), 0 1px 3px 1px rgba(60, 64, 67, 0.15)",
        "google-btn": "0 1px 2px 0 rgba(60,64,67,0.3), 0 1px 3px 1px rgba(60,64,67,0.15)",
        "xs": "0 1px 2px 0 rgba(0, 0, 0, 0.05)"
      }
    }
  },
  plugins: [],
}
