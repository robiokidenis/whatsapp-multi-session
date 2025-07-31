import React, { useState, useEffect } from 'react';
import { useNotification } from '../contexts/NotificationContext';
import { useAuth } from '../contexts/AuthContext';

const Settings = () => {
  const { token } = useAuth();
  const { showError, showSuccess } = useNotification();
  
  // Settings state
  const [settings, setSettings] = useState({
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone, // Default to user's system timezone
    dateFormat: 'YYYY-MM-DD',
    timeFormat: '24h',
    language: 'en',
    notifications: {
      email: true,
      push: true,
      sms: false
    }
  });
  
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  // Common timezones
  const commonTimezones = [
    'America/New_York',
    'America/Chicago', 
    'America/Denver',
    'America/Los_Angeles',
    'Europe/London',
    'Europe/Paris',
    'Europe/Berlin',
    'Asia/Tokyo',
    'Asia/Shanghai',
    'Asia/Kolkata',
    'Asia/Dubai',
    'Asia/Jakarta',
    'Australia/Sydney',
    'Pacific/Auckland'
  ];

  // Load settings
  const loadSettings = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/user/settings', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (response.ok) {
        const data = await response.json();
        setSettings(prev => ({ ...prev, ...data }));
      }
    } catch (error) {
      console.error('Error loading settings:', error);
      // Use default settings
    } finally {
      setLoading(false);
    }
  };

  // Save settings
  const saveSettings = async () => {
    try {
      setSaving(true);
      const response = await fetch('/api/user/settings', {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(settings)
      });
      
      if (response.ok) {
        showSuccess('Settings saved successfully');
        // Store timezone in localStorage for immediate use
        localStorage.setItem('userTimezone', settings.timezone);
      } else {
        showError('Failed to save settings');
      }
    } catch (error) {
      console.error('Error saving settings:', error);
      showError('Failed to save settings');
    } finally {
      setSaving(false);
    }
  };

  useEffect(() => {
    loadSettings();
  }, []);

  // Get current time in selected timezone
  const getCurrentTimeInTimezone = () => {
    try {
      return new Intl.DateTimeFormat('en-US', {
        timeZone: settings.timezone,
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        timeZoneName: 'short'
      }).format(new Date());
    } catch (error) {
      return 'Invalid timezone';
    }
  };

  if (loading) {
    return (
      <div className="p-6 flex justify-center items-center min-h-[400px]">
        <div className="animate-spin h-8 w-8 border-2 border-primary-600 border-t-transparent rounded-full"></div>
      </div>
    );
  }

  return (
    <div className="p-6 max-w-4xl mx-auto">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900 mb-2">Settings</h1>
        <p className="text-gray-600">Manage your preferences and application settings</p>
      </div>

      <div className="space-y-6">
        {/* Timezone Settings */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4 flex items-center">
            <svg className="w-5 h-5 mr-2 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Date & Time Settings
          </h2>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Timezone *
              </label>
              <select
                value={settings.timezone}
                onChange={(e) => setSettings(prev => ({ ...prev, timezone: e.target.value }))}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors"
              >
                <optgroup label="Common Timezones">
                  {commonTimezones.map(tz => (
                    <option key={tz} value={tz}>
                      {tz.replace('_', ' ')} ({new Intl.DateTimeFormat('en', {
                        timeZone: tz,
                        timeZoneName: 'short'
                      }).formatToParts(new Date()).find(part => part.type === 'timeZoneName')?.value})
                    </option>
                  ))}
                </optgroup>
              </select>
              <p className="mt-2 text-sm text-gray-500">
                Current time: <span className="font-medium">{getCurrentTimeInTimezone()}</span>
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Date Format
              </label>
              <select
                value={settings.dateFormat}
                onChange={(e) => setSettings(prev => ({ ...prev, dateFormat: e.target.value }))}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors"
              >
                <option value="YYYY-MM-DD">YYYY-MM-DD (2024-01-15)</option>
                <option value="DD/MM/YYYY">DD/MM/YYYY (15/01/2024)</option>
                <option value="MM/DD/YYYY">MM/DD/YYYY (01/15/2024)</option>
                <option value="DD-MM-YYYY">DD-MM-YYYY (15-01-2024)</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Time Format
              </label>
              <select
                value={settings.timeFormat}
                onChange={(e) => setSettings(prev => ({ ...prev, timeFormat: e.target.value }))}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors"
              >
                <option value="24h">24 Hour (14:30)</option>
                <option value="12h">12 Hour (2:30 PM)</option>
              </select>
            </div>
          </div>
        </div>

        {/* Template Variables Info */}
        <div className="bg-blue-50 rounded-lg p-6 border border-blue-200">
          <h3 className="text-lg font-medium text-blue-900 mb-3 flex items-center">
            <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Template Variables
          </h3>
          <p className="text-blue-800 mb-3">
            The following variables are available for your message templates based on contact data:
          </p>
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
            {[
              { name: 'name', desc: 'Contact name' },
              { name: 'phone', desc: 'Phone number' },
              { name: 'email', desc: 'Email address' },
              { name: 'company', desc: 'Company name' },
              { name: 'position', desc: 'Job position' },
              { name: 'date', desc: 'Current date (your timezone)' },
              { name: 'time', desc: 'Current time (your timezone)' }
            ].map(variable => (
              <div key={variable.name} className="bg-white rounded-lg p-3 border border-blue-200">
                <code className="text-primary-600 font-medium">[{variable.name}]</code>
                <p className="text-xs text-gray-600 mt-1">{variable.desc}</p>
              </div>
            ))}
          </div>
          <p className="text-blue-700 text-sm mt-3">
            📅 Date and time variables will automatically use your selected timezone above.
          </p>
        </div>

        {/* Notification Settings */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4 flex items-center">
            <svg className="w-5 h-5 mr-2 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-5 5v-5zM3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17M17 13v4a2 2 0 01-2 2H9a2 2 0 01-2-2v-4m8 0V9a2 2 0 00-2-2H9a2 2 0 00-2 2v4.01" />
            </svg>
            Notification Preferences
          </h2>
          
          <div className="space-y-4">
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={settings.notifications.email}
                onChange={(e) => setSettings(prev => ({
                  ...prev,
                  notifications: { ...prev.notifications, email: e.target.checked }
                }))}
                className="w-4 h-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
              />
              <span className="ml-3 text-sm text-gray-700">Email notifications</span>
            </label>
            
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={settings.notifications.push}
                onChange={(e) => setSettings(prev => ({
                  ...prev,
                  notifications: { ...prev.notifications, push: e.target.checked }
                }))}
                className="w-4 h-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
              />
              <span className="ml-3 text-sm text-gray-700">Push notifications</span>
            </label>
            
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={settings.notifications.sms}
                onChange={(e) => setSettings(prev => ({
                  ...prev,
                  notifications: { ...prev.notifications, sms: e.target.checked }
                }))}
                className="w-4 h-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
              />
              <span className="ml-3 text-sm text-gray-700">SMS notifications</span>
            </label>
          </div>
        </div>

        {/* Save Button */}
        <div className="flex justify-end">
          <button
            onClick={saveSettings}
            disabled={saving}
            className="px-6 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
          >
            {saving ? (
              <>
                <div className="animate-spin h-4 w-4 border-2 border-white border-t-transparent rounded-full"></div>
                Saving...
              </>
            ) : (
              <>
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
                Save Settings
              </>
            )}
          </button>
        </div>
      </div>
    </div>
  );
};

export default Settings;