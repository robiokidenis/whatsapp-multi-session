import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { useNotification } from '../contexts/NotificationContext';

const ContactGroupsModal = ({ isOpen, onClose, onUpdate }) => {
  const { token } = useAuth();
  const { showNotification } = useNotification();
  
  const [groups, setGroups] = useState([]);
  const [loading, setLoading] = useState(true);
  const [editingGroup, setEditingGroup] = useState(null);
  const [showForm, setShowForm] = useState(false);
  
  const colors = [
    '#3B82F6', // Blue
    '#10B981', // Green
    '#F59E0B', // Yellow
    '#EF4444', // Red
    '#8B5CF6', // Purple
    '#EC4899', // Pink
    '#06B6D4', // Cyan
    '#84CC16', // Lime
    '#F97316', // Orange
    '#6B7280'  // Gray
  ];
  
  useEffect(() => {
    if (isOpen) {
      fetchGroups();
    }
  }, [isOpen]);
  
  const fetchGroups = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/contact-groups', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to fetch groups');
      
      const data = await response.json();
      setGroups(data || []);
    } catch (error) {
      showNotification('Failed to load contact groups', 'error');
      console.error('Error fetching groups:', error);
    } finally {
      setLoading(false);
    }
  };
  
  const handleSaveGroup = async (groupData) => {
    try {
      const url = editingGroup 
        ? `/api/contact-groups/${editingGroup.id}`
        : '/api/contact-groups';
      
      const response = await fetch(url, {
        method: editingGroup ? 'PUT' : 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(groupData)
      });
      
      if (!response.ok) throw new Error('Failed to save group');
      
      showNotification(`Group ${editingGroup ? 'updated' : 'created'} successfully`, 'success');
      setEditingGroup(null);
      setShowForm(false);
      fetchGroups();
      onUpdate();
    } catch (error) {
      showNotification('Failed to save group', 'error');
      console.error('Error saving group:', error);
    }
  };
  
  const handleDeleteGroup = async (groupId) => {
    if (!window.confirm('Delete this group? Contacts in this group will not be deleted.')) {
      return;
    }
    
    try {
      const response = await fetch(`/api/contact-groups/${groupId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to delete group');
      
      showNotification('Group deleted successfully', 'success');
      fetchGroups();
      onUpdate();
    } catch (error) {
      showNotification('Failed to delete group', 'error');
      console.error('Error deleting group:', error);
    }
  };
  
  const handleToggleActive = async (group) => {
    try {
      const response = await fetch(`/api/contact-groups/${group.id}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          is_active: !group.is_active
        })
      });
      
      if (!response.ok) throw new Error('Failed to update group');
      
      showNotification(`Group ${!group.is_active ? 'activated' : 'deactivated'}`, 'success');
      fetchGroups();
      onUpdate();
    } catch (error) {
      showNotification('Failed to update group', 'error');
      console.error('Error updating group:', error);
    }
  };
  
  if (!isOpen) return null;
  
  return (
    <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto">
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-medium text-gray-900">Manage Contact Groups</h3>
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
          {!showForm ? (
            <>
              {/* Groups List */}
              <div className="flex justify-between items-center mb-6">
                <h4 className="text-md font-medium text-gray-900">Contact Groups ({groups.length})</h4>
                <button
                  onClick={() => {
                    setEditingGroup(null);
                    setShowForm(true);
                  }}
                  className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 flex items-center gap-2"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                  </svg>
                  Add Group
                </button>
              </div>
              
              {loading ? (
                <div className="flex justify-center items-center py-8">
                  <svg className="animate-spin h-8 w-8 text-primary-600" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                </div>
              ) : groups.length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                  <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                  </svg>
                  <p className="mt-2">No contact groups yet</p>
                  <p className="text-sm">Create your first group to organize contacts</p>
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {groups.map(group => (
                    <div key={group.id} className="bg-white border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow">
                      <div className="flex items-start justify-between mb-3">
                        <div className="flex items-center gap-3">
                          <div 
                            className="w-4 h-4 rounded-full flex-shrink-0"
                            style={{ backgroundColor: group.color || '#6B7280' }}
                          ></div>
                          <div>
                            <h5 className="font-medium text-gray-900">{group.name}</h5>
                            {group.description && (
                              <p className="text-sm text-gray-500 mt-1">{group.description}</p>
                            )}
                          </div>
                        </div>
                        
                        <div className="flex items-center gap-1">
                          <button
                            onClick={() => {
                              setEditingGroup(group);
                              setShowForm(true);
                            }}
                            className="p-1 text-gray-400 hover:text-gray-600"
                            title="Edit group"
                          >
                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                            </svg>
                          </button>
                          <button
                            onClick={() => handleDeleteGroup(group.id)}
                            className="p-1 text-gray-400 hover:text-red-600"
                            title="Delete group"
                          >
                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                            </svg>
                          </button>
                        </div>
                      </div>
                      
                      <div className="flex items-center justify-between">
                        <div className="text-sm text-gray-500">
                          {group.contact_count || 0} contacts
                        </div>
                        
                        <button
                          onClick={() => handleToggleActive(group)}
                          className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            group.is_active
                              ? 'bg-green-100 text-green-800 hover:bg-green-200'
                              : 'bg-gray-100 text-gray-800 hover:bg-gray-200'
                          }`}
                        >
                          {group.is_active ? 'Active' : 'Inactive'}
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </>
          ) : (
            <>
              {/* Group Form */}
              <div className="mb-6">
                <div className="flex items-center justify-between mb-4">
                  <h4 className="text-md font-medium text-gray-900">
                    {editingGroup ? 'Edit Group' : 'Create New Group'}
                  </h4>
                  <button
                    onClick={() => setShowForm(false)}
                    className="text-sm text-gray-500 hover:text-gray-700"
                  >
                    ← Back to Groups
                  </button>
                </div>
                
                <form onSubmit={(e) => {
                  e.preventDefault();
                  const formData = new FormData(e.target);
                  handleSaveGroup({
                    name: formData.get('name'),
                    description: formData.get('description'),
                    color: formData.get('color'),
                    is_active: formData.get('is_active') === 'true'
                  });
                }}>
                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Group Name *
                      </label>
                      <input
                        type="text"
                        name="name"
                        defaultValue={editingGroup?.name}
                        required
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                        placeholder="e.g., Customers, Leads, VIP"
                      />
                    </div>
                    
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Description
                      </label>
                      <textarea
                        name="description"
                        defaultValue={editingGroup?.description}
                        rows={3}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                        placeholder="Optional description for this group"
                      />
                    </div>
                    
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Color
                      </label>
                      <div className="flex flex-wrap gap-2">
                        {colors.map(color => (
                          <label key={color} className="cursor-pointer">
                            <input
                              type="radio"
                              name="color"
                              value={color}
                              defaultChecked={editingGroup?.color === color || (!editingGroup && color === colors[0])}
                              className="sr-only"
                            />
                            <div 
                              className="w-8 h-8 rounded-full border-2 border-transparent hover:border-gray-300 focus:border-primary-500"
                              style={{ backgroundColor: color }}
                              title={color}
                            ></div>
                          </label>
                        ))}
                      </div>
                    </div>
                    
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Status
                      </label>
                      <select
                        name="is_active"
                        defaultValue={editingGroup?.is_active !== false ? 'true' : 'false'}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                      >
                        <option value="true">Active</option>
                        <option value="false">Inactive</option>
                      </select>
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
                      {editingGroup ? 'Update' : 'Create'} Group
                    </button>
                  </div>
                </form>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default ContactGroupsModal;