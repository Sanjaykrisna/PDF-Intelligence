import React, { useState, useEffect } from 'react';
import eventService from '../services/eventService';

export default function SummaryModal({ document, onClose }) {
  if (!document) return null;

  const docId = document.id || document.document_id;
  const [summaryData, setSummaryData] = useState(document.summary || {});
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!docId) return;

    if (document.summary && document.summary.objective) {
      setSummaryData(document.summary);
      setLoading(false);
      return;
    }

    setLoading(true);
    eventService.getDocumentSummary(docId)
      .then((res) => {
        if (res && res.document && res.document.summary) {
          setSummaryData(res.document.summary);
        } else {
          setSummaryData({});
        }
      })
      .catch((err) => {
        console.error('Failed to fetch summary:', err);
        setSummaryData({});
      })
      .finally(() => setLoading(false));
  }, [docId]);

  const summary = summaryData || {};

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-space-md overflow-y-auto">
      <div 
        onClick={onClose} 
        className="fixed inset-0 bg-[#202124]/40 dark:bg-black/60 backdrop-blur-[2px] transition-opacity" 
      />

      <div className="relative z-10 w-full max-w-3xl bg-white dark:bg-[#151821] rounded-2xl shadow-[0_8px_32px_rgba(32,33,36,0.16)] dark:shadow-2xl border border-[#dadce0] dark:border-[#242938] p-6 sm:p-8 flex flex-col gap-6 max-h-[90vh] overflow-y-auto my-auto animate-in fade-in zoom-in-95 transition-colors duration-200">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-[#dadce0] dark:border-[#242938] pb-4">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-[#fce8e6] dark:bg-[#321c1f] text-[#ea4335] dark:text-[#f87171] border border-[#fad2cf] dark:border-[#452226] flex items-center justify-center">
              <span className="material-symbols-outlined text-[24px]">description</span>
            </div>
            <div>
              <h2 className="font-headline-md text-headline-md text-[#202124] dark:text-[#f1f3f7] font-semibold">
                Executive Document Summary
              </h2>
              <p className="font-label-sm text-label-sm text-[#5f6368] dark:text-[#9aa0a6]">
                {document.title || document.filename}
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-[#5f6368] dark:text-[#9aa0a6] hover:text-[#202124] dark:hover:text-[#f1f3f7] hover:bg-[#f1f3f4] dark:hover:bg-[#222838] transition-colors cursor-pointer"
          >
            <span className="material-symbols-outlined text-[20px]">close</span>
          </button>
        </div>

        {/* Content */}
        <div className="flex flex-col gap-4">
          {summary.objective && (
            <div className="bg-[#f8f9fa] dark:bg-[#1a1e2a] border border-[#dadce0] dark:border-[#282f40] p-4 rounded-xl">
              <h3 className="font-title-sm text-title-sm text-[#1a73e8] dark:text-[#38bdf8] font-semibold mb-1 flex items-center gap-1.5">
                <span className="material-symbols-outlined text-[18px]">target</span>
                Core Objective
              </h3>
              <p className="font-body-md text-body-md text-[#202124] dark:text-[#e2e8f0] leading-relaxed">
                {summary.objective}
              </p>
            </div>
          )}

          {summary.problem && (
            <div className="bg-[#f8f9fa] dark:bg-[#1a1e2a] border border-[#dadce0] dark:border-[#282f40] p-4 rounded-xl">
              <h3 className="font-title-sm text-title-sm text-[#ea4335] dark:text-[#f87171] font-semibold mb-1 flex items-center gap-1.5">
                <span className="material-symbols-outlined text-[18px]">report_problem</span>
                Problem Statement
              </h3>
              <p className="font-body-md text-body-md text-[#202124] dark:text-[#e2e8f0] leading-relaxed">
                {summary.problem}
              </p>
            </div>
          )}

          {summary.solution && (
            <div className="bg-[#f8f9fa] dark:bg-[#1a1e2a] border border-[#dadce0] dark:border-[#282f40] p-4 rounded-xl">
              <h3 className="font-title-sm text-title-sm text-[#34a853] dark:text-[#4ade80] font-semibold mb-1 flex items-center gap-1.5">
                <span className="material-symbols-outlined text-[18px]">check_circle</span>
                Proposed Solution & Architecture
              </h3>
              <p className="font-body-md text-body-md text-[#202124] dark:text-[#e2e8f0] leading-relaxed">
                {summary.solution}
              </p>
            </div>
          )}

          {/* Tech & Algorithms tags */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {Array.isArray(summary.technologies) && summary.technologies.length > 0 && (
              <div className="bg-[#f8f9fa] dark:bg-[#1a1e2a] border border-[#dadce0] dark:border-[#282f40] p-4 rounded-xl">
                <h4 className="font-label-md text-label-md text-[#5f6368] dark:text-[#9aa0a6] font-medium mb-2">Technologies Used</h4>
                <div className="flex flex-wrap gap-1.5">
                  {summary.technologies.map((t, idx) => (
                    <span key={idx} className="px-2.5 py-1 rounded-md bg-[#e8f0fe] dark:bg-[#1e2a44] text-[#1967d2] dark:text-[#8ab4f8] font-label-sm text-label-sm border border-[#d2e3fc] dark:border-[#28385a]">
                      {t}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {Array.isArray(summary.algorithms) && summary.algorithms.length > 0 && (
              <div className="bg-[#f8f9fa] dark:bg-[#1a1e2a] border border-[#dadce0] dark:border-[#282f40] p-4 rounded-xl">
                <h4 className="font-label-md text-label-md text-[#5f6368] dark:text-[#9aa0a6] font-medium mb-2">Algorithms & Methods</h4>
                <div className="flex flex-wrap gap-1.5">
                  {summary.algorithms.map((a, idx) => (
                    <span key={idx} className="px-2.5 py-1 rounded-md bg-[#e6f4ea] dark:bg-[#122b1b] text-[#137333] dark:text-[#4ade80] font-label-sm text-label-sm border border-[#ceead6] dark:border-[#1d472c]">
                      {a}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>

          {summary.results && (
            <div className="bg-[#f8f9fa] dark:bg-[#1a1e2a] border border-[#dadce0] dark:border-[#282f40] p-4 rounded-xl">
              <h3 className="font-title-sm text-title-sm text-[#202124] dark:text-[#f1f3f7] font-semibold mb-1">
                Results & Conclusion
              </h3>
              <p className="font-body-md text-body-md text-[#5f6368] dark:text-[#9aa0a6] leading-relaxed">
                {summary.results}
              </p>
            </div>
          )}

          {!summary.objective && !summary.solution && (
            <div className="p-8 text-center text-[#5f6368] dark:text-[#9aa0a6]">
              <span className="material-symbols-outlined text-[36px] text-[#80868b] dark:text-[#64748b] mb-2">pending</span>
              <p>Summary parsing is currently in progress or not available for this document yet.</p>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex justify-end pt-3 border-t border-[#dadce0] dark:border-[#242938]">
          <button
            onClick={onClose}
            className="px-5 py-2 rounded-lg bg-[#1a73e8] hover:bg-[#1557b0] text-white font-title-sm text-title-sm transition-colors cursor-pointer"
          >
            Close Summary
          </button>
        </div>
      </div>
    </div>
  );
}
