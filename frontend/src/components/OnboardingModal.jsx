import React, { useState } from 'react';

const STEP_DATA = {
  1: {
    icon: 'upload_file',
    badge: 'Awaiting Document Drop',
    badgeClass: 'text-google-blue dark:text-[#38bdf8] bg-google-blue-light dark:bg-[#1e2a44] border-google-blue/20 dark:border-[#38bdf8]/30',
    meta: 'Format: .PDF (Max 20MB)',
    title: 'Ready to index raw document',
    desc: 'Drop your project report, research paper, or system specification. We preserve structure, formulas, and technical tables automatically.',
    activeBg: 'bg-google-blue-light/50 dark:bg-[#1e2a44]/50',
    ring: 'ring-2 ring-google-blue/20 dark:ring-[#38bdf8]/30 border-google-blue/30 dark:border-[#38bdf8]/40',
    iconBg: 'bg-google-blue-light dark:bg-[#1e2a44] text-google-blue dark:text-[#38bdf8]',
    cardBadge: 'text-google-blue dark:text-[#38bdf8] bg-white dark:bg-[#151821] border border-google-blue/20 dark:border-[#38bdf8]/30',
    demo: (
      <div className="w-full md:w-auto flex flex-col sm:flex-row items-center gap-space-sm">
        <div className="w-full sm:w-64 bg-white dark:bg-[#151821] border border-google-gray-300 dark:border-[#282f40] p-space-sm rounded-lg flex items-center justify-between shadow-xs">
          <div className="flex items-center gap-space-xs">
            <span className="material-symbols-filled text-google-green dark:text-[#4ade80] text-[18px]">verified_user</span>
            <span className="font-label-md text-label-md text-google-gray-900 dark:text-[#f1f3f7] font-medium">End-to-End Encryption</span>
          </div>
          <span className="w-2 h-2 rounded-full bg-google-green dark:bg-[#4ade80]" />
        </div>
        <div className="px-space-md py-space-xs rounded-lg bg-google-blue-light dark:bg-[#1e2a44] text-google-blue dark:text-[#38bdf8] border border-google-blue/20 dark:border-[#38bdf8]/30 font-citation-ref text-citation-ref flex items-center gap-space-xs font-medium">
          <span className="material-symbols-outlined text-google-blue dark:text-[#38bdf8] text-[16px]">memory</span>
          <span>Zero Data Retention</span>
        </div>
      </div>
    ),
  },
  2: {
    icon: 'manage_search',
    badge: 'Neural Parse Active',
    badgeClass: 'text-google-red dark:text-[#f87171] bg-google-red-light dark:bg-[#321c1f] border-google-red/20 dark:border-[#f87171]/30',
    meta: 'Embedding Model: gemini-embedding-001',
    title: 'Chunking & high-dimensional indexing',
    desc: 'The document is tokenized into contextual chunks with overlapping semantics and PII redaction to eliminate hallucinations.',
    activeBg: 'bg-google-red-light/40 dark:bg-[#321c1f]/50',
    ring: 'ring-2 ring-google-red/20 dark:ring-[#f87171]/30 border-google-red/30 dark:border-[#f87171]/40',
    iconBg: 'bg-google-red-light dark:bg-[#321c1f] text-google-red dark:text-[#f87171]',
    cardBadge: 'text-google-red dark:text-[#f87171] bg-white dark:bg-[#151821] border border-google-red/20 dark:border-[#f87171]/30',
    demo: (
      <div className="w-full md:w-auto flex flex-col sm:flex-row items-center gap-space-sm">
        <div className="flex items-center gap-space-xs bg-white dark:bg-[#151821] border border-google-gray-300 dark:border-[#282f40] px-space-md py-space-sm rounded-lg shadow-xs">
          <span className="font-citation-ref text-citation-ref text-google-red dark:text-[#f87171] font-bold">1,536 Dim</span>
          <span className="text-google-gray-300 dark:text-[#4b5563]">/</span>
          <span className="font-citation-ref text-citation-ref text-google-gray-900 dark:text-[#f1f3f7] font-medium">Cosine Matrix Generated</span>
        </div>
        <div className="flex items-center gap-space-xs px-space-sm py-space-xs bg-google-red-light dark:bg-[#321c1f] border border-google-red/20 dark:border-[#f87171]/30 rounded-lg font-label-sm text-label-sm text-google-red dark:text-[#f87171] font-medium">
          <span className="material-symbols-outlined text-[16px] animate-spin">refresh</span>
          <span>Vector Chunks Mapped</span>
        </div>
      </div>
    ),
  },
  3: {
    icon: 'forum',
    badge: 'Prompt Ready',
    badgeClass: 'text-google-yellow-dark dark:text-[#fcd34d] bg-google-yellow-light dark:bg-[#2c220f] border-google-yellow-dark/20 dark:border-[#fcd34d]/30',
    meta: 'Query Latency: ~340ms',
    title: 'Query in natural human language',
    desc: 'Ask complex compound questions, request comparisons, or ask for exact engineering metrics without needing SQL or syntax.',
    activeBg: 'bg-google-yellow-light/50 dark:bg-[#2c220f]/50',
    ring: 'ring-2 ring-google-yellow-dark/20 dark:ring-[#fcd34d]/30 border-google-yellow-dark/30 dark:border-[#fcd34d]/40',
    iconBg: 'bg-google-yellow-light dark:bg-[#2c220f] text-google-yellow-dark dark:text-[#fcd34d]',
    cardBadge: 'text-google-yellow-dark dark:text-[#fcd34d] bg-white dark:bg-[#151821] border border-google-yellow-dark/20 dark:border-[#fcd34d]/30',
    demo: (
      <div className="w-full md:w-auto">
        <div className="bg-white dark:bg-[#151821] border border-google-gray-300 dark:border-[#282f40] px-space-md py-space-sm rounded-lg flex flex-col gap-space-2xs shadow-xs">
          <span className="font-label-sm text-label-sm text-google-gray-500 dark:text-[#9aa0a6]">Example User Query</span>
          <span className="font-body-sm text-body-sm text-google-gray-900 dark:text-[#f1f3f7] font-medium">
            "What sensor is used for distance measurement and what is its precision?"
          </span>
        </div>
      </div>
    ),
  },
  4: {
    icon: 'psychology',
    badge: 'Vector Synthesis',
    badgeClass: 'text-google-green dark:text-[#4ade80] bg-google-green-light dark:bg-[#122b1b] border-google-green/20 dark:border-[#4ade80]/30',
    meta: 'Top-k Retrieval: 5 Chunks',
    title: 'Hybrid vector & dense semantic matching',
    desc: 'The engine calculates 70% cosine similarity combined with 30% full-text keyword ranking to identify relevant evidence clusters.',
    activeBg: 'bg-google-green-light/40 dark:bg-[#122b1b]/50',
    ring: 'ring-2 ring-google-green/20 dark:ring-[#4ade80]/30 border-google-green/30 dark:border-[#4ade80]/40',
    iconBg: 'bg-google-green-light dark:bg-[#122b1b] text-google-green dark:text-[#4ade80]',
    cardBadge: 'text-google-green dark:text-[#4ade80] bg-white dark:bg-[#151821] border border-google-green/20 dark:border-[#4ade80]/30',
    demo: (
      <div className="w-full md:w-auto flex flex-col gap-space-xs bg-white dark:bg-[#151821] border border-google-gray-300 dark:border-[#282f40] p-space-sm rounded-lg shadow-xs">
        <div className="flex justify-between items-center text-google-green dark:text-[#4ade80] font-citation-ref text-citation-ref font-semibold gap-4">
          <span>Cosine Similarity</span>
          <span>0.9842 (98.4%)</span>
        </div>
        <div className="w-48 h-1.5 bg-google-gray-200 dark:bg-[#282f40] rounded-full overflow-hidden">
          <div className="h-full bg-google-green dark:bg-[#4ade80] w-[98%]" />
        </div>
      </div>
    ),
  },
  5: {
    icon: 'auto_awesome',
    badge: 'Verified Synthesis',
    badgeClass: 'text-google-blue dark:text-[#38bdf8] bg-google-blue-light dark:bg-[#1e2a44] border-google-blue/20 dark:border-[#38bdf8]/30',
    meta: 'Source Grounding: 100%',
    title: 'Verifiable answers with direct page jumps',
    desc: 'Answers are generated strictly from your document grounding, complete with interactive citations straight to the page number.',
    activeBg: 'bg-google-blue-light/50 dark:bg-[#1e2a44]/50',
    ring: 'ring-2 ring-google-blue/20 dark:ring-[#38bdf8]/30 border-google-blue/30 dark:border-[#38bdf8]/40',
    iconBg: 'bg-google-blue-light dark:bg-[#1e2a44] text-google-blue dark:text-[#38bdf8]',
    cardBadge: 'text-google-blue dark:text-[#38bdf8] bg-white dark:bg-[#151821] border border-google-blue/20 dark:border-[#38bdf8]/30',
    demo: (
      <div className="w-full md:w-auto flex items-center gap-space-sm bg-white dark:bg-[#151821] border border-google-gray-300 dark:border-[#282f40] px-space-md py-space-sm rounded-lg border-l-4 border-l-google-blue dark:border-l-[#38bdf8] shadow-xs">
        <div className="flex flex-col">
          <span className="font-citation-ref text-citation-ref text-google-blue dark:text-[#38bdf8] font-bold">Citation Found</span>
          <span className="font-body-sm text-body-sm text-google-gray-900 dark:text-[#f1f3f7] font-medium">
            HC-SR04 Ultrasonic Transducer [Page 3, Section 2.1]
          </span>
        </div>
        <span className="material-symbols-outlined text-google-blue dark:text-[#38bdf8] text-[18px]">open_in_new</span>
      </div>
    ),
  },
};

