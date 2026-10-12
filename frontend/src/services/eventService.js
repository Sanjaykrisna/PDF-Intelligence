/**
 * eventService.js
 * Centralized API service & Event Bus for PDF_Intelligence application.
 * Handles backend API requests, file uploads with progress, polling, and UI event subscriptions.
 */

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

// Persistent session ID in localStorage for stateless session tracking
const getOrInitSessionId = () => {
  if (typeof window !== 'undefined' && window.localStorage) {
    let sid = window.localStorage.getItem('pdf_intel_session_id');
    if (!sid) {
      sid = typeof crypto !== 'undefined' && crypto.randomUUID
        ? crypto.randomUUID()
        : '00000000-0000-0000-0000-000000000001';
      window.localStorage.setItem('pdf_intel_session_id', sid);
    }
    return sid;
  }
  return '00000000-0000-0000-0000-000000000001';
};
const SESSION_ID = getOrInitSessionId();

class EventService {
  constructor() {
    this.subscribers = {};
  }

  // ==================== EVENT BUS (PUB/SUB) ====================

  /**
   * Subscribe to an event
   * @param {string} event 
   * @param {Function} callback 
   * @returns {Function} unsubscribe function
   */
  subscribe(event, callback) {
    if (!this.subscribers[event]) {
      this.subscribers[event] = [];
    }
    this.subscribers[event].push(callback);
    return () => {
      this.subscribers[event] = this.subscribers[event].filter(cb => cb !== callback);
    };
  }

  /**
   * Emit an event with data to all subscribers
   * @param {string} event 
   * @param {*} data 
   */
  emit(event, data) {
    if (this.subscribers[event]) {
      this.subscribers[event].forEach(callback => {
        try {
          callback(data);
        } catch (err) {
          console.error(`Error in subscriber callback for event ${event}:`, err);
        }
      });
    }
  }

  // ==================== HTTP CORE ====================

  /**
   * Get active ephemeral session ID
   * @returns {string}
   */
  getSessionId() {
    return SESSION_ID;
  }

  async request(endpoint, options = {}) {
    const url = `${API_BASE_URL}${endpoint}`;
    const headers = options.headers || {};

    // Attach active session identifier for stateless session tracking
    headers['X-Session-ID'] = SESSION_ID;

    if (!options.isFormData && !headers['Content-Type']) {
      headers['Content-Type'] = 'application/json';
    }

    const config = {
      ...options,
      headers,
    };

    try {
      const response = await fetch(url, config);

      const rawText = await response.text();
      let parsed;
      try {
        parsed = JSON.parse(rawText);
      } catch {
        parsed = { status: 'E', message: rawText || 'Non-JSON server response' };
      }

      // Check if backend returned structured error string { status: "E", message: "..." }
      if (parsed && parsed.status === 'E') {
        throw new Error(parsed.message || 'Operation failed on server.');
      }

      if (!response.ok) {
        throw new Error(parsed?.message || `HTTP ${response.status}: Request failed`);
      }

      return parsed;
    } catch (err) {
      // Avoid raw technical exposure in UI
      let friendlyMessage = err.message || 'Network error occurred. Please check connection.';
      if (friendlyMessage.includes('Failed to fetch')) {
        friendlyMessage = 'Unable to connect to backend server at ' + API_BASE_URL;
      }
      throw new Error(friendlyMessage);
    }
  }



  // ==================== DOCUMENT & UPLOAD APIS ====================

  /**
   * Upload PDF document with validation and progress tracking
   * @param {File} file 
   * @param {string} title 
   * @param {Function} onProgress (percent: number) => void
   */
  uploadDocument(file, title, onProgress) {
    return new Promise((resolve, reject) => {
      // 1. Client-side PDF validations
      if (!file) {
        return reject(new Error('No file selected for upload.'));
      }
      if (!file.name.toLowerCase().endsWith('.pdf') && file.type !== 'application/pdf') {
        return reject(new Error('Invalid format: Only PDF files (.pdf) are supported.'));
      }
      const maxMB = 20;
      if (file.size > maxMB * 1024 * 1024) {
        return reject(new Error(`File size (${(file.size / (1024 * 1024)).toFixed(1)}MB) exceeds the maximum allowed limit of ${maxMB}MB.`));
      }

      this.emit('document:upload_started', { filename: file.name, size: file.size });

      const formData = new FormData();
      formData.append('file', file);
      formData.append('title', title || file.name);

      const xhr = new XMLHttpRequest();
      xhr.open('POST', `${API_BASE_URL}/api/documents/upload`, true);
      xhr.setRequestHeader('X-Session-ID', SESSION_ID);

      xhr.upload.onprogress = (e) => {
        if (e.lengthComputable) {
          const percent = Math.round((e.loaded / e.total) * 100);
          if (onProgress) onProgress(percent);
          this.emit('document:upload_progress', percent);
        }
      };

      xhr.onload = () => {

        try {
          const res = JSON.parse(xhr.responseText);
          if (res.status === 'S') {
            this.emit('document:upload_success', res);
            resolve(res);
          } else {
            const err = new Error(res.message || 'Upload validation failed.');
            this.emit('document:error', err.message);
            reject(err);
          }
        } catch {
          reject(new Error('Invalid response from server during upload.'));
        }
      };

      xhr.onerror = () => {
        const err = new Error('Network error during file upload.');
        this.emit('document:error', err.message);
        reject(err);
      };

      xhr.send(formData);
    });
  }

