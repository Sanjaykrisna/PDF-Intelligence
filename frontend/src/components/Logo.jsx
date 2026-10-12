import React from 'react';

/**
 * Attractive, modern, futuristic brand logo for PDF_Intelligence.
 * Features an isometric folded neural document silhouette, radiant intelligence spark core,
 * ambient holographic gradient aura, and sleek typography.
 */
export default function Logo({ size = 'md', showText = true, showBadge = true, className = '' }) {
  const isSmall = size === 'sm';
  const isLarge = size === 'lg';

  const iconDims = isSmall
    ? 'w-8 h-8'
    : isLarge
    ? 'w-13 h-13'
    : 'w-10 h-10';

  return (
    <div className={`flex items-center gap-3 select-none group cursor-pointer ${className}`}>
      {/* Dynamic Logo Icon with ambient multi-gradient glow */}
      <div className={`relative ${iconDims} shrink-0 flex items-center justify-center`}>
        {/* Soft Ambient Radial Glow Backdrop */}
        <div className="absolute inset-0 rounded-2xl bg-gradient-to-tr from-[#1a73e8]/30 via-[#7c3aed]/20 to-[#06b6d4]/30 blur-md opacity-70 group-hover:opacity-100 group-hover:scale-110 transition-all duration-300" />

        {/* Icon Frame Container */}
        <div className="relative w-full h-full rounded-xl bg-gradient-to-br from-white/95 to-[#f0f4f9] dark:from-[#1b202c] dark:to-[#11141c] border border-[#d2d9e5] dark:border-[#2d3446] shadow-[0_2px_8px_rgba(26,115,232,0.12)] dark:shadow-[0_4px_16px_rgba(0,0,0,0.4)] flex items-center justify-center p-1.5 transition-transform duration-200 group-hover:scale-105">
          <svg
            className="w-full h-full drop-shadow-sm"
            viewBox="0 0 64 64"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <defs>
              {/* Dynamic Gradients */}
              <linearGradient id="pdfSheetGrad" x1="12" y1="8" x2="52" y2="56" gradientUnits="userSpaceOnUse">
                <stop offset="0%" stopColor="#1a73e8" />
                <stop offset="50%" stopColor="#4f46e5" />
                <stop offset="100%" stopColor="#7c3aed" />
              </linearGradient>

              <linearGradient id="pdfFoldGrad" x1="38" y1="8" x2="52" y2="22" gradientUnits="userSpaceOnUse">
                <stop offset="0%" stopColor="#60a5fa" />
                <stop offset="100%" stopColor="#1a73e8" />
              </linearGradient>

              <linearGradient id="sparkGrad" x1="22" y1="26" x2="42" y2="46" gradientUnits="userSpaceOnUse">
                <stop offset="0%" stopColor="#00f2fe" />
                <stop offset="50%" stopColor="#4facfe" />
                <stop offset="100%" stopColor="#00c6ff" />
              </linearGradient>

              <linearGradient id="amberCoreGrad" x1="32" y1="32" x2="38" y2="38" gradientUnits="userSpaceOnUse">
                <stop offset="0%" stopColor="#f59e0b" />
                <stop offset="100%" stopColor="#ef4444" />
              </linearGradient>

              <filter id="coreGlow" x="-20%" y="-20%" width="140%" height="140%">
                <feGaussianBlur stdDeviation="2" result="blur" />
                <feComposite in="SourceGraphic" in2="blur" operator="over" />
              </filter>
            </defs>

            {/* Back Document Sheet with subtle depth */}
            <path
              d="M18 10C15.7909 10 14 11.7909 14 14V50C14 52.2091 15.7909 54 18 54H46C48.2091 54 50 52.2091 50 50V22L38 10H18Z"
              fill="url(#pdfSheetGrad)"
              fillOpacity="0.12"
            />

            {/* Main Primary PDF Document Outline & Body */}
            <path
              d="M16 12C14.8954 12 14 12.8954 14 14V50C14 51.1046 14.8954 52 16 52H48C49.1046 52 50 51.1046 50 50V22L38 10H18C16.8954 10 16 10.8954 16 12Z"
              stroke="url(#pdfSheetGrad)"
              strokeWidth="2.5"
              strokeLinejoin="round"
            />

            {/* Folded Top-Right Corner */}
            <path
              d="M38 10V20C38 21.1046 38.8954 22 40 22H50"
              stroke="url(#pdfFoldGrad)"
              strokeWidth="2.5"
              strokeLinejoin="round"
              fill="url(#pdfFoldGrad)"
              fillOpacity="0.25"
            />

            {/* Semantic Data Lines (Indexed Document Chunks) */}
            <path
              d="M21 21H32M21 28H28"
              stroke="url(#pdfSheetGrad)"
              strokeWidth="2"
              strokeLinecap="round"
              strokeOpacity="0.6"
            />

            {/* Central Neural Synapse Connections */}
            <path
              d="M24 38L32 34M32 34L40 38M32 34V44"
              stroke="#06b6d4"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeDasharray="2 2"
              className="dark:opacity-80 opacity-60"
            />

            {/* Radiant 4-Point AI Intelligence Spark Core */}
            <path
              d="M32 26C32 30.5 30 32.5 25.5 32.5C30 32.5 32 34.5 32 39C32 34.5 34 32.5 38.5 32.5C34 32.5 32 30.5 32 26Z"
              fill="url(#sparkGrad)"
              filter="url(#coreGlow)"
              className="animate-pulse"
            />

            {/* Core Neural Spark Center Pinpoint */}
            <circle cx="32" cy="32.5" r="1.5" fill="#ffffff" />

            {/* Satellite Vector Nodes */}
            <circle cx="24" cy="38" r="1.8" fill="#10b981" />
            <circle cx="40" cy="38" r="1.8" fill="#3b82f6" />
            <circle cx="32" cy="44" r="1.8" fill="#8b5cf6" />
          </svg>
        </div>
      </div>

      {/* Brand Typography & Tagline */}
      {showText && (
        <div className="flex flex-col leading-none">
          <div className="flex items-center gap-1.5">
            <span className="font-extrabold tracking-tight text-[17px] text-[#1a73e8] dark:text-[#38bdf8] flex items-center">
              PDF
            </span>
            <span className="font-bold tracking-tight text-[17px] text-[#202124] dark:text-[#f1f3f7] flex items-center">
              Intelligence
            </span>
            {showBadge && (
              <span className="ml-1 px-1.5 py-0.5 rounded text-[9px] font-bold uppercase tracking-wider bg-gradient-to-r from-[#1a73e8]/10 to-[#7c3aed]/10 dark:from-[#38bdf8]/20 dark:to-[#818cf8]/20 text-[#1a73e8] dark:text-[#38bdf8] border border-[#1a73e8]/20 dark:border-[#38bdf8]/30">
                AI
              </span>
            )}
          </div>
          <span className="text-[10px] font-medium tracking-wider text-[#5f6368] dark:text-[#94a3b8] uppercase mt-0.5">
            Neural RAG Engine
          </span>
        </div>
      )}
    </div>
  );
}
