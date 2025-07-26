import { useState, useEffect } from 'react';
import axios from 'axios';
import Layout from '../components/Layout';
import SessionCard from '../components/SessionCard';
import QRModal from '../components/QRModal';
import SendMessageModal from '../components/SendMessageModal';
import EditSessionModal from '../components/EditSessionModal';
import UserManagement from '../components/UserManagement';
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
  const [activeTab, setActiveTab] = useState('sessions');

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

  return (
    <Layout>
      {/* Tab Navigation */}
      <div className="mb-6">
        <div className="border-b border-gray-200">
          <nav className="-mb-px flex space-x-8">
            <button
              onClick={() => setActiveTab('sessions')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'sessions'
                  ? 'border-green-500 text-green-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              <i className="fas fa-mobile-alt mr-2"></i>
              WhatsApp Sessions
            </button>
            {user?.role === 'admin' && (
              <button
                onClick={() => setActiveTab('users')}
                className={`py-2 px-1 border-b-2 font-medium text-sm ${
                  activeTab === 'users'
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                <i className="fas fa-users mr-2"></i>
                User Management
              </button>
            )}
          </nav>
        </div>
      </div>

      {/* Tab Content */}
      {activeTab === 'sessions' && (
        <>
          {/* Add Session Form */}
          <div className="bg-white rounded-lg shadow-md p-6 mb-8">
        <h2 className="text-xl font-semibold mb-4 flex items-center">
          <i className="fas fa-plus-circle mr-2 text-green-600"></i>
          Add New WhatsApp Session
        </h2>
        <form onSubmit={createSession} className="flex flex-col md:flex-row gap-4">
          <input
            type="text"
            value={newSession.phone}
            onChange={(e) => setNewSession({ ...newSession, phone: e.target.value })}
            placeholder="Phone number (optional - will auto-generate if empty)"
            className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent"
          />
          <input
            type="text"
            value={newSession.name}
            onChange={(e) => setNewSession({ ...newSession, name: e.target.value })}
            placeholder="Session name (optional)"
            className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent"
          />
          <button
            type="submit"
            className="bg-green-600 hover:bg-green-700 text-white px-6 py-2 rounded-lg font-medium transition duration-200"
          >
            <i className="fas fa-plus mr-2"></i>Add Session
          </button>
        </form>
        
        {message && (
          <div className={`mt-4 p-3 rounded-lg ${
            messageType === 'error'
              ? 'bg-red-100 border border-red-400 text-red-700'
              : 'bg-green-100 border border-green-400 text-green-700'
          }`}>
            {message}
          </div>
        )}
      </div>

      {/* Session Stats */}
      <div className="mb-6">
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-semibold">Sessions ({sessions.length})</h2>
          <button
            onClick={loadSessions}
            className="bg-green-600 hover:bg-green-700 text-white px-3 py-1 rounded text-sm"
          >
            <i className="fas fa-sync-alt mr-1"></i>Refresh
          </button>
        </div>
      </div>

      {/* Sessions Grid */}
      {sessions.length > 0 ? (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {sessions.map((session) => (
            <SessionCard
              key={session.id}
              session={session}
              onShowQR={openQRModal}
              onSendMessage={openSendModal}
              onLogout={logoutSession}
              onDelete={deleteSession}
              onEdit={openEditModal}
            />
          ))}
        </div>
      ) : (
        <div className="text-center py-12">
          <i className="fas fa-mobile-alt text-6xl text-gray-300 mb-4"></i>
          <h3 className="text-xl font-semibold text-gray-600 mb-2">No WhatsApp Sessions</h3>
          <p className="text-gray-500">Add your first WhatsApp session to get started!</p>
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
        </>
      )}

      {/* User Management Tab */}
      {activeTab === 'users' && user?.role === 'admin' && (
        <UserManagement />
      )}
    </Layout>
  );
};

export default Dashboard;