import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { useNotification } from '../contexts/NotificationContext';

const SendBulkMessageModal = ({ isOpen, onClose, selectedContacts, contacts, onComplete }) => {
  const { token } = useAuth();
  const { showNotification } = useNotification();
  
  const [templates, setTemplates] = useState([]);
  const [sessions, setSessions] = useState([]);
  const [loading, setLoading] = useState(false);
  const [selectedTemplate, setSelectedTemplate] = useState(null);
  const [selectedSession, setSelectedSession] = useState('');
  const [delayBetween, setDelayBetween] = useState(2);
  const [randomDelay, setRandomDelay] = useState(true);
  const [customVariables, setCustomVariables] = useState({});
  const [previewMessage, setPreviewMessage] = useState('');
  const [step, setStep] = useState(1); // 1: Setup, 2: Preview, 3: Sending
  const [sendingJob, setSendingJob] = useState(null);
  
  const selectedContactsList = contacts.filter(c => selectedContacts.includes(c.id));
  
  useEffect(() => {
    if (isOpen) {
      fetchTemplates();
      fetchSessions();
    }
  }, [isOpen]);
  
  const fetchTemplates = async () => {
    try {
      const response = await fetch('/api/message-templates?is_active=true', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to fetch templates');
      
      const data = await response.json();
      setTemplates(data || []);
    } catch (error) {
      showNotification('Failed to load templates', 'error');
      console.error('Error fetching templates:', error);
    }
  };
  
  const fetchSessions = async () => {
    try {
      const response = await fetch('/api/sessions', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to fetch sessions');
      
      const data = await response.json();
      setSessions(data.filter(s => s.connected && s.logged_in) || []);
    } catch (error) {
      showNotification('Failed to load sessions', 'error');
      console.error('Error fetching sessions:', error);
    }
  };
  
  const handleTemplateChange = async (templateId) => {
    const template = templates.find(t => t.id === parseInt(templateId));
    setSelectedTemplate(template);
    
    if (template) {
      // Initialize custom variables
      const variables = {};
      if (template.variables) {
        template.variables.forEach(variable => {
          variables[variable.name] = variable.default_value || '';
        });
      }
      setCustomVariables(variables);
      
      // Generate preview
      await generatePreview(template, variables);
    } else {
      setPreviewMessage('');
    }
  };
  
  const generatePreview = async (template, variables = {}) => {
    if (!template || selectedContactsList.length === 0) return;
    
    try {
      const response = await fetch('/api/message-templates/preview', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          template_id: template.id,
          contact_id: selectedContactsList[0].id,
          variables: variables
        })
      });
      
      if (!response.ok) throw new Error('Failed to generate preview');
      
      const data = await response.json();
      setPreviewMessage(data.content);
    } catch (error) {
      console.error('Error generating preview:', error);
      setPreviewMessage(template.content);
    }
  };
  
  const handleVariableChange = async (name, value) => {
    const newVariables = { ...customVariables, [name]: value };
    setCustomVariables(newVariables);
    
    if (selectedTemplate) {
      await generatePreview(selectedTemplate, newVariables);
    }
  };
  
  const handleSendMessages = async () => {
    if (!selectedTemplate || !selectedSession || selectedContactsList.length === 0) {
      showNotification('Please complete all required fields', 'error');
      return;
    }
    
    try {
      setLoading(true);
      setStep(3);
      
      const response = await fetch('/api/bulk-messages', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          session_id: selectedSession,
          template_id: selectedTemplate.id,
          contact_ids: selectedContacts,
          delay_between: delayBetween,
          random_delay: randomDelay,
          variables: customVariables
        })
      });
      
      if (!response.ok) throw new Error('Failed to start bulk messaging');
      
      const data = await response.json();
      setSendingJob(data);
      
      showNotification('Bulk messaging started successfully!', 'success');
      
      // Poll job status
      pollJobStatus(data.job_id);
      
    } catch (error) {
      showNotification('Failed to send messages', 'error');
      console.error('Error sending messages:', error);
      setStep(2);
    } finally {
      setLoading(false);
    }
  };
  
  const pollJobStatus = async (jobId) => {
    const interval = setInterval(async () => {
      try {
        const response = await fetch(`/api/bulk-messages/${jobId}`, {
          headers: {
            'Authorization': `Bearer ${token}`
          }
        });
        
        if (!response.ok) throw new Error('Failed to get job status');
        
        const job = await response.json();
        setSendingJob(job);
        
        if (['completed', 'failed', 'cancelled'].includes(job.status)) {
          clearInterval(interval);
          if (job.status === 'completed') {
            setTimeout(() => {
              onComplete();
            }, 2000);
          }
        }
      } catch (error) {
        console.error('Error polling job status:', error);
        clearInterval(interval);
      }
    }, 2000);
    
    // Clear interval after 5 minutes
    setTimeout(() => clearInterval(interval), 300000);
  };
  
  const calculateEstimatedTime = () => {
    if (!selectedContactsList.length || !delayBetween) return '0 seconds';
    
    const baseTime = (selectedContactsList.length - 1) * delayBetween;
    const withRandomDelay = randomDelay ? baseTime * 1.3 : baseTime;
    const totalSeconds = withRandomDelay + (selectedContactsList.length * 2); // 2s per message
    
    if (totalSeconds < 60) return `${Math.round(totalSeconds)} seconds`;
    if (totalSeconds < 3600) return `${Math.round(totalSeconds / 60)} minutes`;
    return `${Math.round(totalSeconds / 3600)} hours`;
  };
  
  if (!isOpen) return null;
  
  return (
    <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto">
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-medium text-gray-900">
              Send Bulk Message ({selectedContactsList.length} contacts)
            </h3>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          
          {/* Progress Steps */}
          <div className="mt-4">
            <div className="flex items-center justify-center">
              <div className={`flex items-center ${step >= 1 ? 'text-primary-600' : 'text-gray-400'}`}>
                <div className={`w-8 h-8 rounded-full flex items-center justify-center border-2 ${
                  step >= 1 ? 'border-primary-600 bg-primary-50' : 'border-gray-300'
                }`}>
                  {step > 1 ? (
                    <svg className="w-5 h-5 text-primary-600" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                    </svg>
                  ) : (
                    <span className="text-sm font-medium">1</span>
                  )}
                </div>
                <span className="ml-2 text-sm font-medium">Setup</span>
              </div>
              
              <div className={`w-16 h-0.5 mx-4 ${step >= 2 ? 'bg-primary-600' : 'bg-gray-300'}`}></div>
              
              <div className={`flex items-center ${step >= 2 ? 'text-primary-600' : 'text-gray-400'}`}>
                <div className={`w-8 h-8 rounded-full flex items-center justify-center border-2 ${
                  step >= 2 ? 'border-primary-600 bg-primary-50' : 'border-gray-300'
                }`}>
                  {step > 2 ? (
                    <svg className="w-5 h-5 text-primary-600" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                    </svg>
                  ) : (
                    <span className="text-sm font-medium">2</span>
                  )}
                </div>
                <span className="ml-2 text-sm font-medium">Preview</span>
              </div>
              
              <div className={`w-16 h-0.5 mx-4 ${step >= 3 ? 'bg-primary-600' : 'bg-gray-300'}`}></div>
              
              <div className={`flex items-center ${step >= 3 ? 'text-primary-600' : 'text-gray-400'}`}>
                <div className={`w-8 h-8 rounded-full flex items-center justify-center border-2 ${
                  step >= 3 ? 'border-primary-600 bg-primary-50' : 'border-gray-300'
                }`}>
                  <span className="text-sm font-medium">3</span>
                </div>
                <span className="ml-2 text-sm font-medium">Send</span>
              </div>
            </div>
          </div>
        </div>
        
        <div className="p-6">
          {step === 1 && (
            <>
              {/* Setup Step */}
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {/* Left Column */}
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Message Template *
                    </label>
                    <select
                      value={selectedTemplate?.id || ''}
                      onChange={(e) => handleTemplateChange(e.target.value)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                    >
                      <option value="">Select a template</option>
                      {templates.map(template => (
                        <option key={template.id} value={template.id}>
                          {template.name} ({template.type})
                        </option>
                      ))}
                    </select>
                  </div>
                  
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      WhatsApp Session *
                    </label>
                    <select
                      value={selectedSession}
                      onChange={(e) => setSelectedSession(e.target.value)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                    >
                      <option value="">Select a session</option>
                      {sessions.map(session => (
                        <option key={session.id} value={session.id}>
                          {session.name || session.id} ({session.actual_phone})
                        </option>
                      ))}
                    </select>
                  </div>
                  
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Delay Between Messages (seconds)
                    </label>
                    <input
                      type="number"
                      min="1"
                      max="300"
                      value={delayBetween}
                      onChange={(e) => setDelayBetween(parseInt(e.target.value) || 1)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                    />
                    <p className="mt-1 text-sm text-gray-500">
                      Recommended: 2-5 seconds to avoid spam detection
                    </p>
                  </div>
                  
                  <div>
                    <label className="flex items-center">
                      <input
                        type="checkbox"
                        checked={randomDelay}
                        onChange={(e) => setRandomDelay(e.target.checked)}
                        className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                      />
                      <span className="ml-2 text-sm text-gray-700">Add random delay variation (±30%)</span>
                    </label>
                    <p className="mt-1 text-sm text-gray-500">
                      Makes the sending pattern more natural
                    </p>
                  </div>
                  
                  {/* Custom Variables */}
                  {selectedTemplate && selectedTemplate.variables && selectedTemplate.variables.length > 0 && (
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Custom Variables
                      </label>
                      <div className="space-y-2">
                        {selectedTemplate.variables.map(variable => (
                          <div key={variable.name}>
                            <label className="block text-xs text-gray-600">
                              {variable.placeholder} {variable.required && '*'}
                            </label>
                            <input
                              type="text"
                              value={customVariables[variable.name] || ''}
                              onChange={(e) => handleVariableChange(variable.name, e.target.value)}
                              placeholder={variable.default_value}
                              className="w-full px-2 py-1 border border-gray-300 rounded text-sm focus:ring-primary-500 focus:border-primary-500"
                            />
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
                
                {/* Right Column */}
                <div className="space-y-4">
                  {/* Recipients Preview */}
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Recipients ({selectedContactsList.length})
                    </label>
                    <div className="border border-gray-300 rounded-md max-h-40 overflow-y-auto">
                      {selectedContactsList.slice(0, 10).map(contact => (
                        <div key={contact.id} className="px-3 py-2 border-b border-gray-100 last:border-b-0">
                          <div className="text-sm font-medium text-gray-900">{contact.name}</div>
                          <div className="text-xs text-gray-500">{contact.phone}</div>
                        </div>
                      ))}
                      {selectedContactsList.length > 10 && (
                        <div className="px-3 py-2 text-sm text-gray-500">
                          ... and {selectedContactsList.length - 10} more contacts
                        </div>
                      )}
                    </div>
                  </div>
                  
                  {/* Message Preview */}
                  {previewMessage && (
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Message Preview
                      </label>
                      <div className="border border-gray-300 rounded-md p-3 bg-gray-50">
                        <div className="text-sm whitespace-pre-wrap">{previewMessage}</div>
                      </div>
                      <p className="mt-1 text-xs text-gray-500">
                        Preview with data from first recipient
                      </p>
                    </div>
                  )}
                  
                  {/* Estimated Time */}
                  {selectedContactsList.length > 0 && (
                    <div className="bg-blue-50 rounded-lg p-4">
                      <h4 className="text-sm font-medium text-blue-900 mb-2">Sending Summary</h4>
                      <div className="text-sm text-blue-800 space-y-1">
                        <div>Recipients: {selectedContactsList.length} contacts</div>
                        <div>Estimated time: {calculateEstimatedTime()}</div>
                        <div>Delay: {delayBetween}s {randomDelay ? '(with variation)' : ''}</div>
                      </div>
                    </div>
                  )}
                </div>
              </div>
              
              <div className="mt-6 flex justify-end gap-3">
                <button
                  onClick={onClose}
                  className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  onClick={() => setStep(2)}
                  disabled={!selectedTemplate || !selectedSession}
                  className="px-4 py-2 bg-primary-600 text-white rounded-md text-sm font-medium hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Next: Preview
                </button>
              </div>
            </>
          )}
          
          {step === 2 && (
            <>
              {/* Preview Step */}
              <div className="space-y-6">
                <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                  <div className="flex">
                    <svg className="w-5 h-5 text-yellow-400 mr-3 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                    </svg>
                    <div>
                      <h3 className="text-sm font-medium text-yellow-800">Ready to Send</h3>
                      <p className="mt-1 text-sm text-yellow-700">
                        Please review all details before sending messages to {selectedContactsList.length} contacts.
                        This action cannot be undone.
                      </p>
                    </div>
                  </div>
                </div>
                
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div>
                    <h4 className="text-sm font-medium text-gray-900 mb-3">Sending Details</h4>
                    <dl className="space-y-2 text-sm">
                      <div>
                        <dt className="text-gray-500">Template:</dt>
                        <dd className="text-gray-900">{selectedTemplate?.name}</dd>
                      </div>
                      <div>
                        <dt className="text-gray-500">Session:</dt>
                        <dd className="text-gray-900">
                          {sessions.find(s => s.id === selectedSession)?.name || selectedSession}
                        </dd>
                      </div>
                      <div>
                        <dt className="text-gray-500">Recipients:</dt>
                        <dd className="text-gray-900">{selectedContactsList.length} contacts</dd>
                      </div>
                      <div>
                        <dt className="text-gray-500">Delay between messages:</dt>
                        <dd className="text-gray-900">{delayBetween} seconds {randomDelay ? '(with variation)' : ''}</dd>
                      </div>
                      <div>
                        <dt className="text-gray-500">Estimated time:</dt>
                        <dd className="text-gray-900">{calculateEstimatedTime()}</dd>
                      </div>
                    </dl>
                  </div>
                  
                  <div>
                    <h4 className="text-sm font-medium text-gray-900 mb-3">Message Preview</h4>
                    <div className="border border-gray-300 rounded-lg p-3 bg-gray-50">
                      <div className="text-sm whitespace-pre-wrap">{previewMessage}</div>
                    </div>
                  </div>
                </div>
              </div>
              
              <div className="mt-6 flex justify-end gap-3">
                <button
                  onClick={() => setStep(1)}
                  className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
                >
                  Back
                </button>
                <button
                  onClick={handleSendMessages}
                  disabled={loading}
                  className="px-4 py-2 bg-green-600 text-white rounded-md text-sm font-medium hover:bg-green-700 disabled:opacity-50 flex items-center gap-2"
                >
                  {loading ? (
                    <>
                      <svg className="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      Starting...
                    </>
                  ) : (
                    <>
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
                      </svg>
                      Send Messages
                    </>
                  )}
                </button>
              </div>
            </>
          )}
          
          {step === 3 && sendingJob && (
            <>
              {/* Sending Step */}
              <div className="text-center py-8">
                <div className="mb-6">
                  {sendingJob.status === 'running' ? (
                    <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-blue-100 mb-4">
                      <svg className="w-6 h-6 text-blue-600 animate-spin" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                    </div>
                  ) : sendingJob.status === 'completed' ? (
                    <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100 mb-4">
                      <svg className="h-6 w-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                    </div>
                  ) : (
                    <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100 mb-4">
                      <svg className="h-6 w-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </div>
                  )}
                  
                  <h3 className="text-lg font-medium text-gray-900 mb-2">
                    {sendingJob.status === 'running' ? 'Sending Messages...' :
                     sendingJob.status === 'completed' ? 'Messages Sent Successfully!' :
                     sendingJob.status === 'failed' ? 'Sending Failed' :
                     'Processing...'}
                  </h3>
                  
                  <div className="max-w-sm mx-auto mb-6">
                    <div className="bg-gray-200 rounded-full h-2">
                      <div 
                        className={`h-2 rounded-full transition-all duration-300 ${
                          sendingJob.status === 'completed' ? 'bg-green-500' :
                          sendingJob.status === 'failed' ? 'bg-red-500' :
                          'bg-blue-500'
                        }`}
                        style={{ 
                          width: `${sendingJob.progress ? 
                            ((sendingJob.progress.sent + sendingJob.progress.failed) / sendingJob.progress.total) * 100 
                            : 0}%` 
                        }}
                      ></div>
                    </div>
                  </div>
                </div>
                
                {sendingJob.progress && (
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
                    <div className="bg-blue-50 rounded-lg p-4">
                      <div className="text-2xl font-bold text-blue-600">{sendingJob.progress.total}</div>
                      <div className="text-sm text-blue-800">Total</div>
                    </div>
                    <div className="bg-green-50 rounded-lg p-4">
                      <div className="text-2xl font-bold text-green-600">{sendingJob.progress.sent}</div>
                      <div className="text-sm text-green-800">Sent</div>
                    </div>
                    <div className="bg-red-50 rounded-lg p-4">
                      <div className="text-2xl font-bold text-red-600">{sendingJob.progress.failed}</div>
                      <div className="text-sm text-red-800">Failed</div>
                    </div>
                    <div className="bg-yellow-50 rounded-lg p-4">
                      <div className="text-2xl font-bold text-yellow-600">{sendingJob.progress.remaining}</div>
                      <div className="text-sm text-yellow-800">Remaining</div>
                    </div>
                  </div>
                )}
                
                <div className="text-sm text-gray-600 mb-4">
                  Job ID: {sendingJob.id}
                </div>
                
                {sendingJob.status === 'completed' && (
                  <button
                    onClick={onComplete}
                    className="px-6 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700"
                  >
                    Close
                  </button>
                )}
                
                {sendingJob.status === 'running' && (
                  <button
                    onClick={onClose}
                    className="px-6 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700"
                  >
                    Close (Continue in Background)
                  </button>
                )}
                
                {(sendingJob.status === 'failed' || sendingJob.status === 'cancelled') && (
                  <div className="flex justify-center gap-3">
                    <button
                      onClick={() => setStep(1)}
                      className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700"
                    >
                      Try Again
                    </button>
                    <button
                      onClick={onClose}
                      className="px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700"
                    >
                      Close
                    </button>
                  </div>
                )}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default SendBulkMessageModal;