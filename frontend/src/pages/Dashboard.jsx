import { useState, useEffect, useMemo, useCallback } from "react";
import axios from "axios";
import SessionCard from "../components/SessionCard";
import QRModal from "../components/QRModal";
import SendMessageModal from "../components/SendMessageModal";
import EditSessionModal from "../components/EditSessionModal";
import CreateSessionModal from "../components/CreateSessionModal";
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
    if (!window.confirm("Delete this session?")) return;
    try {
      await axios.delete(`/api/sessions/${id}`);
      showSuccess("Session deleted");
      await loadSessions();
    } catch (error) {
      showError("Failed to delete session");
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
    if (!window.confirm(`Delete ${selectedSessions.length} selected sessions?`))
      return;

    setBulkActionLoading(true);
    try {
      await Promise.all(
        selectedSessions.map((id) =>
          axios.delete(`/api/sessions/${id}`).catch(() => null)
        )
      );
      showSuccess(`Deleted ${selectedSessions.length} sessions`);
      setSelectedSessions([]);
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
        {/* CRM Header */}
        <div className="mb-8">
          <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-6">
            <div>
              <div className="flex items-center gap-3 mb-2">
                <h1 className="text-display">Session Management</h1>
              </div>
              <p className="text-secondary">
                Monitor and manage your WhatsApp business connections
              </p>
            </div>
            <button
              onClick={() => setCreateModal(true)}
              className="btn btn-primary btn-lg"
            >
              <svg
                className="w-5 h-5 mr-2"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 4v16m8-8H4"
                />
              </svg>
              New Session
            </button>
          </div>
        </div>

        {/* Improved Stats Dashboard */}
        <div className="grid-stats mb-6">
          <div className="card border border-gray-200">
            <div className="p-3">
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-overline mb-1 text-gray-600">
                    Total Sessions
                  </div>
                  <div className="text-xl font-semibold text-gray-900">{stats.total}</div>
                  <div className="text-xs mt-1 text-gray-500">
                    All registered
                  </div>
                </div>
                <div className="w-6 h-6 bg-gray-50 border border-gray-200 rounded-md flex items-center justify-center">
                  <svg
                    className="w-3 h-3 text-gray-500"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
                    />
                  </svg>
                </div>
              </div>
            </div>
          </div>

          <div className="card border border-success-200 bg-success-25">
            <div className="p-3">
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-overline mb-1 text-success-600">
                    Active Sessions
                  </div>
                  <div className="text-xl font-semibold text-success-800">
                    {stats.connected}
                  </div>
                  <div className="text-xs mt-1 text-success-600">
                    Connected & ready
                  </div>
                </div>
                <div className="w-6 h-6 bg-success-50 border border-success-200 rounded-md flex items-center justify-center">
                  <svg
                    className="w-3 h-3 text-success-600"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                </div>
              </div>
            </div>
          </div>

          <div className="card border border-error-200 bg-error-25">
            <div className="p-3">
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-overline mb-1 text-error-600">
                    Offline Sessions
                  </div>
                  <div className="text-xl font-semibold text-error-800">
                    {stats.offline}
                  </div>
                  <div className="text-xs mt-1 text-error-600">
                    Disconnected
                  </div>
                </div>
                <div className="w-6 h-6 bg-error-50 border border-error-200 rounded-md flex items-center justify-center">
                  <svg
                    className="w-3 h-3 text-error-600"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                </div>
              </div>
            </div>
          </div>
        </div>


        {/* CRM Search & Filters */}
        <div className="card mb-6">
          <div className="p-4">
            <div className="flex flex-col sm:flex-row gap-3 mb-3">
              <div className="flex-1">
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <svg
                      className="w-4 h-4 text-gray-400"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                      />
                    </svg>
                  </div>
                  <input
                    type="text"
                    placeholder="Search sessions..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="w-full py-2.5 pl-9 pr-4 text-sm border border-gray-300 rounded-lg bg-white focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200 placeholder-gray-500"
                  />
                </div>
              </div>
              <button
                onClick={() => setShowFilters(!showFilters)}
                className={`px-4 py-2.5 text-sm font-medium rounded-lg border transition-all duration-200 flex items-center gap-2 ${
                  showFilters 
                    ? "bg-blue-600 text-white border-blue-600 shadow-md" 
                    : "bg-white text-gray-700 border-gray-300 hover:bg-gray-50 hover:border-gray-400"
                }`}
              >
                <svg
                  className="w-4 h-4"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.707A1 1 0 013 7V4z"
                  />
                </svg>
                <span className="hidden sm:inline">Filters</span>
                {filteredSessions.length !== sessions.length && (
                  <span className="inline-flex items-center justify-center w-5 h-5 text-xs font-bold text-white bg-red-500 rounded-full">
                    {filteredSessions.length}
                  </span>
                )}
              </button>
            </div>

            {/* Advanced Filters */}
            {showFilters && (
              <div className="animate-fade-in pt-4 border-t border-gray-200">
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 mb-4">
                  <div>
                    <label className="block text-xs font-medium text-gray-700 mb-1">Status Filter</label>
                    <select
                      value={statusFilter}
                      onChange={(e) => setStatusFilter(e.target.value)}
                      className="w-full py-2 px-3 text-sm border border-gray-300 rounded-lg bg-white focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                    >
                      <option value="all">All Status</option>
                      <option value="connected">Connected</option>
                      <option value="offline">Offline</option>
                    </select>
                  </div>

                  <div className="sm:col-span-1 lg:col-span-3">
                    {filteredSessions.length > 0 && (
                      <div className="flex items-center gap-2 mt-5">
                        <span className="text-xs font-medium text-gray-600">
                          Bulk Actions:
                        </span>
                        <button
                          onClick={handleSelectAll}
                          className="px-3 py-1.5 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 transition-colors duration-200"
                        >
                          {selectedSessions.length === filteredSessions.length
                            ? "Deselect All"
                            : "Select All"}
                        </button>

                        {selectedSessions.length > 0 && (
                          <button
                            onClick={handleBulkDelete}
                            disabled={bulkActionLoading}
                            className="px-3 py-1.5 text-xs font-medium text-white bg-red-600 hover:bg-red-700 disabled:bg-red-400 rounded-md transition-colors duration-200 flex items-center gap-1"
                          >
                            {bulkActionLoading ? (
                              <>
                                <div className="w-3 h-3 border border-white/30 border-t-white rounded-full animate-spin"></div>
                                <span>Deleting...</span>
                              </>
                            ) : (
                              <>
                                <svg
                                  className="w-3 h-3"
                                  fill="none"
                                  stroke="currentColor"
                                  viewBox="0 0 24 24"
                                >
                                  <path
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    strokeWidth={2}
                                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                                  />
                                </svg>
                                <span>Delete ({selectedSessions.length})</span>
                              </>
                            )}
                          </button>
                        )}
                      </div>
                    )}
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* CRM Sessions Grid */}
        <div className="mb-8">
          {filteredSessions.length === 0 ? (
            <div className="card-elevated">
              <div className="card-body text-center py-12">
                <div className="w-20 h-20 bg-gray-100 rounded-2xl flex items-center justify-center mx-auto mb-6">
                  <svg
                    className="w-10 h-10 text-gray-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                    />
                  </svg>
                </div>
                <h3 className="text-title mb-3">
                  {sessions.length === 0
                    ? "No Sessions Created"
                    : "No Matching Sessions"}
                </h3>
                <p className="text-secondary mb-8 max-w-md mx-auto">
                  {sessions.length === 0
                    ? "Get started by creating your first WhatsApp business session to connect with customers."
                    : "No sessions match your current search criteria. Try adjusting your filters or search terms."}
                </p>
                {sessions.length === 0 && (
                  <button
                    onClick={() => setCreateModal(true)}
                    className="btn btn-primary btn-lg"
                  >
                    <svg
                      className="w-5 h-5 mr-2"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M12 4v16m8-8H4"
                      />
                    </svg>
                    Create Your First Session
                  </button>
                )}
              </div>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {filteredSessions.map((session) => {
                const isSelected = selectedSessions.includes(session.id);

                return (
                  <div key={session.id} className="animate-fade-in">
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
    </div>
  );
};

export default Dashboard;
