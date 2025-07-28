import { useState, useEffect, useMemo } from 'react';
import axios from 'axios';
import { useNotification } from '../contexts/NotificationContext';

const UserManagement = () => {
  const { showSuccess, showError, showWarning } = useNotification();
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [roleFilter, setRoleFilter] = useState('all');
  const [statusFilter, setStatusFilter] = useState('all');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [selectedUser, setSelectedUser] = useState(null);
  const [selectedUsers, setSelectedUsers] = useState([]);
  const [showFilters, setShowFilters] = useState(false);
  const [bulkActionLoading, setBulkActionLoading] = useState(false);
  const [formData, setFormData] = useState({
    username: '',
    password: '',
    role: 'user',
    session_limit: 5,
  });

  useEffect(() => {
    loadUsers();
  }, []);

  const stats = useMemo(() => {
    const total = users.length;
    const active = users.filter(u => u.is_active).length;
    const inactive = total - active;
    const admins = users.filter(u => u.role === 'admin').length;
    return { total, active, inactive, admins };
  }, [users]);

  const filteredUsers = useMemo(() => {
    return users.filter(user => {
      const matchesSearch = !searchTerm || 
        user.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
        user.role.toLowerCase().includes(searchTerm.toLowerCase()) ||
        user.id?.toString().includes(searchTerm.toLowerCase());
      
      const matchesRole = roleFilter === 'all' || user.role === roleFilter;
      const matchesStatus = statusFilter === 'all' || 
        (statusFilter === 'active' && user.is_active) ||
        (statusFilter === 'inactive' && !user.is_active);
      
      return matchesSearch && matchesRole && matchesStatus;
    });
  }, [users, searchTerm, roleFilter, statusFilter]);

  const loadUsers = async () => {
    try {
      setLoading(true);
      const response = await axios.get('/api/admin/users');
      
      if (response.data.success) {
        setUsers(response.data.data || []);
      } else {
        showError('Failed to load users');
      }
    } catch (error) {
      showError('Error loading users');
    } finally {
      setLoading(false);
    }
  };


  const handleCreateUser = async (e) => {
    e.preventDefault();
    
    if (!formData.username || !formData.password) {
      showError('Username and password are required');
      return;
    }

    try {
      setLoading(true);
      const response = await axios.post('/api/admin/users', formData);

      if (response.data.success) {
        showSuccess('User created successfully');
        setShowCreateModal(false);
        setFormData({ username: '', password: '', role: 'user', session_limit: 5 });
        await loadUsers();
      }
    } catch (error) {
      showError('Error creating user');
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateUser = async (e) => {
    e.preventDefault();
    
    if (!selectedUser) return;

    const updateData = {};
    if (formData.username && formData.username !== selectedUser.username) {
      updateData.username = formData.username;
    }
    if (formData.password) {
      updateData.password = formData.password;
    }
    if (formData.role && formData.role !== selectedUser.role) {
      updateData.role = formData.role;
    }
    if (formData.session_limit !== undefined && formData.session_limit !== selectedUser.session_limit) {
      updateData.session_limit = formData.session_limit;
    }

    if (Object.keys(updateData).length === 0) {
      showWarning('No changes to update');
      return;
    }

    try {
      setLoading(true);
      const response = await axios.put(`/api/admin/users/${selectedUser.id}`, updateData);

      if (response.data.success) {
        showSuccess('User updated successfully');
        setShowEditModal(false);
        setSelectedUser(null);
        setFormData({ username: '', password: '', role: 'user', session_limit: 5 });
        await loadUsers();
      }
    } catch (error) {
      showError('Error updating user');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteUser = async (user) => {
    if (!window.confirm(`Delete user "${user.username}"?`)) return;

    try {
      setLoading(true);
      const response = await axios.delete(`/api/admin/users/${user.id}`);

      if (response.data.success) {
        showSuccess('User deleted successfully');
        await loadUsers();
      }
    } catch (error) {
      showError('Error deleting user');
    } finally {
      setLoading(false);
    }
  };

  const handleToggleUserStatus = async (user) => {
    try {
      setLoading(true);
      const response = await axios.put(`/api/admin/users/${user.id}`, { 
        is_active: !user.is_active 
      });

      if (response.data.success) {
        showSuccess(`User ${user.is_active ? 'deactivated' : 'activated'}`);
        await loadUsers();
      }
    } catch (error) {
      showError('Error updating user status');
    } finally {
      setLoading(false);
    }
  };

  const handleSelectUser = (userId) => {
    setSelectedUsers(prev => 
      prev.includes(userId) 
        ? prev.filter(id => id !== userId)
        : [...prev, userId]
    );
  };

  const handleSelectAll = () => {
    setSelectedUsers(
      selectedUsers.length === filteredUsers.length 
        ? [] 
        : filteredUsers.map(u => u.id)
    );
  };

  const handleBulkDelete = async () => {
    const nonAdminUsers = selectedUsers.filter(id => {
      const user = users.find(u => u.id === id);
      return user && user.username !== 'admin';
    });
    
    if (nonAdminUsers.length === 0) {
      showError('Cannot delete admin users');
      return;
    }
    
    if (!window.confirm(`Delete ${nonAdminUsers.length} selected users?`)) return;
    
    setBulkActionLoading(true);
    try {
      await Promise.all(
        nonAdminUsers.map(id => axios.delete(`/api/admin/users/${id}`).catch(() => null))
      );
      showSuccess(`Deleted ${nonAdminUsers.length} users`);
      setSelectedUsers([]);
      await loadUsers();
    } catch (error) {
      showError('Bulk delete failed');
    } finally {
      setBulkActionLoading(false);
    }
  };

  const openCreateModal = () => {
    setFormData({ username: '', password: '', role: 'user', session_limit: 5 });
    setShowCreateModal(true);
  };

  const openEditModal = (user) => {
    setSelectedUser(user);
    setFormData({ 
      username: user.username, 
      password: '', 
      role: user.role || 'user',
      session_limit: user.session_limit || 5
    });
    setShowEditModal(true);
  };

  const closeModals = () => {
    setShowCreateModal(false);
    setShowEditModal(false);
    setSelectedUser(null);
    setFormData({ username: '', password: '', role: 'user', session_limit: 5 });
  };

  if (loading && users.length === 0) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="loading-spinner-lg mx-auto mb-4"></div>
          <p className="text-secondary">Loading users...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto px-6 py-8">
        
        {/* CRM Header */}
        <div className="mb-8">
          <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-6">
            <div>
              <div className="flex items-center gap-3 mb-2">
    
                <h1 className="text-display">User Management</h1>
              </div>
              <p className="text-secondary">Manage user accounts, roles, and access permissions</p>
            </div>
            <button 
              onClick={openCreateModal}
              className="btn btn-primary btn-lg"
            >
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
              </svg>
              Add New User
            </button>
          </div>
        </div>

        {/* Improved User Stats */}
        <div className="grid-cols-4 mb-8">
          <div className="card border border-gray-200">
            <div className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-overline mb-1 text-gray-600">Total Users</div>
                  <div className="text-title text-gray-900">{stats.total}</div>
                  <div className="text-caption mt-1 text-gray-500">All registered</div>
                </div>
                <div className="w-8 h-8 bg-gray-50 border border-gray-200 rounded-lg flex items-center justify-center">
                  <svg className="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                  </svg>
                </div>
              </div>
            </div>
          </div>
          
          <div className="card border border-success-200 bg-success-25">
            <div className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-overline mb-1 text-success-600">Active Users</div>
                  <div className="text-title text-success-800">{stats.active}</div>
                  <div className="text-caption mt-1 text-success-600">Currently enabled</div>
                </div>
                <div className="w-8 h-8 bg-success-50 border border-success-200 rounded-lg flex items-center justify-center">
                  <svg className="w-4 h-4 text-success-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
              </div>
            </div>
          </div>
          
          <div className="card border border-error-200 bg-error-25">
            <div className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-overline mb-1 text-error-600">Inactive Users</div>
                  <div className="text-title text-error-800">{stats.inactive}</div>
                  <div className="text-caption mt-1 text-error-600">Disabled accounts</div>
                </div>
                <div className="w-8 h-8 bg-error-50 border border-error-200 rounded-lg flex items-center justify-center">
                  <svg className="w-4 h-4 text-error-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728L5.636 5.636m12.728 12.728L18.364 5.636M5.636 18.364l12.728-12.728" />
                  </svg>
                </div>
              </div>
            </div>
          </div>
          
          <div className="card border border-primary-200 bg-primary-25">
            <div className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-overline mb-1 text-primary-600">Administrators</div>
                  <div className="text-title text-primary-800">{stats.admins}</div>
                  <div className="text-caption mt-1 text-primary-600">Admin privileges</div>
                </div>
                <div className="w-8 h-8 bg-primary-50 border border-primary-200 rounded-lg flex items-center justify-center">
                  <svg className="w-4 h-4 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                  </svg>
                </div>
              </div>
            </div>
          </div>
        </div>


        {/* Search */}
        <div className="card p-6 mb-8">
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="flex-1">
              <input
                type="text"
                placeholder="Search users..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="input"
              />
            </div>
            <button
              onClick={() => setShowFilters(!showFilters)}
              className={`btn ${showFilters ? 'btn-primary' : 'btn-secondary'}`}
            >
              Filters
            </button>
          </div>

          {/* Filters */}
          {showFilters && (
            <div className="animate-fade-in mt-6 pt-6 border-t border-gray-200">
              <div className="flex flex-col sm:flex-row gap-4">
                <select
                  value={roleFilter}
                  onChange={(e) => setRoleFilter(e.target.value)}
                  className="select"
                >
                  <option value="all">All Roles</option>
                  <option value="admin">Admin</option>
                  <option value="user">User</option>
                </select>

                <select
                  value={statusFilter}
                  onChange={(e) => setStatusFilter(e.target.value)}
                  className="select"
                >
                  <option value="all">All Status</option>
                  <option value="active">Active</option>
                  <option value="inactive">Inactive</option>
                </select>

                {filteredUsers.length > 0 && (
                  <div className="flex items-center gap-4">
                    <button
                      onClick={handleSelectAll}
                      className="btn btn-sm btn-secondary"
                    >
                      {selectedUsers.length === filteredUsers.length ? 'Deselect All' : 'Select All'}
                    </button>
                    
                    {selectedUsers.length > 0 && (
                      <button
                        onClick={handleBulkDelete}
                        disabled={bulkActionLoading}
                        className="btn btn-sm btn-danger"
                      >
                        {bulkActionLoading ? 'Deleting...' : `Delete (${selectedUsers.length})`}
                      </button>
                    )}
                  </div>
                )}
              </div>
            </div>
          )}
        </div>

        {/* Users Table */}
        <div className="mb-8">
          {filteredUsers.length === 0 ? (
            <div className="card p-12 text-center">
              <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-6">
                <svg className="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
                </svg>
              </div>
              <h3 className="text-title mb-2">
                {users.length === 0 ? 'No Users Yet' : 'No Matching Users'}
              </h3>
              <p className="text-secondary mb-6 max-w-md mx-auto">
                {users.length === 0 
                  ? 'Create your first user account to get started.'
                  : 'Try adjusting your search or filters.'
                }
              </p>
              {users.length === 0 && (
                <button
                  onClick={openCreateModal}
                  className="btn btn-primary"
                >
                  Create First User
                </button>
              )}
            </div>
          ) : (
            <div className="card">
              <table className="table">
                <thead>
                  <tr>
                    {showFilters && (
                      <th className="w-12">
                        <input
                          type="checkbox"
                          checked={selectedUsers.length === filteredUsers.length && filteredUsers.length > 0}
                          onChange={handleSelectAll}
                          className="w-4 h-4"
                        />
                      </th>
                    )}
                    <th>User</th>
                    <th>Role</th>
                    <th>Status</th>
                    <th>Session Limit</th>
                    <th>Created</th>
                    <th className="text-right">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredUsers.map((user) => {
                    const isSelected = selectedUsers.includes(user.id);
                    
                    return (
                      <tr key={user.id} className={isSelected ? 'bg-gray-50' : ''}>
                        {showFilters && (
                          <td>
                            <input
                              type="checkbox"
                              checked={isSelected}
                              onChange={() => handleSelectUser(user.id)}
                              className="w-4 h-4"
                            />
                          </td>
                        )}
                        <td>
                          <div>
                            <div className="text-body mb-1">
                              {user.username}
                              {user.username === 'admin' && (
                                <span className="badge badge-warning ml-2">Super Admin</span>
                              )}
                            </div>
                            <div className="text-caption">ID: {user.id}</div>
                          </div>
                        </td>
                        <td>
                          <div className={`badge ${user.role === 'admin' ? 'badge-warning' : 'badge'}`}>
                            {user.role === 'admin' ? 'Administrator' : 'User'}
                          </div>
                        </td>
                        <td>
                          <div className={`badge ${user.is_active ? 'badge-success' : 'badge-error'}`}>
                            <div className={`status-dot ${user.is_active ? 'status-dot-success' : 'status-dot-error'}`}></div>
                            {user.is_active ? 'Active' : 'Inactive'}
                          </div>
                        </td>
                        <td>
                          <span className="text-body">
                            {user.session_limit === -1 ? 'Unlimited' : user.session_limit}
                          </span>
                        </td>
                        <td>
                          <div>
                            <div className="text-body">{new Date(user.created_at).toLocaleDateString()}</div>
                            <div className="text-caption">{new Date(user.created_at).toLocaleTimeString()}</div>
                          </div>
                        </td>
                        <td>
                          <div className="flex justify-end gap-2">
                            <button
                              onClick={() => openEditModal(user)}
                              className="btn btn-xs btn-secondary"
                            >
                              Edit
                            </button>
                            {user.username !== 'admin' && (
                              <>
                                <button
                                  onClick={() => handleToggleUserStatus(user)}
                                  className={`btn btn-xs ${user.is_active ? 'btn-warning' : 'btn-success'}`}
                                >
                                  {user.is_active ? 'Deactivate' : 'Activate'}
                                </button>
                                <button
                                  onClick={() => handleDeleteUser(user)}
                                  className="btn btn-xs btn-danger"
                                >
                                  Delete
                                </button>
                              </>
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
      </div>

      {/* Create User Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="card max-w-md w-full animate-scale-in">
            <div className="p-6">
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-title">Create New User</h3>
                <button 
                  onClick={closeModals} 
                  className="btn btn-ghost p-2"
                >
                  ×
                </button>
              </div>
              
              <form onSubmit={handleCreateUser} className="space-y-6">
                <div>
                  <label className="block text-body mb-2">Username</label>
                  <input
                    type="text"
                    value={formData.username}
                    onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                    placeholder="Enter username"
                    className="input"
                    required
                    minLength={3}
                  />
                </div>

                <div>
                  <label className="block text-body mb-2">Password</label>
                  <input
                    type="password"
                    value={formData.password}
                    onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                    placeholder="Enter password"
                    className="input"
                    required
                    minLength={6}
                  />
                </div>

                <div>
                  <label className="block text-body mb-2">Role</label>
                  <select
                    value={formData.role}
                    onChange={(e) => setFormData({ ...formData, role: e.target.value })}
                    className="select"
                    required
                  >
                    <option value="user">User</option>
                    <option value="admin">Administrator</option>
                  </select>
                </div>

                <div>
                  <label className="block text-body mb-2">Session Limit</label>
                  <input
                    type="number"
                    value={formData.session_limit}
                    onChange={(e) => setFormData({ ...formData, session_limit: parseInt(e.target.value) || 5 })}
                    className="input"
                    min="-1"
                    required
                  />
                  <p className="text-caption mt-1">Use -1 for unlimited sessions</p>
                </div>

                <div className="flex gap-3 pt-4">
                  <button
                    type="button"
                    onClick={closeModals}
                    className="btn btn-secondary flex-1"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={loading}
                    className="btn btn-primary flex-1"
                  >
                    {loading ? 'Creating...' : 'Create User'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Edit User Modal */}
      {showEditModal && selectedUser && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="card max-w-md w-full animate-scale-in">
            <div className="p-6">
              <div className="flex items-center justify-between mb-6">
                <div>
                  <h3 className="text-title">Edit User</h3>
                  <p className="text-secondary">@{selectedUser.username}</p>
                </div>
                <button 
                  onClick={closeModals} 
                  className="btn btn-ghost p-2"
                >
                  ×
                </button>
              </div>
              
              <form onSubmit={handleUpdateUser} className="space-y-6">
                <div>
                  <label className="block text-body mb-2">Username</label>
                  <input
                    type="text"
                    value={formData.username}
                    onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                    placeholder="Enter new username"
                    className="input"
                    minLength={3}
                  />
                </div>

                <div>
                  <label className="block text-body mb-2">New Password</label>
                  <input
                    type="password"
                    value={formData.password}
                    onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                    placeholder="Enter new password (optional)"
                    className="input"
                    minLength={6}
                  />
                  <p className="text-caption mt-1">Leave empty to keep current password</p>
                </div>

                <div>
                  <label className="block text-body mb-2">Role</label>
                  <select
                    value={formData.role}
                    onChange={(e) => setFormData({ ...formData, role: e.target.value })}
                    className="select"
                    disabled={selectedUser.username === 'admin'}
                  >
                    <option value="user">User</option>
                    <option value="admin">Administrator</option>
                  </select>
                  {selectedUser.username === 'admin' && (
                    <p className="text-caption mt-1">Admin user role cannot be changed</p>
                  )}
                </div>

                <div>
                  <label className="block text-body mb-2">Session Limit</label>
                  <input
                    type="number"
                    value={formData.session_limit}
                    onChange={(e) => setFormData({ ...formData, session_limit: parseInt(e.target.value) || 5 })}
                    className="input"
                    min="-1"
                  />
                  <p className="text-caption mt-1">Use -1 for unlimited sessions</p>
                </div>

                <div className="flex gap-3 pt-4">
                  <button
                    type="button"
                    onClick={closeModals}
                    className="btn btn-secondary flex-1"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={loading}
                    className="btn btn-primary flex-1"
                  >
                    {loading ? 'Updating...' : 'Update User'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default UserManagement;