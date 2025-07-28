const SessionCard = ({ session, onShowQR, onSendMessage, onLogout, onDelete, onEdit }) => {
  const isConnected = (session.status === 'Connected' || session.connected) && session.logged_in;
  const isLoggedIn = session.logged_in;

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden hover:shadow-md transition-shadow duration-200">
      {/* Session Header */}
      <div className={`p-4 text-white ${
        isConnected 
          ? 'bg-green-600' 
          : 'bg-gray-600'
      }`}>
        <div className="flex items-center justify-between mb-3">
          <div className="flex-1 min-w-0">
            <div className="flex items-center mb-1">
              <div className={`w-2 h-2 rounded-full mr-2 ${
                isConnected ? 'bg-green-300' : 'bg-red-300'
              }`}></div>
              <h3 className="font-semibold text-lg truncate">{session.name || 'Unnamed Session'}</h3>
            </div>
            <div className="space-y-0.5">
              <p className="text-xs opacity-90 truncate">
                ID: {session.phone?.split('@')[0] || session.phone}
              </p>
              {session.actual_phone && (
                <p className="text-xs opacity-90 truncate">
                  Phone: {session.actual_phone?.split('@')[0] || session.actual_phone}
                </p>
              )}
            </div>
          </div>
          
          <div className="flex flex-col items-end space-y-1 ml-3">
            <div className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-white bg-opacity-20 text-white">
              <div className={`w-1.5 h-1.5 rounded-full mr-1.5 ${
                isConnected ? 'bg-green-300' : 'bg-red-300'
              }`}></div>
              {isConnected ? 'Connected' : 'Disconnected'}
            </div>
            
            <div className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
              isLoggedIn 
                ? 'bg-white bg-opacity-20 text-white' 
                : 'bg-white bg-opacity-10 text-white opacity-75'
            }`}>
              <i className={`fas ${
                isLoggedIn ? 'fa-check' : 'fa-times'
              } mr-1 text-xs`}></i>
              {isLoggedIn ? 'Auth' : 'No Auth'}
            </div>
          </div>
        </div>
      </div>

      {/* Session Actions */}
      <div className="p-4">
        <div className="grid grid-cols-2 gap-2 mb-2">
          {!isLoggedIn ? (
            <button
              onClick={() => onShowQR(session)}
              className="col-span-2 bg-blue-600 hover:bg-blue-700 text-white px-3 py-2 rounded-lg font-medium transition-colors duration-200 flex items-center justify-center text-sm"
            >
              <i className="fas fa-qrcode mr-2"></i>Show QR Code
            </button>
          ) : (
            <>
              <button
                onClick={() => onSendMessage(session)}
                className="bg-green-600 hover:bg-green-700 text-white px-3 py-2 rounded-lg font-medium transition-colors duration-200 flex items-center justify-center text-sm"
              >
                <i className="fas fa-paper-plane mr-1.5"></i>Send
              </button>
              <button
                onClick={() => onLogout(session.id)}
                className="bg-red-600 hover:bg-red-700 text-white px-3 py-2 rounded-lg font-medium transition-colors duration-200 flex items-center justify-center text-sm"
              >
                <i className="fas fa-sign-out-alt mr-1.5"></i>Logout
              </button>
            </>
          )}
        </div>
        
        {/* Secondary Actions */}
        <div className="grid grid-cols-2 gap-2">
          <button
            onClick={() => onEdit(session)}
            className="bg-gray-100 hover:bg-gray-200 text-gray-700 px-3 py-1.5 rounded-lg font-medium transition-colors duration-200 flex items-center justify-center text-sm"
          >
            <i className="fas fa-edit mr-1.5"></i>Edit
          </button>
          <button
            onClick={() => onDelete(session.id)}
            className="bg-red-100 hover:bg-red-200 text-red-700 px-3 py-1.5 rounded-lg font-medium transition-colors duration-200 flex items-center justify-center text-sm"
          >
            <i className="fas fa-trash mr-1.5"></i>Delete
          </button>
        </div>
      </div>
    </div>
  );
};

export default SessionCard;