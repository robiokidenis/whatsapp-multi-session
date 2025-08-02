import { useState, useEffect, useMemo, useCallback } from "react";
import axios from "axios";
import SessionCard from "../components/SessionCard";
import QRModal from "../components/QRModal";
import SendMessageModal from "../components/SendMessageModal";
import EditSessionModal from "../components/EditSessionModal";
import CreateSessionModal from "../components/CreateSessionModal";
import DeleteConfirmDialog from "../components/DeleteConfirmDialog";
import { useAuth } from "../contexts/AuthContext";
import { useNotification } from "../contexts/NotificationContext";

const Dashboard = () => {
  const { user } = useAuth();
  const { showSuccess, showError, showWarning } = useNotification();
  const [sessions, setSessions] = useState([]);
  const [qrModal, setQrModal] = useState({ show: false, session: null });
  const [sendModal, setSendModal] = useState({ show: false, session: null });
  const [editModal, setEditModal] = useState({ show: false, session: null });
  const [createModal, setCreateModal] = useState(false);
  const [deleteModal, setDeleteModal] = useState({ show: false, sessionId: null, sessionName: null });
  const [bulkDeleteModal, setBulkDeleteModal] = useState({ show: false, count: 0 });
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [selectedSessions, setSelectedSessions] = useState([]);
  const [bulkActionLoading, setBulkActionLoading] = useState(false);
  const [showFilters, setShowFilters] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    loadSessions();
    const interval = setInterval(loadSessions, 10000);
    return () => clearInterval(interval);
  }, []);

  const loadSessions = async () => {
    try {
      setIsLoading(true);
      const response = await axios.get("/api/sessions");
      setSessions(response.data.data || []);
    } catch (error) {
      console.error("Error loading sessions:", error);
      showError("Failed to load sessions");
    } finally {
      setIsLoading(false);
    }
  };

  const stats = useMemo(() => {
    const total = sessions.length;
    const connected = sessions.filter(
      (s) => (s.status === "Connected" || s.connected) && s.logged_in
    ).length;
    const offline = total - connected;
    return { total, connected, offline };
  }, [sessions]);

  const filteredSessions = useMemo(() => {
    return sessions.filter((session) => {
      const matchesSearch =
        !searchTerm ||
        session.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        session.phone?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        session.id?.toString().includes(searchTerm.toLowerCase());

      const isConnected =
        (session.status === "Connected" || session.connected) &&
        session.logged_in;
      const matchesStatus =
        statusFilter === "all" ||
        (statusFilter === "connected" && isConnected) ||
        (statusFilter === "offline" && !isConnected);

      return matchesSearch && matchesStatus;
    });
  }, [sessions, searchTerm, statusFilter]);

  const handleCreateSuccess = async (message) => {
    showSuccess(message);
    await loadSessions();
  };

  const deleteSession = async (id) => {
    const session = sessions.find(s => s.id === id);
    setDeleteModal({ 
      show: true, 
      sessionId: id, 
      sessionName: session?.name || `Session ${id}`
    });
  };

  const confirmDeleteSession = async () => {
    setDeleteLoading(true);
    try {
      await axios.delete(`/api/sessions/${deleteModal.sessionId}`);
      showSuccess("Session deleted");
      setDeleteModal({ show: false, sessionId: null, sessionName: null });
      await loadSessions();
    } catch (error) {
      showError("Failed to delete session");
    } finally {
      setDeleteLoading(false);
    }
  };

  const logoutSession = async (id) => {
    try {
      await axios.post(`/api/sessions/${id}/logout`);
      showSuccess("Session logged out");
    } catch (error) {
      showWarning("Logout completed");
    } finally {
      await loadSessions();
    }
  };


  const handleSelectSession = (sessionId) => {
    setSelectedSessions((prev) =>
      prev.includes(sessionId)
        ? prev.filter((id) => id !== sessionId)
        : [...prev, sessionId]
    );
  };

  const handleSelectAll = () => {
    setSelectedSessions(
      selectedSessions.length === filteredSessions.length
        ? []
        : filteredSessions.map((s) => s.id)
    );
  };

  const handleBulkDelete = async () => {
    setBulkDeleteModal({ show: true, count: selectedSessions.length });
  };

  const confirmBulkDelete = async () => {
    setBulkActionLoading(true);
    try {
      await Promise.all(
        selectedSessions.map((id) =>
          axios.delete(`/api/sessions/${id}`).catch(() => null)
        )
      );
      showSuccess(`Deleted ${selectedSessions.length} sessions`);
      setSelectedSessions([]);
      setBulkDeleteModal({ show: false, count: 0 });
      await loadSessions();
    } catch (error) {
      showError("Bulk delete failed");
    } finally {
      setBulkActionLoading(false);
    }
  };

  const openQRModal = (session) => setQrModal({ show: true, session });
  const closeQRModal = () => setQrModal({ show: false, session: null });
  const openSendModal = (session) => setSendModal({ show: true, session });
  const closeSendModal = () => setSendModal({ show: false, session: null });
  const openEditModal = (session) => setEditModal({ show: true, session });
  const closeEditModal = () => setEditModal({ show: false, session: null });

  const updateSession = (updatedSession) => {
    setSessions((prev) =>
      prev.map((s) => (s.id === updatedSession.id ? updatedSession : s))
    );
    showSuccess("Session updated");
  };

  if (isLoading && sessions.length === 0) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="loading-spinner-lg mx-auto mb-4"></div>
          <p className="text-secondary">Loading sessions...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto px-6 py-8">
        {/* Enhanced Header */}
        <div className="mb-8">
          <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-6">
            <div>
              <div className="flex items-center gap-4 mb-3">
                <div className="p-3 bg-gradient-to-br from-primary-500 to-primary-600 rounded-2xl shadow-lg">
                  <svg className="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                  </svg>
                </div>
                <div>
                  <h1 className="text-4xl font-bold text-gray-900">Session Management</h1>
                  <p className="text-lg text-gray-600 mt-1">
                    Monitor and manage your WhatsApp business connections
                  </p>
                </div>
              </div>
            </div>
            <button
              onClick={() => setCreateModal(true)}
              className="inline-flex items-center px-6 py-3 bg-gradient-to-r from-primary-600 to-primary-700 text-white font-semibold rounded-xl hover:from-primary-700 hover:to-primary-800 transition-all duration-200 shadow-lg hover:shadow-xl transform hover:scale-[1.02]"
            >
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              Create New Session
            </button>
          </div>
        </div>

        {/* Enhanced Stats Dashboard */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-6 mb-8">
          <div className="group relative bg-gradient-to-br from-gray-50 to-white border border-gray-200 rounded-2xl shadow-sm hover:shadow-lg transition-all duration-300 overflow-hidden">
            <div className="absolute inset-0 bg-gradient-to-br from-gray-400/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
            <div className="relative p-6">
              <div className="flex items-start justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600 mb-2">Total Sessions</p>
                  <p className="text-3xl font-bold text-gray-900 mb-1">{stats.total}</p>
                  <p className="text-sm text-gray-500 flex items-center">
                    <span className="w-2 h-2 bg-gray-400 rounded-full mr-2"></span>
                    All registered sessions
                  </p>
                </div>
                <div className="p-3 bg-gray-100 rounded-xl">
                  <svg className="w-6 h-6 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                  </svg>
                </div>
              </div>
            </div>
          </div>

          <div className="group relative bg-gradient-to-br from-green-50 to-white border border-green-200 rounded-2xl shadow-sm hover:shadow-lg transition-all duration-300 overflow-hidden">
            <div className="absolute inset-0 bg-gradient-to-br from-green-400/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
            <div className="relative p-6">
              <div className="flex items-start justify-between">
                <div>
                  <p className="text-sm font-medium text-green-700 mb-2">Active Sessions</p>
                  <p className="text-3xl font-bold text-green-900 mb-1">{stats.connected}</p>
                  <p className="text-sm text-green-600 flex items-center">
                    <span className="w-2 h-2 bg-green-500 rounded-full mr-2 animate-pulse"></span>
                    Connected & ready
                  </p>
                </div>
                <div className="p-3 bg-green-100 rounded-xl">
                  <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
              </div>
            </div>
          </div>

          <div className="group relative bg-gradient-to-br from-orange-50 to-white border border-orange-200 rounded-2xl shadow-sm hover:shadow-lg transition-all duration-300 overflow-hidden">
            <div className="absolute inset-0 bg-gradient-to-br from-orange-400/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
            <div className="relative p-6">
              <div className="flex items-start justify-between">
                <div>
                  <p className="text-sm font-medium text-orange-700 mb-2">Offline Sessions</p>
                  <p className="text-3xl font-bold text-orange-900 mb-1">{stats.offline}</p>
                  <p className="text-sm text-orange-600 flex items-center">
                    <span className="w-2 h-2 bg-orange-500 rounded-full mr-2"></span>
                    Disconnected
                  </p>
                </div>
                <div className="p-3 bg-orange-100 rounded-xl">
                  <svg className="w-6 h-6 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
              </div>
            </div>
          </div>
        </div>


        {/* Enhanced Search & Filters */}
        <div className="bg-white rounded-2xl shadow-sm border border-gray-200 mb-8">
          <div className="p-6">
            <div className="flex flex-col sm:flex-row gap-4">
              <div className="flex-1">
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                    <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                  </div>
                  <input
                    type="text"
                    placeholder="Search by name, phone, or session ID..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="w-full py-3 pl-12 pr-4 text-sm border border-gray-300 rounded-xl bg-gray-50 focus:bg-white focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200 placeholder-gray-500"
                  />
                </div>
              </div>
              <button
                onClick={() => setShowFilters(!showFilters)}
                className={`px-6 py-3 font-medium rounded-xl transition-all duration-200 flex items-center gap-2 ${
                  showFilters 
                    ? "bg-primary-600 text-white shadow-lg" 
                    : "bg-gray-100 text-gray-700 hover:bg-primary-50 hover:text-primary-700"
                }`}
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.707A1 1 0 013 7V4z" />
                </svg>
                <span>Filters</span>
                {(statusFilter !== 'all' || searchTerm) && (
                  <span className="inline-flex items-center justify-center w-5 h-5 text-xs font-bold bg-white text-primary-600 rounded-full">
                    {filteredSessions.length}
                  </span>
                )}
              </button>
            </div>

            {/* Enhanced Filters */}
            {showFilters && (
              <div className="animate-fade-in mt-6 pt-6 border-t border-gray-200">
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">Connection Status</label>
                    <select
                      value={statusFilter}
                      onChange={(e) => setStatusFilter(e.target.value)}
                      className="w-full py-3 px-4 border border-gray-300 rounded-xl bg-white focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200 text-sm"
                    >
                      <option value="all">All Sessions</option>
                      <option value="connected">🟢 Connected Only</option>
                      <option value="offline">🔴 Offline Only</option>
                    </select>
                  </div>

                  <div className="sm:col-span-1 lg:col-span-3">
                    {filteredSessions.length > 0 && (
                      <div className="flex items-end gap-3">
                        <div className="flex-1">
                          <label className="block text-sm font-medium text-gray-700 mb-2">Bulk Actions</label>
                          <div className="flex items-center gap-2">
                            <button
                              onClick={handleSelectAll}
                              className="px-4 py-3 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-xl hover:bg-gray-50 transition-colors duration-200 flex items-center gap-2"
                            >
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                              </svg>
                              {selectedSessions.length === filteredSessions.length ? "Deselect All" : "Select All"}
                            </button>

                            {selectedSessions.length > 0 && (
                              <button
                                onClick={handleBulkDelete}
                                disabled={bulkActionLoading}
                                className="px-4 py-3 text-sm font-medium text-white bg-red-600 hover:bg-red-700 disabled:bg-red-400 rounded-xl transition-colors duration-200 flex items-center gap-2"
                              >
                                {bulkActionLoading ? (
                                  <>
                                    <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                                    <span>Deleting...</span>
                                  </>
                                ) : (
                                  <>
                                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                    </svg>
                                    <span>Delete ({selectedSessions.length})</span>
                                  </>
                                )}
                              </button>
                            )}
                          </div>
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Enhanced Sessions Grid */}
        <div className="mb-8">
          {filteredSessions.length === 0 ? (
            <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-12 text-center">
              <div className="w-24 h-24 bg-gradient-to-br from-gray-100 to-gray-200 rounded-3xl flex items-center justify-center mx-auto mb-6 shadow-inner">
                <svg className="w-12 h-12 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                </svg>
              </div>
              <h3 className="text-2xl font-bold text-gray-900 mb-3">
                {sessions.length === 0 ? "No Sessions Created" : "No Matching Sessions"}
              </h3>
              <p className="text-gray-600 mb-8 max-w-lg mx-auto">
                {sessions.length === 0
                  ? "Get started by creating your first WhatsApp business session to connect with customers and manage communications."
                  : "No sessions match your current search criteria. Try adjusting your filters or search terms to find what you're looking for."}
              </p>
              {sessions.length === 0 && (
                <button
                  onClick={() => setCreateModal(true)}
                  className="inline-flex items-center px-8 py-4 bg-gradient-to-r from-primary-600 to-primary-700 text-white font-semibold rounded-xl hover:from-primary-700 hover:to-primary-800 transition-all duration-200 shadow-lg hover:shadow-xl transform hover:scale-[1.02]"
                >
                  <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                  </svg>
                  Create Your First Session
                </button>
              )}
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
              {filteredSessions.map((session, index) => {
                const isSelected = selectedSessions.includes(session.id);

                return (
                  <div 
                    key={session.id} 
                    className="animate-fade-in"
                    style={{ animationDelay: `${index * 0.1}s` }}
                  >
                    <SessionCard
                      session={session}
                      onShowQR={openQRModal}
                      onSendMessage={openSendModal}
                      onLogout={logoutSession}
                      onDelete={deleteSession}
                      onEdit={openEditModal}
                      isSelected={isSelected}
                      onSelect={handleSelectSession}
                      showFilters={showFilters}
                    />
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>

      {/* Modals */}
      <CreateSessionModal
        isOpen={createModal}
        onClose={() => setCreateModal(false)}
        onSuccess={handleCreateSuccess}
      />

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
          onSuccess={() => showSuccess("Message sent")}
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

      {/* Delete Confirmation Dialogs */}
      <DeleteConfirmDialog
        isOpen={deleteModal.show}
        onClose={() => setDeleteModal({ show: false, sessionId: null, sessionName: null })}
        onConfirm={confirmDeleteSession}
        title="Delete Session"
        message={`Are you sure you want to delete "${deleteModal.sessionName}"? This will permanently remove the session and all its data.`}
        confirmText="Delete Session"
        loading={deleteLoading}
      />

      <DeleteConfirmDialog
        isOpen={bulkDeleteModal.show}
        onClose={() => setBulkDeleteModal({ show: false, count: 0 })}
        onConfirm={confirmBulkDelete}
        title="Delete Multiple Sessions"
        message={`Are you sure you want to delete ${bulkDeleteModal.count} selected sessions? This will permanently remove all selected sessions and their data.`}
        confirmText={`Delete ${bulkDeleteModal.count} Sessions`}
        loading={bulkActionLoading}
      />
    </div>
  );
};

export default Dashboard;