export default function OnboardingModal({ onComplete, onClose }) {
  const [activeStep, setActiveStep] = useState(1);
  const currentStep = STEP_DATA[activeStep];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-space-sm sm:p-space-md overflow-y-auto">
      {/* Dimmed Workspace Backdrop Overlay */}
      <div 
        onClick={onClose}
        className="fixed inset-0 bg-[#202124]/40 dark:bg-black/60 backdrop-blur-[2px] transition-opacity" 
      />

      {/* Centered Interactive Modal Container (Google Light Theme Card) */}
      <div className="relative z-10 w-full max-w-3xl max-h-[90vh] bg-white dark:bg-[#151821] rounded-2xl p-space-md sm:p-space-lg shadow-[0_8px_32px_rgba(32,33,36,0.16)] dark:shadow-2xl border border-google-gray-300 dark:border-[#242938] flex flex-col gap-space-md overflow-y-auto my-auto animate-in fade-in zoom-in-95 duration-200 transition-colors duration-200">
        {/* Micro Ambient Linear Highlight */}
        <div className="absolute top-0 inset-x-0 h-1 bg-gradient-to-r from-google-blue via-google-red to-google-yellow opacity-90" />

        {/* Modal Header */}
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-space-md">
          <div className="flex flex-col gap-space-xs">
            <div className="inline-flex items-center gap-space-xs self-start px-space-sm py-space-2xs rounded-full bg-google-blue-light dark:bg-[#1e2a44] text-google-blue dark:text-[#38bdf8]">
              <span className="w-1.5 h-1.5 rounded-full bg-google-blue dark:bg-[#38bdf8] animate-ping" />
              <span className="font-label-sm text-label-sm font-semibold tracking-wider uppercase">
                Quick Onboarding • 1 Min Overview
              </span>
            </div>
            <h1 className="font-headline-md text-headline-md text-google-gray-900 dark:text-[#f1f3f7] tracking-tight font-semibold">
              How it works
            </h1>
            <p className="font-body-sm text-body-sm text-google-gray-700 dark:text-[#9aa0a6]">
              Let's take a quick look before you get started with the neural document pipeline.
            </p>
          </div>
          <div className="flex items-center gap-space-xs">
            <div className="hidden sm:flex items-center gap-space-xs px-space-md py-space-xs rounded-lg bg-google-gray-100 dark:bg-[#1e2433] border border-google-gray-200 dark:border-[#282f40] text-google-gray-700 dark:text-[#cbd5e1] font-label-md text-label-md">
              <span className="material-symbols-outlined text-google-green dark:text-[#4ade80] text-[18px]">verified</span>
              <span className="font-medium text-google-gray-900 dark:text-[#f1f3f7]">RAG Certified</span>
            </div>
            {onClose && (
              <button
                type="button"
                onClick={onClose}
                className="p-1.5 text-google-gray-500 dark:text-[#9aa0a6] hover:text-google-gray-900 dark:hover:text-[#f1f3f7] hover:bg-google-gray-100 dark:hover:bg-[#222838] rounded-lg transition-colors cursor-pointer"
                title="Close guide"
              >
                <span className="material-symbols-outlined text-[20px]">close</span>
              </button>
            )}
          </div>
        </div>

        {/* Flow Pipeline Linear Navigation / Tracker Bar */}
        <div className="w-full bg-google-gray-100 dark:bg-[#1a1e2a] border border-google-gray-200 dark:border-[#282f40] rounded-lg p-space-xs flex items-center justify-between relative overflow-x-auto">
          {[1, 2, 3, 4, 5].map((stepNum, idx) => {
            const stepTitles = ['Upload', 'Understand', 'Ask', 'Find', 'Answer'];
            const isCompletedOrActive = stepNum <= activeStep;
            return (
              <React.Fragment key={stepNum}>
                <button
                  type="button"
                  onClick={() => setActiveStep(stepNum)}
                  className={`step-pill group flex items-center gap-space-xs px-space-md py-space-xs rounded-lg transition-all cursor-pointer ${
                    isCompletedOrActive ? 'bg-white dark:bg-[#242c3d] shadow-sm' : 'hover:bg-white dark:hover:bg-[#242c3d]'
                  }`}
                >
                  <span
                    className={`w-5 h-5 rounded-full flex items-center justify-center font-citation-ref text-citation-ref font-bold ${
                      isCompletedOrActive ? 'bg-google-blue dark:bg-[#38bdf8] text-white dark:text-[#0f172a]' : 'bg-google-gray-200 dark:bg-[#282f40] text-google-gray-500 dark:text-[#94a3b8]'
                    }`}
                  >
                    {stepNum}
                  </span>
                  <span
                    className={`font-label-md text-label-md font-semibold ${
                      isCompletedOrActive ? 'text-google-gray-900 dark:text-[#f1f3f7]' : 'text-google-gray-700 dark:text-[#9aa0a6]'
                    }`}
                  >
                    {stepTitles[idx]}
                  </span>
                </button>

                {stepNum < 5 && (
                  <div className="flex-1 flex items-center justify-center px-space-xs">
                    <div className="h-0.5 w-full bg-google-gray-200 dark:bg-[#282f40] relative overflow-hidden rounded-full">
                      <div
                        className={`h-full bg-google-blue dark:bg-[#38bdf8] transition-all duration-500 ${
                          stepNum < activeStep ? 'w-full' : 'w-0'
                        }`}
                      />
                    </div>
                  </div>
                )}
              </React.Fragment>
            );
          })}
        </div>

        {/* 5 Workflow Step Interactive Selector Grid */}
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-space-xs">
          {[
            { id: 1, name: '1. Upload your PDF', sub: 'Choose a text-based PDF and upload it securely.', tag: 'Drag & Drop / Browse', icon: 'upload_file' },
            { id: 2, name: '2. We understand', sub: 'Your document is parsed and prepared for deep questions.', tag: 'Chunking & Embedding', icon: 'manage_search' },
            { id: 3, name: '3. Ask query', sub: 'Ask anything related to your uploaded technical files.', tag: 'Natural Language', icon: 'forum' },
            { id: 4, name: '4. AI finds context', sub: 'Our AI searches your document and retrieves exact vectors.', tag: 'Hybrid RAG Search', icon: 'psychology' },
            { id: 5, name: '5. Get verified answer', sub: 'Receive a synthesized answer with interactive citations.', tag: 'Page Citations', icon: 'auto_awesome' },
          ].map((item) => {
            const isSelected = activeStep === item.id;
            const meta = STEP_DATA[item.id];
            return (
              <div
                key={item.id}
                onClick={() => setActiveStep(item.id)}
                className={`workflow-card cursor-pointer p-space-sm rounded-lg flex flex-col justify-between gap-space-xs shadow-sm transition-all duration-300 ${
                  isSelected
                    ? `${meta.activeBg} ${meta.ring}`
                    : 'bg-white dark:bg-[#1a1e2a] border border-google-gray-300 dark:border-[#282f40] hover:bg-google-gray-50 dark:hover:bg-[#222838]'
                }`}
              >
                <div className="flex items-start justify-between">
                  <div
                    className={`w-8 h-8 rounded-lg flex items-center justify-center shadow-xs ${
                      isSelected ? meta.iconBg : 'bg-google-gray-100 dark:bg-[#242c3d] text-google-gray-500 dark:text-[#9aa0a6]'
                    }`}
                  >
                    <span className="material-symbols-outlined text-[20px]">{item.icon}</span>
                  </div>
                  <span
                    className={`font-label-sm text-label-sm px-space-xs py-space-2xs rounded font-semibold ${
                      isSelected ? meta.cardBadge : 'bg-google-gray-100 dark:bg-[#242c3d] text-google-gray-500 dark:text-[#9aa0a6]'
                    }`}
                  >
                    Step 0{item.id}
                  </span>
                </div>
                <div className="flex flex-col gap-space-2xs">
                  <h2 className="font-label-md text-label-md text-google-gray-900 dark:text-[#f1f3f7] font-semibold leading-tight">{item.name}</h2>
                  <p className="font-citation-ref text-citation-ref text-google-gray-700 dark:text-[#9aa0a6] leading-snug hidden sm:block">{item.sub}</p>
                </div>
                <span className="self-start px-space-xs py-space-2xs rounded-full bg-white dark:bg-[#181c26] text-google-gray-700 dark:text-[#9aa0a6] border border-google-gray-200 dark:border-[#282f40] font-citation-ref text-citation-ref">
                  {item.tag}
                </span>
              </div>
            );
          })}
        </div>

        {/* Live Interactive Step Preview Visualizer Card */}
        <div className="w-full bg-google-gray-50 dark:bg-[#1a1e2a] border border-google-gray-300 dark:border-[#282f40] rounded-lg p-space-md flex flex-col md:flex-row items-center justify-between gap-space-md shadow-sm relative overflow-hidden transition-colors duration-200">
          {/* Dynamic Left Section */}
          <div className="flex items-start gap-space-md max-w-xl w-full">
            <div
              className={`w-12 h-12 rounded-xl flex items-center justify-center shrink-0 border border-current/20 shadow-xs ${currentStep.iconBg}`}
            >
              <span className="material-symbols-outlined text-[28px]">{currentStep.icon}</span>
            </div>
            <div className="flex flex-col gap-space-2xs">
              <div className="flex items-center gap-space-xs flex-wrap">
                <span className={`font-label-sm text-label-sm px-space-xs py-space-2xs rounded font-medium border ${currentStep.badgeClass}`}>
                  {currentStep.badge}
                </span>
                <span className="font-citation-ref text-citation-ref text-google-gray-500 dark:text-[#9aa0a6]">{currentStep.meta}</span>
              </div>
              <div className="font-title-sm text-title-sm text-google-gray-900 dark:text-[#f1f3f7] font-semibold">
                {currentStep.title}
              </div>
              <p className="font-body-sm text-body-sm text-google-gray-700 dark:text-[#9aa0a6] leading-relaxed">{currentStep.desc}</p>
            </div>
          </div>

          {/* Dynamic Right Demo */}
          {currentStep.demo}
        </div>

        {/* Bottom Action & Flow Transition Footer */}
        <div className="flex flex-col sm:flex-row items-center justify-between gap-space-sm pt-space-xs border-t border-google-gray-200 dark:border-[#282f40]">
          <div className="flex items-center gap-space-xs text-google-gray-700 dark:text-[#9aa0a6]">
            <span className="material-symbols-filled text-google-green dark:text-[#4ade80] text-[20px]">check_circle</span>
            <span className="font-label-md text-label-md">
              Step {activeStep} of 5 • Complete setup to unlock PDF workspace
            </span>
          </div>
          <div className="flex items-center gap-space-md w-full sm:w-auto">
            <button
              type="button"
              onClick={onComplete}
              className="w-full sm:w-auto flex items-center justify-center gap-space-sm px-space-xl py-space-md rounded-lg bg-google-blue text-white hover:bg-google-blue-hover active:bg-[#1557b0] transition-all duration-200 font-title-sm text-title-sm font-semibold shadow-sm group cursor-pointer"
            >
              <span>Start Exploring PDF_Intelligence</span>
              <span className="material-symbols-outlined group-hover:translate-x-1 transition-transform text-[20px]">
                arrow_forward
              </span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
