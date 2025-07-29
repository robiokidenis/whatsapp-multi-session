import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { useNotification } from '../contexts/NotificationContext';

const MessageTemplatesModal = ({ isOpen, onClose }) => {
  const { token } = useAuth();
  const { showNotification } = useNotification();
  
  const [templates, setTemplates] = useState([]);
  const [loading, setLoading] = useState(true);
  const [editingTemplate, setEditingTemplate] = useState(null);
  const [showForm, setShowForm] = useState(false);
  const [previewTemplate, setPreviewTemplate] = useState(null);
  const [categories, setCategories] = useState([]);
  
  const templateTypes = [
    { value: 'text', label: 'Text Message' },
    { value: 'image', label: 'Image Message' },
    { value: 'document', label: 'Document Message' }
  ];
  
  const commonVariables = [
    { name: 'name', placeholder: '{{name}}', description: 'Contact name' },
    { name: 'phone', placeholder: '{{phone}}', description: 'Contact phone' },
    { name: 'email', placeholder: '{{email}}', description: 'Contact email' },
    { name: 'company', placeholder: '{{company}}', description: 'Contact company' },
    { name: 'position', placeholder: '{{position}}', description: 'Contact position' }
  ];
  
  useEffect(() => {
    if (isOpen) {
      fetchTemplates();
      fetchCategories();
    }
  }, [isOpen]);
  
  const fetchTemplates = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/message-templates', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to fetch templates');
      
      const data = await response.json();
      setTemplates(data || []);
    } catch (error) {
      showNotification('Failed to load templates', 'error');
      console.error('Error fetching templates:', error);
    } finally {
      setLoading(false);
    }
  };
  
  const fetchCategories = async () => {
    try {
      const response = await fetch('/api/message-templates/categories', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to fetch categories');
      
      const data = await response.json();
      setCategories(data || []);
    } catch (error) {
      console.error('Error fetching categories:', error);
    }
  };
  
  const handleSaveTemplate = async (templateData) => {
    try {
      const url = editingTemplate 
        ? `/api/message-templates/${editingTemplate.id}`
        : '/api/message-templates';
      
      const response = await fetch(url, {
        method: editingTemplate ? 'PUT' : 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(templateData)
      });
      
      if (!response.ok) throw new Error('Failed to save template');
      
      showNotification(`Template ${editingTemplate ? 'updated' : 'created'} successfully`, 'success');
      setEditingTemplate(null);
      setShowForm(false);
      fetchTemplates();
      fetchCategories();
    } catch (error) {
      showNotification('Failed to save template', 'error');
      console.error('Error saving template:', error);
    }
  };
  
  const handleDeleteTemplate = async (templateId) => {
    if (!window.confirm('Delete this template?')) {
      return;
    }
    
    try {
      const response = await fetch(`/api/message-templates/${templateId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to delete template');
      
      showNotification('Template deleted successfully', 'success');
      fetchTemplates();
      fetchCategories();
    } catch (error) {
      showNotification('Failed to delete template', 'error');
      console.error('Error deleting template:', error);
    }
  };
  
  const handlePreviewTemplate = async (template) => {
    try {
      const response = await fetch('/api/message-templates/preview', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          template_id: template.id,
          variables: {
            name: 'John Doe',
            phone: '+1234567890',
            email: 'john@example.com',
            company: 'Example Corp',
            position: 'Manager'
          }
        })
      });
      
      if (!response.ok) throw new Error('Failed to preview template');
      
      const data = await response.json();
      setPreviewTemplate({ ...template, preview: data.content });
    } catch (error) {
      showNotification('Failed to preview template', 'error');
      console.error('Error previewing template:', error);
    }
  };
  
  if (!isOpen) return null;
  
  return (
    <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg max-w-6xl w-full max-h-[90vh] overflow-y-auto">
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-medium text-gray-900">Message Templates</h3>
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
          {!showForm && !previewTemplate ? (
            <>
              {/* Templates List */}
              <div className="flex justify-between items-center mb-6">
                <h4 className="text-md font-medium text-gray-900">Templates ({templates.length})</h4>
                <button
                  onClick={() => {
                    setEditingTemplate(null);
                    setShowForm(true);
                  }}
                  className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 flex items-center gap-2"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                  </svg>
                  Add Template
                </button>
              </div>
              
              {loading ? (
                <div className="flex justify-center items-center py-8">
                  <svg className="animate-spin h-8 w-8 text-primary-600" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                </div>
              ) : templates.length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                  <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  <p className="mt-2">No message templates yet</p>
                  <p className="text-sm">Create your first template for bulk messaging</p>
                </div>
              ) : (
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                  {templates.map(template => (
                    <div key={template.id} className="bg-white border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow">
                      <div className="flex items-start justify-between mb-3">
                        <div className="flex-1">
                          <div className="flex items-center gap-2 mb-1">
                            <h5 className="font-medium text-gray-900">{template.name}</h5>
                            <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                              template.type === 'text' ? 'bg-blue-100 text-blue-800' :
                              template.type === 'image' ? 'bg-green-100 text-green-800' :
                              'bg-purple-100 text-purple-800'
                            }`}>
                              {template.type}
                            </span>
                            {template.category && (
                              <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-700">
                                {template.category}
                              </span>
                            )}
                          </div>
                          
                          <div className="text-sm text-gray-600 mb-2 line-clamp-3">
                            {template.content}
                          </div>
                          
                          <div className="flex items-center gap-4 text-sm text-gray-500">
                            <span>Used {template.usage_count || 0} times</span>
                            <span className={template.is_active ? 'text-green-600' : 'text-red-600'}>
                              {template.is_active ? 'Active' : 'Inactive'}
                            </span>
                          </div>
                        </div>
                        
                        <div className="flex items-center gap-1 ml-4">
                          <button
                            onClick={() => handlePreviewTemplate(template)}
                            className="p-1 text-gray-400 hover:text-blue-600"
                            title="Preview template"
                          >
                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                            </svg>
                          </button>
                          <button
                            onClick={() => {
                              setEditingTemplate(template);
                              setShowForm(true);
                            }}
                            className="p-1 text-gray-400 hover:text-gray-600"
                            title="Edit template"
                          >
                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                            </svg>
                          </button>
                          <button
                            onClick={() => handleDeleteTemplate(template.id)}
                            className="p-1 text-gray-400 hover:text-red-600"
                            title="Delete template"
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
          ) : showForm ? (
            <>
              {/* Template Form */}
              <div className="mb-6">
                <div className="flex items-center justify-between mb-4">
                  <h4 className="text-md font-medium text-gray-900">
                    {editingTemplate ? 'Edit Template' : 'Create New Template'}
                  </h4>
                  <button
                    onClick={() => setShowForm(false)}
                    className="text-sm text-gray-500 hover:text-gray-700"
                  >
                    ← Back to Templates
                  </button>
                </div>
                
                <form onSubmit={(e) => {
                  e.preventDefault();
                  const formData = new FormData(e.target);
                  
                  // Parse variables from form
                  const variables = [];
                  const variableNames = formData.getAll('variable_name[]');
                  const variablePlaceholders = formData.getAll('variable_placeholder[]');
                  const variableDefaults = formData.getAll('variable_default[]');
                  const variableRequired = formData.getAll('variable_required[]');
                  
                  for (let i = 0; i < variableNames.length; i++) {
                    if (variableNames[i] && variablePlaceholders[i]) {
                      variables.push({
                        name: variableNames[i],
                        placeholder: variablePlaceholders[i],
                        default_value: variableDefaults[i] || '',
                        required: variableRequired.includes(i.toString())
                      });
                    }
                  }
                  
                  handleSaveTemplate({
                    name: formData.get('name'),
                    content: formData.get('content'),
                    type: formData.get('type'),
                    category: formData.get('category'),
                    media_url: formData.get('media_url'),
                    media_type: formData.get('media_type'),
                    variables: variables,
                    is_active: formData.get('is_active') === 'true'
                  });
                }}>
                  <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                    {/* Left Column */}
                    <div className="space-y-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Template Name *
                        </label>
                        <input
                          type="text"
                          name="name"
                          defaultValue={editingTemplate?.name}
                          required
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                          placeholder="e.g., Welcome Message, Follow-up"
                        />
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Type *
                        </label>
                        <select
                          name="type"
                          defaultValue={editingTemplate?.type || 'text'}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                        >
                          {templateTypes.map(type => (
                            <option key={type.value} value={type.value}>{type.label}</option>
                          ))}
                        </select>
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Category
                        </label>
                        <input
                          type="text"
                          name="category"
                          defaultValue={editingTemplate?.category}
                          list="categories"
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                          placeholder="e.g., Marketing, Support, Sales"
                        />
                        <datalist id="categories">
                          {categories.map(category => (
                            <option key={category} value={category} />
                          ))}
                        </datalist>
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Media URL (for image/document templates)
                        </label>
                        <input
                          type="url"
                          name="media_url"
                          defaultValue={editingTemplate?.media_url}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                          placeholder="https://example.com/image.jpg"
                        />
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Status
                        </label>
                        <select
                          name="is_active"
                          defaultValue={editingTemplate?.is_active !== false ? 'true' : 'false'}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                        >
                          <option value="true">Active</option>
                          <option value="false">Inactive</option>
                        </select>
                      </div>
                    </div>
                    
                    {/* Right Column */}
                    <div className="space-y-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Message Content *
                        </label>
                        <textarea
                          name="content"
                          defaultValue={editingTemplate?.content}
                          required
                          rows={8}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                          placeholder="Hello {{name}}, welcome to our service! We're excited to have you at {{company}}."
                        />
                        
                        <div className="mt-2 text-sm text-gray-600">
                          <p className="font-medium mb-1">Available Variables:</p>
                          <div className="flex flex-wrap gap-2">
                            {commonVariables.map(variable => (
                              <button
                                key={variable.name}
                                type="button"
                                onClick={(e) => {
                                  const textarea = e.target.closest('.space-y-4').querySelector('textarea[name="content"]');
                                  const cursorPos = textarea.selectionStart;
                                  const textBefore = textarea.value.substring(0, cursorPos);
                                  const textAfter = textarea.value.substring(cursorPos);
                                  textarea.value = textBefore + variable.placeholder + textAfter;
                                  textarea.focus();
                                  textarea.setSelectionRange(cursorPos + variable.placeholder.length, cursorPos + variable.placeholder.length);
                                }}
                                className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-700 hover:bg-gray-200"
                                title={variable.description}
                              >
                                {variable.placeholder}
                              </button>
                            ))}
                          </div>
                        </div>
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
                      {editingTemplate ? 'Update' : 'Create'} Template
                    </button>
                  </div>
                </form>
              </div>
            </>
          ) : previewTemplate ? (
            <>
              {/* Template Preview */}
              <div className="mb-6">
                <div className="flex items-center justify-between mb-4">
                  <h4 className="text-md font-medium text-gray-900">Template Preview</h4>
                  <button
                    onClick={() => setPreviewTemplate(null)}
                    className="text-sm text-gray-500 hover:text-gray-700"
                  >
                    ← Back to Templates
                  </button>
                </div>
                
                <div className="bg-gray-50 rounded-lg p-6">
                  <div className="mb-4">
                    <h5 className="font-medium text-gray-900 mb-2">{previewTemplate.name}</h5>
                    <div className="flex items-center gap-2 mb-4">
                      <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                        previewTemplate.type === 'text' ? 'bg-blue-100 text-blue-800' :
                        previewTemplate.type === 'image' ? 'bg-green-100 text-green-800' :
                        'bg-purple-100 text-purple-800'
                      }`}>
                        {previewTemplate.type}
                      </span>
                      {previewTemplate.category && (
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-700">
                          {previewTemplate.category}
                        </span>
                      )}
                    </div>
                  </div>
                  
                  <div className="bg-white rounded-lg border p-4">
                    <h6 className="text-sm font-medium text-gray-700 mb-2">Preview with sample data:</h6>
                    <div className="whitespace-pre-wrap text-gray-900">
                      {previewTemplate.preview}
                    </div>
                  </div>
                  
                  <div className="mt-4 text-sm text-gray-600">
                    <p><strong>Original template:</strong></p>
                    <div className="bg-gray-100 rounded p-2 mt-1 whitespace-pre-wrap">
                      {previewTemplate.content}
                    </div>
                  </div>
                </div>
              </div>
            </>
          ) : null}
        </div>
      </div>
    </div>
  );
};

export default MessageTemplatesModal;