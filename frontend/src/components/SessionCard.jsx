import { Webhook, WebhookOff } from "lucide-react";

const SessionCard = ({
  session,
  onShowQR,
  onSendMessage,
  onLogout,
  onDelete,
  onEdit,
  isSelected = false,
  onSelect = null,
  showFilters = false,
}) => {
  const isConnected =
    (session.status === "Connected" || session.connected) && session.logged_in;
  const isLoggedIn = session.logged_in;

  return (
    <div
      className={`bg-white rounded-lg border shadow-sm hover:shadow-md transition-all duration-300 overflow-hidden relative animate-fade-in group ${
        isSelected
          ? "border-blue-500 bg-blue-50 ring-2 ring-blue-200"
          : "border-gray-200"
      }`}
    >
      {/* Top Corner Actions */}
      <div className="absolute top-3 right-3 z-10 flex items-center gap-2">
        {/* Edit Button */}
        <button
          onClick={() => onEdit(session)}
          className="w-7 h-7 bg-white/80 backdrop-blur-sm hover:bg-white border border-gray-200 rounded-full shadow-sm hover:shadow-md transition-all duration-200 flex items-center justify-center opacity-0 group-hover:opacity-100"
          title="Edit Session"
        >
          <i className="fas fa-edit text-xs text-blue-600"></i>
        </button>

        {/* Selection Checkbox */}
        {showFilters && onSelect && (
          <input
            type="checkbox"
            checked={isSelected}
            onChange={() => onSelect(session.id)}
            className="w-4 h-4 text-blue-600 bg-white border-2 border-gray-300 rounded focus:ring-blue-500 focus:ring-offset-2 transition-all duration-200 cursor-pointer"
          />
        )}

        {/* Delete Button */}
        <button
          onClick={() => onDelete(session.id)}
          className="w-7 h-7 bg-white/80 backdrop-blur-sm hover:bg-red-50 border border-gray-200 hover:border-red-300 rounded-full shadow-sm hover:shadow-md transition-all duration-200 flex items-center justify-center opacity-0 group-hover:opacity-100"
          title="Delete Session"
        >
          <i className="fas fa-trash text-xs text-red-500"></i>
        </button>
      </div>

      {/* Session Header */}
      <div className="p-5 pb-4">
        <div className="flex items-center mb-4">
          <div className="relative">
            <div className="flex items-center justify-center w-12 h-12 bg-gradient-to-br from-emerald-500 to-green-600 rounded-xl shadow-md">
              <i className="fab fa-whatsapp text-white text-xl"></i>
            </div>
            <div
              className={`absolute -bottom-1 -right-1 w-4 h-4 rounded-full border-2 border-white shadow-md ${
                isConnected ? "bg-green-500" : "bg-red-500"
              }`}
            ></div>
          </div>
          <div className="ml-4 flex-1 min-w-0">
            <h3 className="text-base font-semibold text-gray-900 mb-1 truncate">
              {session.name || "Unnamed Session"}
            </h3>
            {session.actual_phone ? (
              <div className="flex items-center text-sm font-medium text-gray-700">
                <i className="fas fa-phone mr-2 text-green-600"></i>
                <span className="truncate">
                  {session.actual_phone?.split("@")[0] || "Not connected"}
                </span>
              </div>
            ) : (
              <div className="flex items-center text-sm text-gray-500">
                <i className="fas fa-phone-slash mr-2 text-red-400"></i>
                <span>Not connected</span>
              </div>
            )}
            <div className="flex items-center text-xs text-gray-500 mt-1">
              <i className="fas fa-fingerprint mr-2 text-blue-500"></i>
              <span className="font-mono">ID: {session.id}</span>
            </div>
          </div>
        </div>

        {/* Status Badges */}
        <div className="flex flex-wrap items-center gap-2">
          <div
            className={`inline-flex items-center px-3 py-1.5 rounded-full text-xs font-medium border ${
              isConnected
                ? "bg-green-50 text-green-700 border-green-200"
                : "bg-red-50 text-red-700 border-red-200"
            }`}
          >
            <div
              className={`w-2 h-2 rounded-full mr-2 ${
                isConnected ? "bg-green-500" : "bg-red-500"
              }`}
            ></div>
            {isConnected ? "Online" : "Offline"}
          </div>

          <div
            className={`inline-flex items-center px-3 py-1.5 rounded-full text-xs font-medium border ${
              isLoggedIn
                ? "bg-blue-50 text-blue-700 border-blue-200"
                : "bg-orange-50 text-orange-700 border-orange-200"
            }`}
          >
            <i
              className={`fas ${
                isLoggedIn ? "fa-shield-check" : "fa-exclamation-triangle"
              } mr-2 text-xs`}
            ></i>
            {isLoggedIn ? "Authenticated" : "Not Authenticated"}
          </div>

          {/* Webhook Status */}
          <div
            className={`inline-flex items-center  p-1.5   rounded-full text-xs font-medium border ${
              session.webhook_url
                ? "bg-purple-50 text-purple-700 border-purple-200"
                : "bg-gray-50 text-gray-600 border-gray-200"
            }`}
          >
            {session.webhook_url ? (
              <Webhook className="w-3 h-3" title="Webhook Active" />
            ) : (
              <WebhookOff className="w-3 h-3" title="No Webhook" />
            )}
          </div>
        </div>
      </div>

      {/* Session Actions */}
      <div className="px-5 pb-5">
        <div className="border-t border-gray-100 pt-4">
          {!isLoggedIn ? (
            <button
              onClick={() => onShowQR(session)}
              className="w-full bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-medium py-3 px-4 rounded-xl shadow-md hover:shadow-lg transform hover:scale-[1.02] transition-all duration-200 flex items-center justify-center"
            >
              <i className="fas fa-qrcode mr-3 text-lg"></i>
              <span>Connect with QR Code</span>
            </button>
          ) : (
            <div className="flex items-center gap-3">
              <button
                onClick={() => onSendMessage(session)}
                className="flex-1 bg-gradient-to-r from-green-600 to-green-700 hover:from-green-700 hover:to-green-800 text-white font-medium py-3 px-4 rounded-xl shadow-md hover:shadow-lg transform hover:scale-[1.02] transition-all duration-200 flex items-center justify-center"
                title="Send Message"
              >
                <i className="fas fa-paper-plane mr-2"></i>
                <span>Send Message</span>
              </button>
              <button
                onClick={() => onLogout(session.id)}
                className="w-12 h-12 bg-orange-500 hover:bg-orange-600 text-white rounded-xl shadow-md hover:shadow-lg transform hover:scale-[1.02] transition-all duration-200 flex items-center justify-center"
                title="Logout Session"
              >
                <i className="fas fa-sign-out-alt"></i>
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default SessionCard;
