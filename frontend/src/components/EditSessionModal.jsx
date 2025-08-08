import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';

const EditSessionModal = ({ isOpen, onClose, session, onUpdate }) => {
  const { token } = useAuth();
  const [name, setName] = useState('');
  const [webhookUrl, setWebhookUrl] = useState('');
  const [autoReplyText, setAutoReplyText] = useState('');
  const [proxyEnabled, setProxyEnabled] = useState(false);
  const [proxyType, setProxyType] = useState('http');
  const [proxyHost, setProxyHost] = useState('');
  const [proxyPort, setProxyPort] = useState('');
  const [proxyUsername, setProxyUsername] = useState('');
  const [proxyPassword, setProxyPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [testingProxy, setTestingProxy] = useState(false);
  const [proxyTestResult, setProxyTestResult] = useState(null);

  useEffect(() => {
    if (session) {
      setName(session.name || '');
      setWebhookUrl(session.webhook_url || '');
      setAutoReplyText(session.auto_reply_text || '');
      
      // Set proxy configuration
      if (session.proxy_config) {
        setProxyEnabled(session.proxy_config.enabled || false);
        setProxyType(session.proxy_config.type || 'http');
        setProxyHost(session.proxy_config.host || '');
        setProxyPort(session.proxy_config.port ? session.proxy_config.port.toString() : '');
        setProxyUsername(session.proxy_config.username || '');
        setProxyPassword(session.proxy_config.password || '');
      } else {
        setProxyEnabled(false);
        setProxyType('http');
        setProxyHost('');
        setProxyPort('');
        setProxyUsername('');
        setProxyPassword('');
      }
      
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

      // Update proxy configuration
      console.log('Updating proxy config for:', session.id);
      const proxyConfig = proxyEnabled ? {
        enabled: true,
        type: proxyType,
        host: proxyHost.trim(),
        port: parseInt(proxyPort) || 0,
        username: proxyUsername.trim(),
        password: proxyPassword.trim()
      } : { enabled: false };
      
      const proxyResponse = await fetch(`/api/sessions/${session.id}/proxy`, {
        method: 'PUT',
        headers,
        body: JSON.stringify({ proxy_config: proxyConfig })
      });

      console.log('Proxy update response status:', proxyResponse.status);
      
      if (!proxyResponse.ok) {
        const errorText = await proxyResponse.text();
        console.error('Proxy update failed:', errorText);
        throw new Error(`Failed to update proxy configuration: ${proxyResponse.status} ${errorText}`);
      }

      setSuccess('Session updated successfully!');
      
      // Update the session in parent component
      onUpdate({
        ...session,
        name: name.trim(),
        webhook_url: webhookUrl.trim(),
        auto_reply_text: autoReplyText.trim() || null,
        proxy_config: proxyConfig
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

  const testProxyConnection = async () => {
    if (!proxyEnabled || !proxyHost || !proxyPort) {
      setProxyTestResult({ success: false, message: 'Please fill in proxy host and port' });
      return;
    }

    setTestingProxy(true);
    setProxyTestResult(null);

    try {
      const proxyConfig = {
        enabled: true,
        type: proxyType,
        host: proxyHost,
        port: parseInt(proxyPort) || 0,
        username: proxyUsername,
        password: proxyPassword
      };

      const response = await fetch('/api/proxy/test', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          proxy_config: proxyConfig
        })
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const contentType = response.headers.get('content-type');
      if (!contentType || !contentType.includes('application/json')) {
        const text = await response.text();
        throw new Error(`Invalid response format. Expected JSON, got: ${text.substring(0, 100)}...`);
      }

      const data = await response.json();
      
      // Ensure response data is valid
      if (data && typeof data === 'object') {
        setProxyTestResult(data);
      } else {
        setProxyTestResult({
          success: false,
          message: 'Invalid response format from server'
        });
      }
    } catch (error) {
      setProxyTestResult({
        success: false,
        message: error.message || 'Proxy test failed'
      });
    } finally {
      setTestingProxy(false);
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
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] overflow-y-auto">
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
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
            {/* Left Column - Basic Settings */}
            <div className="space-y-6">
              <h4 className="text-lg font-semibold text-gray-900 flex items-center">
                <i className="fas fa-cog mr-2 text-green-500"></i>
                Basic Settings
              </h4>
              
              {/* Session ID Display */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Session ID
                </label>
                <div className="p-3 bg-gray-100 border rounded-lg text-sm text-gray-600 font-mono">
                  {session?.id}
                </div>
              </div>

              {/* Name Input */}
              <div>
                <label htmlFor="sessionName" className="block text-sm font-medium text-gray-700 mb-2">
                  <i className="fas fa-tag mr-2 text-green-500"></i>
                  Session Name
                </label>
                <input
                  type="text"
                  id="sessionName"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Enter session name (optional)"
                  disabled={loading}
                  className="w-full px-3 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 disabled:bg-gray-100 disabled:cursor-not-allowed transition-colors"
                />
              </div>

              {/* Webhook URL Input */}
              <div>
                <label htmlFor="webhookUrl" className="block text-sm font-medium text-gray-700 mb-2">
                  <i className="fas fa-link mr-2 text-green-500"></i>
                  Webhook URL
                </label>
                <input
                  type="url"
                  id="webhookUrl"
                  value={webhookUrl}
                  onChange={(e) => setWebhookUrl(e.target.value)}
                  placeholder="https://your-webhook-url.com (optional)"
                  disabled={loading}
                  className="w-full px-3 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 disabled:bg-gray-100 disabled:cursor-not-allowed transition-colors"
                />
                <p className="text-xs text-gray-500 mt-2">
                  Webhook URL will receive incoming WhatsApp messages for this session
                </p>
              </div>

              {/* Auto Reply Text Input */}
              <div>
                <label htmlFor="autoReplyText" className="block text-sm font-medium text-gray-700 mb-2">
                  <i className="fas fa-reply mr-2 text-green-500"></i>
                  Auto Reply Message
                </label>
                <textarea
                  id="autoReplyText"
                  value={autoReplyText}
                  onChange={(e) => setAutoReplyText(e.target.value)}
                  placeholder="Enter auto reply message (optional)"
                  rows={4}
                  disabled={loading}
                  className="w-full px-3 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 disabled:bg-gray-100 disabled:cursor-not-allowed resize-none transition-colors"
                />
                <p className="text-xs text-gray-500 mt-2">
                  Automatically reply to incoming messages with this text
                </p>
              </div>
            </div>

            {/* Right Column - Proxy Settings */}
            <div className="space-y-6">
              <div className="flex items-center justify-between">
                <h4 className="text-lg font-semibold text-gray-700 flex items-center">
                  <i className="fas fa-shield-alt mr-2 text-green-500"></i>
                  Proxy Settings
                </h4>
                <div className="flex items-center">
                  <input
                    type="checkbox"
                    checked={proxyEnabled}
                    onChange={(e) => setProxyEnabled(e.target.checked)}
                    disabled={loading}
                    className="w-4 h-4 text-green-600 bg-gray-100 border-gray-300 rounded focus:ring-green-500 focus:ring-2 disabled:cursor-not-allowed"
                  />
                  <label className="ml-2 text-sm font-medium text-gray-700">
                    Enable Proxy
                  </label>
                </div>
              </div>

              {proxyEnabled ? (
                <div className="space-y-4 bg-gray-50 p-4 rounded-lg border border-gray-200">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Proxy Type
                      </label>
                      <select
                        value={proxyType}
                        onChange={(e) => setProxyType(e.target.value)}
                        disabled={loading}
                        className="w-full px-3 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 disabled:bg-gray-100 disabled:cursor-not-allowed transition-colors"
                      >
                        <option value="http">HTTP</option>
                        <option value="https">HTTPS</option>
                        <option value="socks5">SOCKS5</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Port
                      </label>
                      <input
                        type="number"
                        value={proxyPort}
                        onChange={(e) => setProxyPort(e.target.value)}
                        placeholder="e.g., 8080"
                        disabled={loading}
                        className="w-full px-3 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 disabled:bg-gray-100 disabled:cursor-not-allowed transition-colors"
                      />
                    </div>
                  </div>
                  
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Proxy Host
                    </label>
                    <input
                      type="text"
                      value={proxyHost}
                      onChange={(e) => setProxyHost(e.target.value)}
                      placeholder="e.g., proxy.example.com or 127.0.0.1"
                      disabled={loading}
                      className="w-full px-3 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 disabled:bg-gray-100 disabled:cursor-not-allowed transition-colors"
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Username (Optional)
                      </label>
                      <input
                        type="text"
                        value={proxyUsername}
                        onChange={(e) => setProxyUsername(e.target.value)}
                        placeholder="Proxy username"
                        disabled={loading}
                        className="w-full px-3 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 disabled:bg-gray-100 disabled:cursor-not-allowed transition-colors"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Password (Optional)
                      </label>
                      <input
                        type="password"
                        value={proxyPassword}
                        onChange={(e) => setProxyPassword(e.target.value)}
                        placeholder="Proxy password"
                        disabled={loading}
                        className="w-full px-3 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 disabled:bg-gray-100 disabled:cursor-not-allowed transition-colors"
                      />
                    </div>
                  </div>

                  {/* Proxy Test Button */}
                  <div className="flex justify-between items-center mb-4">
                    <button
                      type="button"
                      onClick={testProxyConnection}
                      disabled={testingProxy || !proxyHost || !proxyPort}
                      className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white rounded-lg font-medium transition-colors flex items-center"
                    >
                      {testingProxy ? (
                        <>
                          <i className="fas fa-spinner fa-spin mr-2"></i>
                          Testing...
                        </>
                      ) : (
                        <>
                          <i className="fas fa-network-wired mr-2"></i>
                          Test Connection
                        </>
                      )}
                    </button>
                  </div>

                  {/* Proxy Test Result */}
                  {proxyTestResult && (
                    <div className={`mb-4 p-3 rounded-lg border ${
                      proxyTestResult.success 
                        ? 'bg-green-50 border-green-200 text-green-800'
                        : 'bg-red-50 border-red-200 text-red-800'
                    }`}>
                      <div className="flex items-start">
                        <i className={`fas ${proxyTestResult.success ? 'fa-check-circle' : 'fa-exclamation-triangle'} mt-0.5 mr-2`}></i>
                        <div className="text-sm">
                          <p className="font-medium mb-1">
                            {proxyTestResult.success ? 'Connection Successful' : 'Connection Failed'}
                          </p>
                          <p>{proxyTestResult.message}</p>
                          {proxyTestResult.proxy_info && (
                            <p className="mt-1 text-xs opacity-75">
                              {proxyTestResult.proxy_info.type?.toUpperCase()} proxy at {proxyTestResult.proxy_info.host}:{proxyTestResult.proxy_info.port}
                            </p>
                          )}
                        </div>
                      </div>
                    </div>
                  )}

                  <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
                    <div className="flex items-start">
                      <i className="fas fa-info-circle text-blue-500 mt-0.5 mr-2"></i>
                      <div className="text-sm text-blue-800">
                        <p className="font-medium mb-1">Proxy Configuration</p>
                        <p>Update proxy settings for this session to route WhatsApp traffic through your proxy server. Changes take effect after session restart.</p>
                      </div>
                    </div>
                  </div>
                </div>
              ) : (
                <div className="text-center py-12 text-gray-500">
                  <i className="fas fa-shield-alt text-4xl mb-4"></i>
                  <p className="text-lg font-medium mb-2">Proxy Disabled</p>
                  <p className="text-sm">Enable proxy to configure connection settings for this session</p>
                </div>
              )}
            </div>
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
                  Updating Session & Proxy...
                </>
              ) : (
                <>
                  <i className="fas fa-save mr-2"></i>
                  Update Session & Proxy
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