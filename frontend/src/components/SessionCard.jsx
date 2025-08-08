import { Webhook, WebhookOff } from "lucide-react";
import { useState } from "react";

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
  const [showActions, setShowActions] = useState(false);
  const isConnected =
    (session.status === "Connected" || session.connected) && session.logged_in;
  const isLoggedIn = session.logged_in;

  return (
    <div
      className={`group relative bg-white rounded-3xl border shadow-lg hover:shadow-2xl transition-all duration-500 overflow-hidden backdrop-blur-sm h-[300px] flex flex-col ${
        isSelected
          ? "border-primary-400 bg-gradient-to-br from-primary-50 via-white to-primary-25 ring-2 ring-primary-300 shadow-primary-200/50 transform scale-[1.02]"
          : "border-gray-200/60 hover:border-primary-200 hover:shadow-primary-100/30 hover:transform hover:scale-[1.01]"
      }`}
    >
      {/* Subtle Background Pattern */}
      <div className="absolute inset-0 bg-gradient-to-br from-transparent via-gray-50/30 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>
      {/* Clean Top Corner Actions */}
      <div className="absolute top-4 right-4 z-20 flex items-center gap-2">
        {/* Selection Checkbox */}
        {showFilters && onSelect && (
          <div className="relative">
            <input
              type="checkbox"
              checked={isSelected}
              onChange={() => onSelect(session.id)}
              className="w-5 h-5 text-primary-600 bg-white border-2 border-gray-300 rounded-lg focus:ring-primary-500 focus:ring-offset-2 transition-all duration-200 cursor-pointer hover:border-primary-400 shadow-sm"
            />
            {isSelected && (
              <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                <svg className="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                </svg>
              </div>
            )}
          </div>
        )}

        {/* Actions Dropdown */}
        <div className="relative">
          <button
            onClick={() => setShowActions(!showActions)}
            className="w-8 h-8 bg-white/90 backdrop-blur-sm hover:bg-white border border-gray-200 rounded-xl shadow-sm hover:shadow-md transition-all duration-300 flex items-center justify-center opacity-0 group-hover:opacity-100 hover:scale-110"
            title="More Actions"
          >
            <svg className="w-4 h-4 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 5v.01M12 12v.01M12 19v.01" />
            </svg>
          </button>

          {/* Dropdown Menu */}
          {showActions && (
            <div className="absolute right-0 top-full mt-2 w-48 bg-white rounded-xl shadow-xl border border-gray-200 py-2 z-30">
              <button
                onClick={() => {
                  onEdit(session);
                  setShowActions(false);
                }}
                className="w-full px-4 py-2.5 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center transition-colors"
              >
                <svg className="w-4 h-4 mr-3 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
                Edit Session
              </button>
              
              {isLoggedIn && (
                <button
                  onClick={() => {
                    onLogout(session.id);
                    setShowActions(false);
                  }}
                  className="w-full px-4 py-2.5 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center transition-colors"
                >
                  <svg className="w-4 h-4 mr-3 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                  </svg>
                  Logout Session
                </button>
              )}
              
              <div className="border-t border-gray-100 my-1"></div>
              
              <button
                onClick={() => {
                  onDelete(session.id);
                  setShowActions(false);
                }}
                className="w-full px-4 py-2.5 text-left text-sm text-red-600 hover:bg-red-50 flex items-center transition-colors"
              >
                <svg className="w-4 h-4 mr-3 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
                Delete Session
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Click outside to close dropdown */}
      {showActions && (
        <div 
          className="fixed inset-0 z-10" 
          onClick={() => setShowActions(false)}
        ></div>
      )}

      {/* Enhanced Session Header */}
      <div className="p-5 pb-3 flex-1">
        <div className="flex items-center mb-4">
          <div className="relative">
            <div className="flex items-center justify-center w-14 h-14 bg-gradient-to-br from-primary-500 via-primary-600 to-primary-700 rounded-2xl shadow-lg">
              <svg className="w-7 h-7 text-white" fill="currentColor" viewBox="0 0 24 24">
                <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.890-5.335 11.893-11.893A11.821 11.821 0 0020.885 3.488z"/>
              </svg>
            </div>
            <div
              className={`absolute -bottom-0.5 -right-0.5 w-4 h-4 rounded-full border-2 border-white shadow-md ${
                isConnected ? "bg-green-500 animate-pulse" : "bg-red-500"
              }`}
            ></div>
          </div>
          <div className="ml-5 flex-1 min-w-0">
            <h3 className="text-lg font-bold text-gray-900 mb-2 truncate">
              {session.name || "Unnamed Session"}
            </h3>
            {session.actual_phone ? (
              <div className="flex items-center text-sm font-medium text-gray-700 mb-1">
                <div className="w-2 h-2 bg-green-500 rounded-full mr-2 animate-pulse"></div>
                <span className="truncate">
                  {session.actual_phone?.split("@")[0] || "Not connected"}
                </span>
              </div>
            ) : (
              <div className="flex items-center text-sm text-gray-500 mb-1">
                <div className="w-2 h-2 bg-red-500 rounded-full mr-2"></div>
                <span>Not connected</span>
              </div>
            )}
            <div className="flex items-center text-xs text-gray-500">
              <svg className="w-3 h-3 mr-1.5 text-primary-500" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M6.625 2.655A9 9 0 0119 11a1 1 0 11-2 0 7 7 0 00-9.625-6.492 1 1 0 11-.75-1.853zM4.662 4.959A1 1 0 014.75 6.37 6 6 0 0016 11a1 1 0 11-2 0 4 4 0 00-7.438-2.11 1 1 0 01-1.9-.93zM6.5 9a1 1 0 011-1h.01a1 1 0 110 2H7.5a1 1 0 01-1-1z" clipRule="evenodd" />
              </svg>
              <span className="font-mono text-xs">#{session.id}</span>
            </div>
          </div>
        </div>

        {/* Beautiful Status Indicators */}
        <div className="flex items-center justify-between mb-4">
          {/* Auth Status */}
          <div className="flex items-center gap-3">
            {/* Authentication Indicator */}
            <div
              className={`w-10 h-10 rounded-2xl flex items-center justify-center shadow-sm border transition-all duration-300 group-hover:scale-105 ${
                isLoggedIn
                  ? "bg-gradient-to-br from-primary-50 to-blue-100 border-primary-200/60 shadow-primary-100/50"
                  : "bg-gradient-to-br from-orange-50 to-yellow-100 border-orange-200/60 shadow-orange-100/50"
              }`}
              title={isLoggedIn ? "Authenticated" : "Setup Required"}
            >
              <svg className={`w-5 h-5 ${isLoggedIn ? "text-primary-600" : "text-orange-600"}`} fill="currentColor" viewBox="0 0 20 20">
                {isLoggedIn ? (
                  <path fillRule="evenodd" d="M2.166 4.999A11.954 11.954 0 0010 1.944 11.954 11.954 0 0017.834 5c.11.65.166 1.32.166 2.001 0 5.225-3.34 9.67-8 11.317C5.34 16.67 2 12.225 2 7c0-.682.057-1.35.166-2.001zm11.541 3.708a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                ) : (
                  <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                )}
              </svg>
            </div>
          </div>
          
          {/* Feature Indicators */}
          <div className="flex items-center gap-2">
            {/* Webhook Indicator */}
            <div
              className={`w-9 h-9 rounded-xl flex items-center justify-center shadow-sm border transition-all duration-300 group-hover:scale-105 ${
                session.webhook_url
                  ? "bg-gradient-to-br from-purple-50 to-indigo-100 border-purple-200/60 shadow-purple-100/50"
                  : "bg-gradient-to-br from-gray-50 to-gray-100 border-gray-200/60"
              }`}
              title={session.webhook_url ? "Webhook Active" : "No Webhook"}
            >
              {session.webhook_url ? (
                <Webhook className="w-4 h-4 text-purple-600" />
              ) : (
                <WebhookOff className="w-4 h-4 text-gray-400" />
              )}
            </div>
            
            {/* Auto Reply Indicator */}
            <div
              className={`w-9 h-9 rounded-xl flex items-center justify-center shadow-sm border transition-all duration-300 group-hover:scale-105 ${
                session.auto_reply_text
                  ? "bg-gradient-to-br from-teal-50 to-cyan-100 border-teal-200/60 shadow-teal-100/50"
                  : "bg-gradient-to-br from-gray-50 to-gray-100 border-gray-200/60"
              }`}
              title={session.auto_reply_text ? "Auto Reply Active" : "No Auto Reply"}
            >
              <svg className={`w-4 h-4 ${session.auto_reply_text ? "text-teal-600" : "text-gray-400"}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
              </svg>
            </div>
            
            {/* Proxy Indicator */}
            <div
              className={`w-9 h-9 rounded-xl flex items-center justify-center shadow-sm border transition-all duration-300 group-hover:scale-105 ${
                session.proxy_config && session.proxy_config.enabled
                  ? "bg-gradient-to-br from-amber-50 to-orange-100 border-amber-200/60 shadow-amber-100/50"
                  : "bg-gradient-to-br from-gray-50 to-gray-100 border-gray-200/60"
              }`}
              title={
                session.proxy_config && session.proxy_config.enabled
                  ? `Proxy Active (${session.proxy_config.type?.toUpperCase()} - ${session.proxy_config.host}:${session.proxy_config.port})`
                  : "No Proxy"
              }
            >
              <svg className={`w-4 h-4 ${session.proxy_config && session.proxy_config.enabled ? "text-amber-600" : "text-gray-400"}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
              </svg>
            </div>
          </div>
        </div>
      </div>

      {/* Premium Action Button */}
      <div className="px-5 pb-5">
        <div className="border-t border-gray-200/50 pt-4">
          {!isLoggedIn ? (
            <button
              onClick={() => onShowQR(session)}
              className="relative w-full h-12 bg-gradient-to-r from-primary-500 via-primary-600 to-primary-700 hover:from-primary-600 hover:via-primary-700 hover:to-primary-800 text-white font-medium py-3 px-4 rounded-2xl shadow-lg shadow-primary-500/25 hover:shadow-xl hover:shadow-primary-500/40 transform hover:scale-[1.02] transition-all duration-400 flex items-center justify-center group overflow-hidden"
            >
              <div className="absolute inset-0 bg-gradient-to-r from-white/20 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-400"></div>
              <svg className="w-4 h-4 mr-2 group-hover:scale-110 transition-transform duration-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 16h4.01M12 8h4.01M16 4h4.01M4 4h4.01M4 8h4.01M4 12h4.01M4 16h4.01" />
              </svg>
              <span className="relative z-10 text-sm">Connect with QR</span>
            </button>
          ) : (
            <button
              onClick={() => onSendMessage(session)}
              className="relative w-full h-12 bg-gradient-to-r from-green-500 via-green-600 to-emerald-700 hover:from-green-600 hover:via-green-700 hover:to-emerald-800 text-white font-medium py-3 px-4 rounded-2xl shadow-lg shadow-green-500/25 hover:shadow-xl hover:shadow-green-500/40 transform hover:scale-[1.02] transition-all duration-400 flex items-center justify-center group overflow-hidden"
            >
              <div className="absolute inset-0 bg-gradient-to-r from-white/20 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-400"></div>
              <svg className="w-4 h-4 mr-2 group-hover:scale-110 transition-transform duration-300" fill="currentColor" viewBox="0 0 24 24">
                <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.890-5.335 11.893-11.893A11.821 11.821 0 0020.885 3.488z"/>
              </svg>
              <span className="relative z-10 text-sm">Send Message</span>
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default SessionCard;
