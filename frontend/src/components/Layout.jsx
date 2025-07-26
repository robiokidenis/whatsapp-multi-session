import { useAuth } from '../contexts/AuthContext';

const Layout = ({ children }) => {
  const { user, logout } = useAuth();

  return (
    <div className="min-h-screen bg-gray-100">
      {/* Header */}
      <header className="bg-green-600 text-white shadow-lg">
        <div className="container mx-auto px-4 py-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <i className="fab fa-whatsapp text-3xl mr-3"></i>
              <h1 className="text-2xl font-bold">WhatsApp Multi-Session Manager</h1>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-sm">Welcome, {user?.username}!</span>
              <button
                onClick={logout}
                className="bg-red-600 hover:bg-red-700 px-3 py-1 rounded text-sm transition duration-200"
              >
                <i className="fas fa-sign-out-alt mr-1"></i>
                Logout
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto px-4 py-8">
        {children}
      </main>
    </div>
  );
};

export default Layout;