import { useState, useEffect } from 'react';
import axios from 'axios';

const UserManagement = () => {
  const [users, setUsers] = useState([]);
  const [filteredUsers, setFilteredUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [roleFilter, setRoleFilter] = useState('all');
  const [message, setMessage] = useState('');
  const [messageType, setMessageType] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [selectedUser, setSelectedUser] = useState(null);
  const [formData, setFormData] = useState({
    username: '',
    password: '',
    role: 'user',
    session_limit: 5,
  });

  useEffect(() => {
    loadUsers();
  }, []);

  useEffect(() => {
    filterUsers();
  }, [users, searchTerm, roleFilter]);

  const filterUsers = () => {
    let filtered = users;

    if (searchTerm) {
      filtered = filtered.filter(user =>
        user.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
        user.role.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    if (roleFilter !== 'all') {
      filtered = filtered.filter(user => user.role === roleFilter);
    }

    setFilteredUsers(filtered);
  };

  const loadUsers = async () => {
    try {
      setLoading(true);
      const response = await axios.get('/api/admin/users');
      
      if (response.data.success) {
        const userData = response.data.data || [];
        setUsers(userData);
        setFilteredUsers(userData);
      } else {
        showMessage('Failed to load users: ' + (response.data.error || 'Unknown error'), 'error');
      }
    } catch (error) {
      showMessage('Error loading users: ' + (error.response?.data?.error || error.message), 'error');
    } finally {
      setLoading(false);
    }
  };

  const showMessage = (text, type) => {
    setMessage(text);
    setMessageType(type);
    setTimeout(() => {
      setMessage('');
      setMessageType('');
    }, 5000);
  };

  const handleCreateUser = async (e) => {
    e.preventDefault();
    
    if (!formData.username || !formData.password) {
      showMessage('Username and password are required', 'error');
      return;
    }

    try {
      setLoading(true);
      const response = await axios.post('/api/admin/users', {
        username: formData.username,
        password: formData.password,
        role: formData.role,
        session_limit: formData.session_limit,
      });

      if (response.data.success) {
        showMessage('User created successfully!', 'success');
        setShowCreateModal(false);
        setFormData({ username: '', password: '' });
        await loadUsers();
      }
    } catch (error) {
      console.error('Error creating user:', error);
      showMessage('Error creating user: ' + (error.response?.data?.error || error.message), 'error');
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
      showMessage('No changes to update', 'error');
      return;
    }

    try {
      setLoading(true);
      const response = await axios.put(`/api/admin/users/${selectedUser.id}`, updateData);

      if (response.data.success) {
        showMessage('User updated successfully!', 'success');
        setShowEditModal(false);
        setSelectedUser(null);
        setFormData({ username: '', password: '' });
        await loadUsers();
      }
    } catch (error) {
      console.error('Error updating user:', error);
      showMessage('Error updating user: ' + (error.response?.data?.error || error.message), 'error');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteUser = async (user) => {
    if (!window.confirm(`Are you sure you want to delete user "${user.username}"?`)) {
      return;
    }

    try {
      setLoading(true);
      const response = await axios.delete(`/api/admin/users/${user.id}`);

      if (response.data.success) {
        showMessage('User deleted successfully!', 'success');
        await loadUsers();
      }
    } catch (error) {
      console.error('Error deleting user:', error);
      showMessage('Error deleting user: ' + (error.response?.data?.error || error.message), 'error');
    } finally {
      setLoading(false);
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

  return (
    <div>
      {/* Header Section */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 flex items-center">
              <div className="bg-gray-100 p-3 rounded-lg mr-4">
                <i className="fas fa-users text-gray-600 text-xl"></i>
              </div>
              User Management
            </h1>
            <p className="text-gray-600 mt-2">Manage user accounts and permissions</p>
          </div>
          <button
            onClick={openCreateModal}
            className="bg-green-600 hover:bg-green-700 text-white px-6 py-3 rounded-lg font-medium transition-colors duration-200"
          >
            <i className="fas fa-plus mr-2"></i>Add User
          </button>
        </div>
      </div>

      {/* Search and Filter Section */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="flex-1">
            <div className="relative">
              <i className="fas fa-search absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400"></i>
              <input
                type="text"
                placeholder="Search users by username or role..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 transition-colors duration-200"
              />
            </div>
          </div>
          <div className="w-full sm:w-48">
            <select
              value={roleFilter}
              onChange={(e) => setRoleFilter(e.target.value)}
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 transition-colors duration-200"
            >
              <option value="all">All Roles</option>
              <option value="admin">Admin</option>
              <option value="user">User</option>
            </select>
          </div>
        </div>
        
        {/* Stats */}
        <div className="mt-4 flex flex-wrap gap-4">
          <div className="bg-gray-50 px-4 py-2 rounded-lg border border-gray-200">
            <span className="text-gray-700 font-medium">
              Total: {users.length}
            </span>
          </div>
          <div className="bg-gray-50 px-4 py-2 rounded-lg border border-gray-200">
            <span className="text-gray-700 font-medium">
              Filtered: {filteredUsers.length}
            </span>
          </div>
          <div className="bg-gray-50 px-4 py-2 rounded-lg border border-gray-200">
            <span className="text-gray-700 font-medium">
              Active: {users.filter(u => u.is_active).length}
            </span>
          </div>
        </div>
      </div>

      {/* Message Alert */}
      {message && (
        <div className={`mb-6 p-4 rounded-lg border ${
          messageType === 'error'
            ? 'bg-red-50 border-red-200 text-red-800'
            : 'bg-green-50 border-green-200 text-green-800'
        }`}>
          <div className="flex items-center">
            <i className={`fas ${messageType === 'error' ? 'fa-exclamation-circle' : 'fa-check-circle'} mr-3`}></i>
            <span className="font-medium">{message}</span>
          </div>
        </div>
      )}

      {/* Main Content Area */}
      {loading ? (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-12">
          <div className="text-center">
            <div className="inline-flex items-center justify-center w-16 h-16 bg-gray-100 rounded-full mb-4">
              <i className="fas fa-spinner fa-spin text-gray-600 text-2xl"></i>
            </div>
            <h3 className="text-xl font-semibold text-gray-900 mb-2">Loading Users</h3>
            <p className="text-gray-600">Please wait while we fetch the user data...</p>
          </div>
        </div>
      ) : filteredUsers.length > 0 ? (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    User
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Role
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Session Limit
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Created At
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {filteredUsers.map((user, index) => (
                    <tr 
                      key={user.id} 
                      className="hover:bg-gray-50 transition-colors duration-200"
                    >
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center">
                          <div className="flex-shrink-0 h-12 w-12">
                            <div className={`h-12 w-12 rounded-full ${
                              user.role === 'admin' 
                                ? 'bg-gray-700' 
                                : 'bg-gray-500'
                            } flex items-center justify-center`}>
                              <i className={`fas ${
                                user.role === 'admin' ? 'fa-crown' : 'fa-user'
                              } text-white text-lg`}></i>
                            </div>
                          </div>
                          <div className="ml-4">
                            <div className="text-base font-semibold text-gray-900">
                              {user.username}
                            </div>
                            <div className="text-sm text-gray-500">ID: {user.id}</div>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${
                          user.role === 'admin' 
                            ? 'bg-gray-100 text-gray-800 border border-gray-200' 
                            : 'bg-green-100 text-green-800 border border-green-200'
                        }`}>
                          <i className={`fas ${
                            user.role === 'admin' ? 'fa-shield-alt' : 'fa-user-circle'
                          } mr-2`}></i>
                          {user.role === 'admin' ? 'Administrator' : 'User'}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-gray-900">
                          {user.session_limit === -1 ? (
                            <span className="inline-flex items-center text-gray-700">
                              <i className="fas fa-infinity mr-2"></i>
                              Unlimited
                            </span>
                          ) : (
                            <span className="inline-flex items-center text-gray-700">
                              <i className="fas fa-list-ol mr-2"></i>
                              {user.session_limit}
                            </span>
                          )}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${
                          user.is_active 
                            ? 'bg-green-100 text-green-800 border border-green-200' 
                            : 'bg-red-100 text-red-800 border border-red-200'
                        }`}>
                          <div className={`w-2 h-2 rounded-full mr-2 ${
                            user.is_active ? 'bg-green-500' : 'bg-red-500'
                          }`}></div>
                          {user.is_active ? 'Active' : 'Inactive'}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm text-gray-900 font-medium">
                          {new Date(user.created_at).toLocaleDateString()}
                        </div>
                        <div className="text-xs text-gray-500">
                          {new Date(user.created_at).toLocaleTimeString()}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center space-x-3">
                          <button
                            onClick={() => openEditModal(user)}
                            className="inline-flex items-center px-3 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg transition-colors duration-200"
                            title="Edit User"
                          >
                            <i className="fas fa-edit text-sm"></i>
                          </button>
                          {user.username !== 'admin' && (
                            <button
                              onClick={() => handleDeleteUser(user)}
                              className="inline-flex items-center px-3 py-2 bg-red-100 hover:bg-red-200 text-red-700 rounded-lg transition-colors duration-200"
                              title="Delete User"
                            >
                              <i className="fas fa-trash text-sm"></i>
                            </button>
                          )}
                        </div>
                      </td>
                </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        ) : (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-12">
            <div className="text-center">
              <div className="inline-flex items-center justify-center w-24 h-24 bg-gray-100 rounded-full mb-6">
                <i className="fas fa-users text-4xl text-gray-400"></i>
              </div>
              <h3 className="text-2xl font-bold text-gray-900 mb-3">
                {searchTerm || roleFilter !== 'all' ? 'No Matching Users' : 'No Users Found'}
              </h3>
              <p className="text-gray-600 mb-6 max-w-md mx-auto">
                {searchTerm || roleFilter !== 'all' 
                  ? 'Try adjusting your search criteria or filters to find users.'
                  : 'Get started by creating your first user account to manage access to the system.'
                }
              </p>
              {(!searchTerm && roleFilter === 'all') && (
                <button
                  onClick={openCreateModal}
                  className="bg-green-600 hover:bg-green-700 text-white px-6 py-3 rounded-lg font-medium transition-colors duration-200"
                >
                  <i className="fas fa-plus mr-2"></i>Create First User
                </button>
              )}
            </div>
          </div>
        )}

        {/* Create User Modal */}
        {showCreateModal && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
            <div className="bg-white rounded-lg max-w-lg w-full shadow-lg">
              <div className="bg-green-600 rounded-t-lg p-6">
                <div className="flex justify-between items-center">
                  <div className="flex items-center">
                    <div className="bg-white bg-opacity-20 p-2 rounded-lg mr-3">
                      <i className="fas fa-user-plus text-white text-lg"></i>
                    </div>
                    <h3 className="text-xl font-bold text-white">Create New User</h3>
                  </div>
                  <button 
                    onClick={closeModals} 
                    className="text-white hover:bg-white hover:bg-opacity-20 p-2 rounded-lg transition-all duration-200"
                  >
                    <i className="fas fa-times text-lg"></i>
                  </button>
                </div>
                <p className="text-green-100 mt-2">Add a new user to the system with custom permissions</p>
              </div>
              
              <div className="p-6">

                <form onSubmit={handleCreateUser} className="space-y-6">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div>
                      <label className="block text-sm font-semibold text-gray-700 mb-3">
                        <i className="fas fa-user mr-2 text-gray-600"></i>Username
                      </label>
                      <input
                        type="text"
                        value={formData.username}
                        onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                        placeholder="Enter username"
                        className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 transition-colors duration-200"
                        required
                        minLength={3}
                      />
                      <p className="text-xs text-gray-500 mt-1">Minimum 3 characters</p>
                    </div>

                    <div>
                      <label className="block text-sm font-semibold text-gray-700 mb-3">
                        <i className="fas fa-shield-alt mr-2 text-gray-600"></i>Role
                      </label>
                      <select
                        value={formData.role}
                        onChange={(e) => setFormData({ ...formData, role: e.target.value })}
                        className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 transition-colors duration-200"
                        required
                      >
                        <option value="user">Standard User</option>
                        <option value="admin">Administrator</option>
                      </select>
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-semibold text-gray-700 mb-3">
                      <i className="fas fa-lock mr-2 text-gray-600"></i>Password
                    </label>
                    <input
                      type="password"
                      value={formData.password}
                      onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                      placeholder="Enter secure password"
                      className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 transition-colors duration-200"
                      required
                      minLength={6}
                    />
                    <p className="text-xs text-gray-500 mt-1">Minimum 6 characters</p>
                  </div>

                  <div>
                    <label className="block text-sm font-semibold text-gray-700 mb-3">
                      <i className="fas fa-list-ol mr-2 text-gray-600"></i>Session Limit
                    </label>
                    <input
                      type="number"
                      value={formData.session_limit}
                      onChange={(e) => setFormData({ ...formData, session_limit: parseInt(e.target.value) || 5 })}
                      placeholder="Enter session limit"
                      className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 transition-colors duration-200"
                      min="-1"
                      required
                    />
                    <div className="flex items-center mt-2 p-3 bg-gray-50 border border-gray-200 rounded-lg">
                      <i className="fas fa-info-circle text-gray-600 mr-2"></i>
                      <p className="text-xs text-gray-700">Use -1 for unlimited sessions (recommended for admins)</p>
                    </div>
                  </div>

                  <div className="flex gap-4 pt-6 border-t border-gray-200">
                    <button
                      type="button"
                      onClick={closeModals}
                      className="flex-1 px-6 py-3 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors duration-200 font-medium"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      disabled={loading}
                      className="flex-1 bg-green-600 hover:bg-green-700 disabled:bg-gray-400 text-white px-6 py-3 rounded-lg font-medium transition-colors duration-200 disabled:cursor-not-allowed"
                    >
                      {loading ? (
                        <>
                          <i className="fas fa-spinner fa-spin mr-2"></i>
                          Creating User...
                        </>
                      ) : (
                        <>
                          <i className="fas fa-user-plus mr-2"></i>
                          Create User
                        </>
                      )}
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
            <div className="bg-white rounded-lg max-w-lg w-full shadow-lg">
              <div className="bg-green-600 rounded-t-lg p-6">
                <div className="flex justify-between items-center">
                  <div className="flex items-center">
                    <div className="bg-white bg-opacity-20 p-2 rounded-lg mr-3">
                      <i className="fas fa-user-edit text-white text-lg"></i>
                    </div>
                    <div>
                      <h3 className="text-xl font-bold text-white">Edit User</h3>
                      <p className="text-green-100 text-sm">@{selectedUser.username}</p>
                    </div>
                  </div>
                  <button 
                    onClick={closeModals} 
                    className="text-white hover:bg-white hover:bg-opacity-20 p-2 rounded-lg transition-all duration-200"
                  >
                    <i className="fas fa-times text-lg"></i>
                  </button>
                </div>
                <p className="text-green-100 mt-2">Update user information and permissions</p>
              </div>
              
              <div className="p-6">

                <form onSubmit={handleUpdateUser} className="space-y-6">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div>
                      <label className="block text-sm font-semibold text-gray-700 mb-3">
                        <i className="fas fa-user mr-2 text-gray-600"></i>Username
                      </label>
                      <input
                        type="text"
                        value={formData.username}
                        onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                        placeholder="Enter new username"
                        className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 transition-colors duration-200"
                        minLength={3}
                      />
                      <p className="text-xs text-gray-500 mt-1">Leave unchanged if not updating</p>
                    </div>

                    <div>
                      <label className="block text-sm font-semibold text-gray-700 mb-3">
                        <i className="fas fa-shield-alt mr-2 text-gray-600"></i>Role
                      </label>
                      <select
                        value={formData.role}
                        onChange={(e) => setFormData({ ...formData, role: e.target.value })}
                        className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 transition-colors duration-200"
                        disabled={selectedUser.username === 'admin'}
                      >
                        <option value="user">Standard User</option>
                        <option value="admin">Administrator</option>
                      </select>
                      {selectedUser.username === 'admin' && (
                        <div className="flex items-center mt-2 p-2 bg-yellow-50 border border-yellow-200 rounded-lg">
                          <i className="fas fa-lock text-yellow-600 mr-2"></i>
                          <p className="text-xs text-yellow-700">Admin user role cannot be changed</p>
                        </div>
                      )}
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-semibold text-gray-700 mb-3">
                      <i className="fas fa-lock mr-2 text-gray-600"></i>New Password
                    </label>
                    <input
                      type="password"
                      value={formData.password}
                      onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                      placeholder="Enter new password (optional)"
                      className="w-full px-4 py-3 border-2 border-gray-200 rounded-xl focus:ring-2 focus:ring-green-500 focus:border-green-500 transition-all duration-200"
                      minLength={6}
                    />
                    <div className="flex items-center mt-2 p-3 bg-gray-50 border border-gray-200 rounded-lg">
                      <i className="fas fa-info-circle text-gray-600 mr-2"></i>
                      <p className="text-xs text-gray-700">Leave empty to keep current password</p>
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-semibold text-gray-700 mb-3">
                      <i className="fas fa-list-ol mr-2 text-gray-600"></i>Session Limit
                    </label>
                    <input
                      type="number"
                      value={formData.session_limit}
                      onChange={(e) => setFormData({ ...formData, session_limit: parseInt(e.target.value) || 5 })}
                      placeholder="Enter session limit"
                      className="w-full px-4 py-3 border-2 border-gray-200 rounded-xl focus:ring-2 focus:ring-green-500 focus:border-green-500 transition-all duration-200"
                      min="-1"
                    />
                    <div className="flex items-center mt-2 p-3 bg-gray-50 border border-gray-200 rounded-lg">
                      <i className="fas fa-info-circle text-gray-600 mr-2"></i>
                      <p className="text-xs text-gray-700">Use -1 for unlimited sessions</p>
                    </div>
                  </div>

                  <div className="flex gap-4 pt-6 border-t border-gray-200">
                    <button
                      type="button"
                      onClick={closeModals}
                      className="flex-1 px-6 py-3 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors duration-200 font-medium"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      disabled={loading}
                      className="flex-1 bg-green-600 hover:bg-green-700 disabled:bg-gray-400 text-white px-6 py-3 rounded-lg font-medium transition-colors duration-200 disabled:cursor-not-allowed"
                    >
                      {loading ? (
                        <>
                          <i className="fas fa-spinner fa-spin mr-2"></i>
                          Updating User...
                        </>
                      ) : (
                        <>
                          <i className="fas fa-save mr-2"></i>
                          Update User
                        </>
                      )}
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