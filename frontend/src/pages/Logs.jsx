import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';

const LogsPage = () => {
  const { token } = useAuth();
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(false);
  const [filters, setFilters] = useState({
    level: '',
    component: '',
    session_id: '',
    start_time: '',
    end_time: '',
    page: 1,
    page_size: 50
  });
  const [pagination, setPagination] = useState({
    total: 0,
    totalPages: 0,
    page: 1,
    pageSize: 50
  });
  const [levels, setLevels] = useState([]);
  const [components, setComponents] = useState([]);
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [refreshInterval, setRefreshInterval] = useState(null);
  const [loggingStatus, setLoggingStatus] = useState(null);
  const [statusError, setStatusError] = useState(null);

  // Fetch logging status and conditionally fetch other data
  useEffect(() => {
    fetchLoggingStatus();
  }, []);

  // Fetch logs data only if database logging is enabled
  useEffect(() => {
    if (loggingStatus?.database_logging_enabled) {
      fetchLogLevels();
      fetchComponents();
    }
  }, [loggingStatus]);

  // Fetch logs when filters change (only if database logging is enabled)
  useEffect(() => {
    if (loggingStatus?.database_logging_enabled) {
      fetchLogs();
    }
  }, [filters, loggingStatus]);

  // Auto-refresh functionality (only if database logging is enabled)
  useEffect(() => {
    if (autoRefresh && filters.page === 1 && loggingStatus?.database_logging_enabled) {
      const interval = setInterval(() => {
        fetchLogs();
      }, 5000); // Refresh every 5 seconds
      setRefreshInterval(interval);
      return () => clearInterval(interval);
    } else if (refreshInterval) {
      clearInterval(refreshInterval);
      setRefreshInterval(null);
    }
  }, [autoRefresh, filters.page, loggingStatus]);

  // Cleanup interval on unmount
  useEffect(() => {
    return () => {
      if (refreshInterval) {
        clearInterval(refreshInterval);
      }
    };
  }, []);

  const fetchLoggingStatus = async () => {
    try {
      const response = await fetch('/api/admin/logs/status', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      if (response.ok) {
        const data = await response.json();
        setLoggingStatus(data);
        setStatusError(null);
      } else {
        setStatusError('Failed to fetch logging status');
      }
    } catch (error) {
      console.error('Failed to fetch logging status:', error);
      setStatusError('Failed to connect to server');
    }
  };

  const fetchLogLevels = async () => {
    try {
      const response = await fetch('/api/admin/logs/levels', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      if (response.ok) {
        const data = await response.json();
        setLevels(data.levels || []);
      }
    } catch (error) {
      console.error('Failed to fetch log levels:', error);
    }
  };

  const fetchComponents = async () => {
    try {
      const response = await fetch('/api/admin/logs/components', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      if (response.ok) {
        const data = await response.json();
        setComponents(data.components || []);
      }
    } catch (error) {
      console.error('Failed to fetch log components:', error);
    }
  };

  const fetchLogs = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      Object.entries(filters).forEach(([key, value]) => {
        if (value) params.append(key, value);
      });

      const response = await fetch(`/api/admin/logs?${params}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });

      if (response.ok) {
        const data = await response.json();
        setLogs(data.logs || []);
        setPagination({
          total: data.total,
          totalPages: data.total_pages,
          page: data.page,
          pageSize: data.page_size
        });
      }
    } catch (error) {
      console.error('Failed to fetch logs:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleFilterChange = (key, value) => {
    setFilters(prev => ({
      ...prev,
      [key]: value,
      page: 1 // Reset to first page when filters change
    }));
  };

  const handlePageChange = (newPage) => {
    setFilters(prev => ({
      ...prev,
      page: newPage
    }));
  };

  const clearFilters = () => {
    setFilters({
      level: '',
      component: '',
      session_id: '',
      start_time: '',
      end_time: '',
      page: 1,
      page_size: 50
    });
  };

  const formatTimestamp = (timestamp) => {
    return new Date(timestamp * 1000).toLocaleString();
  };

  const getLevelColor = (level) => {
    switch (level?.toLowerCase()) {
      case 'debug': return 'text-gray-600';
      case 'info': return 'text-blue-600';
      case 'warn': return 'text-yellow-600';
      case 'error': return 'text-red-600';
      default: return 'text-gray-800';
    }
  };

  const getLevelBadge = (level) => {
    const colors = {
      debug: 'bg-gray-100 text-gray-800',
      info: 'bg-blue-100 text-blue-800',
      warn: 'bg-yellow-100 text-yellow-800',
      error: 'bg-red-100 text-red-800'
    };
    
    return colors[level?.toLowerCase()] || 'bg-gray-100 text-gray-800';
  };

  const deleteOldLogs = async (days) => {
    if (!confirm(`Are you sure you want to delete logs older than ${days} days?`)) {
      return;
    }

    try {
      const response = await fetch(`/api/admin/logs/cleanup/${days}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });

      if (response.ok) {
        const data = await response.json();
        alert(`✅ Successfully deleted ${data.deleted_count} old log entries`);
        fetchLogs(); // Refresh logs
      } else {
        const errorText = await response.text();
        console.error('Delete failed:', response.status, errorText);
        alert(`❌ Failed to delete logs: ${response.status} - ${errorText}`);
      }
    } catch (error) {
      console.error('Failed to delete old logs:', error);
      alert(`❌ Error deleting logs: ${error.message}`);
    }
  };

  const clearAllLogs = async () => {
    if (!confirm('⚠️ Are you sure you want to DELETE ALL LOGS? This action cannot be undone!')) {
      return;
    }

    // Double confirmation for such a destructive action
    if (!confirm('🚨 FINAL WARNING: This will permanently delete ALL log entries. Type "DELETE ALL" in the next prompt to confirm.')) {
      return;
    }

    const userInput = prompt('Type "DELETE ALL" to confirm (case sensitive):');
    if (userInput !== 'DELETE ALL') {
      alert('❌ Action cancelled - confirmation text did not match');
      return;
    }

    try {
      const response = await fetch('/api/admin/logs/clear', {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });

      if (response.ok) {
        const data = await response.json();
        alert(`✅ Successfully cleared all logs (${data.deleted_count} entries deleted)`);
        fetchLogs(); // Refresh logs
      } else {
        const errorText = await response.text();
        console.error('Clear failed:', response.status, errorText);
        alert(`❌ Failed to clear logs: ${response.status} - ${errorText}`);
      }
    } catch (error) {
      console.error('Failed to clear all logs:', error);
      alert(`❌ Error clearing logs: ${error.message}`);
    }
  };

  // Show loading state while fetching status
  if (loggingStatus === null && !statusError) {
    return (
      <div className="space-y-6">
        <div className="bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <div className="text-center py-8">
              <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
              <p className="mt-2 text-sm text-gray-500">Loading logging status...</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Show error state
  if (statusError) {
    return (
      <div className="space-y-6">
        <div className="bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <div className="text-center py-8">
              <div className="text-red-500 text-lg">❌</div>
              <h3 className="text-lg font-medium text-gray-900 mt-2">Error Loading Logs</h3>
              <p className="text-sm text-gray-500 mt-1">{statusError}</p>
              <button
                onClick={fetchLoggingStatus}
                className="mt-4 bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded text-sm"
              >
                Retry
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Show message when database logging is disabled
  if (loggingStatus && !loggingStatus.database_logging_enabled) {
    return (
      <div className="space-y-6">
        <div className="bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <div className="text-center py-8">
              <div className="text-yellow-500 text-6xl mb-4">📝</div>
              <h3 className="text-xl font-medium text-gray-900 mb-2">Database Logging Disabled</h3>
              <p className="text-gray-600 mb-4 max-w-md mx-auto">
                Database logging is currently disabled in the application configuration. 
                Logs are only being displayed in the console output.
              </p>
              <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6 max-w-lg mx-auto">
                <div className="flex items-start">
                  <div className="text-blue-500 mr-3 mt-1">ℹ️</div>
                  <div className="text-left">
                    <h4 className="text-sm font-medium text-blue-900 mb-1">Current Configuration:</h4>
                    <ul className="text-sm text-blue-800 space-y-1">
                      <li>• Console Logging: <strong>{loggingStatus.console_logging_enabled ? 'Enabled' : 'Disabled'}</strong></li>
                      <li>• Database Logging: <strong>Disabled</strong></li>
                      <li>• Log Level: <strong>{loggingStatus.log_level?.toUpperCase()}</strong></li>
                    </ul>
                  </div>
                </div>
              </div>
              <div className="bg-gray-50 border border-gray-200 rounded-lg p-4 max-w-lg mx-auto">
                <h4 className="text-sm font-medium text-gray-900 mb-2">To Enable Database Logging:</h4>
                <p className="text-sm text-gray-600 text-left">
                  Set the environment variable <code className="bg-gray-200 px-2 py-1 rounded">ENABLE_DATABASE_LOG=true</code> 
                  and restart the application to store logs in the database and view them here.
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <div className="flex justify-between items-center mb-4">
            <div>
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                System Logs
              </h3>
              <p className="text-sm text-gray-500">Latest logs displayed first • Database logging enabled</p>
            </div>
            <div className="flex space-x-2">
              <button
                onClick={() => fetchLogs()}
                className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded text-sm"
              >
                Refresh
              </button>
              <label className="flex items-center space-x-2 text-sm">
                <input
                  type="checkbox"
                  checked={autoRefresh}
                  onChange={(e) => setAutoRefresh(e.target.checked)}
                  className="rounded border-gray-300 text-blue-600 focus:border-blue-500 focus:ring-blue-500"
                  disabled={filters.page > 1 || !loggingStatus?.database_logging_enabled}
                />
                <span className={(filters.page > 1 || !loggingStatus?.database_logging_enabled) ? 'text-gray-400' : 'text-gray-700'}>
                  Auto-refresh {autoRefresh && filters.page === 1 && loggingStatus?.database_logging_enabled ? '(5s)' : ''}
                </span>
              </label>
            </div>
          </div>

          {/* Filters */}
          <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-6 gap-4 mb-4">
            <select
              value={filters.level}
              onChange={(e) => handleFilterChange('level', e.target.value)}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
            >
              <option value="">All Levels</option>
              {levels.map(level => (
                <option key={level} value={level}>{level.toUpperCase()}</option>
              ))}
            </select>

            <select
              value={filters.component}
              onChange={(e) => handleFilterChange('component', e.target.value)}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
            >
              <option value="">All Components</option>
              {components.map(component => (
                <option key={component} value={component}>{component}</option>
              ))}
            </select>

            <input
              type="text"
              placeholder="Session ID"
              value={filters.session_id}
              onChange={(e) => handleFilterChange('session_id', e.target.value)}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
            />

            <input
              type="datetime-local"
              value={filters.start_time ? new Date(filters.start_time * 1000).toISOString().slice(0, 16) : ''}
              onChange={(e) => handleFilterChange('start_time', e.target.value ? Math.floor(new Date(e.target.value).getTime() / 1000) : '')}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
            />

            <input
              type="datetime-local"
              value={filters.end_time ? new Date(filters.end_time * 1000).toISOString().slice(0, 16) : ''}
              onChange={(e) => handleFilterChange('end_time', e.target.value ? Math.floor(new Date(e.target.value).getTime() / 1000) : '')}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
            />

            <button
              onClick={clearFilters}
              className="bg-gray-500 hover:bg-gray-700 text-white font-bold py-2 px-4 rounded"
            >
              Clear
            </button>
          </div>

          {/* Actions */}
          <div className="flex justify-between items-center mb-4">
            <div className="flex flex-wrap gap-2">
              <button
                onClick={() => deleteOldLogs(7)}
                className="bg-orange-500 hover:bg-orange-600 text-white font-bold py-2 px-4 rounded text-sm transition-colors duration-200 shadow-sm"
                title="Delete logs older than 7 days"
              >
                🗑️ Delete 7+ days
              </button>
              <button
                onClick={() => deleteOldLogs(30)}
                className="bg-red-500 hover:bg-red-600 text-white font-bold py-2 px-4 rounded text-sm transition-colors duration-200 shadow-sm"
                title="Delete logs older than 30 days"
              >
                🗑️ Delete 30+ days
              </button>
              <button
                onClick={clearAllLogs}
                className="bg-red-700 hover:bg-red-800 text-white font-bold py-2 px-4 rounded text-sm transition-colors duration-200 shadow-sm border-2 border-red-600"
                title="⚠️ Delete ALL logs - Use with extreme caution!"
              >
                🚨 Clear All Logs
              </button>
            </div>
            {filters.page > 1 && (
              <button
                onClick={() => handlePageChange(1)}
                className="bg-green-500 hover:bg-green-600 text-white font-bold py-2 px-4 rounded text-sm transition-colors duration-200 shadow-sm"
                title="Go to the latest logs (page 1)"
              >
                ↑ Latest Logs
              </button>
            )}
          </div>

          {/* Results count and status */}
          <div className="mb-4 flex justify-between items-center">
            <div className="text-sm text-gray-600">
              Showing {logs.length} of {pagination.total} logs 
              {filters.page > 1 && (
                <span className="ml-2 text-blue-600">
                  (Page {filters.page} of {pagination.totalPages})
                </span>
              )}
            </div>
            <div className="text-xs text-gray-400 flex items-center space-x-2">
              {pagination.total > 0 && (
                <span>Latest: {logs.length > 0 ? formatTimestamp(logs[0].created_at) : 'N/A'}</span>
              )}
              {autoRefresh && filters.page === 1 && (
                <div className="flex items-center space-x-1 text-green-600">
                  <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                  <span>Live</span>
                </div>
              )}
            </div>
          </div>

          {/* Loading */}
          {loading && (
            <div className="text-center py-4">
              <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
            </div>
          )}

          {/* Logs Table */}
          {!loading && (
            <div className="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
              <table className="min-w-full divide-y divide-gray-300">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Time
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Level
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Component
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Message
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Session
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {logs.map((log) => (
                    <tr key={log.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {formatTimestamp(log.created_at)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getLevelBadge(log.level)}`}>
                          {log.level?.toUpperCase()}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {log.component || '-'}
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-900 max-w-lg truncate">
                        <div title={log.message}>
                          {log.message}
                        </div>
                        {log.metadata && Object.keys(log.metadata).length > 0 && (
                          <details className="mt-1">
                            <summary className="text-xs text-gray-500 cursor-pointer">Metadata</summary>
                            <pre className="text-xs text-gray-600 mt-1 whitespace-pre-wrap">
                              {JSON.stringify(log.metadata, null, 2)}
                            </pre>
                          </details>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {log.session_id || '-'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {/* Pagination */}
          {pagination.totalPages > 1 && (
            <div className="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6 mt-4">
              <div className="flex-1 flex justify-between sm:hidden">
                <button
                  onClick={() => handlePageChange(pagination.page - 1)}
                  disabled={pagination.page <= 1}
                  className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
                >
                  Previous
                </button>
                <button
                  onClick={() => handlePageChange(pagination.page + 1)}
                  disabled={pagination.page >= pagination.totalPages}
                  className="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
                >
                  Next
                </button>
              </div>
              <div className="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
                <div>
                  <p className="text-sm text-gray-700">
                    Showing page <span className="font-medium">{pagination.page}</span> of{' '}
                    <span className="font-medium">{pagination.totalPages}</span>
                  </p>
                </div>
                <div>
                  <nav className="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
                    {Array.from({ length: Math.min(5, pagination.totalPages) }, (_, i) => {
                      let pageNum;
                      if (pagination.totalPages <= 5) {
                        pageNum = i + 1;
                      } else if (pagination.page <= 3) {
                        pageNum = i + 1;
                      } else if (pagination.page >= pagination.totalPages - 2) {
                        pageNum = pagination.totalPages - 4 + i;
                      } else {
                        pageNum = pagination.page - 2 + i;
                      }
                      
                      return (
                        <button
                          key={pageNum}
                          onClick={() => handlePageChange(pageNum)}
                          className={`relative inline-flex items-center px-4 py-2 border text-sm font-medium ${
                            pageNum === pagination.page
                              ? 'z-10 bg-indigo-50 border-indigo-500 text-indigo-600'
                              : 'bg-white border-gray-300 text-gray-500 hover:bg-gray-50'
                          }`}
                        >
                          {pageNum}
                        </button>
                      );
                    })}
                  </nav>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default LogsPage;