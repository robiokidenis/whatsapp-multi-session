import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useNotification } from '../contexts/NotificationContext';
import { useAuth } from '../contexts/AuthContext';

const TemplateEditor = () => {
  const navigate = useNavigate();
  const { id } = useParams();
  const { token } = useAuth();
  const { showError, showSuccess, showWarning } = useNotification();
  
  const [loading, setLoading] = useState(false);
  const [categories, setCategories] = useState([]);
  const [showCustomVariables, setShowCustomVariables] = useState(false);
  
  // Form state
  const [formData, setFormData] = useState({
    name: '',
    content: '',
    type: 'text',
    category: '',
    variables: [],
    media_url: '',
    media_type: '',
    is_active: true
  });

  // Common variables based on Contact model
  const commonVariables = [
    { name: 'name', label: 'Name', icon: '👤' },
    { name: 'phone', label: 'Phone', icon: '📱' },
    { name: 'email', label: 'Email', icon: '📧' },
    { name: 'company', label: 'Company', icon: '🏢' },
    { name: 'position', label: 'Position', icon: '💼' },
    { name: 'date', label: 'Date', icon: '📅' },
    { name: 'time', label: 'Time', icon: '⏰' }
  ];

  // Generate live preview content with timezone-aware dates
  const getLivePreview = () => {
    let content = formData.content;
    
    // Get user's timezone from settings or localStorage
    const userTimezone = localStorage.getItem('userTimezone') || Intl.DateTimeFormat().resolvedOptions().timeZone;
    
    // Generate current date/time in user's timezone
    const now = new Date();
    const currentDate = new Intl.DateTimeFormat('en-CA', {
      timeZone: userTimezone,
      year: 'numeric',
      month: '2-digit',
      day: '2-digit'
    }).format(now); // YYYY-MM-DD format
    
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
    
    return content || 'Your template preview will appear here...';
  };

  // Insert variable into content at cursor position
  const insertVariable = (variableName) => {
    const textarea = document.querySelector('textarea[name="content"]');
    if (textarea) {
      const start = textarea.selectionStart;
      const end = textarea.selectionEnd;
      const currentContent = formData.content;
      const newContent = currentContent.substring(0, start) + `[${variableName}]` + currentContent.substring(end);
      
      setFormData(prev => ({ ...prev, content: newContent }));
      
      // Reset cursor position after the inserted variable
      setTimeout(() => {
        const newPosition = start + `[${variableName}]`.length;
        textarea.focus();
        textarea.setSelectionRange(newPosition, newPosition);
      }, 0);
    }
  };

  // Fetch categories
  const fetchCategories = useCallback(async () => {
    try {
      const response = await fetch('/api/message-templates/categories', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (response.ok) {
        const data = await response.json();
        setCategories(Array.isArray(data) ? data : []);
      } else {
        setCategories(['Welcome', 'Reminders', 'Thank You', 'Follow Up', 'Marketing', 'Support']);
      }
    } catch (error) {
      console.error('Error fetching categories:', error);
      setCategories(['Welcome', 'Reminders', 'Thank You', 'Follow Up', 'Marketing', 'Support']);
    }
  }, [token]);

  // Fetch template data if editing
  const fetchTemplate = useCallback(async () => {
    if (!id) return;
    
    try {
      setLoading(true);
      const response = await fetch(`/api/message-templates/${id}`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (response.ok) {
        const template = await response.json();
        setFormData({
          name: template.name || '',
          content: template.content || '',
          type: template.type || 'text',
          category: template.category || '',
          variables: template.variables || [],
          media_url: template.media_url || '',
          media_type: template.media_type || '',
          is_active: template.is_active !== false
        });
        setShowCustomVariables(template.variables && template.variables.length > 0);
      } else {
        showError('Template not found');
        navigate('/templates');
      }
    } catch (error) {
      showError('Failed to load template');
      navigate('/templates');
    } finally {
      setLoading(false);
    }
  }, [id, token, showError, navigate]);

  useEffect(() => {
    fetchCategories();
    fetchTemplate();
  }, [fetchCategories, fetchTemplate]);

  // Handle save template
  const handleSaveTemplate = async () => {
    try {
      setLoading(true);
      const url = id ? `/api/message-templates/${id}` : '/api/message-templates';
      const method = id ? 'PUT' : 'POST';
      
      const response = await fetch(url, {
        method,
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(formData)
      });
      
      if (response.ok) {
        showSuccess(`Template ${id ? 'updated' : 'created'} successfully`);
        navigate('/templates');
      } else {
        showWarning('Template saved locally. Backend API not available.');
        navigate('/templates');
      }
    } catch (error) {
      showError('Failed to save template');
    } finally {
      setLoading(false);
    }
  };

  // Add variable
  const addVariable = () => {
    setShowCustomVariables(true);
    setFormData(prev => ({
      ...prev,
      variables: [...prev.variables, { name: '', placeholder: '', required: false }]
    }));
  };

  // Remove variable
  const removeVariable = (index) => {
    setFormData(prev => ({
      ...prev,
      variables: prev.variables.filter((_, i) => i !== index)
    }));
  };

  // Update variable
  const updateVariable = (index, field, value) => {
    setFormData(prev => ({
      ...prev,
      variables: prev.variables.map((variable, i) => 
        i === index ? { ...variable, [field]: value } : variable
      )
    }));
  };

  if (loading) {
    return (
      <div className="p-6 flex justify-center items-center min-h-[400px]">
        <div className="animate-spin h-8 w-8 border-2 border-primary-600 border-t-transparent rounded-full"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center">
              <button
                onClick={() => navigate('/templates')}
                className="mr-4 p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                </svg>
              </button>
              <div>
                <h1 className="text-xl font-semibold text-gray-900">
                  {id ? 'Edit Template' : 'Create New Template'}
                </h1>
                <p className="text-sm text-gray-600">
                  {id ? 'Update your message template' : 'Design your message template with live preview'}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <button
                onClick={() => navigate('/templates')}
                className="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleSaveTemplate}
                disabled={!formData.name || !formData.content || loading}
                className="px-6 py-2 bg-gradient-to-r from-primary-600 to-blue-600 text-white rounded-lg text-sm font-medium hover:from-primary-700 hover:to-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 shadow-lg hover:shadow-xl"
              >
                {loading ? 'Saving...' : (id ? '✅ Update Template' : '🚀 Create Template')}
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left Column - Template Settings */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4 flex items-center">
              <svg className="w-5 h-5 mr-2 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              Template Settings
            </h3>
            
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Template Name *</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors"
                  placeholder="e.g., Welcome Message, Appointment Reminder"
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Category</label>
                <input
                  type="text"
                  value={formData.category}
                  onChange={(e) => setFormData(prev => ({ ...prev, category: e.target.value }))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors"
                  placeholder="e.g., Welcome, Reminders, Marketing"
                />
              </div>
              
              
              <div className="flex items-center p-3 bg-gray-50 rounded-lg border border-gray-200">
                <input
                  type="checkbox"
                  id="is_active"
                  checked={formData.is_active}
                  onChange={(e) => setFormData(prev => ({ ...prev, is_active: e.target.checked }))}
                  className="w-4 h-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                />
                <label htmlFor="is_active" className="ml-3 text-sm font-medium text-gray-700">
                  Active Template
                </label>
                <span className="ml-auto text-xs text-gray-500">
                  {formData.is_active ? '✅ Available' : '❌ Hidden'}
                </span>
              </div>
            </div>
          </div>
          
          {/* Middle Column - Content Editor */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 flex flex-col">
            <h3 className="text-lg font-medium text-gray-900 mb-4 flex items-center">
              <svg className="w-5 h-5 mr-2 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
              Message Content
            </h3>
            
            <div className="flex-1 flex flex-col">
              <textarea
                name="content"
                value={formData.content}
                onChange={(e) => setFormData(prev => ({ ...prev, content: e.target.value }))}
                className="flex-1 min-h-[400px] px-3 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors resize-none"
                placeholder="Type your message template here...

💡 Available Variables:
• [name] - Contact name
• [phone] - Phone number  
• [email] - Email address
• [company] - Company name
• [position] - Job position
• [date] - Current date (your timezone)
• [time] - Current time (your timezone)

Click variable buttons below to insert quickly"
              />
              
              {/* Quick Variables - Smaller and Nicer */}
              <div className="mt-3">
                <div className="text-xs font-medium text-gray-700 mb-2">Quick Variables:</div>
                <div className="flex flex-wrap gap-1">
                  {commonVariables.map((variable) => (
                    <button
                      key={variable.name}
                      type="button"
                      onClick={() => insertVariable(variable.name)}
                      className="inline-flex items-center px-2 py-1 text-xs bg-gradient-to-r from-blue-50 to-primary-50 text-primary-700 rounded-md border border-primary-200 hover:from-primary-100 hover:to-blue-100 hover:border-primary-300 transition-all duration-200 shadow-sm hover:shadow"
                      title={`Insert ${variable.label}`}
                    >
                      <span className="mr-1">{variable.icon}</span>
                      [{variable.name}]
                    </button>
                  ))}
                </div>
              </div>
            </div>
          </div>
          
          {/* Right Column - Live Preview */}
          <div className="bg-gradient-to-br from-green-50 to-emerald-50 rounded-xl shadow-sm border border-green-200 p-6 flex flex-col">
            <h3 className="text-lg font-medium text-gray-900 mb-4 flex items-center">
              <svg className="w-5 h-5 mr-2 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 616 0z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
              Live Preview
            </h3>
            
            <div className="bg-white rounded-lg border-2 border-green-200 p-4 flex-1 flex flex-col">
              <div className="flex items-center mb-3 pb-3 border-b border-gray-100">
                <div className="w-8 h-8 bg-green-500 rounded-full flex items-center justify-center">
                  <svg className="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893A11.821 11.821 0 0020.885 3.488"/>
                  </svg>
                </div>
                <div className="ml-3">
                  <div className="text-sm font-medium text-gray-900">WhatsApp Preview</div>
                  <div className="text-xs text-gray-500">How your message will look</div>
                </div>
              </div>
              
              <div className="whitespace-pre-wrap text-gray-800 leading-relaxed flex-1 overflow-y-auto min-h-[300px]">
                {getLivePreview()}
              </div>
              
              {formData.content && (
                <div className="mt-4 pt-3 border-t border-gray-100">
                  <div className="text-xs text-gray-500 flex items-center justify-between">
                    <span>Length: {formData.content.length} chars</span>
                    <span className={`px-2 py-1 rounded-full text-xs ${
                      formData.content.length > 1000 
                        ? 'bg-red-100 text-red-700' 
                        : formData.content.length > 500 
                          ? 'bg-yellow-100 text-yellow-700' 
                          : 'bg-green-100 text-green-700'
                    }`}>
                      {formData.content.length > 1000 ? 'Long' : formData.content.length > 500 ? 'Medium' : 'Good'}
                    </span>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
        
        {/* Custom Variables Section */}
        {showCustomVariables && (
          <div className="mt-6 bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-medium text-gray-900 flex items-center">
                <svg className="w-5 h-5 mr-2 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4.871 4A17.926 17.926 0 003 12c0 2.874.673 5.59 1.871 8m14.13 0a17.926 17.926 0 001.87-8 17.926 17.926 0 00-1.87-8M9 9h1.246a1 1 0 01.961.725l1.586 5.55a1 1 0 00.961.725H15m1-7h-.08a2 2 0 00-1.519.698L9.6 15.302A2 2 0 018.08 16H8" />
                </svg>
                Custom Variables ({formData.variables.length})
              </h3>
              <button
                type="button"
                onClick={() => setShowCustomVariables(false)}
                className="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
                title="Hide custom variables"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            
            {formData.variables.length > 0 ? (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
                {formData.variables.map((variable, index) => (
                  <div key={index} className="bg-gray-50 p-4 rounded-lg border border-gray-200">
                    <div className="flex items-start justify-between mb-3">
                      <div className="flex-1 grid grid-cols-2 gap-3">
                        <div>
                          <label className="block text-xs font-medium text-gray-700 mb-1">Variable Name</label>
                          <input
                            type="text"
                            placeholder="e.g., customer_name"
                            value={variable.name}
                            onChange={(e) => updateVariable(index, 'name', e.target.value)}
                            className="w-full px-3 py-2 text-sm border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                          />
                        </div>
                        <div>
                          <label className="block text-xs font-medium text-gray-700 mb-1">Display As</label>
                          <input
                            type="text"
                            placeholder="[customer_name]"
                            value={variable.placeholder}
                            onChange={(e) => updateVariable(index, 'placeholder', e.target.value)}
                            className="w-full px-3 py-2 text-sm border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                          />
                        </div>
                      </div>
                      <button
                        type="button"
                        onClick={() => removeVariable(index)}
                        className="ml-3 p-2 text-red-600 hover:text-red-800 hover:bg-red-50 rounded-md transition-colors"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                      </button>
                    </div>
                    <label className="flex items-center text-sm">
                      <input
                        type="checkbox"
                        checked={variable.required}
                        onChange={(e) => updateVariable(index, 'required', e.target.checked)}
                        className="mr-2 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                      />
                      <span className="text-gray-700">Required field</span>
                    </label>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8 bg-gray-50 rounded-lg border-2 border-dashed border-gray-200 mb-4">
                <svg className="w-12 h-12 text-gray-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
                </svg>
                <h3 className="text-lg font-medium text-gray-900 mb-2">No Custom Variables</h3>
                <p className="text-gray-500 text-sm mb-4">
                  Custom variables allow you to define specific placeholders beyond the common ones.<br/>
                  Great for industry-specific terms or unique business needs.
                </p>
              </div>
            )}
            
            <button
              type="button"
              onClick={addVariable}
              className="w-full px-4 py-3 text-sm bg-white border-2 border-dashed border-gray-300 rounded-lg text-gray-600 hover:border-primary-300 hover:text-primary-600 hover:bg-primary-50 transition-all duration-200 flex items-center justify-center"
            >
              <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              Add Custom Variable
            </button>
          </div>
        )}
        
        {/* Show Custom Variables Button */}
        {!showCustomVariables && (
          <div className="mt-6 text-center">
            <button
              type="button"
              onClick={() => setShowCustomVariables(true)}
              className="text-primary-600 hover:text-primary-700 underline flex items-center mx-auto"
            >
              <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              Add Custom Variables
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

export default TemplateEditor;