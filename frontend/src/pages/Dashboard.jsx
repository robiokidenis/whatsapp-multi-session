import { useState, useEffect } from 'react';
import axios from 'axios';
import SessionCard from '../components/SessionCard';
import QRModal from '../components/QRModal';
import SendMessageModal from '../components/SendMessageModal';
import EditSessionModal from '../components/EditSessionModal';
import { useAuth } from '../contexts/AuthContext';

const Dashboard = () => {
  const { user } = useAuth();
  const [sessions, setSessions] = useState([]);
  const [newSession, setNewSession] = useState({ phone: '', name: '' });
  const [message, setMessage] = useState('');
  const [messageType, setMessageType] = useState('');
  const [qrModal, setQrModal] = useState({ show: false, session: null });
  const [sendModal, setSendModal] = useState({ show: false, session: null });
  const [editModal, setEditModal] = useState({ show: false, session: null });
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');

  useEffect(() => {
    loadSessions();
    // Auto-refresh every 5 seconds
    const interval = setInterval(loadSessions, 5000);
    return () => clearInterval(interval);
  }, []);

  const loadSessions = async () => {
    try {
      const response = await axios.get('/api/sessions');
      setSessions(response.data.data || []);
    } catch (error) {
      console.error('Error loading sessions:', error);
    }
  };

  const createSession = async (e) => {
    e.preventDefault();
    try {
      const response = await axios.post('/api/sessions', {
        phone: newSession.phone || '',
        name: newSession.name,
      });

      if (response.data.success) {
        showMessage('Session created successfully!', 'success');
        setNewSession({ phone: '', name: '' });
        await loadSessions();
      } else {
        showMessage(response.data.error || 'Failed to create session', 'error');
      }
    } catch (error) {
      showMessage('Error creating session: ' + error.message, 'error');
    }
  };

  const deleteSession = async (id) => {
    if (!window.confirm('Are you sure you want to delete this session?')) return;

    try {
      await axios.delete(`/api/sessions/${id}`);
      showMessage('Session deleted successfully!', 'success');
      await loadSessions();
    } catch (error) {
      showMessage('Error deleting session: ' + error.message, 'error');
    }
  };

  const logoutSession = async (id) => {
    try {
      await axios.post(`/api/sessions/${id}/logout`);
      showMessage('Session logged out successfully!', 'success');
    } catch (error) {
      // Even if logout fails due to session errors, still refresh to show current state
      showMessage('Session logout completed (session may have already been disconnected)', 'success');
    } finally {
      // Always refresh sessions list to show current state
      await loadSessions();
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

  const openQRModal = (session) => {
    setQrModal({ show: true, session });
  };

  const closeQRModal = () => {
    setQrModal({ show: false, session: null });
  };

  const openSendModal = (session) => {
    setSendModal({ show: true, session });
  };

  const closeSendModal = () => {
    setSendModal({ show: false, session: null });
  };

  const openEditModal = (session) => {
    setEditModal({ show: true, session });
  };

  const closeEditModal = () => {
    setEditModal({ show: false, session: null });
  };

  const updateSession = (updatedSession) => {
    setSessions(prevSessions =>
      prevSessions.map(session =>
        session.id === updatedSession.id ? updatedSession : session
      )
    );
    showMessage('Session updated successfully!', 'success');
  };

  // Filter sessions based on search and status
  const filteredSessions = sessions.filter(session => {
    const matchesSearch = session.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         session.phone?.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus = statusFilter === 'all' || 
                         (statusFilter === 'connected' && session.status === 'Connected') ||
                         (statusFilter === 'disconnected' && session.status !== 'Connected');
    return matchesSearch && matchesStatus;
  });

  return (
    <div>
      {/* Add Session Form */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
        <div className="flex items-center mb-6">
          <div className="bg-green-600 p-3 rounded-lg mr-4">
            <i className="fas fa-plus-circle text-white text-xl"></i>
          </div>
          <div>
            <h2 className="text-xl font-bold text-gray-900">Create New Session</h2>
            <p className="text-gray-600">Add a new WhatsApp session to manage multiple accounts</p>
          </div>
        </div>
        
        <form onSubmit={createSession} className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Phone Number
              </label>
              <input
                type="text"
                value={newSession.phone}
                onChange={(e) => setNewSession({ ...newSession, phone: e.target.value })}
                placeholder="Enter phone number (optional)"
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 transition-colors duration-200"
              />
              <p className="text-xs text-gray-500 mt-1">Auto-generated if left empty</p>
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Session Name
              </label>
              <input
                type="text"
                value={newSession.name}
                onChange={(e) => setNewSession({ ...newSession, name: e.target.value })}
                placeholder="Enter session name (optional)"
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 transition-colors duration-200"
              />
              <p className="text-xs text-gray-500 mt-1">Helps identify this session</p>
            </div>
          </div>
          
          <div className="flex justify-end">
            <button
              type="submit"
              className="bg-green-600 hover:bg-green-700 text-white px-6 py-2 rounded-lg font-medium transition-colors duration-200"
            >
              <i className="fas fa-plus mr-2"></i>Create Session
            </button>
          </div>
        </form>
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

      {/* Search and Filter Section */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="flex-1">
            <div className="relative">
              <i className="fas fa-search absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400"></i>
              <input
                type="text"
                placeholder="Search sessions by name or phone..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 transition-colors duration-200"
              />
            </div>
          </div>
          <div className="w-full sm:w-48">
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 transition-colors duration-200"
            >
              <option value="all">All Sessions</option>
              <option value="connected">Connected</option>
              <option value="disconnected">Disconnected</option>
            </select>
          </div>
          <button
            onClick={loadSessions}
            className="bg-gray-100 hover:bg-gray-200 text-gray-700 px-4 py-3 rounded-lg font-medium transition-colors duration-200 flex items-center"
          >
            <i className="fas fa-sync-alt mr-2"></i>Refresh
          </button>
        </div>
        
        {/* Stats */}
        <div className="mt-4 flex flex-wrap gap-4">
          <div className="bg-gray-50 px-4 py-2 rounded-lg border border-gray-200">
            <span className="text-gray-700 font-medium">
              Total: {sessions.length}
            </span>
          </div>
          <div className="bg-gray-50 px-4 py-2 rounded-lg border border-gray-200">
            <span className="text-gray-700 font-medium">
              Connected: {sessions.filter(s => s.status === 'Connected').length}
            </span>
          </div>
          <div className="bg-gray-50 px-4 py-2 rounded-lg border border-gray-200">
            <span className="text-gray-700 font-medium">
              Disconnected: {sessions.filter(s => s.status !== 'Connected').length}
            </span>
          </div>
          <div className="bg-gray-50 px-4 py-2 rounded-lg border border-gray-200">
            <span className="text-gray-700 font-medium">
              Filtered: {filteredSessions.length}
            </span>
          </div>
        </div>
      </div>

      {/* Sessions Grid */}
      {filteredSessions.length > 0 ? (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {filteredSessions.map((session, index) => (
            <div 
              key={session.id}
              className="animate-slide-up"
              style={{ animationDelay: `${index * 0.1}s` }}
            >
              <SessionCard
                session={session}
                onShowQR={openQRModal}
                onSendMessage={openSendModal}
                onLogout={logoutSession}
                onDelete={deleteSession}
                onEdit={openEditModal}
              />
            </div>
          ))}
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-12">
          <div className="text-center">
            <div className="inline-flex items-center justify-center w-24 h-24 bg-gray-100 rounded-full mb-6">
              <i className="fas fa-mobile-alt text-4xl text-gray-400"></i>
            </div>
            <h3 className="text-2xl font-bold text-gray-900 mb-3">
              {sessions.length === 0 ? 'No WhatsApp Sessions' : 'No Matching Sessions'}
            </h3>
            <p className="text-gray-600 mb-6 max-w-md mx-auto">
              {sessions.length === 0 
                ? 'Create your first WhatsApp session to start managing multiple accounts from one dashboard.'
                : 'Try adjusting your search criteria or filters to find sessions.'
              }
            </p>
            {sessions.length === 0 && (
              <button
                onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}
                className="bg-green-600 hover:bg-green-700 text-white px-6 py-3 rounded-lg font-medium transition-colors duration-200"
              >
                <i className="fas fa-plus mr-2"></i>Create First Session
              </button>
            )}
          </div>
        </div>
      )}

      {/* Modals */}
      {qrModal.show && (
        <QRModal
          session={qrModal.session}
          onClose={closeQRModal}
          onSuccess={loadSessions}
        />
      )}

      {sendModal.show && (
        <SendMessageModal
          session={sendModal.session}
          onClose={closeSendModal}
          onSuccess={() => showMessage('Message sent successfully!', 'success')}
        />
      )}

      {editModal.show && (
        <EditSessionModal
          isOpen={editModal.show}
          session={editModal.session}
          onClose={closeEditModal}
          onUpdate={updateSession}
        />
      )}
    </div>
  );
};

export default Dashboard;