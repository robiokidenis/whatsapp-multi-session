import { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { useNotification } from '../contexts/NotificationContext';
import axios from 'axios';

const Settings = () => {
  const { user } = useAuth();
  const { showNotification } = useNotification();
  const [activeTab, setActiveTab] = useState('profile');
  const [loading, setLoading] = useState(false);
  
  // API Key Management State
  const [apiKeyInfo, setApiKeyInfo] = useState(null);
  const [showApiKey, setShowApiKey] = useState(false);
  const [generatedApiKey, setGeneratedApiKey] = useState('');
  const [loadingApiKey, setLoadingApiKey] = useState(false);
  
  // Password Change State
  const [passwordForm, setPasswordForm] = useState({
    oldPassword: '',
    newPassword: '',
    confirmPassword: ''
  });
  const [loadingPassword, setLoadingPassword] = useState(false);

  // Load API key info on component mount
  useEffect(() => {
    if (activeTab === 'api-keys') {
      loadApiKeyInfo();
    }
  }, [activeTab]);

  const loadApiKeyInfo = async () => {
    try {
      setLoadingApiKey(true);
      const response = await axios.get('/api/auth/api-key');
      if (response.data.success) {
        setApiKeyInfo(response.data.data);
      }
    } catch (error) {
      console.error('Failed to load API key info:', error);
      showNotification('Failed to load API key information', 'error');
    } finally {
      setLoadingApiKey(false);
    }
  };

  const generateApiKey = async () => {
    try {
      setLoadingApiKey(true);
      const response = await axios.post('/api/auth/api-key');
      if (response.data.success) {
        setGeneratedApiKey(response.data.data.api_key);
        setShowApiKey(true);
        await loadApiKeyInfo(); // Refresh the info
        showNotification('API key generated successfully!', 'success');
      }
    } catch (error) {
      console.error('Failed to generate API key:', error);
      showNotification('Failed to generate API key', 'error');
    } finally {
      setLoadingApiKey(false);
    }
  };

  const revokeApiKey = async () => {
    if (!window.confirm('Are you sure you want to revoke your API key? This action cannot be undone.')) {
      return;
    }

    try {
      setLoadingApiKey(true);
      await axios.delete('/api/auth/api-key');
      setApiKeyInfo({ has_key: false, created_at: null, last_used: null });
      setGeneratedApiKey('');
      setShowApiKey(false);
      showNotification('API key revoked successfully', 'success');
    } catch (error) {
      console.error('Failed to revoke API key:', error);
      showNotification('Failed to revoke API key', 'error');
    } finally {
      setLoadingApiKey(false);
    }
  };

  const changePassword = async (e) => {
    e.preventDefault();
    
    if (passwordForm.newPassword !== passwordForm.confirmPassword) {
      showNotification('New passwords do not match', 'error');
      return;
    }

    if (passwordForm.newPassword.length < 6) {
      showNotification('New password must be at least 6 characters long', 'error');
      return;
    }

    try {
      setLoadingPassword(true);
      await axios.post('/api/auth/change-password', {
        old_password: passwordForm.oldPassword,
        new_password: passwordForm.newPassword
      });
      
      setPasswordForm({ oldPassword: '', newPassword: '', confirmPassword: '' });
      showNotification('Password changed successfully!', 'success');
    } catch (error) {
      console.error('Failed to change password:', error);
      showNotification(
        error.response?.data?.error || 'Failed to change password', 
        'error'
      );
    } finally {
      setLoadingPassword(false);
    }
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text).then(() => {
      showNotification('Copied to clipboard!', 'success');
    }).catch(() => {
      showNotification('Failed to copy to clipboard', 'error');
    });
  };

  const tabs = [
    { id: 'profile', name: 'Profile', icon: '👤' },
    { id: 'api-keys', name: 'API Keys', icon: '🔑' },
    { id: 'security', name: 'Security', icon: '🔒' }
  ];

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Settings</h1>
        <p className="text-gray-600 mt-2">Manage your account settings and preferences</p>
      </div>

      {/* Tab Navigation */}
      <div className="border-b border-gray-200 mb-6">
        <nav className="-mb-px flex space-x-8">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`
                py-2 px-1 border-b-2 font-medium text-sm flex items-center space-x-2
                ${activeTab === tab.id
                  ? 'border-primary-500 text-primary-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }
              `}
            >
              <span>{tab.icon}</span>
              <span>{tab.name}</span>
            </button>
          ))}
        </nav>
      </div>

      {/* Profile Tab */}
      {activeTab === 'profile' && (
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Profile Information</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Username</label>
              <input
                type="text"
                value={user?.username || ''}
                disabled
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm bg-gray-50 text-gray-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Role</label>
              <input
                type="text"
                value={user?.role === 'admin' ? 'Administrator' : 'User'}
                disabled
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm bg-gray-50 text-gray-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Session Limit</label>
              <input
                type="text"
                value={user?.session_limit || 'N/A'}
                disabled
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm bg-gray-50 text-gray-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Account Status</label>
              <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                user?.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
              }`}>
                {user?.is_active ? 'Active' : 'Inactive'}
              </span>
            </div>
          </div>
        </div>
      )}

      {/* API Keys Tab */}
      {activeTab === 'api-keys' && (
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex justify-between items-center mb-4">
            <div>
              <h2 className="text-xl font-semibold text-gray-900">API Key Management</h2>
              <p className="text-gray-600 text-sm mt-1">
                API keys provide an alternative authentication method for programmatic access
              </p>
            </div>
          </div>

          {loadingApiKey ? (
            <div className="flex items-center justify-center p-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
            </div>
          ) : (
            <div className="space-y-6">
              {/* API Key Status */}
              <div className="border rounded-lg p-4 bg-gray-50">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-medium text-gray-900">Current API Key</h3>
                    <p className="text-sm text-gray-600">
                      {apiKeyInfo?.has_key ? 'You have an active API key' : 'No API key generated'}
                    </p>
                  </div>
                  <div className="flex space-x-2">
                    {apiKeyInfo?.has_key ? (
                      <button
                        onClick={revokeApiKey}
                        disabled={loadingApiKey}
                        className="px-4 py-2 border border-red-300 text-red-700 rounded-md hover:bg-red-50 focus:outline-none focus:ring-2 focus:ring-red-500 disabled:opacity-50"
                      >
                        Revoke
                      </button>
                    ) : null}
                    <button
                      onClick={generateApiKey}
                      disabled={loadingApiKey}
                      className="px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50"
                    >
                      {apiKeyInfo?.has_key ? 'Regenerate' : 'Generate'} API Key
                    </button>
                  </div>
                </div>

                {apiKeyInfo?.has_key && (
                  <div className="mt-4 grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <span className="font-medium text-gray-700">Created:</span>
                      <span className="ml-2 text-gray-600">
                        {apiKeyInfo.created_at ? new Date(apiKeyInfo.created_at).toLocaleDateString() : 'N/A'}
                      </span>
                    </div>
                    <div>
                      <span className="font-medium text-gray-700">Last Used:</span>
                      <span className="ml-2 text-gray-600">
                        {apiKeyInfo.last_used ? new Date(apiKeyInfo.last_used).toLocaleDateString() : 'Never'}
                      </span>
                    </div>
                  </div>
                )}
              </div>

              {/* Show Generated API Key */}
              {showApiKey && generatedApiKey && (
                <div className="border-2 border-green-200 rounded-lg p-4 bg-green-50">
                  <div className="flex items-center mb-2">
                    <svg className="w-5 h-5 text-green-600 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <h3 className="font-medium text-green-900">API Key Generated Successfully!</h3>
                  </div>
                  <p className="text-sm text-green-700 mb-3">
                    Please copy your API key now. You won't be able to see it again!
                  </p>
                  <div className="flex items-center space-x-2">
                    <input
                      type="text"
                      value={generatedApiKey}
                      readOnly
                      className="flex-1 px-3 py-2 border border-green-300 rounded-md bg-white font-mono text-sm"
                    />
                    <button
                      onClick={() => copyToClipboard(generatedApiKey)}
                      className="px-3 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-green-500"
                    >
                      Copy
                    </button>
                    <button
                      onClick={() => setShowApiKey(false)}
                      className="px-3 py-2 border border-green-300 text-green-700 rounded-md hover:bg-green-100 focus:outline-none focus:ring-2 focus:ring-green-500"
                    >
                      Hide
                    </button>
                  </div>
                </div>
              )}

              {/* API Key Usage Instructions */}
              <div className="border rounded-lg p-4">
                <h3 className="font-medium text-gray-900 mb-2">Usage Instructions</h3>
                <div className="text-sm text-gray-600 space-y-2">
                  <p>Use your API key in the Authorization header:</p>
                  <div className="bg-gray-100 p-3 rounded font-mono text-xs">
                    Authorization: Bearer your_api_key_here
                  </div>
                  <p className="mt-2">Example with curl:</p>
                  <div className="bg-gray-100 p-3 rounded font-mono text-xs">
                    curl -H "Authorization: Bearer your_api_key_here" \\<br />
                    &nbsp;&nbsp;http://localhost:8080/api/sessions
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Security Tab */}
      {activeTab === 'security' && (
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Security Settings</h2>
          
          {/* Change Password Form */}
          <form onSubmit={changePassword} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Current Password</label>
              <input
                type="password"
                value={passwordForm.oldPassword}
                onChange={(e) => setPasswordForm(prev => ({ ...prev, oldPassword: e.target.value }))}
                required
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">New Password</label>
              <input
                type="password"
                value={passwordForm.newPassword}
                onChange={(e) => setPasswordForm(prev => ({ ...prev, newPassword: e.target.value }))}
                required
                minLength={6}
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Confirm New Password</label>
              <input
                type="password"
                value={passwordForm.confirmPassword}
                onChange={(e) => setPasswordForm(prev => ({ ...prev, confirmPassword: e.target.value }))}
                required
                minLength={6}
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500"
              />
            </div>
            <div className="flex justify-end">
              <button
                type="submit"
                disabled={loadingPassword}
                className="px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50 flex items-center"
              >
                {loadingPassword && (
                  <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                )}
                Change Password
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
};

export default Settings;