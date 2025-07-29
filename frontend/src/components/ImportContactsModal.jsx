import React, { useState, useRef } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { useNotification } from '../contexts/NotificationContext';

const ImportContactsModal = ({ isOpen, onClose, onImportComplete, groups }) => {
  const { token } = useAuth();
  const { showNotification } = useNotification();
  const fileInputRef = useRef(null);
  
  const [importing, setImporting] = useState(false);
  const [importType, setImportType] = useState('csv'); // 'csv' or 'text'
  const [textData, setTextData] = useState('');
  const [selectedGroupId, setSelectedGroupId] = useState('');
  const [detectedContacts, setDetectedContacts] = useState([]);
  const [showPreview, setShowPreview] = useState(false);
  const [importResults, setImportResults] = useState(null);
  const [isDragActive, setIsDragActive] = useState(false);
  
  if (!isOpen) return null;
  
  const handleFileUpload = async (event) => {
    const file = event.target.files[0];
    if (!file) return;
    
    await processFile(file);
  };

  const processFile = async (file) => {
    console.log('Processing file:', {
      name: file.name,
      size: file.size,
      type: file.type
    });

    // Validate file type
    if (!file.name.toLowerCase().endsWith('.csv')) {
      showNotification('Please select a CSV file', 'error');
      return;
    }

    // Validate file size (10MB limit)
    if (file.size > 10 * 1024 * 1024) {
      showNotification('File size must be less than 10MB', 'error');
      return;
    }
    
    try {
      setImporting(true);
      const formData = new FormData();
      formData.append('file', file);
      formData.append('type', 'csv');
      
      console.log('Making API request to /api/contacts/detect');
      console.log('FormData contents:', {
        file: file.name,
        type: 'csv'
      });
      
      const response = await fetch('/api/contacts/detect', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`
        },
        body: formData
      });
      
      console.log('Response received:', {
        status: response.status,
        statusText: response.statusText,
        contentType: response.headers.get('content-type')
      });
      
      if (!response.ok) {
        const errorText = await response.text();
        console.error('API Error Response:', errorText);
        showNotification(`API Error: ${response.status} ${response.statusText}`, 'error');
        return;
      }
      
      const data = await response.json();
      console.log('API Response data:', data);
      setDetectedContacts(data.contacts || []);
      setShowPreview(true);
    } catch (error) {
      showNotification('Failed to analyze file', 'error');
      console.error('Error analyzing file:', error);
    } finally {
      setImporting(false);
    }
  };

  // Drag and drop handlers
  const dropRef = useRef(null);
  
  const handleDragEnter = (e) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragActive(true);
  };

  const handleDragLeave = (e) => {
    e.preventDefault();
    e.stopPropagation();
    
    // Only deactivate if leaving the drop area completely
    const rect = dropRef.current?.getBoundingClientRect();
    if (rect && (e.clientX < rect.left || e.clientX > rect.right || 
                 e.clientY < rect.top || e.clientY > rect.bottom)) {
      setIsDragActive(false);
    }
  };

  const handleDragOver = (e) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleDrop = (e) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragActive(false);

    const files = e.dataTransfer.files;
    if (files && files.length > 0) {
      processFile(files[0]);
    }
  };
  
  const handleTextAnalysis = async () => {
    if (!textData.trim()) {
      showNotification('Please enter contact data', 'error');
      return;
    }
    
    try {
      setImporting(true);
      const response = await fetch('/api/contacts/detect', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          type: 'text',
          data: textData
        })
      });
      
      if (!response.ok) throw new Error('Failed to analyze text');
      
      const data = await response.json();
      setDetectedContacts(data.contacts || []);
      setShowPreview(true);
    } catch (error) {
      showNotification('Failed to analyze text', 'error');
      console.error('Error analyzing text:', error);
    } finally {
      setImporting(false);
    }
  };
  
  const handleImport = async () => {
    if (detectedContacts.length === 0) {
      showNotification('No contacts to import', 'error');
      return;
    }
    
    try {
      setImporting(true);
      
      const contactsData = detectedContacts.map(contact => ({
        name: contact.name,
        phone: contact.phone,
        email: contact.email || '',
        company: contact.company || '',
        position: contact.position || '',
        group_id: selectedGroupId || null,
        notes: `Imported from ${importType}${contact.source ? ` - ${contact.source}` : ''}`,
        is_active: true
      }));
      
      const response = await fetch('/api/contacts/import', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          contacts: contactsData
        })
      });
      
      if (!response.ok) throw new Error('Failed to import contacts');
      
      const results = await response.json();
      setImportResults(results);
      
      if (results.success > 0) {
        showNotification(`Successfully imported ${results.success} contacts`, 'success');
        onImportComplete();
      }
    } catch (error) {
      showNotification('Failed to import contacts', 'error');
      console.error('Error importing contacts:', error);
    } finally {
      setImporting(false);
    }
  };
  
  const resetModal = () => {
    setImportType('csv');
    setTextData('');
    setSelectedGroupId('');
    setDetectedContacts([]);
    setShowPreview(false);
    setImportResults(null);
    setIsDragActive(false);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };
  
  const handleClose = () => {
    resetModal();
    onClose();
  };
  
  return (
    <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto">
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-medium text-gray-900">Import Contacts</h3>
            <button
              onClick={handleClose}
              className="text-gray-400 hover:text-gray-600"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
        
        <div className="p-6">
          {!showPreview && !importResults ? (
            <>
              {/* Import Type Selection */}
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-3">Import Method</label>
                <div className="flex gap-4">
                  <label className="flex items-center">
                    <input
                      type="radio"
                      name="importType"
                      value="csv"
                      checked={importType === 'csv'}
                      onChange={(e) => setImportType(e.target.value)}
                      className="mr-2"
                    />
                    CSV File Upload
                  </label>
                  <label className="flex items-center">
                    <input
                      type="radio"
                      name="importType"
                      value="text"
                      checked={importType === 'text'}
                      onChange={(e) => setImportType(e.target.value)}
                      className="mr-2"
                    />
                    Text Input
                  </label>
                </div>
              </div>
              
              {/* CSV Upload */}
              {importType === 'csv' && (
                <div className="mb-6">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Upload CSV File
                  </label>
                  <div 
                    ref={dropRef}
                    className={`border-2 border-dashed rounded-lg p-8 text-center transition-all duration-200 cursor-pointer ${
                      isDragActive 
                        ? 'border-primary-500 bg-primary-50 scale-105' 
                        : importing 
                          ? 'border-gray-200 bg-gray-50' 
                          : 'border-gray-300 hover:border-primary-400 hover:bg-gray-50'
                    }`}
                    onClick={() => !importing && fileInputRef.current?.click()}
                    onDragEnter={handleDragEnter}
                    onDragLeave={handleDragLeave}
                    onDragOver={handleDragOver}
                    onDrop={handleDrop}
                  >
                    {importing ? (
                      <div className="flex flex-col items-center">
                        <svg className="animate-spin h-12 w-12 text-primary-600 mb-2" fill="none" viewBox="0 0 24 24">
                          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        <span className="text-sm font-medium text-gray-900">Processing file...</span>
                        <p className="text-xs text-gray-500 mt-1">Please wait while we analyze your CSV</p>
                      </div>
                    ) : (
                      <>
                        <div className="flex flex-col items-center">
                          <svg 
                            className={`h-16 w-16 mb-4 transition-colors ${isDragActive ? 'text-primary-500' : 'text-gray-400'}`} 
                            fill="none" 
                            stroke="currentColor" 
                            viewBox="0 0 24 24"
                          >
                            <path 
                              strokeLinecap="round" 
                              strokeLinejoin="round" 
                              strokeWidth={1.5} 
                              d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" 
                            />
                          </svg>
                          
                          <div className="space-y-2">
                            <p className={`text-lg font-medium ${isDragActive ? 'text-primary-600' : 'text-gray-900'}`}>
                              {isDragActive ? 'Drop your CSV file here!' : 'Drag and drop your CSV file'}
                            </p>
                            <p className="text-sm text-gray-500">
                              or click to browse • CSV files up to 10MB
                            </p>
                          </div>
                        </div>
                        <input
                          ref={fileInputRef}
                          id="file-upload"
                          name="file-upload"
                          type="file"
                          accept=".csv"
                          className="hidden"
                          onChange={handleFileUpload}
                          disabled={importing}
                          onClick={(e) => e.target.value = ''}
                        />
                      </>
                    )}
                  </div>
                  
                  <div className="mt-4 text-sm text-gray-600">
                    <p className="font-medium">CSV Format Tips:</p>
                    <ul className="mt-1 list-disc list-inside text-xs space-y-1">
                      <li>Include headers like: name, phone, email, company, position</li>
                      <li>Phone numbers can be in any format (will be auto-detected)</li>
                      <li>The system will automatically detect and map columns</li>
                      <li>Supports multiple languages and formats</li>
                    </ul>
                  </div>
                </div>
              )}
              
              {/* Text Input */}
              {importType === 'text' && (
                <div className="mb-6">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Paste Contact Data
                  </label>
                  <textarea
                    value={textData}
                    onChange={(e) => setTextData(e.target.value)}
                    rows={10}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                    placeholder="Enter contact information, one per line:&#10;John Doe +1234567890 john@example.com&#10;Jane Smith +0987654321 XYZ Corp&#10;+1555000111 Bob Johnson&#10;..."
                  />
                  
                  <div className="mt-4 text-sm text-gray-600">
                    <p className="font-medium">Text Format Tips:</p>
                    <ul className="mt-1 list-disc list-inside text-xs space-y-1">
                      <li>One contact per line</li>
                      <li>Can include: name, phone, email, company in any order</li>
                      <li>The system will automatically detect and parse fields</li>
                      <li>Phone numbers will be validated and formatted</li>
                    </ul>
                  </div>
                </div>
              )}
              
              {/* Group Selection */}
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Assign to Group (Optional)
                </label>
                <select
                  value={selectedGroupId}
                  onChange={(e) => setSelectedGroupId(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                >
                  <option value="">No Group</option>
                  {groups.map(group => (
                    <option key={group.id} value={group.id}>{group.name}</option>
                  ))}
                </select>
              </div>

              {/* Action Buttons for Initial Step */}
              <div className="flex justify-end gap-3">
                <button
                  onClick={handleClose}
                  className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
                >
                  Cancel
                </button>
                {importType === 'csv' && (
                  <button
                    onClick={() => fileInputRef.current?.click()}
                    disabled={importing}
                    className="px-4 py-2 bg-primary-600 text-white rounded-md text-sm font-medium hover:bg-primary-700 disabled:opacity-50 flex items-center gap-2"
                  >
                    {importing ? (
                      <>
                        <svg className="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        Processing...
                      </>
                    ) : (
                      <>
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                        </svg>
                        Choose File
                      </>
                    )}
                  </button>
                )}
                {importType === 'text' && (
                  <button
                    onClick={handleTextAnalysis}
                    disabled={importing || !textData.trim()}
                    className="px-4 py-2 bg-primary-600 text-white rounded-md text-sm font-medium hover:bg-primary-700 disabled:opacity-50 flex items-center gap-2"
                  >
                    {importing ? (
                      <>
                        <svg className="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        Analyzing...
                      </>
                    ) : (
                      <>
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                        </svg>
                        Analyze Text
                      </>
                    )}
                  </button>
                )}
              </div>
            </>
          ) : showPreview && !importResults ? (
            <>
              {/* Preview */}
              <div className="mb-6">
                <div className="flex items-center justify-between mb-4">
                  <h4 className="text-lg font-medium text-gray-900">
                    Preview Detected Contacts ({detectedContacts.length})
                  </h4>
                  <button
                    onClick={() => setShowPreview(false)}
                    className="text-sm text-gray-500 hover:text-gray-700"
                  >
                    ← Back to Import
                  </button>
                </div>
                
                {detectedContacts.length === 0 ? (
                  <div className="text-center py-8 text-gray-500">
                    <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.081 16.5c-.77.833.192 2.5 1.732 2.5z" />
                    </svg>
                    <p className="mt-2">No valid contacts detected</p>
                    <p className="text-sm">Please check your data format and try again</p>
                  </div>
                ) : (
                  <div className="overflow-x-auto">
                    <table className="w-full border border-gray-200 rounded-lg">
                      <thead className="bg-gray-50">
                        <tr>
                          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Phone</th>
                          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Email</th>
                          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Company</th>
                          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Confidence</th>
                          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Source</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-gray-200">
                        {detectedContacts.map((contact, index) => (
                          <tr key={index} className="hover:bg-gray-50">
                            <td className="px-4 py-2 text-sm text-gray-900">{contact.name || '-'}</td>
                            <td className="px-4 py-2 text-sm text-gray-900 font-mono">{contact.phone || '-'}</td>
                            <td className="px-4 py-2 text-sm text-gray-900">{contact.email || '-'}</td>
                            <td className="px-4 py-2 text-sm text-gray-900">{contact.company || '-'}</td>
                            <td className="px-4 py-2">
                              <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                                contact.confidence >= 0.8 ? 'bg-green-100 text-green-800' :
                                contact.confidence >= 0.6 ? 'bg-yellow-100 text-yellow-800' :
                                'bg-red-100 text-red-800'
                              }`}>
                                {Math.round(contact.confidence * 100)}%
                              </span>
                            </td>
                            <td className="px-4 py-2 text-xs text-gray-500">{contact.source}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
              
              {detectedContacts.length > 0 && (
                <div className="flex justify-end gap-3">
                  <button
                    onClick={() => setShowPreview(false)}
                    className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
                  >
                    Back
                  </button>
                  <button
                    onClick={handleImport}
                    disabled={importing}
                    className="px-4 py-2 bg-primary-600 text-white rounded-md text-sm font-medium hover:bg-primary-700 disabled:opacity-50 flex items-center gap-2"
                  >
                    {importing ? (
                      <>
                        <svg className="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        Importing...
                      </>
                    ) : (
                      <>
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                        </svg>
                        Import {detectedContacts.length} Contacts
                      </>
                    )}
                  </button>
                </div>
              )}
            </>
          ) : importResults ? (
            <>
              {/* Import Results */}
              <div className="text-center py-8">
                <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100 mb-4">
                  <svg className="h-6 w-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                </div>
                
                <h3 className="text-lg font-medium text-gray-900 mb-2">Import Complete!</h3>
                
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
                  <div className="bg-green-50 rounded-lg p-4">
                    <div className="text-2xl font-bold text-green-600">{importResults.success}</div>
                    <div className="text-sm text-green-800">Imported</div>
                  </div>
                  <div className="bg-yellow-50 rounded-lg p-4">
                    <div className="text-2xl font-bold text-yellow-600">{importResults.duplicates}</div>
                    <div className="text-sm text-yellow-800">Duplicates</div>
                  </div>
                  <div className="bg-red-50 rounded-lg p-4">
                    <div className="text-2xl font-bold text-red-600">{importResults.failed}</div>
                    <div className="text-sm text-red-800">Failed</div>
                  </div>
                  <div className="bg-blue-50 rounded-lg p-4">
                    <div className="text-2xl font-bold text-blue-600">{importResults.total}</div>
                    <div className="text-sm text-blue-800">Total</div>
                  </div>
                </div>
                
                {importResults.errors && importResults.errors.length > 0 && (
                  <div className="mb-6">
                    <h4 className="text-sm font-medium text-gray-900 mb-2">Errors:</h4>
                    <div className="bg-red-50 rounded-lg p-4 max-h-32 overflow-y-auto">
                      {importResults.errors.map((error, index) => (
                        <div key={index} className="text-sm text-red-800">{error}</div>
                      ))}
                    </div>
                  </div>
                )}
                
                <button
                  onClick={handleClose}
                  className="px-6 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700"
                >
                  Close
                </button>
              </div>
            </>
          ) : null}
        </div>
      </div>
    </div>
  );
};

export default ImportContactsModal;