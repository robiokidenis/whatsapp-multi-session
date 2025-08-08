import { useState } from 'react';
import axios from 'axios';

const CreateSessionModal = ({ isOpen, onClose, onSuccess }) => {
  const [formData, setFormData] = useState({
    phone: '',
    name: '',
    autoReplyText: '',
    proxyEnabled: false,
    proxyType: 'http',
    proxyHost: '',
    proxyPort: '',
    proxyUsername: '',
    proxyPassword: ''
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [testingProxy, setTestingProxy] = useState(false);
  const [proxyTestResult, setProxyTestResult] = useState(null);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      // Build proxy config object
      const proxyConfig = formData.proxyEnabled ? {
        enabled: true,
        type: formData.proxyType,
        host: formData.proxyHost,
        port: parseInt(formData.proxyPort) || 0,
        username: formData.proxyUsername,
        password: formData.proxyPassword
      } : null;

      const response = await axios.post('/api/sessions', {
        phone: formData.phone || '',
        name: formData.name,
        auto_reply_text: formData.autoReplyText || null,
        proxy_config: proxyConfig,
      });

      if (response.data.success) {
        setFormData({ 
          phone: '', 
          name: '', 
          autoReplyText: '',
          proxyEnabled: false,
          proxyType: 'http',
          proxyHost: '',
          proxyPort: '',
          proxyUsername: '',
          proxyPassword: ''
        });
        onSuccess('Session created successfully!');
        onClose();
      } else {
        setError(response.data.error || 'Failed to create session');
      }
    } catch (error) {
      setError('Error creating session: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target;
    setFormData({
      ...formData,
      [name]: type === 'checkbox' ? checked : value,
    });
    
    // Clear proxy test result when proxy settings change
    if (name.startsWith('proxy')) {
      setProxyTestResult(null);
    }
  };

  const testProxyConnection = async () => {
    if (!formData.proxyEnabled || !formData.proxyHost || !formData.proxyPort) {
      setProxyTestResult({ success: false, message: 'Please fill in proxy host and port' });
      return;
    }

    setTestingProxy(true);
    setProxyTestResult(null);

    try {
      const proxyConfig = {
        enabled: true,
        type: formData.proxyType,
        host: formData.proxyHost,
        port: parseInt(formData.proxyPort) || 0,
        username: formData.proxyUsername,
        password: formData.proxyPassword
      };

      const response = await axios.post('/api/proxy/test', {
        proxy_config: proxyConfig
      });

      // Ensure response data is valid
      if (response.data && typeof response.data === 'object') {
        setProxyTestResult(response.data);
      } else {
        setProxyTestResult({
          success: false,
          message: 'Invalid response format from server'
        });
      }
    } catch (error) {
      setProxyTestResult({
        success: false,
        message: error.response?.data?.message || error.message || 'Proxy test failed'
      });
    } finally {
      setTestingProxy(false);
    }
  };

  const handleClose = () => {
    setFormData({ 
      phone: '', 
      name: '', 
      autoReplyText: '',
      proxyEnabled: false,
      proxyType: 'http',
      proxyHost: '',
      proxyPort: '',
      proxyUsername: '',
      proxyPassword: ''
    });
    setError('');
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full max-h-[90vh] overflow-y-auto">
        {/* Modal Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200">
          <div className="flex items-center">
            <div className="flex items-center justify-center w-10 h-10 bg-primary-500 rounded-lg mr-3 shadow-sm">
              <i className="fas fa-plus text-white"></i>
            </div>
            <div>
              <h3 className="text-lg font-semibold text-gray-900">Create New Session</h3>
              <p className="text-sm text-gray-600">Add a new WhatsApp session</p>
            </div>
          </div>
          <button
            onClick={handleClose}
            className="text-gray-400 hover:text-gray-600 transition-colors"
          >
            <i className="fas fa-times text-lg"></i>
          </button>
        </div>

        {/* Modal Body */}
        <form onSubmit={handleSubmit} className="p-6">
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-900 mb-2">
                <i className="fas fa-mobile-alt mr-2 text-primary-500"></i>
                Phone Number
              </label>
              <input
                type="text"
                name="phone"
                value={formData.phone}
                onChange={handleChange}
                placeholder="Enter phone number (optional)"
                className="w-full px-3 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors"
              />
              <p className="text-xs text-gray-500 mt-1">
                Auto-generated if left empty
              </p>
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-900 mb-2">
                <i className="fas fa-tag mr-2 text-primary-500"></i>
                Session Name
              </label>
              <input
                type="text"
                name="name"
                value={formData.name}
                onChange={handleChange}
                placeholder="Enter session name (optional)"
                className="w-full px-3 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors"
              />
              <p className="text-xs text-gray-500 mt-1">
                Helps identify this session
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-900 mb-2">
                <i className="fas fa-reply mr-2 text-primary-500"></i>
                Auto Reply Message
              </label>
              <textarea
                name="autoReplyText"
                value={formData.autoReplyText}
                onChange={handleChange}
                placeholder="Enter auto reply message (optional)"
                rows={3}
                className="w-full px-3 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors resize-none"
              />
              <p className="text-xs text-gray-500 mt-1">
                Automatically reply to incoming messages with this text
              </p>
            </div>

            {/* Proxy Settings */}
            <div className="border-t border-gray-200 pt-4">
              <div className="flex items-center justify-between mb-4">
                <label className="flex items-center text-sm font-medium text-gray-900">
                  <i className="fas fa-shield-alt mr-2 text-primary-500"></i>
                  Proxy Settings
                </label>
                <div className="flex items-center">
                  <input
                    type="checkbox"
                    name="proxyEnabled"
                    checked={formData.proxyEnabled}
                    onChange={handleChange}
                    className="w-4 h-4 text-primary-600 bg-gray-100 border-gray-300 rounded focus:ring-primary-500 focus:ring-2"
                  />
                  <label className="ml-2 text-sm font-medium text-gray-700">
                    Enable Proxy
                  </label>
                </div>
              </div>

              {formData.proxyEnabled && (
                <div className="space-y-4 bg-gray-50 p-4 rounded-lg">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Proxy Type
                      </label>
                      <select
                        name="proxyType"
                        value={formData.proxyType}
                        onChange={handleChange}
                        className="w-full px-3 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors"
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
                        name="proxyPort"
                        value={formData.proxyPort}
                        onChange={handleChange}
                        placeholder="e.g., 8080"
                        className="w-full px-3 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors"
                      />
                    </div>
                  </div>
                  
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Proxy Host
                    </label>
                    <input
                      type="text"
                      name="proxyHost"
                      value={formData.proxyHost}
                      onChange={handleChange}
                      placeholder="e.g., proxy.example.com or 127.0.0.1"
                      className="w-full px-3 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors"
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Username (Optional)
                      </label>
                      <input
                        type="text"
                        name="proxyUsername"
                        value={formData.proxyUsername}
                        onChange={handleChange}
                        placeholder="Proxy username"
                        className="w-full px-3 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Password (Optional)
                      </label>
                      <input
                        type="password"
                        name="proxyPassword"
                        value={formData.proxyPassword}
                        onChange={handleChange}
                        placeholder="Proxy password"
                        className="w-full px-3 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors"
                      />
                    </div>
                  </div>

                  {/* Proxy Test Button */}
                  <div className="flex justify-between items-center mb-4">
                    <button
                      type="button"
                      onClick={testProxyConnection}
                      disabled={testingProxy || !formData.proxyHost || !formData.proxyPort}
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
                        <p>Configure proxy settings for this session to route WhatsApp traffic through your proxy server. This helps with IP masking and geographic distribution.</p>
                      </div>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>

          {error && (
            <div className="mt-4 bg-red-50 border border-red-200 text-red-800 px-3 py-2 rounded-lg">
              <div className="flex items-center text-sm">
                <i className="fas fa-exclamation-circle mr-2"></i>
                <span>{error}</span>
              </div>
            </div>
          )}

          {/* Modal Footer */}
          <div className="flex justify-end gap-3 mt-6 pt-4 border-t border-gray-200">
            <button
              type="button"
              onClick={handleClose}
              className="px-4 py-2 text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg font-medium transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="bg-primary-500 hover:bg-primary-600 text-white px-4 py-2 rounded-lg font-medium transition-colors flex items-center disabled:opacity-50"
            >
              {loading ? (
                <>
                  <i className="fas fa-spinner fa-spin mr-2"></i>
                  Creating...
                </>
              ) : (
                <>
                  <i className="fas fa-plus mr-2"></i>
                  Create Session
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default CreateSessionModal;