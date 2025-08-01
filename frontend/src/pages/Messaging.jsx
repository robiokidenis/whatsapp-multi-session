import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { useNotification } from '../contexts/NotificationContext';
import { useAuth } from '../contexts/AuthContext';

const Messaging = () => {
  const { token } = useAuth();
  const { showError, showSuccess, showWarning } = useNotification();
  
  // State
  const [contacts, setContacts] = useState([]);
  const [contactGroups, setContactGroups] = useState([]);
  const [templates, setTemplates] = useState([]);
  const [sessions, setSessions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  
  // Selection state
  const [selectionMode, setSelectionMode] = useState('individual'); // 'individual', 'group', 'all'
  const [selectedContacts, setSelectedContacts] = useState([]);
  const [selectedGroups, setSelectedGroups] = useState([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedGroupFilter, setSelectedGroupFilter] = useState('');
  
  // Message state
  const [message, setMessage] = useState('');
  const [selectedTemplate, setSelectedTemplate] = useState('');
  const [selectedSession, setSelectedSession] = useState('');
  const [messageType, setMessageType] = useState('text'); // 'text', 'media'
  const [mediaFile, setMediaFile] = useState(null);
  const [scheduleTime, setScheduleTime] = useState('');
  
  
  // Fetch contacts
  const fetchContacts = useCallback(async () => {
    try {
      setLoading(true);
      const params = new URLSearchParams({
        page: 1,
        limit: 1000, // Get all contacts for messaging
        ...(searchQuery && { query: searchQuery }),
        ...(selectedGroupFilter && { group_id: selectedGroupFilter })
      });
      
      const response = await fetch(`/api/contacts?${params}`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to fetch contacts');
      
      const data = await response.json();
      setContacts(data.contacts || []);
    } catch (error) {
      showError('Failed to load contacts');
      console.error('Error fetching contacts:', error);
    } finally {
      setLoading(false);
    }
  }, [token, searchQuery, selectedGroupFilter, showError]);
  
  // Fetch contact groups
  const fetchContactGroups = useCallback(async () => {
    try {
      const response = await fetch('/api/contact-groups', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to fetch contact groups');
      
      const data = await response.json();
      setContactGroups(data || []);
    } catch (error) {
      console.error('Error fetching contact groups:', error);
    }
  }, [token]);
  
  // Fetch templates
  const fetchTemplates = useCallback(async () => {
    try {
      const response = await fetch('/api/message-templates', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) {
        console.warn('Templates API not available:', response.status);
        setTemplates([]); // Set empty array if templates not available
        return;
      }
      
      const data = await response.json();
      setTemplates(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error('Error fetching templates:', error);
      setTemplates([]); // Fallback to empty array
    }
  }, [token]);
  
  // Fetch sessions
  const fetchSessions = useCallback(async () => {
    try {
      const response = await fetch('/api/sessions', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to fetch sessions');
      
      const data = await response.json();
      // Handle both response formats
      const sessionList = data.data || data || [];
      
      console.log('Fetched sessions:', sessionList); // Debug log
      
      // Filter sessions - prioritize connected ones but show all available
      const connectedSessions = sessionList.filter(session => 
        (session.status === 'Connected' || session.connected) && session.logged_in
      );
      
      console.log('Connected sessions:', connectedSessions); // Debug log
      
      // If no connected sessions, show all sessions so user can see what's available
      const availableSessions = connectedSessions.length > 0 ? connectedSessions : sessionList;
      setSessions(availableSessions);
      
      // Auto-select first connected session if available
      if (connectedSessions.length > 0 && !selectedSession) {
        setSelectedSession(connectedSessions[0].id);
      }
    } catch (error) {
      console.error('Error fetching sessions:', error);
      showError('Failed to load WhatsApp sessions');
    }
  }, [token, selectedSession, showError]);
  
  // Initial load
  useEffect(() => {
    fetchContacts();
    fetchContactGroups();
    fetchTemplates();
    fetchSessions();
  }, [fetchContacts, fetchContactGroups, fetchTemplates, fetchSessions]);
  
  // Handle individual contact selection
  const handleContactSelection = (contactId) => {
    setSelectedContacts(prev => {
      if (prev.includes(contactId)) {
        return prev.filter(id => id !== contactId);
      }
      return [...prev, contactId];
    });
  };
  
  // Handle group selection
  const handleGroupSelection = (groupId) => {
    setSelectedGroups(prev => {
      if (prev.includes(groupId)) {
        return prev.filter(id => id !== groupId);
      }
      return [...prev, groupId];
    });
  };
  
  // Handle select all contacts
  const handleSelectAllContacts = () => {
    if (selectedContacts.length === contacts.length) {
      setSelectedContacts([]);
    } else {
      setSelectedContacts(contacts.map(c => c.id));
    }
  };
  
  // Handle template selection
  const handleTemplateSelect = (template) => {
    setSelectedTemplate(template.id);
    setMessage(template.content);
    setMessageType(template.type || 'text');
  };
  
  // Get target contacts based on selection mode
  const getTargetContacts = useMemo(() => {
    switch (selectionMode) {
      case 'individual':
        return contacts.filter(c => selectedContacts.includes(c.id));
      case 'group':
        return contacts.filter(c => 
          c.groups?.some(g => selectedGroups.includes(g.id)) ||
          (c.group_id && selectedGroups.includes(c.group_id))
        );
      case 'all':
        return contacts;
      default:
        return [];
    }
  }, [selectionMode, selectedContacts, selectedGroups, contacts]);
  
  // Handle send message
  const handleSendMessage = async () => {
    if (!selectedSession) {
      showError('Please select a WhatsApp session');
      return;
    }
    
    // Check if selected session is connected
    const selectedSessionObj = sessions.find(s => s.id === selectedSession);
    if (selectedSessionObj) {
      const isConnected = (selectedSessionObj.status === 'Connected' || selectedSessionObj.connected) && selectedSessionObj.logged_in;
      if (!isConnected) {
        showError('Selected session is not connected. Please select an online session or connect it first.');
        return;
      }
    }
    
    if (!message.trim()) {
      showError('Please enter a message');
      return;
    }
    
    const targetContacts = getTargetContacts;
    if (targetContacts.length === 0) {
      showError('Please select at least one contact or group');
      return;
    }
    
    if (!selectedTemplate && !message.trim()) {
      showError('Please select a message template or enter a custom message');
      return;
    }
    
    try {
      setSending(true);
      
      const requestData = {
        session_id: selectedSession,
        contact_ids: targetContacts.map(c => c.id),
        delay_between: 5, // 5 seconds delay between messages
        random_delay: false,
        variables: {} // Template variables can be added here
      };
      
      // Add either template_id or message
      if (selectedTemplate) {
        requestData.template_id = parseInt(selectedTemplate);
      } else {
        requestData.message = message;
      }
      
      // Add group ID if selecting by group
      if (selectionMode === 'group' && selectedGroups.length > 0) {
        requestData.group_id = selectedGroups[0]; // Use first selected group
      }
      
      const response = await fetch('/api/bulk-messages', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(requestData)
      });
      
      if (!response.ok) {
        const error = await response.text();
        throw new Error(error || 'Failed to send messages');
      }
      
      const result = await response.json();
      showSuccess(`Bulk messaging job started. Job ID: ${result.job_id}`);
      
      // Reset form
      setMessage('');
      setSelectedTemplate('');
      setSelectedContacts([]);
      setSelectedGroups([]);
      setMediaFile(null);
      setScheduleTime('');
      
    } catch (error) {
      showError(error.message || 'Failed to send messages');
      console.error('Error sending messages:', error);
    } finally {
      setSending(false);
    }
  };
  
  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 mb-2">Messaging Center</h1>
            <p className="text-gray-600">Send bulk messages to your contacts and manage templates</p>
          </div>
          <button
            onClick={() => window.location.href = '/message-queue'}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center gap-2"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
            </svg>
            View Queue
          </button>
        </div>
      </div>
      
      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-white rounded-lg shadow p-4">
          <div className="text-sm text-gray-500">Total Contacts</div>
          <div className="text-2xl font-bold text-gray-900">{contacts.length}</div>
        </div>
        <div className="bg-white rounded-lg shadow p-4">
          <div className="text-sm text-gray-500">Contact Groups</div>
          <div className="text-2xl font-bold text-blue-600">{contactGroups.length}</div>
        </div>
        <div className="bg-white rounded-lg shadow p-4">
          <div className="text-sm text-gray-500">Templates</div>
          <div className="text-2xl font-bold text-green-600">{templates.length}</div>
        </div>
        <div className="bg-white rounded-lg shadow p-4">
          <div className="text-sm text-gray-500">Target Recipients</div>
          <div className="text-2xl font-bold text-purple-600">{getTargetContacts.length}</div>
        </div>
      </div>
      
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column - Recipient Selection */}
        <div className="lg:col-span-1">
          <div className="bg-white rounded-lg shadow">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-medium text-gray-900">Select Recipients</h3>
            </div>
            
            <div className="p-6">
              {/* Selection Mode */}
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-3">Selection Mode</label>
                <div className="space-y-2">
                  <label className="flex items-center">
                    <input
                      type="radio"
                      name="selectionMode"
                      value="individual"
                      checked={selectionMode === 'individual'}
                      onChange={(e) => setSelectionMode(e.target.value)}
                      className="mr-3 text-primary-600 focus:ring-primary-500"
                    />
                    <span className="text-sm">Select Individual Contacts</span>
                  </label>
                  <label className="flex items-center">
                    <input
                      type="radio"
                      name="selectionMode"
                      value="group"
                      checked={selectionMode === 'group'}
                      onChange={(e) => setSelectionMode(e.target.value)}
                      className="mr-3 text-primary-600 focus:ring-primary-500"
                    />
                    <span className="text-sm">Select by Groups</span>
                  </label>
                  <label className="flex items-center">
                    <input
                      type="radio"
                      name="selectionMode"
                      value="all"
                      checked={selectionMode === 'all'}
                      onChange={(e) => setSelectionMode(e.target.value)}
                      className="mr-3 text-primary-600 focus:ring-primary-500"
                    />
                    <span className="text-sm">All Contacts</span>
                  </label>
                </div>
              </div>
              
              {/* Search and Filter */}
              {selectionMode !== 'all' && (
                <div className="mb-6">
                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Search Contacts</label>
                      <input
                        type="text"
                        placeholder="Search by name or phone..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                      />
                    </div>
                    
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Filter by Group</label>
                      <select
                        value={selectedGroupFilter}
                        onChange={(e) => setSelectedGroupFilter(e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                      >
                        <option value="">All Groups</option>
                        {contactGroups.map(group => (
                          <option key={group.id} value={group.id}>{group.name}</option>
                        ))}
                      </select>
                    </div>
                  </div>
                </div>
              )}
              
              {/* Group Selection */}
              {selectionMode === 'group' && (
                <div className="mb-6">
                  <label className="block text-sm font-medium text-gray-700 mb-3">Select Groups</label>
                  <div className="space-y-2 max-h-60 overflow-y-auto">
                    {contactGroups.map(group => (
                      <label key={group.id} className="flex items-center p-2 hover:bg-gray-50 rounded">
                        <input
                          type="checkbox"
                          checked={selectedGroups.includes(group.id)}
                          onChange={() => handleGroupSelection(group.id)}
                          className="mr-3 text-primary-600 focus:ring-primary-500"
                        />
                        <div className="flex items-center flex-1">
                          <div 
                            className="w-3 h-3 rounded-full mr-2"
                            style={{ backgroundColor: group.color }}
                          />
                          <span className="text-sm">{group.name}</span>
                          <span className="text-xs text-gray-500 ml-auto">
                            {contacts.filter(c => 
                              c.groups?.some(g => g.id === group.id) || c.group_id === group.id
                            ).length} contacts
                          </span>
                        </div>
                      </label>
                    ))}
                  </div>
                </div>
              )}
              
              {/* Individual Contact Selection */}
              {selectionMode === 'individual' && (
                <div className="mb-6">
                  <div className="flex items-center justify-between mb-3">
                    <label className="block text-sm font-medium text-gray-700">Select Contacts</label>
                    <button
                      onClick={handleSelectAllContacts}
                      className="text-xs text-primary-600 hover:text-primary-800"
                    >
                      {selectedContacts.length === contacts.length ? 'Deselect All' : 'Select All'}
                    </button>
                  </div>
                  <div className="space-y-2 max-h-80 overflow-y-auto">
                    {loading ? (
                      <div className="text-center py-4">
                        <div className="animate-spin h-6 w-6 border-2 border-primary-600 border-t-transparent rounded-full mx-auto"></div>
                      </div>
                    ) : contacts.length === 0 ? (
                      <div className="text-center py-4 text-gray-500">No contacts found</div>
                    ) : (
                      contacts.map(contact => (
                        <label key={contact.id} className="flex items-center p-2 hover:bg-gray-50 rounded">
                          <input
                            type="checkbox"
                            checked={selectedContacts.includes(contact.id)}
                            onChange={() => handleContactSelection(contact.id)}
                            className="mr-3 text-primary-600 focus:ring-primary-500"
                          />
                          <div className="flex-1 min-w-0">
                            <div className="text-sm font-medium text-gray-900 truncate">
                              {contact.name}
                            </div>
                            <div className="text-xs text-gray-500 truncate">
                              {contact.phone}
                              {contact.company && ` • ${contact.company}`}
                            </div>
                          </div>
                          {contact.groups && contact.groups.length > 0 && (
                            <div className="flex gap-1">
                              {contact.groups.slice(0, 2).map(group => (
                                <div 
                                  key={group.id}
                                  className="w-2 h-2 rounded-full"
                                  style={{ backgroundColor: group.color }}
                                  title={group.name}
                                />
                              ))}
                              {contact.groups.length > 2 && (
                                <span className="text-xs text-gray-400">+{contact.groups.length - 2}</span>
                              )}
                            </div>
                          )}
                        </label>
                      ))
                    )}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
        
        {/* Right Column - Message Composition */}
        <div className="lg:col-span-2">
          <div className="bg-white rounded-lg shadow">
            <div className="px-6 py-4 border-b border-gray-200">
              <div className="flex items-center justify-between">
                <h3 className="text-lg font-medium text-gray-900">Compose Message</h3>
                <button
                  onClick={() => window.location.href = '/templates'}
                  className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors flex items-center gap-2"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  Manage Templates
                </button>
              </div>
            </div>
            
            <div className="p-6">
              {/* Session Selection */}
              <div className="mb-6">
                <div className="flex items-center justify-between mb-2">
                  <label className="block text-sm font-medium text-gray-700">
                    WhatsApp Session <span className="text-red-500">*</span>
                  </label>
                  <button
                    onClick={fetchSessions}
                    className="text-xs text-primary-600 hover:text-primary-800 flex items-center gap-1"
                  >
                    <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                    </svg>
                    Refresh
                  </button>
                </div>
                {sessions.length === 0 ? (
                  <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
                    <p className="text-sm text-yellow-800 mb-2">
                      No WhatsApp sessions found.
                    </p>
                    <p className="text-xs text-yellow-700">
                      Go to <a href="/dashboard" className="underline">Session Management</a> to create and connect a WhatsApp session first.
                    </p>
                  </div>
                ) : (
                  <select
                    value={selectedSession}
                    onChange={(e) => setSelectedSession(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                    required
                  >
                    <option value="">Select a session</option>
                    {sessions.map(session => {
                      const isConnected = (session.status === 'Connected' || session.connected) && session.logged_in;
                      return (
                        <option key={session.id} value={session.id}>
                          {session.name || session.phone || session.id} - {isConnected ? '🟢 Online' : '🔴 Offline'}
                        </option>
                      );
                    })}
                  </select>
                )}
              </div>
              
              {/* Template Selection */}
              <div className="mb-6">
                <div className="flex items-center justify-between mb-2">
                  <label className="block text-sm font-medium text-gray-700">
                    Message Template <span className="text-gray-400">(Optional)</span>
                  </label>
                  {selectedTemplate && (
                    <button
                      onClick={() => {
                        setSelectedTemplate('');
                        setMessageType('text');
                      }}
                      className="text-xs text-primary-600 hover:text-primary-800"
                    >
                      Clear Selection
                    </button>
                  )}
                </div>
                {templates.length === 0 ? (
                  <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg text-sm text-blue-700">
                    No templates available. You can send custom messages directly.
                  </div>
                ) : (
                  <>
                    <p className="text-xs text-gray-500 mb-2">
                      Select a template for quick messaging or leave unselected to write a custom message
                    </p>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                      {templates.slice(0, 4).map(template => (
                    <button
                      key={template.id}
                      onClick={() => handleTemplateSelect(template)}
                      className={`p-3 text-left border rounded-lg transition-colors ${
                        selectedTemplate === template.id
                          ? 'border-primary-500 bg-primary-50'
                          : 'border-gray-200 hover:border-gray-300'
                      }`}
                    >
                      <div className="font-medium text-sm text-gray-900">{template.name}</div>
                      <div className="text-xs text-gray-500 mt-1 truncate">{template.content}</div>
                    </button>
                  ))}
                    </div>
                  </>
                )}
              </div>
              
              {/* Message Type */}
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-3">Message Type</label>
                <div className="flex gap-4">
                  <label className="flex items-center">
                    <input
                      type="radio"
                      name="messageType"
                      value="text"
                      checked={messageType === 'text'}
                      onChange={(e) => setMessageType(e.target.value)}
                      className="mr-2 text-primary-600 focus:ring-primary-500"
                    />
                    <span className="text-sm">Text Message</span>
                  </label>
                  <label className="flex items-center">
                    <input
                      type="radio"
                      name="messageType"
                      value="media"
                      checked={messageType === 'media'}
                      onChange={(e) => setMessageType(e.target.value)}
                      className="mr-2 text-primary-600 focus:ring-primary-500"
                    />
                    <span className="text-sm">Media Message</span>
                  </label>
                </div>
              </div>
              
              {/* Message Content */}
              <div className="mb-6">
                <div className="flex items-center justify-between mb-2">
                  <label className="block text-sm font-medium text-gray-700">
                    Message Content {!selectedTemplate && <span className="text-red-500">*</span>}
                  </label>
                  {!selectedTemplate && (
                    <span className="text-xs text-primary-600 bg-primary-50 px-2 py-1 rounded">
                      Custom Message Mode
                    </span>
                  )}
                  {selectedTemplate && (
                    <span className="text-xs text-gray-500 bg-gray-100 px-2 py-1 rounded">
                      Using Template
                    </span>
                  )}
                </div>
                <textarea
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  rows={6}
                  placeholder={selectedTemplate ? "Message from template (edit as needed)..." : "Type your custom message here... You can use variables like {{name}}, {{company}}, etc."}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                  readOnly={false}
                />
                <div className="mt-2 flex justify-between text-xs text-gray-500">
                  <div>
                    {!selectedTemplate && "Available variables: {{name}}, {{phone}}, {{email}}, {{company}}, {{position}}"}
                  </div>
                  <div>
                    {message.length} characters
                  </div>
                </div>
              </div>
              
              {/* Media Upload */}
              {messageType === 'media' && (
                <div className="mb-6">
                  <label className="block text-sm font-medium text-gray-700 mb-2">Media File</label>
                  <div className="border-2 border-dashed border-gray-300 rounded-lg p-6 text-center">
                    <input
                      type="file"
                      accept="image/*,video/*,audio/*"
                      onChange={(e) => setMediaFile(e.target.files[0])}
                      className="hidden"
                      id="media-upload"
                    />
                    <label htmlFor="media-upload" className="cursor-pointer">
                      <div className="text-gray-400 mb-2">
                        <svg className="mx-auto h-12 w-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                        </svg>
                      </div>
                      <div className="text-sm text-gray-600">
                        {mediaFile ? mediaFile.name : 'Click to upload media file'}
                      </div>
                    </label>
                  </div>
                </div>
              )}
              
              {/* Schedule */}
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-2">Schedule (Optional)</label>
                <input
                  type="datetime-local"
                  value={scheduleTime}
                  onChange={(e) => setScheduleTime(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                />
              </div>
              
              {/* Preview */}
              <div className="mb-6 p-4 bg-gray-50 rounded-lg">
                <div className="text-sm font-medium text-gray-700 mb-2">Preview</div>
                <div className="text-sm text-gray-600">
                  <div><strong>Recipients:</strong> {getTargetContacts.length} contacts</div>
                  <div><strong>Message:</strong> {message || 'No message content'}</div>
                  {scheduleTime && <div><strong>Scheduled:</strong> {new Date(scheduleTime).toLocaleString()}</div>}
                </div>
              </div>
              
              {/* Send Button */}
              <div className="flex justify-end">
                <button
                  onClick={handleSendMessage}
                  disabled={sending || !message.trim() || getTargetContacts.length === 0 || !selectedSession}
                  className="px-6 py-3 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
                >
                  {sending ? (
                    <>
                      <svg className="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      Sending...
                    </>
                  ) : (
                    <>
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
                      </svg>
                      Send Message
                    </>
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Messaging;