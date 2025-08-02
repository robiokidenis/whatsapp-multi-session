import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';

const EditSessionModal = ({ isOpen, onClose, session, onUpdate }) => {
  const { token } = useAuth();
  const [name, setName] = useState('');
  const [webhookUrl, setWebhookUrl] = useState('');
  const [autoReplyText, setAutoReplyText] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    if (session) {
      setName(session.name || '');
      setWebhookUrl(session.webhook_url || '');
      setAutoReplyText(session.auto_reply_text || '');
      setError('');
      setSuccess('');
    }
  }, [session]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    setSuccess('');

    try {
      console.log('Edit session - Token:', token ? 'Present' : 'Missing');
      
      if (!token) {
        throw new Error('No authentication token found. Please login again.');
      }
      
      const headers = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      };

      // Update session name
      console.log('Updating session name for:', session.id, 'to:', name.trim());
      const nameResponse = await fetch(`/api/sessions/${session.id}/name`, {
        method: 'PUT',
        headers,
        body: JSON.stringify({ name: name.trim() })
      });

      console.log('Name update response status:', nameResponse.status);
      
      if (!nameResponse.ok) {
        const errorText = await nameResponse.text();
        console.error('Name update failed:', errorText);
        throw new Error(`Failed to update session name: ${nameResponse.status} ${errorText}`);
      }

      // Update webhook URL
      console.log('Updating webhook URL for:', session.id, 'to:', webhookUrl.trim());
      const webhookResponse = await fetch(`/api/sessions/${session.id}/webhook`, {
        method: 'PUT',
        headers,
        body: JSON.stringify({ webhook_url: webhookUrl.trim() })
      });

      console.log('Webhook update response status:', webhookResponse.status);
      
      if (!webhookResponse.ok) {
        const errorText = await webhookResponse.text();
        console.error('Webhook update failed:', errorText);
        throw new Error(`Failed to update webhook URL: ${webhookResponse.status} ${errorText}`);
      }

      // Update auto reply text
      console.log('Updating auto reply text for:', session.id, 'to:', autoReplyText.trim());
      const autoReplyResponse = await fetch(`/api/sessions/${session.id}/auto-reply`, {
        method: 'PUT',
        headers,
        body: JSON.stringify({ auto_reply_text: autoReplyText.trim() || null })
      });

      console.log('Auto reply update response status:', autoReplyResponse.status);
      
      if (!autoReplyResponse.ok) {
        const errorText = await autoReplyResponse.text();
        console.error('Auto reply update failed:', errorText);
        throw new Error(`Failed to update auto reply text: ${autoReplyResponse.status} ${errorText}`);
      }

      setSuccess('Session updated successfully!');
      
      // Update the session in parent component
      onUpdate({
        ...session,
        name: name.trim(),
        webhook_url: webhookUrl.trim(),
        auto_reply_text: autoReplyText.trim() || null
      });

      // Close modal after a short delay
      setTimeout(() => {
        onClose();
        setSuccess('');
      }, 1500);

    } catch (err) {
      setError(err.message || 'Failed to update session');
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    if (!loading) {
      onClose();
      setError('');
      setSuccess('');
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b">
          <h3 className="text-lg font-semibold text-gray-900">Edit Session</h3>
          <button
            onClick={handleClose}
            disabled={loading}
            className={`text-gray-400 hover:text-gray-600 text-xl font-bold ${
              loading ? 'cursor-not-allowed opacity-50' : ''
            }`}
          >
            ×
          </button>
        </div>

        {/* Content */}
        <form onSubmit={handleSubmit} className="p-6">
          {/* Session ID Display */}
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Session ID
            </label>
            <div className="p-2 bg-gray-100 border rounded text-sm text-gray-600">
              {session?.id}
            </div>
          </div>

          {/* Name Input */}
          <div className="mb-4">
            <label htmlFor="sessionName" className="block text-sm font-medium text-gray-700 mb-1">
              Session Name
            </label>
            <input
              type="text"
              id="sessionName"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Enter session name (optional)"
              disabled={loading}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-green-500 focus:border-green-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
            />
          </div>

          {/* Webhook URL Input */}
          <div className="mb-4">
            <label htmlFor="webhookUrl" className="block text-sm font-medium text-gray-700 mb-1">
              Webhook URL
            </label>
            <input
              type="url"
              id="webhookUrl"
              value={webhookUrl}
              onChange={(e) => setWebhookUrl(e.target.value)}
              placeholder="https://your-webhook-url.com (optional)"
              disabled={loading}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-green-500 focus:border-green-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
            />
            <p className="text-xs text-gray-500 mt-1">
              Webhook URL will receive incoming WhatsApp messages for this session
            </p>
          </div>

          {/* Auto Reply Text Input */}
          <div className="mb-6">
            <label htmlFor="autoReplyText" className="block text-sm font-medium text-gray-700 mb-1">
              Auto Reply Message
            </label>
            <textarea
              id="autoReplyText"
              value={autoReplyText}
              onChange={(e) => setAutoReplyText(e.target.value)}
              placeholder="Enter auto reply message (optional)"
              rows={3}
              disabled={loading}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-green-500 focus:border-green-500 disabled:bg-gray-100 disabled:cursor-not-allowed resize-none"
            />
            <p className="text-xs text-gray-500 mt-1">
              Automatically reply to incoming messages with this text
            </p>
          </div>

          {/* Error Message */}
          {error && (
            <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
              <i className="fas fa-exclamation-triangle mr-2"></i>
              {error}
            </div>
          )}

          {/* Success Message */}
          {success && (
            <div className="mb-4 p-3 bg-green-100 border border-green-400 text-green-700 rounded">
              <i className="fas fa-check-circle mr-2"></i>
              {success}
            </div>
          )}

          {/* Actions */}
          <div className="flex gap-3">
            <button
              type="button"
              onClick={handleClose}
              disabled={loading}
              className={`flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50 transition duration-200 ${
                loading ? 'cursor-not-allowed opacity-50' : ''
              }`}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className={`flex-1 px-4 py-2 bg-green-600 text-white rounded-md transition duration-200 ${
                loading
                  ? 'cursor-not-allowed opacity-50'
                  : 'hover:bg-green-700'
              }`}
            >
              {loading ? (
                <>
                  <i className="fas fa-spinner fa-spin mr-2"></i>
                  Updating...
                </>
              ) : (
                <>
                  <i className="fas fa-save mr-2"></i>
                  Update
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default EditSessionModal;