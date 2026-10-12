import React, { useState, useEffect, useRef } from 'react';
import eventService from '../services/eventService';
import SummaryModal from '../components/SummaryModal';
import InsightsModal from '../components/InsightsModal';
import ThemeToggle from '../components/ThemeToggle';

export default function Conversation({ onOpenOnboarding }) {
  // Document State
  const [document, setDocument] = useState(null);
  const [insights, setInsights] = useState(null);
  const [viewMode, setViewMode] = useState('chat'); // 'chat' or 'library'
  const [documentsList, setDocumentsList] = useState([]);
  const [loadingDocuments, setLoadingDocuments] = useState(false);

  // Modals for specific document from library or active doc
  const [activeModalDoc, setActiveModalDoc] = useState(null);
  const [activeModalInsights, setActiveModalInsights] = useState(null);

  // Upload State
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [processingStage, setProcessingStage] = useState('');
  const [uploadError, setUploadError] = useState('');
  const [isDragOver, setIsDragOver] = useState(false);
  const fileInputRef = useRef(null);

  // Chat State
  const [messages, setMessages] = useState([]);
  const [inputQuery, setInputQuery] = useState('');
  const [chatLoading, setChatLoading] = useState(false);
  const [chatError, setChatError] = useState('');
  const [copiedIndex, setCopiedIndex] = useState(null);
  const messagesEndRef = useRef(null);

  // Modals
  const [showSummaryModal, setShowSummaryModal] = useState(false);
  const [showInsightsModal, setShowInsightsModal] = useState(false);

  // System & AI Info
  const [systemInfo, setSystemInfo] = useState(null);

  // 3-second dialog prompt when clicking Workbench without an active document
  const [showWorkbenchPrompt, setShowWorkbenchPrompt] = useState(false);
  const workbenchPromptTimeoutRef = useRef(null);

  useEffect(() => {
    return () => {
      if (workbenchPromptTimeoutRef.current) {
        clearTimeout(workbenchPromptTimeoutRef.current);
      }
    };
  }, []);

  const handleWorkbenchClick = () => {
    if (document) {
      setViewMode('chat');
    } else {
      if (workbenchPromptTimeoutRef.current) {
        clearTimeout(workbenchPromptTimeoutRef.current);
      }
      setShowWorkbenchPrompt(true);
      workbenchPromptTimeoutRef.current = setTimeout(() => {
        setShowWorkbenchPrompt(false);
      }, 3000);
    }
  };

  // Auto-scroll chat
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages, chatLoading]);

  // Initial Load: Fetch System Info, Latest Document & Documents List
  useEffect(() => {
    loadSystemInfo();
    loadLatestDocument();
    loadDocuments();
  }, []);

  const loadSystemInfo = async () => {
    try {
      const data = await eventService.getSystemInfo();
      if (data && data.ai) {
        setSystemInfo(data);
      }
    } catch (err) {
      console.error('Failed to load system info:', err);
    }
  };

  const loadDocuments = async () => {
    setLoadingDocuments(true);
    try {
      const data = await eventService.getDocuments();
      if (data && data.status === 'S' && Array.isArray(data.documents)) {
        // Strictly capped at 3 documents maximum
        const list = data.documents.slice(0, 3).map(d => ({ ...d, id: d.id || d.document_id }));
        setDocumentsList(list);
        return list;
      }
    } catch (err) {
      console.error('Failed to load documents list:', err);
    } finally {
      setLoadingDocuments(false);
    }
    return [];
  };

  const loadLatestDocument = async () => {
    try {
      const data = await eventService.getLatestDocument();
      if (data && data.status === 'S' && data.document) {
        const unifiedDoc = { ...data.document, id: data.document.id || data.document.document_id };
        setDocument(unifiedDoc);
        // Also fetch insights
        fetchDocumentInsights(unifiedDoc.id);
        // Add welcome message
        setMessages([
          {
            id: 'init-1',
            sender: 'ai',
            time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
            title: 'AI Assistant',
            text: `Document "${unifiedDoc.title || unifiedDoc.filename}" is loaded and ready. You can ask questions, extract specifications, or explore data from this document.`,
            sources: [],
          }
        ]);
      } else {
        // No document yet, switch to library view
        setViewMode('library');
      }
    } catch {
      setViewMode('library');
    }
  };

  const fetchDocumentInsights = async (docId) => {
    try {
      const data = await eventService.getDocumentInsights(docId);
      if (data && data.status === 'S' && data.insights) {
        setInsights(data.insights);
      }
    } catch (err) {
      console.error('Failed to fetch insights:', err);
    }
  };



  // Select a document from library to be the active chat document
  const handleSelectActiveDocument = (doc) => {
    const unifiedDoc = { ...doc, id: doc.id || doc.document_id };
    setDocument(unifiedDoc);
    fetchDocumentInsights(unifiedDoc.id);
    setViewMode('chat');
    setMessages([
      {
        id: 'switch-' + Date.now(),
        sender: 'ai',
        time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
        title: 'AI Assistant',
        text: `Active context switched to "${unifiedDoc.title || unifiedDoc.filename}". Ask any questions about this document.`,
        sources: [],
      }
    ]);
  };

  // Delete a document from library
  const handleDeleteDocument = async (docId, docName) => {
    if (!window.confirm(`Delete "${docName || 'this document'}" from your library? All indexed vector chunks and embeddings will be removed.`)) {
      return;
    }
    try {
      // Optimistically update list immediately
      setDocumentsList((prev) => prev.filter(d => d.id !== docId && d.document_id !== docId));
      await eventService.deleteDocument(docId);
      const remaining = await loadDocuments();
      if (document && (document.id === docId || document.document_id === docId)) {
        if (remaining && remaining.length > 0) {
          const nextDoc = { ...remaining[0], id: remaining[0].id || remaining[0].document_id };
          setDocument(nextDoc);
          fetchDocumentInsights(nextDoc.id);
        } else {
          setDocument(null);
          setInsights(null);
          setMessages([]);
        }
      }
    } catch (err) {
      console.error('Failed to delete document:', err);
      alert(err.message || 'Failed to delete document');
      loadDocuments();
    }
  };

  const handleOpenSummaryModal = (doc) => {
    setActiveModalDoc(doc);
    setShowSummaryModal(true);
  };

  const handleOpenInsightsModal = async (doc) => {
    setActiveModalDoc(doc);
    setActiveModalInsights(null);
    setShowInsightsModal(true);
    try {
      const data = await eventService.getDocumentInsights(doc.id || doc.document_id);
      if (data && data.status === 'S' && data.insights) {
        setActiveModalInsights(data.insights);
      }
    } catch (err) {
      console.error('Failed to load insights for doc:', err);
    }
  };

  // Handle File Upload
  const handleFileUpload = async (file) => {
    if (!file) return;

    if (!file.name.toLowerCase().endsWith('.pdf') && file.type !== 'application/pdf') {
      setUploadError('Invalid file format: Please select a text-based PDF document (.pdf).');
      return;
    }

    const maxMB = systemInfo?.ai?.validation?.max_file_size_mb || 20;
    if (file.size > maxMB * 1024 * 1024) {
      setUploadError(`File exceeds maximum size limit of ${maxMB}MB.`);
      return;
    }

    setUploadError('');
    setUploading(true);
    setUploadProgress(0);
    setProcessingStage('Uploading document to secure RAG pipeline...');

    try {
      const res = await eventService.uploadDocument(file, file.name, (progress) => {
        setUploadProgress(progress);
        if (progress >= 100) {
          setProcessingStage('Validating PDF structure & extracting text...');
        }
      });

      if (res && res.document_id) {
        const docId = res.document_id;
        setProcessingStage('Neural parsing: Chunking, PII redaction & vector embeddings...');

        // Poll until ready
        const readyDoc = await eventService.pollDocumentUntilReady(docId, (status, stage, doc) => {
          setProcessingStage(`${stage}...`);
          if (doc) setDocument({ ...doc, id: doc.id || doc.document_id || docId });
        });

        const unifiedDoc = { ...readyDoc, id: readyDoc?.id || readyDoc?.document_id || docId };
        setDocument(unifiedDoc);
        await fetchDocumentInsights(docId);
        await loadDocuments();
        setUploading(false);
        setViewMode('chat');

        // Add welcome message for new document
        setMessages([
          {
            id: Date.now().toString(),
            sender: 'ai',
            time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
            title: 'PDF_Intelligence Synthesis',
            badge: 'Ready & Grounded',
            latency: '240ms',
            text: `Document "${file.name}" has been processed and indexed into ${readyDoc.total_chunks || 'multiple'} vector chunks. What would you like to explore?`,
            sources: [],
          }
        ]);
      }
    } catch (err) {
      setUploadError(err.message || 'Failed to process PDF. Please ensure it is a valid text-based PDF.');
      setUploading(false);
    }
  };

  // Handle Drag & Drop
  const handleDragOver = (e) => {
    e.preventDefault();
    setIsDragOver(true);
  };

  const handleDragLeave = (e) => {
    e.preventDefault();
    setIsDragOver(false);
  };

  const handleDrop = (e) => {
    e.preventDefault();
    setIsDragOver(false);
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      handleFileUpload(e.dataTransfer.files[0]);
    }
  };

  // Handle Sending a Question
  const handleSendQuery = async (queryText = null) => {
    const query = (queryText || inputQuery).trim();
    if (!query || chatLoading) return;

    const activeDocId = document?.id || document?.document_id;
    if (!activeDocId) {
      setViewMode('upload');
      setUploadError('Please upload a PDF document before asking questions.');
      return;
    }

    const userMessageId = Date.now().toString();
    const userMsg = {
      id: userMessageId,
      sender: 'user',
      time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
      text: query,
    };

    setMessages((prev) => [...prev, userMsg]);
    setInputQuery('');
    setChatLoading(true);
    setChatError('');

    const startTime = performance.now();

    try {
      const res = await eventService.askQuestion(activeDocId, query, 1);
      const latency = `${Math.round(performance.now() - startTime)}ms`;

      const aiMsg = {
        id: (Date.now() + 1).toString(),
        sender: 'ai',
        time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
        title: 'AI Response',
        text: res.answer || 'No answer generated for this query.',
        sources: Array.isArray(res.sources) ? res.sources : [],
        confidence: res.confidence,
      };

      setMessages((prev) => [...prev, aiMsg]);
    } catch (err) {
      setChatError(err.message || 'Failed to get answer from AI engine.');
      const errorAiMsg = {
        id: (Date.now() + 1).toString(),
        sender: 'ai',
        isError: true,
        time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
        title: 'Error',
        text: `Unable to answer: ${err.message}. Please try rephrasing your query.`,
        sources: [],
      };
      setMessages((prev) => [...prev, errorAiMsg]);
    } finally {
      setChatLoading(false);
    }
  };

  const handleCopyText = (text, idx) => {
    navigator.clipboard.writeText(text);
    setCopiedIndex(idx);
    setTimeout(() => setCopiedIndex(null), 2000);
  };


  const aiConfig = systemInfo?.ai;
  const primaryModel = aiConfig?.primary?.model || 'gemini-3.5-flash-lite';
  const primaryProvider = aiConfig?.primary?.provider || 'Google Gemini';
  const embedModel = aiConfig?.embeddings?.model || 'gemini-embedding-001';
  const embedDim = aiConfig?.embeddings?.dimension || 1536;

  return (
    <div className="bg-bg-surface dark:bg-[#0c0d12] font-body-md text-body-md text-on-surface dark:text-[#f1f3f7] h-screen flex flex-col antialiased overflow-hidden transition-colors duration-200">
      {/* ==================== HEADER ==================== */}
      <header className="fixed top-0 left-0 right-0 h-16 z-50 bg-[#ffffff]/95 dark:bg-[#111318]/95 backdrop-blur-md border-b border-[#dadce0] dark:border-[#222631] shadow-[0_1px_3px_rgba(60,64,67,0.08)] dark:shadow-none transition-colors duration-200">
        <div className="h-16 w-full px-margin-desktop flex items-center justify-between">
          <div className="flex items-center gap-space-md">
            <div
              onClick={handleWorkbenchClick}
              className="flex items-center gap-2.5 cursor-pointer group"
              title="PDF_Intelligence Workbench"
            >
              {/* Distinct & attractive modern PDF Intelligence brand icon container */}
              <div className="w-12 h-12 rounded-2xl bg-white dark:bg-[#161922] border border-[#dadce0] dark:border-[#2a3040] shadow-sm flex items-center justify-center p-2 group-hover:border-[#1a73e8] dark:group-hover:border-[#38bdf8] transition-colors">
                <svg className="w-full h-full" viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <defs>
                    <linearGradient id="piGrad1" x1="8" y1="6" x2="40" y2="42" gradientUnits="userSpaceOnUse">
                      <stop offset="0%" stopColor="#1A73E8" />
                      <stop offset="60%" stopColor="#4F46E5" />
                      <stop offset="100%" stopColor="#7C3AED" />
                    </linearGradient>
                    <linearGradient id="piGradFold" x1="28" y1="6" x2="40" y2="18" gradientUnits="userSpaceOnUse">
                      <stop offset="0%" stopColor="#60A5FA" />
                      <stop offset="100%" stopColor="#2563EB" />
                    </linearGradient>
                    <linearGradient id="piSparkGrad" x1="18" y1="20" x2="32" y2="34" gradientUnits="userSpaceOnUse">
                      <stop offset="0%" stopColor="#00F2FE" />
                      <stop offset="50%" stopColor="#38BDF8" />
                      <stop offset="100%" stopColor="#06B6D4" />
                    </linearGradient>
                  </defs>

                  {/* Soft background sheet depth */}
                  <path
                    d="M14 8C11.7909 8 10 9.7909 10 12V36C10 38.2091 11.7909 40 14 40H34C36.2091 40 38 38.2091 38 36V18L28 8H14Z"
                    fill="url(#piGrad1)"
                    fillOpacity="0.12"
                  />

                  {/* Main Folded Document Silhouette */}
                  <path
                    d="M13 9C11.3431 9 10 10.3431 10 12V36C10 37.6569 11.3431 39 13 39H35C36.6569 39 38 37.6569 38 36V17.5L29.5 9H13Z"
                    stroke="url(#piGrad1)"
                    strokeWidth="2.5"
                    strokeLinejoin="round"
                  />

                  {/* Folded Top-Right Corner */}
                  <path
                    d="M29 9V16C29 17.1046 29.8954 18 31 18H38"
                    stroke="url(#piGradFold)"
                    strokeWidth="2.5"
                    strokeLinejoin="round"
                    fill="url(#piGradFold)"
                    fillOpacity="0.25"
                  />

                  {/* Document Text / Vector Stream lines */}
                  <line x1="15" y1="17" x2="23" y2="17" stroke="url(#piGrad1)" strokeWidth="2" strokeLinecap="round" strokeOpacity="0.7" />
                  <line x1="15" y1="22" x2="20" y2="22" stroke="url(#piGrad1)" strokeWidth="2" strokeLinecap="round" strokeOpacity="0.7" />

                  {/* Neural AI Synapses */}
                  <path
                    d="M18 31L24 27M24 27L30 31M24 27V34"
                    stroke="#06B6D4"
                    strokeWidth="1.5"
                    strokeLinecap="round"
                    strokeDasharray="2 1.5"
                  />

                  {/* 4-Point Radiant Intelligence Spark */}
                  <path
                    d="M24 21C24 24.5 22.5 26 19 26C22.5 26 24 27.5 24 31C24 27.5 25.5 26 29 26C25.5 26 24 24.5 24 21Z"
                    fill="url(#piSparkGrad)"
                  />

                  {/* Spark Center Pinpoint */}
                  <circle cx="24" cy="26" r="1.3" fill="#FFFFFF" />

                  {/* Satellite Neural Nodes */}
                  <circle cx="18" cy="31" r="1.5" fill="#10B981" />
                  <circle cx="30" cy="31" r="1.5" fill="#3B82F6" />
                  <circle cx="24" cy="34" r="1.5" fill="#8B5CF6" />
                </svg>
              </div>
              <span className="font-title-sm text-[17px] text-[#202124] dark:text-[#f1f3f7] tracking-tight font-semibold">
                PDF_Intelligence
              </span>
            </div>
            <span className="hidden sm:inline-flex items-center px-2.5 py-0.5 rounded-full bg-[#f1f3f4] dark:bg-[#1a1e28] text-[#3c4043] dark:text-[#9aa0a6] font-label-sm text-label-sm border border-[#dadce0] dark:border-[#2a3040] font-medium">
              <span className="w-1.5 h-1.5 rounded-full bg-[#34a853] mr-1.5 animate-pulse" />
              RAG Engine v2.4
            </span>
          </div>

          <nav className="hidden md:flex items-center gap-space-lg">
            <button
              type="button"
              onClick={handleWorkbenchClick}
              className={`transition-colors font-title-sm rounded-lg px-space-sm py-space-xs cursor-pointer ${viewMode === 'chat'
                ? 'bg-[#e8f0fe] dark:bg-[#1e2a44] text-[#1967d2] dark:text-[#8ab4f8]'
                : 'text-[#5f6368] dark:text-[#94a3b8] hover:text-[#202124] dark:hover:text-[#f1f3f7]'
                }`}
            >
              Workbench
            </button>
            <button
              type="button"
              onClick={() => {
                setViewMode('library');
                loadDocuments();
              }}
              className={`transition-colors font-medium rounded-lg px-space-sm py-space-xs cursor-pointer flex items-center gap-1.5 ${viewMode === 'library'
                ? 'bg-[#e8f0fe] dark:bg-[#1e2a44] text-[#1967d2] dark:text-[#8ab4f8]'
                : 'text-[#5f6368] dark:text-[#94a3b8] hover:text-[#202124] dark:hover:text-[#f1f3f7]'
                }`}
            >
              <span>Document Library</span>
              <span className="text-[11px] px-1.5 py-0.2 bg-[#e8eaed] dark:bg-[#252c3d] text-[#3c4043] dark:text-[#cbd5e1] rounded-full font-semibold">
                {documentsList.length}/3
              </span>
            </button>
            <button
              type="button"
              onClick={onOpenOnboarding}
              className="text-[#5f6368] dark:text-[#94a3b8] hover:text-[#202124] dark:hover:text-[#f1f3f7] transition-colors font-medium flex items-center gap-1 cursor-pointer"
            >
              <span className="material-symbols-outlined text-[17px] text-[#1a73e8] dark:text-[#38bdf8]">help_outline</span>
              <span>How It Works</span>
            </button>
          </nav>

          <div className="flex items-center gap-space-sm">
            {/* Theme Toggle (Light / Dark mode) */}
            <ThemeToggle />

            <div className="flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-[#f8f9fa] dark:bg-[#1a1e28] border border-[#dadce0] dark:border-[#2a3040] text-[#3c4043] dark:text-[#cbd5e1] font-label-sm text-xs font-medium">
              <span className="w-2 h-2 rounded-full bg-[#34a853]" />
              <span>Session Active</span>
            </div>
          </div>
        </div>
      </header>

      {/* ==================== MAIN DUAL-PANEL WORKSPACE ==================== */}
      <main className="w-full pt-16 bg-background flex-1 flex flex-col h-screen overflow-hidden">
        <div className="w-full flex-1 grid grid-cols-1 lg:grid-cols-12 h-[calc(100vh-4rem)] overflow-hidden">
          {/* LEFT / CENTER: Flex-1 Main Conversation & Document Area (Col 8 on Large) */}
          <div className="lg:col-span-8 flex flex-col h-[calc(100vh-4rem)] overflow-hidden bg-[#ffffff] dark:bg-[#0e1017] relative border-r border-[#dadce0] dark:border-[#222631] transition-colors duration-200">
            {/* Top Section: Document Context Card (Only shown in Chat Workbench view) */}
            {viewMode === 'chat' && document && (
              <div className="shrink-0 px-margin-desktop pt-space-md pb-2 bg-white dark:bg-[#0e1017] transition-colors duration-200">
                <div className="w-full rounded-xl bg-white dark:bg-[#151821] border border-[#dadce0] dark:border-[#242938] p-space-md shadow-[0_1px_3px_rgba(60,64,67,0.08)] dark:shadow-none relative transition-colors duration-200">
                  <div className="flex flex-col md:flex-row md:items-center justify-between gap-space-md">
                    <div className="flex items-start md:items-center gap-space-md">
                      <div className="w-12 h-12 rounded-lg bg-[#fce8e6] dark:bg-[#321c1f] flex items-center justify-center text-[#ea4335] dark:text-[#f87171] border border-[#fad2cf] dark:border-[#452226] shrink-0">
                        <span className="material-symbols-outlined text-[28px]">picture_as_pdf</span>
                      </div>
                      <div>
                        <div className="flex items-center gap-space-sm flex-wrap">
                          <h1 className="font-title-sm text-title-sm text-[#202124] dark:text-[#f1f3f7] tracking-tight font-semibold">
                            {document.filename || document.title || 'Uploaded Document'}
                          </h1>
                          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full bg-[#e6f4ea] dark:bg-[#122b1b] text-[#137333] dark:text-[#4ade80] font-label-sm text-label-sm border border-[#ceead6] dark:border-[#1d472c] font-medium">
                            <span className="material-symbols-outlined text-[13px] text-[#34a853] dark:text-[#4ade80]">check_circle</span>
                            {document.status === 'COMPLETED' ? 'Document ready' : document.status || 'Processing'}
                          </span>
                        </div>
                        <p className="font-label-sm text-label-sm text-[#5f6368] dark:text-[#9aa0a6] mt-1 flex items-center gap-2 flex-wrap">
                          <span>{document.file_size ? `${(document.file_size / (1024 * 1024)).toFixed(1)} MB` : 'PDF'}</span>
                          <span>•</span>
                          <span>{document.page_count || insights?.page_count || 1} Pages</span>
                          <span>•</span>
                          <span className="text-[#137333] dark:text-[#4ade80] font-medium">OCR Text Ready</span>
                          <span>•</span>
                          <span className="text-[#80868b] dark:text-[#64748b]">
                            {insights ? `${insights.total_chunks} chunks indexed` : 'Indexed'}
                          </span>
                        </p>
                      </div>
                    </div>

                    {/* Document Action Pills */}
                    <div className="flex items-center gap-space-xs flex-wrap">
                      <button
                        type="button"
                        onClick={() => {
                          setViewMode(viewMode === 'library' ? 'chat' : 'library');
                          loadDocuments();
                        }}
                        className="px-space-sm py-1.5 rounded-lg bg-white dark:bg-[#1c202d] hover:bg-[#f8f9fa] dark:hover:bg-[#252c3e] text-[#3c4043] dark:text-[#cbd5e1] hover:text-[#202124] dark:hover:text-[#f1f3f7] text-body-sm font-medium border border-[#dadce0] dark:border-[#2e3547] flex items-center gap-1.5 shadow-xs transition-all cursor-pointer"
                      >
                        <span className="material-symbols-outlined text-[16px] text-[#1a73e8] dark:text-[#38bdf8]">swap_horiz</span>
                        <span>{viewMode === 'library' ? 'Back to Workbench' : 'Document Library'}</span>
                      </button>
                      <button
                        type="button"
                        onClick={() => setShowInsightsModal(true)}
                        className="px-space-sm py-1.5 rounded-lg bg-white dark:bg-[#1c202d] hover:bg-[#f8f9fa] dark:hover:bg-[#252c3e] text-[#3c4043] dark:text-[#cbd5e1] hover:text-[#202124] dark:hover:text-[#f1f3f7] text-body-sm font-medium border border-[#dadce0] dark:border-[#2e3547] flex items-center gap-1.5 shadow-xs transition-all cursor-pointer"
                      >
                        <span className="material-symbols-outlined text-[16px] text-[#188038] dark:text-[#4ade80]">account_tree</span>
                        <span>View Source Index</span>
                      </button>
                      <button
                        type="button"
                        onClick={() => setShowSummaryModal(true)}
                        className="px-space-sm py-1.5 rounded-lg bg-white dark:bg-[#1c202d] hover:bg-[#f8f9fa] dark:hover:bg-[#252c3e] text-[#3c4043] dark:text-[#cbd5e1] hover:text-[#202124] dark:hover:text-[#f1f3f7] text-body-sm font-medium border border-[#dadce0] dark:border-[#2e3547] flex items-center gap-1.5 shadow-xs transition-all cursor-pointer"
                      >
                        <span className="material-symbols-outlined text-[16px] text-[#ea4335] dark:text-[#f87171]">download</span>
                        <span>Summary</span>
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            )}

            {/* VIEW 1: Active Interactive Conversation Thread */}
            {viewMode === 'chat' && (
              <div className="flex-1 flex flex-col space-y-space-lg overflow-y-auto px-margin-desktop py-space-sm min-h-0">
                {/* Empty State suggestions if only welcome message exists */}
                {messages.length <= 1 && (
                  <div className="p-space-lg rounded-2xl bg-[#f8f9fa] dark:bg-[#13161f] border border-[#dadce0] dark:border-[#242938] flex flex-col gap-3 my-2 transition-colors duration-200">
                    <div className="flex items-center gap-2 text-google-blue dark:text-[#38bdf8] font-title-sm font-semibold">
                      <span className="material-symbols-outlined text-[20px]">lightbulb</span>
                      <span>Suggested Questions:</span>
                    </div>
                    <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
                      {[
                        "What is the main objective and problem statement?",
                        "What sensors, components, or hardware are used?",
                        "Summarize the key results and advantages.",
                      ].map((prompt, pIdx) => (
                        <button
                          key={pIdx}
                          type="button"
                          onClick={() => handleSendQuery(prompt)}
                          className="p-3 bg-white dark:bg-[#1a1e2a] rounded-xl border border-[#dadce0] dark:border-[#2e3547] hover:border-google-blue dark:hover:border-[#38bdf8] hover:bg-[#e8f0fe]/30 dark:hover:bg-[#1e2a44]/40 text-left text-body-sm text-[#202124] dark:text-[#cbd5e1] transition-all cursor-pointer hover:shadow-xs active:scale-[0.98]"
                        >
                          "{prompt}"
                        </button>
                      ))}
                    </div>
                  </div>
                )}

                {/* Messages Stream */}
                {messages.map((msg, index) => {
                  if (msg.sender === 'user') {
                    return (
                      <div key={msg.id || index} className="flex justify-end items-end gap-space-sm group">
                        <div className="flex flex-col items-end max-w-xl">
                          <div className="flex items-center gap-2 mb-1">
                            <span className="font-label-sm text-label-sm text-[#5f6368] dark:text-[#9aa0a6] font-medium">You</span>
                            <span className="font-label-sm text-label-sm text-[#80868b] dark:text-[#64748b]">{msg.time}</span>
                          </div>
                          <div className="rounded-2xl rounded-tr-xs bg-gradient-to-r from-[#1a73e8] to-[#1557b0] text-white px-space-lg py-space-md shadow-sm">
                            <p className="font-body-md text-body-md font-medium text-white">{msg.text}</p>
                          </div>
                        </div>
                        <div className="w-9 h-9 rounded-full bg-[#1a73e8] flex items-center justify-center text-white shadow-xs shrink-0">
                          <span className="material-symbols-outlined text-[20px]">person</span>
                        </div>
                      </div>
                    );
                  }

                  // AI Message Card
                  return (
                    <div key={msg.id || index} className="flex items-start gap-space-md max-w-3xl">
                      {/* AI Sparkle Avatar */}
                      <div className="w-9 h-9 rounded-xl bg-[#e8f0fe] dark:bg-[#1e2a44] border border-[#d2e3fc] dark:border-[#28385a] flex items-center justify-center text-[#1a73e8] dark:text-[#38bdf8] shrink-0 shadow-xs">
                        <span className="material-symbols-outlined text-[20px]">auto_awesome</span>
                      </div>

                      {/* AI Message Card */}
                      <div className="flex-1 rounded-2xl bg-white dark:bg-[#151821] border border-[#dadce0] dark:border-[#242938] p-space-lg shadow-[0_1px_3px_rgba(60,64,67,0.08)] dark:shadow-none relative overflow-hidden transition-all">
                        {/* Left Google Blue rail accent within the radius */}
                        <div className="absolute left-0 top-0 bottom-0 w-1 bg-gradient-to-b from-[#1a73e8] via-[#34a853] to-[#1a73e8]" />

                        {/* Clean Topic Header */}
                        <div className="flex items-center justify-between gap-2 mb-3">
                          <div className="flex items-center gap-2">
                            <span className="font-title-sm text-[14px] font-semibold text-[#202124] dark:text-[#f1f3f7] flex items-center gap-1.5 tracking-wide">
                              <span className="material-symbols-outlined text-[16px] text-[#1a73e8] dark:text-[#38bdf8]">auto_awesome</span>
                              {msg.title && msg.title !== 'PDF_Intelligence Synthesis' ? msg.title : 'AI Response'}
                            </span>
                          </div>
                          <span className="text-[11px] text-[#80868b] dark:text-[#64748b] font-normal">
                            {msg.time}
                          </span>
                        </div>

                        {/* Message Content */}
                        <div className={`font-body-md text-body-md leading-relaxed whitespace-pre-line ${msg.isError ? 'text-[#c5221f] dark:text-[#f87171]' : 'text-[#202124] dark:text-[#e2e8f0]'
                          }`}>
                          {msg.text}
                        </div>

                        {/* Citations & Actions Toolbar */}
                        <div className="mt-4 pt-3 border-t border-[#dadce0] dark:border-[#242938] flex flex-wrap items-center justify-between gap-3">
                          {/* Source Citations */}
                          <div className="flex items-center gap-2 flex-wrap">
                            {Array.isArray(msg.sources) && msg.sources.length > 0 && (
                              msg.sources.map((src, sIdx) => (
                                <div key={sIdx} className="relative group/cite">
                                  <button
                                    type="button"
                                    className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-[#e8f0fe] dark:bg-[#1e283d] hover:bg-[#d2e3fc] dark:hover:bg-[#28385a] active:scale-95 text-[#1967d2] dark:text-[#8ab4f8] border border-[#d2e3fc] dark:border-[#2c3d5e] font-label-md text-label-md transition-all shadow-xs cursor-pointer"
                                  >
                                    <span className="material-symbols-outlined text-[14px]">bookmark</span>
                                    <span className="font-semibold">Source · Page {src.page}</span>
                                    {src.section && (
                                      <span className="text-[11px] text-[#5f6368] dark:text-[#94a3b8]">
                                        · {src.section}
                                      </span>
                                    )}
                                  </button>
                                  {/* Citation Tooltip */}
                                  <div className="absolute bottom-full left-0 mb-2 w-80 p-3.5 rounded-xl bg-white dark:bg-[#1b1f2b] border border-[#dadce0] dark:border-[#2e3547] shadow-[0_4px_16px_rgba(60,64,67,0.15)] dark:shadow-[0_8px_24px_rgba(0,0,0,0.5)] opacity-0 pointer-events-none group-hover/cite:opacity-100 group-hover/cite:pointer-events-auto transition-all duration-200 z-30">
                                    <div className="flex items-center justify-between border-b border-[#dadce0] dark:border-[#2e3547] pb-2 mb-2">
                                      <span className="font-label-sm text-label-sm text-[#1967d2] dark:text-[#8ab4f8] font-semibold">
                                        {src.section || 'Source Grounding'}
                                      </span>
                                      <span className="font-citation-ref text-citation-ref text-[#80868b] dark:text-[#64748b]">
                                        p.{src.page}
                                      </span>
                                    </div>
                                    <p className="font-body-sm text-body-sm text-[#3c4043] dark:text-[#cbd5e1] italic leading-relaxed">
                                      Verified against document page {src.page}.
                                    </p>
                                  </div>
                                </div>
                              ))
                            )}
                          </div>

                          {/* Toolbar Action Buttons with tactile gestures */}
                          <div className="flex items-center gap-1 text-[#5f6368] dark:text-[#9aa0a6]">
                            <button
                              type="button"
                              onClick={() => handleCopyText(msg.text, index)}
                              className={`p-1.5 rounded-lg transition-all duration-150 active:scale-90 cursor-pointer flex items-center gap-1 ${copiedIndex === index
                                  ? 'bg-[#e6f4ea] dark:bg-[#122b1b] text-[#137333] dark:text-[#4ade80] border border-[#ceead6] dark:border-[#1d472c]'
                                  : 'hover:bg-[#f1f3f4] dark:hover:bg-[#222838] hover:text-[#202124] dark:hover:text-[#f1f3f7]'
                                }`}
                              title={copiedIndex === index ? 'Copied to clipboard' : 'Copy Answer'}
                            >
                              <span className="material-symbols-outlined text-[17px]">
                                {copiedIndex === index ? 'check' : 'content_copy'}
                              </span>
                              {copiedIndex === index && <span className="text-xs font-medium">Copied</span>}
                            </button>

                            <button
                              type="button"
                              onClick={() => {
                                const lastUserQuery = messages.filter(m => m.sender === 'user').slice(-1)[0]?.text;
                                if (lastUserQuery) handleSendQuery(lastUserQuery);
                              }}
                              className="p-1.5 rounded-lg hover:bg-[#f1f3f4] dark:hover:bg-[#222838] hover:text-[#1a73e8] dark:hover:text-[#38bdf8] active:scale-90 transition-all duration-150 cursor-pointer"
                              title="Regenerate"
                            >
                              <span className="material-symbols-outlined text-[17px]">replay</span>
                            </button>
                          </div>
                        </div>
                      </div>
                    </div>
                  );
                })}

                {/* Loading indicator */}
                {chatLoading && (
                  <div className="flex items-start gap-space-md max-w-2xl">
                    <div className="w-9 h-9 rounded-xl bg-[#e8f0fe] dark:bg-[#1e2a44] border border-[#d2e3fc] dark:border-[#28385a] flex items-center justify-center text-[#1a73e8] dark:text-[#38bdf8] shrink-0 shadow-xs">
                      <span className="material-symbols-outlined text-[20px] animate-spin">refresh</span>
                    </div>
                    <div className="p-4 rounded-2xl bg-white dark:bg-[#151821] border border-[#dadce0] dark:border-[#242938] shadow-sm flex items-center gap-3">
                      <span className="w-2.5 h-2.5 rounded-full bg-google-blue dark:bg-[#38bdf8] animate-ping" />
                      <span className="text-body-md text-[#5f6368] dark:text-[#cbd5e1] font-medium">
                        Generating answer...
                      </span>
                    </div>
                  </div>
                )}

                <div ref={messagesEndRef} />
              </div>
            )}

            {/* VIEW 2: Document Library Management (Up to 3 Documents) */}
            {viewMode === 'library' && (
              <div className="flex-1 flex flex-col space-y-5 overflow-y-auto px-margin-desktop py-space-md min-h-0">
                {/* Header & Quota Information */}
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-[#f8f9fa] dark:bg-[#13161f] border border-[#dadce0] dark:border-[#242938] p-4 rounded-xl transition-colors duration-200">
                  <div>
                    <h2 className="font-title-sm text-title-sm text-[#202124] dark:text-[#f1f3f7] font-semibold flex items-center gap-2">
                      <span className="material-symbols-outlined text-[#1a73e8] dark:text-[#38bdf8] text-[22px]">folder_special</span>
                      Document Library
                    </h2>
                    <p className="font-label-sm text-label-sm text-[#5f6368] dark:text-[#9aa0a6] mt-0.5">
                      Manage your uploaded knowledge base. Up to 3 active documents maintained.
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold border ${documentsList.length >= 3
                      ? 'bg-[#fef7e0] dark:bg-[#2c220f] text-[#b06000] dark:text-[#fcd34d] border-[#fce8b2] dark:border-[#453715]'
                      : 'bg-[#e6f4ea] dark:bg-[#122b1b] text-[#137333] dark:text-[#4ade80] border-[#ceead6] dark:border-[#1d472c]'
                      }`}>
                      <span className={`w-2 h-2 rounded-full ${documentsList.length >= 3 ? 'bg-[#f29900]' : 'bg-[#34a853]'}`} />
                      {documentsList.length} of 3 Slots Used
                    </span>
                    <button
                      type="button"
                      onClick={() => loadDocuments()}
                      disabled={loadingDocuments}
                      className="p-1.5 rounded-lg border border-[#dadce0] dark:border-[#2e3547] bg-white dark:bg-[#1c202d] hover:bg-[#f1f3f4] dark:hover:bg-[#252c3e] text-[#5f6368] dark:text-[#9aa0a6] hover:text-[#202124] dark:hover:text-[#f1f3f7] transition-colors cursor-pointer"
                      title="Refresh library"
                    >
                      <span className={`material-symbols-outlined text-[18px] ${loadingDocuments ? 'animate-spin' : ''}`}>refresh</span>
                    </button>
                  </div>
                </div>

                {/* Document Cards List (Up to 3 Documents) */}
                {documentsList.length > 0 && (
                  <div className="grid grid-cols-1 gap-3.5">
                    {documentsList.map((doc, idx) => {
                      const isActive = document && (document.id === doc.id || document.document_id === doc.id);
                      return (
                        <div
                          key={doc.id || idx}
                          className={`p-4 sm:p-5 rounded-2xl border transition-all bg-white dark:bg-[#151821] flex flex-col md:flex-row md:items-center justify-between gap-4 ${isActive
                            ? 'border-[#1a73e8] dark:border-[#38bdf8] shadow-[0_2px_8px_rgba(26,115,232,0.15)] ring-1 ring-[#1a73e8]/20'
                            : 'border-[#dadce0] dark:border-[#242938] hover:border-[#b8bdc7] dark:hover:border-[#384156] shadow-xs'
                            }`}
                        >
                          <div className="flex items-start gap-3.5 min-w-0 flex-1">
                            <div className="w-11 h-11 rounded-xl bg-[#fce8e6] dark:bg-[#321c1f] text-[#ea4335] dark:text-[#f87171] border border-[#fad2cf] dark:border-[#452226] flex items-center justify-center shrink-0">
                              <span className="material-symbols-outlined text-[24px]">picture_as_pdf</span>
                            </div>
                            <div className="min-w-0 flex-1">
                              <div className="flex items-center gap-2 flex-wrap">
                                <h3 className="font-title-sm text-title-sm text-[#202124] dark:text-[#f1f3f7] font-semibold truncate" title={doc.title || doc.filename}>
                                  {doc.title || doc.filename}
                                </h3>
                                {isActive && (
                                  <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-[#e8f0fe] dark:bg-[#1e2a44] text-[#1967d2] dark:text-[#8ab4f8] text-[11px] font-semibold border border-[#d2e3fc] dark:border-[#28385a]">
                                    <span className="w-1.5 h-1.5 rounded-full bg-[#1a73e8] dark:bg-[#38bdf8]" />
                                    Active in Workbench
                                  </span>
                                )}
                                <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-medium border ${doc.status === 'COMPLETED'
                                  ? 'bg-[#e6f4ea] dark:bg-[#122b1b] text-[#137333] dark:text-[#4ade80] border-[#ceead6] dark:border-[#1d472c]'
                                  : doc.status === 'ERROR'
                                    ? 'bg-[#fce8e6] dark:bg-[#321c1f] text-[#c5221f] dark:text-[#f87171] border-[#fad2cf] dark:border-[#452226]'
                                    : 'bg-[#e8f0fe] dark:bg-[#1e2a44] text-[#1967d2] dark:text-[#8ab4f8] border-[#d2e3fc] dark:border-[#28385a]'
                                  }`}>
                                  <span className="material-symbols-outlined text-[13px]">
                                    {doc.status === 'COMPLETED' ? 'check_circle' : doc.status === 'ERROR' ? 'error' : 'refresh'}
                                  </span>
                                  {doc.status === 'COMPLETED' ? 'Ready' : doc.status || 'Processing'}
                                </span>
                              </div>

                              <div className="flex items-center gap-2.5 text-[12px] text-[#5f6368] dark:text-[#9aa0a6] mt-1.5 flex-wrap font-body-sm">
                                <span>{doc.file_size ? `${(doc.file_size / (1024 * 1024)).toFixed(2)} MB` : '< 0.1 MB'}</span>
                                <span>•</span>
                                <span>{doc.page_count || 1} Pages</span>
                                <span>•</span>
                                <span>{doc.total_chunks || 0} chunks</span>
                                <span>•</span>
                                <span>{doc.pii_detected ? 'PII Redacted' : 'Clean'}</span>
                                <span>•</span>
                                <span className="text-[#80868b] dark:text-[#64748b]">
                                  {doc.created_at ? new Date(doc.created_at).toLocaleDateString([], { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }) : 'Recently'}
                                </span>
                              </div>
                            </div>
                          </div>

                          {/* Card Action Buttons */}
                          <div className="flex items-center gap-2 shrink-0 flex-wrap justify-end">
                            {!isActive ? (
                              <button
                                type="button"
                                onClick={() => handleSelectActiveDocument(doc)}
                                className="px-3 py-1.5 rounded-lg bg-[#1a73e8] hover:bg-[#1557b0] text-white text-body-sm font-medium transition-colors flex items-center gap-1.5 shadow-xs cursor-pointer"
                              >
                                <span className="material-symbols-outlined text-[16px]">chat</span>
                                <span>Open in Chat</span>
                              </button>
                            ) : (
                              <button
                                type="button"
                                onClick={() => setViewMode('chat')}
                                className="px-3 py-1.5 rounded-lg bg-[#e8f0fe] dark:bg-[#1e2a44] text-[#1967d2] dark:text-[#8ab4f8] text-body-sm font-medium transition-colors flex items-center gap-1.5 border border-[#d2e3fc] dark:border-[#28385a] cursor-pointer"
                              >
                                <span className="material-symbols-outlined text-[16px]">arrow_forward</span>
                                <span>Go to Chat</span>
                              </button>
                            )}

                            <button
                              type="button"
                              onClick={() => handleDeleteDocument(doc.id || doc.document_id, doc.filename || doc.title)}
                              className="p-1.5 rounded-lg bg-white dark:bg-[#1c202d] hover:bg-[#fce8e6] dark:hover:bg-[#321c1f] text-[#5f6368] dark:text-[#9aa0a6] hover:text-[#c5221f] dark:hover:text-[#f87171] border border-[#dadce0] dark:border-[#2e3547] hover:border-[#fad2cf] dark:hover:border-[#452226] transition-colors cursor-pointer"
                              title="Delete document"
                            >
                              <span className="material-symbols-outlined text-[18px]">delete</span>
                            </button>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}

                {/* Upload Section in Library */}
                <div
                  onDragOver={handleDragOver}
                  onDragLeave={handleDragLeave}
                  onDrop={handleDrop}
                  className={`w-full bg-[#ffffff] dark:bg-[#151821] border-2 border-dashed rounded-2xl p-5 text-center flex flex-col items-center justify-center transition-all shadow-xs ${isDragOver ? 'border-[#1a73e8] dark:border-[#38bdf8] bg-[#e8f0fe]/20 dark:bg-[#1e2a44]/20' : 'border-[#dadce0] dark:border-[#2e3547] hover:border-[#1a73e8] dark:hover:border-[#38bdf8]'
                    }`}
                >
                  <input
                    type="file"
                    ref={fileInputRef}
                    id="pdf-file-upload"
                    accept="application/pdf,.pdf"
                    className="block w-px h-px opacity-0 overflow-hidden"
                    onChange={(e) => {
                      if (e.target.files && e.target.files[0]) {
                        handleFileUpload(e.target.files[0]);
                      }
                    }}
                  />

                  {uploading ? (
                    <div className="w-full max-w-md flex flex-col gap-3 my-2">
                      <div className="flex justify-between items-center text-label-sm font-medium text-[#202124] dark:text-[#f1f3f7]">
                        <span>{processingStage}</span>
                        <span>{uploadProgress}%</span>
                      </div>
                      <div className="w-full h-2 bg-[#dadce0] dark:bg-[#2e3547] rounded-full overflow-hidden">
                        <div
                          className="h-full bg-[#1a73e8] transition-all duration-300"
                          style={{ width: `${uploadProgress}%` }}
                        />
                      </div>
                      <span className="text-xs text-[#5f6368] dark:text-[#9aa0a6]">Multi-phase pipeline active: Clean PII → Embeddings → pgvector</span>
                    </div>
                  ) : (
                    <div className="flex flex-col items-center gap-2.5">
                      <div className="flex flex-col sm:flex-row items-center gap-2.5">
                        <button
                          type="button"
                          onClick={() => fileInputRef.current?.click()}
                          className="px-4 py-2 rounded-lg bg-[#1a73e8] hover:bg-[#1557b0] text-white text-body-sm font-medium transition-all shadow-xs flex items-center gap-2 cursor-pointer"
                        >
                          <span className="material-symbols-outlined text-[18px]">folder_open</span>
                          <span>Upload PDF Document</span>
                        </button>
                        <span className="text-body-sm text-[#80868b] dark:text-[#64748b]">or drag and drop PDF here</span>
                      </div>

                      {documentsList.length >= 3 ? (
                        <p className="text-[12px] text-[#b06000] dark:text-[#fcd34d] bg-[#fef7e0] dark:bg-[#2c220f] px-3 py-1 rounded-md border border-[#fce8b2] dark:border-[#453715] flex items-center gap-1.5">
                          <span className="material-symbols-outlined text-[15px]">info</span>
                          Capacity full (3 of 3). Uploading an additional PDF will automatically replace your oldest document.
                        </p>
                      ) : (
                        <p className="text-[12px] text-[#5f6368] dark:text-[#9aa0a6]">
                          {3 - documentsList.length} of 3 document slots available. Supported: text-based PDF up to {systemInfo?.ai?.validation?.max_file_size_mb || 20}MB.
                        </p>
                      )}
                    </div>
                  )}

                  {uploadError && (
                    <div className="mt-3 p-2.5 rounded-lg bg-[#fce8e6] dark:bg-[#321c1f] border border-[#fad2cf] dark:border-[#452226] text-[#c5221f] dark:text-[#f87171] text-body-sm flex items-center gap-2">
                      <span className="material-symbols-outlined text-[18px]">error</span>
                      <span>{uploadError}</span>
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* Bottom Floating Message Input Area: STRICTLY only rendered in Chat / Workbench mode */}
            {viewMode === 'chat' && (
              <div className="shrink-0 px-margin-desktop py-3 bg-white dark:bg-[#0e1017] border-t border-[#f1f3f4] dark:border-[#1e222e] transition-colors duration-200">
                <form
                  onSubmit={(e) => {
                    e.preventDefault();
                    handleSendQuery();
                  }}
                  className="w-full rounded-2xl bg-[#ffffff] dark:bg-[#151821] border border-[#dadce0] dark:border-[#242938] p-space-sm shadow-[0_2px_8px_rgba(60,64,67,0.1)] dark:shadow-none focus-within:border-[#1a73e8] dark:focus-within:border-[#388bfd] transition-all"
                >
                  <div className="flex items-center gap-space-sm px-space-xs">
                    <button
                      type="button"
                      onClick={() => fileInputRef.current?.click()}
                      className="p-2 rounded-xl text-[#5f6368] dark:text-[#9aa0a6] hover:text-[#202124] dark:hover:text-[#f1f3f7] hover:bg-[#f1f3f4] dark:hover:bg-[#222838] transition-colors cursor-pointer"
                      title="Attach or Replace Document"
                    >
                      <span className="material-symbols-outlined text-[22px]">attach_file</span>
                    </button>

                    <input
                      type="text"
                      value={inputQuery}
                      onChange={(e) => setInputQuery(e.target.value)}
                      disabled={chatLoading}
                      placeholder="Ask something about your PDF... (e.g. 'Summarize Section 2.1 specifications')"
                      className="flex-1 bg-transparent border-0 text-[#202124] dark:text-[#f1f3f7] placeholder:text-[#80868b] dark:placeholder:text-[#64748b] focus:outline-none font-body-md text-body-md py-2 min-w-0"
                    />

                    <button
                      type="button"
                      className="p-2 rounded-xl text-[#5f6368] dark:text-[#9aa0a6] hover:text-[#202124] dark:hover:text-[#f1f3f7] hover:bg-[#f1f3f4] dark:hover:bg-[#222838] transition-colors hidden sm:flex cursor-pointer"
                      title="Voice Input"
                    >
                      <span className="material-symbols-outlined text-[22px]">mic</span>
                    </button>

                    <button
                      type="submit"
                      disabled={chatLoading || !inputQuery.trim()}
                      className="h-10 px-space-md rounded-xl bg-[#1a73e8] hover:bg-[#1557b0] text-white flex items-center justify-center gap-1.5 transition-all shadow-sm font-medium text-body-sm shrink-0 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
                    >
                      <span>Send</span>
                      <span className="material-symbols-outlined text-[18px]">arrow_upward</span>
                    </button>
                  </div>
                </form>

                <div className="flex items-center justify-between px-space-xs mt-1.5">
                  <p className="font-label-sm text-label-sm text-[#80868b] dark:text-[#64748b] flex items-center gap-1.5">
                    <span className="material-symbols-outlined text-[14px] text-[#1a73e8] dark:text-[#38bdf8]">schedule</span>
                    <span>Session-based conversation • Messages reset on reload • Citations verified</span>
                  </p>
                  <span className="hidden md:inline font-citation-ref text-citation-ref text-[#80868b] dark:text-[#64748b]">
                    Press ⏎ to submit
                  </span>
                </div>
              </div>
            )}
          </div>

          {/* ==================== RIGHT STATIC AI / LLM PANEL ==================== */}
          <div className="lg:col-span-4 h-[calc(100vh-4rem)] overflow-y-auto bg-[#f8f9fa] dark:bg-[#111318] p-margin-desktop flex flex-col space-y-space-lg shrink-0 border-l border-[#dadce0] dark:border-[#222631] transition-colors duration-200">
            {/* Panel Header */}
            <div className="border-b border-[#dadce0] dark:border-[#222631] pb-space-md">
              <div className="flex items-center gap-2 mb-1">
                <span className="material-symbols-outlined text-[#1a73e8] dark:text-[#38bdf8] text-[22px]">neurology</span>
                <h2 className="font-headline-md text-headline-md text-[#202124] dark:text-[#f1f3f7] font-semibold">
                  AI used by this service
                </h2>
              </div>
              <p className="font-body-sm text-body-sm text-[#5f6368] dark:text-[#9aa0a6]">
                Models orchestrating neural indexing, semantic search, and streaming RAG synthesis.
              </p>
            </div>

            {/* Provider & Model Cards */}
            <div className="rounded-xl bg-[#ffffff] dark:bg-[#151821] border border-[#dadce0] dark:border-[#242938] p-space-sm shadow-[0_1px_2px_rgba(60,64,67,0.06)] dark:shadow-none divide-y divide-[#f1f3f4] dark:divide-[#1e222e] transition-colors duration-200">
              {/* Primary LLM */}
              <div className="py-2 px-1 flex items-center justify-between">
                <div className="flex items-center gap-2.5 min-w-0">
                  <div className="w-7 h-7 rounded-lg bg-[#e8f0fe] dark:bg-[#1e2a44] flex items-center justify-center text-[#1a73e8] dark:text-[#38bdf8] shrink-0">
                    <span className="material-symbols-outlined text-[16px]">smart_toy</span>
                  </div>
                  <div className="min-w-0">
                    <div className="flex items-center gap-2">
                      <h3 className="font-title-sm text-[13px] text-[#202124] dark:text-[#f1f3f7] font-semibold truncate">
                        {primaryModel}
                      </h3>
                      <span className="text-[11px] text-[#5f6368] dark:text-[#9aa0a6]">{primaryProvider}</span>
                    </div>
                    <span className="font-label-sm text-[11px] text-[#80868b] dark:text-[#64748b] block truncate">
                      Primary Model · 340ms avg
                    </span>
                  </div>
                </div>
                <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-[#e6f4ea] dark:bg-[#122b1b] text-[#137333] dark:text-[#4ade80] font-label-sm text-[11px] border border-[#ceead6] dark:border-[#1d472c] font-medium shrink-0">
                  <span className="w-1.5 h-1.5 rounded-full bg-[#34a853] dark:bg-[#4ade80]" />
                  Available
                </span>
              </div>

              {/* Embeddings */}
              <div className="py-2 px-1 flex items-center justify-between">
                <div className="flex items-center gap-2.5 min-w-0">
                  <div className="w-7 h-7 rounded-lg bg-[#e8f0fe] dark:bg-[#1e2a44] flex items-center justify-center text-[#1a73e8] dark:text-[#38bdf8] shrink-0">
                    <span className="material-symbols-outlined text-[16px]">scatter_plot</span>
                  </div>
                  <div className="min-w-0">
                    <div className="flex items-center gap-2">
                      <h3 className="font-title-sm text-[13px] text-[#202124] dark:text-[#f1f3f7] font-semibold truncate">
                        {embedModel}
                      </h3>
                      <span className="text-[11px] text-[#5f6368] dark:text-[#9aa0a6]">{embedDim}d</span>
                    </div>
                    <span className="font-label-sm text-[11px] text-[#80868b] dark:text-[#64748b] block truncate">
                      Neural Embeddings · Cosine
                    </span>
                  </div>
                </div>
                <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-[#e6f4ea] dark:bg-[#122b1b] text-[#137333] dark:text-[#4ade80] font-label-sm text-[11px] border border-[#ceead6] dark:border-[#1d472c] font-medium shrink-0">
                  <span className="w-1.5 h-1.5 rounded-full bg-[#34a853] dark:bg-[#4ade80]" />
                  Active
                </span>
              </div>
            </div>

            {/* Live Vector Pipeline Parameters */}
            <div className="p-space-md rounded-xl bg-white dark:bg-[#151821] border border-[#dadce0] dark:border-[#242938] flex flex-col gap-2.5 shadow-xs transition-colors duration-200">
              <span className="font-label-md text-label-md text-[#202124] dark:text-[#f1f3f7] font-semibold flex items-center gap-1.5">
                <span className="material-symbols-outlined text-[18px] text-[#1a73e8] dark:text-[#38bdf8]">tune</span>
                Hybrid Retrieval Tuning
              </span>
              <div className="grid grid-cols-2 gap-2 text-label-sm text-[#5f6368]">
                <div className="p-2 bg-[#f8f9fa] dark:bg-[#1a1e2a] rounded-lg border border-[#e8eaed] dark:border-[#282f40]">
                  <span className="block text-[10px] text-[#80868b] dark:text-[#64748b]">Vector Weight</span>
                  <span className="font-semibold text-[#202124] dark:text-[#f1f3f7]">70% Cosine</span>
                </div>
                <div className="p-2 bg-[#f8f9fa] dark:bg-[#1a1e2a] rounded-lg border border-[#e8eaed] dark:border-[#282f40]">
                  <span className="block text-[10px] text-[#80868b] dark:text-[#64748b]">Keyword Weight</span>
                  <span className="font-semibold text-[#202124] dark:text-[#f1f3f7]">30% BM25</span>
                </div>
                <div className="p-2 bg-[#f8f9fa] dark:bg-[#1a1e2a] rounded-lg border border-[#e8eaed] dark:border-[#282f40]">
                  <span className="block text-[10px] text-[#80868b] dark:text-[#64748b]">Top-k Candidates</span>
                  <span className="font-semibold text-[#202124] dark:text-[#f1f3f7]">5 Chunks</span>
                </div>
                <div className="p-2 bg-[#f8f9fa] dark:bg-[#1a1e2a] rounded-lg border border-[#e8eaed] dark:border-[#282f40]">
                  <span className="block text-[10px] text-[#80868b] dark:text-[#64748b]">PII Scrubbing</span>
                  <span className="font-semibold text-[#137333] dark:text-[#4ade80]">Active Filter</span>
                </div>
              </div>
            </div>

            {/* Mini Telemetry / Embeddings Health Vector Graph */}
            <div className="p-space-md rounded-xl bg-[#ffffff] dark:bg-[#151821] border border-[#dadce0] dark:border-[#242938] flex items-center justify-between shadow-xs transition-colors duration-200">
              <div>
                <span className="font-label-sm text-label-sm text-[#80868b] dark:text-[#64748b] block">Active Chunk Cache</span>
                <span className="font-title-sm text-title-sm text-[#202124] dark:text-[#f1f3f7] font-semibold">98.8% Hit Rate</span>
              </div>
              <div className="h-10 w-24 flex items-end gap-1">
                <div className="w-2 bg-[#d2e3fc] dark:bg-[#1e2a44] h-4 rounded-xs" />
                <div className="w-2 bg-[#8ab4f8] dark:bg-[#2563eb] h-6 rounded-xs" />
                <div className="w-2 bg-[#d2e3fc] dark:bg-[#1e2a44] h-5 rounded-xs" />
                <div className="w-2 bg-[#8ab4f8] dark:bg-[#2563eb] h-8 rounded-xs" />
                <div className="w-2 bg-[#1a73e8] dark:text-[#38bdf8] h-9 rounded-xs bg-[#1a73e8]" />
                <div className="w-2 bg-[#34a853] dark:bg-[#22c55e] h-10 rounded-xs" />
              </div>
            </div>

            {/* System Status / Footer inside right panel */}
            <div className="pt-4 mt-auto border-t border-[#dadce0] dark:border-[#222631] text-body-sm font-body-sm text-[#5f6368] dark:text-[#9aa0a6] flex flex-col gap-1.5">
              <div className="flex items-center justify-between text-xs">
                <span className="inline-flex items-center gap-1.5 font-medium text-[#137333] dark:text-[#4ade80]">
                  <span className="w-2 h-2 rounded-full bg-[#34a853] dark:bg-[#4ade80] inline-block animate-pulse" />
                  Operational
                </span>
                <span className="text-[#80868b] dark:text-[#64748b]">Latency: 142ms</span>
              </div>
              <span className="text-[11px] text-[#80868b] dark:text-[#64748b]">
                © 2026 PDF_Intelligence. Vector indexing and neural synthesis pipeline.
              </span>
            </div>
          </div>
        </div>
      </main>

      {/* Modals */}
      {showSummaryModal && (
        <SummaryModal
          document={activeModalDoc || document}
          onClose={() => {
            setShowSummaryModal(false);
            setActiveModalDoc(null);
          }}
        />
      )}

      {showInsightsModal && (
        <InsightsModal
          insights={activeModalDoc ? activeModalInsights : insights}
          document={activeModalDoc || document}
          onClose={() => {
            setShowInsightsModal(false);
            setActiveModalDoc(null);
            setActiveModalInsights(null);
          }}
        />
      )}

      {/* 3-Second Workbench Upload Document Prompt Dialog */}
      {showWorkbenchPrompt && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 dark:bg-black/70 backdrop-blur-xs transition-opacity duration-200">
          <div className="w-full max-w-sm rounded-2xl bg-white dark:bg-[#151821] p-6 shadow-2xl border border-[#dadce0] dark:border-[#242938] relative overflow-hidden transition-transform duration-200">
            {/* Dismiss button */}
            <button
              type="button"
              onClick={() => {
                if (workbenchPromptTimeoutRef.current) clearTimeout(workbenchPromptTimeoutRef.current);
                setShowWorkbenchPrompt(false);
              }}
              className="absolute top-3.5 right-3.5 text-[#5f6368] dark:text-[#9aa0a6] hover:text-[#202124] dark:hover:text-[#f1f3f7] p-1.5 rounded-full hover:bg-[#f1f3f4] dark:hover:bg-[#222838] transition-colors cursor-pointer"
              title="Close"
            >
              <span className="material-symbols-outlined text-[20px]">close</span>
            </button>

            <div className="flex flex-col items-center text-center">
              <div className="w-14 h-14 rounded-2xl bg-[#e8f0fe] dark:bg-[#1e2a44] text-[#1a73e8] dark:text-[#38bdf8] border border-[#d2e3fc] dark:border-[#28385a] flex items-center justify-center mb-3.5 shadow-xs">
                <span className="material-symbols-outlined text-[30px]">upload_file</span>
              </div>

              <h3 className="font-title-sm text-title-sm text-[#202124] dark:text-[#f1f3f7] font-semibold tracking-tight">
                Upload Document to Continue
              </h3>

              <p className="font-body-sm text-body-sm text-[#5f6368] dark:text-[#9aa0a6] mt-2 leading-relaxed">
                Please upload a PDF document first to activate the interactive AI workbench.
              </p>

              <button
                type="button"
                onClick={() => {
                  if (workbenchPromptTimeoutRef.current) clearTimeout(workbenchPromptTimeoutRef.current);
                  setShowWorkbenchPrompt(false);
                  setViewMode('library');
                  fileInputRef.current?.click();
                }}
                className="mt-4 w-full py-2.5 px-4 rounded-xl bg-[#1a73e8] hover:bg-[#1557b0] text-white font-medium text-body-sm transition-all shadow-sm flex items-center justify-center gap-2 cursor-pointer active:scale-98"
              >
                <span className="material-symbols-outlined text-[18px]">add</span>
                <span>Upload PDF Document</span>
              </button>

              {/* 3-Second Countdown Progress Bar */}
              <div className="w-full mt-4 flex items-center gap-2 text-xs text-[#80868b] dark:text-[#64748b]">
                <span className="material-symbols-outlined text-[14px]">timer</span>
                <span>Auto-closing in 3s</span>
                <div className="flex-1 bg-[#f1f3f4] dark:bg-[#222838] h-1.5 rounded-full overflow-hidden ml-1">
                  <div
                    className="h-full bg-[#1a73e8] dark:bg-[#38bdf8] rounded-full"
                    style={{
                      animation: 'shrinkProgress 3s linear forwards',
                    }}
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
