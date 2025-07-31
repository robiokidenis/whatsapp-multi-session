import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useNotification } from '../contexts/NotificationContext';
import { useAuth } from '../contexts/AuthContext';

const Templates = () => {
  const navigate = useNavigate();
  const { token } = useAuth();
  const { showError, showSuccess, showWarning } = useNotification();
  
  // State
  const [templates, setTemplates] = useState([]);
  const [categories, setCategories] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState('');
  const [showActiveOnly, setShowActiveOnly] = useState(true);
  
  // Preview state
  const [showPreview, setShowPreview] = useState(false);
  const [previewContent, setPreviewContent] = useState('');
  
  // Fetch templates
  const fetchTemplates = useCallback(async () => {
    try {
      setLoading(true);
      const params = new URLSearchParams();
      if (showActiveOnly) params.append('is_active', 'true');
      
      const response = await fetch(`/api/message-templates?${params}`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) {
        if (response.status === 404) {
          showWarning('Template API not yet implemented. Using demo data.');
          setTemplates(getDemoTemplates());
          return;
        }
        throw new Error('Failed to fetch templates');
      }
      
      const data = await response.json();
      setTemplates(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error('Error fetching templates:', error);
      showWarning('Using demo templates. Backend templates not available.');
      setTemplates(getDemoTemplates());
    } finally {
      setLoading(false);
    }
  }, [token, showActiveOnly, showWarning]);
  
  // Fetch categories
  const fetchCategories = useCallback(async () => {
    try {
      const response = await fetch('/api/message-templates/categories', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) {
        setCategories(getDemoCategories());
        return;
      }
      
      const data = await response.json();
      setCategories(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error('Error fetching categories:', error);
      setCategories(getDemoCategories());
    }
  }, [token]);
  
  // Demo data for when backend is not available
  const getDemoTemplates = () => [
    {
      id: 1,
      name: 'Welcome Message',
      content: 'Hello [name]! Welcome to our service. We\'re excited to have you with us!',
      type: 'text',
      category: 'Welcome',
      is_active: true,
      usage_count: 45,
      created_at: new Date().toISOString(),
      variables: [
        { name: 'name', placeholder: '[name]', required: true }
      ]
    },
    {
      id: 2,
      name: 'Appointment Reminder',
      content: 'Hi [name], this is a reminder for your appointment on [date] at [time]. Please confirm your attendance.',
      type: 'text',
      category: 'Reminders',
      is_active: true,
      usage_count: 28,
      created_at: new Date().toISOString(),
      variables: [
        { name: 'name', placeholder: '[name]', required: true },
        { name: 'date', placeholder: '[date]', required: true },
        { name: 'time', placeholder: '[time]', required: true }
      ]
    },
    {
      id: 3,
      name: 'Business Introduction',
      content: 'Hello [name], I\'m reaching out from [company]. We specialize in helping businesses like yours improve efficiency. Would you be interested in a quick call?',
      type: 'text',
      category: 'Business',
      is_active: true,
      usage_count: 62,
      created_at: new Date().toISOString(),
      variables: [
        { name: 'name', placeholder: '[name]', required: true },
        { name: 'company', placeholder: '[company]', required: true }
      ]
    },
    {
      id: 4,
      name: 'Follow Up',
      content: 'Hi [name], I hope you\'re doing well at [company]. As a [position], you might find our new solution interesting. Let me know if you\'d like more details!',
      type: 'text',
      category: 'Follow Up',
      is_active: true,
      usage_count: 15,
      created_at: new Date().toISOString(),
      variables: [
        { name: 'name', placeholder: '[name]', required: true },
        { name: 'company', placeholder: '[company]', required: true },
        { name: 'position', placeholder: '[position]', required: true }
      ]
    }
  ];
  
  const getDemoCategories = () => [
    'Welcome', 'Reminders', 'Thank You', 'Follow Up', 'Marketing', 'Support'
  ];
  
  // Initial load
  useEffect(() => {
    fetchTemplates();
    fetchCategories();
  }, [fetchTemplates, fetchCategories]);
  
  // Filter templates
  const filteredTemplates = templates.filter(template => {
    if (searchQuery && !template.name.toLowerCase().includes(searchQuery.toLowerCase()) &&
        !template.content.toLowerCase().includes(searchQuery.toLowerCase())) {
      return false;
    }
    if (selectedCategory && template.category !== selectedCategory) {
      return false;
    }
    return true;
  });
  
  // Handle create template
  const handleCreateTemplate = () => {
    navigate('/templates/new');
  };
  
  // Handle delete template
  const handleDeleteTemplate = async (template) => {
    if (!window.confirm(`Delete template "${template.name}"?`)) return;
    
    try {
      const response = await fetch(`/api/message-templates/${template.id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) {
        showWarning('Template deleted locally. Backend API not available.');
      } else {
        showSuccess('Template deleted successfully');
      }
      
      setTemplates(prev => prev.filter(t => t.id !== template.id));
    } catch (error) {
      showError('Failed to delete template');
      console.error('Error deleting template:', error);
    }
  };
  
  // Handle preview template
  const handlePreviewTemplate = (template) => {
    let content = template.content;
    
    // Get user's timezone from settings or localStorage
    const userTimezone = localStorage.getItem('userTimezone') || Intl.DateTimeFormat().resolvedOptions().timeZone;
    
    // Generate current date/time in user's timezone
    const now = new Date();
    const currentDate = new Intl.DateTimeFormat('en-CA', {
      timeZone: userTimezone,
      year: 'numeric',
      month: '2-digit',
      day: '2-digit'
    }).format(now);
    
    const currentTime = new Intl.DateTimeFormat('en-US', {
      timeZone: userTimezone,
      hour: '2-digit',
      minute: '2-digit',
      hour12: true
    }).format(now);
    
    // Replace contact model variables with sample data
    content = content.replace(/\[name\]/g, 'John Doe');
    content = content.replace(/\[phone\]/g, '+1 (555) 123-4567');
    content = content.replace(/\[email\]/g, 'john.doe@example.com');
    content = content.replace(/\[company\]/g, 'Acme Corp');
    content = content.replace(/\[position\]/g, 'Marketing Manager');
    content = content.replace(/\[date\]/g, currentDate);
    content = content.replace(/\[time\]/g, currentTime);
    
    setPreviewContent(content);
    setShowPreview(true);
  };
  
  // Handle edit template
  const handleEditTemplate = (template) => {
    navigate(`/templates/edit/${template.id}`);
  };
  
  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-2">Message Templates</h1>
        <p className="text-gray-600">Create and manage reusable message templates for your campaigns</p>
      </div>
      
      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-white rounded-lg shadow p-4">
          <div className="text-sm text-gray-500">Total Templates</div>
          <div className="text-2xl font-bold text-gray-900">{templates.length}</div>
        </div>
        <div className="bg-white rounded-lg shadow p-4">
          <div className="text-sm text-gray-500">Active Templates</div>
          <div className="text-2xl font-bold text-green-600">{templates.filter(t => t.is_active).length}</div>
        </div>
        <div className="bg-white rounded-lg shadow p-4">
          <div className="text-sm text-gray-500">Categories</div>
          <div className="text-2xl font-bold text-blue-600">{categories.length}</div>
        </div>
        <div className="bg-white rounded-lg shadow p-4">
          <div className="text-sm text-gray-500">Total Usage</div>
          <div className="text-2xl font-bold text-purple-600">{templates.reduce((sum, t) => sum + (t.usage_count || 0), 0)}</div>
        </div>
      </div>
      
      {/* Actions and Filters */}
      <div className="bg-white rounded-lg shadow mb-6">
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            {/* Left side - Filters */}
            <div className="flex flex-col sm:flex-row gap-4">
              <div className="w-full sm:w-80">
                <input
                  type="text"
                  placeholder="Search templates..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                />
              </div>
              
              <select
                value={selectedCategory}
                onChange={(e) => setSelectedCategory(e.target.value)}
                className="px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
              >
                <option value="">All Categories</option>
                {categories.map(category => (
                  <option key={category} value={category}>{category}</option>
                ))}
              </select>
              
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={showActiveOnly}
                  onChange={(e) => setShowActiveOnly(e.target.checked)}
                  className="mr-2 text-primary-600 focus:ring-primary-500"
                />
                <span className="text-sm text-gray-700">Active only</span>
              </label>
            </div>
            
            {/* Right side - Actions */}
            <button
              onClick={handleCreateTemplate}
              className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors flex items-center gap-2"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              Create Template
            </button>
          </div>
        </div>
      </div>
      
      {/* Templates Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {loading ? (
          <div className="col-span-full flex justify-center items-center py-12">
            <div className="animate-spin h-8 w-8 border-2 border-primary-600 border-t-transparent rounded-full"></div>
          </div>
        ) : filteredTemplates.length === 0 ? (
          <div className="col-span-full text-center py-12">
            <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <h3 className="mt-2 text-sm font-medium text-gray-900">No templates found</h3>
            <p className="mt-1 text-sm text-gray-500">Get started by creating a new template.</p>
            <button
              onClick={handleCreateTemplate}
              className="mt-6 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
            >
              Create Template
            </button>
          </div>
        ) : (
          filteredTemplates.map(template => (
            <div key={template.id} className="bg-white rounded-lg shadow hover:shadow-md transition-shadow">
              <div className="p-6">
                <div className="flex items-start justify-between mb-4">
                  <div className="flex-1">
                    <h3 className="text-lg font-medium text-gray-900 mb-1">{template.name}</h3>
                    <div className="flex items-center gap-2">
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                        {template.type}
                      </span>
                      {template.category && (
                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                          {template.category}
                        </span>
                      )}
                      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                        template.is_active 
                          ? 'bg-green-100 text-green-800' 
                          : 'bg-red-100 text-red-800'
                      }`}>
                        {template.is_active ? 'Active' : 'Inactive'}
                      </span>
                    </div>
                  </div>
                  
                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => handlePreviewTemplate(template)}
                      className="p-2 text-gray-400 hover:text-gray-600 transition-colors"
                      title="Preview"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                      </svg>
                    </button>
                    <button
                      onClick={() => handleEditTemplate(template)}
                      className="p-2 text-gray-400 hover:text-gray-600 transition-colors"
                      title="Edit"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                      </svg>
                    </button>
                    <button
                      onClick={() => handleDeleteTemplate(template)}
                      className="p-2 text-gray-400 hover:text-red-600 transition-colors"
                      title="Delete"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    </button>
                  </div>
                </div>
                
                <div className="mb-4">
                  <p className="text-sm text-gray-600 line-clamp-3">{template.content}</p>
                </div>
                
                <div className="flex items-center justify-between text-sm text-gray-500">
                  <span>Used {template.usage_count || 0} times</span>
                  <span>{new Date(template.created_at).toLocaleDateString()}</span>
                </div>
                
                {template.variables && template.variables.length > 0 && (
                  <div className="mt-3 pt-3 border-t border-gray-200">
                    <div className="text-xs text-gray-500 mb-1">Variables:</div>
                    <div className="flex flex-wrap gap-1">
                      {template.variables.map((variable, index) => (
                        <span key={index} className="inline-flex items-center px-2 py-1 rounded text-xs bg-gray-100 text-gray-700">
                          {variable.placeholder || `[${variable.name}]`}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
          ))
        )}
      </div>
      
      {/* Preview Modal */}
      {showPreview && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg max-w-2xl w-full">
            <div className="px-6 py-4 border-b border-gray-200">
              <div className="flex items-center justify-between">
                <h3 className="text-lg font-medium text-gray-900">Template Preview</h3>
                <button
                  onClick={() => setShowPreview(false)}
                  className="text-gray-400 hover:text-gray-600"
                >
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
            
            <div className="p-6">
              <div className="bg-gray-50 rounded-lg p-4 border-l-4 border-primary-500">
                <div className="text-sm text-gray-600 mb-2">Preview with sample data:</div>
                <div className="text-gray-900 whitespace-pre-wrap">{previewContent}</div>
              </div>
            </div>
            
            <div className="px-6 py-4 bg-gray-50 border-t border-gray-200 flex justify-end">
              <button
                onClick={() => setShowPreview(false)}
                className="px-4 py-2 bg-primary-600 text-white rounded-md text-sm font-medium hover:bg-primary-700"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Templates;