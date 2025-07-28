const SessionCard = ({ session, onShowQR, onSendMessage, onLogout, onDelete, onEdit }) => {
  const isConnected = (session.status === 'Connected' || session.connected) && session.logged_in;
  const isLoggedIn = session.logged_in;

  return (
    <div className="bg-white rounded-lg border border-gray-200 shadow-sm hover:shadow-md transition-all duration-200 overflow-hidden">
      {/* Session Header */}
      <div className="p-5 border-b border-gray-100">
        <div className="flex items-start justify-between mb-4">
          <div className="flex items-center">
            <div className="relative">
              <div className="flex items-center justify-center w-11 h-11 bg-blue-500 rounded-lg shadow-sm">
                <i className="fab fa-whatsapp text-white text-lg"></i>
              </div>
              <div className={`absolute -bottom-1 -right-1 w-3 h-3 rounded-full border-2 border-white ${
                isConnected ? 'bg-green-500' : 'bg-gray-400'
              }`}></div>
            </div>
            <div className="ml-3">
              <h3 className="font-semibold text-gray-900 text-base mb-1">{session.name || 'Unnamed Session'}</h3>
              <div className="flex items-center text-sm text-gray-500">
                <i className="fas fa-mobile-alt mr-2"></i>
                <span>{session.phone?.split('@')[0] || session.phone}</span>
              </div>
              {session.actual_phone && (
                <div className="flex items-center text-sm text-gray-500 mt-1">
                  <i className="fas fa-phone mr-2"></i>
                  <span>{session.actual_phone?.split('@')[0] || session.actual_phone}</span>
                </div>
              )}
            </div>
          </div>
        </div>
        
        {/* Status Badges */}
        <div className="flex items-center gap-2">
          <div className={`inline-flex items-center px-2.5 py-1 rounded text-xs font-medium ${
            isConnected 
              ? 'bg-green-100 text-green-700' 
              : 'bg-gray-100 text-gray-600'
          }`}>
            <div className={`w-2 h-2 rounded-full mr-2 ${
              isConnected ? 'bg-green-500' : 'bg-gray-400'
            }`}></div>
            {isConnected ? 'Online' : 'Offline'}
          </div>
          
          <div className={`inline-flex items-center px-2.5 py-1 rounded text-xs font-medium ${
            isLoggedIn 
              ? 'bg-blue-100 text-blue-700' 
              : 'bg-gray-100 text-gray-600'
          }`}>
            <i className={`fas ${isLoggedIn ? 'fa-shield-check' : 'fa-exclamation-triangle'} mr-2`}></i>
            {isLoggedIn ? 'Authenticated' : 'Not Authenticated'}
          </div>
        </div>
      </div>

      {/* Session Actions */}
      <div className="p-5">
        {/* Primary Actions */}
        <div className="space-y-3 mb-4">
          {!isLoggedIn ? (
            <button
              onClick={() => onShowQR(session)}
              className="w-full bg-blue-500 hover:bg-blue-600 text-white py-2.5 px-4 rounded-lg font-medium flex items-center justify-center transition-colors duration-200"
            >
              <i className="fas fa-qrcode mr-2"></i>
              Connect QR Code
            </button>
          ) : (
            <div className="grid grid-cols-2 gap-3">
              <button
                onClick={() => onSendMessage(session)}
                className="bg-blue-500 hover:bg-blue-600 text-white py-2.5 px-3 rounded-lg font-medium transition-colors duration-200 flex items-center justify-center text-sm"
              >
                <i className="fas fa-paper-plane mr-2"></i>
                Send
              </button>
              <button
                onClick={() => onLogout(session.id)}
                className="bg-purple-500 hover:bg-purple-600 text-white py-2.5 px-3 rounded-lg font-medium transition-colors duration-200 flex items-center justify-center text-sm"
              >
                <i className="fas fa-sign-out-alt mr-2"></i>
                Logout
              </button>
            </div>
          )}
        </div>
        
        {/* Secondary Actions */}
        <div className="grid grid-cols-2 gap-3">
          <button
            onClick={() => onEdit(session)}
            className="bg-gray-100 hover:bg-gray-200 text-gray-700 py-2 px-3 rounded-lg font-medium transition-colors duration-200 flex items-center justify-center text-sm"
          >
            <i className="fas fa-edit mr-2"></i>
            Edit
          </button>
          <button
            onClick={() => onDelete(session.id)}
            className="bg-gray-100 hover:bg-gray-200 text-gray-600 hover:text-gray-700 py-2 px-3 rounded-lg font-medium transition-colors duration-200 flex items-center justify-center text-sm"
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