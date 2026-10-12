import React from 'react';
import { useTheme } from '../context/ThemeContext';

export default function ThemeToggle({ className = '' }) {
  const { isDark, toggleTheme } = useTheme();

  return (
    <button
      type="button"
      onClick={toggleTheme}
      className={`relative p-2 rounded-xl border transition-all duration-200 cursor-pointer flex items-center justify-center group ${
        isDark
          ? 'bg-[#1b202c] border-[#2e3547] text-[#fbbf24] hover:bg-[#252c3d] shadow-[0_0_12px_rgba(251,191,36,0.15)]'
          : 'bg-[#f8f9fa] border-[#dadce0] text-[#5f6368] hover:text-[#1a73e8] hover:bg-[#f1f3f4] shadow-xs'
      } ${className}`}
      title={isDark ? 'Switch to Light Mode (Ctrl+D)' : 'Switch to Dark Mode (Ctrl+D)'}
      aria-label="Toggle theme"
    >
      <div className="relative w-5 h-5 flex items-center justify-center">
        {isDark ? (
          <span className="material-symbols-outlined text-[20px] text-[#fbbf24] rotate-0 transition-transform duration-300 group-hover:rotate-45">
            light_mode
          </span>
        ) : (
          <span className="material-symbols-outlined text-[20px] text-[#5f6368] group-hover:text-[#1a73e8] rotate-0 transition-transform duration-300 group-hover:-rotate-12">
            dark_mode
          </span>
        )}
      </div>
      <span className="sr-only">Toggle theme</span>
    </button>
  );
}
