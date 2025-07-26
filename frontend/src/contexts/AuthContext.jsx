import { createContext, useContext, useState, useEffect } from 'react';
import axios from 'axios';

const AuthContext = createContext();

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [token, setToken] = useState(localStorage.getItem('auth_token'));

  useEffect(() => {
    if (token) {
      axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
      const userData = localStorage.getItem('user_data');
      if (userData) {
        setUser(JSON.parse(userData));
      }
    }
    setLoading(false);
  }, [token]);

  useEffect(() => {
    // Add response interceptor to handle 401 errors
    const interceptor = axios.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response && error.response.status === 401) {
          logout();
        }
        return Promise.reject(error);
      }
    );

    return () => {
      axios.interceptors.response.eject(interceptor);
    };
  }, []);

  const login = async (username, password) => {
    try {
      const response = await axios.post('/api/login', {
        username,
        password,
      });

      if (response.data.success) {
        const { token: newToken, user: userData } = response.data;
        
        localStorage.setItem('auth_token', newToken);
        localStorage.setItem('user_data', JSON.stringify(userData));
        
        setToken(newToken);
        setUser(userData);
        
        axios.defaults.headers.common['Authorization'] = `Bearer ${newToken}`;
        
        return { success: true };
      }
    } catch (error) {
      return {
        success: false,
        error: error.response?.data || 'Login failed',
      };
    }
  };

  const logout = () => {
    localStorage.removeItem('auth_token');
    localStorage.removeItem('user_data');
    delete axios.defaults.headers.common['Authorization'];
    
    setToken(null);
    setUser(null);
  };

  const value = {
    user,
    token,
    loading,
    login,
    logout,
    isAuthenticated: !!user && !!token,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};