import { useState, useEffect } from 'react';
import axios from 'axios';

const SendMessageModal = ({ session, onClose, onSuccess }) => {
  const [messageType, setMessageType] = useState('text'); // 'text' or 'location'
  const [formData, setFormData] = useState({
    to: '',
    message: '',
    latitude: '',
    longitude: '',
  });
  const [sending, setSending] = useState(false);
  const [groups, setGroups] = useState([]);
  const [loadingGroups, setLoadingGroups] = useState(false);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [recipientType, setRecipientType] = useState('personal'); // 'personal' or 'group'
  const [searchTerm, setSearchTerm] = useState('');
  const [filteredGroups, setFilteredGroups] = useState([]);

  useEffect(() => {
    if (recipientType === 'group') {
      loadGroups();
    }
  }, [recipientType, session?.id]);

  useEffect(() => {
    if (groups.length > 0) {
      const filtered = groups.filter(group => {
        if (!group) return false;
        
        const groupName = (group.name || group.subject || '').toLowerCase();
        const groupId = (group.id || '').toLowerCase();
        const search = (searchTerm || '').toLowerCase();
        
        return groupName.includes(search) || groupId.includes(search);
      });
      setFilteredGroups(filtered);
    } else {
      setFilteredGroups([]);
    }
  }, [groups, searchTerm]);

  const loadGroups = async () => {
    if (!session?.id) return;
    
    setLoadingGroups(true);
    try {
      const response = await axios.get(`/api/sessions/${session.id}/groups`);
      if (response.data.success && response.data.data) {
        // Transform the group data to match our expected format
        const transformedGroups = response.data.data.map(group => ({
          id: group.jid || group.id,
          name: group.name || group.subject || 'Unnamed Group',
          subject: group.name || group.subject,
          jid: group.jid || group.id,
          participants: group.participants || [], // This may not be available in the basic response
          owner: group.owner,
          created: group.created
        }));
        
        // Filter out any invalid group objects
        const validGroups = transformedGroups.filter(group => 
          group && typeof group === 'object' && group.id
        );
        setGroups(validGroups);
      } else {
        setGroups([]);
      }
    } catch (error) {
      console.error('Error loading groups:', error);
      // Log the full error for debugging
      if (error.response) {
        console.error('Response data:', error.response.data);
        console.error('Response status:', error.response.status);
      }
      setGroups([]);
    } finally {
      setLoadingGroups(false);
    }
  };

  const copyToClipboard = async (text) => {
    try {
      await navigator.clipboard.writeText(text);
      // You could add a toast notification here
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  };

  const handleGroupSelect = (groupId, groupName) => {
    if (!groupId) return;
    
    setFormData({ ...formData, to: groupId });
    setSearchTerm(groupName || '');
    setShowSuggestions(false);
  };

  const handleInputChange = (e) => {
    const value = e.target.value;
    setFormData({ ...formData, to: value });
    
    if (recipientType === 'group') {
      setSearchTerm(value);
      setShowSuggestions(value.length > 0);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!formData.to) return;
    
    if (messageType === 'text' && !formData.message) return;
    if (messageType === 'location' && (!formData.latitude || !formData.longitude)) return;

    setSending(true);
    
    try {
      let response;
      
      if (messageType === 'text') {
        response = await axios.post(`/api/sessions/${session.id}/send`, {
          to: formData.to,
          message: formData.message,
        });
      } else if (messageType === 'location') {
        response = await axios.post(`/api/sessions/${session.id}/send-location`, {
          to: formData.to,
          latitude: parseFloat(formData.latitude),
          longitude: parseFloat(formData.longitude),
        });
      }

      if (response.data.success) {
        onSuccess();
        onClose();
      }
    } catch (error) {
      console.error('Error sending message:', error);
    } finally {
      setSending(false);
    }
  };

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  const getCurrentLocation = () => {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (position) => {
          setFormData({
            ...formData,
            latitude: position.coords.latitude.toString(),
            longitude: position.coords.longitude.toString(),
          });
        },
        (error) => {
          console.error('Error getting location:', error);
          alert('Unable to retrieve your location. Please enter coordinates manually.');
        }
      );
    } else {
      alert('Geolocation is not supported by this browser.');
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-3xl shadow-2xl max-w-2xl w-full max-h-[90vh] overflow-hidden border border-gray-200/50 animate-scale-in">
        {/* Enhanced Header */}
        <div className="relative bg-gradient-to-r from-primary-500 via-primary-600 to-primary-700 p-6">
          <div className="absolute inset-0 bg-gradient-to-r from-white/10 to-transparent"></div>
          <div className="relative flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div className="w-12 h-12 bg-white/20 backdrop-blur-sm rounded-2xl flex items-center justify-center shadow-lg border border-white/30">
                <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
                </svg>
              </div>
              <div>
                <h3 className="text-xl font-bold text-white">Send Message</h3>
                <p className="text-primary-100 text-sm">
                  Session: {session?.name || `#${session?.id}`}
                </p>
              </div>
            </div>
            <button
              onClick={onClose}
              className="w-10 h-10 bg-white/20 backdrop-blur-sm hover:bg-white/30 rounded-xl flex items-center justify-center transition-all duration-200 border border-white/30"
            >
              <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        {/* Enhanced Content Area */}
        <div className="p-6 overflow-y-auto max-h-[calc(90vh-200px)]">
          {/* Message Type Selector */}
          <div className="mb-8">
            <div className="text-sm font-semibold text-gray-700 mb-4 flex items-center">
              <svg className="w-4 h-4 mr-2 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 4V2a1 1 0 011-1h8a1 1 0 011 1v2m-9 3v10a2 2 0 002 2h6a2 2 0 002-2V7M7 7h10M7 7l1-3m9 3l-1-3" />
              </svg>
              Message Type
            </div>
            <div className="grid grid-cols-2 gap-3">
              <button
                type="button"
                onClick={() => setMessageType('text')}
                className={`relative p-4 rounded-2xl border-2 transition-all duration-300 group ${
                  messageType === 'text' 
                    ? 'border-primary-500 bg-gradient-to-br from-primary-50 to-primary-100 shadow-lg shadow-primary-200/50' 
                    : 'border-gray-200 bg-white hover:border-primary-300 hover:shadow-md'
                }`}
              >
                <div className="flex flex-col items-center">
                  <div className={`w-12 h-12 rounded-xl flex items-center justify-center mb-3 transition-all duration-300 ${
                    messageType === 'text' 
                      ? 'bg-primary-500 text-white shadow-lg' 
                      : 'bg-gray-100 text-gray-600 group-hover:bg-primary-100 group-hover:text-primary-600'
                  }`}>
                    <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                    </svg>
                  </div>
                  <span className={`font-medium transition-colors duration-300 ${
                    messageType === 'text' ? 'text-primary-700' : 'text-gray-700'
                  }`}>
                    Text Message
                  </span>
                  <span className="text-xs text-gray-500 mt-1">Send text content</span>
                </div>
              </button>
              <button
                type="button"
                onClick={() => setMessageType('location')}
                className={`relative p-4 rounded-2xl border-2 transition-all duration-300 group ${
                  messageType === 'location' 
                    ? 'border-primary-500 bg-gradient-to-br from-primary-50 to-primary-100 shadow-lg shadow-primary-200/50' 
                    : 'border-gray-200 bg-white hover:border-primary-300 hover:shadow-md'
                }`}
              >
                <div className="flex flex-col items-center">
                  <div className={`w-12 h-12 rounded-xl flex items-center justify-center mb-3 transition-all duration-300 ${
                    messageType === 'location' 
                      ? 'bg-primary-500 text-white shadow-lg' 
                      : 'bg-gray-100 text-gray-600 group-hover:bg-primary-100 group-hover:text-primary-600'
                  }`}>
                    <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                  </div>
                  <span className={`font-medium transition-colors duration-300 ${
                    messageType === 'location' ? 'text-primary-700' : 'text-gray-700'
                  }`}>
                    Send Location
                  </span>
                  <span className="text-xs text-gray-500 mt-1">Share coordinates</span>
                </div>
              </button>
            </div>
          </div>

          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Recipient Type Selector */}
            <div className="mb-8">
              <div className="text-sm font-semibold text-gray-700 mb-4 flex items-center">
                <svg className="w-4 h-4 mr-2 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 9l4-4 4 4m0 6l-4 4-4-4" />
                </svg>
                Send To
              </div>
              <div className="grid grid-cols-2 gap-3">
                <button
                  type="button"
                  onClick={() => {
                    setRecipientType('personal');
                    setFormData({ ...formData, to: '' });
                    setSearchTerm('');
                    setShowSuggestions(false);
                  }}
                  className={`relative p-4 rounded-2xl border-2 transition-all duration-300 group ${
                    recipientType === 'personal' 
                      ? 'border-blue-500 bg-gradient-to-br from-blue-50 to-blue-100 shadow-lg shadow-blue-200/50' 
                      : 'border-gray-200 bg-white hover:border-blue-300 hover:shadow-md'
                  }`}
                >
                  <div className="flex flex-col items-center">
                    <div className={`w-12 h-12 rounded-xl flex items-center justify-center mb-3 transition-all duration-300 ${
                      recipientType === 'personal' 
                        ? 'bg-blue-500 text-white shadow-lg' 
                        : 'bg-gray-100 text-gray-600 group-hover:bg-blue-100 group-hover:text-blue-600'
                    }`}>
                      <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                      </svg>
                    </div>
                    <span className={`font-medium transition-colors duration-300 ${
                      recipientType === 'personal' ? 'text-blue-700' : 'text-gray-700'
                    }`}>
                      Personal Chat
                    </span>
                    <span className="text-xs text-gray-500 mt-1">Individual contact</span>
                  </div>
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setRecipientType('group');
                    setFormData({ ...formData, to: '' });
                    setSearchTerm('');
                    setShowSuggestions(false);
                  }}
                  className={`relative p-4 rounded-2xl border-2 transition-all duration-300 group ${
                    recipientType === 'group' 
                      ? 'border-green-500 bg-gradient-to-br from-green-50 to-green-100 shadow-lg shadow-green-200/50' 
                      : 'border-gray-200 bg-white hover:border-green-300 hover:shadow-md'
                  }`}
                >
                  <div className="flex flex-col items-center">
                    <div className={`w-12 h-12 rounded-xl flex items-center justify-center mb-3 transition-all duration-300 ${
                      recipientType === 'group' 
                        ? 'bg-green-500 text-white shadow-lg' 
                        : 'bg-gray-100 text-gray-600 group-hover:bg-green-100 group-hover:text-green-600'
                    }`}>
                      <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                      </svg>
                    </div>
                    <span className={`font-medium transition-colors duration-300 ${
                      recipientType === 'group' ? 'text-green-700' : 'text-gray-700'
                    }`}>
                      Group Chat
                    </span>
                    <span className="text-xs text-gray-500 mt-1">Multiple members</span>
                  </div>
                </button>
              </div>
            </div>

            {/* Recipient Input */}
            <div className="mb-6">
              <div className="text-sm font-semibold text-gray-700 mb-4 flex items-center">
                <svg className="w-4 h-4 mr-2 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
                {recipientType === 'personal' ? 'Phone Number' : 'Group Selection'}
              </div>
              <div className="relative">
                <div className="relative">
                  <input
                    type="text"
                    name="to"
                    value={formData.to}
                    onChange={handleInputChange}
                    onFocus={() => recipientType === 'group' && setShowSuggestions(true)}
                    placeholder={
                      recipientType === 'personal' 
                        ? "6281234567890 or 6281234567890@s.whatsapp.net"
                        : "Type group name or ID..."
                    }
                    className="w-full px-4 py-3 pr-12 border-2 border-gray-200 rounded-2xl bg-gray-50 focus:bg-white focus:border-primary-500 focus:ring-4 focus:ring-primary-100 transition-all duration-300 placeholder-gray-500"
                    required
                  />
                  <div className="absolute right-3 top-1/2 transform -translate-y-1/2 flex items-center gap-2">
                    {recipientType === 'personal' && (
                      <svg className="w-5 h-5 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
                      </svg>
                    )}
                    {recipientType === 'group' && (
                      <button
                        type="button"
                        onClick={() => setShowSuggestions(!showSuggestions)}
                        className="p-1 text-green-500 hover:text-green-700 hover:bg-green-50 rounded-lg transition-colors"
                        title="Browse Groups"
                      >
                        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                        </svg>
                      </button>
                    )}
                  </div>
                </div>
                <p className="text-xs text-gray-500 mt-2 flex items-center">
                  <svg className="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  {recipientType === 'personal' 
                    ? 'Enter WhatsApp number with country code (e.g., 628123456789)'
                    : 'Start typing to search groups by name or select from the list'
                  }
                </p>
              </div>
              
              {/* Enhanced Group Autocomplete */}
              {recipientType === 'group' && showSuggestions && (
                <div className="mt-4 border border-gray-200 rounded-2xl bg-white shadow-2xl overflow-hidden backdrop-blur-sm">
                  <div className="p-4 border-b border-gray-100 bg-gradient-to-r from-green-50 to-emerald-50">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <svg className="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                        </svg>
                        <h4 className="font-semibold text-green-800">
                          Available Groups
                          {filteredGroups.length > 0 && (
                            <span className="ml-2 px-2 py-1 bg-green-200 text-green-800 text-xs rounded-full">
                              {filteredGroups.length}
                            </span>
                          )}
                        </h4>
                      </div>
                      {loadingGroups && (
                        <div className="w-5 h-5 border-2 border-green-300 border-t-green-600 rounded-full animate-spin"></div>
                      )}
                    </div>
                  </div>
                  
                  <div className="max-h-64 overflow-y-auto">
                    {loadingGroups ? (
                      <div className="p-8 text-center">
                        <div className="w-8 h-8 border-2 border-gray-300 border-t-primary-600 rounded-full animate-spin mx-auto mb-3"></div>
                        <p className="text-gray-500">Loading groups...</p>
                      </div>
                    ) : filteredGroups.length === 0 ? (
                      <div className="p-8 text-center">
                        <svg className="w-12 h-12 text-gray-300 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                        </svg>
                        <p className="text-gray-500 font-medium">No groups found</p>
                        <p className="text-gray-400 text-sm mt-1">
                          {searchTerm ? 'Try different search terms' : 'No groups available for this session'}
                        </p>
                      </div>
                    ) : (
                      <div className="divide-y divide-gray-100">
                        {filteredGroups.map((group, index) => {
                          if (!group || !group.id) return null;
                          
                          return (
                          <div 
                            key={group.id || index}
                            className="p-4 hover:bg-gradient-to-r hover:from-green-25 hover:to-emerald-25 transition-all duration-200 cursor-pointer group"
                            onClick={() => handleGroupSelect(group.id, group.name || group.subject)}
                          >
                            <div className="flex items-center justify-between">
                              <div className="flex-1 min-w-0">
                                <div className="flex items-center gap-3">
                                  <div className="relative">
                                    <div className="w-12 h-12 bg-gradient-to-br from-green-500 to-emerald-600 rounded-2xl flex items-center justify-center shadow-lg group-hover:shadow-xl group-hover:scale-105 transition-all duration-200">
                                      <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                                      </svg>
                                    </div>
                                    <div className="absolute -bottom-1 -right-1 w-4 h-4 bg-green-500 rounded-full border-2 border-white"></div>
                                  </div>
                                  <div className="flex-1 min-w-0">
                                    <h5 className="font-semibold text-gray-900 truncate group-hover:text-green-800 transition-colors">
                                      {group.name || group.subject || 'Unnamed Group'}
                                    </h5>
                                    <p className="text-sm text-gray-500 font-mono truncate bg-gray-100 px-2 py-1 rounded-lg mt-1">
                                      {group.id}
                                    </p>
                                    {group.participants && (
                                      <div className="flex items-center gap-1 mt-2">
                                        <svg className="w-3 h-3 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
                                        </svg>
                                        <span className="text-xs text-gray-500">
                                          {group.participants.length} members
                                        </span>
                                      </div>
                                    )}
                                  </div>
                                </div>
                              </div>
                              <div className="flex items-center gap-2 ml-4 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
                                <button
                                  type="button"
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    copyToClipboard(group.id);
                                  }}
                                  className="p-2 text-gray-400 hover:text-primary-600 hover:bg-primary-50 rounded-xl transition-all duration-200 hover:scale-110"
                                  title="Copy Group ID"
                                >
                                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                                  </svg>
                                </button>
                                <div className="w-8 h-8 bg-green-500 rounded-xl flex items-center justify-center shadow-lg">
                                  <svg className="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                                  </svg>
                                </div>
                              </div>
                            </div>
                          </div>
                        );
                        })}
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>

            {/* Enhanced Text Message Field */}
            {messageType === 'text' && (
              <div className="mb-6">
                <div className="text-sm font-semibold text-gray-700 mb-4 flex items-center">
                  <svg className="w-4 h-4 mr-2 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                  </svg>
                  Message Content
                </div>
                <div className="relative">
                  <textarea
                    name="message"
                    value={formData.message}
                    onChange={handleChange}
                    placeholder="Type your message here... You can use emojis 😊 and line breaks"
                    rows="5"
                    className="w-full px-4 py-3 border-2 border-gray-200 rounded-2xl bg-gray-50 focus:bg-white focus:border-primary-500 focus:ring-4 focus:ring-primary-100 transition-all duration-300 placeholder-gray-500 resize-none"
                    required
                  />
                  <div className="absolute bottom-3 right-3 flex items-center gap-2">
                    <span className="text-xs text-gray-400">
                      {formData.message.length} characters
                    </span>
                    <svg className="w-4 h-4 text-primary-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                    </svg>
                  </div>
                </div>
                <p className="text-xs text-gray-500 mt-2 flex items-center">
                  <svg className="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  Your message will be sent exactly as typed, including formatting and emojis
                </p>
              </div>
            )}

            {/* Location Fields */}
            {messageType === 'location' && (
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div className="text-body-medium">Location Coordinates</div>
                  <button
                    type="button"
                    onClick={getCurrentLocation}
                    className="btn btn-sm btn-secondary"
                  >
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                    Use Current Location
                  </button>
                </div>
                
                <div className="grid grid-cols-2 gap-4">
                  <div className="input-group">
                    <label className="input-label">Latitude</label>
                    <input
                      type="number"
                      name="latitude"
                      value={formData.latitude}
                      onChange={handleChange}
                      placeholder="-6.200000"
                      step="any"
                      className="input"
                      required
                    />
                  </div>
                  <div className="input-group">
                    <label className="input-label">Longitude</label>
                    <input
                      type="number"
                      name="longitude"
                      value={formData.longitude}
                      onChange={handleChange}
                      placeholder="106.816666"
                      step="any"
                      className="input"
                      required
                    />
                  </div>
                </div>
                
                {formData.latitude && formData.longitude && (
                  <div className="p-4 bg-gray-50 rounded-lg border border-gray-200">
                    <div className="text-body-medium mb-2">Preview Location</div>
                    <div className="text-caption text-gray-600">
                      Latitude: {formData.latitude}<br />
                      Longitude: {formData.longitude}
                    </div>
                    <a
                      href={`https://www.google.com/maps?q=${formData.latitude},${formData.longitude}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-primary-600 hover:text-primary-700 text-caption underline mt-2 inline-block"
                    >
                      View on Google Maps →
                    </a>
                  </div>
                )}
              </div>
            )}
          </form>
        </div>

        {/* Enhanced Footer */}
        <div className="p-6 border-t border-gray-200 bg-gradient-to-r from-gray-50 to-gray-100">
          <div className="flex gap-4">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-6 py-3 bg-white border-2 border-gray-300 text-gray-700 font-medium rounded-2xl hover:bg-gray-50 hover:border-gray-400 transition-all duration-300 flex items-center justify-center"
            >
              <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
              Cancel
            </button>
            <button
              onClick={handleSubmit}
              disabled={sending || !formData.to || 
                (messageType === 'text' && !formData.message) ||
                (messageType === 'location' && (!formData.latitude || !formData.longitude))
              }
              className="flex-1 px-6 py-3 bg-gradient-to-r from-primary-600 to-primary-700 hover:from-primary-700 hover:to-primary-800 disabled:from-gray-400 disabled:to-gray-500 text-white font-semibold rounded-2xl shadow-lg hover:shadow-xl disabled:shadow-none transform hover:scale-[1.02] disabled:scale-100 transition-all duration-300 flex items-center justify-center"
            >
              {sending ? (
                <>
                  <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></div>
                  Sending...
                </>
              ) : (
                <>
                  <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
                  </svg>
                  Send {messageType === 'text' ? 'Message' : 'Location'}
                </>
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SendMessageModal;