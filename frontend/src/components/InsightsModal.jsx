import React from 'react';

export default function InsightsModal({ insights, document, onClose }) {
  if (!insights) {
    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center p-space-md overflow-y-auto">
        <div onClick={onClose} className="fixed inset-0 bg-[#202124]/40 dark:bg-black/60 backdrop-blur-[2px] transition-opacity" />
        <div className="relative z-10 w-full max-w-md bg-white dark:bg-[#151821] rounded-2xl shadow-[0_8px_32px_rgba(32,33,36,0.16)] dark:shadow-2xl border border-[#dadce0] dark:border-[#242938] p-6 text-center flex flex-col items-center gap-4 transition-colors duration-200">
          <span className="material-symbols-outlined text-[36px] text-[#1a73e8] dark:text-[#38bdf8] animate-spin">refresh</span>
          <p className="font-body-md text-body-md text-[#5f6368] dark:text-[#9aa0a6]">Loading technical insights and source index...</p>
          <button onClick={onClose} className="px-4 py-1.5 rounded-lg border border-[#dadce0] dark:border-[#282f40] text-body-sm font-medium hover:bg-[#f1f3f4] dark:hover:bg-[#222838] text-[#3c4043] dark:text-[#cbd5e1] cursor-pointer">
            Cancel
          </button>
        </div>
      </div>
    );
  }

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
            <div className="w-10 h-10 rounded-xl bg-[#e6f4ea] dark:bg-[#122b1b] text-[#137333] dark:text-[#4ade80] border border-[#ceead6] dark:border-[#1d472c] flex items-center justify-center">
              <span className="material-symbols-outlined text-[24px]">account_tree</span>
            </div>
            <div>
              <h2 className="font-headline-md text-headline-md text-[#202124] dark:text-[#f1f3f7] font-semibold">
                Document Technical Insights & Source Index
              </h2>
              <p className="font-label-sm text-label-sm text-[#5f6368] dark:text-[#9aa0a6]">
                {insights.filename || insights.title}
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

        {/* Metrics Grid */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <div className="p-3.5 bg-[#f8f9fa] dark:bg-[#1a1e2a] border border-[#dadce0] dark:border-[#282f40] rounded-xl text-center">
            <span className="font-label-sm text-label-sm text-[#5f6368] dark:text-[#9aa0a6] block">Total Pages</span>
            <span className="font-headline-md text-headline-md text-[#202124] dark:text-[#f1f3f7] font-semibold">
              {insights.page_count || 0}
            </span>
          </div>

          <div className="p-3.5 bg-[#f8f9fa] dark:bg-[#1a1e2a] border border-[#dadce0] dark:border-[#282f40] rounded-xl text-center">
            <span className="font-label-sm text-label-sm text-[#5f6368] dark:text-[#9aa0a6] block">Text Pages</span>
            <span className="font-headline-md text-headline-md text-[#34a853] dark:text-[#4ade80] font-semibold">
              {insights.text_page_count || 0}
            </span>
          </div>

          <div className="p-3.5 bg-[#f8f9fa] dark:bg-[#1a1e2a] border border-[#dadce0] dark:border-[#282f40] rounded-xl text-center">
            <span className="font-label-sm text-label-sm text-[#5f6368] dark:text-[#9aa0a6] block">Total Chunks</span>
            <span className="font-headline-md text-headline-md text-[#1a73e8] dark:text-[#38bdf8] font-semibold">
              {insights.total_chunks || 0}
            </span>
          </div>

          <div className="p-3.5 bg-[#f8f9fa] dark:bg-[#1a1e2a] border border-[#dadce0] dark:border-[#282f40] rounded-xl text-center">
            <span className="font-label-sm text-label-sm text-[#5f6368] dark:text-[#9aa0a6] block">Word Count</span>
            <span className="font-headline-md text-headline-md text-[#202124] dark:text-[#f1f3f7] font-semibold">
              {(insights.total_words || 0).toLocaleString()}
            </span>
          </div>
        </div>

        {/* Character & PII Telemetry */}
        <div className="p-4 bg-[#f8f9fa] dark:bg-[#1a1e2a] border border-[#dadce0] dark:border-[#282f40] rounded-xl flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <span className="material-symbols-outlined text-[#1a73e8] dark:text-[#38bdf8] text-[20px]">fingerprint</span>
            <div>
              <span className="font-title-sm text-title-sm text-[#202124] dark:text-[#f1f3f7] font-medium block">
                Total Characters: {(insights.total_characters || 0).toLocaleString()}
              </span>
              <span className="font-label-sm text-label-sm text-[#5f6368] dark:text-[#9aa0a6]">
                Normalized and tokenized for 1536-dimensional vector embedding
              </span>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <span className="font-label-sm text-label-sm text-[#5f6368] dark:text-[#9aa0a6]">PII Redaction:</span>
            {insights.pii_detected ? (
              <span className="px-2.5 py-0.5 rounded-full bg-[#fef7e0] dark:bg-[#2c220f] text-[#b06000] dark:text-[#fcd34d] font-label-sm text-label-sm font-medium border border-[#feefa3] dark:border-[#453715]">
                PII Redacted
              </span>
            ) : (
              <span className="px-2.5 py-0.5 rounded-full bg-[#e6f4ea] dark:bg-[#122b1b] text-[#137333] dark:text-[#4ade80] font-label-sm text-label-sm font-medium border border-[#ceead6] dark:border-[#1d472c]">
                Clean / No PII
              </span>
            )}
          </div>
        </div>

        {/* Validation Checks Audit Table */}
        <div className="flex flex-col gap-2">
          <h3 className="font-title-sm text-title-sm text-[#202124] dark:text-[#f1f3f7] font-semibold flex items-center gap-1.5">
            <span className="material-symbols-outlined text-[18px] text-[#1a73e8] dark:text-[#38bdf8]">verified_user</span>
            Automated Validation Pipeline Audits
          </h3>

          <div className="border border-[#dadce0] dark:border-[#282f40] rounded-xl overflow-hidden">
            <table className="w-full text-left text-body-sm">
              <thead className="bg-[#f1f3f4] dark:bg-[#1f2433] text-[#5f6368] dark:text-[#9aa0a6] font-label-md text-label-md border-b border-[#dadce0] dark:border-[#282f40]">
                <tr>
                  <th className="py-2.5 px-4 font-semibold">Check</th>
                  <th className="py-2.5 px-4 font-semibold">Status</th>
                  <th className="py-2.5 px-4 font-semibold">Details</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#f1f3f4] dark:divide-[#282f40]">
                {Array.isArray(insights.validation_checks) && insights.validation_checks.length > 0 ? (
                  insights.validation_checks.map((check, idx) => (
                    <tr key={idx} className="hover:bg-[#f8f9fa] dark:hover:bg-[#1c2130] transition-colors">
                      <td className="py-2.5 px-4 font-medium text-[#202124] dark:text-[#f1f3f7]">
                        {check.check_name}
                      </td>
                      <td className="py-2.5 px-4">
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded-full font-citation-ref text-[10px] font-semibold uppercase ${
                            check.status === 'PASS'
                              ? 'bg-[#e6f4ea] dark:bg-[#122b1b] text-[#137333] dark:text-[#4ade80] border border-[#ceead6] dark:border-[#1d472c]'
                              : check.status === 'WARN'
                              ? 'bg-[#fef7e0] dark:bg-[#2c220f] text-[#b06000] dark:text-[#fcd34d] border border-[#feefa3] dark:border-[#453715]'
                              : 'bg-[#fce8e6] dark:bg-[#321c1f] text-[#ea4335] dark:text-[#f87171] border border-[#fad2cf] dark:border-[#452226]'
                          }`}
                        >
                          {check.status}
                        </span>
                      </td>
                      <td className="py-2.5 px-4 text-[#5f6368] dark:text-[#9aa0a6]">
                        {check.message}
                      </td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan={3} className="py-4 px-4 text-center text-[#80868b] dark:text-[#64748b]">
                      No audit checks recorded.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* Footer */}
        <div className="flex justify-end pt-3 border-t border-[#dadce0] dark:border-[#242938]">
          <button
            onClick={onClose}
            className="px-5 py-2 rounded-lg bg-[#1a73e8] hover:bg-[#1557b0] text-white font-title-sm text-title-sm transition-colors cursor-pointer"
          >
            Close Insights
          </button>
        </div>
      </div>
    </div>
  );
}
