import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { useNotification } from '../contexts/NotificationContext';

const AutoReplyModal = ({ isOpen, onClose }) => {
  const { token } = useAuth();
  const { showNotification } = useNotification();
  
  const [autoReplies, setAutoReplies] = useState([]);
  const [sessions, setSessions] = useState([]);
  const [templates, setTemplates] = useState([]);
  const [loading, setLoading] = useState(true);
  const [editingAutoReply, setEditingAutoReply] = useState(null);
  const [showForm, setShowForm] = useState(false);
  
  const triggerTypes = [
    { value: 'keyword', label: 'Keyword Match' },
    { value: 'time_based', label: 'Time Based' },
    { value: 'welcome', label: 'Welcome Message' }
  ];
  
  const timeUnits = [
    { value: 'minutes', label: 'Minutes' },
    { value: 'hours', label: 'Hours' },
    { value: 'days', label: 'Days' }
  ];
  
  useEffect(() => {
    if (isOpen) {
      fetchAutoReplies();
      fetchSessions();
      fetchTemplates();
    }
  }, [isOpen]);
  
  const fetchAutoReplies = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/auto-replies', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to fetch auto replies');
      
      const data = await response.json();
      setAutoReplies(data || []);
    } catch (error) {
      showNotification('Failed to load auto replies', 'error');
      console.error('Error fetching auto replies:', error);
    } finally {
      setLoading(false);
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
      console.error('Error fetching sessions:', error);
    }
  };
  
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
      console.error('Error fetching templates:', error);
    }
  };
  
  const handleSaveAutoReply = async (autoReplyData) => {
    try {
      const url = editingAutoReply 
        ? `/api/auto-replies/${editingAutoReply.id}`
        : '/api/auto-replies';
      
      const response = await fetch(url, {
        method: editingAutoReply ? 'PUT' : 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(autoReplyData)
      });
      
      if (!response.ok) throw new Error('Failed to save auto reply');
      
      showNotification(`Auto reply ${editingAutoReply ? 'updated' : 'created'} successfully`, 'success');
      setEditingAutoReply(null);
      setShowForm(false);
      fetchAutoReplies();
    } catch (error) {
      showNotification('Failed to save auto reply', 'error');
      console.error('Error saving auto reply:', error);
    }
  };
  
  const handleDeleteAutoReply = async (autoReplyId) => {
    if (!window.confirm('Delete this auto reply rule?')) {
      return;
    }
    
    try {
      const response = await fetch(`/api/auto-replies/${autoReplyId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to delete auto reply');
      
      showNotification('Auto reply deleted successfully', 'success');
      fetchAutoReplies();
    } catch (error) {
      showNotification('Failed to delete auto reply', 'error');
      console.error('Error deleting auto reply:', error);
    }
  };
  
  const handleToggleActive = async (autoReply) => {
    try {
      const response = await fetch(`/api/auto-replies/${autoReply.id}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          is_active: !autoReply.is_active
        })
      });
      
      if (!response.ok) throw new Error('Failed to update auto reply');
      
      showNotification(`Auto reply ${!autoReply.is_active ? 'activated' : 'deactivated'}`, 'success');
      fetchAutoReplies();
    } catch (error) {
      showNotification('Failed to update auto reply', 'error');
      console.error('Error updating auto reply:', error);
    }
  };
  
  if (!isOpen) return null;
  
  return (
    <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg max-w-6xl w-full max-h-[90vh] overflow-y-auto">
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-medium text-gray-900">Auto Reply Rules</h3>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
        
        <div className="p-6">
          {!showForm ? (
            <>
              {/* Auto Replies List */}
              <div className="flex justify-between items-center mb-6">
                <h4 className="text-md font-medium text-gray-900">Auto Reply Rules ({autoReplies.length})</h4>
                <button
                  onClick={() => {
                    setEditingAutoReply(null);
                    setShowForm(true);
                  }}
                  className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 flex items-center gap-2"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                  </svg>
                  Add Auto Reply
                </button>
              </div>
              
              {loading ? (
                <div className="flex justify-center items-center py-8">
                  <svg className="animate-spin h-8 w-8 text-primary-600" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                </div>
              ) : autoReplies.length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                  <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <p className="mt-2">No auto reply rules yet</p>
                  <p className="text-sm">Create your first rule to automatically respond to messages</p>
                </div>
              ) : (
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                  {autoReplies.map(autoReply => (
                    <div key={autoReply.id} className="bg-white border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow">
                      <div className="flex items-start justify-between mb-3">
                        <div className="flex-1">
                          <div className="flex items-center gap-2 mb-1">
                            <h5 className="font-medium text-gray-900">{autoReply.name}</h5>
                            <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                              autoReply.trigger_type === 'keyword' ? 'bg-blue-100 text-blue-800' :
                              autoReply.trigger_type === 'time_based' ? 'bg-green-100 text-green-800' :
                              'bg-purple-100 text-purple-800'
                            }`}>
                              {triggerTypes.find(t => t.value === autoReply.trigger_type)?.label}
                            </span>
                          </div>
                          
                          <div className="text-sm text-gray-600 mb-2">
                            {autoReply.trigger_type === 'keyword' && autoReply.keywords && (
                              <div>
                                <span className="font-medium">Keywords:</span> {autoReply.keywords.join(', ')}
                              </div>
                            )}
                            {autoReply.trigger_type === 'time_based' && (
                              <div>
                                <span className="font-medium">Trigger:</span> After {autoReply.delay_amount} {autoReply.delay_unit}
                              </div>
                            )}
                            {autoReply.start_time && autoReply.end_time && (
                              <div>
                                <span className="font-medium">Active:</span> {autoReply.start_time} - {autoReply.end_time}
                              </div>
                            )}
                          </div>
                          
                          <div className="text-sm text-gray-600 mb-2 line-clamp-2">
                            <span className="font-medium">Message:</span> {autoReply.message_content || 'Template based'}
                          </div>
                          
                          <div className="flex items-center gap-4 text-sm text-gray-500">
                            <span>Used {autoReply.usage_count || 0} times</span>
                            <span>Daily limit: {autoReply.daily_limit || 'Unlimited'}</span>
                            <span className={autoReply.is_active ? 'text-green-600' : 'text-red-600'}>
                              {autoReply.is_active ? 'Active' : 'Inactive'}
                            </span>
                          </div>
                        </div>
                        
                        <div className="flex items-center gap-1 ml-4">
                          <button
                            onClick={() => handleToggleActive(autoReply)}
                            className={`p-1 rounded ${
                              autoReply.is_active 
                                ? 'text-green-600 hover:text-green-800' 
                                : 'text-gray-400 hover:text-green-600'
                            }`}
                            title={autoReply.is_active ? 'Deactivate' : 'Activate'}
                          >
                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                          </button>
                          <button
                            onClick={() => {
                              setEditingAutoReply(autoReply);
                              setShowForm(true);
                            }}
                            className="p-1 text-gray-400 hover:text-gray-600"
                            title="Edit auto reply"
                          >
                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                            </svg>
                          </button>
                          <button
                            onClick={() => handleDeleteAutoReply(autoReply.id)}
                            className="p-1 text-gray-400 hover:text-red-600"
                            title="Delete auto reply"
                          >
                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                            </svg>
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </>
          ) : (
            <>
              {/* Auto Reply Form */}
              <div className="mb-6">
                <div className="flex items-center justify-between mb-4">
                  <h4 className="text-md font-medium text-gray-900">
                    {editingAutoReply ? 'Edit Auto Reply' : 'Create New Auto Reply'}
                  </h4>
                  <button
                    onClick={() => setShowForm(false)}
                    className="text-sm text-gray-500 hover:text-gray-700"
                  >
                    ← Back to Auto Replies
                  </button>
                </div>
                
                <form onSubmit={(e) => {
                  e.preventDefault();
                  const formData = new FormData(e.target);
                  
                  // Parse keywords
                  const keywordsText = formData.get('keywords') || '';
                  const keywords = keywordsText.split(',').map(k => k.trim()).filter(k => k);
                  
                  handleSaveAutoReply({
                    name: formData.get('name'),
                    session_id: formData.get('session_id'),
                    trigger_type: formData.get('trigger_type'),
                    keywords: keywords.length > 0 ? keywords : null,
                    delay_amount: formData.get('delay_amount') ? parseInt(formData.get('delay_amount')) : null,
                    delay_unit: formData.get('delay_unit') || null,
                    template_id: formData.get('template_id') || null,
                    message_content: formData.get('message_content') || null,
                    start_time: formData.get('start_time') || null,
                    end_time: formData.get('end_time') || null,
                    daily_limit: formData.get('daily_limit') ? parseInt(formData.get('daily_limit')) : null,
                    is_active: formData.get('is_active') === 'true'
                  });
                }}>
                  <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                    {/* Left Column */}
                    <div className="space-y-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Rule Name *
                        </label>
                        <input
                          type="text"
                          name="name"
                          defaultValue={editingAutoReply?.name}
                          required
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                          placeholder="e.g., Welcome Message, Support Hours"
                        />
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          WhatsApp Session *
                        </label>
                        <select
                          name="session_id"
                          defaultValue={editingAutoReply?.session_id}
                          required
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
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Trigger Type *
                        </label>
                        <select
                          name="trigger_type"
                          defaultValue={editingAutoReply?.trigger_type || 'keyword'}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                        >
                          {triggerTypes.map(type => (
                            <option key={type.value} value={type.value}>{type.label}</option>
                          ))}
                        </select>
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Keywords (comma separated)
                        </label>
                        <input
                          type="text"
                          name="keywords"
                          defaultValue={editingAutoReply?.keywords?.join(', ')}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                          placeholder="hello, hi, support, help"
                        />
                        <p className="mt-1 text-sm text-gray-500">
                          Messages containing these keywords will trigger the auto reply
                        </p>
                      </div>
                      
                      <div className="grid grid-cols-2 gap-3">
                        <div>
                          <label className="block text-sm font-medium text-gray-700 mb-1">
                            Delay Amount
                          </label>
                          <input
                            type="number"
                            name="delay_amount"
                            min="1"
                            defaultValue={editingAutoReply?.delay_amount}
                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                          />
                        </div>
                        <div>
                          <label className="block text-sm font-medium text-gray-700 mb-1">
                            Delay Unit
                          </label>
                          <select
                            name="delay_unit"
                            defaultValue={editingAutoReply?.delay_unit || 'minutes'}
                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                          >
                            {timeUnits.map(unit => (
                              <option key={unit.value} value={unit.value}>{unit.label}</option>
                            ))}
                          </select>
                        </div>
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Daily Reply Limit per Contact
                        </label>
                        <input
                          type="number"
                          name="daily_limit"
                          min="1"
                          defaultValue={editingAutoReply?.daily_limit}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                          placeholder="Leave empty for unlimited"
                        />
                      </div>
                    </div>
                    
                    {/* Right Column */}
                    <div className="space-y-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Message Template
                        </label>
                        <select
                          name="template_id"
                          defaultValue={editingAutoReply?.template_id || ''}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                        >
                          <option value="">Use custom message below</option>
                          {templates.map(template => (
                            <option key={template.id} value={template.id}>
                              {template.name} ({template.type})
                            </option>
                          ))}
                        </select>
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Custom Message
                        </label>
                        <textarea
                          name="message_content"
                          defaultValue={editingAutoReply?.message_content}
                          rows={6}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                          placeholder="Hello {{name}}, thank you for your message. We'll get back to you soon!"
                        />
                        <p className="mt-1 text-sm text-gray-500">
                          Use template above or enter custom message. Variables like {{name}}, {{phone}} are supported.
                        </p>
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                          Active Hours (Optional)
                        </label>
                        <div className="grid grid-cols-2 gap-3">
                          <div>
                            <label className="block text-xs text-gray-600">Start Time</label>
                            <input
                              type="time"
                              name="start_time"
                              defaultValue={editingAutoReply?.start_time}
                              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                            />
                          </div>
                          <div>
                            <label className="block text-xs text-gray-600">End Time</label>
                            <input
                              type="time"
                              name="end_time"
                              defaultValue={editingAutoReply?.end_time}
                              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                            />
                          </div>
                        </div>
                        <p className="mt-1 text-sm text-gray-500">
                          Leave empty to be active 24/7
                        </p>
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Status
                        </label>
                        <select
                          name="is_active"
                          defaultValue={editingAutoReply?.is_active !== false ? 'true' : 'false'}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                        >
                          <option value="true">Active</option>
                          <option value="false">Inactive</option>
                        </select>
                      </div>
                    </div>
                  </div>
                  
                  <div className="mt-6 flex justify-end gap-3">
                    <button
                      type="button"
                      onClick={() => setShowForm(false)}
                      className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700"
                    >
                      {editingAutoReply ? 'Update' : 'Create'} Auto Reply
                    </button>
                  </div>
                </form>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default AutoReplyModal;