import { useState, useEffect, useMemo, useCallback } from "react";
import axios from "axios";
import SessionCard from "../components/SessionCard";
import QRModal from "../components/QRModal";
import SendMessageModal from "../components/SendMessageModal";
import EditSessionModal from "../components/EditSessionModal";
import CreateSessionModal from "../components/CreateSessionModal";
import { useAuth } from "../contexts/AuthContext";

const Dashboard = () => {
  const { user } = useAuth();
  const [sessions, setSessions] = useState([]);
  const [message, setMessage] = useState("");
  const [messageType, setMessageType] = useState("");
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
      showMessage("Failed to load sessions", "error");
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
    showMessage(message, "success");
    await loadSessions();
  };

  const deleteSession = async (id) => {
    if (!window.confirm("Delete this session?")) return;
    try {
      await axios.delete(`/api/sessions/${id}`);
      showMessage("Session deleted", "success");
      await loadSessions();
    } catch (error) {
      showMessage("Failed to delete session", "error");
    }
  };

  const logoutSession = async (id) => {
    try {
      await axios.post(`/api/sessions/${id}/logout`);
      showMessage("Session logged out", "success");
    } catch (error) {
      showMessage("Logout completed", "warning");
    } finally {
      await loadSessions();
    }
  };

  const showMessage = (text, type) => {
    setMessage(text);
    setMessageType(type);
    setTimeout(() => {
      setMessage("");
      setMessageType("");
    }, 4000);
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
      showMessage(`Deleted ${selectedSessions.length} sessions`, "success");
      setSelectedSessions([]);
      await loadSessions();
    } catch (error) {
      showMessage("Bulk delete failed", "error");
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
    showMessage("Session updated", "success");
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
        <div className="grid-stats mb-8">
          <div className="card border border-gray-200">
            <div className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-overline mb-1 text-gray-600">
                    Total Sessions
                  </div>
                  <div className="text-title text-gray-900">{stats.total}</div>
                  <div className="text-caption mt-1 text-gray-500">
                    All registered
                  </div>
                </div>
                <div className="w-8 h-8 bg-gray-50 border border-gray-200 rounded-lg flex items-center justify-center">
                  <svg
                    className="w-4 h-4 text-gray-500"
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
            <div className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-overline mb-1 text-success-600">
                    Active Sessions
                  </div>
                  <div className="text-title text-success-800">
                    {stats.connected}
                  </div>
                  <div className="text-caption mt-1 text-success-600">
                    Connected & ready
                  </div>
                </div>
                <div className="w-8 h-8 bg-success-50 border border-success-200 rounded-lg flex items-center justify-center">
                  <svg
                    className="w-4 h-4 text-success-600"
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
            <div className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-overline mb-1 text-error-600">
                    Offline Sessions
                  </div>
                  <div className="text-title text-error-800">
                    {stats.offline}
                  </div>
                  <div className="text-caption mt-1 text-error-600">
                    Disconnected
                  </div>
                </div>
                <div className="w-8 h-8 bg-error-50 border border-error-200 rounded-lg flex items-center justify-center">
                  <svg
                    className="w-4 h-4 text-error-600"
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

        {/* CRM Message Alert */}
        {message && (
          <div
            className={`card-elevated mb-6 border-l-4 ${
              messageType === "error"
                ? "border-l-error"
                : messageType === "warning"
                ? "border-l-warning"
                : "border-l-success"
            }`}
          >
            <div className="card-body">
              <div className="flex items-center gap-3">
                <div
                  className={`status-dot ${
                    messageType === "error"
                      ? "status-dot-error"
                      : messageType === "warning"
                      ? "status-dot-warning"
                      : "status-dot-success"
                  }`}
                ></div>
                <span className="text-body-medium">{message}</span>
                <button
                  onClick={() => setMessage("")}
                  className="ml-auto p-1 text-gray-400 hover:text-gray-600 transition-colors"
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
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        )}

        {/* CRM Search & Filters */}
        <div className="card mb-8">
          <div className="card-body">
            <div className="flex flex-col sm:flex-row gap-4 mb-4">
              <div className="flex-1">
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <svg
                      className="w-5 h-5 text-gray-400"
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
                    placeholder="Search sessions by name, phone, or ID..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="input pl-10"
                  />
                </div>
              </div>
              <button
                onClick={() => setShowFilters(!showFilters)}
                className={`btn ${
                  showFilters ? "btn-primary" : "btn-secondary"
                }`}
              >
                <svg
                  className="w-4 h-4 mr-2"
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
                Filters {showFilters ? "& Actions" : ""}
              </button>
            </div>

            {/* Advanced Filters */}
            {showFilters && (
              <div className="animate-fade-in pt-6 border-t border-gray-200">
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
                  <div>
                    <label className="input-label">Status Filter</label>
                    <select
                      value={statusFilter}
                      onChange={(e) => setStatusFilter(e.target.value)}
                      className="select"
                    >
                      <option value="all">All Status</option>
                      <option value="connected">Connected</option>
                      <option value="offline">Offline</option>
                    </select>
                  </div>

                  <div className="sm:col-span-1 lg:col-span-3">
                    {filteredSessions.length > 0 && (
                      <div className="flex items-center gap-3 mt-6">
                        <span className="text-body-medium text-gray-700">
                          Bulk Actions:
                        </span>
                        <button
                          onClick={handleSelectAll}
                          className="btn btn-sm btn-secondary"
                        >
                          {selectedSessions.length === filteredSessions.length
                            ? "Deselect All"
                            : "Select All"}
                        </button>

                        {selectedSessions.length > 0 && (
                          <button
                            onClick={handleBulkDelete}
                            disabled={bulkActionLoading}
                            className="btn btn-sm btn-danger"
                          >
                            {bulkActionLoading ? (
                              <>
                                <div className="loading-spinner mr-2"></div>
                                Deleting...
                              </>
                            ) : (
                              <>
                                <svg
                                  className="w-4 h-4 mr-1"
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
                                Delete ({selectedSessions.length})
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
            <div className="grid-auto">
              {filteredSessions.map((session) => {
                const isConnected =
                  (session.status === "Connected" || session.connected) &&
                  session.logged_in;
                const isSelected = selectedSessions.includes(session.id);

                return (
                  <div key={session.id} className="animate-fade-in">
                    <div
                      className={`card-elevated hover:shadow-xl transition-all duration-300 group relative overflow-hidden ${
                        isSelected
                          ? "ring-2 ring-primary-500 border-primary-200 bg-primary-25"
                          : "hover:border-primary-200"
                      }`}
                    >
                      {/* Status Indicator Bar */}
                      <div
                        className={`absolute top-0 left-0 right-0 h-1 ${
                          isConnected
                            ? "bg-gradient-to-r from-success-500 to-success-400"
                            : "bg-gradient-to-r from-error-500 to-error-400"
                        }`}
                      ></div>

                      {/* Selection Checkbox */}
                      {showFilters && (
                        <div className="absolute top-4 right-4 z-10">
                          <div className="relative">
                            <input
                              type="checkbox"
                              checked={isSelected}
                              onChange={() => handleSelectSession(session.id)}
                              className="w-5 h-5 text-primary-600 border-2 border-gray-300 rounded-lg focus:ring-primary-500 focus:ring-offset-2 transition-colors"
                            />
                            {isSelected && (
                              <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                                <svg
                                  className="w-3 h-3 text-white"
                                  fill="currentColor"
                                  viewBox="0 0 20 20"
                                >
                                  <path
                                    fillRule="evenodd"
                                    d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                                    clipRule="evenodd"
                                  />
                                </svg>
                              </div>
                            )}
                          </div>
                        </div>
                      )}

                      <div className="p-6">
                        {/* Header Section */}
                        <div className="flex items-start gap-4 mb-6">
                          {/* WhatsApp Icon with Status */}
                          <div className="relative flex-shrink-0">
                            <div
                              className={`w-14 h-14 rounded-2xl flex bg-green-500 items-center justify-center shadow-md ${
                                isConnected
                                  ? "bg-gradient-to-br from-success-500 to-success-600"
                                  : "bg-gradient-to-br from-gray-400 to-gray-500"
                              }`}
                            >
                              <svg
                                className="w-8 h-8 text-white"
                                fill="currentColor"
                                viewBox="0 0 24 24"
                              >
                                <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893A11.821 11.821 0 0020.885 3.690z" />
                              </svg>
                            </div>
                            {/* Pulse Effect for Connected Sessions */}
                            {isConnected && (
                              <div className="absolute inset-0 rounded-2xl bg-success-500 opacity-25 animate-pulse"></div>
                            )}
                          </div>

                          {/* Session Info */}
                          <div className="flex-1 min-w-0">
                            <div className="flex items-start justify-between">
                              <div className="flex-1 min-w-0 flex-nowrap">
                                <h3 className="text-subtitle truncate mb-1 group-hover:text-primary-700 transition-colors">
                                  {session.name || `Session ${session.id}`}
                                </h3>
                                <p className="text-secondary mb-2 flex items-center gap-2">
                                  <svg
                                    className="w-4 h-4 flex-shrink-0"
                                    fill="none"
                                    stroke="currentColor"
                                    viewBox="0 0 24 24"
                                  >
                                    <path
                                      strokeLinecap="round"
                                      strokeLinejoin="round"
                                      strokeWidth={2}
                                      d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"
                                    />
                                  </svg>
                                  <span className="truncate">
                                    {session.phone || "No phone configured"}
                                  </span>
                                </p>

                                {/* Enhanced Status Badge */}
                                <div className="flex items-center gap-2 mb-1 flex-nowrap w-full">
                                  <div
                                    className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-full text-sm font-medium ${
                                      isConnected
                                        ? "bg-success-50 text-success-700 border border-success-200"
                                        : "bg-error-50 text-error-700 border border-error-200"
                                    }`}
                                  >
                                    <div
                                      className={` rounded-full text-nowrap ${
                                        isConnected
                                          ? "bg-success-500"
                                          : "bg-error-500"
                                      } ${isConnected ? "animate-pulse" : ""}`}
                                    ></div>
                                    {isConnected ? "Active" : "Disconnected"}
                                  </div>

                                </div>
                              </div>

                              {/* Session ID Badge */}
                              <div className="text-right flex-shrink-0 ml-4">
                                <div className="px-2 py-1 bg-gray-100 rounded-lg">
                                  <div className="text-caption text-gray-500 mb-0.5"></div>
                                  <div className="text-body-medium text-gray-900 font-mono">
                                    {session.id}
                                  </div>
                                </div>
                              </div>
                            </div>
                          </div>
                        </div>

                        {/* Action Buttons */}
                        <div className="flex gap-3">
                          {!isConnected ? (
                            <button
                              onClick={() => openQRModal(session)}
                              className="btn btn-primary flex-1 group/btn"
                            >
                              <svg
                                className="w-4 h-4 mr-2 group-hover/btn:scale-110 transition-transform"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                              >
                                <path
                                  strokeLinecap="round"
                                  strokeLinejoin="round"
                                  strokeWidth={2}
                                  d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z"
                                />
                              </svg>
                              Connect Session
                            </button>
                          ) : (
                            <button
                              onClick={() => openSendModal(session)}
                              className="btn btn-success flex-1 group/btn"
                            >
                              <svg
                                className="w-4 h-4 mr-2 group-hover/btn:scale-110 transition-transform"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                              >
                                <path
                                  strokeLinecap="round"
                                  strokeLinejoin="round"
                                  strokeWidth={2}
                                  d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
                                />
                              </svg>
                              Send Message
                            </button>
                          )}

                          <div className="flex gap-2">
                            <button
                              onClick={() => openEditModal(session)}
                              className="btn btn-secondary group/btn"
                              title="Edit Session"
                            >
                              <svg
                                className="w-4 h-4 group-hover/btn:scale-110 transition-transform"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                              >
                                <path
                                  strokeLinecap="round"
                                  strokeLinejoin="round"
                                  strokeWidth={2}
                                  d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                                />
                              </svg>
                            </button>

                            <button
                              onClick={() => deleteSession(session.id)}
                              className="btn btn-danger group/btn"
                              title="Delete Session"
                            >
                              <svg
                                className="w-4 h-4 group-hover/btn:scale-110 transition-transform"
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
                            </button>
                          </div>
                        </div>
                      </div>

                      {/* Hover Overlay Effect */}
                      <div className="absolute inset-0 bg-gradient-to-br from-primary-500/5 to-primary-600/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none"></div>
                    </div>
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
          onSuccess={() => showMessage("Message sent", "success")}
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
