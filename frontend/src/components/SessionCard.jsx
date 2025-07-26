const SessionCard = ({ session, onShowQR, onSendMessage, onLogout, onDelete, onEdit }) => {
  return (
    <div className="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-lg transition duration-200">
      {/* Session Header */}
      <div className="bg-gradient-to-r from-green-500 to-green-600 text-white p-4">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="font-semibold text-lg">{session.name || 'Unnamed'}</h3>
            <p className="text-green-100 text-sm">ID: {session.phone}</p>
            {session.actual_phone && (
              <p className="text-green-100 text-xs">Phone: {session.actual_phone}</p>
            )}
          </div>
          <div className="text-right">
            <div className="flex items-center mb-1">
              <i className={`fas fa-circle text-xs mr-1 ${
                session.connected ? 'text-green-300' : 'text-red-300'
              }`}></i>
              <span className="text-xs">
                {session.connected ? 'Connected' : 'Disconnected'}
              </span>
            </div>
            <div className="flex items-center">
              <i className={`fas fa-user-check text-xs mr-1 ${
                session.logged_in ? 'text-green-300' : 'text-gray-300'
              }`}></i>
              <span className="text-xs">
                {session.logged_in ? 'Logged In' : 'Not Logged In'}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Session Actions */}
      <div className="p-4">
        <div className="flex flex-wrap gap-2">
          {!session.logged_in && (
            <button
              onClick={() => onShowQR(session)}
              className="flex-1 bg-blue-600 hover:bg-blue-700 text-white px-3 py-2 rounded text-sm font-medium transition duration-200"
            >
              <i className="fas fa-qrcode mr-1"></i>Show QR
            </button>
          )}
          
          {session.logged_in && (
            <>
              <button
                onClick={() => onLogout(session.id)}
                className="flex-1 bg-orange-600 hover:bg-orange-700 text-white px-3 py-2 rounded text-sm font-medium transition duration-200"
              >
                <i className="fas fa-sign-out-alt mr-1"></i>Logout
              </button>
              <button
                onClick={() => onSendMessage(session)}
                className="flex-1 bg-green-600 hover:bg-green-700 text-white px-3 py-2 rounded text-sm font-medium transition duration-200"
              >
                <i className="fas fa-paper-plane mr-1"></i>Send
              </button>
            </>
          )}
          
          <button
            onClick={() => onEdit(session)}
            className="bg-gray-600 hover:bg-gray-700 text-white px-3 py-2 rounded text-sm font-medium transition duration-200"
          >
            <i className="fas fa-edit mr-1"></i>Edit
          </button>
          
          <button
            onClick={() => onDelete(session.id)}
            className="bg-red-600 hover:bg-red-700 text-white px-3 py-2 rounded text-sm font-medium transition duration-200"
          >
            <i className="fas fa-trash mr-1"></i>Delete
          </button>
        </div>
      </div>
    </div>
  );
};

export default SessionCard;