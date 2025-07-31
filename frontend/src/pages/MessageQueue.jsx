import React, { useState, useEffect, useCallback } from 'react';
import { useNotification } from '../contexts/NotificationContext';
import { useAuth } from '../contexts/AuthContext';

const MessageQueue = () => {
  const { token } = useAuth();
  const { showError, showSuccess, showWarning } = useNotification();
  
  // State
  const [jobs, setJobs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedJob, setSelectedJob] = useState(null);
  const [showPreviewModal, setShowPreviewModal] = useState(false);
  const [jobDetails, setJobDetails] = useState(null);
  const [refreshInterval, setRefreshInterval] = useState(null);
  
  // Filters
  const [statusFilter, setStatusFilter] = useState('all');
  const [searchQuery, setSearchQuery] = useState('');

  // Fetch bulk messaging jobs
  const fetchJobs = useCallback(async () => {
    try {
      const response = await fetch('/api/bulk-messages', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to fetch jobs');
      
      const data = await response.json();
      setJobs(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error('Error fetching jobs:', error);
      showError('Failed to load message queue');
    } finally {
      setLoading(false);
    }
  }, [token, showError]);

  // Fetch job details
  const fetchJobDetails = useCallback(async (jobId) => {
    try {
      const response = await fetch(`/api/bulk-messages/${jobId}`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to fetch job details');
      
      const data = await response.json();
      setJobDetails(data);
    } catch (error) {
      console.error('Error fetching job details:', error);
      showError('Failed to load job details');
    }
  }, [token, showError]);

  // Cancel job
  const cancelJob = async (jobId) => {
    if (!window.confirm('Are you sure you want to cancel this messaging job?')) {
      return;
    }

    try {
      const response = await fetch(`/api/bulk-messages/${jobId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to cancel job');
      
      showSuccess('Job cancelled successfully');
      fetchJobs(); // Refresh the list
    } catch (error) {
      console.error('Error cancelling job:', error);
      showError('Failed to cancel job');
    }
  };

  // Initial load and auto-refresh
  useEffect(() => {
    fetchJobs();
    
    // Set up auto-refresh every 5 seconds for active jobs
    const interval = setInterval(fetchJobs, 5000);
    setRefreshInterval(interval);
    
    return () => {
      if (interval) clearInterval(interval);
    };
  }, []); // Empty dependency array to run only once

  // Handle job preview
  const handlePreview = (job) => {
    setSelectedJob(job);
    fetchJobDetails(job.job_id || job.id);
    setShowPreviewModal(true);
  };

  // Filter jobs based on status and search
  const filteredJobs = jobs.filter(job => {
    const matchesStatus = statusFilter === 'all' || job.status === statusFilter;
    const matchesSearch = !searchQuery || 
      (job.job_id && job.job_id.toLowerCase().includes(searchQuery.toLowerCase())) ||
      (job.template_name && job.template_name.toLowerCase().includes(searchQuery.toLowerCase()));
    
    return matchesStatus && matchesSearch;
  });

  // Get status color
  const getStatusColor = (status) => {
    switch (status) {
      case 'pending': return 'bg-yellow-100 text-yellow-800';
      case 'running': return 'bg-blue-100 text-blue-800';
      case 'completed': return 'bg-green-100 text-green-800';
      case 'failed': return 'bg-red-100 text-red-800';
      case 'cancelled': return 'bg-gray-100 text-gray-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  // Get progress percentage
  const getProgressPercentage = (job) => {
    const total = job.total_contacts || 0;
    const sent = job.sent_count || 0;
    const failed = job.failed_count || 0;
    const processed = sent + failed;
    
    return total > 0 ? Math.round((processed / total) * 100) : 0;
  };

  // Format date
  const formatDate = (dateString) => {
    if (!dateString) return '-';
    return new Date(dateString).toLocaleString();
  };

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 mb-2">Message Queue</h1>
            <p className="text-gray-600">Monitor and manage your bulk messaging jobs</p>
          </div>
          <button
            onClick={fetchJobs}
            className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors flex items-center gap-2"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            Refresh
          </button>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg shadow mb-6">
        <div className="p-4 border-b border-gray-200">
          <div className="flex flex-col md:flex-row gap-4">
            <div className="flex-1">
              <input
                type="text"
                placeholder="Search by Job ID or Template name..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
              />
            </div>
            <div>
              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                className="px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
              >
                <option value="all">All Status</option>
                <option value="pending">Pending</option>
                <option value="running">Running</option>
                <option value="completed">Completed</option>
                <option value="failed">Failed</option>
                <option value="cancelled">Cancelled</option>
              </select>
            </div>
          </div>
        </div>
      </div>

      {/* Jobs List */}
      <div className="bg-white rounded-lg shadow">
        {loading ? (
          <div className="p-8 text-center">
            <div className="animate-spin h-8 w-8 border-2 border-primary-600 border-t-transparent rounded-full mx-auto mb-4"></div>
            <p className="text-gray-500">Loading message queue...</p>
          </div>
        ) : filteredJobs.length === 0 ? (
          <div className="p-8 text-center">
            <svg className="w-12 h-12 text-gray-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2M4 13h2m8-8v12M9 17l6-6" />
            </svg>
            <p className="text-gray-500 mb-2">No messaging jobs found</p>
            <p className="text-sm text-gray-400">Start a bulk messaging campaign to see jobs here</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Job Info
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Progress
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Recipients
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Started
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {filteredJobs.map((job) => {
                  const progress = getProgressPercentage(job);
                  return (
                    <tr key={job.job_id || job.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div>
                          <div className="text-sm font-medium text-gray-900">
                            {job.job_id || job.id || 'Unknown'}
                          </div>
                          <div className="text-sm text-gray-500">
                            {job.template_name || 'Template N/A'}
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getStatusColor(job.status)}`}>
                          {job.status || 'unknown'}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center">
                          <div className="flex-1">
                            <div className="flex justify-between text-sm mb-1">
                              <span className="text-gray-600">{progress}%</span>
                              <span className="text-gray-500">
                                {job.sent_count || 0}/{job.total_contacts || 0}
                              </span>
                            </div>
                            <div className="w-full bg-gray-200 rounded-full h-2">
                              <div
                                className={`h-2 rounded-full transition-all duration-300 ${
                                  job.status === 'completed' ? 'bg-green-500' :
                                  job.status === 'failed' ? 'bg-red-500' :
                                  job.status === 'running' ? 'bg-blue-500' :
                                  'bg-yellow-500'
                                }`}
                                style={{ width: `${progress}%` }}
                              ></div>
                            </div>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        <div>
                          <div className="text-green-600">✓ {job.sent_count || 0} sent</div>
                          {(job.failed_count || 0) > 0 && (
                            <div className="text-red-600">✗ {job.failed_count} failed</div>
                          )}
                          {(job.pending_count || 0) > 0 && (
                            <div className="text-yellow-600">⏳ {job.pending_count} pending</div>
                          )}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {formatDate(job.started_at || job.created_at)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                        <div className="flex gap-2">
                          <button
                            onClick={() => handlePreview(job)}
                            className="text-primary-600 hover:text-primary-900"
                          >
                            View
                          </button>
                          {(job.status === 'pending' || job.status === 'running') && (
                            <button
                              onClick={() => cancelJob(job.job_id || job.id)}
                              className="text-red-600 hover:text-red-900"
                            >
                              Cancel
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Preview Modal */}
      {showPreviewModal && selectedJob && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg max-w-4xl w-full max-h-[80vh] overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
              <h3 className="text-lg font-medium text-gray-900">
                Job Details: {selectedJob.job_id || selectedJob.id}
              </h3>
              <button
                onClick={() => setShowPreviewModal(false)}
                className="text-gray-400 hover:text-gray-600"
              >
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            
            <div className="p-6 overflow-y-auto max-h-[60vh]">
              {jobDetails ? (
                <div className="space-y-6">
                  {/* Job Summary */}
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className="bg-gray-50 p-4 rounded-lg">
                      <h4 className="font-medium text-gray-900 mb-2">Status</h4>
                      <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getStatusColor(jobDetails.status)}`}>
                        {jobDetails.status || 'unknown'}
                      </span>
                    </div>
                    <div className="bg-gray-50 p-4 rounded-lg">
                      <h4 className="font-medium text-gray-900 mb-2">Progress</h4>
                      <div className="text-2xl font-bold text-primary-600">
                        {getProgressPercentage(jobDetails)}%
                      </div>
                      <div className="text-sm text-gray-500">
                        {jobDetails.sent_count || 0} of {jobDetails.total_contacts || 0} sent
                      </div>
                    </div>
                    <div className="bg-gray-50 p-4 rounded-lg">
                      <h4 className="font-medium text-gray-900 mb-2">Timeline</h4>
                      <div className="text-sm text-gray-600">
                        <div>Started: {formatDate(jobDetails.started_at)}</div>
                        {jobDetails.completed_at && (
                          <div>Completed: {formatDate(jobDetails.completed_at)}</div>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* Message Preview */}
                  {jobDetails.message_content && (
                    <div>
                      <h4 className="font-medium text-gray-900 mb-3">Message Preview</h4>
                      <div className="bg-gray-50 p-4 rounded-lg border">
                        <div className="whitespace-pre-wrap text-sm">
                          {jobDetails.message_content}
                        </div>
                      </div>
                    </div>
                  )}

                  {/* Recipients Details */}
                  {jobDetails.messages && jobDetails.messages.length > 0 && (
                    <div>
                      <h4 className="font-medium text-gray-900 mb-3">Recipients Status</h4>
                      <div className="bg-gray-50 rounded-lg max-h-64 overflow-y-auto">
                        <table className="min-w-full">
                          <thead className="bg-gray-100 sticky top-0">
                            <tr>
                              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Contact</th>
                              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Sent At</th>
                            </tr>
                          </thead>
                          <tbody className="divide-y divide-gray-200">
                            {jobDetails.messages.map((message, index) => (
                              <tr key={index}>
                                <td className="px-4 py-2 text-sm">
                                  <div>
                                    <div className="font-medium">{message.contact_name || 'Unknown'}</div>
                                    <div className="text-gray-500">{message.contact_phone}</div>
                                  </div>
                                </td>
                                <td className="px-4 py-2">
                                  <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getStatusColor(message.status)}`}>
                                    {message.status}
                                  </span>
                                </td>
                                <td className="px-4 py-2 text-sm text-gray-500">
                                  {formatDate(message.sent_at)}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  )}
                </div>
              ) : (
                <div className="text-center py-8">
                  <div className="animate-spin h-8 w-8 border-2 border-primary-600 border-t-transparent rounded-full mx-auto mb-4"></div>
                  <p className="text-gray-500">Loading job details...</p>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default MessageQueue;