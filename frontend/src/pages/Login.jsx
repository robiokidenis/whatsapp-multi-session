import { useState, useEffect } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { useNotification } from '../contexts/NotificationContext';

const Login = () => {
  const { login, isAuthenticated } = useAuth();
  const { showError } = useNotification();
  const [formData, setFormData] = useState({
    username: '',
    password: '',
  });
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [rememberMe, setRememberMe] = useState(false);
  const [mounted, setMounted] = useState(false);
  
  useEffect(() => {
    setMounted(true);
    // Check for saved username
    const savedUsername = localStorage.getItem('savedUsername');
    if (savedUsername) {
      setFormData(prev => ({ ...prev, username: savedUsername }));
      setRememberMe(true);
    }
  }, []);

  if (isAuthenticated) {
    return <Navigate to="/" replace />;
  }

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!formData.username || !formData.password) {
      showError('Please enter username and password');
      return;
    }

    setLoading(true);

    // Save username if remember me is checked
    if (rememberMe) {
      localStorage.setItem('savedUsername', formData.username);
    } else {
      localStorage.removeItem('savedUsername');
    }

    const result = await login(formData.username, formData.password);
    
    if (!result.success) {
      showError(result.error);
      
      // Add shake animation on error
      const form = e.target;
      form.classList.add('animate-shake');
      setTimeout(() => form.classList.remove('animate-shake'), 500);
      
      // If rate limited, disable the form for a short time
      if (result.rateLimited) {
        // You could add additional UI feedback here like disabling the button
        // or showing a countdown timer
      }
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
    <div className="min-h-screen bg-gradient-to-br from-emerald-50 via-teal-50 to-cyan-50 flex items-center justify-center p-4 relative overflow-hidden">
      {/* Animated background elements */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute -top-40 -right-40 w-60 h-60 bg-emerald-200 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-blob"></div>
        <div className="absolute -bottom-40 -left-40 w-60 h-60 bg-teal-200 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-blob animation-delay-2000"></div>
        <div className="absolute top-40 left-40 w-60 h-60 bg-cyan-200 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-blob animation-delay-4000"></div>
      </div>
      
      <div className={`w-full max-w-md mx-auto relative ${mounted ? 'animate-fadeIn' : 'opacity-0'}`}>
        <div className="bg-white/95 backdrop-blur-lg rounded-2xl shadow-2xl overflow-hidden border border-white/20">
          {/* Header with gradient */}
          <div className="relative bg-gradient-to-r from-emerald-500 to-teal-600 text-white p-6 text-center">
            <div className="absolute inset-0 bg-black/10"></div>
            <div className="relative">
              <div className="inline-flex items-center justify-center w-16 h-16 bg-white/20 backdrop-blur-sm rounded-xl mb-3 animate-pulse">
                <i className="fab fa-whatsapp text-3xl text-white"></i>
              </div>
              <h1 className="text-2xl font-bold mb-1 tracking-tight">WhatsApp Manager</h1>
              <p className="text-emerald-100/90 font-medium text-sm">Multi-Session Platform</p>
            </div>
          </div>

          {/* Login Form */}
          <div className="p-6">
            <div className="text-center mb-6">
              <h2 className="text-xl font-bold text-gray-900 mb-1">Welcome Back</h2>
              <p className="text-gray-600 text-sm">Sign in to manage your sessions</p>
            </div>

            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-1">
                <label className="block text-sm font-semibold text-gray-700">
                  Username
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <i className="fas fa-user text-gray-400"></i>
                  </div>
                  <input
                    type="text"
                    name="username"
                    autoComplete="username"
                    value={formData.username}
                    onChange={handleChange}
                    required
                    className="w-full pl-10 pr-3 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent transition-all duration-200 bg-gray-50/50 hover:bg-white text-sm"
                    placeholder="Enter your username"
                  />
                </div>
              </div>

              <div className="space-y-1">
                <label className="block text-sm font-semibold text-gray-700">
                  Password
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <i className="fas fa-lock text-gray-400"></i>
                  </div>
                  <input
                    type={showPassword ? "text" : "password"}
                    name="password"
                    autoComplete="current-password"
                    value={formData.password}
                    onChange={handleChange}
                    required
                    className="w-full pl-10 pr-12 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent transition-all duration-200 bg-gray-50/50 hover:bg-white text-sm"
                    placeholder="Enter your password"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-500 hover:text-gray-700 transition-colors"
                  >
                    <i className={`fas ${showPassword ? 'fa-eye-slash' : 'fa-eye'} text-sm`}></i>
                  </button>
                </div>
              </div>

              {/* Remember me checkbox */}
              <div className="flex items-center justify-between">
                <label className="flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    checked={rememberMe}
                    onChange={(e) => setRememberMe(e.target.checked)}
                    className="w-4 h-4 text-emerald-600 bg-gray-100 border-gray-300 rounded focus:ring-emerald-500 focus:ring-2"
                  />
                  <span className="ml-2 text-sm text-gray-700">Remember me</span>
                </label>
                <a href="#" className="text-sm text-emerald-600 hover:text-emerald-700 font-medium transition-colors">
                  Forgot password?
                </a>
              </div>


              <button
                type="submit"
                disabled={loading || !formData.username || !formData.password}
                className="w-full bg-gradient-to-r from-emerald-500 to-teal-600 text-white py-2.5 rounded-lg font-semibold text-sm
                         hover:from-emerald-600 hover:to-teal-700 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:ring-offset-2 
                         disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:from-emerald-500 disabled:hover:to-teal-600
                         transform transition-all duration-200 hover:scale-[1.02] active:scale-[0.98]"
              >
                {loading ? (
                  <span className="flex items-center justify-center">
                    <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Signing in...
                  </span>
                ) : (
                  <span className="flex items-center justify-center">
                    <i className="fas fa-sign-in-alt mr-2"></i>
                    Sign In
                  </span>
                )}
              </button>
            </form>
          </div>

          {/* Footer */}
          <div className="bg-gradient-to-b from-gray-50 to-gray-100 px-6 py-4 text-center border-t border-gray-200">
            <div className="flex items-center justify-center text-gray-700 mb-1">
              <i className="fas fa-shield-alt mr-2 text-emerald-600 text-sm"></i>
              <span className="font-semibold text-sm">Secure Authentication</span>
            </div>
            <p className="text-xs text-gray-600">Protected by enterprise-grade security</p>
            <div className="mt-3 flex items-center justify-center space-x-3 text-xs text-gray-500">
              <span>© 2024 WhatsApp Manager</span>
              <span>•</span>
              <a href="#" className="hover:text-emerald-600 transition-colors">Privacy</a>
              <span>•</span>
              <a href="#" className="hover:text-emerald-600 transition-colors">Terms</a>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Login;