  /**
   * Retrieve document summary by ID
   * @param {string} documentId 
   */
  async getDocumentSummary(documentId) {
    return await this.request(`/api/documents/${documentId}/summary`, { method: 'GET' });
  }

  /**
   * Retrieve document analytics & validation insights by ID
   * @param {string} documentId 
   */
  async getDocumentInsights(documentId) {
    return await this.request(`/api/documents/${documentId}/insights`, { method: 'GET' });
  }

  /**
   * Retrieve the user's latest uploaded document
   */
  async getLatestDocument() {
    return await this.request('/api/documents/latest', { method: 'GET' });
  }

  /**
   * Retrieve list of previously uploaded documents for the user (up to 3)
   */
  async getDocuments() {
    return await this.request('/api/documents', { method: 'GET' });
  }

  /**
   * Delete a document by ID (purges DB record, chunks, audits, and storage file)
   * @param {string} documentId 
   */
  async deleteDocument(documentId) {
    try {
      const res = await this.request(`/api/documents/${documentId}`, { method: 'DELETE' });
      this.emit('document:deleted', documentId);
      return res;
    } catch (primaryErr) {
      // Fallback to POST /api/documents/delete if DELETE method is blocked by any proxy/browser
      try {
        const res = await this.request('/api/documents/delete', {
          method: 'POST',
          body: JSON.stringify({ id: documentId, document_id: documentId }),
        });
        this.emit('document:deleted', documentId);
        return res;
      } catch {
        throw primaryErr;
      }
    }
  }

  /**
   * Poll document until background parsing/chunking/embedding is completed
   * @param {string} documentId 
   * @param {Function} onStatusUpdate (status: string, stage: string) => void
   */
  async pollDocumentUntilReady(documentId, onStatusUpdate, maxSeconds = 90) {
    const startTime = Date.now();
    const pollInterval = 1500;

    const stages = [
      'Uploading',
      'Validating',
      'Extracting',
      'Sanitizing',
      'Chunking',
      'Embedding',
      'Ready',
    ];

    let currentStageIndex = 1;

    while ((Date.now() - startTime) < (maxSeconds * 1000)) {
      try {
        const data = await this.getDocumentInsights(documentId);
        if (data && data.status === 'S' && data.insights) {
          const doc = data.insights;
          const status = doc.status;

          // Progress simulated stage advance if still processing
          if (currentStageIndex < 5) {
            currentStageIndex++;
          }

          if (status === 'COMPLETED') {
            if (onStatusUpdate) onStatusUpdate('COMPLETED', 'Ready', doc);
            this.emit('document:ready', doc);
            return doc;
          }

          if (status === 'FAILED' || status === 'ERROR') {
            const auditFailure = doc.validation_checks?.find(c => c.status === 'FAIL')?.message;
            const err = new Error(auditFailure || 'Document processing failed during extraction or embedding.');
            if (onStatusUpdate) onStatusUpdate('FAILED', 'Failed', doc);
            this.emit('document:error', err.message);
            throw err;
          }

          const currentStage = stages[currentStageIndex];
          if (onStatusUpdate) onStatusUpdate(status, currentStage, doc);
          this.emit('document:processing_status', { status, stage: currentStage, doc });
        }
      } catch (err) {
        if (err.message.includes('failed')) throw err;
      }

      await new Promise(res => setTimeout(res, pollInterval));
    }

    throw new Error('Document processing timed out. Please refresh or check server status.');
  }

  // ==================== CHAT & QUESTION ANSWERING ====================

  /**
   * Ask question regarding the document
   * @param {string} documentId 
   * @param {string} question 
   * @param {number} level (1 = Standard, 2 = Deep Reasoning)
   */
  async askQuestion(documentId, question, level = 1) {
    if (!question || !question.trim()) {
      throw new Error('Question text cannot be empty.');
    }
    this.emit('chat:loading', true);
    try {
      const data = await this.request(`/api/documents/${documentId}/ask`, {
        method: 'POST',
        body: JSON.stringify({
          question: question.trim(),
          level,
        }),
      });

      this.emit('chat:answer', data);
      return data;
    } finally {
      this.emit('chat:loading', false);
    }
  }

  // ==================== SYSTEM & AI CAPABILITIES ====================

  /**
   * Get active system info, AI models, embedding dimensions, and fallback configurations
   */
  async getSystemInfo() {
    try {
      const data = await this.request('/api/system/info', { method: 'GET' });
      if (data && data.status === 'S') {
        this.emit('system:info_updated', data);
        return data;
      }
      return null;
    } catch {
      // Return sensible defaults if server system info endpoint is momentarily unreachable
      return {
        status: 'S',
        ai: {
          primary: {
            provider: 'Google Gemini',
            model: 'gemini-3.5-flash-lite',
            status: 'Available',
            latency_avg: '340ms avg',
          },
          fallback: {
            provider: 'Ollama',
            model: 'llama3.2',
            status: 'Ready',
            mode: 'Local AI Fallback',
          },
          embeddings: {
            provider: 'Google Gemini',
            model: 'gemini-embedding-001',
            dimension: 1536,
            status: 'Ready',
          },
          retrieval: {
            hybrid_search: true,
            vector_weight: 0.70,
            keyword_weight: 0.30,
            top_k: 5,
            similarity_threshold: 0.50,
          },
          validation: {
            max_file_size_mb: 20,
            max_pages: 100,
          },
        },
      };
    }
  }
}

export const eventService = new EventService();
export default eventService;
