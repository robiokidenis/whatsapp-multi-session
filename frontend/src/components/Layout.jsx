import { useAuth } from '../contexts/AuthContext';
import { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';

const Layout = ({ children }) => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  
  const activeTab = location.pathname.includes('user-management') ? 'users' : 'sessions';

  const handleTabChange = (tab) => {
    if (tab === 'sessions') {
      navigate('/');
    } else if (tab === 'users') {
      navigate('/user-management');
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Professional Header */}
      <header className="bg-white border-b border-gray-200 shadow-sm">
        <div className="max-w-7xl mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            {/* Brand Section */}
            <div className="flex items-center">
              <div className="bg-green-600 p-3 rounded-lg mr-4">
                <i className="fab fa-whatsapp text-white text-2xl"></i>
              </div>
              <div>
                <h1 className="text-2xl font-bold text-gray-900">
                  WhatsApp Manager
                </h1>
                <p className="text-gray-600 text-sm">Multi-Session Platform</p>
              </div>
            </div>

            {/* User Section */}
            <div className="flex items-center gap-4">
              {/* User Info */}
              <div className="hidden sm:flex items-center bg-gray-100 rounded-lg px-4 py-2">
                <div className="flex items-center justify-center w-8 h-8 bg-gray-300 rounded-full mr-3">
                  <i className={`fas ${user?.role === 'admin' ? 'fa-user-shield' : 'fa-user'} text-gray-600 text-sm`}></i>
                </div>
                <div>
                  <div className="text-sm font-medium text-gray-900">{user?.username}</div>
                  <div className="text-xs text-gray-500 capitalize">{user?.role}</div>
                </div>
              </div>

              {/* Logout Button */}
              <button
                onClick={logout}
                className="bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-lg font-medium transition-colors duration-200 flex items-center"
              >
                <i className="fas fa-sign-out-alt mr-2"></i>
                <span className="hidden sm:inline">Logout</span>
              </button>
            </div>
          </div>

          {/* Navigation Tabs */}
          <div className="mt-6 border-b border-gray-200">
            <nav className="flex space-x-8">
              <button
                onClick={() => handleTabChange('sessions')}
                className={`pb-4 px-1 border-b-2 font-medium text-sm transition-colors duration-200 flex items-center ${
                  activeTab === 'sessions'
                    ? 'border-green-600 text-green-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                <i className="fas fa-mobile-alt mr-2"></i>
                Sessions
              </button>
              {user?.role === 'admin' && (
                <button
                  onClick={() => handleTabChange('users')}
                  className={`pb-4 px-1 border-b-2 font-medium text-sm transition-colors duration-200 flex items-center ${
                    activeTab === 'users'
                      ? 'border-green-600 text-green-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                  }`}
                >
                  <i className="fas fa-users mr-2"></i>
                  User Management
                </button>
              )}
            </nav>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-6 py-8">
        {children}
      </main>
    </div>
  );
};

export default Layout;