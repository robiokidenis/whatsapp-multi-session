const SessionCard = ({ 
  session, 
  onShowQR, 
  onSendMessage, 
  onLogout, 
  onDelete, 
  onEdit,
  isSelected = false,
  onSelect = null,
  showFilters = false 
}) => {
  const isConnected = (session.status === 'Connected' || session.connected) && session.logged_in;
  const isLoggedIn = session.logged_in;

  return (
    <div className={`bg-white rounded-xl border shadow-lg hover:shadow-xl transition-all duration-300 overflow-hidden relative animate-fade-in ${
      isSelected ? 'border-blue-500 bg-blue-50 ring-2 ring-blue-200' : 'border-gray-200'
    }`}>
      {/* Selection Checkbox */}
      {showFilters && onSelect && (
        <div className="absolute top-5 right-5 z-10">
          <input
            type="checkbox"
            checked={isSelected}
            onChange={() => onSelect(session.id)}
            className="w-5 h-5 text-blue-600 bg-white border-2 border-gray-300 rounded focus:ring-blue-500 focus:ring-offset-2 transition-all duration-200 cursor-pointer"
          />
        </div>
      )}
      
      {/* Session Header */}
      <div className="p-6 border-b border-gray-100 bg-gradient-to-r from-gray-50 to-white">
        <div className="flex items-start justify-between mb-4">
          <div className="flex items-center">
            <div className="relative">
              <div className="flex items-center justify-center w-14 h-14 bg-gradient-to-br from-emerald-500 to-green-600 rounded-xl shadow-md hover:shadow-lg transition-shadow duration-200">
                <i className="fab fa-whatsapp text-white text-2xl"></i>
              </div>
              <div className={`absolute -bottom-1 -right-1 w-4 h-4 rounded-full border-3 border-white shadow-md ${
                isConnected ? 'bg-green-500' : 'bg-red-500'
              }`}></div>
            </div>
            <div className="ml-3">
              <h3 className="text-lg font-semibold text-gray-900 mb-2">{session.name || 'Unnamed Session'}</h3>
              {session.actual_phone ? (
                <div className="flex items-center text-sm font-medium text-gray-800">
                  <i className="fas fa-phone mr-2 text-green-600"></i>
                  <span>{session.actual_phone?.split('@')[0] || 'Not connected'}</span>
                </div>
              ) : (
                <div className="flex items-center text-sm text-gray-500">
                  <i className="fas fa-phone-slash mr-2 text-red-500"></i>
                  <span>Not connected</span>
                </div>
              )}
              <div className="flex items-center text-xs text-gray-500 mt-2">
                <i className="fas fa-fingerprint mr-2 text-blue-500"></i>
                <span className="font-mono">ID: {session.id}</span>
              </div>
            </div>
          </div>
        </div>
        
        {/* Status Badges */}
        <div className="flex items-center gap-3 mt-4">
          <div className={`inline-flex items-center px-3 py-1.5 rounded-full text-xs font-semibold border ${
            isConnected 
              ? 'bg-green-100 text-green-800 border-green-200' 
              : 'bg-red-100 text-red-800 border-red-200'
          }`}>
            <div className={`w-2 h-2 rounded-full mr-2 ${
              isConnected ? 'bg-green-500' : 'bg-red-500'
            }`}></div>
            {isConnected ? 'Online' : 'Offline'}
          </div>
          
          <div className={`inline-flex items-center px-3 py-1.5 rounded-full text-xs font-semibold border ${
            isLoggedIn 
              ? 'bg-blue-100 text-blue-800 border-blue-200' 
              : 'bg-orange-100 text-orange-800 border-orange-200'
          }`}>
            <i className={`fas ${isLoggedIn ? 'fa-shield-check' : 'fa-exclamation-triangle'} mr-2`}></i>
            {isLoggedIn ? 'Authenticated' : 'Not Authenticated'}
          </div>
        </div>
      </div>

      {/* Session Actions */}
      <div className="p-6 bg-gray-50">
        {/* Primary Actions */}
        <div className="space-y-3 mb-5">
          {!isLoggedIn ? (
            <button
              onClick={() => onShowQR(session)}
              className="w-full bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-semibold py-3 px-6 rounded-xl shadow-lg hover:shadow-xl transform hover:scale-[1.02] transition-all duration-200 flex items-center justify-center"
            >
              <i className="fas fa-qrcode mr-3 text-lg"></i>
              Connect QR Code
            </button>
          ) : (
            <div className="grid grid-cols-2 gap-4">
              <button
                onClick={() => onSendMessage(session)}
                className="bg-gradient-to-r from-green-600 to-green-700 hover:from-green-700 hover:to-green-800 text-white font-semibold py-2.5 px-4 rounded-lg shadow-md hover:shadow-lg transform hover:scale-[1.02] transition-all duration-200 flex items-center justify-center"
              >
                <i className="fas fa-paper-plane mr-2"></i>
                Send
              </button>
              <button
                onClick={() => onLogout(session.id)}
                className="bg-gradient-to-r from-orange-500 to-orange-600 hover:from-orange-600 hover:to-orange-700 text-white font-semibold py-2.5 px-4 rounded-lg shadow-md hover:shadow-lg transform hover:scale-[1.02] transition-all duration-200 flex items-center justify-center"
              >
                <i className="fas fa-sign-out-alt mr-2"></i>
                Logout
              </button>
            </div>
          )}
        </div>
        
        {/* Secondary Actions */}
        <div className="grid grid-cols-2 gap-4">
          <button
            onClick={() => onEdit(session)}
            className="bg-white hover:bg-gray-100 text-gray-700 font-medium py-2.5 px-4 rounded-lg border border-gray-300 shadow-sm hover:shadow-md transition-all duration-200 flex items-center justify-center"
          >
            <i className="fas fa-edit mr-2 text-blue-600"></i>
            Edit
          </button>
          <button
            onClick={() => onDelete(session.id)}
            className="bg-white hover:bg-red-50 text-red-600 hover:text-red-700 font-medium py-2.5 px-4 rounded-lg border border-red-300 shadow-sm hover:shadow-md transition-all duration-200 flex items-center justify-center"
          >
            <i className="fas fa-trash mr-2"></i>
            Delete
          </button>
        </div>
      </div>
    </div>
  );
};

export default SessionCard;