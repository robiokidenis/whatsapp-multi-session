import { useState } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const Login = () => {
  const { login, isAuthenticated } = useAuth();
  const [formData, setFormData] = useState({
    username: '',
    password: '',
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  if (isAuthenticated) {
    return <Navigate to="/" replace />;
  }

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!formData.username || !formData.password) {
      setError('Please enter username and password');
      return;
    }

    setLoading(true);
    setError('');

    const result = await login(formData.username, formData.password);
    
    if (!result.success) {
      setError(result.error);
    }
    
    setLoading(false);
  };

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  return (
    <div className="min-h-screen bg-gray-100 flex items-center justify-center">
      <div className="bg-white shadow-2xl rounded-lg max-w-md w-full mx-4">
        {/* Header */}
        <div className="bg-green-600 text-white p-6 rounded-t-lg text-center">
          <i className="fab fa-whatsapp text-5xl mb-4"></i>
          <h1 className="text-2xl font-bold mb-2">WhatsApp Manager</h1>
          <p className="text-green-100">Administrative Login</p>
        </div>

        {/* Login Form */}
        <div className="p-6">
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                <i className="fas fa-user mr-2"></i>Username
              </label>
              <input
                type="text"
                name="username"
                value={formData.username}
                onChange={handleChange}
                required
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent transition duration-200"
                placeholder="Enter your username"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                <i className="fas fa-lock mr-2"></i>Password
              </label>
              <input
                type="password"
                name="password"
                value={formData.password}
                onChange={handleChange}
                required
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent transition duration-200"
                placeholder="Enter your password"
              />
            </div>

            {error && (
              <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded-lg">
                <i className="fas fa-exclamation-circle mr-2"></i>{error}
              </div>
            )}

            <button
              type="submit"
              disabled={loading || !formData.username || !formData.password}
              className="w-full bg-green-600 hover:bg-green-700 disabled:bg-gray-400 text-white font-medium py-3 px-4 rounded-lg transition duration-200 flex items-center justify-center"
            >
              {loading ? (
                <>
                  <i className="fas fa-spinner fa-spin mr-2"></i>
                  Logging in...
                </>
              ) : (
                <>
                  <i className="fas fa-sign-in-alt mr-2"></i>
                  Login
                </>
              )}
            </button>
          </form>
        </div>

        {/* Footer */}
        <div className="bg-gray-50 px-6 py-4 rounded-b-lg text-center text-sm text-gray-600">
          <p>
            <i className="fas fa-shield-alt mr-1"></i>Secure Administrative Access
          </p>
          <p className="mt-2">Only authorized administrators can access this system</p>
        </div>
      </div>
    </div>
  );
};

export default Login;